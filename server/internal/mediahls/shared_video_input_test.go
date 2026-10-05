package mediahls

import (
	"bytes"
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

type sharedInputTranscoder struct {
	*fakeTranscoder
	units chan types.EncodedMediaUnit
	failNext atomic.Bool
}

func (encoder *sharedInputTranscoder) Push(unit types.EncodedMediaUnit) bool {
	if encoder.failNext.Swap(false) {
		return false
	}
	unit.Data = bytes.Clone(unit.Data)
	encoder.units <- unit
	return true
}

func TestSharedVideoInputPreservesUnitsAndReportsPeerFailure(t *testing.T) {
	provider := newFakeHLSProvider()
	subscription, err := provider.Subscribe(context.Background(), types.SourceSubscriptionRequest{
		Kind: types.MediaKindVideo, Selector: types.MediaSelector{Type: types.MediaSelectorTypeExact, ID: "high"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer subscription.Close()
	packager := newPackager(provider, fakeTranscoderFactory{})
	var workers []*packagerWorker
	var encoders []*sharedInputTranscoder
	for _, variant := range FixedVariants() {
		encoder := &sharedInputTranscoder{
			fakeTranscoder: &fakeTranscoder{variant: variant, samples: make(chan types.Sample, 1), drops: make(chan struct{}, 1)},
			units: make(chan types.EncodedMediaUnit, 2),
		}
		encoders = append(encoders, encoder)
		workers = append(workers, &packagerWorker{
			track: &trackState{id: variant.ID, variant: variant, generation: 1},
			source: subscription.Source(), transcoder: encoder,
		})
	}
	owner := workers[0]
	owner.subscription, owner.inputPeers = subscription, workers[1:]
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); packager.pumpInput(ctx, owner) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("shared video input did not stop")
		}
	})
	unit := types.EncodedMediaUnit{
		Generation: 1, Sequence: 37, PTS: 30*time.Second + 837*time.Millisecond,
		DTS: 30*time.Second + 797*time.Millisecond, PTSValid: true, DTSValid: true,
		Duration: 40*time.Millisecond, Keyframe: true, CapturedAt: time.Unix(123, 456), Data: []byte{1, 2, 3},
	}
	provider.subscription("high").emit(unit)
	for index, encoder := range encoders {
		select {
		case got := <-encoder.units:
			if got.Generation != unit.Generation || got.Sequence != unit.Sequence || got.PTS != unit.PTS ||
				got.DTS != unit.DTS || got.PTSValid != unit.PTSValid || got.DTSValid != unit.DTSValid ||
				got.Duration != unit.Duration || got.Keyframe != unit.Keyframe || !got.CapturedAt.Equal(unit.CapturedAt) || !bytes.Equal(got.Data, unit.Data) {
				t.Fatalf("rendition %s changed the shared encoded unit", workers[index].track.id)
			}
		case <-time.After(time.Second):
			t.Fatalf("rendition %s missed shared input", workers[index].track.id)
		}
	}
	encoders[1].failNext.Store(true)
	unit.Sequence++
	unit.PTS += unit.Duration
	unit.DTS += unit.Duration
	provider.subscription("high").emit(unit)
	select {
	case reason := <-packager.restart:
		if reason != "worker_failure" {
			t.Fatalf("shared video peer failure reason = %s", reason)
		}
	case <-time.After(time.Second):
		t.Fatal("shared video peer rejection did not restart the generation")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("shared video input continued after peer rejection")
	}
	select {
	case <-encoders[2].units:
		t.Fatal("later rendition received a unit after peer rejection")
	default:
	}
	if !bytes.Equal(unit.Data, []byte{1, 2, 3}) || provider.count() != 1 {
		t.Fatal("shared video input mutated bytes or added subscriptions")
	}
}

type sharedConstructionFactory struct {
	encoders []*fakeTranscoder
	failVideo string
}

func (factory *sharedConstructionFactory) NewAudio(types.MediaSource) (transcoder, error) {
	encoder := &fakeTranscoder{audio: true, samples: make(chan types.Sample, 1), drops: make(chan struct{}, 1)}
	factory.encoders = append(factory.encoders, encoder)
	return encoder, nil
}

func (factory *sharedConstructionFactory) NewVideo(variant Variant, _ types.MediaSource) (transcoder, error) {
	if variant.ID == factory.failVideo {
		return nil, ErrCodecUnsupported
	}
	encoder := &fakeTranscoder{variant: variant, samples: make(chan types.Sample, 1), drops: make(chan struct{}, 1)}
	factory.encoders = append(factory.encoders, encoder)
	return encoder, nil
}

func TestSharedVideoConstructionFailureClosesOwnedResources(t *testing.T) {
	provider := newFakeHLSProvider()
	factory := &sharedConstructionFactory{failVideo: "low"}
	packager := newPackager(provider, factory)
	_, cancel, err := packager.startGeneration(context.Background(), "startup")
	defer cancel()
	if !errors.Is(err, ErrCodecUnsupported) {
		t.Fatalf("shared video construction error = %v", err)
	}
	if provider.count() != 2 || len(factory.encoders) != 3 {
		t.Fatal("failed construction opened unexpected subscriptions/encoders")
	}
	for _, id := range []string{"audio", "high"} {
		if provider.subscription(id).closed.Load() != 1 {
			t.Fatalf("owned subscription %s was not closed once", id)
		}
	}
	for _, encoder := range factory.encoders {
		if encoder.closed.Load() != 1 {
			t.Fatal("constructed encoder was not closed once")
		}
	}
}

func TestSharedVideoConstructionStopsBothInputsAndAllEncoders(t *testing.T) {
	provider := newFakeHLSProvider()
	factory := &sharedConstructionFactory{}
	packager := newPackager(provider, factory)
	workers, cancel, err := packager.startGeneration(context.Background(), "startup")
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	stopped := make(chan struct{})
	go func() { defer close(stopped); stopPackagerWorkers(workers) }()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("shared video workers did not stop")
	}
	if provider.count() != 2 || len(factory.encoders) != 4 || len(workers) != 4 {
		t.Fatal("shared worker set has unexpected ownership")
	}
	for _, id := range []string{"audio", "high"} {
		if provider.subscription(id).closed.Load() != 1 {
			t.Fatalf("owned subscription %s was not closed once", id)
		}
	}
	for _, worker := range workers {
		if (worker.track.id == "medium" || worker.track.id == "low") && worker.subscription != nil {
			t.Fatal("scaled rendition owns another source subscription")
		}
	}
	for _, encoder := range factory.encoders {
		if encoder.closed.Load() != 1 {
			t.Fatal("shared encoder was not closed once")
		}
	}
}
