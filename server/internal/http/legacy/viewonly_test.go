package legacy

import (
	"net/http/httptest"
	"testing"

	oldEvent "github.com/m1k1o/neko/server/internal/http/legacy/event"
)

const testViewOnlyToken = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestViewOnlyTokenFromRequest(t *testing.T) {
	r := httptest.NewRequest("GET", "http://example.test/ws?username=Guest", nil)
	r.Header.Set("Sec-WebSocket-Protocol", "neko-view."+testViewOnlyToken)

	token, protocol, err := viewOnlyTokenFromRequest(r)
	if err != nil {
		t.Fatalf("valid subprotocol rejected: %v", err)
	}
	if token != testViewOnlyToken || protocol != "neko-view."+testViewOnlyToken {
		t.Fatalf("unexpected token/protocol: %q %q", token, protocol)
	}
}

func TestViewOnlyTokenFromRequestRejectsInvalidOrDuplicateValues(t *testing.T) {
	invalid := httptest.NewRequest("GET", "http://example.test/ws", nil)
	invalid.Header.Set("Sec-WebSocket-Protocol", "neko-view.short")
	if _, _, err := viewOnlyTokenFromRequest(invalid); err == nil {
		t.Fatal("invalid token shape was accepted")
	}

	duplicate := httptest.NewRequest("GET", "http://example.test/ws", nil)
	duplicate.Header.Set("Sec-WebSocket-Protocol", "neko-view."+testViewOnlyToken+", neko-view."+testViewOnlyToken)
	if _, _, err := viewOnlyTokenFromRequest(duplicate); err == nil {
		t.Fatal("duplicate view-only protocols were accepted")
	}
}

func TestViewOnlyLegacyEventAllowed(t *testing.T) {
	for _, eventName := range []string{
		oldEvent.CLIENT_HEARTBEAT,
		oldEvent.SIGNAL_OFFER,
		oldEvent.SIGNAL_ANSWER,
		oldEvent.SIGNAL_CANDIDATE,
	} {
		if !viewOnlyLegacyEventAllowed(eventName) {
			t.Errorf("signalling event %q was denied", eventName)
		}
	}

	for _, eventName := range []string{
		oldEvent.CONTROL_REQUEST,
		oldEvent.CONTROL_CLIPBOARD,
		oldEvent.CONTROL_KEYBOARD,
		oldEvent.CHAT_MESSAGE,
		oldEvent.CHAT_EMOTE,
		oldEvent.FILETRANSFER_REFRESH,
		oldEvent.SCREEN_SET,
		oldEvent.ADMIN_CONTROL,
	} {
		if viewOnlyLegacyEventAllowed(eventName) {
			t.Errorf("interactive event %q was allowed", eventName)
		}
	}
}
