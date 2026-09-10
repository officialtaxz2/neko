package websocket

import (
	"testing"

	"github.com/m1k1o/neko/server/pkg/types/event"
)

func TestViewOnlyEventAllowed(t *testing.T) {
	allowed := []string{
		event.CLIENT_HEARTBEAT,
		event.SIGNAL_REQUEST,
		event.SIGNAL_RESTART,
		event.SIGNAL_OFFER,
		event.SIGNAL_ANSWER,
		event.SIGNAL_CANDIDATE,
		event.SIGNAL_VIDEO,
		event.SIGNAL_AUDIO,
	}
	for _, eventName := range allowed {
		if !viewOnlyEventAllowed(eventName) {
			t.Errorf("receive-media event %q was denied", eventName)
		}
	}

	denied := []string{
		event.SYSTEM_LOGS,
		event.CONTROL_REQUEST,
		event.CONTROL_MOVE,
		event.CONTROL_SCROLL,
		event.CONTROL_BUTTONPRESS,
		event.CONTROL_KEYPRESS,
		event.CONTROL_TOUCHBEGIN,
		event.CLIPBOARD_SET,
		event.KEYBOARD_MAP,
		event.SCREEN_SET,
		event.SEND_UNICAST,
		event.SEND_BROADCAST,
		"chat/message",
		"filetransfer/update",
		"plugin/future-interactive-event",
	}
	for _, eventName := range denied {
		if viewOnlyEventAllowed(eventName) {
			t.Errorf("interactive event %q was allowed", eventName)
		}
	}
}
