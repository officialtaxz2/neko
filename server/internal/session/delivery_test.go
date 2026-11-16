package session

import (
	"testing"

	"github.com/m1k1o/neko/server/pkg/types"
)

// fakeDelivery embeds the real interface so we only implement the methods sessions call.
type fakeDelivery struct {
	types.MediaDelivery
	closedWith types.MediaDeliveryCloseReason
}

func (d *fakeDelivery) SetPaused(bool) error { return nil }

func (d *fakeDelivery) CloseWithReason(reason types.MediaDeliveryCloseReason) error {
	d.closedWith = reason
	return nil
}

// Fork: like the webrtc peer upstream destroys in TestPeerDestroyedBeforeConnected,
// a media delivery of any backend that has not become active yet is revoked with
// its session, otherwise it keeps the capture pipelines running.
func TestMediaDeliveryClosedBeforeActive(t *testing.T) {
	tests := []struct {
		name string
		end  func(manager *SessionManagerCtx, session types.Session) error
	}{
		{"disconnect", func(manager *SessionManagerCtx, session types.Session) error {
			return manager.Disconnect(session.ID())
		}},
		{"delete", func(manager *SessionManagerCtx, session types.Session) error {
			return manager.Delete(session.ID())
		}},
		{"profile change", func(manager *SessionManagerCtx, session types.Session) error {
			return manager.Update(session.ID(), types.MemberProfile{CanConnect: true, CanLogin: true})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, session := newTestSession(t, types.MemberProfile{CanConnect: true, CanLogin: true, CanWatch: true})
			delivery := &fakeDelivery{}

			session.SetMediaDelivery(delivery)
			if session.State().IsWatching {
				t.Fatal("session is watching before the delivery became active")
			}

			if err := tt.end(manager, session); err != nil {
				t.Fatalf("could not end session: %v", err)
			}
			if delivery.closedWith != types.MediaDeliveryCloseRevoked {
				t.Errorf("delivery close reason = %q, want %q", delivery.closedWith, types.MediaDeliveryCloseRevoked)
			}
		})
	}
}
