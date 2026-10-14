package types

import (
	"reflect"
	"testing"
)

func TestRestrictViewOnlyRemovesInteractiveCapabilities(t *testing.T) {
	profile := MemberProfile{
		IsViewOnly:         true,
		IsAdmin:            true,
		CanLogin:           true,
		CanConnect:         true,
		CanWatch:           true,
		CanHost:            true,
		CanShareMedia:      true,
		CanAccessClipboard: true,
		SendsInactiveCursor: true,
	}

	restricted := profile.RestrictViewOnly()
	if !restricted.IsViewOnly || !restricted.CanLogin || !restricted.CanConnect || !restricted.CanWatch {
		t.Fatalf("view-only identity and receive permissions changed: %+v", restricted)
	}
	if restricted.IsAdmin || restricted.CanHost || restricted.CanShareMedia ||
		restricted.CanAccessClipboard || restricted.SendsInactiveCursor {
		t.Fatalf("interactive capability survived view-only restriction: %+v", restricted)
	}
	if restricted.IsInteractive() {
		t.Fatal("view-only profile reported itself as interactive")
	}
}

func TestRestrictViewOnlyLeavesMemberProfileUnchanged(t *testing.T) {
	profile := MemberProfile{IsAdmin: true, CanHost: true, CanShareMedia: true}
	if got := profile.RestrictViewOnly(); !reflect.DeepEqual(got, profile) {
		t.Fatalf("ordinary profile changed: got %+v want %+v", got, profile)
	}
}

func TestNewViewOnlyMemberProfile(t *testing.T) {
	profile := NewViewOnlyMemberProfile("viewer")
	if profile.Name != "viewer" || !profile.IsViewOnly || !profile.CanLogin ||
		!profile.CanConnect || !profile.CanWatch {
		t.Fatalf("unexpected view-only receive profile: %+v", profile)
	}
	if profile.IsAdmin || profile.CanHost || profile.CanShareMedia ||
		profile.CanAccessClipboard || profile.SendsInactiveCursor {
		t.Fatalf("view-only profile contains interactive capability: %+v", profile)
	}
	if canSend, ok := profile.Plugins["chat.can_send"].(bool); !ok || canSend {
		t.Fatalf("view-only chat send permission = %#v, want false", profile.Plugins["chat.can_send"])
	}
	if enabled, ok := profile.Plugins["filetransfer.enabled"].(bool); !ok || enabled {
		t.Fatalf("view-only file transfer permission = %#v, want false", profile.Plugins["filetransfer.enabled"])
	}
}
