package mediahls

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

type fakeHLSProvider struct {
	mu   sync.Mutex
	subs map[string]*fakeHLSSubscription
}

func newFakeHLSProvider() *fakeHLSProvider { return &fakeHLSProvider{subs: make(map[string]*fakeHLSSubscription)} }

func (*fakeHLSProvider) Sources(kind types.MediaKind) []types.MediaSource {
	if kind == types.MediaKindAudio {
		return []types.MediaSource{{ID:"audio", Kind:kind, Codec:types.MediaCodec{Name:"opus", MIMEType:"audio/opus", ClockRate:48_000, Channels:2}, Generation:1}}
	}
	return []types.MediaSource{
		{ID:"high", Kind:kind, Codec:types.MediaCodec{Name:"vp8", MIMEType:"video/vp8", ClockRate:90_000}, Generation:1},
		{ID:"medium", Kind:kind, Codec:types.MediaCodec{Name:"vp8", MIMEType:"video/vp8", ClockRate:90_000}, Generation:1},
		{ID:"low", Kind:kind, Codec:types.MediaCodec{Name:"vp8", MIMEType:"video/vp8", ClockRate:90_000}, Generation:1},
	}
}

func (provider *fakeHLSProvider) Subscribe(_ context.Context, request types.SourceSubscriptionRequest) (types.MediaSubscription, error) {
	sources := provider.Sources(request.Kind)
	source, ok := exactSource(sources, request.Selector.ID)
	if !ok { return nil, types.ErrMediaSourceNotFound }
	subscription := &fakeHLSSubscription{id:source.ID, source:source, events:make(chan types.MediaEvent, 64)}
	if source.Kind == types.MediaKindVideo {
		for _, variant := range FixedVariants() {
			if variant.ID == source.ID {
				subscription.source.Width = variant.Width
				subscription.source.Height = variant.Height
				subscription.source.FrameRateNumerator = variant.FrameRate
				subscription.source.FrameRateDenominator = 1
				break
			}
		}
	}
	subscription.events <- types.MediaEvent{Type:types.MediaEventTypeFormat, Source:subscription.source}
	provider.mu.Lock()
	provider.subs[source.ID] = subscription
	provider.mu.Unlock()
	return subscription, nil
}

func (provider *fakeHLSProvider) subscription(id string) *fakeHLSSubscription {
	provider.mu.Lock(); defer provider.mu.Unlock()
	return provider.subs[id]
}

func (provider *fakeHLSProvider) count() int {
	provider.mu.Lock(); defer provider.mu.Unlock()
	return len(provider.subs)
}

type fakeHLSSubscription struct {
	id string
	source types.MediaSource
	events chan types.MediaEvent
	closed atomic.Int32
}

func (subscription *fakeHLSSubscription) ID() string { return subscription.id }
func (subscription *fakeHLSSubscription) Source() types.MediaSource { return subscription.source }
func (subscription *fakeHLSSubscription) Events() <-chan types.MediaEvent { return subscription.events }
func (*fakeHLSSubscription) Switch(context.Context, types.MediaSelector) error { return errors.New("not supported") }
func (*fakeHLSSubscription) SetPaused(bool) error { return nil }
func (subscription *fakeHLSSubscription) Close() error { subscription.closed.Add(1); return nil }
func (subscription *fakeHLSSubscription) emit(unit types.EncodedMediaUnit) {
	subscription.events <- types.MediaEvent{Type:types.MediaEventTypeUnit, Source:subscription.source, Unit:unit}
}

type fakeTranscoderFactory struct{}

func (fakeTranscoderFactory) NewAudio(types.MediaSource) (transcoder, error) {
	return &fakeTranscoder{audio:true, samples:make(chan types.Sample, 64), drops:make(chan struct{}, 1)}, nil
}
func (fakeTranscoderFactory) NewVideo(variant Variant, _ types.MediaSource) (transcoder, error) {
	return &fakeTranscoder{variant:variant, samples:make(chan types.Sample, 64), drops:make(chan struct{}, 1)}, nil
}

type fakeTranscoder struct {
	audio bool
	variant Variant
	samples chan types.Sample
	drops chan struct{}
	closed atomic.Int32
}

func (transcoder *fakeTranscoder) Samples() <-chan types.Sample { return transcoder.samples }
func (transcoder *fakeTranscoder) Drops() <-chan struct{} { return transcoder.drops }
func (*fakeTranscoder) CapturedAt(sample types.Sample) time.Time { return sample.Timestamp }
func (transcoder *fakeTranscoder) Push(unit types.EncodedMediaUnit) bool {
	sample := types.Sample{Data:append([]byte(nil),unit.Data...), Timestamp:unit.CapturedAt, PTS:unit.PTS, DTS:unit.DTS, PTSValid:true, DTSValid:true, Duration:unit.Duration, DeltaUnit:!unit.Keyframe}
	if transcoder.audio {
		sample.CodecConfig=[]byte{0x11,0x90}
	} else {
		sample.CodecConfig=[]byte{1,100,0,31,0xff,0xe1,0,1,0x67,1,0,1,0x68}
		sample.Width=transcoder.variant.Width; sample.Height=transcoder.variant.Height
		sample.FrameRateNumerator=transcoder.variant.FrameRate; sample.FrameRateDenominator=1
	}
	transcoder.samples <- sample
	return true
}
func (transcoder *fakeTranscoder) Close() { transcoder.closed.Add(1) }

func TestPackagerSharesOneWorkerSetAndPublishesBothModes(t *testing.T) {
	provider := newFakeHLSProvider()
	packager := newPackager(provider, fakeTranscoderFactory{})
	defer packager.Shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	first := make(chan error,1)
	go func(){ first <- packager.Acquire(ctx,ModeHLS) }()
	waitHLSCondition(t, func() bool { return provider.count()==2 })

	now := time.Now()
	emit := func(id string) {
		subscription := provider.subscription(id)
		for second:=0; second<=19; second++ {
			keyframe := id=="audio" || second%2==0
			subscription.emit(types.EncodedMediaUnit{Generation:1, Sequence:uint64(second+1), PTS:time.Duration(second)*time.Second, PTSValid:true, DTS:time.Duration(second)*time.Second, DTSValid:true, Duration:time.Second, Keyframe:keyframe, CapturedAt:now.Add(time.Duration(second)*time.Second), Data:[]byte{byte(second+1)}})
		}
	}
	emit("high")
	waitHLSCondition(t, func() bool { packager.mu.Lock(); defer packager.mu.Unlock(); return packager.anchorSet })
	emit("audio")
	if err:=<-first; err!=nil { t.Fatal(err) }
	if err:=packager.Acquire(ctx,ModeLLHLS); err!=nil { t.Fatal(err) }
	if provider.count()!=2 { t.Fatalf("provider subscriptions = %d",provider.count()) }
	master,err:=packager.Master(); if err!=nil||len(master)==0 { t.Fatalf("master = %q, %v",master,err) }
	conventional,err:=packager.Playlist(ctx,ModeHLS,"high",PlaylistDirectives{}); if err!=nil||len(conventional)==0 { t.Fatalf("HLS playlist = %q, %v",conventional,err) }
	lowLatency,err:=packager.Playlist(ctx,ModeLLHLS,"high",PlaylistDirectives{}); if err!=nil||len(lowLatency)==0 { t.Fatalf("LL-HLS playlist = %q, %v",lowLatency,err) }
	if _,ok:=packager.Object("high","seg-1.m4s");!ok { t.Fatal("first parent segment missing") }
	packager.Release(); packager.Release(); packager.Shutdown()
}

func TestAggregateRetentionEvictsOnlyUnadvertisedOldestObjects(t *testing.T) {
	packager := newPackager(newFakeHLSProvider(), fakeTranscoderFactory{})
	packager.retained["high"] = map[ObjectKind][]string{}
	for sequence := uint64(1); sequence <= 4; sequence++ {
		uri := fmt.Sprintf("seg-%d.m4s", sequence)
		object, err := NewMediaObject(uri, "high", ObjectSegment, 1, sequence, 0, "video/mp4", []byte{byte(sequence)})
		if err != nil {
			t.Fatal(err)
		}
		if err := packager.store.Publish(object); err != nil {
			t.Fatal(err)
		}
		packager.retained["high"][ObjectSegment] = append(packager.retained["high"][ObjectSegment], uri)
	}
	if !packager.evictUnadvertisedLocked() {
		t.Fatal("unadvertised object was not evicted")
	}
	if _, ok := packager.store.Get("high", "seg-1.m4s"); ok {
		t.Fatal("oldest unadvertised parent remains")
	}
	for sequence := uint64(2); sequence <= 4; sequence++ {
		if _, ok := packager.store.Get("high", fmt.Sprintf("seg-%d.m4s", sequence)); !ok {
			t.Fatalf("advertised parent %d was evicted", sequence)
		}
	}
	if packager.evictUnadvertisedLocked() {
		t.Fatal("advertised retention floor was evicted")
	}
}

func TestGenerationResetPreservesBoundedOldObjects(t *testing.T) {
	packager := newPackager(newFakeHLSProvider(), fakeTranscoderFactory{})
	object, err := NewMediaObject("init-1.mp4", "audio", ObjectInit, 1, 0, 0, "audio/mp4", []byte{1})
	if err != nil {
		t.Fatal(err)
	}
	if err := packager.store.Publish(object); err != nil {
		t.Fatal(err)
	}
	packager.retained["audio"] = map[ObjectKind][]string{ObjectInit: []string{"init-1.mp4"}}
	packager.resetGenerationLocked()
	if _, ok := packager.store.Get("audio", "init-1.mp4"); !ok {
		t.Fatal("generation reset discarded the previous immutable object")
	}
}

func TestWorkerRequiresCompleteStableProviderFormat(t *testing.T) {
	variant := FixedVariants()[0]
	source := types.MediaSource{
		ID: variant.SourceID, Kind: types.MediaKindVideo,
		Codec: types.MediaCodec{Name:"vp8", MIMEType:"video/vp8", ClockRate:90_000},
		Generation:1,
	}
	worker := &packagerWorker{track:&trackState{id:variant.ID, variant:variant}, source:source}
	if !workerFormatIdentityMatches(worker, source) || workerFormatMatches(worker, source) {
		t.Fatal("cold provider identity was not kept pending for complete caps")
	}
	// high is the full desktop: every complete geometry and rate is scaled to
	// the advertised renditions, not only the 1280x720 high output.
	for _, size := range []struct{ width, height, fps uint32 }{
		{1280, 720, 25}, {1920, 1080, 25}, {2560, 1440, 30}, {1280, 1024, 25}, {854, 480, 20},
	} {
		// The opened subscription had no caps yet (cold capture).
		complete := source
		complete.Width, complete.Height = size.width, size.height
		complete.FrameRateNumerator, complete.FrameRateDenominator = size.fps, 1
		if !workerFormatMatches(worker, complete) {
			t.Fatalf("complete %dx%d@%d provider format was rejected", size.width, size.height, size.fps)
		}
	}
	complete := source
	complete.Width, complete.Height = 1920, 1080
	complete.FrameRateNumerator, complete.FrameRateDenominator = 25, 1
	for name, mutate := range map[string]func(*types.MediaSource){
		"zero width": func(source *types.MediaSource) { source.Width = 0 },
		"zero height": func(source *types.MediaSource) { source.Height = 0 },
		"zero frame-rate numerator": func(source *types.MediaSource) { source.FrameRateNumerator = 0 },
		"zero frame-rate denominator": func(source *types.MediaSource) { source.FrameRateDenominator = 0 },
		"zero generation": func(source *types.MediaSource) { source.Generation = 0 },
		"other source": func(source *types.MediaSource) { source.ID = "medium" },
		"other codec": func(source *types.MediaSource) { source.Codec.Name = "h264" },
	} {
		candidate := complete
		mutate(&candidate)
		if workerFormatMatches(worker, candidate) {
			t.Fatalf("%s provider format was accepted", name)
		}
	}
	// Once a complete format is known, from the opened subscription or a ready
	// FORMAT, a geometry or rate change restarts the generation instead of
	// silently feeding running encoders.
	for _, ready := range []bool{false, true} {
		worker.source, worker.formatReady = complete, ready
		if !workerFormatMatches(worker, complete) {
			t.Fatalf("unchanged provider format was rejected (ready=%t)", ready)
		}
		resized := complete
		resized.Width, resized.Height = 1280, 720
		if workerFormatMatches(worker, resized) {
			t.Fatalf("resized provider format was accepted (ready=%t)", ready)
		}
		retimed := complete
		retimed.FrameRateNumerator = 30
		if workerFormatMatches(worker, retimed) {
			t.Fatalf("retimed provider format was accepted (ready=%t)", ready)
		}
	}
}

func waitHLSCondition(t *testing.T, condition func() bool) {
	t.Helper()
	deadline:=time.Now().Add(time.Second)
	for !condition() {
		if time.Now().After(deadline) { t.Fatal("condition timed out") }
		time.Sleep(time.Millisecond)
	}
}
