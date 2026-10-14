package mediahls

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/pkg/types"
)

type bootstrapContextKey struct{}

type bootstrapAttachment struct {
	mode string
}

type Backend struct {
	packager *Packager
	leases   *LeaseStore
	logger   zerolog.Logger

	mu         sync.Mutex
	deliveries map[string]*Delivery
	closed     bool
}

func NewBackend(provider types.EncodedMediaProvider, leases *LeaseStore) (*Backend, error) {
	if leases == nil {
		return nil, ErrInvalidConfig
	}
	packager, err := NewPackager(provider)
	if err != nil {
		return nil, err
	}
	return &Backend{
		packager: packager,
		leases: leases,
		logger: log.With().Str("module", "mediahls").Str("submodule", "delivery").Logger(),
		deliveries: make(map[string]*Delivery),
	}, nil
}

func (*Backend) Name() string { return BackendName }

func (*Backend) Capabilities() types.MediaBackendCapabilities {
	return types.MediaBackendCapabilities{ReceiveAudio: true, ReceiveVideo: true, ServerSideSelection: true}
}

func (backend *Backend) Open(ctx context.Context, lease types.MediaLease, request types.MediaDeliveryRequest) (types.MediaDelivery, error) {
	if ctx == nil || lease == nil || request.Backend != BackendName || !request.Audio || !request.Video {
		return nil, types.ErrMediaDeliveryNotAllowed
	}
	attachment, ok := ctx.Value(bootstrapContextKey{}).(*bootstrapAttachment)
	if !ok || attachment == nil || !ValidMode(attachment.mode) {
		return nil, types.ErrMediaDeliveryNotAllowed
	}
	backend.mu.Lock()
	closed := backend.closed
	backend.mu.Unlock()
	if closed {
		return nil, ErrPackagerClosed
	}
	packagerHeld := false
	if !request.InitialPaused {
		if err := backend.packager.Acquire(ctx, attachment.mode); err != nil {
			return nil, err
		}
		packagerHeld = true
	}
	offer, err := backend.leases.Issue(LeaseBinding{SessionID: lease.SessionID(), Mode: attachment.mode}, time.Now())
	if err != nil {
		if packagerHeld {
			backend.packager.Release()
		}
		return nil, err
	}
	state := "opening"
	if request.InitialPaused {
		state = "paused"
		if !backend.leases.SetPaused(lease.SessionID(), true) {
			backend.leases.Invalidate(offer.PublicID)
			return nil, ErrLeaseNotFound
		}
	}
	deliveryContext, deliveryCancel := context.WithCancel(context.Background())
	delivery := &Delivery{
		lease: lease, backend: backend, offer: offer, mode: attachment.mode,
		context: deliveryContext, cancel: deliveryCancel,
		done: make(chan struct{}), state: state, packagerHeld: packagerHeld,
	}
	backend.mu.Lock()
	if backend.closed {
		backend.mu.Unlock()
		deliveryCancel()
		backend.leases.Invalidate(offer.PublicID)
		if packagerHeld {
			backend.packager.Release()
		}
		return nil, ErrPackagerClosed
	}
	backend.deliveries[offer.PublicID] = delivery
	backend.mu.Unlock()
	hlsLeases.WithLabelValues(attachment.mode, state).Inc()
	backend.logger.Info().Str("session_id", delivery.SessionID()).Str("mode", delivery.mode).Str("state", state).Msg("HLS lease opened")
	return delivery, nil
}

func (backend *Backend) Delivery(publicID, sessionID string) (*Delivery, bool) {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	delivery, ok := backend.deliveries[publicID]
	return delivery, ok && delivery.SessionID() == sessionID
}

func (backend *Backend) remove(delivery *Delivery) {
	backend.mu.Lock()
	if current := backend.deliveries[delivery.offer.PublicID]; current == delivery {
		delete(backend.deliveries, delivery.offer.PublicID)
	}
	backend.mu.Unlock()
}

func (backend *Backend) Shutdown() {
	backend.mu.Lock()
	if backend.closed {
		backend.mu.Unlock()
		return
	}
	backend.closed = true
	deliveries := make([]*Delivery, 0, len(backend.deliveries))
	for _, delivery := range backend.deliveries {
		deliveries = append(deliveries, delivery)
	}
	backend.mu.Unlock()
	for _, delivery := range deliveries {
		_ = delivery.CloseWithReason(types.MediaDeliveryCloseShutdown)
	}
	for _, delivery := range deliveries {
		<-delivery.Done()
	}
	backend.packager.Shutdown()
}

type Delivery struct {
	lease   types.MediaLease
	backend *Backend
	offer   LeaseOffer
	mode    string
	context context.Context
	cancel  context.CancelFunc
	done    chan struct{}

	startOnce    sync.Once
	closeOnce    sync.Once
	mu           sync.Mutex
	state        string
	closed       bool
	packagerHeld bool
}

func (delivery *Delivery) ID() string            { return delivery.lease.ID() }
func (delivery *Delivery) SessionID() string     { return delivery.lease.SessionID() }
func (delivery *Delivery) Backend() string       { return BackendName }
func (delivery *Delivery) Done() <-chan struct{} { return delivery.done }
func (delivery *Delivery) Offer() LeaseOffer     { return delivery.offer }
func (delivery *Delivery) Mode() string          { return delivery.mode }

func (delivery *Delivery) Start() {
	delivery.startOnce.Do(func() { go delivery.watchExpiry() })
}

func (delivery *Delivery) watchExpiry() {
	for {
		expiresAt, err := delivery.backend.leases.Expiration(delivery.offer.PublicID, delivery.offer.Secret, time.Now())
		if err != nil {
			delivery.finish(true)
			return
		}
		wait := time.Until(expiresAt)
		if wait < 0 {
			wait = 0
		}
		timer := time.NewTimer(wait)
		select {
		case <-delivery.done:
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (delivery *Delivery) MarkActive() bool {
	delivery.mu.Lock()
	if delivery.closed {
		delivery.mu.Unlock()
		return false
	}
	if delivery.state == "active" {
		delivery.mu.Unlock()
		return true
	}
	if delivery.state != "opening" {
		delivery.mu.Unlock()
		return false
	}
	previous := delivery.state
	delivery.state = "active"
	delivery.mu.Unlock()
	hlsLeases.WithLabelValues(delivery.mode, previous).Dec()
	hlsLeases.WithLabelValues(delivery.mode, "active").Inc()
	delivery.backend.logger.Info().Str("session_id", delivery.SessionID()).Str("mode", delivery.mode).Str("state", "active").Msg("HLS lease state changed")
	if delivery.lease.SetState(types.MediaDeliveryStateActive) {
		return true
	}
	delivery.finish(true)
	return false
}

func (delivery *Delivery) SetPaused(paused bool) error {
	if !paused {
		delivery.mu.Lock()
		if delivery.closed {
			delivery.mu.Unlock()
			return ErrPackagerClosed
		}
		if delivery.state != "paused" {
			delivery.mu.Unlock()
			return nil
		}
		delivery.state = "opening"
		delivery.mu.Unlock()
		hlsLeases.WithLabelValues(delivery.mode, "paused").Dec()
		hlsLeases.WithLabelValues(delivery.mode, "opening").Inc()
		if !delivery.lease.SetState(types.MediaDeliveryStateOpening) {
			delivery.finish(true)
			return ErrPackagerClosed
		}
		delivery.backend.logger.Info().Str("session_id", delivery.SessionID()).Str("mode", delivery.mode).Str("state", "opening").Msg("HLS lease resume warming")
		go delivery.resume()
		return nil
	}
	delivery.mu.Lock()
	if delivery.closed {
		delivery.mu.Unlock()
		return nil
	}
	if delivery.state == "paused" {
		delivery.mu.Unlock()
		return nil
	}
	if !delivery.backend.leases.SetPaused(delivery.SessionID(), true) {
		delivery.mu.Unlock()
		return ErrLeaseNotFound
	}
	previous := delivery.state
	delivery.state = "paused"
	held := delivery.packagerHeld
	delivery.packagerHeld = false
	delivery.mu.Unlock()
	hlsLeases.WithLabelValues(delivery.mode, previous).Dec()
	hlsLeases.WithLabelValues(delivery.mode, "paused").Inc()
	delivery.backend.logger.Info().Str("session_id", delivery.SessionID()).Str("mode", delivery.mode).Str("state", "paused").Msg("HLS lease paused")
	if held {
		delivery.backend.packager.Release()
	}
	delivery.lease.SetState(types.MediaDeliveryStateOpening)
	return nil
}

func (delivery *Delivery) resume() {
	timeout := ConventionalReadyWindow
	if delivery.mode == ModeLLHLS {
		timeout = LowLatencyReadyWindow
	}
	ctx, cancel := context.WithTimeout(delivery.context, timeout)
	err := delivery.backend.packager.Acquire(ctx, delivery.mode)
	cancel()
	if err != nil {
		delivery.mu.Lock()
		failed := !delivery.closed && delivery.state == "opening"
		delivery.mu.Unlock()
		if failed {
			delivery.backend.logger.Warn().Err(err).Str("session_id", delivery.SessionID()).Str("mode", delivery.mode).Msg("HLS lease resume failed")
			delivery.finish(true)
		}
		return
	}
	delivery.mu.Lock()
	if delivery.closed || delivery.state != "opening" {
		delivery.mu.Unlock()
		delivery.backend.packager.Release()
		return
	}
	if !delivery.backend.leases.SetPaused(delivery.SessionID(), false) {
		delivery.mu.Unlock()
		delivery.backend.packager.Release()
		delivery.finish(true)
		return
	}
	delivery.packagerHeld = true
	delivery.mu.Unlock()
	delivery.backend.logger.Info().Str("session_id", delivery.SessionID()).Str("mode", delivery.mode).Str("state", "opening").Msg("HLS lease resumed")
}

func (delivery *Delivery) Close() error {
	delivery.finish(true)
	return nil
}

func (delivery *Delivery) CloseWithReason(_ types.MediaDeliveryCloseReason) error {
	delivery.finish(true)
	return nil
}

func (delivery *Delivery) finish(invalidate bool) {
	delivery.closeOnce.Do(func() {
		delivery.mu.Lock()
		delivery.closed = true
		state := delivery.state
		held := delivery.packagerHeld
		delivery.packagerHeld = false
		delivery.mu.Unlock()
		delivery.cancel()
		if invalidate {
			delivery.backend.leases.Invalidate(delivery.offer.PublicID)
		}
		delivery.backend.remove(delivery)
		if held {
			delivery.backend.packager.Release()
		}
		hlsLeases.WithLabelValues(delivery.mode, state).Dec()
		delivery.backend.logger.Info().Str("session_id", delivery.SessionID()).Str("mode", delivery.mode).Str("state", "closed").Msg("HLS lease closed")
		delivery.lease.SetState(types.MediaDeliveryStateClosed)
		close(delivery.done)
	})
}

func hlsOpenErrorStatus(err error) int {
	switch {
	case errors.Is(err, ErrLeasePaused):
		return 503
	case errors.Is(err, ErrPackagerNotReady), errors.Is(err, ErrCodecUnsupported), errors.Is(err, types.ErrMediaSourceNotFound):
		return 503
	case errors.Is(err, ErrLeaseLimit):
		return 429
	case errors.Is(err, types.ErrMediaDeliveryNotAllowed):
		return 403
	default:
		return 500
	}
}
