package webrtc

import (
	"testing"

	"github.com/pion/rtcp"

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

type keyframeSubscription struct {
	types.MediaSubscription
	requests int
}

func (subscription *keyframeSubscription) RequestKeyframe() {
	subscription.requests++
}

func TestRequestsKeyframeMatchesPictureLossForThisStream(t *testing.T) {
	const ssrc = 1234
	tests := []struct {
		name    string
		packets []rtcp.Packet
		want    bool
	}{
		{"pli", []rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: ssrc}}, true},
		{"fir", []rtcp.Packet{&rtcp.FullIntraRequest{FIR: []rtcp.FIREntry{{SSRC: ssrc}}}}, true},
		{"pli in compound report", []rtcp.Packet{&rtcp.ReceiverReport{}, &rtcp.PictureLossIndication{MediaSSRC: ssrc}}, true},
		{"pli for another stream", []rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: ssrc + 1}}, false},
		{"receiver report and nack only", []rtcp.Packet{&rtcp.ReceiverReport{}, &rtcp.TransportLayerNack{MediaSSRC: ssrc}}, false},
	}
	for _, test := range tests {
		if got := requestsKeyframe(test.packets, ssrc); got != test.want {
			t.Errorf("%s: requestsKeyframe() = %v, want %v", test.name, got, test.want)
		}
	}
}

func TestTrackPassesKeyframeRequestToSubscription(t *testing.T) {
	subscription := &keyframeSubscription{}
	track := &Track{kind: types.MediaKindVideo, subscription: subscription}

	track.requestKeyframe()
	if subscription.requests != 1 {
		t.Fatalf("subscription got %d keyframe requests, want 1", subscription.requests)
	}

	// no subscription yet, or one that cannot request keyframes
	(&Track{kind: types.MediaKindVideo}).requestKeyframe()
}
