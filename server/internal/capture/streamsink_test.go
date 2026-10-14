package capture

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

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
