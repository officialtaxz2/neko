package handler

import (
	"errors"
	"testing"

	"github.com/pion/webrtc/v4"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

type signalSession struct { types.Session; sent int }
func (*signalSession) Profile() types.MemberProfile { return types.MemberProfile{CanWatch: true} }
func (*signalSession) PrivateModeEnabled() bool { return false }
func (session *signalSession) Send(string, any) { session.sent++ }
type signalSelector struct { types.StreamSelectorManager; ids []string }
func (selector signalSelector) IDs() []string { return selector.ids }
type signalCapture struct { types.CaptureManager; ids []string }
func (capture signalCapture) Video() types.StreamSelectorManager { return signalSelector{ids: capture.ids} }
type signalPeer struct { types.WebRTCPeer; videoErr, audioErr error; destroyed bool }
func (peer *signalPeer) SetVideo(types.PeerVideoRequest) error { return peer.videoErr }
func (peer *signalPeer) SetAudio(types.PeerAudioRequest) error { return peer.audioErr }
func (peer *signalPeer) Video() types.PeerVideo { return types.PeerVideo{} }
func (peer *signalPeer) Audio() types.PeerAudio { return types.PeerAudio{} }
func (peer *signalPeer) Destroy() { peer.destroyed = true }
type signalRTC struct { types.WebRTCManager; peer *signalPeer; opened bool }
func (rtc *signalRTC) CreatePeer(types.Session) (*webrtc.SessionDescription, types.WebRTCPeer, error) {
	rtc.opened = true
	return &webrtc.SessionDescription{}, rtc.peer, nil
}
func (*signalRTC) ICEServers() []types.ICEServer { return nil }

func TestSignalRequestRejectsEmptySelectionBeforeCreatingPeer(t *testing.T) {
	rtc := &signalRTC{}
	handler := &MessageHandlerCtx{capture: signalCapture{}, webrtc: rtc}
	if err := handler.signalRequest(&signalSession{}, &message.SignalRequest{}); err == nil { t.Fatal("empty video IDs accepted") }
	if rtc.opened { t.Fatal("empty selection created a peer") }
}

func TestSignalRequestClosesPartialPeerAndRetainsSuccessfulPeer(t *testing.T) {
	for _, stage := range []string{"video", "audio", "success"} {
		t.Run(stage, func(t *testing.T) {
			peer := &signalPeer{}
			if stage == "video" { peer.videoErr = errors.New("invalid video") }
			if stage == "audio" { peer.audioErr = errors.New("invalid audio") }
			session := &signalSession{}
			handler := &MessageHandlerCtx{capture: signalCapture{ids: []string{"main"}}, webrtc: &signalRTC{peer: peer}}
			err := handler.signalRequest(session, &message.SignalRequest{})
			if stage == "success" {
				if err != nil || peer.destroyed || session.sent != 1 { t.Fatalf("successful peer not retained: %v", err) }
			} else if err == nil || !peer.destroyed || session.sent != 0 { t.Fatalf("partial peer leaked at %s", stage) }
		})
	}
}
