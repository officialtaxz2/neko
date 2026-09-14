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
