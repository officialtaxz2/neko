package capture

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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
