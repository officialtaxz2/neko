package websocket

import "github.com/m1k1o/neko/server/pkg/types/event"

// viewOnlyEventAllowed is the WebSocket ingress allowlist for passive
// sessions. Signalling can select receive media, but every interactive core or
// plugin event is denied before it reaches a handler.
func viewOnlyEventAllowed(eventName string) bool {
	switch eventName {
	case event.CLIENT_HEARTBEAT,
		event.SIGNAL_REQUEST,
		event.SIGNAL_RESTART,
		event.SIGNAL_OFFER,
		event.SIGNAL_ANSWER,
		event.SIGNAL_CANDIDATE,
		event.SIGNAL_VIDEO,
		event.SIGNAL_AUDIO,
		event.MEDIA_CAPABILITIES_REQUEST,
		event.MEDIA_CREATE:
		return true
	default:
		return false
	}
}
