//go:build hlsintegration && hlsdiagnostic

package mediahls

import (
	"context"
	"errors"
	"testing"
	"time"
)

// This is an isolated diagnosis of the unchanged 71a14d21 packager, not a
// positive acceptance test. The artificial source phases exercise a condition
// omitted by the aligned fixture; they are not measured target capture skew.
// Keep the extra tag out of normal codec-image/full-preparation checks.
func TestDiagnosticSkewedVideoKeyframesBlockReadiness(t *testing.T) {
	if fixedVariants[1].SourceID == fixedVariants[0].SourceID {
		t.Skip("expected-defect fixture is only for the pinned pre-repair independent-source packager")
	}
	if err := validateGSTTranscoderElements(); err != nil {
		t.Fatal(err)
	}
	provider := &codecFixtureProvider{
		origin: time.Now(), errors: make(chan error, 4),
		clockOffsets: map[string]time.Duration{
			"high": 800 * time.Millisecond,
			"medium": 50 * time.Millisecond,
			"low": 100 * time.Millisecond,
		},
	}
	diagnostics := &codecFixtureDiagnostics{tracks: make(map[string]*codecFixtureTrace)}
	packager := newPackager(provider, diagnostics)
	diagnostics.packager = packager
	defer packager.Shutdown()
	defer diagnostics.log(t)
	ctx, cancel := context.WithTimeout(context.Background(), ConventionalReadyWindow+time.Second)
	defer cancel()
	if err := packager.Acquire(ctx, ModeHLS); !errors.Is(err, ErrPackagerNotReady) {
		if err == nil {
			packager.Release()
		}
		t.Fatalf("skew diagnostic expected not-ready, got %v", err)
	}
	select {
	case fixtureErr := <-provider.errors:
		t.Fatalf("skew diagnostic source failed: %v", fixtureErr)
	default:
	}
	packager.mu.Lock()
	generation := packager.generation
	anchor := packager.anchorPTS
	packager.mu.Unlock()
	if generation != 1 {
		t.Fatalf("skew diagnostic restarted into generation %d", generation)
	}
	for _, id := range []string{"audio", "high", "medium", "low"} {
		track := packager.track(id)
		if track == nil {
			t.Fatalf("skew diagnostic track %s missing", id)
		}
		track.mu.RLock()
		initReady, ready, failed := track.initReady, track.hlsReady, track.failed
		parents, parts, partIndex := len(track.segments), len(track.parts), track.partIndex
		track.mu.RUnlock()
		if !initReady || failed {
			t.Fatalf("skew diagnostic track %s init=%t failed=%t", id, initReady, failed)
		}
		if id == "audio" || id == "high" {
			if !ready || parents != 3 {
				t.Fatalf("skew diagnostic healthy track %s ready=%t parents=%d", id, ready, parents)
			}
		} else if ready || parents != 0 || parts != 0 || partIndex != -1 {
			t.Fatalf("skew diagnostic blocked track %s ready=%t parents=%d parts=%d index=%d", id, ready, parents, parts, partIndex)
		}
		diagnostics.mu.Lock()
		trace := diagnostics.tracks[id]
		diagnostics.mu.Unlock()
		if trace == nil {
			t.Fatalf("skew diagnostic trace %s missing", id)
		}
		trace.mu.Lock()
		outputs, rejected, keyframes := trace.outputs, trace.rejected, len(trace.keyframes)
		observedKeyframes := append([]codecFixtureOutput(nil), trace.keyframes...)
		trace.mu.Unlock()
		if outputs < 100 || rejected != 0 || (id != "audio" && keyframes != 4) {
			t.Fatalf("skew diagnostic track %s outputs=%d rejected=%d keyframes=%d", id, outputs, rejected, keyframes)
		}
		if id == "medium" || id == "low" {
			for index, frame := range observedKeyframes {
				elapsed := frame.DTS - anchor
				if !frame.PTSValid || !frame.DTSValid || !frame.Keyframe ||
					(index == 0 && elapsed >= 0) ||
					(index > 0 && (elapsed <= 0 || int64(elapsed/PartDuration)%2 != 1)) {
					t.Fatalf("skew diagnostic track %s IDR %d does not show the pre-anchor/odd-bucket phase: %+v", id, index, frame)
				}
			}
		}
	}
	t.Log("PHASE_DIAGNOSTIC reproduced not-ready: audio/high ready; medium/low initial admission blocked despite flowing IDRs; generation=1")
}
