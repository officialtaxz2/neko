package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/pkg/types"
)

func TestViewOnlySessionsAreNotPersistedOrRestored(t *testing.T) {
	file := filepath.Join(t.TempDir(), "sessions.json")
	manager := New(&config.Session{File: file})

	if _, _, err := manager.Create("viewer", types.NewViewOnlyMemberProfile("Viewer")); err != nil {
		t.Fatalf("could not create view-only session: %v", err)
	}
	if _, _, err := manager.Create("member", types.MemberProfile{Name: "Member"}); err != nil {
		t.Fatalf("could not create ordinary session: %v", err)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("could not read persisted sessions: %v", err)
	}
	var saved []types.SessionProfile
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("could not decode persisted sessions: %v", err)
	}
	if len(saved) != 1 || saved[0].Id != "member" {
		t.Fatalf("persisted sessions = %+v, want only ordinary member", saved)
	}

	legacyViewOnly := types.SessionProfile{
		Id:      "stale-viewer",
		Token:   "stale-token",
		Profile: types.NewViewOnlyMemberProfile("Stale Viewer"),
	}
	data, err = json.Marshal(append(saved, legacyViewOnly))
	if err != nil {
		t.Fatalf("could not encode legacy sessions: %v", err)
	}
	if err := os.WriteFile(file, data, 0o600); err != nil {
		t.Fatalf("could not write legacy sessions: %v", err)
	}

	restored := New(&config.Session{File: file})
	if _, ok := restored.Get("member"); !ok {
		t.Fatal("ordinary session was not restored")
	}
	if _, ok := restored.Get("stale-viewer"); ok {
		t.Fatal("view-only session was restored")
	}
}
