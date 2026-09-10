package webrtc

import (
	"testing"

	"github.com/m1k1o/neko/server/internal/webrtc/payload"
)

func TestViewOnlyDataEventAllowed(t *testing.T) {
	if !viewOnlyDataEventAllowed(payload.OP_PING) {
		t.Fatal("data-channel ping was denied")
	}

	denied := []uint8{
		payload.OP_MOVE,
		payload.OP_SCROLL,
		payload.OP_KEY_DOWN,
		payload.OP_KEY_UP,
		payload.OP_BTN_DOWN,
		payload.OP_BTN_UP,
		payload.OP_TOUCH_BEGIN,
		payload.OP_TOUCH_UPDATE,
		payload.OP_TOUCH_END,
	}
	for _, eventName := range denied {
		if viewOnlyDataEventAllowed(eventName) {
			t.Errorf("interactive data-channel event %#x was allowed", eventName)
		}
	}
}
