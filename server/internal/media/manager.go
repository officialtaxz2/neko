package media

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/pkg/types"
)

const deliveryShutdownTimeout = 5 * time.Second

var (
	mediaDeliveries = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "deliveries",
		Namespace: "neko",
		Subsystem: "media",
		Help:      "Current participant media deliveries by backend and lifecycle state.",
	}, []string{"backend", "state"})
	mediaDeliveryOpens = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "delivery_opens_total",
		Namespace: "neko",
		Subsystem: "media",
		Help:      "Participant media delivery open attempts.",
	}, []string{"backend", "result"})
	mediaDeliveryCloses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "delivery_closes_total",
		Namespace: "neko",
		Subsystem: "media",
		Help:      "Participant media deliveries closed by the central manager.",
	}, []string{"backend", "state"})
)

type deliveryEntry struct {
	lease    *lease
	delivery types.MediaDelivery
	session  types.Session
	state    types.MediaDeliveryState
}

type ManagerCtx struct {
	logger   zerolog.Logger
	sessions types.SessionManager
	serial   atomic.Uint64

	mu         sync.Mutex
	backends   map[string]types.MediaBackend
	deliveries map[string]*deliveryEntry
	shutdown   bool
}

func New(sessions types.SessionManager) *ManagerCtx {
	manager := &ManagerCtx{
		logger:     log.With().Str("module", "media").Logger(),
		sessions:   sessions,
		backends:   make(map[string]types.MediaBackend),
		deliveries: make(map[string]*deliveryEntry),
	}

	sessions.OnDeleted(func(session types.Session) {
		manager.CloseSession(session.ID())
	})
	sessions.OnProfileChanged(func(session types.Session, profile, _ types.MemberProfile) {
		if !profile.CanWatch {
			manager.CloseSession(session.ID())
		}
	})

	return manager
}

func (manager *ManagerCtx) Register(backend types.MediaBackend) error {
	if backend == nil || backend.Name() == "" {
		return errors.New("media backend name cannot be empty")
	}

	manager.mu.Lock()
	defer manager.mu.Unlock()
	if manager.shutdown {
		return errors.New("media delivery manager is shut down")
	}
	if _, exists := manager.backends[backend.Name()]; exists {
		return fmt.Errorf("%w: %s", types.ErrMediaBackendAlreadyExists, backend.Name())
	}
	manager.backends[backend.Name()] = backend
	return nil
}

func (manager *ManagerCtx) Backends() []types.MediaBackendDescriptor {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	backends := make([]types.MediaBackendDescriptor, 0, len(manager.backends))
	for _, backend := range manager.backends {
		backends = append(backends, types.MediaBackendDescriptor{
			Name:         backend.Name(),
			Capabilities: backend.Capabilities(),
		})
	}
	sort.Slice(backends, func(i, j int) bool {
		return backends[i].Name < backends[j].Name
	})
	return backends
}

func (manager *ManagerCtx) Open(ctx context.Context, session types.Session, request types.MediaDeliveryRequest) (types.MediaDelivery, error) {
	if session == nil || !session.Profile().CanWatch {
		return nil, types.ErrMediaDeliveryNotAllowed
	}
	current, ok := manager.sessions.Get(session.ID())
	if !ok || current != session {
		return nil, types.ErrMediaDeliveryNotAllowed
	}

	manager.mu.Lock()
	if manager.shutdown {
		manager.mu.Unlock()
		return nil, errors.New("media delivery manager is shut down")
	}
	backend, ok := manager.backends[request.Backend]
	manager.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", types.ErrMediaBackendNotFound, request.Backend)
	}

	capabilities := backend.Capabilities()
	request.Audio = request.Audio && capabilities.ReceiveAudio
	request.Video = request.Video && capabilities.ReceiveVideo
	if !request.Audio && !request.Video {
		return nil, types.ErrMediaDeliveryNotAllowed
	}

	lease := &lease{
		id:        request.Backend + "-" + strconv.FormatUint(manager.serial.Add(1), 10),
		sessionID: session.ID(),
		backend:   request.Backend,
		manager:   manager,
		state:     types.MediaDeliveryStateOpening,
		valid:     true,
	}

	mediaDeliveryOpens.WithLabelValues(request.Backend, "attempt").Inc()
	delivery, err := backend.Open(ctx, lease, request)
	if err != nil {
		lease.invalidate()
		mediaDeliveryOpens.WithLabelValues(request.Backend, "error").Inc()
		return nil, err
	}
	var done <-chan struct{}
	if delivery != nil {
		done = delivery.Done()
	}
	if delivery == nil || done == nil || delivery.ID() != lease.ID() || delivery.SessionID() != session.ID() || delivery.Backend() != request.Backend {
		lease.invalidate()
		if delivery != nil {
			_ = delivery.Close()
		}
		mediaDeliveryOpens.WithLabelValues(request.Backend, "error").Inc()
		return nil, errors.New("media backend returned a delivery outside its lease")
	}

	entry := &deliveryEntry{
		lease:    lease,
		delivery: delivery,
		session:  session,
		state:    types.MediaDeliveryStateOpening,
	}

	manager.mu.Lock()
	if manager.shutdown {
		manager.mu.Unlock()
		lease.invalidate()
		_ = delivery.Close()
		mediaDeliveryOpens.WithLabelValues(request.Backend, "error").Inc()
		return nil, errors.New("media delivery manager is shut down")
	}
	current, exists := manager.sessions.Get(session.ID())
	if !exists || current != session || !session.Profile().CanWatch {
		manager.mu.Unlock()
		lease.invalidate()
		_ = delivery.Close()
		mediaDeliveryOpens.WithLabelValues(request.Backend, "error").Inc()
		return nil, types.ErrMediaDeliveryNotAllowed
	}
	previous := manager.deliveries[session.ID()]
	manager.deliveries[session.ID()] = entry
	manager.mu.Unlock()

	mediaDeliveries.WithLabelValues(request.Backend, string(types.MediaDeliveryStateOpening)).Inc()
	mediaDeliveryOpens.WithLabelValues(request.Backend, "success").Inc()
	session.SetMediaDelivery(delivery)
	lease.activate()

	if previous != nil {
		previous.lease.invalidate()
		_ = previous.delivery.Close()
		manager.finishPrevious(previous)
	}

	go func() {
		<-done
		lease.SetState(types.MediaDeliveryStateClosed)
	}()

	return delivery, nil
}

func (manager *ManagerCtx) finishPrevious(entry *deliveryEntry) {
	mediaDeliveries.WithLabelValues(entry.delivery.Backend(), string(entry.state)).Dec()
	mediaDeliveryCloses.WithLabelValues(entry.delivery.Backend(), "replaced").Inc()
}

func (manager *ManagerCtx) setState(lease *lease, state types.MediaDeliveryState) bool {
	manager.mu.Lock()
	entry, ok := manager.deliveries[lease.sessionID]
	if !ok || entry.lease != lease {
		manager.mu.Unlock()
		return false
	}
	oldState := entry.state
	if oldState == state {
		manager.mu.Unlock()
		return true
	}
	entry.state = state
	closed := state == types.MediaDeliveryStateClosed || state == types.MediaDeliveryStateFailed
	if closed {
		delete(manager.deliveries, lease.sessionID)
	}
	manager.mu.Unlock()

	mediaDeliveries.WithLabelValues(entry.delivery.Backend(), string(oldState)).Dec()
	if closed {
		lease.invalidate()
		mediaDeliveryCloses.WithLabelValues(entry.delivery.Backend(), string(state)).Inc()
	} else {
		mediaDeliveries.WithLabelValues(entry.delivery.Backend(), string(state)).Inc()
	}

	entry.session.SetMediaDeliveryActive(entry.delivery, state == types.MediaDeliveryStateActive)
	return true
}

func (manager *ManagerCtx) isCurrent(lease *lease) bool {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	entry, ok := manager.deliveries[lease.sessionID]
	return ok && entry.lease == lease
}

func (manager *ManagerCtx) CloseSession(sessionID string) {
	manager.mu.Lock()
	entry := manager.deliveries[sessionID]
	manager.mu.Unlock()
	if entry != nil {
		_ = entry.delivery.Close()
	}
}

func (manager *ManagerCtx) Shutdown() error {
	manager.mu.Lock()
	if manager.shutdown {
		manager.mu.Unlock()
		return nil
	}
	manager.shutdown = true
	deliveries := make([]types.MediaDelivery, 0, len(manager.deliveries))
	for _, entry := range manager.deliveries {
		deliveries = append(deliveries, entry.delivery)
	}
	manager.mu.Unlock()

	for _, delivery := range deliveries {
		_ = delivery.Close()
	}

	deadline := time.NewTimer(deliveryShutdownTimeout)
	defer deadline.Stop()
	for _, delivery := range deliveries {
		select {
		case <-delivery.Done():
		case <-deadline.C:
			manager.logger.Warn().Msg("timed out waiting for media deliveries to close")
			return errors.New("timed out waiting for media deliveries to close")
		}
	}
	return nil
}

type lease struct {
	id        string
	sessionID string
	backend   string
	manager   *ManagerCtx

	mu        sync.Mutex
	state     types.MediaDeliveryState
	valid     bool
	activated bool
}

func (lease *lease) ID() string {
	return lease.id
}

func (lease *lease) SessionID() string {
	return lease.sessionID
}

func (lease *lease) Backend() string {
	return lease.backend
}

func (lease *lease) activate() {
	lease.mu.Lock()
	lease.activated = true
	state := lease.state
	valid := lease.valid
	lease.mu.Unlock()
	if valid && state != types.MediaDeliveryStateOpening {
		lease.manager.setState(lease, state)
	}
}

func (lease *lease) SetState(state types.MediaDeliveryState) bool {
	lease.mu.Lock()
	if !lease.valid {
		lease.mu.Unlock()
		return false
	}
	lease.state = state
	activated := lease.activated
	lease.mu.Unlock()
	if !activated {
		return true
	}
	return lease.manager.setState(lease, state)
}

func (lease *lease) Valid() bool {
	lease.mu.Lock()
	valid := lease.valid
	activated := lease.activated
	lease.mu.Unlock()
	return valid && (!activated || lease.manager.isCurrent(lease))
}

func (lease *lease) invalidate() {
	lease.mu.Lock()
	lease.valid = false
	lease.mu.Unlock()
}
