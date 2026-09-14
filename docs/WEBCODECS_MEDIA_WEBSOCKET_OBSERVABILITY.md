# WebCodecs/media-WebSocket observability

This document is the credential-safe observability surface for the default-off `webcodecs-ws` receive prototype. It is intentionally separate from the stable deployment. The prototype exports only bounded labels; ticket values, session IDs, client addresses, origins, request URLs and event/control payloads must never become metric labels or dashboard variables.

Use [`../docker-compose.webcodecs-ws.yaml`](../docker-compose.webcodecs-ws.yaml) only for explicit prototype runs. [`../deploy/collect-webcodecs-media.sh`](../deploy/collect-webcodecs-media.sh) captures the fixed metric families and host/container resource snapshots into a private directory outside the repository. Raw logs and metric snapshots can still reveal deployment timing and topology, so commit only a reviewed, sanitized result summary.

## Fixed metric inventory

| Signal | Labels | Purpose |
| --- | --- | --- |
| `neko_media_websocket_connections` | `state` | Current prototype delivery lifecycle counts |
| `neko_media_websocket_handshakes_total` | `result` | Bounded pre-/post-upgrade results |
| `neko_media_websocket_records_total` | `kind`, `type` | Written FORMAT/UNIT/DISCONTINUITY/END records |
| `neko_media_websocket_bytes_total` | `kind` | Encoded audio/video payload bytes, excluding envelope bytes |
| `neko_media_websocket_drops_total` | `kind`, `stage`, `reason` | Provider/egress drops and their bounded cause |
| `neko_media_websocket_resyncs_total` | `reason` | Delivery-local recovery triggers |
| `neko_media_websocket_write_duration_seconds` | none | Socket write-time histogram |
| `neko_media_websocket_queue_depth` | none | Observed egress record-depth histogram |
| `neko_media_websocket_queue_bytes` | none | Observed egress payload-byte histogram |
| `neko_media_websocket_client_lag_milliseconds` | none | Bounded client progress/lag histogram |
| `neko_media_deliveries` and related `neko_media_*` metrics | bounded backend/source/kind/state | Central delivery/subscription cleanup and reuse |
| `neko_capture_pipelines_active`, `neko_capture_streamsink_*` | existing bounded capture labels | Encoder/pipeline reuse and encoded-source byte comparison |
| `process_cpu_seconds_total`, `process_resident_memory_bytes`, `go_goroutines`, `go_memstats_heap_alloc_bytes` | standard process/runtime labels | Resource and leak checks |

Queue and lag metrics are histograms of observations, not per-participant current-value gauges. Use them for distributions and cap evidence together with drop/close counters, lifecycle counts and the test matrix; do not infer a particular participant from an aggregate bucket.

Every upgraded media delivery writes one completion log with bounded `close_code`, internal `close_reason` and `failed` fields. Accepted browser resyncs additionally log bounded `resync_kind` and `resync_reason`; queue reasons distinguish video-compressed, audio-compressed, decoded-audio and AudioWorklet overflow while the Prometheus reason label uses the corresponding fixed `client_*` value. These fields distinguish backpressure causes such as `resync_limit` without recording a ticket, URL, Origin, address, media payload or control payload. The delivery logger also carries the existing session identifier, so collected raw logs remain private evidence and must be sanitized before anything is committed.

## PromQL panels and checks

Connection lifecycle and handshakes:

```promql
sum by (state) (neko_media_websocket_connections)
sum by (result) (increase(neko_media_websocket_handshakes_total[5m]))
sum by (backend, state) (neko_media_deliveries)
```

Traffic and record mix:

```promql
sum by (kind) (rate(neko_media_websocket_bytes_total[5m]) * 8)
sum by (kind, type) (increase(neko_media_websocket_records_total[5m]))
sum by (backend, source_id, kind) (neko_media_subscriptions_active)
```

Backpressure, recovery and write time:

```promql
sum by (kind, stage, reason) (increase(neko_media_websocket_drops_total[5m]))
sum by (reason) (increase(neko_media_websocket_resyncs_total[5m]))
histogram_quantile(0.95, sum by (le) (rate(neko_media_websocket_write_duration_seconds_bucket[5m])))
histogram_quantile(0.95, sum by (le) (rate(neko_media_websocket_queue_depth_bucket[5m])))
histogram_quantile(0.95, sum by (le) (rate(neko_media_websocket_queue_bytes_bucket[5m])))
histogram_quantile(0.95, sum by (le) (rate(neko_media_websocket_client_lag_milliseconds_bucket[5m])))
```

Resource comparison:

```promql
rate(process_cpu_seconds_total[5m]) * 100
process_resident_memory_bytes
go_goroutines
go_memstats_heap_alloc_bytes
sum(neko_capture_pipelines_active)
```

Cleanup deltas and payload accounting for a bounded test window:

```promql
sum by (backend) (increase(neko_media_delivery_opens_total[5m]))
  -
sum by (backend) (increase(neko_media_delivery_closes_total[5m]))
sum(increase(neko_media_websocket_bytes_total[5m]))
sum(increase(neko_capture_streamsink_bytes[5m]))
```

The first cleanup query is the open-minus-close delta for each backend; a nonzero value can be an intentionally still-open delivery, so compare it with the lifecycle gauge and the known viewer count. The payload comparison is meaningful only when the selected source and observation window are aligned and no unrelated listener changes occur. The WebSocket byte counter intentionally excludes the 64-byte record envelope and lifecycle/control JSON, so record deltas must be used to add the documented overhead before applying the 10% acceptance bound.

## Snapshot workflow

Run from a clean exact `testing` commit on the real target server:

```bash
RESULT_DIR="../neko-webcodecs-results-$(date -u +%Y%m%dT%H%M%SZ)"
./deploy/collect-webcodecs-media.sh init "$RESULT_DIR"
./deploy/collect-webcodecs-media.sh snapshot "$RESULT_DIR" baseline-start
# Run one bounded phase.
./deploy/collect-webcodecs-media.sh snapshot "$RESULT_DIR" baseline-end
./deploy/collect-webcodecs-media.sh logs "$RESULT_DIR" 30m
```

Do not paste or archive `.env`, cookies, request headers, browser storage, media WebSocket URLs, event payloads or packet captures containing the one-time subprotocol ticket. Use [`WEBCODECS_MEDIA_WEBSOCKET_RESULTS_TEMPLATE.md`](WEBCODECS_MEDIA_WEBSOCKET_RESULTS_TEMPLATE.md) for the reviewed record.
