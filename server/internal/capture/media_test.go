package capture

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

type fakeMediaStream struct {
	mu          sync.Mutex
	id          string
	codec       codec.RTPCodec
	bitrate     uint64
	nominal     uint64
	generation  uint64
	listeners   map[types.SampleListener]struct{}
	addCount    int
	removeCount int
}

func newFakeMediaStream(id string, mediaCodec codec.RTPCodec, bitrate, nominal uint64) *fakeMediaStream {
	return &fakeMediaStream{
		id:         id,
		codec:      mediaCodec,
		bitrate:    bitrate,
		nominal:    nominal,
		generation: 1,
		listeners:  make(map[types.SampleListener]struct{}),
	}
}

func (stream *fakeMediaStream) ID() string             { return stream.id }
func (stream *fakeMediaStream) Codec() codec.RTPCodec  { return stream.codec }
func (stream *fakeMediaStream) Bitrate() uint64        { return stream.bitrate }
func (stream *fakeMediaStream) NominalBitrate() uint64 { return stream.nominal }
func (stream *fakeMediaStream) Generation() uint64     { return stream.generation }
func (stream *fakeMediaStream) CreatePipeline() error  { return nil }
func (stream *fakeMediaStream) DestroyPipeline()       {}
func (stream *fakeMediaStream) Started() bool          { return stream.ListenersCount() > 0 }
func (stream *fakeMediaStream) ListenersCount() int {
	stream.mu.Lock()
	defer stream.mu.Unlock()
	return len(stream.listeners)
}
func (stream *fakeMediaStream) AddListener(listener types.SampleListener) error {
	stream.mu.Lock()
	stream.listeners[listener] = struct{}{}
	stream.addCount++
	stream.mu.Unlock()
	return nil
}
func (stream *fakeMediaStream) RemoveListener(listener types.SampleListener) error {
	stream.mu.Lock()
	delete(stream.listeners, listener)
	stream.removeCount++
	stream.mu.Unlock()
	return nil
}
func (stream *fakeMediaStream) MoveListenerTo(listener types.SampleListener, target types.StreamSinkManager) error {
	if err := target.AddListener(listener); err != nil {
		return err
	}
	return stream.RemoveListener(listener)
}
func (stream *fakeMediaStream) emit(sample types.Sample) {
	stream.mu.Lock()
	listeners := make([]types.SampleListener, 0, len(stream.listeners))
	for listener := range stream.listeners {
		listeners = append(listeners, listener)
	}
	stream.mu.Unlock()
	for _, listener := range listeners {
		listener.WriteSample(sample)
	}
}

type fakeMediaSelector struct {
	codec   codec.RTPCodec
	ids     []string
	streams map[string]types.StreamSinkManager
}

func (selector *fakeMediaSelector) IDs() []string         { return append([]string(nil), selector.ids...) }
func (selector *fakeMediaSelector) Codec() codec.RTPCodec { return selector.codec }
func (selector *fakeMediaSelector) GetStream(request types.StreamSelector) (types.StreamSinkManager, bool) {
	stream, ok := selector.streams[request.ID]
	return stream, ok
}

type countingDropObserver struct {
	mu    sync.Mutex
	drops int
}

func (observer *countingDropObserver) OnMediaSubscriptionDrop(types.MediaDrop) {
	observer.mu.Lock()
	observer.drops++
	observer.mu.Unlock()
}

func (observer *countingDropObserver) count() int {
	observer.mu.Lock()
	defer observer.mu.Unlock()
	return observer.drops
}

func nextMediaEvent(t *testing.T, subscription types.MediaSubscription) types.MediaEvent {
	t.Helper()
	select {
	case event, ok := <-subscription.Events():
		if !ok {
			t.Fatal("media subscription closed before the expected event")
		}
		return event
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for media event")
		return types.MediaEvent{}
	}
}

func closeAndDrainMediaSubscription(t *testing.T, subscription types.MediaSubscription) {
	t.Helper()
	if err := subscription.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
	deadline := time.After(time.Second)
	for {
		select {
		case _, ok := <-subscription.Events():
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("subscription event channel did not close")
		}
	}
}

func newTestMediaProvider() (*captureMediaProvider, *fakeMediaStream, *fakeMediaStream, *fakeMediaStream) {
	audio := newFakeMediaStream("audio", codec.Opus(), 128_000, 0)
	high := newFakeMediaStream("high", codec.VP8(), 2_000_000, 1_996_800)
	low := newFakeMediaStream("low", codec.VP8(), 330_000, 332_800)
	selector := &fakeMediaSelector{
		codec: codec.VP8(),
		ids:   []string{"high", "low"},
		streams: map[string]types.StreamSinkManager{
			"high": high,
			"low":  low,
		},
	}
	return newMediaProvider(audio, selector), audio, high, low
}

func TestMediaProviderDiscoveryDemandAndKeyframeAdmission(t *testing.T) {
	provider, _, high, _ := newTestMediaProvider()
	sources := provider.Sources(types.MediaKindVideo)
	if len(sources) != 2 || sources[0].ID != "high" || sources[1].ID != "low" {
		t.Fatalf("video source order = %#v, want high then low", sources)
	}

	subscription, err := provider.Subscribe(context.Background(), types.SourceSubscriptionRequest{
		Kind:           types.MediaKindVideo,
		Selector:       types.MediaSelector{ID: "high"},
		QueueCapacity:  2,
		OverflowPolicy: types.MediaOverflowDropNewest,
		Backend:        "test",
	})
	if err != nil {
		t.Fatalf("Subscribe() error: %v", err)
	}
	if high.addCount != 1 || high.ListenersCount() != 1 {
		t.Fatalf("first subscription demand = adds %d/listeners %d, want 1/1", high.addCount, high.ListenersCount())
	}
	if event := nextMediaEvent(t, subscription); event.Type != types.MediaEventTypeFormat {
		t.Fatalf("first event = %s, want format", event.Type)
	}

	high.emit(types.Sample{Generation: 1, Sequence: 1, DeltaUnit: true, Timestamp: time.Now(), Duration: time.Millisecond})
	high.emit(types.Sample{Generation: 1, Sequence: 2, DeltaUnit: false, Timestamp: time.Now(), Duration: time.Millisecond, Data: []byte{1}})
	event := nextMediaEvent(t, subscription)
	if event.Type != types.MediaEventTypeUnit || !event.Unit.Keyframe || event.Unit.Sequence != 2 {
		t.Fatalf("admitted event = %#v, want keyframe sequence 2", event)
	}

	if err := subscription.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
	if err := subscription.Close(); err != nil {
		t.Fatalf("second Close() error: %v", err)
	}
	if event := nextMediaEvent(t, subscription); event.Type != types.MediaEventTypeEnd {
		t.Fatalf("final event = %s, want end", event.Type)
	}
	select {
	case _, ok := <-subscription.Events():
		if ok {
			t.Fatal("subscription event channel remained open after end")
		}
	case <-time.After(time.Second):
		t.Fatal("subscription event channel did not close after end")
	}
	if high.removeCount != 1 || high.ListenersCount() != 0 {
		t.Fatalf("last subscription release = removes %d/listeners %d, want 1/0", high.removeCount, high.ListenersCount())
	}
}

func TestMediaSubscriptionPublishesFormatBeforeRacingFirstUnit(t *testing.T) {
	provider, _, high, _ := newTestMediaProvider()
	subscription, err := provider.Subscribe(context.Background(), types.SourceSubscriptionRequest{
		Kind: types.MediaKindVideo, Selector: types.MediaSelector{ID: "high"}, QueueCapacity: 2,
		OverflowPolicy: types.MediaOverflowDropNewest, Backend: "test",
	})
	if err != nil {
		t.Fatalf("Subscribe() error: %v", err)
	}
	high.emit(types.Sample{
		Generation: 1, Sequence: 1, Timestamp: time.Now(), DeltaUnit: false,
		Width: 1280, Height: 720, FrameRateNumerator: 25, FrameRateDenominator: 1,
		Data: []byte{1},
	})

	if event := nextMediaEvent(t, subscription); event.Type != types.MediaEventTypeFormat {
		t.Fatalf("first racing event = %#v, want format", event)
	}
	for attempts := 0; attempts < 2; attempts++ {
		event := nextMediaEvent(t, subscription)
		if event.Type == types.MediaEventTypeUnit {
			closeAndDrainMediaSubscription(t, subscription)
			return
		}
		if event.Type != types.MediaEventTypeFormat {
			t.Fatalf("event before first unit = %#v, want updated format", event)
		}
	}
	t.Fatal("first encoded unit was not published after its format")
}

func TestMediaSubscriptionSwitchPauseResumeAndDiscontinuity(t *testing.T) {
	provider, _, high, low := newTestMediaProvider()
	subscription, err := provider.Subscribe(context.Background(), types.SourceSubscriptionRequest{
		Kind: types.MediaKindVideo, Selector: types.MediaSelector{ID: "high"}, QueueCapacity: 2,
		OverflowPolicy: types.MediaOverflowDropNewest, Backend: "test",
	})
	if err != nil {
		t.Fatalf("Subscribe() error: %v", err)
	}
	nextMediaEvent(t, subscription) // initial format

	if err := subscription.Switch(context.Background(), types.MediaSelector{ID: "low"}); err != nil {
		t.Fatalf("Switch() error: %v", err)
	}
	if event := nextMediaEvent(t, subscription); event.Type != types.MediaEventTypeDiscontinuity || event.Discontinuity.Reason != "source_switch" {
		t.Fatalf("switch event = %#v, want source_switch discontinuity", event)
	}
	if event := nextMediaEvent(t, subscription); event.Type != types.MediaEventTypeFormat || event.Source.ID != "low" {
		t.Fatalf("switch format = %#v, want low", event)
	}
	if high.ListenersCount() != 0 || low.ListenersCount() != 1 {
		t.Fatalf("switch demand high/low = %d/%d, want 0/1", high.ListenersCount(), low.ListenersCount())
	}

	low.emit(types.Sample{Generation: 1, Sequence: 3, DeltaUnit: true, Timestamp: time.Now()})
	low.emit(types.Sample{Generation: 1, Sequence: 4, DeltaUnit: false, Timestamp: time.Now(), Data: []byte{4}})
	if event := nextMediaEvent(t, subscription); event.Type != types.MediaEventTypeUnit || event.Source.ID != "low" || event.Unit.Sequence != 4 {
		t.Fatalf("switched unit = %#v, want low keyframe", event)
	}

	if err := subscription.SetPaused(true); err != nil {
		t.Fatalf("SetPaused(true) error: %v", err)
	}
	nextMediaEvent(t, subscription) // paused discontinuity
	nextMediaEvent(t, subscription) // paused format
	if low.ListenersCount() != 0 {
		t.Fatalf("paused listeners = %d, want 0", low.ListenersCount())
	}
	if err := subscription.SetPaused(false); err != nil {
		t.Fatalf("SetPaused(false) error: %v", err)
	}
	if event := nextMediaEvent(t, subscription); event.Type != types.MediaEventTypeDiscontinuity || event.Discontinuity.Reason != "resumed" {
		t.Fatalf("resume event = %#v, want resumed discontinuity", event)
	}
	nextMediaEvent(t, subscription) // resumed format
	if low.ListenersCount() != 1 {
		t.Fatalf("resumed listeners = %d, want 1", low.ListenersCount())
	}
	closeAndDrainMediaSubscription(t, subscription)
}

func TestMediaSubscriptionTimingGenerationAndFormatOrdering(t *testing.T) {
	provider, _, high, _ := newTestMediaProvider()
	provider.startedAt = time.Unix(100, 0)
	subscription, err := provider.Subscribe(context.Background(), types.SourceSubscriptionRequest{
		Kind: types.MediaKindVideo, Selector: types.MediaSelector{ID: "high"}, QueueCapacity: 2,
		OverflowPolicy: types.MediaOverflowDropNewest, Backend: "test",
	})
	if err != nil {
		t.Fatalf("Subscribe() error: %v", err)
	}
	nextMediaEvent(t, subscription)

	high.generation = 2
	high.emit(types.Sample{
		Generation: 2, Sequence: 10, Timestamp: provider.startedAt.Add(5 * time.Second),
		PTS: time.Second, PTSValid: true, DTS: 800 * time.Millisecond, DTSValid: true,
		Duration: 40 * time.Millisecond, DeltaUnit: false, Width: 1280, Height: 720,
		FrameRateNumerator: 25, FrameRateDenominator: 1, Data: []byte{1, 2, 3},
	})
	if event := nextMediaEvent(t, subscription); event.Type != types.MediaEventTypeDiscontinuity || event.Discontinuity.Generation != 2 {
		t.Fatalf("generation event = %#v, want generation 2 discontinuity", event)
	}
	format := nextMediaEvent(t, subscription)
	if format.Type != types.MediaEventTypeFormat || format.Source.Width != 1280 || format.Source.Height != 720 {
		t.Fatalf("format event = %#v, want 1280x720", format)
	}
	unit := nextMediaEvent(t, subscription)
	if unit.Type != types.MediaEventTypeUnit || unit.Unit.Generation != 2 || unit.Unit.Sequence != 10 {
		t.Fatalf("unit lifecycle = %#v, want generation 2 sequence 10", unit)
	}
	if unit.Unit.PTS != 5*time.Second || unit.Unit.DTS != 4800*time.Millisecond || !unit.Unit.DTSValid {
		t.Fatalf("normalized timing = PTS %v DTS %v valid %v", unit.Unit.PTS, unit.Unit.DTS, unit.Unit.DTSValid)
	}

	high.emit(types.Sample{
		Generation: 2, Sequence: 11, Timestamp: provider.startedAt.Add(6 * time.Second),
		PTS: 2 * time.Second, PTSValid: true, DTS: 1800 * time.Millisecond, DTSValid: true,
		Duration: 40 * time.Millisecond,
		DeltaUnit: false, Width: 640, Height: 360,
		FrameRateNumerator: 25, FrameRateDenominator: 1, Data: []byte{4, 5, 6},
	})
	if event := nextMediaEvent(t, subscription); event.Type != types.MediaEventTypeDiscontinuity || event.Discontinuity.Generation != 3 {
		t.Fatalf("caps-change event = %#v, want generation 3 discontinuity", event)
	}
	if event := nextMediaEvent(t, subscription); event.Type != types.MediaEventTypeFormat || event.Source.Width != 640 {
		t.Fatalf("caps-change format = %#v, want 640x360", event)
	}
	if event := nextMediaEvent(t, subscription); event.Type != types.MediaEventTypeUnit || event.Unit.Generation != 3 {
		t.Fatalf("caps-change unit = %#v, want generation 3", event)
	}

	high.emit(types.Sample{
		Generation: 2, Sequence: 12, Timestamp: provider.startedAt.Add(7 * time.Second),
		PTS: 3 * time.Second, PTSValid: true, DTS: 100 * time.Millisecond, DTSValid: true,
		Duration: 40 * time.Millisecond, DeltaUnit: false, Width: 640, Height: 360,
		FrameRateNumerator: 25, FrameRateDenominator: 1, Data: []byte{7, 8, 9},
	})
	if event := nextMediaEvent(t, subscription); event.Type != types.MediaEventTypeDiscontinuity || event.Discontinuity.Generation != 4 || event.Discontinuity.Reason != "timestamp_reset" {
		t.Fatalf("timestamp-reset event = %#v, want generation 4 discontinuity", event)
	}
	nextMediaEvent(t, subscription) // timestamp-reset format
	if event := nextMediaEvent(t, subscription); event.Type != types.MediaEventTypeUnit || event.Unit.Generation != 4 {
		t.Fatalf("timestamp-reset unit = %#v, want generation 4", event)
	}
	closeAndDrainMediaSubscription(t, subscription)
}

func TestMediaSubscriptionOverflowIsLocalAndNonBlocking(t *testing.T) {
	provider, _, high, _ := newTestMediaProvider()
	observer := &countingDropObserver{}
	slow, err := provider.Subscribe(context.Background(), types.SourceSubscriptionRequest{
		Kind: types.MediaKindVideo, Selector: types.MediaSelector{ID: "high"}, QueueCapacity: 2,
		OverflowPolicy: types.MediaOverflowDropNewest, Backend: "slow", Observer: observer,
	})
	if err != nil {
		t.Fatalf("slow Subscribe() error: %v", err)
	}
	fast, err := provider.Subscribe(context.Background(), types.SourceSubscriptionRequest{
		Kind: types.MediaKindVideo, Selector: types.MediaSelector{ID: "high"}, QueueCapacity: 2,
		OverflowPolicy: types.MediaOverflowDropNewest, Backend: "fast",
	})
	if err != nil {
		t.Fatalf("fast Subscribe() error: %v", err)
	}
	nextMediaEvent(t, slow)
	nextMediaEvent(t, fast)

	fastUnits := make(chan struct{}, 1)
	go func() {
		for event := range fast.Events() {
			if event.Type == types.MediaEventTypeUnit {
				select {
				case fastUnits <- struct{}{}:
				default:
				}
			}
		}
	}()

	finished := make(chan struct{})
	go func() {
		for index := 0; index < 100; index++ {
			high.emit(types.Sample{Generation: 1, Sequence: uint64(index + 1), Timestamp: time.Now(), DeltaUnit: false, Data: []byte{1}})
		}
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("sample fan-out blocked on a slow subscription")
	}
	select {
	case <-fastUnits:
	case <-time.After(time.Second):
		t.Fatal("healthy subscription received no media while another overflowed")
	}
	if observer.count() == 0 {
		t.Fatal("slow subscription did not report bounded queue overflow")
	}
	slowDrained := make(chan struct{})
	go func() {
		for range slow.Events() {
		}
		close(slowDrained)
	}()
	_ = slow.Close()
	_ = fast.Close()
	select {
	case <-slowDrained:
	case <-time.After(time.Second):
		t.Fatal("slow subscription did not close after its consumer resumed")
	}
}
