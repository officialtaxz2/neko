package mediahls

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

// The second Samples call proves the prior pump has moved past its initial
// output. The repaired pump keeps that output until high establishes anchor.
type startupObservedTranscoder struct {
	*fakeTranscoder
	first       chan struct{}
	reentered   chan struct{}
	firstOnce   sync.Once
	sampleCalls int
}

func (encoder *startupObservedTranscoder) Samples() <-chan types.Sample {
	encoder.sampleCalls++ // only the output pump calls this method
	if encoder.sampleCalls == 2 {
		close(encoder.reentered)
	}
	return encoder.samples
}

func (encoder *startupObservedTranscoder) CapturedAt(sample types.Sample) time.Time {
	encoder.firstOnce.Do(func() { close(encoder.first) })
	return sample.Timestamp
}

func startupWorker(t *testing.T, id string) (*packagerWorker, *startupObservedTranscoder) {
	t.Helper()
	var variant Variant
	for _, candidate := range FixedVariants() {
		if candidate.ID == id {
			variant = candidate
			break
		}
	}
	if variant.ID == "" {
		t.Fatalf("unknown fixture rendition %s", id)
	}
	encoder := &startupObservedTranscoder{
		fakeTranscoder: &fakeTranscoder{
			variant: variant, samples: make(chan types.Sample, WorkerHandoffCapacity),
			drops: make(chan struct{}, 1),
		},
		first: make(chan struct{}), reentered: make(chan struct{}),
	}
	track := &trackState{
		id: id, variant: variant, generation: 1, timescale: 90_000,
		baseMSN: 1, nextMSN: 1, partIndex: -1,
		muxer: newFragmentMuxer(90_000), lastOutput: time.Now(),
	}
	return &packagerWorker{track: track, transcoder: encoder}, encoder
}

func startupUnit(second int) types.EncodedMediaUnit {
	pts := 30*time.Second + time.Duration(second)*time.Second
	return types.EncodedMediaUnit{
		Generation: 1, Sequence: uint64(second + 1),
		PTS: pts, DTS: pts, PTSValid: true, DTSValid: true,
		Duration: time.Second, Keyframe: second%2 == 0,
		CapturedAt: time.Unix(0, 0).Add(pts), Data: []byte{byte(second + 1)},
	}
}

func startStartupPump(t *testing.T, packager *Packager, worker *packagerWorker) (context.CancelFunc, <-chan struct{}) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		packager.pumpOutput(ctx, worker)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
			packager.Shutdown()
		case <-time.After(time.Second):
			t.Error("startup output pump did not stop on cancellation")
		}
	})
	return cancel, done
}

func waitStartupSignal(t *testing.T, signal <-chan struct{}) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(time.Second):
		t.Fatal("startup fixture signal timed out")
	}
}

func TestPackagerRetainsInitialSamplesUntilAnchor(t *testing.T) {
	packager := newPackager(newFakeHLSProvider(), fakeTranscoderFactory{})
	medium, encoder := startupWorker(t, "medium")
	high, highEncoder := startupWorker(t, "high")
	packager.tracks["medium"], packager.tracks["high"] = medium.track, high.track
	startStartupPump(t, packager, medium)
	encoder.Push(startupUnit(0))
	waitStartupSignal(t, encoder.first)
	packager.signalUpdate() // an unrelated publication wake is not an anchor
	// Delay high after observing medium's initial IDR. Old code confirms it
	// consumed/discarded that IDR by entering its next select. New code waits.
	select {
	case <-encoder.reentered:
	case <-time.After(100 * time.Millisecond):
	}
	highEncoder.Push(startupUnit(0))
	if err := packager.acceptSample(high.track, <-highEncoder.samples); err != nil {
		t.Fatal(err)
	}
	for second := 1; second <= 18; second++ {
		encoder.Push(startupUnit(second))
	}
	// Completion of MSN 3 is the progress barrier. It is the third parent for
	// retained initial output and only the second for the prior lost-IDR path.
	waitHLSCondition(t, func() bool {
		medium.track.mu.RLock()
		defer medium.track.mu.RUnlock()
		return medium.track.nextMSN >= 4
	})
	medium.track.mu.RLock()
	defer medium.track.mu.RUnlock()
	if !medium.track.hlsReady || len(medium.track.segments) != 3 || medium.track.segments[0].Sequence != 1 {
		t.Fatal("initial pre-anchor IDR was lost; three parents were not ready at 18 seconds")
	}
}

func TestPackagerAnchorWaitCancels(t *testing.T) {
	packager := newPackager(newFakeHLSProvider(), fakeTranscoderFactory{})
	worker, encoder := startupWorker(t, "medium")
	packager.tracks["medium"] = worker.track
	cancel, done := startStartupPump(t, packager, worker)
	encoder.Push(startupUnit(0))
	waitStartupSignal(t, encoder.first)
	cancel()
	waitStartupSignal(t, done)
	worker.track.mu.RLock()
	defer worker.track.mu.RUnlock()
	if worker.track.initReady || worker.track.dtsSet {
		t.Fatal("unanchored cancelled sample was published")
	}
	select {
	case <-packager.restart:
		t.Fatal("cancellation requested a generation restart")
	default:
	}
}

func TestPackagerAnchorWaitHonorsOverflow(t *testing.T) {
	for _, closed := range []bool{false, true} {
		name := "drop"
		if closed {
			name = "closed"
		}
		t.Run(name, func(t *testing.T) {
			packager := newPackager(newFakeHLSProvider(), fakeTranscoderFactory{})
			worker, encoder := startupWorker(t, "medium")
			packager.tracks["medium"] = worker.track
			_, done := startStartupPump(t, packager, worker)
			encoder.Push(startupUnit(0))
			waitStartupSignal(t, encoder.first)
			if closed {
				close(encoder.drops)
			} else {
				encoder.drops <- struct{}{}
			}
			waitStartupSignal(t, done)
			select {
			case reason := <-packager.restart:
				if reason != "worker_failure" {
					t.Fatalf("overflow restart reason = %s", reason)
				}
			default:
				t.Fatal("unanchored overflow did not request a restart")
			}
		})
	}
}
