package mediaws

import (
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/event"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

const (
	BackendName = "webcodecs-ws"
	MediaPath   = "/api/media/ws"

	TicketCreatesPerMinute = 6
	TicketCreateBurst      = 2
)

type creationBucket struct {
	tokens float64
	last   time.Time
}

type Negotiator struct {
	logger   zerolog.Logger
	sessions types.SessionManager
	provider types.EncodedMediaProvider
	tickets  *TicketStore

	mu      sync.Mutex
	buckets map[string]creationBucket
	now     func() time.Time
}

func NewNegotiator(sessions types.SessionManager, provider types.EncodedMediaProvider, tickets *TicketStore) *Negotiator {
	if tickets == nil {
		tickets = NewTicketStore()
	}
	negotiator := &Negotiator{
		logger:   log.With().Str("module", "mediaws").Str("submodule", "negotiation").Logger(),
		sessions: sessions,
		provider: provider,
		tickets:  tickets,
		buckets:  make(map[string]creationBucket),
		now:      time.Now,
	}

	sessions.OnDeleted(func(session types.Session) {
		negotiator.tickets.InvalidateSession(session.ID())
		negotiator.forgetRateState(session.ID())
	})
	sessions.OnDisconnected(func(session types.Session) {
		// The session manager emits this only after its existing reconnect grace
		// period has elapsed or an explicit disconnect happens.
		negotiator.tickets.InvalidateSession(session.ID())
	})
	sessions.OnProfileChanged(func(session types.Session, new, _ types.MemberProfile) {
		if !new.CanWatch {
			negotiator.tickets.InvalidateSession(session.ID())
		}
	})
	return negotiator
}

func (negotiator *Negotiator) Tickets() *TicketStore {
	return negotiator.tickets
}

func (negotiator *Negotiator) Handler(session types.Session, data types.WebSocketMessage) bool {
	switch data.Event {
	case event.MEDIA_CAPABILITIES_REQUEST:
		request := message.MediaCapabilitiesRequest{}
		if err := DecodeStrictJSON(data.Payload, &request); err != nil {
			negotiator.logReject(session, data.Event, "invalid_payload")
			return true
		}
		if err := RequireJSONFields(data.Payload, "version"); err != nil {
			negotiator.logReject(session, data.Event, "invalid_payload")
			return true
		}
		negotiator.capabilities(session, request)
		return true

	case event.MEDIA_CREATE:
		request := message.MediaCreate{}
		if err := DecodeStrictJSON(data.Payload, &request); err != nil {
			negotiator.logReject(session, data.Event, "invalid_payload")
			return true
		}
		if err := RequireJSONFields(data.Payload, "version", "backend", "audio", "video"); err != nil {
			negotiator.logReject(session, data.Event, "invalid_payload")
			return true
		}
		negotiator.create(session, request)
		return true
	default:
		return false
	}
}

func (negotiator *Negotiator) capabilities(session types.Session, request message.MediaCapabilitiesRequest) {
	current, ok := negotiator.currentWatchSession(session)
	if !ok || request.Version != int(ProtocolVersion) {
		negotiator.logReject(session, event.MEDIA_CAPABILITIES_REQUEST, "not_allowed")
		return
	}

	current.Send(event.MEDIA_CAPABILITIES, message.MediaCapabilities{
		Version:  int(ProtocolVersion),
		Backend:  BackendName,
		Protocol: ProtocolName,
		Audio:    negotiator.capabilitySources(types.MediaKindAudio),
		Video:    negotiator.capabilitySources(types.MediaKindVideo),
	})
}

func (negotiator *Negotiator) create(session types.Session, request message.MediaCreate) {
	current, ok := negotiator.currentWatchSession(session)
	if !ok || request.Version != int(ProtocolVersion) || request.Backend != BackendName || request.Video == nil {
		negotiator.logReject(session, event.MEDIA_CREATE, "not_allowed")
		return
	}
	if !negotiator.allowCreate(current.ID(), negotiator.now()) {
		negotiator.logReject(session, event.MEDIA_CREATE, "rate_limited")
		return
	}

	video, ok := negotiator.normalizeChoice(types.MediaKindVideo, request.Video)
	if !ok {
		negotiator.logReject(session, event.MEDIA_CREATE, "unsupported_video")
		return
	}
	audio := NegotiatedKind{}
	if request.Audio != nil {
		audio, ok = negotiator.normalizeChoice(types.MediaKindAudio, request.Audio)
		if !ok {
			negotiator.logReject(session, event.MEDIA_CREATE, "unsupported_audio")
			return
		}
	}

	now := negotiator.now()
	offer, err := negotiator.tickets.Issue(TicketBinding{
		SessionID: current.ID(),
		Backend:   BackendName,
		Version:   ProtocolVersion,
		Request: NegotiatedRequest{
			Audio: audio,
			Video: video,
		},
	}, now)
	if err != nil {
		negotiator.logger.Error().Err(err).Str("session_id", current.ID()).Msg("media ticket creation failed")
		return
	}

	current.Send(event.MEDIA_OFFER, message.MediaOffer{
		Version:     int(ProtocolVersion),
		Backend:     BackendName,
		Protocol:    ProtocolName,
		Path:        MediaPath,
		Ticket:      offer.Ticket,
		ExpiresInMS: offer.ExpiresAt.Sub(now).Milliseconds(),
	})
}

func (negotiator *Negotiator) currentWatchSession(session types.Session) (types.Session, bool) {
	if session == nil {
		return nil, false
	}
	current, ok := negotiator.sessions.Get(session.ID())
	if !ok || current != session || !current.Profile().CanWatch || !current.State().IsConnected {
		return nil, false
	}
	return current, true
}

func (negotiator *Negotiator) capabilitySources(kind types.MediaKind) []message.MediaCapabilitySource {
	sources := negotiator.provider.Sources(kind)
	result := make([]message.MediaCapabilitySource, 0, len(sources))
	for _, source := range sources {
		if !supportedSource(source) {
			continue
		}
		result = append(result, message.MediaCapabilitySource{
			ID:                   source.ID,
			Codec:                normalizedCodec(source),
			MIMEType:             canonicalMIMEType(source.Kind),
			ClockRate:            source.Codec.ClockRate,
			Channels:             source.Codec.Channels,
			CodedWidth:           source.Width,
			CodedHeight:          source.Height,
			FrameRateNumerator:   source.FrameRateNumerator,
			FrameRateDenominator: source.FrameRateDenominator,
			NominalBitrate:       source.NominalBitrate,
			Selector: types.MediaSelector{
				Type: types.MediaSelectorTypeExact,
				ID:   source.ID,
			},
		})
	}
	return result
}

func (negotiator *Negotiator) normalizeChoice(kind types.MediaKind, choice *message.MediaCreateChoice) (NegotiatedKind, bool) {
	if choice == nil || choice.SourceID == "" || len(choice.SourceID) > MaxSourceIDLength || choice.Codec == "" {
		return NegotiatedKind{}, false
	}
	for _, source := range negotiator.provider.Sources(kind) {
		if source.ID != choice.SourceID || !supportedSource(source) {
			continue
		}
		codec := normalizedCodec(source)
		if choice.Codec != codec {
			return NegotiatedKind{}, false
		}
		return NegotiatedKind{Enabled: true, SourceID: source.ID, Codec: codec}, true
	}
	return NegotiatedKind{}, false
}

func supportedSource(source types.MediaSource) bool {
	if source.ID == "" || len(source.ID) > MaxSourceIDLength || len(source.Codec.Config) != 0 || source.NominalBitrate > MaxSafeInteger {
		return false
	}
	codec := normalizedCodec(source)
	switch source.Kind {
	case types.MediaKindVideo:
		dimensionsValid := (source.Width == 0 && source.Height == 0) ||
			(source.Width > 0 && source.Width <= MaxVideoDimension && source.Height > 0 && source.Height <= MaxVideoDimension)
		frameRateValid := (source.FrameRateNumerator == 0 && source.FrameRateDenominator == 0) ||
			(source.FrameRateNumerator > 0 && source.FrameRateDenominator > 0)
		return codec == "vp8" && strings.EqualFold(source.Codec.MIMEType, "video/VP8") && source.Codec.ClockRate == 90000 && source.Codec.Channels == 0 && dimensionsValid && frameRateValid
	case types.MediaKindAudio:
		return codec == "opus" && strings.EqualFold(source.Codec.MIMEType, "audio/opus") && source.Codec.ClockRate == 48000 && source.Codec.Channels == 2 &&
			source.Width == 0 && source.Height == 0 && source.FrameRateNumerator == 0 && source.FrameRateDenominator == 0
	default:
		return false
	}
}

func normalizedCodec(source types.MediaSource) string {
	return strings.ToLower(strings.TrimSpace(source.Codec.Name))
}

func canonicalMIMEType(kind types.MediaKind) string {
	if kind == types.MediaKindVideo {
		return "video/VP8"
	}
	return "audio/opus"
}

func (negotiator *Negotiator) allowCreate(sessionID string, now time.Time) bool {
	negotiator.mu.Lock()
	defer negotiator.mu.Unlock()

	bucket := negotiator.buckets[sessionID]
	if bucket.last.IsZero() {
		bucket.tokens = TicketCreateBurst
		bucket.last = now
	} else if now.After(bucket.last) {
		elapsed := now.Sub(bucket.last).Minutes()
		bucket.tokens += elapsed * TicketCreatesPerMinute
		if bucket.tokens > TicketCreateBurst {
			bucket.tokens = TicketCreateBurst
		}
		bucket.last = now
	}
	if bucket.tokens < 1 {
		negotiator.buckets[sessionID] = bucket
		return false
	}
	bucket.tokens--
	negotiator.buckets[sessionID] = bucket
	return true
}

func (negotiator *Negotiator) forgetRateState(sessionID string) {
	negotiator.mu.Lock()
	delete(negotiator.buckets, sessionID)
	negotiator.mu.Unlock()
}

func (negotiator *Negotiator) logReject(session types.Session, eventName, reason string) {
	if session == nil {
		return
	}
	negotiator.logger.Warn().
		Str("session_id", session.ID()).
		Str("event", eventName).
		Str("reason", reason).
		Msg("media negotiation rejected")
}
