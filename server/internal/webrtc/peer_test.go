package webrtc

import "testing"

func TestEstimatedBitrateSupportsUpgrade(t *testing.T) {
	tests := []struct {
		name             string
		targetBitrate    int
		referenceBitrate uint64
		threshold        float64
		want             bool
	}{
		{
			name:             "default threshold accepts more than fifteen percent headroom",
			targetBitrate:    1_160_000,
			referenceBitrate: 1_000_000,
			threshold:        0.15,
			want:             true,
		},
		{
			name:             "configured medium tier rejects insufficient estimate",
			targetBitrate:    613_076,
			referenceBitrate: 748_800,
			threshold:        0.15,
			want:             false,
		},
		{
			name:             "configured medium tier accepts sufficient estimate",
			targetBitrate:    950_000,
			referenceBitrate: 748_800,
			threshold:        0.15,
			want:             true,
		},
		{
			name:             "negative threshold disables the check",
			targetBitrate:    0,
			referenceBitrate: 1_000_000,
			threshold:        -1,
			want:             true,
		},
		{
			name:             "zero reference bitrate is not actionable",
			targetBitrate:    1_000_000,
			referenceBitrate: 0,
			threshold:        0.15,
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := estimatedBitrateSupportsUpgrade(tt.targetBitrate, tt.referenceBitrate, tt.threshold); got != tt.want {
				t.Fatalf("estimatedBitrateSupportsUpgrade(%d, %d, %v) = %v, want %v", tt.targetBitrate, tt.referenceBitrate, tt.threshold, got, tt.want)
			}
		})
	}
}

func TestReferenceBitrateForUpgrade(t *testing.T) {
	if got, want := referenceBitrateForUpgrade(200_000, 332_800), uint64(332_800); got != want {
		t.Fatalf("nominal reference = %d, want %d", got, want)
	}
	if got, want := referenceBitrateForUpgrade(200_000, 0), uint64(200_000); got != want {
		t.Fatalf("measured fallback = %d, want %d", got, want)
	}
}
