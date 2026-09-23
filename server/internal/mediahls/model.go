package mediahls

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/m1k1o/neko/server/pkg/types/message"
)

const (
	BackendName   = "hls"
	BootstrapPath = "/api/media/hls/session"
	ProtocolVersion uint8 = 1

	ModeHLS   = "hls"
	ModeLLHLS = "ll-hls"

	MaximumLeases           = 128
	MaximumRequests         = 512
	MaximumLeaseRequests    = 4
	MaximumLeaseBlocking    = 2
	MaximumBlockingRequests = 64
	MaximumBootstrapBody    = 2 * 1024
	MaximumPlaylistBytes    = 64 * 1024
	MaximumInitBytes        = 2 * 1024 * 1024
	MaximumPartBytes        = 1 * 1024 * 1024
	MaximumSegmentBytes     = 8 * 1024 * 1024
	MaximumRetainedBytes    = 64 * 1024 * 1024
	MaximumRetainedParents  = 6
	MaximumRetainedParts    = 42
	MaximumRetainedInits    = 2

	BootstrapCreatesPerMinute = 2
	BootstrapCreateBurst      = 2
	KeepAlivesPerMinute       = 6
	KeepAliveBurst            = 2
	PlaylistsPerSecond        = 12
	PlaylistBurst             = 24
	ObjectsPerSecond          = 32
	ObjectBurst               = 64
)

const (
	LeaseLifetime      = 30 * time.Second
	BlockingReloadWait = 7 * time.Second
	HeaderReadTimeout  = 5 * time.Second
	ResponseWriteTimeout = 30 * time.Second
)

var (
	ErrInvalidConfig = errors.New("invalid HLS configuration")
	ErrInvalidMode   = errors.New("invalid HLS mode")
)

type Config struct {
	AllowedOrigins []string
	TrustedProxies []string
	Modes          []string
	MaxLeases      int
	MaxRequests    int
}

type Variant struct {
	ID               string
	SourceID         string
	Width            uint32
	Height           uint32
	FrameRate        uint32
	AverageBandwidth uint64
	Bandwidth        uint64
}

var fixedVariants = []Variant{
	{ID: "high", SourceID: "high", Width: 1280, Height: 720, FrameRate: 25, AverageBandwidth: 3_128_000, Bandwidth: 4_000_000},
	{ID: "medium", SourceID: "medium", Width: 854, Height: 480, FrameRate: 20, AverageBandwidth: 1_228_000, Bandwidth: 1_500_000},
	{ID: "low", SourceID: "low", Width: 640, Height: 360, FrameRate: 15, AverageBandwidth: 493_000, Bandwidth: 650_000},
}

func FixedVariants() []Variant {
	return slices.Clone(fixedVariants)
}

func NormalizeConfig(config Config) (Config, error) {
	if config.MaxLeases == 0 {
		config.MaxLeases = MaximumLeases
	}
	if config.MaxRequests == 0 {
		config.MaxRequests = MaximumRequests
	}
	if config.MaxLeases < 1 || config.MaxLeases > MaximumLeases {
		return Config{}, fmt.Errorf("%w: max leases must be between 1 and %d", ErrInvalidConfig, MaximumLeases)
	}
	if config.MaxRequests < 1 || config.MaxRequests > MaximumRequests {
		return Config{}, fmt.Errorf("%w: max requests must be between 1 and %d", ErrInvalidConfig, MaximumRequests)
	}
	if len(config.AllowedOrigins) == 0 {
		return Config{}, fmt.Errorf("%w: at least one exact HTTPS origin is required", ErrInvalidConfig)
	}
	modes := make([]string, 0, len(config.Modes))
	seen := map[string]struct{}{}
	for _, mode := range config.Modes {
		mode = strings.TrimSpace(mode)
		if !ValidMode(mode) {
			return Config{}, fmt.Errorf("%w: unsupported mode %q", ErrInvalidConfig, mode)
		}
		if _, ok := seen[mode]; ok {
			return Config{}, fmt.Errorf("%w: duplicate mode %q", ErrInvalidConfig, mode)
		}
		seen[mode] = struct{}{}
		modes = append(modes, mode)
	}
	if len(modes) == 0 {
		return Config{}, fmt.Errorf("%w: at least one mode is required", ErrInvalidConfig)
	}
	config.Modes = modes
	if _, err := NewSecurityPolicy(config.AllowedOrigins, config.TrustedProxies); err != nil {
		return Config{}, err
	}
	return config, nil
}

func ValidMode(mode string) bool {
	return mode == ModeHLS || mode == ModeLLHLS
}

func capabilityVariants() []message.MediaHLSVariant {
	result := make([]message.MediaHLSVariant, 0, len(fixedVariants))
	for _, variant := range fixedVariants {
		result = append(result, message.MediaHLSVariant{
			ID: variant.ID, SourceID: variant.SourceID,
			VideoCodec: "avc1.64001f", Container: "video/mp4",
			Width: variant.Width, Height: variant.Height, FrameRate: variant.FrameRate,
			AverageBandwidth: variant.AverageBandwidth, Bandwidth: variant.Bandwidth,
		})
	}
	return result
}

func capabilityLimits(config Config) message.MediaHLSLimits {
	return message.MediaHLSLimits{
		MaxLeases: config.MaxLeases, MaxRequests: config.MaxRequests,
		MaxRequestsPerLease: MaximumLeaseRequests,
		MaxBlockingPerLease: MaximumLeaseBlocking,
		MaxBlockingRequests: MaximumBlockingRequests,
		IdleExpiresInMS: int(LeaseLifetime.Milliseconds()),
		BlockingReloadWaitMS: int(BlockingReloadWait.Milliseconds()),
		MaximumPlaylistBytes: MaximumPlaylistBytes,
	}
}
