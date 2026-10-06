# HLS rolling-playlist and seek-only progress repair — 2026-10-07

Status: **IMPLEMENTED / STATICALLY REVIEWED / TARGET CHECKS PENDING**.
The live application remains exact `7dcc3c5e4ba0279f35727eb4bed2e86bf721ad3b`
with conventional HLS enabled. No new activation or runtime verification was
performed in Codex. Preserve that live image and all prior private evidence.

## Supplied five-minute browser trace

The operator supplied the complete `hls-browser-trace-v1` JSON from the existing
PC/Helium page. Its reason is `five_minute_limit`, duration 300,001 ms; these
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
the difference between audio and video. A repaired live result remains pending.

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
preparation. New preparation/tests/images are **PENDING / NOT EXECUTED IN CODEX**.

After a supplied successful preparation, use the existing exact-candidate
deployment procedure, then one bounded live check past several 24-second window
advances with moving picture/audio and concurrent WebRTC. Request further
diagnostics only if that path still fails. Reliable cold start, room actions,
authorization/lifecycle, TV/native/LL-HLS, resource/isolation and dependency
maintenance remain open. `master` stays at `d9105ef8`.
