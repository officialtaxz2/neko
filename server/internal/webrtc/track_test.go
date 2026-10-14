package webrtc

import (
	"testing"

	"github.com/m1k1o/neko/server/pkg/types"
)

func TestWebRTCTrackUsesSingleTwoUnitDropNewestSubscription(t *testing.T) {
	track := &Track{kind: types.MediaKindVideo}
	request := track.subscriptionRequest("high")

	if request.QueueCapacity != 2 {
		t.Fatalf("WebRTC queue capacity = %d, want 2", request.QueueCapacity)
	}
	if request.OverflowPolicy != types.MediaOverflowDropNewest {
		t.Fatalf("WebRTC overflow policy = %q, want drop_newest", request.OverflowPolicy)
	}
	if request.Backend != "webrtc" || request.Kind != types.MediaKindVideo || request.Selector.ID != "high" {
		t.Fatalf("WebRTC subscription request = %#v", request)
	}
	if request.Observer != track {
		t.Fatal("WebRTC track is not observing its provider-owned queue drops")
	}
}
