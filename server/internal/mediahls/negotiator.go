package mediahls

import (
	"slices"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/event"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

type creationBucket struct { tokens float64; last time.Time }

type Negotiator struct {
	logger zerolog.Logger
	sessions types.SessionManager
	tickets *TicketStore
	config Config
	mu sync.Mutex
	buckets map[string]creationBucket
	now func() time.Time
}

func NewNegotiator(sessions types.SessionManager, tickets *TicketStore, config Config) (*Negotiator, error) {
	normalized, err := NormalizeConfig(config)
	if err != nil { return nil, err }
	if sessions == nil { return nil, ErrInvalidConfig }
	if tickets == nil { tickets = NewTicketStore() }
	negotiator := &Negotiator{logger: log.With().Str("module", "mediahls").Str("submodule", "negotiation").Logger(), sessions: sessions, tickets: tickets, config: normalized, buckets: map[string]creationBucket{}, now: time.Now}
	sessions.OnDeleted(func(session types.Session){ negotiator.tickets.InvalidateSession(session.ID()); negotiator.forgetRateState(session.ID()) })
	sessions.OnDisconnected(func(session types.Session){ negotiator.tickets.InvalidateSession(session.ID()) })
	sessions.OnProfileChanged(func(session types.Session, new, _ types.MemberProfile){ if !new.CanWatch { negotiator.tickets.InvalidateSession(session.ID()) } })
	return negotiator, nil
}

func (negotiator *Negotiator) Tickets() *TicketStore { return negotiator.tickets }

func (negotiator *Negotiator) Handler(session types.Session, data types.WebSocketMessage) bool {
	switch data.Event {
	case event.MEDIA_HLS_CAPABILITIES_REQUEST:
		request := message.MediaHLSCapabilitiesRequest{}
		if DecodeStrictJSON(data.Payload, &request) != nil || RequireJSONFields(data.Payload, "version", "mode") != nil {
			negotiator.logReject(session, data.Event, "invalid_payload"); return true
		}
		negotiator.capabilities(session, request); return true
	case event.MEDIA_HLS_CREATE:
		request := message.MediaHLSCreate{}
		if DecodeStrictJSON(data.Payload, &request) != nil || RequireJSONFields(data.Payload, "version", "backend", "mode") != nil {
			negotiator.logReject(session, data.Event, "invalid_payload"); return true
		}
		negotiator.create(session, request); return true
	default:
		return false
	}
}

func (negotiator *Negotiator) capabilities(session types.Session, request message.MediaHLSCapabilitiesRequest) {
	current, ok := negotiator.currentWatchSession(session)
	if !ok || request.Version != int(ProtocolVersion) || !negotiator.modeEnabled(request.Mode) {
		negotiator.logReject(session, event.MEDIA_HLS_CAPABILITIES_REQUEST, "not_allowed"); return
	}
	current.Send(event.MEDIA_HLS_CAPABILITIES, message.MediaHLSCapabilities{
		Version: int(ProtocolVersion), Backend: BackendName, Modes: slices.Clone(negotiator.config.Modes),
		Container: "fmp4", VideoCodec: "h264-high-3.1", AudioCodec: "aac-lc", AudioRate: 48_000,
		Variants: capabilityVariants(), Limits: capabilityLimits(negotiator.config),
	})
}

func (negotiator *Negotiator) create(session types.Session, request message.MediaHLSCreate) {
	current, ok := negotiator.currentWatchSession(session)
	if !ok || request.Version != int(ProtocolVersion) || request.Backend != BackendName || !negotiator.modeEnabled(request.Mode) {
		negotiator.logReject(session, event.MEDIA_HLS_CREATE, "not_allowed"); return
	}
	if !negotiator.allowCreate(current.ID(), negotiator.now()) {
		negotiator.logReject(session, event.MEDIA_HLS_CREATE, "rate_limited"); return
	}
	now := negotiator.now()
	offer, err := negotiator.tickets.Issue(TicketBinding{SessionID: current.ID(), Backend: BackendName, Version: ProtocolVersion, Mode: request.Mode, Request: BootstrapRequest{Mode: request.Mode, VariantIDs: []string{"high", "medium", "low"}, Audio: true}}, now)
	if err != nil { negotiator.logger.Error().Err(err).Str("session_id", current.ID()).Msg("HLS bootstrap ticket creation failed"); return }
	current.Send(event.MEDIA_HLS_OFFER, message.MediaHLSOffer{Version: int(ProtocolVersion), Backend: BackendName, Mode: request.Mode, Path: BootstrapPath, Ticket: offer.Ticket, ExpiresInMS: offer.ExpiresAt.Sub(now).Milliseconds()})
}

func (negotiator *Negotiator) currentWatchSession(session types.Session) (types.Session, bool) {
	if session == nil { return nil, false }
	current, ok := negotiator.sessions.Get(session.ID())
	if !ok || current != session || !current.Profile().CanWatch || !current.State().IsConnected { return nil, false }
	return current, true
}

func (negotiator *Negotiator) modeEnabled(mode string) bool { return slices.Contains(negotiator.config.Modes, mode) }

func (negotiator *Negotiator) allowCreate(sessionID string, now time.Time) bool {
	negotiator.mu.Lock(); defer negotiator.mu.Unlock()
	bucket := negotiator.buckets[sessionID]
	if bucket.last.IsZero() { bucket.tokens, bucket.last = BootstrapCreateBurst, now
	} else if now.After(bucket.last) { bucket.tokens += now.Sub(bucket.last).Minutes()*BootstrapCreatesPerMinute; if bucket.tokens > BootstrapCreateBurst { bucket.tokens = BootstrapCreateBurst }; bucket.last = now }
	if bucket.tokens < 1 { negotiator.buckets[sessionID] = bucket; return false }
	bucket.tokens--; negotiator.buckets[sessionID] = bucket; return true
}

func (negotiator *Negotiator) forgetRateState(sessionID string) { negotiator.mu.Lock(); delete(negotiator.buckets, sessionID); negotiator.mu.Unlock() }

func (negotiator *Negotiator) logReject(session types.Session, eventName, reason string) {
	if session == nil { return }
	negotiator.logger.Warn().Str("session_id", session.ID()).Str("event", eventName).Str("reason", reason).Msg("HLS negotiation rejected")
}
