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

### Bounded progress sampling

`deploy/diagnose-hls-playback.sh REPOSITORY OUTPUT_DIR --progress` adds an
optional read-only live-progress sample to the existing two-argument helper.
Use the reviewed helper extracted outside Git while retaining the exact
prepared application checkout. It checks the preparation marker and running
image ID against the private prepared image record before sampling. No service
restart, build, proxy change or credentialed media request is made.

Two loopback metrics scrapes 12 seconds apart distinguish capture bytes/HLS
source units, published media bytes, HTTP results, source generations, new
capture pipelines and HLS generations/discontinuities. Only fixed enumerated
labels, numeric values and fixed log markers are printed; raw logs/metrics
remain private. The selected WebRTC-open counts are aggregate, without peer IDs.
Counter deltas require `delta_valid: true`: successful scrapes, matching process
start/container metadata and no observed counter regression. Missing series
alone are not zero production; counter reset or missing process metadata
invalidates deltas. Multiple viewers can contribute HTTP counts, and server
publication/fetching does not demonstrate browser decoding. A reload loses the
earlier browser state; if a freeze recurs, sample before Retry/reload when
possible. Do not infer a join-caused freeze from cumulative event counts.

The [exact-7dcc moving-picture/freeze record](HLS_CLIENT_STABILITY_REVIEW_2026-10-05.md#moving-first-picture-then-freeze--2026-10-07)
requested this step. The supplied `02526dca` run passed syntax/diagnostic checks,
confirmed the exact-7dcc live image and valid 12.012-second deltas, but captured
idle HLS: no leases/subscriptions/workers or interval publication/HTTP events.
The operator confirmed HLS was closed/logged out. This is expected idle cleanup,
not evidence of an active freeze. The subsequent corrected `4593a6f9` progress
sample passed with valid 12.011-second deltas at the same exact-7dcc image:
one active lease, both subscriptions/all workers, positive source-unit/media
publication and successful HTTP requests, with no interval generation/
discontinuity/drop/capture-creation/WebRTC-open increment. The operator confirms
this sample was taken during frozen picture/audio loss without reload, then a
later "HLS HTTP connection failed; retry manually". The later HTTP failure was
outside this interval; server flow does not prove successful browser decoding.
The supplied `inspect-hls-browser-timing.js` result shows one 1,001-ms master
without response metadata, but a full 250-entry buffer omits later failures.
Status-zero entries can contain body data; terminal cleanup cleared the current
video. The complete five-minute `trace-hls-browser.js` result is now supplied:
392 observed HTTP completions, 180 retained/212 discarded; no lost media/event/
slow rows. It captures a fixed 513-frame/24-second buffer after initial playback,
continued medium-video loading, audio stopped after four segments and six-second
seek jumps. Only bootstrap exceeded the 900-ms slow threshold; no new terminal
HTTP error is captured. Exact-8741 preparation then passed with Prepare-Exitcode 0,
both old-defect reproductions, 66 client tests/type/build, repaired wire/HLS-package
checks and candidate images. Private evidence is
`/opt/docker/nekoNew/neko-hls-results-8741f7880d9a`; checkout is exact-8741 and
live exact-7dcc was retained. NEXT [activate the prepared image](HLS_PLAYLIST_WINDOW_REPAIR_2026-10-07.md#next-activate-the-prepared-image)
and check sustained HLS picture/audio alongside WebRTC. Supplied target evidence,
NOT EXECUTED IN CODEX; activation/live acceptance remain pending. No new build/trace.

The progress and saved-startup label parsers were statically corrected to admit
digits before their unchanged fixed allowlists, retaining the `vp8` codec that
was previously dropped. Prior accepted HLS/lease/source rows remain valid;
omitted VP8 values in the earlier summary are unknown. The corrected progress
parser's target gate passed and now includes VP8 rows, with helper blob
`60cf47ddf8489c2397a5770a4c9523ae09a50b56`. The corrected saved-startup parser
has no fresh target execution. Supplied target results, **NOT EXECUTED IN CODEX**.

### Historical worker/startup signals

`HLS worker restart requested` now records only fixed variant, generation,
stage and reason fields before the existing worker_failure request. Input
push failure, output overflow/closure, anchor-wait overflow/closure and audio
output stall are distinguished without media bytes, credentials or new metric
labels. Each requesting pump/monitor returns; the log is bounded by worker
exits in each generation. The private playback diagnostic counts this marker;
the saved-startup summary exports only allowlisted variant/stage/reason values.
An absent marker is not proof that a worker never failed.

The isolated [worker startup diagnostic](../deploy/diagnose-hls-worker-startup.sh)
uses the rebuilt 414 codec image and retains up to three observations per track
(the first two replaced workers and the latest). Three fresh processes include
failed checks in their output; diagnostic exit 0 does not pass preparation.
The deliberate anchor-overflow lifecycle test also produces restart markers:
interpret them within the corresponding `=== RUN` / test-result interval.
Keep live HLS disabled until the failed full-preparation gate is resolved.
The supplied 53034495 run passed all six checks in three fresh processes,
with smooth readiness at 18.11 s in generation 1 each; it did not reproduce
the earlier worker failure. The [controlled audio-anchor A/B](../deploy/validate-hls-audio-anchor.sh)
at repair 71a14d21 subsequently passed with Audio-Anchor-Exitcode 0. Identical
observed fixtures on both sides reproduced the old blocked-drainage unit and
a real audio/anchor/queue_full restart under 256 ms injected high-input delay.
Its positive side passed eight checks in each of three fresh processes plus
scene cuts once (25 top-level passes); all seven real-codec fixtures stayed in
generation 1 with zero rejected pushes. Normal readiness was 18.10/18.11/18.11 s,
delayed-high readiness 18.09 s each and scene cuts 30.09 s. The production repair
drains AAC during startup instead of holding its first sample; only video waits
for high. AAC admitted after the anchor keeps its actual presentation offset.
The injected delay is a test condition, not a deployment setting; this
controlled reproduction does not identify the earlier unobserved restart or
live all-stream outage. The A/B changed no preparation marker or service.
Full exact-71a14d21 preparation then passed with Repair-Prepare-Exitcode 0:
all nine selected startup checks passed in the rebuilt GStreamer 1.26.2 image,
with normal/delayed-high/scene-cut fixtures in generation 1, and base/Brave
images were built. The service stayed unchanged. Its supplied tail omits
earlier client/Go/fuzz output; preceding stages are covered by the script's
reported final success, not fresh per-check counts in this excerpt. The private
snapshot/success marker is under /opt/docker/nekoNew/neko-hls-results-71a14d2174da.
Dependency-audit exit 1 remains an open report item, not security acceptance.
Default-off deployment of my-neko/brave:hls-71a14d2174da then passed with
Baseline-Exitcode 0, healthy service, a private baseline snapshot and 2/2
disabled bootstrap/media probes returning 404. The operator confirmed the
requested normal-browser check works without HLS. Same-image conventional HLS
activation then passed with Enable-Exitcode 0, healthy service, a private enable
snapshot and 17/17 public plus 2/2 cleartext-denial probes. The subsequent live
HLS attempt failed after connecting with "HLS bootstrap failed; retry manually";
the operator reported only WebRTC streaming works. Read-only diagnosis passed:
950 log lines, one not-ready bootstrap and one started/idle-stopped generation,
759/564 medium/low keyframe-admission drops, and cumulative part/segment bytes
only for audio/high. Current zero object/running gauges follow idle teardown;
they do not mean publication never occurred. No sampled exit/OOM was found.
Same-image default-off restoration passed, with a healthy service and 2/2
disabled-route probes; the operator confirmed normal login/picture/audio.
The isolated clock-phase diagnosis then reproduced the defect in all three
cold runs at 24.02 seconds, while its aligned control passed at 18.11 seconds.
The common high-source fan-out repair A/B then passed at a7ffb8b1 with
Shared-Clock-Exitcode 0: old defect reproduced once, all fifteen positive
checks passed three cold runs plus one scene-cut check (46 positive top-level
passes). All ten codec fixtures stayed in generation 1; normal/delayed/skewed
fixture durations were 18.08/18.07/18.83 seconds each run, scene cuts 30.07 seconds.
Full exact-a7ffb8b1 preparation then passed with Repair-Prepare-Exitcode 0:
thirteen selected native/startup checks passed in the rebuilt GStreamer 1.26.2
image; all four codec fixtures stayed in generation 1 (normal/delayed/skewed
18.08/18.07/18.83 seconds, scene cuts 30.08 seconds), and base/Brave images were
built. The supplied tail omits earlier client/Go/fuzz output; final script
success covers those stages without separately shown fresh counts. The running
default-off 71 service stayed unchanged. The new private snapshot/success marker
is under `/opt/docker/nekoNew/neko-hls-results-a7ffb8b13448`.
Default-off exact-repair deployment then passed with Baseline-Exitcode 0,
healthy service, a private baseline snapshot and 2/2 disabled-route probes;
the operator reported the requested normal browser check works. Same-image
activation then passed with Enable-Exitcode 0 and all nineteen denial probes.
The operator reported first HLS picture and the compact streaming label, then
the initial-readiness error around 30 seconds; retry restored HLS and WebRTC
kept working. The supplied read-only diagnosis and isolated client gate then
passed: old deadline fault reproduced and all 52 repaired tests/type/build
passed, with the live service unchanged. Scoped exact-73d5ff6d preparation then
passed with Client-Image-Exitcode 0: fresh client bundle `index-CrHQRMnq.js`,
cached unchanged server/runtime layers, both base/Brave images and private
snapshot/marker under `/opt/docker/nekoNew/neko-hls-results-73d5ff6d2911`.
Inherited backend evidence is separate from the passed exact client gate.
Exact-73 default-off deployment then passed healthy with Baseline-Exitcode 0,
a private baseline snapshot and 2/2 disabled-route probes. The operator confirmed
the requested normal browser check works. Same-image activation then passed
healthy with Enable-Exitcode 0, private enable snapshot and 19/19 denial probes.
On PC/Helium the operator reports an initial Retry and one frozen-picture reload,
then working HLS while WebRTC continued working. "HLS failed" was confirmed
without its detailed error/timing; the roughly 20–30-second lag is not a measured
latency gate. Checkout/live image is now 73d5ff6d with HLS enabled; the saved
prior a7ff image remains under the 73 rollback tag. Existing safe counters can
narrow the server stages, but cannot reconstruct the erased browser error.
Read-only diagnosis subsequently passed with Diagnostic-Exitcode 0: healthy
image ID `sha256:cf8913d83d2e6fd2b1e5ca3e2af5dbf984f5432ed65722ee738a20775cf1b0cb`,
987 sampled application lines, zero restarts/OOM/sample exits/fixed error markers,
one active HLS lease/four running workers, six successful bootstraps and 394
successful segments. One bootstrap was not-ready; the exact browser attempt
and elapsed readiness time are not correlated. Two startup-labelled generations,
two ready markers and one idle stop are consistent with another cold cycle.
Counts are cumulative across participants/retries, not a five-minute playback
or audible-audio result. The operator cannot confirm that interval and mentions
possible random reconnects/room actions; no further ad-hoc check is requested now.
NEXT [consolidated startup/frozen-picture/room-event investigation](HLS_CLIENT_READINESS_REPAIR_2026-10-05.md#next-consolidate-startup-frozen-picture-and-room-event-investigation).
The actual
live IDR phases remain unmeasured. The repair owns one audio/high subscription
and four workers; its three high-resolution decoders require a fresh CPU/RSS
comparison. **NOT EXECUTED IN CODEX; supplied reproduction and isolated repair
A/B, full preparation, default-off deployment/browser and enabled CLI/HTTP
passed; sustained enabled playback acceptance pending.**

`HLS playback did not become ready; retry manually` is the client player's
30-second initial-readiness deadline, distinct from bootstrap rejection and
the ongoing `HLS playback stalled; retry manually` progress watchdog. The old
client failed to cancel its initial deadline after `canplay`/`playing`, so a
briefly low readyState at the deadline could end already-started playback.
Repair 73d5ff6d cancels that deadline on current-player readiness; its target
client gate passed, while new images/deployment remain pending. Live media counters alone cannot prove a
browser buffer state. Capture the safe summary without restarting a working
WebRTC service; raw logs stay private and cumulative retries are not correlated
to one browser failure.

The latest supplied a7ff diagnosis returned exit 0: healthy image
`sha256:e3517c887622e04065a7fec5fae1902c03e2e4e5470946915b9b99457e2992b2`, zero
restarts, OOM false and no sampled Neko exits/fixed error markers in 977 log
lines. Two HLS bootstraps succeeded; all four tracks published parts/parents
in startup generation 1. Counts include 486 successful segments, 479 playlists,
1,408 masters, 92 keepalives and seven init requests. Lease-close/idle-stop
markers and zero current objects/leases/packagers are consistent with later
idle cleanup, not evidence of a crash. Earlier not-ready/rejection counters
remain cumulative. This proves server delivery within its scope, not sustained
browser playback, latency or a failure-free device matrix.

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
