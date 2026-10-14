package session

import (
	"testing"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/pkg/types"
)

func TestViewOnlySessionMarkerCannotBeRemovedByProfileUpdate(t *testing.T) {
	manager := New(&config.Session{})
	session, _, err := manager.Create("viewer", types.NewViewOnlyMemberProfile("Viewer"))
	if err != nil {
		t.Fatalf("could not create view-only session: %v", err)
	}

	err = manager.Update(session.ID(), types.MemberProfile{
		Name:               "Escalated Viewer",
		IsAdmin:            true,
		CanLogin:           true,
		CanConnect:         true,
		CanWatch:           true,
		CanHost:            true,
		CanShareMedia:      true,
		CanAccessClipboard: true,
		SendsInactiveCursor: true,
	})
	if err != nil {
		t.Fatalf("could not update view-only session: %v", err)
	}

	profile := session.Profile()
	if !profile.IsViewOnly || profile.IsAdmin || profile.CanHost ||
		profile.CanShareMedia || profile.CanAccessClipboard || profile.SendsInactiveCursor {
		t.Fatalf("profile update escaped view-only boundary: %+v", profile)
	}
}
