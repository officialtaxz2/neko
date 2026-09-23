package mediahls

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/m1k1o/neko/server/pkg/types"
)

type fakeMediaLease struct {
	id        string
	sessionID string
	state     types.MediaDeliveryState
	valid     bool
}

func (lease *fakeMediaLease) ID() string        { return lease.id }
func (lease *fakeMediaLease) SessionID() string { return lease.sessionID }
func (*fakeMediaLease) Backend() string          { return BackendName }
func (lease *fakeMediaLease) SetState(state types.MediaDeliveryState) bool {
	if !lease.valid {
		return false
	}
	lease.state = state
	return true
}
func (lease *fakeMediaLease) Valid() bool { return lease.valid }

func TestInitiallyPrivateDeliveryAttachesPausedWithoutStartingPackager(t *testing.T) {
	leases, err := NewLeaseStore(MaximumLeases, MaximumRequests)
	if err != nil {
		t.Fatal(err)
	}
	packager := newPackager(newFakeHLSProvider(), fakeTranscoderFactory{})
	backend := &Backend{
		packager:   packager,
		leases:     leases,
		logger:     zerolog.Nop(),
		deliveries: make(map[string]*Delivery),
	}
	lease := &fakeMediaLease{id: "hls-1", sessionID: "viewer", state: types.MediaDeliveryStateOpening, valid: true}
	ctx := context.WithValue(context.Background(), bootstrapContextKey{}, &bootstrapAttachment{mode: ModeHLS})
	opened, err := backend.Open(ctx, lease, types.MediaDeliveryRequest{
		Backend: BackendName, Audio: true, Video: true, InitialPaused: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	delivery := opened.(*Delivery)
	if delivery.state != "paused" || delivery.packagerHeld || packager.refs != 0 || packager.running {
		t.Fatalf("paused delivery state = %q, held=%t, refs=%d, running=%t", delivery.state, delivery.packagerHeld, packager.refs, packager.running)
	}
	if _, err := leases.Authenticate(delivery.offer.PublicID, delivery.offer.Secret, false, time.Now()); !errors.Is(err, ErrLeasePaused) {
		t.Fatalf("paused lease authentication = %v", err)
	}
	if err := delivery.Close(); err != nil {
		t.Fatal(err)
	}
}
