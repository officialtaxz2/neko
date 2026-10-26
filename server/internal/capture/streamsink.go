package capture

import (
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/pkg/gst"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

var moveSinkListenerMu = sync.Mutex{}

var _ types.StreamSinkNominalBitrateProvider = (*StreamSinkManagerCtx)(nil)
var _ types.MediaKeyframeRequester = (*StreamSinkManagerCtx)(nil)

// minKeyframeRequestInterval bounds forced keyframes per stream. All viewers of
// a tier share one encoder, and a browser repeats its request every 200 ms
// while it waits.
const minKeyframeRequestInterval = 500 * time.Millisecond

type StreamSinkManagerCtx struct {
	id string
	// nominal encoded stream rate in bits per second, or 0 if unspecified
	nominalBitrate uint64

	// wait for a keyframe before sending samples
	waitForKf bool

	// Encoded stream bitrate in bits per second. The bandwidth estimator uses
	// the same unit, so keep the byte-to-bit conversion at sample ingestion.
	bitrate    atomic.Uint64
	generation atomic.Uint64
	sequence   atomic.Uint64
	brBuckets  map[int]float64
	bitrateMu  sync.Mutex

	logger zerolog.Logger
	mu     sync.Mutex
	wg     sync.WaitGroup

	codec      codec.RTPCodec
	pipeline   gst.Pipeline
	pipelineMu sync.Mutex
	pipelineFn func() (string, error)
	// last forced keyframe, guarded by pipelineMu
	lastKeyframeRequest time.Time

	listeners   map[uintptr]types.SampleListener
	listenersKf map[uintptr]types.SampleListener // keyframe lobby
	listenersMu sync.Mutex

	// metrics
	currentListeners prometheus.Gauge
	currentBitrate   prometheus.Gauge
	totalBytes       prometheus.Counter
	frames           prometheus.Counter
	keyframes        prometheus.Counter
	keyframeRequests prometheus.Counter
	pipelinesCounter prometheus.Counter
	pipelinesActive  prometheus.Gauge
}

func streamSinkNew(codec codec.RTPCodec, pipelineFn func() (string, error), id string, nominalBitrate int) *StreamSinkManagerCtx {
	logger := log.With().
		Str("module", "capture").
		Str("submodule", "stream-sink").
		Str("id", id).Logger()

	manager := &StreamSinkManagerCtx{
		id: id,

		// only wait for keyframes if the codec is video
		waitForKf: codec.IsVideo(),

		brBuckets: map[int]float64{},

		logger:     logger,
		codec:      codec,
		pipelineFn: pipelineFn,

		listeners:   map[uintptr]types.SampleListener{},
		listenersKf: map[uintptr]types.SampleListener{},

		// metrics
		currentListeners: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "streamsink_listeners",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Current number of listeners for a pipeline.",
			ConstLabels: map[string]string{
				"video_id":   id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		}),
		currentBitrate: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "streamsink_bitrate",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Current encoded pipeline bitrate in bits per second.",
			ConstLabels: map[string]string{
				"video_id":   id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		}),
		totalBytes: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "streamsink_bytes",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of bytes created by the pipeline.",
			ConstLabels: map[string]string{
				"video_id":   id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		}),
		frames: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "streamsink_frames",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of video frames created by the pipeline; a rate below the configured fps means the encoder falls behind.",
			ConstLabels: map[string]string{
				"video_id":   id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		}),
		keyframes: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "streamsink_keyframes",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of video keyframes created by the pipeline.",
			ConstLabels: map[string]string{
				"video_id":   id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		}),
		keyframeRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "streamsink_keyframe_requests",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of keyframes forced on the encoder, after rate limiting.",
			ConstLabels: map[string]string{
				"video_id":   id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		}),
		pipelinesCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name:      "pipelines_total",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of created pipelines.",
			ConstLabels: map[string]string{
				"submodule":  "streamsink",
				"video_id":   id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		}),
		pipelinesActive: promauto.NewGauge(prometheus.GaugeOpts{
			Name:      "pipelines_active",
			Namespace: "neko",
			Subsystem: "capture",
			Help:      "Total number of active pipelines.",
			ConstLabels: map[string]string{
				"submodule":  "streamsink",
				"video_id":   id,
				"codec_name": codec.Name,
				"codec_type": codec.Type.String(),
			},
		}),
	}
	if nominalBitrate > 0 {
		manager.nominalBitrate = uint64(nominalBitrate)
	}

	return manager
}

func (manager *StreamSinkManagerCtx) shutdown() {
	manager.logger.Info().Msgf("shutdown")

	manager.listenersMu.Lock()
	for key := range manager.listeners {
		delete(manager.listeners, key)
	}
	for key := range manager.listenersKf {
		delete(manager.listenersKf, key)
	}
	manager.listenersMu.Unlock()

	manager.DestroyPipeline()
	manager.wg.Wait()
}

func (manager *StreamSinkManagerCtx) ID() string {
	return manager.id
}

func (manager *StreamSinkManagerCtx) Bitrate() uint64 {
	return manager.bitrate.Load()
}

func (manager *StreamSinkManagerCtx) NominalBitrate() uint64 {
	return manager.nominalBitrate
}

func (manager *StreamSinkManagerCtx) Generation() uint64 {
	return manager.generation.Load()
}

func (manager *StreamSinkManagerCtx) Codec() codec.RTPCodec {
	return manager.codec
}

func (manager *StreamSinkManagerCtx) start() error {
	if len(manager.listeners)+len(manager.listenersKf) == 0 {
		err := manager.CreatePipeline()
		if err != nil && !errors.Is(err, types.ErrCapturePipelineAlreadyExists) {
			return err
		}

		manager.logger.Info().Msgf("first listener, starting")
	}

	return nil
}

func (manager *StreamSinkManagerCtx) stop() {
	if len(manager.listeners)+len(manager.listenersKf) == 0 {
		manager.DestroyPipeline()
		manager.logger.Info().Msgf("last listener, stopping")
	}
}

func (manager *StreamSinkManagerCtx) addListener(listener types.SampleListener) {
	ptr := reflect.ValueOf(listener).Pointer()

	manager.listenersMu.Lock()
	if manager.waitForKf {
		// if we're waiting for a keyframe, add it to the keyframe lobby
		manager.listenersKf[ptr] = listener
	} else {
		// otherwise, add it as a regular listener
		manager.listeners[ptr] = listener
	}
	manager.listenersMu.Unlock()

	manager.logger.Debug().Interface("ptr", ptr).Msgf("adding listener")
	manager.currentListeners.Set(float64(manager.ListenersCount()))

	// the new listener waits for a keyframe, ask for one now; while the lobby
	// is not empty, onSample repeats the request
	manager.RequestKeyframe()
}

// RequestKeyframe forces a keyframe from the video encoder, at most once per
// minKeyframeRequestInterval. A failed request is not counted, so the next
// one is not held back.
func (manager *StreamSinkManagerCtx) RequestKeyframe() {
	if !manager.waitForKf {
		return
	}

	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	now := time.Now()
	if now.Sub(manager.lastKeyframeRequest) < minKeyframeRequestInterval {
		return
	}
	if !manager.pipeline.EmitVideoKeyframe() {
		manager.logger.Warn().Msg("failed to request keyframe")
		return
	}
	manager.lastKeyframeRequest = now
	manager.keyframeRequests.Inc()
}

func (manager *StreamSinkManagerCtx) removeListener(listener types.SampleListener) {
	ptr := reflect.ValueOf(listener).Pointer()

	manager.listenersMu.Lock()
	delete(manager.listeners, ptr)
	delete(manager.listenersKf, ptr) //	if it's a keyframe listener, remove it too
	manager.listenersMu.Unlock()

	manager.logger.Debug().Interface("ptr", ptr).Msgf("removing listener")
	manager.currentListeners.Set(float64(manager.ListenersCount()))
}

func (manager *StreamSinkManagerCtx) AddListener(listener types.SampleListener) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if listener == nil {
		return errors.New("listener cannot be nil")
	}

	// start if stopped
	if err := manager.start(); err != nil {
		return err
	}

	// add listener
	manager.addListener(listener)

	return nil
}

func (manager *StreamSinkManagerCtx) RemoveListener(listener types.SampleListener) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if listener == nil {
		return errors.New("listener cannot be nil")
	}

	// remove listener
	manager.removeListener(listener)

	// stop if started
	manager.stop()

	return nil
}

// moving listeners between streams ensures, that target pipeline is running
// before listener is added, and stops source pipeline if there are 0 listeners
func (manager *StreamSinkManagerCtx) MoveListenerTo(listener types.SampleListener, stream types.StreamSinkManager) error {
	if listener == nil {
		return errors.New("listener cannot be nil")
	}

	targetStream, ok := stream.(*StreamSinkManagerCtx)
	if !ok {
		return errors.New("target stream manager does not support moving listeners")
	}

	// we need to acquire both mutextes, from source stream and from target stream
	// in order to do that safely (without possibility of deadlock) we need third
	// global mutex, that ensures atomic locking

	// lock global mutex
	moveSinkListenerMu.Lock()

	// lock source stream
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// lock target stream
	targetStream.mu.Lock()
	defer targetStream.mu.Unlock()

	// unlock global mutex
	moveSinkListenerMu.Unlock()

	// start if stopped
	if err := targetStream.start(); err != nil {
		return err
	}

	// swap listeners
	manager.removeListener(listener)
	targetStream.addListener(listener)

	// stop if started
	manager.stop()

	return nil
}

func (manager *StreamSinkManagerCtx) ListenersCount() int {
	manager.listenersMu.Lock()
	defer manager.listenersMu.Unlock()

	return len(manager.listeners) + len(manager.listenersKf)
}

func (manager *StreamSinkManagerCtx) Started() bool {
	return manager.ListenersCount() > 0
}

func (manager *StreamSinkManagerCtx) CreatePipeline() error {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline != nil {
		return types.ErrCapturePipelineAlreadyExists
	}

	pipelineStr, err := manager.pipelineFn()
	if err != nil {
		return err
	}

	manager.logger.Info().
		Str("codec", manager.codec.Name).
		Str("src", pipelineStr).
		Msgf("creating pipeline")

	manager.pipeline, err = gst.CreatePipeline(pipelineStr)
	if err != nil {
		return err
	}
	manager.logger.Info().Msg("capture pipeline parsed")

	manager.pipeline.AttachAppsink("appsink")
	manager.logger.Info().Msg("capture appsink attached")
	manager.generation.Add(1)
	manager.sequence.Store(0)
	manager.logger.Info().Msg("capture pipeline play started")
	manager.pipeline.Play()
	manager.logger.Info().Msg("capture pipeline play completed")

	pipeline := manager.pipeline
	video := manager.codec.IsVideo()
	manager.wg.Go(func() { manager.emitSamples(pipeline, video) })

	manager.pipelinesCounter.Inc()
	manager.pipelinesActive.Set(1)

	return nil
}

// emitSamples passes the pipeline's samples to the listeners until the
// pipeline is destroyed.
func (manager *StreamSinkManagerCtx) emitSamples(pipeline gst.Pipeline, video bool) {
	manager.logger.Debug().Msg("started emitting samples")

	lastPTS := time.Duration(-1)
	// samples that may have been queued before a dropped one
	beforeDrop := 0
	for {
		sample, ok := <-pipeline.Sample()
		if !ok {
			manager.logger.Debug().Msg("stopped emitting samples")
			return
		}

		if video {
			sample.Duration = sampleDuration(lastPTS, sample)
			if sample.PTSValid {
				lastPTS = sample.PTS
			}
			manager.frames.Inc()

			// The pipeline dropped an encoded frame because this loop fell
			// behind. Later delta frames reference it, so every listener waits
			// for the next keyframe instead of showing a corrupted picture until
			// then. The drop happened while the sample channel was full, so a
			// keyframe among the samples queued then may precede it and must
			// not end the wait. If the drop happened only after this sample was
			// read, all samples still in the channel precede it, so this one
			// and the next cap(Sample()) samples count as queued.
			if beforeDrop > 0 {
				beforeDrop--
			}
			select {
			case _, ok := <-pipeline.Dropped():
				if ok {
					beforeDrop = cap(pipeline.Sample()) + 1
					manager.awaitKeyframe()
				}
			default:
			}
		}

		manager.onSample(sample, beforeDrop == 0)
	}
}

func (manager *StreamSinkManagerCtx) saveSampleBitrate(timestamp time.Time, sampleBytes float64) {
	manager.bitrateMu.Lock()
	defer manager.bitrateMu.Unlock()
	// get unix timestamp in seconds
	sec := timestamp.Unix()
	// last bucket is timestamp rounded to 3 seconds - 1 second
	last := int((sec - 1) % 3)
	// current bucket is timestamp rounded to 3 seconds
	curr := int(sec % 3)
	// next bucket is timestamp rounded to 3 seconds + 1 second
	next := int((sec + 1) % 3)

	if manager.brBuckets[next] != 0 {
		bitrate := uint64(manager.brBuckets[last])
		manager.bitrate.Store(bitrate)
		manager.currentBitrate.Set(float64(bitrate))
		// empty next bucket
		manager.brBuckets[next] = 0
	}

	// Samples are measured in bytes, while both the Pion bandwidth estimator
	// and StreamSinkManager.Bitrate use bits per second.
	manager.brBuckets[curr] += sampleBytes * 8
}

// maxSampleGap bounds the PTS gap used as a sample duration. Longer gaps mean a
// stalled source rather than late frames; the media WebSocket rejects unit
// durations above 10 seconds, so keep the buffer's own duration then.
const maxSampleGap = 10 * time.Second

// sampleDuration returns how far a consumer's clock (the RTP timestamp) must advance for this
// video sample: the time since the previous one. A buffer's own duration is not that: ximagesrc
// gives a frame it captures late only the time left until the next frame slot, and frames
// dropped before the encoder leave gaps. Advancing by it makes the stream's clock run slow
// under load, and browsers grow their jitter buffer to compensate (up to seconds).
// The gap is known only when the next sample arrives, so it is applied one frame late; the
// clock itself stays right.
func sampleDuration(lastPTS time.Duration, sample types.Sample) time.Duration {
	if !sample.PTSValid || lastPTS < 0 || sample.PTS <= lastPTS || sample.PTS-lastPTS > maxSampleGap {
		return sample.Duration
	}
	return sample.PTS - lastPTS
}

// awaitKeyframe moves every listener into the keyframe lobby and asks the
// encoder for a keyframe.
func (manager *StreamSinkManagerCtx) awaitKeyframe() {
	manager.listenersMu.Lock()
	for k, v := range manager.listeners {
		manager.listenersKf[k] = v
	}
	manager.listeners = make(map[uintptr]types.SampleListener)
	manager.listenersMu.Unlock()

	manager.logger.Debug().Msg("pipeline dropped a frame, waiting for keyframe")
	manager.RequestKeyframe()
}

// onSample passes a sample to the listeners. A keyframe ends the wait of the
// listeners in the keyframe lobby only if afterDrop is set, that is if it
// cannot precede a dropped frame.
func (manager *StreamSinkManagerCtx) onSample(sample types.Sample, afterDrop bool) {
	sample.Generation = manager.generation.Load()
	sample.Sequence = manager.sequence.Add(1)
	manager.listenersMu.Lock()

	// save to metrics
	length := float64(sample.Length)
	manager.totalBytes.Add(length)
	manager.saveSampleBitrate(sample.Timestamp, length)

	// if is not delta unit -> it can be decoded independently -> it is a keyframe
	keyframe := manager.waitForKf && !sample.DeltaUnit
	if keyframe && afterDrop && len(manager.listenersKf) > 0 {
		// if current sample is a keyframe, move listeners from
		// keyframe lobby to actual listeners map and clear lobby
		for k, v := range manager.listenersKf {
			manager.listeners[k] = v
		}
		manager.listenersKf = make(map[uintptr]types.SampleListener)
	}
	// a listener still waits, so the last request may have been lost
	waiting := len(manager.listenersKf) > 0

	// copy listeners before releasing lock to avoid holding it during dispatch
	listeners := make([]types.SampleListener, 0, len(manager.listeners))
	for _, l := range manager.listeners {
		listeners = append(listeners, l)
	}
	manager.listenersMu.Unlock()

	if keyframe {
		manager.keyframes.Inc()
	}
	if waiting {
		manager.RequestKeyframe()
	}

	for _, l := range listeners {
		l.WriteSample(sample)
	}
}

func (manager *StreamSinkManagerCtx) DestroyPipeline() {
	manager.pipelineMu.Lock()
	defer manager.pipelineMu.Unlock()

	if manager.pipeline == nil {
		return
	}

	manager.pipeline.Destroy()
	manager.logger.Info().Msgf("destroying pipeline")
	manager.pipeline = nil

	manager.pipelinesActive.Set(0)

	manager.bitrateMu.Lock()
	defer manager.bitrateMu.Unlock()
	manager.brBuckets = make(map[int]float64)
	manager.bitrate.Store(0)
	manager.currentBitrate.Set(0)
}
