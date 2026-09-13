package mediaws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/pkg/types"
)

const (
	ReadyTimeout               = 5 * time.Second
	FeedbackTimeout            = 5 * time.Second
	RenderedProgressTimeout    = 3 * time.Second
	ProgressRecoveryTimeout    = 2 * time.Second
	PingPeriod                 = 10 * time.Second
	PongTimeout                = 20 * time.Second
	WriteTimeout               = 2 * time.Second
	CloseTimeout               = time.Second
	ControlRecordsPerSecond    = 10
	ControlRecordBurst         = 20
	ClientResyncsPerMinute     = 5
	MaximumResyncsPerWindow    = 3
	ResyncWindow               = 30 * time.Second
	OutboundRateWindow         = 5 * time.Second
	MaximumOutboundBytesPerSec = 16 * 1024 * 1024
	SentTimestampHistory       = 512
)

var (
	ErrMissingSocketAttachment = errors.New("media websocket attachment missing")
	ErrBackendClosed           = errors.New("media websocket delivery closed")
	ErrCodecUnsupported        = errors.New("media websocket codec or format unsupported")
)

type socketAttachmentContextKey struct{}

type socketAttachment struct {
	connection *websocket.Conn
}

type Backend struct {
	provider types.EncodedMediaProvider
}

func NewBackend(provider types.EncodedMediaProvider) *Backend {
	return &Backend{provider: provider}
}

func (*Backend) Name() string {
	return BackendName
}

func (*Backend) Capabilities() types.MediaBackendCapabilities {
	return types.MediaBackendCapabilities{
		ReceiveAudio: true,
		ReceiveVideo: true,
	}
}

func (backend *Backend) Open(ctx context.Context, lease types.MediaLease, request types.MediaDeliveryRequest) (types.MediaDelivery, error) {
	if ctx == nil || lease == nil || backend.provider == nil || request.Backend != BackendName || !request.Video {
		return nil, types.ErrMediaDeliveryNotAllowed
	}
	attachment, ok := ctx.Value(socketAttachmentContextKey{}).(*socketAttachment)
	if !ok || attachment == nil || attachment.connection == nil {
		return nil, ErrMissingSocketAttachment
	}
	if !request.Audio && (request.AudioSelector != (types.MediaSelector{})) {
		return nil, types.ErrMediaDeliveryNotAllowed
	}
	if request.VideoSelector.Type != types.MediaSelectorTypeExact || request.VideoSelector.ID == "" {
		return nil, types.ErrMediaSourceNotFound
	}
	if request.Audio && (request.AudioSelector.Type != types.MediaSelectorTypeExact || request.AudioSelector.ID == "") {
		return nil, types.ErrMediaSourceNotFound
	}

	delivery := newDelivery(lease, attachment.connection)
	video, err := backend.provider.Subscribe(delivery.ctx, types.SourceSubscriptionRequest{
		Kind:           types.MediaKindVideo,
		Selector:       request.VideoSelector,
		QueueCapacity:  ProviderVideoQueueCapacity,
		OverflowPolicy: types.MediaOverflowDropNewest,
		Backend:        BackendName,
		Observer:       delivery,
	})
	if err != nil {
		delivery.cancel()
		return nil, fmt.Errorf("subscribe media websocket video: %w", err)
	}
	if !supportedSource(video.Source()) {
		_ = video.Close()
		delivery.cancel()
		return nil, ErrCodecUnsupported
	}
	delivery.addTrack(KindVideo, video)

	if request.Audio {
		audio, err := backend.provider.Subscribe(delivery.ctx, types.SourceSubscriptionRequest{
			Kind:           types.MediaKindAudio,
			Selector:       request.AudioSelector,
			QueueCapacity:  ProviderAudioQueueCapacity,
			OverflowPolicy: types.MediaOverflowDropNewest,
			Backend:        BackendName,
			Observer:       delivery,
		})
		if err != nil {
			_ = video.Close()
			delivery.cancel()
			return nil, fmt.Errorf("subscribe media websocket audio: %w", err)
		}
		if !supportedSource(audio.Source()) {
			_ = audio.Close()
			_ = video.Close()
			delivery.cancel()
			return nil, ErrCodecUnsupported
		}
		delivery.addTrack(KindAudio, audio)
	}
	if request.InitialPaused {
		delivery.mu.Lock()
		delivery.paused = true
		delivery.mu.Unlock()
		for _, track := range delivery.tracks {
			if err := track.subscription.SetPaused(true); err != nil {
				for _, cleanup := range delivery.tracks {
					_ = cleanup.subscription.Close()
				}
				delivery.cancel()
				return nil, fmt.Errorf("pause initial media websocket subscription: %w", err)
			}
		}
	}

	return delivery, nil
}

type deliveryTrack struct {
	subscription              types.MediaSubscription
	source                    types.MediaSource
	generation                uint64
	nextSequence              uint64
	formatSent                bool
	awaitKeyframe             bool
	transitioning             bool
	pendingReason             string
	suppressProviderTransition bool
	lastRendered              uint64
	lastProviderGen           uint64
	sentUnits                 uint64
	sentTimeline              []sentTimestamp
}

type sentTimestamp struct {
	sequence uint64
	pts      uint64
}

type resyncRequest struct {
	kind   Kind
	reason string
}

type closeRequest struct {
	endReason string
	code      int
	reason    string
	failed    bool
	skipEnd   bool
}

type outboundSample struct {
	at    time.Time
	bytes int
}

type Delivery struct {
	logger     zerolog.Logger
	lease      types.MediaLease
	connection *websocket.Conn
	ctx        context.Context
	cancel     context.CancelFunc
	queue      *deliveryQueue
	done       chan struct{}
	doneOnce   sync.Once
	closeOnce  sync.Once
	startMu    sync.Mutex
	started    bool
	aborted    bool
	workers    sync.WaitGroup
	closing    chan closeRequest
	resync     chan resyncRequest

	mu                       sync.Mutex
	tracks                   map[Kind]*deliveryTrack
	ready                    bool
	paused                   bool
	createdAt                time.Time
	readyAt                  time.Time
	lastFeedbackAt           time.Time
	lastProgressAt           time.Time
	progressRecoveryDeadline time.Time
	sentSinceProgress        bool
	skewExceededAt           time.Time
	controlTokens            float64
	controlRefillAt          time.Time
	clientResyncs            []time.Time
	resyncs                  []time.Time
	pendingCommonResync      string
	requestedClose           *closeRequest
	connectionState          string
	outbound                 []outboundSample
}

func newDelivery(lease types.MediaLease, connection *websocket.Conn) *Delivery {
	ctx, cancel := context.WithCancel(context.Background())
	now := time.Now()
	return &Delivery{
		logger:          log.With().Str("module", "mediaws").Str("submodule", "delivery").Str("session_id", lease.SessionID()).Logger(),
		lease:           lease,
		connection:      connection,
		ctx:             ctx,
		cancel:          cancel,
		queue:           newDeliveryQueue(),
		done:            make(chan struct{}),
		closing:         make(chan closeRequest, 1),
		resync:          make(chan resyncRequest, 8),
		tracks:          make(map[Kind]*deliveryTrack),
		createdAt:       now,
		controlTokens:   ControlRecordBurst,
		controlRefillAt: now,
	}
}

func (delivery *Delivery) addTrack(kind Kind, subscription types.MediaSubscription) {
	delivery.mu.Lock()
	delivery.tracks[kind] = &deliveryTrack{
		subscription:  subscription,
		source:        subscription.Source(),
		generation:    1,
		awaitKeyframe: kind == KindVideo,
	}
	delivery.mu.Unlock()
}

func (delivery *Delivery) ID() string            { return delivery.lease.ID() }
func (delivery *Delivery) SessionID() string     { return delivery.lease.SessionID() }
func (delivery *Delivery) Backend() string       { return BackendName }
func (delivery *Delivery) Done() <-chan struct{} { return delivery.done }
func (delivery *Delivery) Start() {
	delivery.startMu.Lock()
	if delivery.started || delivery.aborted {
		delivery.startMu.Unlock()
		return
	}
	delivery.started = true
	delivery.startMu.Unlock()
	go delivery.run()
}
func (delivery *Delivery) Close() error {
	delivery.close(normalClose())
	return nil
}
func (delivery *Delivery) CloseWithReason(reason types.MediaDeliveryCloseReason) error {
	delivery.close(closeRequestForReason(reason))
	return nil
}

func closeRequestForReason(reason types.MediaDeliveryCloseReason) closeRequest {
	switch reason {
	case types.MediaDeliveryCloseRevoked:
		return closeRequest{endReason: "revoked", code: 4403, reason: "authorization_revoked"}
	case types.MediaDeliveryCloseReplaced:
		return closeRequest{endReason: "replaced", code: 4408, reason: "delivery_replaced"}
	case types.MediaDeliveryCloseShutdown:
		return closeRequest{endReason: "shutdown", code: websocket.CloseGoingAway, reason: "server_shutdown"}
	default:
		return normalClose()
	}
}

func normalClose() closeRequest {
	return closeRequest{endReason: "normal", code: websocket.CloseNormalClosure, reason: "normal"}
}

func (delivery *Delivery) invalidLeaseClose() closeRequest {
	if reasoner, ok := delivery.lease.(types.MediaLeaseCloseReason); ok {
		if reason := reasoner.CloseReason(); reason != "" {
			return closeRequestForReason(reason)
		}
	}
	return attachmentClose("lease_invalid")
}

func (delivery *Delivery) close(request closeRequest) {
	delivery.closeOnce.Do(func() {
		delivery.mu.Lock()
		delivery.requestedClose = &request
		delivery.mu.Unlock()
		delivery.cancel()
		if request.skipEnd {
			delivery.queue.stop()
		} else {
			end := mustLifecycleRecord(Record{
				Type: RecordEnd,
				Kind: KindNone,
				Metadata: mustJSON(EndMetadata{Schema: "neko.media.end/1", Reason: request.endReason}),
			})
			delivery.queue.finish(end)
		}
		delivery.setConnectionState("closing")
		delivery.closing <- request
		delivery.startMu.Lock()
		abort := !delivery.started
		if abort {
			delivery.aborted = true
		}
		delivery.startMu.Unlock()
		if abort {
			delivery.writePreStartClose(request)
			delivery.finish(request)
		}
	})
}

func (delivery *Delivery) writePreStartClose(request closeRequest) {
	if delivery.connection == nil {
		return
	}
	for {
		record, ok := delivery.queue.pop()
		if !ok {
			break
		}
		if err := delivery.writeRecord(record); err != nil {
			return
		}
	}
	deadline := time.Now().Add(CloseTimeout)
	_ = delivery.connection.SetWriteDeadline(deadline)
	_ = delivery.connection.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(request.code, request.reason), deadline)
}

func (delivery *Delivery) run() {
	delivery.mu.Lock()
	stateUnset := delivery.connectionState == ""
	delivery.mu.Unlock()
	if stateUnset {
		delivery.setConnectionState("opening")
	}
	delivery.connection.SetReadLimit(MaxControlRecord)
	_ = delivery.connection.SetReadDeadline(time.Now().Add(PongTimeout))
	delivery.connection.SetPongHandler(func(string) error {
		return delivery.connection.SetReadDeadline(time.Now().Add(PongTimeout))
	})
	delivery.connection.SetPingHandler(func(string) error {
		// Browser clients do not originate protocol pings. Keeping this handler
		// write-free preserves the delivery's single socket-writer boundary.
		return nil
	})
	delivery.connection.SetCloseHandler(func(code int, text string) error {
		// The reader reports the peer close; the writer owns the close reply.
		return &websocket.CloseError{Code: code, Text: text}
	})

	delivery.mu.Lock()
	tracks := make([]*deliveryTrack, 0, len(delivery.tracks))
	for _, track := range delivery.tracks {
		tracks = append(tracks, track)
	}
	delivery.mu.Unlock()
	for _, track := range tracks {
		delivery.startWorker(func() { delivery.pump(track) })
	}
	delivery.startWorker(delivery.readControls)
	delivery.startWorker(delivery.monitor)
	delivery.startWorker(delivery.handleResyncs)

	result := delivery.writeLoop()
	delivery.cancel()
	for _, track := range tracks {
		_ = track.subscription.Close()
	}
	if delivery.connection != nil {
		_ = delivery.connection.Close()
	}
	delivery.workers.Wait()
	delivery.finish(result)
}

func (delivery *Delivery) finish(result closeRequest) {
	delivery.doneOnce.Do(func() {
		delivery.cancel()
		delivery.mu.Lock()
		tracks := make([]*deliveryTrack, 0, len(delivery.tracks))
		for _, track := range delivery.tracks {
			tracks = append(tracks, track)
		}
		delivery.mu.Unlock()
		for _, track := range tracks {
			_ = track.subscription.Close()
		}
		if delivery.connection != nil {
			_ = delivery.connection.Close()
		}
		if result.failed {
			delivery.lease.SetState(types.MediaDeliveryStateFailed)
		} else {
			delivery.lease.SetState(types.MediaDeliveryStateClosed)
		}
		delivery.clearConnectionState()
		close(delivery.done)
	})
}

func (delivery *Delivery) startWorker(worker func()) {
	delivery.workers.Add(1)
	go func() {
		defer delivery.workers.Done()
		worker()
	}()
}

func (delivery *Delivery) setConnectionState(state string) {
	delivery.mu.Lock()
	old := delivery.connectionState
	if old == state || (old == "closing" && state != "closing") {
		delivery.mu.Unlock()
		return
	}
	delivery.connectionState = state
	delivery.mu.Unlock()
	if old != "" {
		mediaWebSocketConnections.WithLabelValues(old).Dec()
	}
	mediaWebSocketConnections.WithLabelValues(state).Inc()
}

func (delivery *Delivery) clearConnectionState() {
	delivery.mu.Lock()
	old := delivery.connectionState
	delivery.connectionState = ""
	delivery.mu.Unlock()
	if old != "" {
		mediaWebSocketConnections.WithLabelValues(old).Dec()
	}
}

func (delivery *Delivery) SetPaused(paused bool) error {
	delivery.mu.Lock()
	if delivery.ctx.Err() != nil {
		delivery.mu.Unlock()
		return ErrBackendClosed
	}
	if delivery.paused == paused {
		delivery.mu.Unlock()
		return nil
	}
	delivery.paused = paused
	if paused {
		delivery.sentSinceProgress = false
		delivery.progressRecoveryDeadline = time.Time{}
		delivery.skewExceededAt = time.Time{}
	} else {
		now := time.Now()
		if delivery.ready {
			delivery.readyAt = now
		} else {
			delivery.createdAt = now
		}
		delivery.lastFeedbackAt = time.Time{}
		delivery.lastProgressAt = now
		delivery.progressRecoveryDeadline = time.Time{}
		delivery.sentSinceProgress = false
		delivery.skewExceededAt = time.Time{}
	}
	tracks := make([]*deliveryTrack, 0, len(delivery.tracks))
	for _, track := range delivery.tracks {
		if !paused {
			track.suppressProviderTransition = true
			track.transitioning = track.transitioning || track.formatSent
		}
		tracks = append(tracks, track)
	}
	delivery.mu.Unlock()
	delivery.queue.clearMedia(KindNone)

	for _, track := range tracks {
		if err := track.subscription.SetPaused(paused); err != nil {
			delivery.close(backendClose("pause_failed"))
			return err
		}
	}
	if paused {
		return nil
	}

	delivery.mu.Lock()
	formatsPublished := false
	for _, track := range tracks {
		track.source = track.subscription.Source()
		track.pendingReason = ""
		formatsPublished = formatsPublished || track.formatSent
	}
	delivery.paused = false
	delivery.mu.Unlock()
	if formatsPublished {
		delivery.requestResync(KindNone, "resumed")
	}
	return nil
}

func (delivery *Delivery) OnMediaSubscriptionDrop(drop types.MediaDrop) {
	kind := kindFromMedia(drop.Kind)
	if kind == KindNone {
		return
	}
	mediaWebSocketDrops.WithLabelValues(metricKind(kind), "provider", "queue_full").Inc()
	if kind == KindAudio {
		kind = KindNone
	}
	delivery.requestResync(kind, "server_overflow")
}

func (delivery *Delivery) requestResync(kind Kind, reason string) {
	delivery.mu.Lock()
	if delivery.ctx.Err() != nil {
		delivery.mu.Unlock()
		return
	}
	if kind == KindNone {
		if !delivery.allFormatsLocked() || delivery.paused {
			delivery.pendingCommonResync = preferResyncReason(delivery.pendingCommonResync, reason)
			for _, track := range delivery.tracks {
				track.transitioning = track.transitioning || track.formatSent
			}
			delivery.mu.Unlock()
			return
		}
		reason = preferResyncReason(delivery.pendingCommonResync, reason)
		delivery.pendingCommonResync = ""
		for _, track := range delivery.tracks {
			track.transitioning = true
		}
	} else if track := delivery.tracks[kind]; track != nil {
		if !track.formatSent || delivery.paused {
			track.pendingReason = preferResyncReason(track.pendingReason, reason)
			track.transitioning = track.transitioning || track.formatSent
			delivery.mu.Unlock()
			return
		}
		track.transitioning = true
	}
	delivery.mu.Unlock()
	select {
	case delivery.resync <- resyncRequest{kind: kind, reason: reason}:
	default:
		delivery.close(backpressureClose("resync_queue_full"))
	}
}

func (delivery *Delivery) pump(track *deliveryTrack) {
	for {
		select {
		case <-delivery.ctx.Done():
			return
		case event, ok := <-track.subscription.Events():
			if !ok {
				if delivery.ctx.Err() == nil {
					delivery.close(backendClose("subscription_closed"))
				}
				return
			}
			if err := delivery.handleMediaEvent(track, event); err != nil {
				delivery.logger.Warn().Str("reason", "invalid_source_event").Msg("media delivery failed")
				if errors.Is(err, ErrCodecUnsupported) {
					delivery.close(codecClose("unsupported_format"))
				} else {
					delivery.close(backendClose("invalid_source_event"))
				}
				return
			}
		}
	}
}

func (delivery *Delivery) handleMediaEvent(track *deliveryTrack, event types.MediaEvent) error {
	kind := kindFromMedia(event.Source.Kind)
	if kind == KindNone {
		return ErrCodecUnsupported
	}
	delivery.mu.Lock()
	current, ok := delivery.tracks[kind]
	if !ok || current != track || event.Source.ID != track.subscription.Source().ID {
		delivery.mu.Unlock()
		return errors.New("media event source mismatch")
	}
	if delivery.paused {
		delivery.mu.Unlock()
		return nil
	}

	switch event.Type {
	case types.MediaEventTypeFormat:
		if !supportedSource(event.Source) {
			delivery.mu.Unlock()
			return ErrCodecUnsupported
		}
		if !completeSupportedSource(event.Source) {
			track.source = event.Source
			delivery.mu.Unlock()
			return nil
		}
		reason := track.pendingReason
		track.pendingReason = ""
		if track.suppressProviderTransition && track.formatSent {
			track.source = event.Source
			track.lastProviderGen = event.Source.Generation
			track.suppressProviderTransition = false
			delivery.mu.Unlock()
			return nil
		}
		if !track.formatSent {
			track.source = event.Source
			track.lastProviderGen = event.Source.Generation
			track.suppressProviderTransition = false
			record, err := delivery.initialFormatLocked(kind, track)
			if err == nil && !delivery.queue.pushLifecycle(record) {
				err = errors.New("lifecycle queue full")
			}
			pending := delivery.takePendingCommonResyncLocked()
			delivery.mu.Unlock()
			if err != nil {
				return err
			}
			if pending != "" {
				delivery.requestResync(KindNone, pending)
			}
			return nil
		}
		if reason == "" && track.lastProviderGen == event.Source.Generation && sameFormat(track.source, event.Source) {
			delivery.mu.Unlock()
			return nil
		}
		if reason == "" {
			reason = "format_change"
		}
		track.source = event.Source
		track.lastProviderGen = event.Source.Generation
		track.transitioning = true
		delivery.mu.Unlock()
		delivery.requestResync(kind, normalizeDiscontinuityReason(reason))
		return nil

	case types.MediaEventTypeDiscontinuity:
		if event.Discontinuity.Reason == "paused" {
			delivery.mu.Unlock()
			return nil
		}
		if event.Discontinuity.Reason == "resumed" && track.suppressProviderTransition {
			delivery.mu.Unlock()
			return nil
		}
		track.pendingReason = normalizeDiscontinuityReason(event.Discontinuity.Reason)
		track.transitioning = track.formatSent
		delivery.mu.Unlock()
		return nil

	case types.MediaEventTypeUnit:
		if !delivery.ready || !track.formatSent || track.transitioning {
			delivery.mu.Unlock()
			mediaWebSocketDrops.WithLabelValues(metricKind(kind), "egress", "not_ready").Inc()
			return nil
		}
		if event.Unit.Generation != 0 && track.lastProviderGen != 0 && event.Unit.Generation != track.lastProviderGen {
			delivery.mu.Unlock()
			delivery.requestResync(kind, "source_restart")
			return nil
		}
		if kind == KindVideo {
			encodedKeyframe, valid := vp8Keyframe(event.Unit.Data)
			if !valid || encodedKeyframe != event.Unit.Keyframe {
				delivery.mu.Unlock()
				return errors.New("VP8 keyframe marker mismatch")
			}
			if track.awaitKeyframe && !event.Unit.Keyframe {
				delivery.mu.Unlock()
				mediaWebSocketDrops.WithLabelValues("video", "egress", "keyframe_admission").Inc()
				return nil
			}
			if event.Unit.Keyframe {
				track.awaitKeyframe = false
			}
		}
		record, err := delivery.unitRecordLocked(kind, track, event.Unit)
		if err == nil {
			track.nextSequence++
		}
		if err != nil {
			delivery.mu.Unlock()
			return err
		}
		queued := delivery.queue.pushMedia(record)
		if !queued {
			if kind == KindAudio {
				for _, affected := range delivery.tracks {
					affected.transitioning = affected.transitioning || affected.formatSent
				}
			} else {
				track.transitioning = true
			}
		}
		delivery.mu.Unlock()
		if !queued {
			mediaWebSocketDrops.WithLabelValues(metricKind(kind), "egress", "queue_full").Inc()
			if kind == KindAudio {
				kind = KindNone
			}
			delivery.requestResync(kind, "server_overflow")
			return nil
		}
		return nil

	case types.MediaEventTypeEnd:
		delivery.mu.Unlock()
		delivery.close(backendClose("source_ended"))
		return nil
	default:
		delivery.mu.Unlock()
		return errors.New("unsupported media event")
	}
}

func (delivery *Delivery) initialFormatLocked(kind Kind, track *deliveryTrack) (queuedRecord, error) {
	track.generation = 1
	track.nextSequence = 0
	track.formatSent = true
	track.awaitKeyframe = kind == KindVideo
	track.transitioning = false
	return makeFormatRecord(kind, track.generation, track.source)
}

func (delivery *Delivery) allFormatsLocked() bool {
	if len(delivery.tracks) == 0 {
		return false
	}
	for _, track := range delivery.tracks {
		if !track.formatSent || !completeSupportedSource(track.source) {
			return false
		}
	}
	return true
}

func (delivery *Delivery) takePendingCommonResyncLocked() string {
	if delivery.pendingCommonResync == "" || !delivery.allFormatsLocked() || delivery.paused {
		return ""
	}
	reason := delivery.pendingCommonResync
	delivery.pendingCommonResync = ""
	return reason
}

func preferResyncReason(current, candidate string) string {
	if current == "server_overflow" || candidate == "" {
		return current
	}
	if candidate == "server_overflow" || current == "" {
		return candidate
	}
	return current
}

func (delivery *Delivery) unitRecordLocked(kind Kind, track *deliveryTrack, unit types.EncodedMediaUnit) (queuedRecord, error) {
	duration, err := normalizedDuration(kind, track.source, unit)
	if err != nil {
		return queuedRecord{}, err
	}
	pts, err := durationMicros(unit.PTS)
	if err != nil {
		return queuedRecord{}, err
	}
	dts := uint64(0)
	if unit.DTSValid {
		dts, err = durationMicros(unit.DTS)
		if err != nil {
			return queuedRecord{}, err
		}
	}
	flags := uint8(0)
	if unit.Keyframe && kind == KindVideo {
		flags |= FlagKeyframe
	}
	if unit.PTSValid {
		flags |= FlagPTSValid
	}
	if unit.DTSValid {
		flags |= FlagDTSValid
	}
	record := Record{
		Type:       RecordUnit,
		Kind:       kind,
		Flags:      flags,
		TrackID:    trackID(kind),
		Generation: track.generation,
		Sequence:   track.nextSequence,
		PTS:        int64(pts),
		DTS:        int64(dts),
		Duration:   duration,
		Payload:    unit.Data,
	}
	encoded, err := MarshalRecord(record)
	if err != nil {
		return queuedRecord{}, err
	}
	return queuedRecord{record: record, encoded: encoded, payloadBytes: len(record.Payload)}, nil
}

func makeFormatRecord(kind Kind, generation uint64, source types.MediaSource) (queuedRecord, error) {
	if !completeSupportedSource(source) {
		return queuedRecord{}, errors.New("incomplete or unsupported source format")
	}
	metadata := FormatMetadata{
		Schema:               "neko.media.format/1",
		SourceID:             source.ID,
		SourceGeneration:     source.Generation,
		Codec:                normalizedCodec(source),
		MIMEType:             canonicalMIMEType(source.Kind),
		ClockRate:            source.Codec.ClockRate,
		Channels:             source.Codec.Channels,
		CodedWidth:           source.Width,
		CodedHeight:          source.Height,
		DisplayWidth:         source.Width,
		DisplayHeight:        source.Height,
		FrameRateNumerator:   source.FrameRateNumerator,
		FrameRateDenominator: source.FrameRateDenominator,
		NominalBitrate:       source.NominalBitrate,
	}
	record := Record{
		Type:       RecordFormat,
		Kind:       kind,
		TrackID:    trackID(kind),
		Generation: generation,
		Metadata:   mustJSON(metadata),
	}
	encoded, err := MarshalRecord(record)
	if err != nil {
		return queuedRecord{}, err
	}
	return queuedRecord{record: record, encoded: encoded}, nil
}

func makeDiscontinuityRecord(kind Kind, generation uint64, reason string) (queuedRecord, error) {
	record := Record{
		Type:       RecordDiscontinuity,
		Kind:       kind,
		TrackID:    trackID(kind),
		Generation: generation,
		Metadata: mustJSON(DiscontinuityMetadata{
			Schema: "neko.media.discontinuity/1",
			Reason: reason,
		}),
	}
	encoded, err := MarshalRecord(record)
	if err != nil {
		return queuedRecord{}, err
	}
	return queuedRecord{record: record, encoded: encoded}, nil
}

func mustLifecycleRecord(record Record) queuedRecord {
	encoded, err := MarshalRecord(record)
	if err != nil {
		panic(err)
	}
	return queuedRecord{record: record, encoded: encoded, payloadBytes: len(record.Payload)}
}

func mustJSON(value any) []byte {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return encoded
}

func completeSupportedSource(source types.MediaSource) bool {
	return source.Generation > 0 && source.Generation <= MaxSafeInteger && supportedSource(source) &&
		(source.Kind != types.MediaKindVideo || (source.Width > 0 && source.Height > 0 && source.FrameRateNumerator > 0 && source.FrameRateDenominator > 0))
}

func sameFormat(left, right types.MediaSource) bool {
	return left.ID == right.ID && left.Generation == right.Generation && normalizedCodec(left) == normalizedCodec(right) &&
		left.Codec.ClockRate == right.Codec.ClockRate && left.Codec.Channels == right.Codec.Channels &&
		left.Width == right.Width && left.Height == right.Height && left.FrameRateNumerator == right.FrameRateNumerator &&
		left.FrameRateDenominator == right.FrameRateDenominator && left.NominalBitrate == right.NominalBitrate
}

func normalizeDiscontinuityReason(reason string) string {
	switch reason {
	case "source_switch", "source_restart", "format_change", "timestamp_reset", "server_overflow", "browser_resync", "resumed":
		return reason
	default:
		return "source_restart"
	}
}

func kindFromMedia(kind types.MediaKind) Kind {
	switch kind {
	case types.MediaKindAudio:
		return KindAudio
	case types.MediaKindVideo:
		return KindVideo
	default:
		return KindNone
	}
}

func trackID(kind Kind) uint32 {
	if kind == KindAudio {
		return 1
	}
	return 2
}

func durationMicros(value time.Duration) (uint64, error) {
	if value < 0 {
		return 0, errors.New("negative media timestamp")
	}
	micros := uint64(value / time.Microsecond)
	if micros > MaxSafeInteger {
		return 0, errors.New("media timestamp exceeds safe range")
	}
	return micros, nil
}

func normalizedDuration(kind Kind, source types.MediaSource, unit types.EncodedMediaUnit) (uint64, error) {
	if unit.Duration > 0 {
		micros, err := durationMicros(unit.Duration)
		if err == nil && micros > 0 && micros <= MaxUnitDurationMicros {
			return micros, nil
		}
		return 0, errors.New("invalid media duration")
	}
	if kind == KindVideo && source.FrameRateNumerator > 0 && source.FrameRateDenominator > 0 {
		duration := time.Second * time.Duration(source.FrameRateDenominator) / time.Duration(source.FrameRateNumerator)
		micros, err := durationMicros(duration)
		if err == nil && micros > 0 && micros <= MaxUnitDurationMicros {
			return micros, nil
		}
	}
	if kind == KindAudio {
		if duration, ok := opusPacketDurationMicros(unit.Data); ok {
			return duration, nil
		}
	}
	return 0, errors.New("media duration unavailable")
}

func opusPacketDurationMicros(packet []byte) (uint64, bool) {
	if len(packet) == 0 {
		return 0, false
	}
	configuration := packet[0] >> 3
	var frameDuration uint64
	switch {
	case configuration < 12:
		frameDuration = []uint64{10000, 20000, 40000, 60000}[configuration%4]
	case configuration < 16:
		frameDuration = []uint64{10000, 20000}[configuration%2]
	default:
		frameDuration = []uint64{2500, 5000, 10000, 20000}[configuration%4]
	}
	frames := uint64(1)
	switch packet[0] & 0x03 {
	case 1, 2:
		frames = 2
	case 3:
		if len(packet) < 2 {
			return 0, false
		}
		frames = uint64(packet[1] & 0x3f)
		if frames == 0 || frames > 48 {
			return 0, false
		}
	}
	duration := frameDuration * frames
	return duration, duration > 0 && duration <= 120000
}

func vp8Keyframe(payload []byte) (bool, bool) {
	if len(payload) == 0 {
		return false, false
	}
	return payload[0]&0x01 == 0, true
}

func backendClose(reason string) closeRequest {
	return closeRequest{endReason: "backend_error", code: 4500, reason: reason, failed: true}
}

func backpressureClose(reason string) closeRequest {
	return closeRequest{endReason: "backend_error", code: 4413, reason: reason, failed: true}
}
