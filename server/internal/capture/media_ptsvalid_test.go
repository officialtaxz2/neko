package capture

import (
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

func TestNormalizeUnitPreservesPTSValidity(t *testing.T) {
	startedAt := time.Unix(1_700_000_000, 0)
	sample := types.Sample{
		Timestamp:  startedAt.Add(time.Second),
		Generation: 1,
		PTS:        250 * time.Millisecond,
		PTSValid:   true,
		Duration:   40 * time.Millisecond,
	}
	subscription := &captureMediaSubscription{
		provider:    &captureMediaProvider{startedAt: startedAt},
		timelineGen: 1,
	}
	unit := subscription.normalizeUnitLocked(sample)
	if !unit.PTSValid {
		t.Fatal("valid source PTS was not preserved")
	}

	sample.PTSValid = false
	subscription = &captureMediaSubscription{
		provider:    &captureMediaProvider{startedAt: startedAt},
		timelineGen: 1,
	}
	unit = subscription.normalizeUnitLocked(sample)
	if unit.PTSValid {
		t.Fatal("synthesized PTS was reported as source-valid")
	}
}
