package media_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/internal/config"
	mediadelivery "github.com/m1k1o/neko/server/internal/media"
	"github.com/m1k1o/neko/server/internal/session"
	"github.com/m1k1o/neko/server/pkg/types"
)

type fakeBackend struct {
	mu           sync.Mutex
	capabilities types.MediaBackendCapabilities
	lastLease    types.MediaLease
	lastRequest  types.MediaDeliveryRequest
	deliveries   []*fakeDelivery
}

func (backend *fakeBackend) Name() string { return "fake" }
func (backend *fakeBackend) Capabilities() types.MediaBackendCapabilities {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	return backend.capabilities
}
func (backend *fakeBackend) Open(_ context.Context, lease types.MediaLease, request types.MediaDeliveryRequest) (types.MediaDelivery, error) {
	delivery := &fakeDelivery{
		id:        lease.ID(),
		sessionID: lease.SessionID(),
		backend:   lease.Backend(),
		done:      make(chan struct{}),
	}
	backend.mu.Lock()
	backend.lastLease = lease
	backend.lastRequest = request
	backend.deliveries = append(backend.deliveries, delivery)
	backend.mu.Unlock()
	return delivery, nil
}

func (backend *fakeBackend) lease() types.MediaLease {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	return backend.lastLease
}

func (backend *fakeBackend) request() types.MediaDeliveryRequest {
	backend.mu.Lock()
	defer backend.mu.Unlock()
	return backend.lastRequest
}

type fakeDelivery struct {
	id        string
	sessionID string
	backend   string
	done      chan struct{}
	closeOnce sync.Once
	paused    bool
}

func (delivery *fakeDelivery) ID() string                 { return delivery.id }
func (delivery *fakeDelivery) SessionID() string          { return delivery.sessionID }
func (delivery *fakeDelivery) Backend() string            { return delivery.backend }
func (delivery *fakeDelivery) Done() <-chan struct{}       { return delivery.done }
func (delivery *fakeDelivery) SetPaused(paused bool) error { delivery.paused = paused; return nil }
func (delivery *fakeDelivery) Close() error {
	delivery.closeOnce.Do(func() { close(delivery.done) })
	return nil
}

func eventually(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("condition did not become true")
		}
		time.Sleep(time.Millisecond)
	}
}

func newDeliveryTestManager(t *testing.T) (*mediadelivery.ManagerCtx, *session.SessionManagerCtx, *fakeBackend) {
	t.Helper()
	sessions := session.New(&config.Session{})
	manager := mediadelivery.New(sessions)
	backend := &fakeBackend{
		capabilities: types.MediaBackendCapabilities{ReceiveAudio: true, ReceiveVideo: true},
	}
	if err := manager.Register(backend); err != nil {
		t.Fatalf("Register() error: %v", err)
	}
	return manager, sessions, backend
}

func TestDeliveryManagerIntersectsRequestedReceiveCapabilities(t *testing.T) {
	manager, sessions, backend := newDeliveryTestManager(t)
	backend.mu.Lock()
	backend.capabilities = types.MediaBackendCapabilities{ReceiveVideo: true}
	backend.mu.Unlock()
	member, _, err := sessions.Create("member", types.MemberProfile{CanLogin: true, CanConnect: true, CanWatch: true})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if _, err := manager.Open(context.Background(), member, types.MediaDeliveryRequest{Backend: "fake", Audio: true}); !errors.Is(err, types.ErrMediaDeliveryNotAllowed) {
		t.Fatalf("audio-only Open() error = %v, want ErrMediaDeliveryNotAllowed", err)
	}
	delivery, err := manager.Open(context.Background(), member, types.MediaDeliveryRequest{Backend: "fake", Audio: true, Video: true})
	if err != nil {
		t.Fatalf("audio/video Open() error: %v", err)
	}
	request := backend.request()
	if request.Audio || !request.Video {
		t.Fatalf("intersected request = %#v, want video only", request)
	}
	_ = delivery.Close()
}

func TestDeliveryManagerDeniesCanWatchFalseAndAllowsViewOnly(t *testing.T) {
	manager, sessions, backend := newDeliveryTestManager(t)
	denied, _, err := sessions.Create("denied", types.MemberProfile{CanLogin: true, CanConnect: true})
	if err != nil {
		t.Fatalf("Create(denied) error: %v", err)
	}
	if _, err := manager.Open(context.Background(), denied, types.MediaDeliveryRequest{Backend: "fake", Video: true}); !errors.Is(err, types.ErrMediaDeliveryNotAllowed) {
		t.Fatalf("Open(denied) error = %v, want ErrMediaDeliveryNotAllowed", err)
	}

	viewer, _, err := sessions.Create("viewer", types.NewViewOnlyMemberProfile("Viewer"))
	if err != nil {
		t.Fatalf("Create(viewer) error: %v", err)
	}
	delivery, err := manager.Open(context.Background(), viewer, types.MediaDeliveryRequest{Backend: "fake", Audio: true, Video: true})
	if err != nil {
		t.Fatalf("Open(view-only) error: %v", err)
	}
	if delivery.SessionID() != viewer.ID() || !viewer.Profile().IsViewOnly {
		t.Fatal("view-only delivery did not retain its passive session identity")
	}
	if !backend.lease().SetState(types.MediaDeliveryStateActive) {
		t.Fatal("active state was rejected for the current view-only delivery")
	}
	if !viewer.State().IsWatching {
		t.Fatal("active delivery did not set backend-neutral watching state")
	}

	manager.CloseSession(viewer.ID())
	eventually(t, func() bool { return !viewer.State().IsWatching && viewer.GetMediaDelivery() == nil })
}

func TestDeliveryManagerRevokesOnProfileChangeAndSessionDelete(t *testing.T) {
	manager, sessions, backend := newDeliveryTestManager(t)
	member, _, err := sessions.Create("member", types.MemberProfile{CanLogin: true, CanConnect: true, CanWatch: true})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if _, err := manager.Open(context.Background(), member, types.MediaDeliveryRequest{Backend: "fake", Video: true}); err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	backend.lease().SetState(types.MediaDeliveryStateActive)
	if err := sessions.Update(member.ID(), types.MemberProfile{CanLogin: true, CanConnect: true, CanWatch: false}); err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	eventually(t, func() bool { return member.GetMediaDelivery() == nil && !member.State().IsWatching })

	second, _, err := sessions.Create("second", types.MemberProfile{CanLogin: true, CanConnect: true, CanWatch: true})
	if err != nil {
		t.Fatalf("Create(second) error: %v", err)
	}
	if _, err := manager.Open(context.Background(), second, types.MediaDeliveryRequest{Backend: "fake", Video: true}); err != nil {
		t.Fatalf("Open(second) error: %v", err)
	}
	backend.lease().SetState(types.MediaDeliveryStateActive)
	if err := sessions.Delete(second.ID()); err != nil {
		t.Fatalf("Delete(second) error: %v", err)
	}
	eventually(t, func() bool {
		select {
		case <-backend.deliveries[len(backend.deliveries)-1].Done():
			return true
		default:
			return false
		}
	})
}

func TestDeliveryManagerReplacesOnePrimaryDeliveryAndShutsDown(t *testing.T) {
	manager, sessions, backend := newDeliveryTestManager(t)
	member, _, err := sessions.Create("member", types.MemberProfile{CanLogin: true, CanConnect: true, CanWatch: true})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	first, err := manager.Open(context.Background(), member, types.MediaDeliveryRequest{Backend: "fake", Video: true})
	if err != nil {
		t.Fatalf("first Open() error: %v", err)
	}
	second, err := manager.Open(context.Background(), member, types.MediaDeliveryRequest{Backend: "fake", Video: true})
	if err != nil {
		t.Fatalf("second Open() error: %v", err)
	}
	select {
	case <-first.Done():
	case <-time.After(time.Second):
		t.Fatal("replacement did not close the previous primary delivery")
	}
	if member.GetMediaDelivery() != second {
		t.Fatal("replacement delivery is not attached to the session")
	}
	if err := manager.Shutdown(); err != nil {
		t.Fatalf("Shutdown() error: %v", err)
	}
	select {
	case <-second.Done():
	case <-time.After(time.Second):
		t.Fatal("Shutdown() did not close the active delivery")
	}
}
