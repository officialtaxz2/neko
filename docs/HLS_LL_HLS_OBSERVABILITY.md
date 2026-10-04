# HLS / LL-HLS observability

Phase 4 repository assets, 2026-10-04. Runtime acceptance is **PENDING**; all
project checks are **NOT EXECUTED IN CODEX**. The fixed contract and thresholds
remain in [HLS_LL_HLS.md](HLS_LL_HLS.md).

## Private evidence

On the real server, from the repository root:

```bash
bash deploy/collect-hls-media.sh init ../neko-hls-results
bash deploy/collect-hls-media.sh snapshot ../neko-hls-results no-viewer
bash deploy/collect-hls-media.sh snapshot ../neko-hls-results one-viewer
```

The directory must be outside Git, must not be an ancestor of the repository,
and must be new or already private with mode `0700`. Files are created under
`umask 077`. The collector records UTC, exact Git commit/cleanliness, selected
image/health/restart fields, Docker versions, selected Prometheus samples and
container resource counters. It copies the pending results template once.
It never archives `.env`, full `docker inspect`, HTTP headers or raw logs.

For a changed bind port or path prefix, set `NEKO_METRICS_URL` to the actual
credential-free loopback HTTP `/metrics` URL. No public metrics exposure is
required. HLS metric labels contain no session, lease ID or credential; existing
WebRTC samples can contain peer IDs. Keep the collected files private and share
only a reviewed summary. Do not export a browser HAR or command containing a
real ticket/cookie/media URL into this directory or the conversation.

## Signals and interpretation

| Metric family | Evidence |
| --- | --- |
| `neko_media_hls_leases` | Mode and opening/active/paused lease counts; not proof that a frame is displayed |
| `neko_media_hls_bootstrap_total`, `neko_media_hls_requests_total` | Fixed success/failure outcomes, without request paths |
| `neko_media_hls_requests`, `neko_media_hls_blocked_reloads` | Active media handlers and waiting LL reloads; not the keepalive concurrency limiter's complete internal count |
| `neko_media_hls_packagers`, `neko_media_hls_packager_starts_total` | Shared audio/high/medium/low workers and their lifecycle |
| `neko_media_hls_generations_total`, `neko_media_hls_drops_total` | Discontinuities and bounded pipeline/queue failures |
| `neko_media_hls_objects`, `neko_media_hls_retained_bytes` | Memory-only immutable init/part/segment retention |
| `neko_media_hls_published_bytes_total` | Publication rate; parts and parents overlap, so measure parent segments separately |
| `neko_media_hls_publish_delay_seconds`, `neko_media_hls_request_duration_seconds` | Publication/HTTP delay, neither measures glass-to-glass latency |
| `neko_media_*`, `neko_capture_*`, `neko_websocket_*`, `neko_webrtc_*` | Shared subscriptions, capture demand and healthy-viewer isolation |
| `process_*`, `go_*`, `docker stats` | Process CPU/RSS, Go heap/goroutines and container resources |

No scrape system is required for snapshots. If Prometheus is already available,
these queries can accompany the five-minute phases:

```promql
sum(neko_media_hls_leases) by (mode, state)
sum(neko_media_hls_packagers) by (variant, state)
sum(neko_media_hls_retained_bytes)
sum(rate(neko_media_hls_drops_total[5m])) by (variant, stage, reason)
8 * sum(rate(neko_media_hls_published_bytes_total{kind="segment"}[5m])) by (variant)
histogram_quantile(0.95, sum(rate(neko_media_hls_request_duration_seconds_bucket[5m])) by (le, mode, resource))
100 * rate(process_cpu_seconds_total[5m])
process_resident_memory_bytes
sum(neko_media_subscriptions_active{backend="hls"}) by (kind)
```

Process CPU above is percent of one CPU core; normalize by host CPU count for
the total-host 80% threshold. Container CPU, Go heap and process RSS have
different meanings. Snapshot differences alone do not establish sustained CPU
or request p95; sample throughout each phase and record the measurement method.

Run no-viewer, one-viewer, three-same-rendition, three-variant and slow-viewer
phases for five minutes each, with healthy WebRTC/WebCodecs viewers present.
Additional viewers must not add packager workers. Retained objects must remain
within 64 MiB; provider queue capacities remain 64 and worker handoff remains
eight. After the final **unpaused lease**, allow the 15-second idle grace for
workers/subscriptions/retention to reach zero. A paused authenticated lease may
remain and renew without keeping the packager alive. After closing every lease,
lease/request counts must also reach zero. Keep the fixed CPU/RSS, latency and
isolation gates; record omissions rather than silently declaring acceptance.
