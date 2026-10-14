package websocket

import (
	"slices"
	"testing"

	"github.com/m1k1o/neko/server/pkg/types/event"
)

func TestMediaNegotiationPayloadsAreNotLogged(t *testing.T) {
	for _, eventName := range []string{
		event.MEDIA_CAPABILITIES_REQUEST,
		event.MEDIA_CAPABILITIES,
		event.MEDIA_CREATE,
		event.MEDIA_OFFER,
		event.MEDIA_HLS_CAPABILITIES_REQUEST,
		event.MEDIA_HLS_CAPABILITIES,
		event.MEDIA_HLS_CREATE,
		event.MEDIA_HLS_OFFER,
	} {
		if !slices.Contains(nologEvents, eventName) {
			t.Errorf("media negotiation event %q is payload-logged", eventName)
		}
	}
}
