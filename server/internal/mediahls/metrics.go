package mediahls

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/m1k1o/neko/server/pkg/types"
)

var (
	hlsLeases = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "leases", Namespace: "neko", Subsystem: "media_hls", Help: "Current HLS playback leases by mode and bounded state."}, []string{"mode", "state"})
	hlsBootstrap = promauto.NewCounterVec(prometheus.CounterOpts{Name: "bootstrap_total", Namespace: "neko", Subsystem: "media_hls", Help: "HLS bootstrap attempts by mode and bounded result."}, []string{"mode", "result"})
	hlsRequests = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "requests", Namespace: "neko", Subsystem: "media_hls", Help: "Current HLS HTTP requests by resource and bounded state."}, []string{"resource", "state"})
	hlsRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{Name: "requests_total", Namespace: "neko", Subsystem: "media_hls", Help: "HLS HTTP requests by mode, resource and bounded result."}, []string{"mode", "resource", "result"})
	hlsRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{Name: "request_duration_seconds", Namespace: "neko", Subsystem: "media_hls", Help: "HLS HTTP request duration by mode and resource.", Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 7, 15, 30}}, []string{"mode", "resource"})
	hlsBlockedReloads = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "blocked_reloads", Namespace: "neko", Subsystem: "media_hls", Help: "Current LL-HLS blocking reloads by bounded state."}, []string{"state"})
	hlsPackagers = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "packagers", Namespace: "neko", Subsystem: "media_hls", Help: "Current shared HLS packager workers by rendition and bounded state."}, []string{"variant", "state"})
	hlsPackagerStarts = promauto.NewCounterVec(prometheus.CounterOpts{Name: "packager_starts_total", Namespace: "neko", Subsystem: "media_hls", Help: "HLS packager worker starts by rendition and bounded result."}, []string{"variant", "result"})
	hlsGenerations = promauto.NewCounterVec(prometheus.CounterOpts{Name: "generations_total", Namespace: "neko", Subsystem: "media_hls", Help: "HLS generation transitions by rendition and bounded reason."}, []string{"variant", "reason"})
	hlsObjects = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "objects", Namespace: "neko", Subsystem: "media_hls", Help: "Retained immutable HLS objects by rendition and kind."}, []string{"variant", "kind"})
	hlsRetainedBytes = promauto.NewGaugeVec(prometheus.GaugeOpts{Name: "retained_bytes", Namespace: "neko", Subsystem: "media_hls", Help: "Retained immutable HLS bytes by rendition and kind."}, []string{"variant", "kind"})
	hlsPublishedBytes = promauto.NewCounterVec(prometheus.CounterOpts{Name: "published_bytes_total", Namespace: "neko", Subsystem: "media_hls", Help: "Published HLS bytes by rendition and kind."}, []string{"variant", "kind"})
	hlsPublishDelay = promauto.NewHistogramVec(prometheus.HistogramOpts{Name: "publish_delay_seconds", Namespace: "neko", Subsystem: "media_hls", Help: "Delay from encoded media capture to HLS object publication.", Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 1.5, 2, 3, 6, 12}}, []string{"variant", "kind"})
	hlsDrops = promauto.NewCounterVec(prometheus.CounterOpts{Name: "drops_total", Namespace: "neko", Subsystem: "media_hls", Help: "HLS bounded drops by rendition, stage, kind and reason."}, []string{"variant", "stage", "kind", "reason"})
)

func metricMode(value string) string {
	if value == ModeLLHLS {
		return ModeLLHLS
	}
	return ModeHLS
}

func metricVariant(value string) string {
	switch value {
	case "audio", "high", "medium", "low":
		return value
	default:
		return "unknown"
	}
}

func metricResource(value ResourceKind) string {
	switch value {
	case ResourceMaster, ResourcePlaylist, ResourceInit, ResourceSegment, ResourcePart, ResourceKeepAlive:
		return string(value)
	default:
		return "bootstrap"
	}
}

func metricObjectKind(kind ObjectKind) string {
	switch kind {
	case ObjectInit:
		return "init"
	case ObjectPart:
		return "part"
	case ObjectSegment:
		return "segment"
	default:
		return "unknown"
	}
}

func metricMediaKind(kind types.MediaKind) string {
	switch kind {
	case types.MediaKindAudio, types.MediaKindVideo:
		return string(kind)
	default:
		return "unknown"
	}
}
