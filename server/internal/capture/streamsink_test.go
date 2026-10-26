package capture

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/m1k1o/neko/server/pkg/gst"
	"github.com/m1k1o/neko/server/pkg/types"
)

func TestSaveSampleBitrateUsesBitsPerSecond(t *testing.T) {
	manager := &StreamSinkManagerCtx{
		brBuckets:      map[int]float64{},
		currentBitrate: prometheus.NewGauge(prometheus.GaugeOpts{Name: "test_streamsink_bitrate"}),
	}

	// Advancing through all three ring buckets publishes the completed middle
	// bucket. 125 encoded bytes in one second are 1,000 bits per second.
	manager.saveSampleBitrate(time.Unix(0, 0), 125)
	manager.saveSampleBitrate(time.Unix(1, 0), 125)
	manager.saveSampleBitrate(time.Unix(2, 0), 125)

	if got, want := manager.Bitrate(), uint64(1_000); got != want {
		t.Fatalf("Bitrate() = %d, want %d bits per second", got, want)
	}
}

func TestNominalBitrate(t *testing.T) {
	manager := &StreamSinkManagerCtx{nominalBitrate: 748_800}

	if got, want := manager.NominalBitrate(), uint64(748_800); got != want {
		t.Fatalf("NominalBitrate() = %d, want %d bits per second", got, want)
	}
}

func TestSampleDuration(t *testing.T) {
	const frame = 40 * time.Millisecond
	tests := []struct {
		name     string
		lastPTS  time.Duration
		pts      time.Duration
		ptsValid bool
		duration time.Duration
		want     time.Duration
	}{
		{"first sample keeps its duration", -1, 0, true, frame, frame},
		{"next frame", 0, frame, true, frame, frame},
		{"late capture: ximagesrc's short duration is not the gap", 0, frame, true, 2 * time.Millisecond, frame},
		{"two frames dropped before the encoder", 0, 3 * frame, true, frame, 3 * frame},
		{"unknown pts", 0, frame, false, frame, frame},
		{"timestamps going backwards", 2 * frame, frame, true, frame, frame},
		{"stalled source keeps its duration", 0, maxSampleGap + frame, true, frame, frame},
	}
	for _, tt := range tests {
		sample := types.Sample{PTS: tt.pts, PTSValid: tt.ptsValid, Duration: tt.duration}
		if got := sampleDuration(tt.lastPTS, sample); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

type fakeKeyframePipeline struct {
	gst.Pipeline
	emitted int
	fail    bool
}

func (pipeline *fakeKeyframePipeline) EmitVideoKeyframe() bool {
	if pipeline.fail {
		return false
	}
	pipeline.emitted++
	return true
}

type recordingListener struct {
	samples []types.Sample
}

func (listener *recordingListener) WriteSample(sample types.Sample) {
	listener.samples = append(listener.samples, sample)
}

func newKeyframeTestSink(pipeline gst.Pipeline) *StreamSinkManagerCtx {
	return &StreamSinkManagerCtx{
		waitForKf:        true,
		pipeline:         pipeline,
		brBuckets:        map[int]float64{},
		listeners:        map[uintptr]types.SampleListener{},
		listenersKf:      map[uintptr]types.SampleListener{},
		currentListeners: prometheus.NewGauge(prometheus.GaugeOpts{Name: "test_listeners"}),
		currentBitrate:   prometheus.NewGauge(prometheus.GaugeOpts{Name: "test_bitrate"}),
		totalBytes:       prometheus.NewCounter(prometheus.CounterOpts{Name: "test_bytes"}),
		frames:           prometheus.NewCounter(prometheus.CounterOpts{Name: "test_frames"}),
		keyframes:        prometheus.NewCounter(prometheus.CounterOpts{Name: "test_keyframes"}),
		keyframeRequests: prometheus.NewCounter(prometheus.CounterOpts{Name: "test_keyframe_requests"}),
	}
}

func TestRequestKeyframeIsRateLimited(t *testing.T) {
	pipeline := &fakeKeyframePipeline{}
	manager := newKeyframeTestSink(pipeline)

	manager.RequestKeyframe()
	manager.RequestKeyframe()
	if pipeline.emitted != 1 {
		t.Fatalf("emitted %d keyframes for two immediate requests, want 1", pipeline.emitted)
	}

	manager.lastKeyframeRequest = time.Now().Add(-minKeyframeRequestInterval)
	manager.RequestKeyframe()
	if pipeline.emitted != 2 {
		t.Fatalf("emitted %d keyframes after the interval, want 2", pipeline.emitted)
	}
}

func TestFailedKeyframeRequestDoesNotHoldBackTheNext(t *testing.T) {
	pipeline := &fakeKeyframePipeline{fail: true}
	manager := newKeyframeTestSink(pipeline)

	manager.RequestKeyframe()
	pipeline.fail = false
	manager.RequestKeyframe()
	if pipeline.emitted != 1 {
		t.Fatalf("emitted %d keyframes after a failed request, want 1", pipeline.emitted)
	}
}

func TestRequestKeyframeWithoutVideoPipeline(t *testing.T) {
	audio := newKeyframeTestSink(&fakeKeyframePipeline{})
	audio.waitForKf = false
	audio.RequestKeyframe()
	if got := audio.pipeline.(*fakeKeyframePipeline).emitted; got != 0 {
		t.Fatalf("audio stream emitted %d keyframes", got)
	}

	stopped := newKeyframeTestSink(nil)
	stopped.RequestKeyframe()
}

func TestWaitingListenerRepeatsKeyframeRequest(t *testing.T) {
	pipeline := &fakeKeyframePipeline{}
	manager := newKeyframeTestSink(pipeline)
	listener := &recordingListener{}

	manager.addListener(listener)
	if pipeline.emitted != 1 {
		t.Fatalf("joining listener emitted %d keyframes, want 1", pipeline.emitted)
	}

	// the requested keyframe got lost: the next delta frame repeats the request
	manager.lastKeyframeRequest = time.Now().Add(-minKeyframeRequestInterval)
	manager.onSample(types.Sample{DeltaUnit: true, Length: 1}, true)
	if pipeline.emitted != 2 {
		t.Fatalf("waiting listener emitted %d keyframes, want 2", pipeline.emitted)
	}
	if len(listener.samples) != 0 {
		t.Fatal("waiting listener received a delta frame")
	}

	manager.onSample(types.Sample{DeltaUnit: false, Length: 1}, true)
	manager.lastKeyframeRequest = time.Now().Add(-minKeyframeRequestInterval)
	manager.onSample(types.Sample{DeltaUnit: true, Length: 1}, true)
	if len(listener.samples) != 2 || listener.samples[0].DeltaUnit {
		t.Fatalf("listener received %d samples after the keyframe, want keyframe then delta", len(listener.samples))
	}
	if pipeline.emitted != 2 {
		t.Fatalf("emitted %d keyframes with nobody waiting, want 2", pipeline.emitted)
	}
}

func TestPipelineDropMakesListenersWaitForKeyframe(t *testing.T) {
	pipeline := &fakeKeyframePipeline{}
	manager := newKeyframeTestSink(pipeline)
	listener := &recordingListener{}
	manager.addListener(listener)
	manager.onSample(types.Sample{DeltaUnit: false, Length: 1}, true)

	manager.lastKeyframeRequest = time.Time{}
	manager.awaitKeyframe()
	if pipeline.emitted != 2 {
		t.Fatalf("dropped frame emitted %d keyframes in total, want 2", pipeline.emitted)
	}

	manager.onSample(types.Sample{DeltaUnit: true, Length: 1}, true)
	if len(listener.samples) != 1 {
		t.Fatal("listener received a delta frame after a dropped frame")
	}
	manager.onSample(types.Sample{DeltaUnit: false, Length: 1}, true)
	if len(listener.samples) != 2 {
		t.Fatal("listener did not resume with the next keyframe")
	}
}

func TestKeyframeQueuedBeforeDropDoesNotEndTheWait(t *testing.T) {
	manager := newKeyframeTestSink(&fakeKeyframePipeline{})
	listener := &recordingListener{}
	manager.addListener(listener)
	manager.onSample(types.Sample{DeltaUnit: false, Length: 1}, true)
	manager.awaitKeyframe()

	// this keyframe may have been queued before the dropped frame
	manager.onSample(types.Sample{DeltaUnit: false, Length: 1}, false)
	manager.onSample(types.Sample{DeltaUnit: true, Length: 1}, true)
	if len(listener.samples) != 1 {
		t.Fatal("a keyframe that may precede the dropped frame ended the wait")
	}
	manager.onSample(types.Sample{DeltaUnit: false, Length: 1}, true)
	if len(listener.samples) != 2 {
		t.Fatal("listener did not resume with a keyframe after the drop")
	}
}

// samplePipeline delivers samples like gst.Pipeline: a sample channel of the
// same capacity and a drop signal.
type samplePipeline struct {
	fakeKeyframePipeline
	sample  chan types.Sample
	dropped chan struct{}
}

func (pipeline *samplePipeline) Sample() chan types.Sample { return pipeline.sample }
func (pipeline *samplePipeline) Dropped() <-chan struct{}  { return pipeline.dropped }

type signalingListener struct {
	samples chan types.Sample
}

func (listener *signalingListener) WriteSample(sample types.Sample) {
	listener.samples <- sample
}

func TestDropAfterReadingASampleCoversTheWholeQueue(t *testing.T) {
	pipeline := &samplePipeline{
		sample:  make(chan types.Sample, 4),
		dropped: make(chan struct{}, 1),
	}
	manager := newKeyframeTestSink(pipeline)
	listener := &signalingListener{samples: make(chan types.Sample, 16)}
	manager.addListener(listener)

	done := make(chan struct{})
	go func() {
		defer close(done)
		manager.emitSamples(pipeline, true)
	}()

	pipeline.sample <- types.Sample{Sequence: 1, DeltaUnit: false, Length: 1}
	if sample := <-listener.samples; sample.DeltaUnit {
		t.Fatal("listener did not start with the keyframe")
	}

	// The drop is signalled before the loop checks it, after the next sample
	// was read: all samples queued behind that one precede the drop, the last
	// of them a keyframe.
	pipeline.dropped <- struct{}{}
	for _, delta := range []bool{true, true, true, true, false, false, true} {
		pipeline.sample <- types.Sample{DeltaUnit: delta, Length: 1}
	}
	close(pipeline.sample)
	<-done
	close(listener.samples)

	var got []bool
	for sample := range listener.samples {
		got = append(got, sample.DeltaUnit)
	}
	if len(got) != 2 || got[0] || !got[1] {
		t.Fatalf("listener received delta flags %v after the drop, want the second keyframe and its delta", got)
	}
}
