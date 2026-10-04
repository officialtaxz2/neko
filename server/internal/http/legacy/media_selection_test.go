package legacy

import (
	"net/http/httptest"
	"testing"
)

func TestWebCodecsMediaSelectionIsExact(t *testing.T) {
	tests := []struct {
		query    string
		selected bool
	}{
		{"", false},
		{"?media=webrtc", false},
		{"?media=webcodecs-ws-extra", false},
		{"?media=webcodecs-ws", true},
		{"?media=webcodecs-ws&media=webcodecs-ws", false},
	}

	for _, test := range tests {
		request := httptest.NewRequest("GET", "http://neko.invalid/ws"+test.query, nil)
		if selected := webCodecsMediaSelected(request); selected != test.selected {
			t.Fatalf("query %q selected = %v, want %v", test.query, selected, test.selected)
		}
	}
}

func TestReceiveOnlySelectionIsExactAndDoesNotChangeTheDefault(t *testing.T) {
	for _, query := range []string{"?media=hls", "?media=ll-hls", "?media=webcodecs-ws"} {
		if !receiveOnlyMediaSelected(httptest.NewRequest("GET", "https://neko.invalid/ws"+query, nil)) {
			t.Fatalf("receive-only opt-in rejected: %s", query)
		}
	}
	for _, query := range []string{"", "?media=webrtc", "?media=HLS", "?media=hls-extra", "?media=hls&media=ll-hls", "?media=hls&media=hls"} {
		if receiveOnlyMediaSelected(httptest.NewRequest("GET", "https://neko.invalid/ws"+query, nil)) {
			t.Fatalf("non-exact opt-in skipped WebRTC: %s", query)
		}
	}
}
