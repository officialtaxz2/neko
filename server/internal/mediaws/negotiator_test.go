package mediaws

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/internal/config"
	sessionmanager "github.com/m1k1o/neko/server/internal/session"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/event"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

type negotiationProvider struct {
	audio []types.MediaSource
	video []types.MediaSource
}

func (provider *negotiationProvider) Sources(kind types.MediaKind) []types.MediaSource {
	if kind == types.MediaKindAudio {
		return provider.audio
	}
	if kind == types.MediaKindVideo {
		return provider.video
	}
	return nil
}

func (*negotiationProvider) Subscribe(context.Context, types.SourceSubscriptionRequest) (types.MediaSubscription, error) {
	panic("phase 1 negotiation must not subscribe to capture")
}

type negotiationPeer struct {
	events map[string][]any
}

func newNegotiationPeer() *negotiationPeer {
	return &negotiationPeer{events: make(map[string][]any)}
}

func (peer *negotiationPeer) Send(eventName string, payload any) {
	peer.events[eventName] = append(peer.events[eventName], payload)
}

func (*negotiationPeer) Ping() error     { return nil }
func (*negotiationPeer) Destroy(string) {}

func (peer *negotiationPeer) last(eventName string) (any, bool) {
	items := peer.events[eventName]
	if len(items) == 0 {
		return nil, false
	}
	return items[len(items)-1], true
}

func newNegotiationFixture(t *testing.T, canWatch bool) (*Negotiator, *sessionmanager.SessionManagerCtx, types.Session, *negotiationPeer, time.Time) {
	t.Helper()
	sessions := sessionmanager.New(&config.Session{})
	viewer, _, err := sessions.Create("viewer", types.MemberProfile{
		Name:       "viewer",
		CanLogin:   true,
		CanConnect: true,
		CanWatch:   canWatch,
	})
	if err != nil {
		t.Fatal(err)
	}
	peer := newNegotiationPeer()
	viewer.ConnectWebSocketPeer(peer)
	provider := &negotiationProvider{
		audio: []types.MediaSource{{
			ID:   "audio",
			Kind: types.MediaKindAudio,
			Codec: types.MediaCodec{
				Name:      "opus",
				MIMEType:  "audio/opus",
				ClockRate: 48000,
				Channels:  2,
			},
			NominalBitrate: 128000,
		}},
		video: []types.MediaSource{{
			ID:   "high",
			Kind: types.MediaKindVideo,
			Codec: types.MediaCodec{
				Name:      "VP8",
				MIMEType:  "video/vp8",
				ClockRate: 90000,
			},
			// A cold capture source does not know dimensions until Phase 2 opens
			// its bounded provider subscription and receives initial FORMAT.
			NominalBitrate: 1_996_800,
		}},
	}
	now := time.Unix(1_700_000_000, 0)
	negotiator := NewNegotiator(sessions, provider, nil)
	negotiator.now = func() time.Time { return now }
	return negotiator, sessions, viewer, peer, now
}

func TestNegotiatorAdvertisesColdVP8AndOpusSources(t *testing.T) {
	negotiator, _, viewer, peer, _ := newNegotiationFixture(t, true)
	handled := negotiator.Handler(viewer, types.WebSocketMessage{
		Event:   event.MEDIA_CAPABILITIES_REQUEST,
		Payload: json.RawMessage(`{"version":1}`),
	})
	if !handled {
		t.Fatal("capability event was not handled")
	}
	raw, ok := peer.last(event.MEDIA_CAPABILITIES)
	if !ok {
		t.Fatal("capabilities were not sent")
	}
	capabilities, ok := raw.(message.MediaCapabilities)
	if !ok {
		t.Fatalf("capabilities type = %T", raw)
	}
	if capabilities.Backend != BackendName || capabilities.Protocol != ProtocolName || len(capabilities.Audio) != 1 || len(capabilities.Video) != 1 {
		t.Fatalf("capabilities = %#v", capabilities)
	}
	if capabilities.Video[0].MIMEType != "video/VP8" || capabilities.Video[0].CodedWidth != 0 || capabilities.Video[0].CodedHeight != 0 {
		t.Fatalf("cold video capability = %#v", capabilities.Video[0])
	}
	if capabilities.Audio[0].MIMEType != "audio/opus" || capabilities.Audio[0].Channels != 2 {
		t.Fatalf("audio capability = %#v", capabilities.Audio[0])
	}
}

func TestNegotiatorCreatesBoundTicketAndRevokesItWithCanWatch(t *testing.T) {
	negotiator, sessions, viewer, peer, now := newNegotiationFixture(t, true)
	request := json.RawMessage(`{"version":1,"backend":"webcodecs-ws","audio":{"source_id":"audio","codec":"opus"},"video":{"source_id":"high","codec":"vp8"}}`)
	if !negotiator.Handler(viewer, types.WebSocketMessage{Event: event.MEDIA_CREATE, Payload: request}) {
		t.Fatal("create event was not handled")
	}
	raw, ok := peer.last(event.MEDIA_OFFER)
	if !ok {
		t.Fatal("offer was not sent")
	}
	offer, ok := raw.(message.MediaOffer)
	if !ok {
		t.Fatalf("offer type = %T", raw)
	}
	if len(offer.Ticket) != TicketEncodedLength || offer.ExpiresInMS != TicketLifetime.Milliseconds() || offer.Path != MediaPath {
		t.Fatalf("offer = %#v", offer)
	}
	binding, err := negotiator.Tickets().Redeem(offer.Ticket, now)
	if err != nil {
		t.Fatal(err)
	}
	if binding.SessionID != viewer.ID() || binding.Request.Video.SourceID != "high" || binding.Request.Audio.SourceID != "audio" {
		t.Fatalf("binding = %#v", binding)
	}

	if !negotiator.Handler(viewer, types.WebSocketMessage{Event: event.MEDIA_CREATE, Payload: request}) {
		t.Fatal("second create event was not handled")
	}
	raw, ok = peer.last(event.MEDIA_OFFER)
	if !ok {
		t.Fatal("second offer was not sent")
	}
	offer = raw.(message.MediaOffer)
	profile := viewer.Profile()
	profile.CanWatch = false
	if err := sessions.Update(viewer.ID(), profile); err != nil {
		t.Fatal(err)
	}
	if _, err := negotiator.Tickets().Redeem(offer.Ticket, now); !errors.Is(err, ErrTicketGone) {
		t.Fatalf("revoked ticket error = %v", err)
	}
}

func TestNegotiatorRequiresCanWatchAndRateLimitsTicketCreation(t *testing.T) {
	denied, _, deniedViewer, deniedPeer, _ := newNegotiationFixture(t, false)
	denied.Handler(deniedViewer, types.WebSocketMessage{
		Event:   event.MEDIA_CAPABILITIES_REQUEST,
		Payload: json.RawMessage(`{"version":1}`),
	})
	if _, ok := deniedPeer.last(event.MEDIA_CAPABILITIES); ok {
		t.Fatal("non-watcher received capabilities")
	}

	negotiator, _, viewer, peer, _ := newNegotiationFixture(t, true)
	request := json.RawMessage(`{"version":1,"backend":"webcodecs-ws","audio":null,"video":{"source_id":"high","codec":"vp8"}}`)
	for attempt := 0; attempt < TicketCreateBurst+1; attempt++ {
		negotiator.Handler(viewer, types.WebSocketMessage{Event: event.MEDIA_CREATE, Payload: request})
	}
	if offers := len(peer.events[event.MEDIA_OFFER]); offers != TicketCreateBurst {
		t.Fatalf("offers = %d, want %d", offers, TicketCreateBurst)
	}
}

func TestNegotiatorRequiresExplicitAudioChoice(t *testing.T) {
	negotiator, _, viewer, peer, _ := newNegotiationFixture(t, true)
	negotiator.Handler(viewer, types.WebSocketMessage{
		Event:   event.MEDIA_CREATE,
		Payload: json.RawMessage(`{"version":1,"backend":"webcodecs-ws","video":{"source_id":"high","codec":"vp8"}}`),
	})
	if _, ok := peer.last(event.MEDIA_OFFER); ok {
		t.Fatal("create request with omitted audio field received an offer")
	}
}

func TestSupportedSourceRejectsConfigAndWrongClock(t *testing.T) {
	source := types.MediaSource{
		ID:   "high",
		Kind: types.MediaKindVideo,
		Codec: types.MediaCodec{Name: "vp8", MIMEType: "video/VP8", ClockRate: 90000},
	}
	if !supportedSource(source) {
		t.Fatal("valid cold VP8 source rejected")
	}
	source.Codec.Config = []byte{1}
	if supportedSource(source) {
		t.Fatal("VP8 source with decoder config accepted")
	}
	source.Codec.Config = nil
	source.Codec.ClockRate = 80000
	if supportedSource(source) {
		t.Fatal("VP8 source with wrong clock accepted")
	}

	source.Codec.ClockRate = 90000
	source.Width = MaxVideoDimension + 1
	if supportedSource(source) {
		t.Fatal("VP8 source with over-limit dimensions accepted")
	}
	source.Width = 1280
	if supportedSource(source) {
		t.Fatal("VP8 source with only one known dimension accepted")
	}
	source.Height = 720
	source.FrameRateNumerator = 30
	if supportedSource(source) {
		t.Fatal("VP8 source with incomplete frame rate accepted")
	}
}
