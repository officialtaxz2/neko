package handler

import (
	"errors"
	"testing"

	"github.com/pion/webrtc/v4"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

// The fakes below embed the real interfaces so we only implement the methods
// signalRequest actually calls.
type fakeSession struct {
	types.Session
	profile types.MemberProfile
	sent    int
}

func (s *fakeSession) ID() string { return "test-session" }

func (s *fakeSession) Profile() types.MemberProfile { return s.profile }

func (s *fakeSession) PrivateModeEnabled() bool { return false }

func (s *fakeSession) Send(string, any) { s.sent++ }

type fakeWebRTC struct {
	types.WebRTCManager
	peer *fakePeer
	// peer returned by the next CreatePeer, a new one if nil
	next *fakePeer
}

func (w *fakeWebRTC) ICEServers(types.Session) []types.ICEServer { return nil }

func (w *fakeWebRTC) CreatePeer(types.Session) (*webrtc.SessionDescription, types.WebRTCPeer, error) {
	w.peer = w.next
	if w.peer == nil {
		w.peer = &fakePeer{}
	}
	return &webrtc.SessionDescription{SDP: "fake-offer"}, w.peer, nil
}

type fakePeer struct {
	types.WebRTCPeer
	video     types.PeerVideoRequest
	videoErr  error
	audioErr  error
	destroyed bool
}

func (p *fakePeer) SetPaused(bool) error { return nil }

func (p *fakePeer) SetVideo(v types.PeerVideoRequest) error {
	p.video = v
	return p.videoErr
}

func (p *fakePeer) Video() types.PeerVideo { return types.PeerVideo{} }

func (p *fakePeer) SetAudio(types.PeerAudioRequest) error { return p.audioErr }

func (p *fakePeer) Destroy() { p.destroyed = true }

func (p *fakePeer) Audio() types.PeerAudio { return types.PeerAudio{} }

type fakeCapture struct {
	types.CaptureManager
	ids []string
}

func (c *fakeCapture) Video() types.StreamSelectorManager {
	return &fakeSelector{ids: c.ids}
}

type fakeSelector struct {
	types.StreamSelectorManager
	ids []string
}

func (s *fakeSelector) IDs() []string { return s.ids }

func newSignalTestHandler(capture types.CaptureManager) (*MessageHandlerCtx, *fakeWebRTC) {
	webrtcManager := &fakeWebRTC{}
	h := &MessageHandlerCtx{
		capture: capture,
		webrtc:  webrtcManager,
	}
	return h, webrtcManager
}

func watchSession() *fakeSession {
	return &fakeSession{profile: types.MemberProfile{CanWatch: true}}
}

// On master, an empty video id list made signalRequest index videos[0] and
// panic the whole server. It must return a descriptive error instead.
func TestSignalRequestNoVideoStreams(t *testing.T) {
	h, _ := newSignalTestHandler(&fakeCapture{ids: nil})

	payload := &message.SignalRequest{}
	if err := h.signalRequest(watchSession(), payload); err == nil {
		t.Fatal("expected an error when no video streams are configured, got nil")
	}
}

// When streams exist and the client sends no selector, the first video id is
// used as the default stream.
func TestSignalRequestDefaultsToFirstVideo(t *testing.T) {
	h, webrtcManager := newSignalTestHandler(&fakeCapture{ids: []string{"high", "low", "medium"}})

	payload := &message.SignalRequest{}
	if err := h.signalRequest(watchSession(), payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	selector := webrtcManager.peer.video.Selector
	if selector == nil {
		t.Fatal("expected a default selector to be set")
	}
	if selector.ID != "high" {
		t.Fatalf("expected default video id %q, got %q", "high", selector.ID)
	}
	if selector.Type != types.StreamSelectorTypeExact {
		t.Fatalf("expected exact selector type, got %v", selector.Type)
	}
}

// Fork: the selection is validated before a peer exists, so an empty one
// never creates a peer that would keep the capture pipelines running.
func TestSignalRequestNoVideoStreamsCreatesNoPeer(t *testing.T) {
	h, webrtcManager := newSignalTestHandler(&fakeCapture{ids: nil})

	if err := h.signalRequest(watchSession(), &message.SignalRequest{}); err == nil {
		t.Fatal("expected an error when no video streams are configured, got nil")
	}
	if webrtcManager.peer != nil {
		t.Fatal("an empty selection created a peer")
	}
}

// Fork: a peer that fails to initialize is destroyed and never offered,
// while a successful one is kept and offered once.
func TestSignalRequestClosesPartialPeerAndRetainsSuccessfulPeer(t *testing.T) {
	for _, stage := range []string{"video", "audio", "success"} {
		t.Run(stage, func(t *testing.T) {
			h, webrtcManager := newSignalTestHandler(&fakeCapture{ids: []string{"main"}})
			peer := &fakePeer{}
			switch stage {
			case "video":
				peer.videoErr = errors.New("invalid video")
			case "audio":
				peer.audioErr = errors.New("invalid audio")
			}
			webrtcManager.next = peer
			session := watchSession()

			err := h.signalRequest(session, &message.SignalRequest{})
			if stage == "success" {
				if err != nil || peer.destroyed || session.sent != 1 {
					t.Fatalf("successful peer not retained: err %v, destroyed %v, sent %d", err, peer.destroyed, session.sent)
				}
				return
			}
			if err == nil || !peer.destroyed || session.sent != 0 {
				t.Fatalf("partial peer leaked at %s: err %v, destroyed %v, sent %d", stage, err, peer.destroyed, session.sent)
			}
		})
	}
}
