# Opt-in Adaptive Quality Profile

This document describes the experimental multi-pipeline and per-peer bandwidth-estimator profile for the repository's Brave Compose baseline. The normal `docker-compose.yaml` remains the stable, single-pipeline default. Adaptive quality is enabled only when `docker-compose.adaptive.yaml` is supplied as a second Compose file.

The pipeline values below are reproducible starting points. The estimator timings, encoder constraints and asymmetric upgrade threshold include target-server tuning candidates described below, but the profile remains unaccepted until the affected phases and final recovery pass with recorded measurements. Do not promote the overlay to the default before that acceptance.

## Profile contents

`deploy/adaptive-quality.yaml` defines three VP8 pipelines in required highest-to-lowest order:

| ID | Resolution at 1920×1080 | Frame rate | Nominal encoder target |
| --- | ---: | ---: | ---: |
| `high` | 1920×1080 | 25 fps | 1,996,800 bit/s |
| `medium` | 1280×720 | 20 fps | 748,800 bit/s |
| `low` | 960×540 | 15 fps | 332,800 bit/s |

For other desktop sizes, `medium` scales both dimensions to roughly two thirds and `low` to roughly one half; both expressions produce even dimensions. All pipelines use the same VP8 codec as required by Neko's stream selector. The first ID, `high`, is the initial/default stream.

The overlay mounts the profile read-only at `/etc/neko/adaptive-quality.yaml` and sets `NEKO_CONFIG` to that path. Existing Compose environment values still override corresponding YAML values, so the tracked base configuration continues to supply screen, ICE and deployment settings.

The estimator is active rather than passive, starts at 2.5 Mbit/s and uses explicit timing/threshold values. Debug logging is enabled for the estimator. These values are deliberately kept in one mounted file so a measurement-led tuning change produces a reviewable diff.

The first bounded target-server constraint run on 2026-09-10 isolated the constrained viewer successfully, but took about 30 seconds to leave `high` and cascaded from `medium` to `low` about 15 seconds later. The next run reduced `stalled_duration` from 24 to 8 seconds and increased `downgrade_backoff` from 10 to 30 seconds. It left `high` within 20 seconds and recovered from `low` through `medium` to `high` within 40 seconds, but still cascaded to `low` during the 1.3 Mbit/s phase.

The phase measurements identified encoder overshoot rather than host saturation: the nominal 1,996,800 bit/s `high` stream measured as high as 4,394,400 bit/s, while the nominal 499,200 bit/s `low` stream measured between 702,400 and 1,230,040 bit/s. The profile had limited `max-quantizer` to 20, 24 and 28; libvpx can let those quality floors override rate-control targets. [GStreamer documents 63 as the default worst-quality bound](https://gstreamer.freedesktop.org/documentation/vpx/GstVPXEnc.html?gi-language=c), while the [WebM real-time CBR guidance](https://www.webmproject.org/docs/encoder-parameters/) recommends a 50–63 range and demonstrates 56. Raising `max-quantizer` to 56 brought the unconstrained `high` stream to 1,996,120 bit/s against its 1,996,800 bit/s target. All three viewers stayed smooth on `high`, with zero peer-local audio/video drops.

The constrained rerun rejected encoder tuning alone as sufficient. The healthy viewers remained smooth on `high`, but the constrained viewer oscillated upward while the impairment was unchanged: during the 1.3 Mbit/s phase it moved from `medium` to `low`, back to `medium`, and then to `high`; during the 0.7 Mbit/s phase it spent most of the phase on `medium` before reaching `low`. `Low` was the only constrained tier that became visibly usable in part of the 1.3 Mbit/s run, while returning to an unconstrained path restored smooth playback immediately.

Source tracing explains that oscillation. The original `diff_threshold` both protects the current tier from downgrade and permits an upgrade by comparing the estimate with the **current** tier's measured rate. With approximately 2:1 adjacent targets, the old 0.15 upgrade value needed only 15% spare over the lower tier, not enough capacity for the next tier. The server now has a separate `upgrade_diff_threshold`; its default remains 0.15 for compatibility. Its focused target-server test passed, and a profile value of 1.30 removed upward oscillation under both unchanged constraints.

That stable-selection rerun still failed constrained-viewer usability. At 1.3 Mbit/s the viewer reached `low` after 60 seconds and held it; at 0.7 Mbit/s it remained on `low`; both healthy viewers stayed smooth on `high` with zero peer-local drops. `Low` was usable under the 1.3 Mbit/s constraint but showed repeated freezes and audio loss at 0.7 Mbit/s. The low stream measured 494,296 bit/s and shared the shaped path with 128,400 bit/s of audio, consuming about 623 kbit/s before RTP/UDP/IP overhead and retransmission. The 0.7 Mbit/s shaper consequently stayed at capacity and dropped another 153,969 packets. After restoration playback was immediately smooth, but conservative estimation returned through `medium` within 90 seconds and reached `high` only later.

The final measurement-led candidate budgets transport headroom instead of changing more timers: `medium` is 748,800 bit/s, `low` is 332,800 bit/s, and all tiers allow VP8's full `max-quantizer: 63` range so the quality floor cannot defeat those limits. The widest adjacent nominal target ratio is about 2.67:1, so the profile now uses `upgrade_diff_threshold: 1.75`, requiring an estimate 2.75 times the current measured rate. This should allow prompt unimpeded recovery while preventing `medium` to `high` at 1.3 Mbit/s and `low` to `medium` at 0.7 Mbit/s. It is the last tuning candidate before accepting or explicitly rejecting this profile.

## Before activation

Run these commands only on the real target server. Preserve the currently working image under a rollback tag, then rebuild because the adaptive unit includes server-side bitrate/metrics changes:

```bash
docker image tag my-neko/brave:latest my-neko/brave:pre-adaptive
cd server
go test ./internal/capture -run '^TestSaveSampleBitrateUsesBitsPerSecond$'
go test ./internal/webrtc -run '^TestEstimatedBitrateSupportsUpgrade$'
./build
cd ..
./build my-neko/base:latest -y
./build my-neko/brave:latest -y
```

When the target host deliberately has no Go toolchain, validate through the repository's server image instead:

```bash
VALIDATION_IMAGE="my-neko/server-validation:$(git rev-parse --short=8 HEAD)"
docker build -t "$VALIDATION_IMAGE" ./server
docker run --rm "$VALIDATION_IMAGE" \
  go test ./internal/capture -run '^TestSaveSampleBitrateUsesBitsPerSecond$'
docker run --rm "$VALIDATION_IMAGE" \
  go test ./internal/webrtc -run '^TestEstimatedBitrateSupportsUpgrade$'
./build my-neko/base:latest -y
./build my-neko/brave:latest -y
```

If `.env` sets another `NEKO_IMAGE`, use that image name for the backup and build commands. Keep the rollback tag local and never copy credentials into a tracked file.

Validate the merged Compose model without printing resolved secrets, recreate the service, and inspect its state:

```bash
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml config --quiet
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml up -d --force-recreate
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml ps
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml logs --since=5m neko
```

Required startup evidence:

- the `neko` service is healthy/running and uses the intended locally built image;
- the log contains `preflight complete with config file` and no configuration/GStreamer error;
- joining a viewer starts `high`, shown by `set video` with `video_id=high` and by the metrics below;
- the existing Brave profile, downloads and managed policy still load from their original mounts.

## Observability

With the default loopback HTTP port, inspect the relevant metrics on the target server:

```bash
curl -fsS http://127.0.0.1:8082/metrics \
  | grep -E 'neko_(capture_(streamsink_(bitrate|listeners|bytes)|pipelines_active)|webrtc_(receiver_estimated_target_bitrate|track_dropped_samples_total|video_listeners))'
```

Replace `8082` when `NEKO_HTTP_PORT` has been changed. The important series are:

- `neko_capture_streamsink_bitrate`: measured encoded pipeline rate in bit/s, labeled by `video_id`;
- `neko_webrtc_receiver_estimated_target_bitrate`: estimator target in bit/s, labeled by `session_id`;
- `neko_webrtc_video_listeners`: the selected tier for each session (`1` is active);
- `neko_webrtc_track_dropped_samples_total`: cumulative peer-local queue drops, labeled by `session_id` and `kind`;
- `neko_capture_streamsink_listeners` and `neko_capture_pipelines_active`: demand and active encoder pipelines;
- `neko_capture_streamsink_bytes`: cumulative encoded output, useful for an independent rate calculation.

The bitrate gauge and estimator target deliberately use the same bit/s unit. A completed or closed session can leave cumulative counters in the Prometheus registry; use `neko_webrtc_connection_state == 5` or a current `video_listeners == 1` series to identify active sessions. Evaluate queue-drop counter deltas during a phase, not their lifetime totals.

For switch decisions and their measured inputs, follow the estimator logs:

```bash
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml logs -f neko \
  | grep -E 'got bitrate from estimator|downgraded video stream|upgraded video stream|set video|dropping sample'
```

`got bitrate from estimator` includes `target_bitrate`, measured `stream_bitrate`, their ratio and trend. The trace-only `dropping sample` line may not be visible at the normal log level; the counter is the authoritative drop diagnostic.

## Resource cost

Pipelines are started on demand. If every viewer is on `high`, only that video pipeline is needed. If viewers occupy all three tiers, up to three VP8 capture/scale/encoder pipelines run concurrently. Each tier is configured with up to four encoder threads, so CPU demand can rise sharply; thread count is not a promise of exact core use. Scaling, queues and encoded buffers also add memory pressure.

During a switch, the destination pipeline is started before the listener is moved and the unused source pipeline is stopped afterward. Short-lived encoder overlap is therefore expected. Outbound media bandwidth remains per viewer and should roughly follow that viewer's selected tier plus audio, RTP/RTCP and network overhead.

Record host CPU, memory, load, pipeline gauges and packet/drop behavior while all three tiers are active. Lower `threads`, frame rates or target bitrates only from target-server evidence. If the host itself becomes saturated, slow-peer isolation cannot protect healthy viewers from shared CPU exhaustion.

## Healthy-plus-constrained-viewer acceptance

Use two healthy viewers (`H1`, `H2`) and one independently constrained viewer (`C`). A constraint must apply only to `C`'s receive path. Browser HTTP throttling is not sufficient evidence because WebRTC media normally uses UDP; use a router, VM/network namespace or OS shaper whose scope and actual throughput can be verified. Record the shaper/tool, client IP, browser versions, server commit and image ID.

When combining a rate limit with `netem`, bound and inspect the queue explicitly. The default 1,000-packet limit produced several seconds of backlog and an invalid frozen-viewer result during the first target-server attempt. Verify non-zero shaped traffic, bounded backlog and an unshaped healthy-viewer path from `tc -s` output; record kernel-shaper drops separately from `neko_webrtc_track_dropped_samples_total`, which measures the earlier peer-local track queue.

Take a metrics snapshot at the beginning and end of every phase and keep the estimator log running. Map `H1`, `H2` and `C` to `session_id` values by joining them one at a time and observing the newly active `video_listeners` series.

The repository includes a non-destructive collection helper and result template. It reads only the filtered adaptive metrics and logs and never reads `.env`. Keep its output outside the repository because raw evidence can contain session IDs and host details:

```bash
ADAPTIVE_RESULT_DIR="../neko-adaptive-results-$(date -u +%Y%m%dT%H%M%SZ)"
./deploy/collect-adaptive-quality.sh init "$ADAPTIVE_RESULT_DIR"
```

At every phase boundary, take a named snapshot. Use `baseline-start`, `baseline-end`, `medium-start`, `medium-end`, `low-start`, `low-end`, `restored-start` and `restored-end` so before/after counter deltas remain unambiguous:

```bash
./deploy/collect-adaptive-quality.sh snapshot "$ADAPTIVE_RESULT_DIR" baseline-start
# Run the phase, then:
./deploy/collect-adaptive-quality.sh snapshot "$ADAPTIVE_RESULT_DIR" baseline-end
```

After the final recovery checks, capture the relevant estimator/switch log window and complete `RESULTS.md` in the evidence directory:

```bash
./deploy/collect-adaptive-quality.sh logs "$ADAPTIVE_RESULT_DIR" 60m
```

Set `NEKO_METRICS_URL` for the helper when the metrics endpoint is not `http://127.0.0.1:8082/metrics`. The helper records the running container's configured image and immutable image ID without dumping its environment or resolved Compose configuration. Review and sanitize the completed result before adding a summary to the repository.

1. Start `H1`, wait for video/audio/control, then start `H2` and finally `C`. Leave all paths unconstrained for 60 seconds.
2. Confirm all three sessions have `video_id="high"`, target estimates are being updated, and no healthy-session video-drop counter increases during a further 60-second baseline.
3. Limit only `C` to 1.3 Mbit/s receive bandwidth, 80 ms added round-trip latency and 1% packet loss for 90 seconds. Keep `H1` and `H2` actively displaying the same changing desktop content.
4. Confirm `C` changes from `high` to `medium`; `H1` and `H2` must remain on `high`, remain responsive and show no new peer-local video drops. Correlate the switch with `target_bitrate`, `stream_bitrate` and a downgrade log entry.
5. Tighten only `C` to 0.7 Mbit/s receive bandwidth, 120 ms added round-trip latency and 2% packet loss for 90 seconds.
6. Confirm `C` changes to `low`. Its video-drop counter may increase, but any increase must remain labeled with `C`'s session. `H1` and `H2` must still remain on `high` without a new drop delta or visible latency/freeze.
7. Remove all impairment from `C`, provide at least 5 Mbit/s receive bandwidth, and wait up to 90 seconds.
8. Confirm `C` returns through `medium` to `high`, while `H1` and `H2` never changed tier. Confirm unused `medium` and `low` pipelines return to zero listeners and become inactive.
9. Repeat the complete throttle/restore sequence with the legacy client/protocol path when that path is part of the deployment. It must request automatic selection and exhibit the same peer-local behavior.
10. Exercise refresh/rejoin and a transient network interruption for `C`; recovery must not move or stall either healthy session.

Acceptance requires all of the following:

- only `C` changes quality during impairment and recovers to `high` afterward;
- both healthy viewers remain usable, stay on `high`, and acquire no new peer-local video-drop count;
- estimator and measured stream bitrates are plausible bit/s values rather than differing by an unexplained factor of eight;
- no global capture stall, reconnect storm, GStreamer error or sustained host resource saturation occurs;
- pipeline listener/active gauges return to the expected state after recovery and disconnect;
- exact phase measurements and any deviations are recorded before tuning values are changed.

Use this result table in the server-side record:

| Phase | H1 tier / drop delta | H2 tier / drop delta | C tier / drop delta | C target bit/s | Stream bit/s | Host CPU/RAM | Result/notes |
| --- | --- | --- | --- | ---: | ---: | --- | --- |
| baseline |  |  |  |  |  |  |  |
| C 1.3 Mbit/s |  |  |  |  |  |  |  |
| C 0.7 Mbit/s |  |  |  |  |  |  |  |
| restored |  |  |  |  |  |  |  |

## Rollback

To disable adaptive quality while retaining the current image, tear down the merged model and recreate from the base Compose file only:

```bash
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml down
docker compose -f docker-compose.yaml up -d --force-recreate
docker compose -f docker-compose.yaml config --quiet
```

The recreated service must no longer contain the adaptive config mount or `NEKO_CONFIG`, and it must expose only the default `main` stream. Host profile/download data are bind-mounted and are not removed by these commands.

If rollback of the newly built server image is also required, use the backup tag created before activation:

```bash
NEKO_IMAGE=my-neko/brave:pre-adaptive docker compose -f docker-compose.yaml up -d --force-recreate
```

Adapt the image name if the deployment did not use `my-neko/brave:latest`. After rollback, repeat login, playback, control, profile/policy and healthy-viewer smoke checks.
