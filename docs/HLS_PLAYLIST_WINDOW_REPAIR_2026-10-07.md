# HLS rolling-playlist and seek-only progress repair — 2026-10-07

Status: **EXACT-8741 PREPARATION, ACTIVATION AND BOUNDED FIVE-MINUTE PLAYBACK PASSED / GROUPED ACCEPTANCE OPEN**.
Target checkout and live deployment are
`8741f7880d9a709e6dc17924d2af0564a949ccd9`, conventional HLS enabled.
The operator confirmed moving HLS picture/audio for at least five minutes
without Retry/reload, with WebRTC working alongside it. This is bounded
PC/Helium acceptance, not full Phase 4 or a device-wide stability claim.
All runtime/browser results are supplied, **NOT EXECUTED IN CODEX**.
Preserve the previous exact-7dcc image and all prior private evidence.

## Supplied exact-7dcc five-minute browser trace

The operator supplied the complete `hls-browser-trace-v1` JSON from the old
exact-7dcc PC/Helium page. Its reason is `five_minute_limit`, duration 300,001 ms; these
are supplied browser observations, **NOT EXECUTED IN CODEX**. The attachment
initially truncated the JSON; the existing report was subsequently copied in
full without requesting another test, Retry or reload.

| Observation | Result |
| --- | --- |
| Initial state | Old terminal HTTP-connection detail until about 8 seconds |
| Manual Retry/bootstrap | Started 8,625 ms; completed after 18,182 ms, with body metadata |
| Playback start | `playing` at 27,149 ms; moving samples from 28 seconds |
| Freeze | `waiting` at 51,099 ms; media time 23.948 seconds, 513 frames |
| Buffered audio/video intersection | Fixed at 12.008–24.019 seconds throughout the later freeze |
| Later media state | Unpaused, no media error; readyState 1, frames fixed at 513 |
| Apparent time progress | Seeks every six seconds, from 30 to 270 seconds; no later `seeked`/`canplay`/`playing` |
| Audio requests | Two child playlists, four segments, one init |
| Video requests | Low: two segments; high: two; medium: 44 with 45 child playlists |
| Watch traffic | 269 masters, 17 keepalives; normal fast completions in retained tail |
| Bounded output | 392 request completions; 180 retained HTTP rows, 212 discarded; no discarded media/event/slow rows |
| Long requests | Only the 18.182-second bootstrap reached the 900-ms slow threshold |
| UI after Retry | Compact HLS through the end; no new terminal HTTP failure captured |

The elapsed five minutes are a **captured playback failure**, not successful
five-minute acceptance. Bodies/playlist parser errors are not captured. A zero
Resource Timing status with body metadata is not proof of HTTP failure. This
trace does not reproduce the earlier later-terminal HTTP error or correlate a
WebRTC participant/room action with the freeze.

## Confirmed source defects and bounded repairs

`server/internal/mediahls/packager.go` marks the first complete segment of each
generation with `EXT-X-DISCONTINUITY`, retaining three complete parents. At the
fourth parent (24 seconds), eviction removed that first tag without advancing
the track's `EXT-X-DISCONTINUITY-SEQUENCE`. Consequently overlapping segment
URIs changed discontinuity numbers across reloads. This violates
[RFC 8216 §6.2.2](https://www.rfc-editor.org/rfc/rfc8216.html#section-6.2.2).
The repair increments the **track-local** base for each removed discontinuity
tag before trimming; removal of ordinary segments does not advance it. It
preserves the global generation counter and existing window/codec/retention
policies. The rule applies to complete-parent windows in both playlist modes.

Pinned hls.js 1.7.3's
[overlap check](https://github.com/video-dev/hls.js/blob/v1.7.3/src/utils/level-helper.ts)
rejects changed continuity counters. Its
[base playlist controller](https://github.com/video-dev/hls.js/blob/v1.7.3/src/controller/base-playlist-controller.ts)
returns on that parsing error before scheduling the next reload. This is a
strong explanation for audio stopping after four segments and the buffer
ending around 24 seconds, while new video variants still load. The supplied
trace does **not** expose the parser error or prove which ABR transition caused
the difference between audio and video. The repaired five-minute live gate below
passed; the uncaptured parser error remains an inferred cause of the old freeze.

The client watchdog separately treated every changed `currentTime` as playback
progress. Captured seek jumps indefinitely hid the frozen buffer/frames. The
repair accepts only forward movement with current data while not seeking,
updates the baseline even on excluded samples, and records the seek target
without refreshing the progress deadline. This also covers a seek that completes
between watchdog samples. Healthy forward playback with readyState 2 remains
accepted; pause/resume, startup/HTTP budgets and manual backend choice stay
unchanged. No automatic Retry or transport fallback is introduced.

## Target preparation and next acceptance

[`deploy/prepare-hls-playlist-window.sh`](../deploy/prepare-hls-playlist-window.sh)
is a separate target-only helper. Supply the repository, exact-7dcc private
evidence directory, a **new** private output directory and a full candidate SHA.
It verifies the clean testing history, bounded changed-file scope and live
image against the baseline preparation record. In isolated candidate copies it:

1. Requires the old client's seek-only stall assertion to fail, then runs the
   repaired full client suite, TypeScript check and build.
2. Builds the candidate server validation image, requires the old packager's
   retained-segment continuity assertion to fail, then checks the repaired
   wire playlists through six complete parents for all four tracks in both
   modes and runs the full HLS package suite without cached test results.
3. Fast-forwards the target checkout only after those checks, records private
   evidence, builds candidate base/Brave tags and verifies the same live
   container/image is still running. Only full success writes the validation
   marker used by the existing deployment helper.

The new tests exercise wire-level overlap and actual controller behavior with
bounded deterministic doubles; they do not replace a browser/media gate.
The LL-HLS check covers rolling complete-parent windows, not startup partial
segments, blocking reloads or native/LL-HLS browser acceptance.
Source codecs/transcoders, security configuration, dependencies and runtime
configuration are unchanged. Prior exact-8f native/fuzz evidence is inherited
where supplied, **not freshly rerun or relabelled as candidate results**.
No production container, Caddy configuration or active session is restarted by
preparation. The exact-8741 target preparation below has now passed; none of
those tests/builds were executed in Codex.

## Supplied exact-8741 preparation — 2026-10-07

The operator supplied the preparation output with **Prepare-Exitcode: 0**.

| Gate | Supplied result |
| --- | --- |
| Old client control | Seek-only jumps hide the controlled stall, as required |
| Repaired client | 66 tests passed; zero failures/skips; TypeScript and build passed |
| Old server control | Retained-segment discontinuity changes at the fourth parent, as required |
| Repaired wire check | `TestPackagerRollingPlaylistPreservesDiscontinuityNumbers` passed for all four tracks/both modes over six simulated parents |
| Full HLS package | Uncached `go test ./internal/mediahls -count=1` passed |
| Application checkout | Fast-forwarded from exact-7dcc to exact-8741 |
| Candidate images | Server validation, `my-neko/base:hls-8741f7880d9a` and `my-neko/brave:hls-8741f7880d9a` built |
| Client image bundle | `index-DPO2ZBEO.js`; fresh image client build shown |
| Live service | Same exact-7dcc container/image retained through preparation |
| Private evidence | `/opt/docker/nekoNew/neko-hls-results-8741f7880d9a`; snapshot and successful validation marker recorded |

These controlled results confirm the old defects and the repaired automated
paths. They do not prove that the browser freeze is fixed or close TV/native/
LL-HLS acceptance. There is no fresh codec integration or fuzz result in this
block. npm's dependency audit warnings remain part of open dependency maintenance;
no package update or audit remediation is claimed.

## Supplied exact-8741 activation and playback — 2026-10-07

The operator supplied the completed activation block below with
**Activate-Exitcode: 0**. It is an execution record; do not repeat it.
The existing deployer checked the prepared image and recorded private init/enable
evidence. Keep the target checkout at **exact-8741**; later documentation commits
are not a replacement image checkpoint. No additional pull or build is needed.

```bash
set +e
bash -Ee -o pipefail <<'NEKO_HLS_ACTIVATE_REPAIR'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "8741f7880d9a709e6dc17924d2af0564a949ccd9"

umask 077
output=/opt/docker/nekoNew/neko-hls-results-8741f7880d9a
test "$(stat -c %a "$output")" = 700
bash deploy/deploy-hls-media.sh enable "$output"
NEKO_HLS_ACTIVATE_REPAIR
printf 'Activate-Exitcode: %s\n' "$?"
```

| Gate | Supplied result |
| --- | --- |
| Application commit | `8741f7880d9a709e6dc17924d2af0564a949ccd9` |
| Deployment helper blob | `c6f52dc80fdf605ec908f3fe3856ce23e015e494` |
| Live image/service | `my-neko/brave:hls-8741f7880d9a`; healthy after recreation |
| Private evidence | Init and enable snapshot in `/opt/docker/nekoNew/neko-hls-results-8741f7880d9a` |
| Initial browser report | Playback worked without frozen pictures |
| Bounded browser interval | Operator explicitly confirmed at least five minutes of moving HLS picture/audio without Retry or reload |
| Concurrent WebRTC | Operator confirmed WebRTC continued working throughout that interval |

This closes the requested bounded conventional-HLS sustained-playback gate on
the existing PC/Helium setup. The duration and audio/WebRTC behavior are operator
observations, not a new automated browser trace or server measurement. No repeated
cold-start distribution, room-action sequence, native player or TV result was
supplied. The positive repaired run supports the repair; it does not expose the
old hls.js parser error or prove a WebRTC-join cause. No further diagnosis is
required for this passed interval.

## Remaining acceptance

Continue the [grouped Phase 4 plan](HLS_LL_HLS_VALIDATION.md) with the passed
exact-8741 preparation, activation and bounded browser interval recorded. Do not
repeat these gates without a new failure, source change or other concrete reason.
No additional server command or ad-hoc operator check is requested now.

- Repeated cold-start reliability and the existing startup latency target.
- A documented room-action/reconnect interval and authorization/lease lifecycle.
- TV/native/LL-HLS compatibility and grouped device/resource/isolation checks.
- Open dependency maintenance and final integrated security/stability acceptance.

Keep conventional HLS opt-in, WebRTC the default and backend switching manual.
Preserve prior images/evidence. `master` stays at `d9105ef8` until explicitly
authorized promotion after the remaining grouped acceptance.
