package webrtc

import "testing"

func TestEstimatedBitrateSupportsUpgrade(t *testing.T) {
	tests := []struct {
		name          string
		targetBitrate int
		streamBitrate uint64
		threshold     float64
		want          bool
	}{
		{
			name:          "default threshold accepts more than fifteen percent headroom",
			targetBitrate: 1_160_000,
			streamBitrate: 1_000_000,
			threshold:     0.15,
			want:          true,
		},
		{
			name:          "two-to-one tiers reject insufficient next-tier capacity",
			targetBitrate: 1_300_000,
			streamBitrate: 998_400,
			threshold:     1.30,
			want:          false,
		},
		{
			name:          "two-to-one tiers accept next-tier capacity with headroom",
			targetBitrate: 2_400_000,
			streamBitrate: 998_400,
			threshold:     1.30,
			want:          true,
		},
		{
			name:          "negative threshold disables the check",
			targetBitrate: 0,
			streamBitrate: 1_000_000,
			threshold:     -1,
			want:          true,
		},
		{
			name:          "zero stream bitrate is not actionable",
			targetBitrate: 1_000_000,
			streamBitrate: 0,
			threshold:     0.15,
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := estimatedBitrateSupportsUpgrade(tt.targetBitrate, tt.streamBitrate, tt.threshold); got != tt.want {
				t.Fatalf("estimatedBitrateSupportsUpgrade(%d, %d, %v) = %v, want %v", tt.targetBitrate, tt.streamBitrate, tt.threshold, got, tt.want)
			}
		})
	}
}
