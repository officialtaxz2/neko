package mediaws

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	mediaWebSocketConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "connections",
		Namespace: "neko",
		Subsystem: "media_websocket",
		Help:      "Current experimental media WebSocket connections by lifecycle state.",
	}, []string{"state"})
	mediaWebSocketHandshakes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "handshakes_total",
		Namespace: "neko",
		Subsystem: "media_websocket",
		Help:      "Experimental media WebSocket attachment attempts by bounded result.",
	}, []string{"result"})
	mediaWebSocketRecords = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "records_total",
		Namespace: "neko",
		Subsystem: "media_websocket",
		Help:      "Media WebSocket protocol records written by media kind and record type.",
	}, []string{"kind", "type"})
	mediaWebSocketBytes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "bytes_total",
		Namespace: "neko",
		Subsystem: "media_websocket",
		Help:      "Encoded media payload bytes written to media WebSockets.",
	}, []string{"kind"})
	mediaWebSocketDrops = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "drops_total",
		Namespace: "neko",
		Subsystem: "media_websocket",
		Help:      "Media WebSocket local drops by kind, bounded stage and reason.",
	}, []string{"kind", "stage", "reason"})
	mediaWebSocketResyncs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name:      "resyncs_total",
		Namespace: "neko",
		Subsystem: "media_websocket",
		Help:      "Media WebSocket delivery resynchronizations by bounded reason.",
	}, []string{"reason"})
	mediaWebSocketWriteSeconds = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:      "write_duration_seconds",
		Namespace: "neko",
		Subsystem: "media_websocket",
		Help:      "Media WebSocket record write duration.",
		Buckets:   []float64{0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2},
	})
	mediaWebSocketQueueDepth = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:      "queue_depth",
		Namespace: "neko",
		Subsystem: "media_websocket",
		Help:      "Observed per-delivery server egress record depth.",
		Buckets:   []float64{0, 1, 2, 4, 8, 12, 16, 20, 24, 28},
	})
	mediaWebSocketQueueBytes = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:      "queue_bytes",
		Namespace: "neko",
		Subsystem: "media_websocket",
		Help:      "Observed per-delivery encoded payload bytes queued for egress.",
		Buckets:   []float64{0, 64 * 1024, 256 * 1024, 1024 * 1024, 4 * 1024 * 1024, 8 * 1024 * 1024, 12 * 1024 * 1024, 16 * 1024 * 1024},
	})
	mediaWebSocketClientLagMS = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:      "client_lag_milliseconds",
		Namespace: "neko",
		Subsystem: "media_websocket",
		Help:      "Bounded lag reported by the media WebSocket client.",
		Buckets:   []float64{0, 20, 40, 80, 120, 200, 300, 500, 1000, 2000, 5000, 10000},
	})
)

func metricKind(kind Kind) string {
	switch kind {
	case KindAudio:
		return "audio"
	case KindVideo:
		return "video"
	default:
		return "none"
	}
}

func metricRecordType(recordType RecordType) string {
	switch recordType {
	case RecordFormat:
		return "format"
	case RecordUnit:
		return "unit"
	case RecordDiscontinuity:
		return "discontinuity"
	case RecordEnd:
		return "end"
	default:
		return "invalid"
	}
}
