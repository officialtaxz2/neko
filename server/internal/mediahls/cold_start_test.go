package mediahls

import (
	"context"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

// Sources is a pre-demand generation; Subscribe represents starting capture.
type coldHLSProvider struct {
	*fakeHLSProvider
	initialReason string
}

func (provider coldHLSProvider) Subscribe(ctx context.Context, request types.SourceSubscriptionRequest) (types.MediaSubscription, error) {
	opened, err := provider.fakeHLSProvider.Subscribe(ctx, request)
	if err != nil {
		return nil, err
	}
	subscription := opened.(*fakeHLSSubscription)
	<-subscription.events // replace the original fixture FORMAT
	subscription.source.Generation++
	complete := subscription.source
	cold := complete
	cold.Width, cold.Height = 0, 0
	cold.FrameRateNumerator, cold.FrameRateDenominator = 0, 0
	subscription.events <- types.MediaEvent{Type: types.MediaEventTypeFormat, Source: cold}
	subscription.events <- types.MediaEvent{
		Type: types.MediaEventTypeDiscontinuity, Source: complete,
		Discontinuity: types.MediaDiscontinuity{Generation: complete.Generation, Reason: provider.initialReason},
	}
	subscription.events <- types.MediaEvent{Type: types.MediaEventTypeFormat, Source: complete}
	return subscription, nil
}

func TestColdCaptureStartBindsOpenedGenerationAndInitialCaps(t *testing.T) {
	for _, reason := range []string{"format_change", "source_restart"} {
		t.Run(reason, func(t *testing.T) { testColdCaptureStart(t, reason) })
	}
}

func testColdCaptureStart(t *testing.T, reason string) {
	t.Helper()
	provider := coldHLSProvider{fakeHLSProvider: newFakeHLSProvider(), initialReason: reason}
	packager := newPackager(provider, fakeTranscoderFactory{})
	variant := FixedVariants()[0]
	planningSource, _ := exactSource(provider.Sources(types.MediaKindVideo), variant.SourceID)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	worker, err := packager.newWorker(ctx, variant.ID, variant, planningSource, 1, 1, 1, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	defer worker.transcoder.Close()
	defer worker.subscription.Close()
	if worker.source.Generation != 2 || planningSource.Generation != 1 {
		t.Fatal("worker retained the pre-demand generation")
	}
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		packager.pumpInput(ctx, worker)
	}()
	provider.subscription(variant.ID).emit(types.EncodedMediaUnit{
		Generation: 2, Keyframe: true, PTS: time.Second, DTS: time.Second,
		PTSValid: true, DTSValid: true, Data: []byte{1},
	})
	select {
	case <-worker.transcoder.Samples():
	case reason := <-packager.restart:
		t.Fatalf("initial generation/caps completion restarted capture: %s", reason)
	case <-time.After(time.Second):
		t.Fatal("first encoded unit did not reach the transcoder")
	}
	// A genuine restart after admission must still recreate the worker set.
	next := worker.source
	next.Generation++
	provider.subscription(variant.ID).events <- types.MediaEvent{
		Type: types.MediaEventTypeDiscontinuity, Source: next,
		Discontinuity: types.MediaDiscontinuity{Generation: next.Generation, Reason: "source_restart"},
	}
	select {
	case reason := <-packager.restart:
		if reason != "source_restart" {
			t.Fatalf("restart reason = %s", reason)
		}
	case <-time.After(time.Second):
		t.Fatal("genuine source restart was swallowed")
	}
	<-stopped
}

func TestColdCapsWithDifferentGenerationOrDimensionsStillRestart(t *testing.T) {
	for _, test := range []struct {
		name, reason     string
		changeGeneration bool
	}{
		{"format_generation", "format_change", true},
		{"format_dimensions", "format_change", false},
		{"restart_generation", "source_restart", true},
		{"restart_dimensions", "source_restart", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			variant := FixedVariants()[0]
			complete := types.MediaSource{
				ID: variant.ID, Kind: types.MediaKindVideo, Generation: 2,
				Codec: types.MediaCodec{Name: "vp8"},
				Width: variant.Width, Height: variant.Height,
				FrameRateNumerator: variant.FrameRate, FrameRateDenominator: 1,
			}
			worker := &packagerWorker{track: &trackState{id: variant.ID, variant: variant}, source: complete}
			if test.changeGeneration {
				complete.Generation++
			} else {
				complete.Width++
			}
			events := make(chan types.MediaEvent, 1)
			events <- types.MediaEvent{
				Type: types.MediaEventTypeDiscontinuity, Source: complete,
				Discontinuity: types.MediaDiscontinuity{Generation: complete.Generation, Reason: test.reason},
			}
			worker.subscription = &fakeHLSSubscription{events: events}
			packager := newPackager(newFakeHLSProvider(), fakeTranscoderFactory{})
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			packager.pumpInput(ctx, worker)
			select {
			case reason := <-packager.restart:
				if reason != test.reason {
					t.Fatalf("restart reason = %s", reason)
				}
			default:
				t.Fatal("incompatible format was ignored")
			}
		})
	}
}
