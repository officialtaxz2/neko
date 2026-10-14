package legacy

import (
	"net/http/httptest"
	"testing"
)

func TestReceiveOnlySelectionIsExactAndDoesNotChangeTheDefault(t *testing.T) {
	for _, query := range []string{"?media=hls", "?media=ll-hls"} {
		if !receiveOnlyMediaSelected(httptest.NewRequest("GET", "https://neko.invalid/ws"+query, nil)) {
			t.Fatalf("receive-only opt-in rejected: %s", query)
		}
	}
	// The removed WebCodecs selector must fall back to the normal WebRTC path.
	for _, query := range []string{"", "?media=webrtc", "?media=webcodecs-ws", "?media=HLS", "?media=hls-extra", "?media=hls&media=ll-hls", "?media=hls&media=hls"} {
		if receiveOnlyMediaSelected(httptest.NewRequest("GET", "https://neko.invalid/ws"+query, nil)) {
			t.Fatalf("non-exact opt-in skipped WebRTC: %s", query)
		}
	}
}
