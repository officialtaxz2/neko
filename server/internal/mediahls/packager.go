package mediahls

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/pkg/types"
)

const (
	PackagerIdleGrace       = 15 * time.Second
	RenditionFailureWindow  = 12 * time.Second
	ConventionalReadyWindow = 24 * time.Second
	LowLatencyReadyWindow   = 6 * time.Second
	PartDuration            = time.Second
	ParentDuration          = 6 * time.Second
)

var (
	ErrCodecUnsupported = errors.New("HLS source codec unsupported")
	ErrPackagerNotReady = errors.New("HLS packager not ready")
	ErrPackagerClosed   = errors.New("HLS packager closed")
	ErrObjectNotFound   = errors.New("HLS object not found")
)

type packagedPart struct {
	part     Part
	sequence uint64
}

type trackState struct {
	mu sync.RWMutex

	id                    string
	audio                 bool
	variant               Variant
	timescale             uint32
	generation            uint64
	discontinuitySequence uint64
	startedAt             time.Time
	baseMSN               uint64
	initURI               string
	initReady             bool
	codecConfig          []byte
	failed                bool
	lastOutput            time.Time
	segments              []Segment
	parts                 []packagedPart
	nextMSN               uint64
	nextPart              uint64
	llReady               bool
	hlsReady              bool

	muxer          *fragmentMuxer
	partIndex      int64
	partSamples    []fragmentSample
	partBytes      int
	partCapturedAt time.Time
	segmentData    []byte
	lastDTS        uint64
	dtsSet         bool
}

type packagerWorker struct {
	track        *trackState
	subscription types.MediaSubscription
	transcoder   transcoder
	source       types.MediaSource
	formatReady  bool
	metricActive bool
	wg           sync.WaitGroup
}

type Packager struct {
	logger   zerolog.Logger
	provider types.EncodedMediaProvider
	factory  transcoderFactory

	mu                    sync.Mutex
	store                 *ObjectStore
	tracks                map[string]*trackState
	retained              map[string]map[ObjectKind][]string
	readyModes            map[string]bool
	refs                  int
	running               bool
	stopping              bool
	closed                bool
	generation            uint64
	discontinuitySequence uint64
	nextMSN               uint64
	anchorSet             bool
	anchorPTS             time.Duration
	anchorWall            time.Time
	notify                chan struct{}
	restart               chan string
	runCancel             context.CancelFunc
	runDone               chan struct{}
	idle                  *time.Timer
}

func NewPackager(provider types.EncodedMediaProvider) (*Packager, error) {
	if provider == nil {
		return nil, ErrInvalidConfig
	}
	if err := validateGSTTranscoderElements(); err != nil {
		return nil, err
	}
	return newPackager(provider, gstTranscoderFactory{}), nil
}

func newPackager(provider types.EncodedMediaProvider, factory transcoderFactory) *Packager {
	return &Packager{
		logger:   log.With().Str("module", "mediahls").Str("submodule", "packager").Logger(),
		provider: provider,
		factory:  factory,
		store:      NewObjectStore(),
		tracks:     make(map[string]*trackState),
		retained:   make(map[string]map[ObjectKind][]string),
		readyModes: make(map[string]bool),
		nextMSN:    1,
		notify:     make(chan struct{}),
		restart:    make(chan string, 2),
	}
}

func (packager *Packager) Acquire(ctx context.Context, mode string) error {
	if !ValidMode(mode) {
		return ErrInvalidMode
	}
	for {
		packager.mu.Lock()
		if packager.closed {
			packager.mu.Unlock()
			return ErrPackagerClosed
		}
		if packager.stopping {
			notify := packager.notify
			packager.mu.Unlock()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-notify:
				continue
			}
		}
		packager.refs++
		if packager.idle != nil {
			packager.idle.Stop()
			packager.idle = nil
		}
		if !packager.running {
			packager.startLocked()
		}
		packager.mu.Unlock()
		break
	}

	timeout := ConventionalReadyWindow
	if mode == ModeLLHLS {
		timeout = LowLatencyReadyWindow
	}
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := packager.waitReady(waitCtx, mode); err != nil {
		packager.Release()
		return err
	}
	return nil
}

func (packager *Packager) Release() {
	packager.mu.Lock()
	defer packager.mu.Unlock()
	if packager.refs > 0 {
		packager.refs--
	}
	if packager.refs != 0 || !packager.running || packager.closed || packager.idle != nil {
		return
	}
	packager.idle = time.AfterFunc(PackagerIdleGrace, packager.stopIdle)
	packager.logger.Info().Dur("grace", PackagerIdleGrace).Msg("HLS packager idle stop scheduled")
}

func (packager *Packager) stopIdle() {
	packager.mu.Lock()
	if packager.refs != 0 || packager.closed {
		packager.idle = nil
		packager.mu.Unlock()
		return
	}
	cancel := packager.runCancel
	done := packager.runDone
	packager.running = false
	packager.stopping = true
	packager.runCancel = nil
	packager.runDone = nil
	packager.idle = nil
	packager.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
	packager.mu.Lock()
	packager.clearPublicationLocked()
	packager.stopping = false
	packager.signalLocked()
	packager.mu.Unlock()
	packager.logger.Info().Msg("HLS packager stopped after idle grace")
}

func (packager *Packager) Shutdown() {
	packager.mu.Lock()
	if packager.closed {
		packager.mu.Unlock()
		return
	}
	packager.closed = true
	if packager.idle != nil {
		packager.idle.Stop()
		packager.idle = nil
	}
	cancel := packager.runCancel
	done := packager.runDone
	packager.running = false
	packager.runCancel = nil
	packager.runDone = nil
	packager.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
	packager.mu.Lock()
	packager.clearPublicationLocked()
	packager.signalLocked()
	packager.mu.Unlock()
}

func (packager *Packager) startLocked() {
	ctx, cancel := context.WithCancel(context.Background())
	packager.running = true
	packager.runCancel = cancel
	packager.runDone = make(chan struct{})
	done := packager.runDone
	go func() {
		defer close(done)
		packager.supervise(ctx)
	}()
}

func (packager *Packager) supervise(ctx context.Context) {
	reason := "startup"
	for ctx.Err() == nil {
		workers, generationCancel, err := packager.startGeneration(ctx, reason)
		if err != nil {
			packager.logger.Error().Err(err).Str("reason", reason).Msg("HLS generation start failed")
			select {
			case <-ctx.Done():
				return
			case reason = <-packager.restart:
			case <-time.After(time.Second):
				reason = "worker_failure"
			}
			continue
		}

		select {
		case <-ctx.Done():
			generationCancel()
			stopPackagerWorkers(workers)
			return
		case reason = <-packager.restart:
			generationCancel()
			stopPackagerWorkers(workers)
			packager.drainRestarts()
		}
	}
}

func (packager *Packager) startGeneration(parent context.Context, reason string) ([]*packagerWorker, context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(parent)
	packager.mu.Lock()
	packager.generation++
	packager.discontinuitySequence++
	generation := packager.generation
	discontinuity := packager.discontinuitySequence
	baseMSN := packager.nextMSN
	startedAt := time.Now().UTC()
	packager.anchorSet = false
	packager.anchorPTS = 0
	packager.anchorWall = time.Time{}
	packager.resetGenerationLocked()
	packager.mu.Unlock()

	workers := make([]*packagerWorker, 0, 4)
	cleanupOnError := func(err error) ([]*packagerWorker, context.CancelFunc, error) {
		cancel()
		closePackagerWorkers(workers, false)
		return nil, func() {}, err
	}

	audioSources := packager.provider.Sources(types.MediaKindAudio)
	if len(audioSources) != 1 {
		hlsPackagerStarts.WithLabelValues("audio", "error").Inc()
		return cleanupOnError(ErrCodecUnsupported)
	}
	audioWorker, err := packager.newWorker(ctx, "audio", Variant{}, audioSources[0], generation, discontinuity, baseMSN, startedAt)
	if err != nil {
		return cleanupOnError(err)
	}
	workers = append(workers, audioWorker)

	for _, variant := range FixedVariants() {
		source, ok := exactSource(packager.provider.Sources(types.MediaKindVideo), variant.SourceID)
		if !ok {
			hlsPackagerStarts.WithLabelValues(variant.ID, "error").Inc()
			return cleanupOnError(types.ErrMediaSourceNotFound)
		}
		worker, err := packager.newWorker(ctx, variant.ID, variant, source, generation, discontinuity, baseMSN, startedAt)
		if err != nil {
			return cleanupOnError(err)
		}
		workers = append(workers, worker)
	}

	packager.mu.Lock()
	for _, worker := range workers {
		packager.tracks[worker.track.id] = worker.track
		worker.metricActive = true
		hlsPackagers.WithLabelValues(worker.track.id, "running").Inc()
		hlsPackagerStarts.WithLabelValues(worker.track.id, "success").Inc()
		hlsGenerations.WithLabelValues(worker.track.id, metricGenerationReason(reason)).Inc()
	}
	packager.signalLocked()
	packager.mu.Unlock()
	packager.logger.Info().Uint64("generation", generation).Str("reason", reason).Msg("HLS packager generation started")

	for _, worker := range workers {
		worker.wg.Add(2)
		go func(worker *packagerWorker) { defer worker.wg.Done(); packager.pumpInput(ctx, worker) }(worker)
		go func(worker *packagerWorker) { defer worker.wg.Done(); packager.pumpOutput(ctx, worker) }(worker)
	}
	go packager.monitorGeneration(ctx, workers, startedAt)
	return workers, cancel, nil
}

func (packager *Packager) newWorker(ctx context.Context, id string, variant Variant, source types.MediaSource, generation, discontinuity, baseMSN uint64, startedAt time.Time) (*packagerWorker, error) {
	audio := id == "audio"
	if audio {
		if source.Codec.Name != "opus" || source.Codec.ClockRate != 48_000 || source.Codec.Channels != 2 {
			return nil, ErrCodecUnsupported
		}
	} else if source.Codec.Name != "vp8" {
		return nil, ErrCodecUnsupported
	}
	subscription, err := packager.provider.Subscribe(ctx, types.SourceSubscriptionRequest{
		Kind:           source.Kind,
		Selector:       types.MediaSelector{Type: types.MediaSelectorTypeExact, ID: source.ID},
		QueueCapacity:  ProviderQueueCapacity,
		OverflowPolicy: types.MediaOverflowDropNewest,
		Backend:        BackendName,
		Observer:       packager,
	})
	if err != nil {
		hlsPackagerStarts.WithLabelValues(id, "error").Inc()
		return nil, err
	}
	var encoder transcoder
	if audio {
		encoder, err = packager.factory.NewAudio(source)
	} else {
		encoder, err = packager.factory.NewVideo(variant, source)
	}
	if err != nil {
		_ = subscription.Close()
		hlsPackagerStarts.WithLabelValues(id, "error").Inc()
		return nil, err
	}
	timescale := uint32(90_000)
	if audio {
		timescale = 48_000
	}
	track := &trackState{
		id: id, audio: audio, variant: variant, timescale: timescale,
		generation: generation, discontinuitySequence: discontinuity,
		startedAt: startedAt, baseMSN: baseMSN, nextMSN: baseMSN,
		muxer: newFragmentMuxer(timescale), partIndex: -1,
		lastOutput: startedAt,
	}
	return &packagerWorker{
		track: track, subscription: subscription, transcoder: encoder,
		source: types.CloneMediaSource(source),
	}, nil
}

func stopPackagerWorkers(workers []*packagerWorker) {
	closePackagerWorkers(workers, true)
}

func closePackagerWorkers(workers []*packagerWorker, countMetrics bool) {
	for _, worker := range workers {
		_ = worker.subscription.Close()
	}
	for _, worker := range workers {
		worker.wg.Wait()
		worker.transcoder.Close()
		if countMetrics && worker.metricActive {
			worker.track.mu.RLock()
			failed := worker.track.failed
			worker.track.mu.RUnlock()
			if failed {
				hlsPackagers.WithLabelValues(worker.track.id, "failed").Dec()
			} else {
				hlsPackagers.WithLabelValues(worker.track.id, "running").Dec()
			}
		}
	}
}

func exactSource(sources []types.MediaSource, id string) (types.MediaSource, bool) {
	for _, source := range sources {
		if source.ID == id {
			return source, true
		}
	}
	return types.MediaSource{}, false
}

func (packager *Packager) pumpInput(ctx context.Context, worker *packagerWorker) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-worker.subscription.Events():
			if ctx.Err() != nil {
				return
			}
			if !ok || event.Type == types.MediaEventTypeEnd {
				packager.requestRestart("source_end")
				return
			}
			switch event.Type {
			case types.MediaEventTypeFormat:
				if !workerFormatIdentityMatches(worker, event.Source) {
					packager.requestRestart("format_change")
					return
				}
				if !worker.track.audio && (event.Source.Width == 0 || event.Source.Height == 0 || event.Source.FrameRateNumerator == 0 || event.Source.FrameRateDenominator == 0) {
					// A cold provider first advertises identity/generation, then
					// publishes its complete caps before the first admissible unit.
					continue
				}
				if !workerFormatMatches(worker, event.Source) {
					packager.requestRestart("format_change")
					return
				}
				if worker.source.Generation != 0 && event.Source.Generation != worker.source.Generation {
					packager.requestRestart("source_restart")
					return
				}
				worker.source = types.CloneMediaSource(event.Source)
				worker.formatReady = true
			case types.MediaEventTypeUnit:
				if !worker.formatReady || event.Unit.Generation == 0 || event.Unit.Generation != worker.source.Generation {
					packager.requestRestart("source_restart")
					return
				}
				if !worker.transcoder.Push(event.Unit) {
					kind := "video"
					if worker.track.audio {
						kind = "audio"
					}
					hlsDrops.WithLabelValues(worker.track.id, "worker", kind, "queue_full").Inc()
					packager.requestRestart("worker_failure")
					return
				}
			case types.MediaEventTypeDiscontinuity:
				packager.requestRestart(normalizeGenerationReason(event.Discontinuity.Reason))
				return
			}
		}
	}
}

func workerFormatMatches(worker *packagerWorker, source types.MediaSource) bool {
	if !workerFormatIdentityMatches(worker, source) {
		return false
	}
	if worker.track.audio {
		return source.Kind == types.MediaKindAudio && source.Codec.Name == "opus" && source.Codec.ClockRate == 48_000 && source.Codec.Channels == 2
	}
	return source.Kind == types.MediaKindVideo && source.Codec.Name == "vp8" &&
		source.Width == worker.track.variant.Width && source.Height == worker.track.variant.Height &&
		source.FrameRateNumerator == worker.track.variant.FrameRate && source.FrameRateDenominator == 1
}

func workerFormatIdentityMatches(worker *packagerWorker, source types.MediaSource) bool {
	if worker == nil || worker.track == nil || source.Generation == 0 || source.ID != worker.source.ID || source.Kind != worker.source.Kind || source.Codec.Name != worker.source.Codec.Name {
		return false
	}
	if worker.track.audio {
		return source.Kind == types.MediaKindAudio && source.Codec.Name == "opus" && source.Codec.ClockRate == 48_000 && source.Codec.Channels == 2
	}
	return source.Kind == types.MediaKindVideo && source.Codec.Name == "vp8"
}

func (packager *Packager) pumpOutput(ctx context.Context, worker *packagerWorker) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-worker.transcoder.Drops():
			if ctx.Err() != nil {
				return
			}
			if !ok {
				if ctx.Err() == nil {
					packager.requestRestart("worker_failure")
				}
				return
			}
			kind := "video"
			if worker.track.audio {
				kind = "audio"
			}
			hlsDrops.WithLabelValues(worker.track.id, "worker", kind, "queue_full").Inc()
			packager.requestRestart("worker_failure")
			return
		case sample, ok := <-worker.transcoder.Samples():
			if ctx.Err() != nil {
				return
			}
			if !ok {
				if ctx.Err() == nil {
					packager.requestRestart("worker_failure")
				}
				return
			}
			worker.track.mu.RLock()
			failed := worker.track.failed
			worker.track.mu.RUnlock()
			if failed && !sample.DeltaUnit {
				packager.logger.Info().Str("variant", worker.track.id).Msg("HLS rendition rejoined on fresh keyframe")
				packager.requestRestart("rendition_rejoin")
				return
			}
			sample.Timestamp = worker.transcoder.CapturedAt(sample)
			if err := packager.acceptSample(worker.track, sample); err != nil {
				packager.logger.Warn().Err(err).Str("variant", worker.track.id).Msg("HLS sample rejected")
				packager.requestRestart("format_change")
				return
			}
		}
	}
}

func (packager *Packager) acceptSample(track *trackState, sample types.Sample) error {
	if len(sample.Data) == 0 || !sample.PTSValid {
		return ErrInvalidObject
	}
	dts := sample.PTS
	if sample.DTSValid {
		dts = sample.DTS
	}
	anchor, anchorWall, ok := packager.timeline(track, dts, !sample.DeltaUnit)
	if !ok || dts < anchor {
		return nil
	}

	track.mu.Lock()
	defer track.mu.Unlock()
	track.lastOutput = time.Now()
	if !track.audio && (sample.Width != track.variant.Width || sample.Height != track.variant.Height || sample.FrameRateNumerator != track.variant.FrameRate || sample.FrameRateDenominator != 1) {
		return errors.New("HLS transcoder output format does not match advertised rendition")
	}
	if track.initReady && len(sample.CodecConfig) > 0 && !slices.Equal(sample.CodecConfig, track.codecConfig) {
		return errors.New("HLS transcoder codec configuration changed")
	}
	if !track.initReady {
		if len(sample.CodecConfig) == 0 {
			return nil
		}
		var init []byte
		var err error
		if track.audio {
			init, err = buildAudioInit(sample.CodecConfig)
		} else {
			init, err = buildVideoInit(track.variant.Width, track.variant.Height, sample.CodecConfig)
		}
		if err != nil {
			return err
		}
		uri := fmt.Sprintf("init-%d.mp4", track.generation)
		if err := packager.publishLocked(track, ObjectInit, uri, 0, 0, init, sample.Timestamp); err != nil {
			return err
		}
		track.initURI = uri
		track.initReady = true
		track.codecConfig = slices.Clone(sample.CodecConfig)
	}

	elapsed := dts - anchor
	partIndex := int64(elapsed / PartDuration)
	if track.partIndex < 0 {
		// All renditions join a generation only at a common parent boundary;
		// video additionally waits for the aligned encoder IDR instead of
		// turning harmless startup skew into a generation restart loop.
		if partIndex%6 != 0 || (!track.audio && sample.DeltaUnit) {
			if !track.audio && sample.DeltaUnit {
				hlsDrops.WithLabelValues(track.id, "packager", "video", "keyframe_admission").Inc()
			}
			return nil
		}
	}
	if track.partIndex >= 0 && partIndex < track.partIndex {
		return errors.New("HLS transcode timeline regressed")
	}
	if track.partIndex >= 0 && partIndex > track.partIndex && !track.audio && partIndex%2 == 0 && sample.DeltaUnit {
		// Do not start an advertised independent part with a delta frame. A
		// bounded encoder/keyframe delay is dropped locally until its IDR arrives.
		hlsDrops.WithLabelValues(track.id, "packager", "video", "keyframe_admission").Inc()
		return nil
	}
	if track.partIndex >= 0 && partIndex > track.partIndex {
		if partIndex != track.partIndex+1 {
			return errors.New("HLS transcode timeline gap")
		}
		if err := packager.finishPartLocked(track, anchorWall); err != nil {
			return err
		}
	}
	if track.partIndex < 0 || partIndex != track.partIndex {
		track.partIndex = partIndex
		if track.startedAt.Before(anchorWall) {
			track.startedAt = anchorWall
		}
		track.partSamples = nil
		track.partBytes = 0
		track.partCapturedAt = sample.Timestamp
		if !track.audio && partIndex%2 == 0 && sample.DeltaUnit {
			return errors.New("HLS two-second boundary did not begin with IDR")
		}
	}

	duration := sample.Duration
	if duration <= 0 {
		if track.audio {
			duration = 1024 * time.Second / 48_000
		} else {
			duration = time.Second / time.Duration(track.variant.FrameRate)
		}
	}
	ptsTicks := durationTicks(sample.PTS-anchor, track.timescale)
	dtsTicks := durationTicks(dts-anchor, track.timescale)
	durationValue := durationTicks(duration, track.timescale)
	if durationValue == 0 || durationValue > uint64(^uint32(0)) {
		return ErrInvalidObject
	}
	if track.dtsSet && dtsTicks < track.lastDTS {
		return errors.New("HLS decode timestamp regressed")
	}
	if track.partBytes+len(sample.Data) > MaximumPartBytes {
		return ErrObjectLimit
	}
	track.partSamples = append(track.partSamples, fragmentSample{
		data: slices.Clone(sample.Data), dts: dtsTicks, pts: ptsTicks,
		duration: uint32(durationValue), keyframe: !sample.DeltaUnit,
	})
	track.lastDTS = dtsTicks
	track.dtsSet = true
	track.partBytes += len(sample.Data)
	return nil
}

func (packager *Packager) timeline(track *trackState, dts time.Duration, keyframe bool) (time.Duration, time.Time, bool) {
	packager.mu.Lock()
	defer packager.mu.Unlock()
	if !packager.anchorSet && track.id == "high" && keyframe {
		packager.anchorSet = true
		packager.anchorPTS = dts
		packager.anchorWall = time.Now().UTC()
		packager.signalLocked()
	}
	return packager.anchorPTS, packager.anchorWall, packager.anchorSet
}

func (packager *Packager) finishPartLocked(track *trackState, anchorWall time.Time) error {
	if len(track.partSamples) == 0 || track.partIndex < 0 {
		return ErrInvalidObject
	}
	data, err := track.muxer.fragment(track.partSamples)
	if err != nil {
		return err
	}
	sequence := track.baseMSN + uint64(track.partIndex/6)
	partNumber := uint64(track.partIndex % 6)
	uri := fmt.Sprintf("part-%d-%d.m4s", sequence, partNumber)
	if err := packager.publishLocked(track, ObjectPart, uri, sequence, partNumber, data, track.partCapturedAt); err != nil {
		return err
	}
	track.parts = append(track.parts, packagedPart{sequence: sequence, part: Part{URI: uri, Duration: 1, Independent: track.audio || partNumber%2 == 0}})
	track.nextMSN = sequence
	track.nextPart = partNumber + 1
	if len(track.segmentData)+len(data) > MaximumSegmentBytes {
		return ErrObjectLimit
	}
	track.segmentData = append(track.segmentData, data...)
	if partNumber == 5 {
		if len(track.parts) != 6 {
			return ErrInvalidObject
		}
		segmentURI := fmt.Sprintf("seg-%d.m4s", sequence)
		if err := packager.publishLocked(track, ObjectSegment, segmentURI, sequence, 0, track.segmentData, track.partCapturedAt); err != nil {
			return err
		}
		track.segments = append(track.segments, Segment{
			URI: segmentURI, Sequence: sequence, Duration: 6,
			ProgramDateTime: anchorWall.Add(time.Duration(sequence-track.baseMSN) * ParentDuration),
			Discontinuity: len(track.segments) == 0,
		})
		if len(track.segments) > 3 {
			track.segments = slices.Clone(track.segments[len(track.segments)-3:])
		}
		track.parts = nil
		track.segmentData = nil
		track.nextMSN = sequence + 1
		track.nextPart = 0
	}
	if len(track.parts) >= 3 || len(track.segments) > 0 {
		track.llReady = true
	}
	if len(track.segments) == 3 {
		track.hlsReady = true
	}
	packager.mu.Lock()
	if sequence >= packager.nextMSN {
		packager.nextMSN = sequence + 1
	}
	packager.signalLocked()
	packager.mu.Unlock()
	return nil
}

func (packager *Packager) publishLocked(track *trackState, kind ObjectKind, uri string, sequence, part uint64, data []byte, capturedAt time.Time) error {
	contentType := "video/mp4"
	if track.audio {
		contentType = "audio/mp4"
	}
	object, err := NewMediaObject(uri, track.id, kind, track.generation, sequence, part, contentType, data)
	if err != nil {
		return err
	}
	maximum := MaximumRetainedParents
	if kind == ObjectPart {
		maximum = MaximumRetainedParts
	} else if kind == ObjectInit {
		maximum = MaximumRetainedInits
	}
	packager.mu.Lock()
	defer packager.mu.Unlock()
	byKind := packager.retained[track.id]
	if byKind == nil {
		byKind = make(map[ObjectKind][]string)
		packager.retained[track.id] = byKind
	}
	retained := byKind[kind]
	for len(retained) >= maximum {
		packager.evictObjectLocked(track.id, kind, retained[0])
		retained = retained[1:]
	}
	byKind[kind] = retained
	for packager.store.RetainedBytes()+len(data) > MaximumRetainedBytes {
		if !packager.evictUnadvertisedLocked() {
			return ErrObjectLimit
		}
	}
	// Aggregate eviction can remove an object from this same rendition/kind.
	// Re-read the authoritative slice so an evicted URI is never reintroduced
	// into retention bookkeeping by the append below.
	retained = byKind[kind]
	if err := packager.store.Publish(object); err != nil {
		return err
	}
	byKind[kind] = append(byKind[kind], uri)
	label := metricObjectKind(kind)
	hlsObjects.WithLabelValues(track.id, label).Inc()
	hlsRetainedBytes.WithLabelValues(track.id, label).Add(float64(len(data)))
	hlsPublishedBytes.WithLabelValues(track.id, label).Add(float64(len(data)))
	if !capturedAt.IsZero() {
		delay := time.Since(capturedAt).Seconds()
		if delay >= 0 {
			hlsPublishDelay.WithLabelValues(track.id, label).Observe(delay)
		}
	}
	return nil
}

func (packager *Packager) evictUnadvertisedLocked() bool {
	type candidate struct {
		variant string
		kind    ObjectKind
		object  MediaObject
	}
	var oldest candidate
	found := false
	minimum := map[ObjectKind]int{
		ObjectInit:    1,
		ObjectPart:    6,
		ObjectSegment: 3,
	}
	for _, variant := range []string{"audio", "high", "medium", "low"} {
		byKind := packager.retained[variant]
		for _, kind := range []ObjectKind{ObjectInit, ObjectPart, ObjectSegment} {
			retained := byKind[kind]
			if len(retained) <= minimum[kind] {
				continue
			}
			object, ok := packager.store.Get(variant, retained[0])
			if !ok {
				byKind[kind] = retained[1:]
				return true
			}
			if !found || olderMediaObject(object, oldest.object) {
				oldest = candidate{variant: variant, kind: kind, object: object}
				found = true
			}
		}
	}
	if !found {
		return false
	}
	retained := packager.retained[oldest.variant][oldest.kind]
	packager.evictObjectLocked(oldest.variant, oldest.kind, retained[0])
	packager.retained[oldest.variant][oldest.kind] = retained[1:]
	return true
}

func olderMediaObject(left, right MediaObject) bool {
	if left.generation != right.generation {
		return left.generation < right.generation
	}
	if left.sequence != right.sequence {
		return left.sequence < right.sequence
	}
	if left.part != right.part {
		return left.part < right.part
	}
	return left.kind < right.kind
}

func (packager *Packager) evictObjectLocked(variant string, kind ObjectKind, uri string) {
	if object, ok := packager.store.Get(variant, uri); ok && packager.store.Remove(variant, uri) {
		label := metricObjectKind(kind)
		hlsObjects.WithLabelValues(variant, label).Dec()
		hlsRetainedBytes.WithLabelValues(variant, label).Sub(float64(object.Size()))
	}
}

func (packager *Packager) objectStore() *ObjectStore {
	packager.mu.Lock()
	defer packager.mu.Unlock()
	return packager.store
}

func (packager *Packager) monitorGeneration(ctx context.Context, workers []*packagerWorker, startedAt time.Time) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if now.Sub(startedAt) < RenditionFailureWindow {
				continue
			}
			for _, worker := range workers {
				track := worker.track
				track.mu.Lock()
				stalled := now.Sub(track.lastOutput) >= RenditionFailureWindow
				if stalled && !track.failed {
					track.failed = true
					hlsPackagers.WithLabelValues(track.id, "running").Dec()
					hlsPackagers.WithLabelValues(track.id, "failed").Inc()
					packager.logger.Warn().Str("variant", track.id).Msg("HLS rendition removed after output stall")
					packager.signalUpdate()
				}
				track.mu.Unlock()
				if stalled && track.audio {
					packager.requestRestart("worker_failure")
					return
				}
			}
		}
	}
}

func (packager *Packager) OnMediaSubscriptionDrop(drop types.MediaDrop) {
	hlsDrops.WithLabelValues(metricVariant(drop.SourceID), "provider", metricMediaKind(drop.Kind), "queue_full").Inc()
	packager.requestRestart("provider_overflow")
}

func (packager *Packager) requestRestart(reason string) {
	select {
	case packager.restart <- normalizeGenerationReason(reason):
	default:
	}
}

func (packager *Packager) drainRestarts() {
	for {
		select {
		case <-packager.restart:
		default:
			return
		}
	}
}

func (packager *Packager) waitReady(ctx context.Context, mode string) error {
	for {
		packager.mu.Lock()
		if packager.closed {
			packager.mu.Unlock()
			return ErrPackagerClosed
		}
		running := packager.running
		tracks := make(map[string]*trackState, len(packager.tracks))
		for id, track := range packager.tracks {
			tracks[id] = track
		}
		notify := packager.notify
		packager.mu.Unlock()
		if running && tracksReady(tracks, mode) {
			packager.recordReady(mode)
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %s", ErrPackagerNotReady, mode)
		case <-notify:
		}
	}
}

func tracksReady(tracks map[string]*trackState, mode string) bool {
	if len(tracks) != 4 {
		return false
	}
	for _, id := range []string{"audio", "high", "medium", "low"} {
		track := tracks[id]
		if track == nil {
			return false
		}
		track.mu.RLock()
		ready := track.initReady && !track.failed && ((mode == ModeHLS && track.hlsReady) || (mode == ModeLLHLS && track.llReady))
		track.mu.RUnlock()
		if !ready {
			return false
		}
	}
	return true
}

func (packager *Packager) Master() ([]byte, error) {
	packager.mu.Lock()
	audio := packager.tracks["audio"]
	tracks := make(map[string]*trackState, len(packager.tracks))
	for id, track := range packager.tracks {
		tracks[id] = track
	}
	packager.mu.Unlock()
	if audio == nil {
		return nil, ErrPackagerNotReady
	}
	audio.mu.RLock()
	audioReady := audio.initReady && !audio.failed
	audio.mu.RUnlock()
	if !audioReady {
		return nil, ErrPackagerNotReady
	}
	variants := make([]MasterVariant, 0, 3)
	for _, variant := range FixedVariants() {
		track := tracks[variant.ID]
		if track == nil {
			continue
		}
		track.mu.RLock()
		available := track.initReady && !track.failed
		track.mu.RUnlock()
		if !available {
			continue
		}
		variants = append(variants, MasterVariant{ID: variant.ID, URI: variant.ID + "/index.m3u8", Bandwidth: variant.Bandwidth, AverageBandwidth: variant.AverageBandwidth, Width: variant.Width, Height: variant.Height, FrameRate: float64(variant.FrameRate), VideoCodec: "avc1.64001f"})
	}
	return (MasterPlaylist{AudioURI: "audio/index.m3u8", Variants: variants}).Render()
}

func (packager *Packager) Position(variant string) (uint64, uint64, error) {
	track := packager.track(variant)
	if track == nil {
		return 0, 0, ErrObjectNotFound
	}
	track.mu.RLock()
	defer track.mu.RUnlock()
	if !track.initReady || track.failed {
		return 0, 0, ErrPackagerNotReady
	}
	return track.nextMSN, track.nextPart, nil
}

func (packager *Packager) Playlist(ctx context.Context, mode, variant string, directives PlaylistDirectives) ([]byte, error) {
	if mode == ModeLLHLS && directives.HasMSN {
		if err := packager.waitPosition(ctx, variant, directives); err != nil {
			return nil, err
		}
	}
	track := packager.track(variant)
	if track == nil {
		return nil, ErrObjectNotFound
	}
	track.mu.RLock()
	defer track.mu.RUnlock()
	if !track.initReady || track.failed || (mode == ModeHLS && !track.hlsReady) || (mode == ModeLLHLS && !track.llReady) {
		return nil, ErrPackagerNotReady
	}
	playlist := MediaPlaylist{
		Mode: mode, MediaSequence: track.baseMSN,
		DiscontinuitySequence: track.discontinuitySequence,
		MapURI: track.initURI, Segments: slices.Clone(track.segments),
	}
	if len(playlist.Segments) > 0 {
		playlist.MediaSequence = track.segments[0].Sequence
	}
	if mode == ModeHLS {
		return playlist.Render()
	}
	playlist.Parts = make([]Part, 0, len(track.parts))
	for _, part := range track.parts {
		playlist.Parts = append(playlist.Parts, part.part)
	}
	playlist.PartsProgramDateTime = track.startedAt.Add(time.Duration(track.nextMSN-track.baseMSN) * ParentDuration)
	playlist.PreloadHint = fmt.Sprintf("part-%d-%d.m4s", track.nextMSN, track.nextPart)
	playlist.RenditionReports = packager.renditionReportsLocked(variant)
	return playlist.Render()
}

func (packager *Packager) waitPosition(ctx context.Context, variant string, directives PlaylistDirectives) error {
	for {
		currentMSN, currentPart, err := packager.Position(variant)
		if err != nil {
			return err
		}
		if currentMSN > directives.MSN || (currentMSN == directives.MSN && directives.HasPart && currentPart > directives.Part) {
			return nil
		}
		packager.mu.Lock()
		notify := packager.notify
		packager.mu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-notify:
		}
	}
}

func (packager *Packager) renditionReportsLocked(exclude string) []RenditionReport {
	packager.mu.Lock()
	tracks := make(map[string]*trackState, len(packager.tracks))
	for id, track := range packager.tracks {
		tracks[id] = track
	}
	packager.mu.Unlock()
	reports := make([]RenditionReport, 0, 3)
	for _, id := range []string{"audio", "high", "medium", "low"} {
		if id == exclude {
			continue
		}
		track := tracks[id]
		if track == nil {
			continue
		}
		track.mu.RLock()
		if track.initReady && !track.failed {
			lastMSN, lastPart := track.nextMSN, uint64(0)
			if track.nextPart > 0 {
				lastPart = track.nextPart - 1
			} else if lastMSN > track.baseMSN {
				lastMSN--
				lastPart = 5
			}
			reports = append(reports, RenditionReport{URI: "../" + id + "/index.m3u8", LastMSN: lastMSN, LastPart: lastPart})
		}
		track.mu.RUnlock()
	}
	return reports
}

func (packager *Packager) Object(variant, uri string) (MediaObject, bool) {
	return packager.objectStore().Get(variant, uri)
}

func (packager *Packager) WaitObject(ctx context.Context, variant, uri string) (MediaObject, bool) {
	for {
		if object, ok := packager.Object(variant, uri); ok {
			return object, true
		}
		packager.mu.Lock()
		if packager.closed || !packager.running {
			packager.mu.Unlock()
			return MediaObject{}, false
		}
		notify := packager.notify
		packager.mu.Unlock()
		select {
		case <-ctx.Done():
			return MediaObject{}, false
		case <-notify:
		}
	}
}

func (packager *Packager) track(id string) *trackState {
	packager.mu.Lock()
	defer packager.mu.Unlock()
	return packager.tracks[id]
}

func (packager *Packager) resetGenerationLocked() {
	packager.tracks = make(map[string]*trackState)
	packager.readyModes = make(map[string]bool)
	packager.anchorSet = false
}

func (packager *Packager) clearPublicationLocked() {
	packager.store = NewObjectStore()
	packager.retained = make(map[string]map[ObjectKind][]string)
	packager.resetGenerationLocked()
	for _, variant := range []string{"audio", "high", "medium", "low"} {
		for _, kind := range []string{"init", "part", "segment"} {
			hlsObjects.WithLabelValues(variant, kind).Set(0)
			hlsRetainedBytes.WithLabelValues(variant, kind).Set(0)
		}
	}
}

func (packager *Packager) recordReady(mode string) {
	packager.mu.Lock()
	if !packager.readyModes[mode] {
		packager.readyModes[mode] = true
		packager.logger.Info().Str("mode", mode).Uint64("generation", packager.generation).Msg("HLS packager ready")
	}
	packager.mu.Unlock()
}

func (packager *Packager) signalUpdate() {
	packager.mu.Lock()
	packager.signalLocked()
	packager.mu.Unlock()
}

func (packager *Packager) signalLocked() {
	close(packager.notify)
	packager.notify = make(chan struct{})
}

func durationTicks(value time.Duration, timescale uint32) uint64 {
	if value <= 0 || timescale == 0 {
		return 0
	}
	seconds := uint64(value / time.Second)
	nanos := uint64(value % time.Second)
	return seconds*uint64(timescale) + (nanos*uint64(timescale)+uint64(time.Second)/2)/uint64(time.Second)
}

func normalizeGenerationReason(reason string) string {
	switch reason {
	case "startup", "format_change", "timestamp_reset", "source_end", "worker_failure", "rendition_rejoin":
		return reason
	case "source_restart", "source_switch":
		return "source_restart"
	case "provider_overflow", "queue_full":
		return "provider_overflow"
	case "resume", "resumed":
		return "resume"
	default:
		return "worker_failure"
	}
}

func metricGenerationReason(reason string) string { return normalizeGenerationReason(reason) }
