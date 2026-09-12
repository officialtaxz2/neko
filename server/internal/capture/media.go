package capture

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

const maximumMediaSubscriptionQueue = 64

var mediaSubscriptionSerial atomic.Uint64

var (
	mediaSubscriptionsActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "subscriptions_active",
		Namespace: "neko",
		Subsystem: "media",
		Help:      "Active encoded-media source subscriptions.",
	}, []string{"backend", "source_id", "kind"})
	mediaSubscriptionQueueCapacity = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "subscription_queue_capacity",
		Namespace: "neko",
		Subsystem: "media",
		Help:      "Configured bounded source-subscription queue capacity.",
		Buckets:   []float64{1, 2, 4, 8, 16, 32, 64},
	}, []string{"backend", "source_id", "kind"})
	mediaSubscriptionQueueDepth = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "subscription_queue_depth",
		Namespace: "neko",
		Subsystem: "media",
		Help:      "Observed encoded-unit queue depth after enqueue and dequeue.",
		Buckets:   []float64{0, 1, 2, 4, 8, 16, 32, 64},
	}, []string{"backend", "source_id", "kind"})
	mediaSubscriptionDeliveredUnits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "subscription_delivered_units_total",
		Namespace: "neko",
		Subsystem: "media",
		Help:      "Encoded media units delivered to subscription consumers.",
	}, []string{"backend", "source_id", "kind"})
	mediaSubscriptionDeliveredBytes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "subscription_delivered_bytes_total",
		Namespace: "neko",
		Subsystem: "media",
		Help:      "Encoded media bytes delivered to subscription consumers.",
	}, []string{"backend", "source_id", "kind"})
	mediaSubscriptionDroppedUnits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "subscription_dropped_units_total",
		Namespace: "neko",
		Subsystem: "media",
		Help:      "Encoded media units dropped locally by a bounded subscription.",
	}, []string{"backend", "source_id", "kind", "reason"})
	mediaSubscriptionDiscontinuities = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "subscription_discontinuities_total",
		Namespace: "neko",
		Subsystem: "media",
		Help:      "Encoded-media subscription discontinuities.",
	}, []string{"backend", "source_id", "kind", "reason"})
	mediaSourceGeneration = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "source_generation",
		Namespace: "neko",
		Subsystem: "media",
		Help:      "Current encoded-media source generation.",
	}, []string{"source_id", "kind"})
)

type streamGenerationProvider interface {
	Generation() uint64
}

type captureMediaProvider struct {
	audio     types.StreamSinkManager
	video     types.StreamSelectorManager
	startedAt time.Time

	sourcesMu sync.Mutex
	sources   map[string]cachedMediaSource
}

type cachedMediaSource struct {
	source            types.MediaSource
	captureGeneration uint64
	awaitingFormat    bool
	lastDTS           time.Duration
	lastDTSValid      bool
	reason            string
}

func newMediaProvider(audio types.StreamSinkManager, video types.StreamSelectorManager) *captureMediaProvider {
	provider := &captureMediaProvider{
		audio:     audio,
		video:     video,
		startedAt: time.Now(),
		sources:   make(map[string]cachedMediaSource),
	}

	provider.describe(types.MediaKindAudio, audio)
	for _, id := range video.IDs() {
		if stream, ok := video.GetStream(types.MediaSelector{ID: id}); ok {
			provider.describe(types.MediaKindVideo, stream)
		}
	}

	return provider
}

func mediaSourceKey(kind types.MediaKind, id string) string {
	return string(kind) + "/" + id
}

func (provider *captureMediaProvider) describe(kind types.MediaKind, stream types.StreamSinkManager) types.MediaSource {
	mediaCodec := mediaCodecFromRTP(stream.Codec())
	source := types.MediaSource{
		ID:      stream.ID(),
		Kind:    kind,
		Codec:   mediaCodec,
		Bitrate: stream.Bitrate(),
	}
	if nominal, ok := stream.(types.StreamSinkNominalBitrateProvider); ok {
		source.NominalBitrate = nominal.NominalBitrate()
	}
	captureGeneration := uint64(0)
	if generation, ok := stream.(streamGenerationProvider); ok {
		captureGeneration = generation.Generation()
	}

	key := mediaSourceKey(kind, stream.ID())
	provider.sourcesMu.Lock()
	awaitingFormat := captureGeneration != 0
	if previous, ok := provider.sources[key]; ok {
		source.Width = previous.source.Width
		source.Height = previous.source.Height
		source.FrameRateNumerator = previous.source.FrameRateNumerator
		source.FrameRateDenominator = previous.source.FrameRateDenominator
		source.Generation = previous.source.Generation
		awaitingFormat = previous.awaitingFormat
		if captureGeneration != previous.captureGeneration {
			source.Generation++
			awaitingFormat = true
			previous.lastDTSValid = false
			previous.reason = "source_restart"
		} else if !mediaCodecEqual(previous.source.Codec, source.Codec) {
			source.Generation++
			previous.lastDTSValid = false
			previous.reason = "format_change"
		}
		provider.sources[key] = previous
	} else {
		source.Generation = captureGeneration
	}
	cached := provider.sources[key]
	cached.source = types.CloneMediaSource(source)
	cached.captureGeneration = captureGeneration
	cached.awaitingFormat = awaitingFormat
	provider.sources[key] = cached
	provider.sourcesMu.Unlock()
	mediaSourceGeneration.WithLabelValues(source.ID, string(source.Kind)).Set(float64(source.Generation))

	return types.CloneMediaSource(source)
}

func mediaCodecFromRTP(source codec.RTPCodec) types.MediaCodec {
	parameters := make(map[string]string)
	for _, parameter := range strings.Split(source.Capability.SDPFmtpLine, ";") {
		parameter = strings.TrimSpace(parameter)
		if parameter == "" {
			continue
		}
		parts := strings.SplitN(parameter, "=", 2)
		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}
		parameters[parts[0]] = value
	}

	return types.MediaCodec{
		Name:       source.Name,
		MIMEType:   source.Capability.MimeType,
		ClockRate:  source.Capability.ClockRate,
		Channels:   source.Capability.Channels,
		Parameters: parameters,
	}
}

func (provider *captureMediaProvider) Sources(kind types.MediaKind) []types.MediaSource {
	streams := make([]types.StreamSinkManager, 0)
	switch kind {
	case types.MediaKindAudio:
		streams = append(streams, provider.audio)
	case types.MediaKindVideo:
		for _, id := range provider.video.IDs() {
			if stream, ok := provider.video.GetStream(types.MediaSelector{ID: id}); ok {
				streams = append(streams, stream)
			}
		}
	default:
		return nil
	}

	sources := make([]types.MediaSource, 0, len(streams))
	for _, stream := range streams {
		source := provider.describe(kind, stream)
		sources = append(sources, source)
	}
	return sources
}

func (provider *captureMediaProvider) observeSample(current types.MediaSource, sample types.Sample) (types.MediaSource, string) {
	key := mediaSourceKey(current.Kind, current.ID)
	provider.sourcesMu.Lock()
	cached, ok := provider.sources[key]
	if !ok {
		cached = cachedMediaSource{source: types.CloneMediaSource(current)}
	}
	source := cached.source
	generationChanged := false
	if sample.Generation != 0 && sample.Generation != cached.captureGeneration {
		source.Generation++
		cached.captureGeneration = sample.Generation
		cached.awaitingFormat = true
		cached.lastDTSValid = false
		cached.reason = "source_restart"
		generationChanged = true
	}
	if sample.DTSValid && cached.lastDTSValid && sample.DTS < cached.lastDTS && !generationChanged {
		source.Generation++
		cached.reason = "timestamp_reset"
		generationChanged = true
	}

	formatChanged := false
	if sample.Width != 0 && sample.Width != source.Width {
		source.Width = sample.Width
		formatChanged = true
	}
	if sample.Height != 0 && sample.Height != source.Height {
		source.Height = sample.Height
		formatChanged = true
	}
	if sample.FrameRateNumerator != 0 && sample.FrameRateNumerator != source.FrameRateNumerator {
		source.FrameRateNumerator = sample.FrameRateNumerator
		formatChanged = true
	}
	if sample.FrameRateDenominator != 0 && sample.FrameRateDenominator != source.FrameRateDenominator {
		source.FrameRateDenominator = sample.FrameRateDenominator
		formatChanged = true
	}
	if formatChanged && !generationChanged && !cached.awaitingFormat {
		source.Generation++
		cached.reason = "format_change"
		generationChanged = true
	}
	cached.awaitingFormat = false
	cached.lastDTS = sample.DTS
	cached.lastDTSValid = sample.DTSValid

	cached.source = types.CloneMediaSource(source)
	provider.sources[key] = cached
	provider.sourcesMu.Unlock()
	if generationChanged {
		mediaSourceGeneration.WithLabelValues(source.ID, string(source.Kind)).Set(float64(source.Generation))
	}
	return types.CloneMediaSource(source), cached.reason
}

func mediaSourceFormatChanged(previous, current types.MediaSource) bool {
	return previous.Generation != current.Generation ||
		!mediaCodecEqual(previous.Codec, current.Codec) ||
		previous.Width != current.Width ||
		previous.Height != current.Height ||
		previous.FrameRateNumerator != current.FrameRateNumerator ||
		previous.FrameRateDenominator != current.FrameRateDenominator
}

func mediaCodecEqual(previous, current types.MediaCodec) bool {
	if previous.Name != current.Name ||
		previous.MIMEType != current.MIMEType ||
		previous.ClockRate != current.ClockRate ||
		previous.Channels != current.Channels ||
		!bytes.Equal(previous.Config, current.Config) ||
		len(previous.Parameters) != len(current.Parameters) {
		return false
	}
	for key, value := range previous.Parameters {
		currentValue, ok := current.Parameters[key]
		if !ok || currentValue != value {
			return false
		}
	}
	return true
}

func (provider *captureMediaProvider) resolve(kind types.MediaKind, selector types.MediaSelector) (types.MediaSource, types.StreamSinkManager, bool) {
	source, ok := types.SelectMediaSource(provider.Sources(kind), selector)
	if !ok {
		return types.MediaSource{}, nil, false
	}

	if kind == types.MediaKindAudio {
		return source, provider.audio, source.ID == provider.audio.ID()
	}
	stream, ok := provider.video.GetStream(types.MediaSelector{ID: source.ID})
	return source, stream, ok
}

func (provider *captureMediaProvider) Subscribe(ctx context.Context, request types.SourceSubscriptionRequest) (types.MediaSubscription, error) {
	if request.Backend == "" {
		return nil, errors.New("media subscription backend cannot be empty")
	}
	if request.QueueCapacity <= 0 || request.QueueCapacity > maximumMediaSubscriptionQueue {
		return nil, fmt.Errorf("media subscription queue capacity must be between 1 and %d", maximumMediaSubscriptionQueue)
	}
	if request.OverflowPolicy == "" {
		request.OverflowPolicy = types.MediaOverflowDropNewest
	}
	if request.OverflowPolicy != types.MediaOverflowDropNewest {
		return nil, fmt.Errorf("unsupported media overflow policy: %s", request.OverflowPolicy)
	}

	source, stream, ok := provider.resolve(request.Kind, request.Selector)
	if !ok {
		return nil, types.ErrMediaSourceNotFound
	}
	if ctx == nil {
		ctx = context.Background()
	}

	subscription := &captureMediaSubscription{
		id:             strconv.FormatUint(mediaSubscriptionSerial.Add(1), 10),
		provider:       provider,
		backend:        request.Backend,
		observer:       request.Observer,
		overflowPolicy: request.OverflowPolicy,
		queueCapacity:  request.QueueCapacity,
		stream:         stream,
		source:         source,
		awaitKeyframe:  request.Kind == types.MediaKindVideo,
		signal:         make(chan struct{}, 1),
		events:         make(chan types.MediaEvent),
		done:           make(chan struct{}),
		abort:          make(chan struct{}),
	}

	go subscription.run()

	// The subscription remains inadmissible until capture demand has started and
	// the initial format is queued. A racing appsink sample is therefore dropped
	// locally rather than overtaking the format event.
	if err := stream.AddListener(subscription); err != nil {
		subscription.closeWithoutListener()
		return nil, err
	}

	subscription.mu.Lock()
	source = provider.describe(request.Kind, stream)
	subscription.source = source
	subscription.resetTimelineLocked(source.Generation)
	subscription.replacePendingLocked(types.MediaEvent{
		Type:   types.MediaEventTypeFormat,
		Source: types.CloneMediaSource(source),
	})
	subscription.ready = true
	subscription.mu.Unlock()
	subscription.notify()

	mediaSubscriptionsActive.WithLabelValues(request.Backend, source.ID, string(source.Kind)).Inc()
	mediaSubscriptionQueueCapacity.WithLabelValues(request.Backend, source.ID, string(source.Kind)).Observe(float64(request.QueueCapacity))

	go func() {
		select {
		case <-ctx.Done():
			_ = subscription.Close()
		case <-subscription.done:
		}
	}()

	return subscription, nil
}

type captureMediaSubscription struct {
	id             string
	provider       *captureMediaProvider
	backend        string
	observer       types.MediaSubscriptionObserver
	overflowPolicy types.MediaOverflowPolicy
	queueCapacity  int

	mu               sync.Mutex
	stream           types.StreamSinkManager
	source           types.MediaSource
	paused           bool
	ready            bool
	closed           bool
	awaitKeyframe    bool
	formatPublished  bool
	pending          []types.MediaEvent
	pendingUnits     int
	timelineGen      uint64
	timelineOffset   time.Duration
	timelineSet      bool
	lastDTS          time.Duration
	fallbackSequence uint64

	signal chan struct{}
	events chan types.MediaEvent
	done   chan struct{}
	abort  chan struct{}
}

func (subscription *captureMediaSubscription) ID() string {
	return subscription.id
}

func (subscription *captureMediaSubscription) Source() types.MediaSource {
	subscription.mu.Lock()
	defer subscription.mu.Unlock()

	source := types.CloneMediaSource(subscription.source)
	observed := subscription.provider.describe(source.Kind, subscription.stream)
	source.Bitrate = observed.Bitrate
	source.NominalBitrate = observed.NominalBitrate
	return types.CloneMediaSource(source)
}

func (subscription *captureMediaSubscription) Events() <-chan types.MediaEvent {
	return subscription.events
}

func (subscription *captureMediaSubscription) notify() {
	select {
	case subscription.signal <- struct{}{}:
	default:
	}
}

func (subscription *captureMediaSubscription) run() {
	defer close(subscription.events)
	defer close(subscription.done)

	for {
		subscription.mu.Lock()
		if len(subscription.pending) == 0 {
			subscription.mu.Unlock()
			select {
			case <-subscription.signal:
			case <-subscription.abort:
				return
			}
			continue
		}

		event := subscription.pending[0]
		subscription.pending = subscription.pending[1:]
		if event.Type == types.MediaEventTypeUnit {
			subscription.pendingUnits--
			mediaSubscriptionQueueDepth.WithLabelValues(subscription.backend, event.Source.ID, string(event.Source.Kind)).Observe(float64(subscription.pendingUnits))
		}
		subscription.mu.Unlock()

		select {
		case subscription.events <- event:
		case <-subscription.abort:
			return
		}
		if event.Type == types.MediaEventTypeFormat {
			subscription.mu.Lock()
			subscription.formatPublished = true
			subscription.mu.Unlock()
		}
		if event.Type == types.MediaEventTypeUnit {
			mediaSubscriptionDeliveredUnits.WithLabelValues(subscription.backend, event.Source.ID, string(event.Source.Kind)).Inc()
			mediaSubscriptionDeliveredBytes.WithLabelValues(subscription.backend, event.Source.ID, string(event.Source.Kind)).Add(float64(len(event.Unit.Data)))
		}
		if event.Type == types.MediaEventTypeEnd {
			return
		}
	}
}

func (subscription *captureMediaSubscription) replacePendingLocked(events ...types.MediaEvent) {
	subscription.pending = append(subscription.pending[:0], events...)
	subscription.pendingUnits = 0
}

func (subscription *captureMediaSubscription) transitionLocked(reason string, source types.MediaSource) {
	subscription.source = source
	subscription.awaitKeyframe = source.Kind == types.MediaKindVideo
	subscription.resetTimelineLocked(source.Generation)
	if !subscription.formatPublished {
		subscription.replacePendingLocked(types.MediaEvent{
			Type:   types.MediaEventTypeFormat,
			Source: types.CloneMediaSource(source),
		})
		return
	}
	subscription.replacePendingLocked(
		types.MediaEvent{
			Type:   types.MediaEventTypeDiscontinuity,
			Source: types.CloneMediaSource(source),
			Discontinuity: types.MediaDiscontinuity{
				Generation: source.Generation,
				Reason:     reason,
			},
		},
		types.MediaEvent{
			Type:   types.MediaEventTypeFormat,
			Source: types.CloneMediaSource(source),
		},
	)
	mediaSubscriptionDiscontinuities.WithLabelValues(subscription.backend, source.ID, string(source.Kind), reason).Inc()
}

func (subscription *captureMediaSubscription) resetTimelineLocked(generation uint64) {
	subscription.timelineGen = generation
	subscription.timelineOffset = 0
	subscription.timelineSet = false
	subscription.lastDTS = 0
	subscription.fallbackSequence = 0
}

func (subscription *captureMediaSubscription) Switch(_ context.Context, selector types.MediaSelector) error {
	source, target, ok := subscription.provider.resolve(subscription.source.Kind, selector)
	if !ok {
		return types.ErrMediaSourceNotFound
	}

	subscription.mu.Lock()
	if subscription.closed {
		subscription.mu.Unlock()
		return types.ErrMediaSubscriptionClosed
	}
	if target == subscription.stream {
		subscription.mu.Unlock()
		return nil
	}
	oldStream := subscription.stream
	oldSource := subscription.source
	oldAwaitKeyframe := subscription.awaitKeyframe
	paused := subscription.paused
	subscription.ready = false
	subscription.awaitKeyframe = source.Kind == types.MediaKindVideo
	subscription.mu.Unlock()

	if !paused {
		if err := oldStream.MoveListenerTo(subscription, target); err != nil {
			subscription.mu.Lock()
			subscription.ready = true
			subscription.source = oldSource
			subscription.awaitKeyframe = oldAwaitKeyframe
			subscription.mu.Unlock()
			return err
		}
		mediaSubscriptionsActive.WithLabelValues(subscription.backend, oldSource.ID, string(oldSource.Kind)).Dec()
	}

	source = subscription.provider.describe(source.Kind, target)
	subscription.mu.Lock()
	subscription.stream = target
	subscription.transitionLocked("source_switch", source)
	subscription.ready = !paused
	subscription.mu.Unlock()
	subscription.notify()

	if !paused {
		mediaSubscriptionsActive.WithLabelValues(subscription.backend, source.ID, string(source.Kind)).Inc()
		mediaSubscriptionQueueCapacity.WithLabelValues(subscription.backend, source.ID, string(source.Kind)).Observe(float64(subscription.queueCapacity))
	}
	return nil
}

func (subscription *captureMediaSubscription) SetPaused(paused bool) error {
	subscription.mu.Lock()
	if subscription.closed {
		subscription.mu.Unlock()
		return types.ErrMediaSubscriptionClosed
	}
	if subscription.paused == paused {
		subscription.mu.Unlock()
		return nil
	}
	subscription.paused = paused
	subscription.ready = false
	stream := subscription.stream
	source := subscription.source
	subscription.mu.Unlock()

	if paused {
		if err := stream.RemoveListener(subscription); err != nil {
			subscription.mu.Lock()
			subscription.paused = false
			subscription.ready = true
			subscription.mu.Unlock()
			return err
		}
		mediaSubscriptionsActive.WithLabelValues(subscription.backend, source.ID, string(source.Kind)).Dec()
		subscription.mu.Lock()
		subscription.transitionLocked("paused", source)
		subscription.mu.Unlock()
		subscription.notify()
		return nil
	}

	if err := stream.AddListener(subscription); err != nil {
		subscription.mu.Lock()
		subscription.paused = true
		subscription.mu.Unlock()
		return err
	}

	source = subscription.provider.describe(source.Kind, stream)
	subscription.mu.Lock()
	subscription.transitionLocked("resumed", source)
	subscription.ready = true
	subscription.mu.Unlock()
	subscription.notify()
	mediaSubscriptionsActive.WithLabelValues(subscription.backend, source.ID, string(source.Kind)).Inc()
	return nil
}

func (subscription *captureMediaSubscription) Close() error {
	subscription.mu.Lock()
	if subscription.closed {
		subscription.mu.Unlock()
		return nil
	}
	subscription.closed = true
	subscription.ready = false
	paused := subscription.paused
	stream := subscription.stream
	source := subscription.source
	subscription.mu.Unlock()

	var removeErr error
	if !paused {
		removeErr = stream.RemoveListener(subscription)
		if removeErr == nil {
			mediaSubscriptionsActive.WithLabelValues(subscription.backend, source.ID, string(source.Kind)).Dec()
		}
	}

	subscription.mu.Lock()
	subscription.replacePendingLocked(types.MediaEvent{
		Type:   types.MediaEventTypeEnd,
		Source: types.CloneMediaSource(source),
	})
	subscription.mu.Unlock()
	subscription.notify()
	return removeErr
}

func (subscription *captureMediaSubscription) closeWithoutListener() {
	subscription.mu.Lock()
	subscription.closed = true
	subscription.mu.Unlock()
	close(subscription.abort)
}

func (subscription *captureMediaSubscription) WriteSample(sample types.Sample) {
	subscription.mu.Lock()
	if subscription.closed || subscription.paused || !subscription.ready {
		subscription.mu.Unlock()
		return
	}

	source, reason := subscription.provider.observeSample(subscription.source, sample)
	formatChanged := mediaSourceFormatChanged(subscription.source, source)
	if formatChanged {
		if reason == "" {
			reason = "format_change"
		}
		subscription.transitionLocked(reason, source)
	}

	keyframe := !sample.DeltaUnit
	if subscription.awaitKeyframe && !keyframe {
		subscription.mu.Unlock()
		mediaSubscriptionDroppedUnits.WithLabelValues(subscription.backend, source.ID, string(source.Kind), "keyframe_admission").Inc()
		return
	}
	if keyframe {
		subscription.awaitKeyframe = false
	}

	sample.Generation = source.Generation
	unit := subscription.normalizeUnitLocked(sample)
	event := types.MediaEvent{
		Type:   types.MediaEventTypeUnit,
		Source: types.CloneMediaSource(source),
		Unit:   unit,
	}
	if subscription.pendingUnits >= subscription.queueCapacity {
		subscription.mu.Unlock()
		drop := types.MediaDrop{
			Backend:  subscription.backend,
			Kind:     source.Kind,
			SourceID: source.ID,
			Reason:   "queue_full",
		}
		mediaSubscriptionDroppedUnits.WithLabelValues(drop.Backend, drop.SourceID, string(drop.Kind), drop.Reason).Inc()
		if subscription.observer != nil {
			subscription.observer.OnMediaSubscriptionDrop(drop)
		}
		return
	}

	subscription.pending = append(subscription.pending, event)
	subscription.pendingUnits++
	depth := subscription.pendingUnits
	subscription.mu.Unlock()
	mediaSubscriptionQueueDepth.WithLabelValues(subscription.backend, source.ID, string(source.Kind)).Observe(float64(depth))
	subscription.notify()
}

func (subscription *captureMediaSubscription) normalizeUnitLocked(sample types.Sample) types.EncodedMediaUnit {
	capturedAt := sample.Timestamp
	if capturedAt.IsZero() {
		capturedAt = time.Now()
	}
	if sample.Generation != subscription.timelineGen {
		subscription.resetTimelineLocked(sample.Generation)
	}

	if !subscription.timelineSet {
		if sample.PTSValid {
			subscription.timelineOffset = capturedAt.Sub(subscription.provider.startedAt) - sample.PTS
		} else {
			subscription.timelineOffset = 0
		}
		subscription.timelineSet = true
	}

	pts := capturedAt.Sub(subscription.provider.startedAt)
	if sample.PTSValid {
		pts = subscription.timelineOffset + sample.PTS
	}
	if pts < 0 {
		pts = 0
	}

	dts := time.Duration(0)
	if sample.DTSValid {
		dts = subscription.timelineOffset + sample.DTS
		if dts < subscription.lastDTS {
			dts = subscription.lastDTS
		}
		subscription.lastDTS = dts
	}

	sequence := sample.Sequence
	if sequence == 0 {
		subscription.fallbackSequence++
		sequence = subscription.fallbackSequence
	}

	return types.EncodedMediaUnit{
		Generation: sample.Generation,
		Sequence:   sequence,
		PTS:        pts,
		PTSValid:   sample.PTSValid,
		DTS:        dts,
		DTSValid:   sample.DTSValid,
		Duration:   sample.Duration,
		Keyframe:   !sample.DeltaUnit,
		CapturedAt: capturedAt,
		Data:       sample.Data,
	}
}
