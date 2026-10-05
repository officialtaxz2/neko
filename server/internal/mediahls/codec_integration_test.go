//go:build hlsintegration

package mediahls

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/gst"
	"github.com/m1k1o/neko/server/pkg/types"
)

// Real VP8/Opus fixture encoders feed the production H.264/AAC transcoders and
// packager. The fixture supplies a shared, nonzero input clock; it does not
// stand in for target capture skew, authorization, proxy or browser acceptance.
type codecFixtureProvider struct {
	origin    time.Time
	errors    chan error
	sceneCuts bool
}

func (provider *codecFixtureProvider) Sources(kind types.MediaKind) []types.MediaSource {
	return newFakeHLSProvider().Sources(kind)
}

func (provider *codecFixtureProvider) Subscribe(ctx context.Context, request types.SourceSubscriptionRequest) (types.MediaSubscription, error) {
	source, ok := exactSource(provider.Sources(request.Kind), request.Selector.ID)
	if !ok {
		return nil, types.ErrMediaSourceNotFound
	}
	source.Generation = 2 // starting source demand advances the planning generation
	pipelineSource := "audiotestsrc is-live=true wave=sine samplesperbuffer=960 ! audio/x-raw,format=S16LE,rate=48000,channels=2 ! opusenc bitrate=128000 ! appsink name=appsink max-buffers=8 drop=true sync=false"
	var rate uint32 = 50
	if source.Kind == types.MediaKindVideo {
		pattern := "ball"
		if provider.sceneCuts {
			pattern = "black"
		}
		for _, variant := range FixedVariants() {
			if variant.SourceID == source.ID {
				source.Width, source.Height = variant.Width, variant.Height
				source.FrameRateNumerator, source.FrameRateDenominator = variant.FrameRate, 1
				rate = variant.FrameRate
				pipelineSource = fmt.Sprintf("videotestsrc name=fixture_video is-live=true pattern=%s ! video/x-raw,format=I420,width=%d,height=%d,framerate=%d/1 ! vp8enc deadline=1 cpu-used=8 keyframe-max-dist=%d ! appsink name=appsink max-buffers=8 drop=true sync=false", pattern, variant.Width, variant.Height, rate, rate*2)
			}
		}
	}
	pipeline, err := gst.CreatePipelineWithSampleCapacity(pipelineSource, 64)
	if err != nil {
		return nil, err
	}
	child, cancel := context.WithCancel(ctx)
	subscription := &codecFixtureSubscription{
		source: source, pipeline: pipeline, cancel: cancel,
		events: make(chan types.MediaEvent, 64), done: make(chan struct{}),
	}
	if source.Kind == types.MediaKindVideo {
		cold := source
		cold.Width, cold.Height, cold.FrameRateNumerator, cold.FrameRateDenominator = 0, 0, 0, 0
		subscription.events <- types.MediaEvent{Type: types.MediaEventTypeFormat, Source: cold}
		subscription.events <- types.MediaEvent{
			Type: types.MediaEventTypeDiscontinuity, Source: source,
			// Cold demand's cached reason is source_restart even though caps
			// complete the already-opened generation rather than advancing it.
			Discontinuity: types.MediaDiscontinuity{Generation: source.Generation, Reason: "source_restart"},
		}
	}
	subscription.events <- types.MediaEvent{Type: types.MediaEventTypeFormat, Source: source}
	pipeline.AttachAppsink("appsink")
	go func() {
		defer close(subscription.done)
		defer close(subscription.events)
		var sequence uint64
		var scene int
		for {
			select {
			case <-child.Done():
				return
			case <-pipeline.Dropped():
				select {
				case provider.errors <- errors.New("fixture source handoff overflow"):
				default:
				}
				return
			case sample := <-pipeline.Sample():
				// A nonzero common provider clock exercises late-room startup,
				// videorate's initial gap and encoder segment normalization.
				pts := 30*time.Second + time.Duration(sequence)*time.Second/time.Duration(rate)
				sequence++
				if provider.sceneCuts && source.Kind == types.MediaKindVideo {
					// Hard cuts at 1.3 s deliberately do not follow the 2 s GOP.
					// videotestsrc's black/white enum values are 2/3.
					nextScene := int((pts - 30*time.Second) / (1300 * time.Millisecond))
					if nextScene != scene {
						if !pipeline.SetPropInt("fixture_video", "pattern", 2+nextScene%2) {
							select {
								case provider.errors <- errors.New("fixture scene change failed"):
								default:
								}
								return
							}
							scene = nextScene
						}
				}
				unit := types.EncodedMediaUnit{
					Generation: source.Generation, Sequence: sequence,
					PTS: pts, DTS: pts, PTSValid: true, DTSValid: true,
					Duration: time.Second / time.Duration(rate), Keyframe: !sample.DeltaUnit,
					CapturedAt: provider.origin.Add(pts - 30*time.Second), Data: sample.Data,
				}
				select {
				case subscription.events <- types.MediaEvent{Type: types.MediaEventTypeUnit, Source: source, Unit: unit}:
				case <-child.Done():
					return
				}
			}
		}
	}()
	pipeline.Play()
	return subscription, nil
}

type codecFixtureSubscription struct {
	source    types.MediaSource
	pipeline  gst.Pipeline
	cancel    context.CancelFunc
	events    chan types.MediaEvent
	done      chan struct{}
	closeOnce sync.Once
}

func (subscription *codecFixtureSubscription) ID() string { return subscription.source.ID }
func (subscription *codecFixtureSubscription) Source() types.MediaSource {
	return types.CloneMediaSource(subscription.source)
}
func (subscription *codecFixtureSubscription) Events() <-chan types.MediaEvent {
	return subscription.events
}
func (*codecFixtureSubscription) Switch(context.Context, types.MediaSelector) error {
	return errors.New("fixture has fixed sources")
}
func (*codecFixtureSubscription) SetPaused(bool) error { return errors.New("fixture has no pause") }
func (subscription *codecFixtureSubscription) Close() error {
	subscription.closeOnce.Do(func() {
		subscription.cancel()
		<-subscription.done
		subscription.pipeline.Destroy()
	})
	return nil
}

func TestRealCodecsReachConventionalPackagerReadiness(t *testing.T) {
	if err := validateGSTTranscoderElements(); err != nil {
		t.Fatal(err)
	}
	for _, element := range []string{"audiotestsrc", "videotestsrc", "opusenc", "vp8enc"} {
		if err := gst.CheckElement(element); err != nil {
			t.Fatal(err)
		}
	}
	provider := &codecFixtureProvider{origin: time.Now(), errors: make(chan error, 4)}
	packager := newPackager(provider, gstTranscoderFactory{})
	defer packager.Shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), ConventionalReadyWindow+time.Second)
	defer cancel()
	if err := packager.Acquire(ctx, ModeHLS); err != nil {
		select {
		case fixtureErr := <-provider.errors:
			t.Fatalf("real codec readiness: %v; fixture: %v", err, fixtureErr)
		default:
			t.Fatalf("real VP8/Opus to H.264/AAC packager did not become ready: %v", err)
		}
	}
	defer packager.Release()
	for _, id := range []string{"audio", "high", "medium", "low"} {
		track := packager.track(id)
		track.mu.RLock()
		initURI := track.initURI
		segments := append([]Segment(nil), track.segments...)
		track.mu.RUnlock()
		init, ok := packager.Object(id, initURI)
		if !ok || init.Size() == 0 || len(segments) < 3 {
			t.Fatalf("rendition %s lacks codec init or three conventional parents", id)
		}
		if segment, ok := packager.Object(id, segments[len(segments)-1].URI); !ok || segment.Size() == 0 {
			t.Fatalf("rendition %s lacks retained parent bytes", id)
		}
	}
	if _, err := packager.Master(); err != nil {
		t.Fatal(err)
	}
	packager.mu.Lock()
	generation := packager.generation
	packager.mu.Unlock()
	if generation != 1 {
		t.Fatalf("cold fixture caused %d packager generations", generation)
	}
}

// This test can also be mounted into the prior codec-validation image. Its
// scene cuts expose GOP phase changes that the smooth startup fixture misses.
func TestRealSceneCutsPreservePackagerGeneration(t *testing.T) {
	if err := validateGSTTranscoderElements(); err != nil {
		t.Fatal(err)
	}
	for _, element := range []string{"audiotestsrc", "videotestsrc", "opusenc", "vp8enc"} {
		if err := gst.CheckElement(element); err != nil {
			t.Fatal(err)
		}
	}
	provider := &codecFixtureProvider{origin: time.Now(), errors: make(chan error, 4), sceneCuts: true}
	packager := newPackager(provider, gstTranscoderFactory{})
	defer packager.Shutdown()
	ctx, cancel := context.WithTimeout(context.Background(), ConventionalReadyWindow+2*ParentDuration+4*time.Second)
	defer cancel()
	if err := packager.Acquire(ctx, ModeHLS); err != nil {
		t.Fatalf("scene-cut packager did not become ready: %v", err)
	}
	defer packager.Release()
	initial := make(map[string]uint64)
	for _, id := range []string{"audio", "high", "medium", "low"} {
		track := packager.track(id)
		if track == nil {
			t.Fatalf("scene cuts removed initial rendition %s", id)
		}
		track.mu.RLock()
		if len(track.segments) != 3 {
			track.mu.RUnlock()
			t.Fatalf("scene-cut rendition %s lacks three initial parents", id)
		}
		initial[id] = track.segments[len(track.segments)-1].Sequence
		track.mu.RUnlock()
	}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			t.Fatal("scene-cut packager stopped advancing complete parents")
		case err := <-provider.errors:
			t.Fatal(err)
		case <-ticker.C:
		}
		packager.mu.Lock()
		generation := packager.generation
		packager.mu.Unlock()
		if generation != 1 {
			t.Fatalf("scene cuts restarted packaging: generation=%d", generation)
		}
		advanced := true
		for _, id := range []string{"audio", "high", "medium", "low"} {
			track := packager.track(id)
			if track == nil {
				t.Fatalf("scene cuts removed rendition %s", id)
			}
			track.mu.RLock()
			ready := track.initReady && track.hlsReady && !track.failed && len(track.segments) == 3
			var last Segment
			if len(track.segments) > 0 {
				last = track.segments[len(track.segments)-1]
			}
			track.mu.RUnlock()
			if !ready {
				t.Fatalf("scene cuts made rendition %s unplayable", id)
			}
			if _, ok := packager.Object(id, last.URI); !ok {
				t.Fatalf("rendition %s lost its latest complete parent", id)
			}
			if last.Sequence < initial[id]+2 {
				advanced = false
			}
		}
		if advanced {
			if _, err := packager.Master(); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
}
