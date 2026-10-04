# HLS startup repair and unresolved login incident — 2026-10-04

Status: source changes statically reviewed on `testing`; the exact 80020d99
target automated/image gate and both real-codec integration tests passed.
Deployment/live acceptance of that image remain pending. Tests/builds/codec
execution are supplied target evidence, **NOT EXECUTED IN CODEX**. No live
HLS-playback or unique original-login-cause claim.

## Supplied target evidence

The read-only helper from `d191b8ea` completed with exit code 0 against application
`93f1fa637ae3f14ba41d1bfef39e860d43993ff1` and prepared image
`sha256:d217eccd941810f2436c4e3372ecd796b9e77b7758ecbaf5ed9e0c2a2a7fc4a7`.
It captured 311 application log lines, found the container running/healthy with
zero Docker restarts and `oom_killed=false`, and found no Neko process exit in
the bounded ten-minute Supervisor sample. Loopback metrics responded.

The HLS summary contained one `not_ready` and one `backend_error` bootstrap,
two worker-set starts and one `source_restart` generation transition for each
of audio/high/medium/low. All four running-packager gauges were zero at collection.
No packager-ready or lease-open marker appeared in the sample. No successful
bootstrap/publication was demonstrated. Counters also include earlier synthetic
denial probes; missing metric families are not proof of zero historical work.

`gst-inspect-1.0` was unavailable in the application image; this does not establish
missing GStreamer libraries/plugins or their version. The new codec validation
image installs the tools and reports its version separately.

The operator also reported normal login timing out before media selection, and
tentatively reported `/ws` HTTP 101. If confirmed, that proves the public socket
upgrade, not authentication or session initialization: legacy `/ws` upgrades
before its internal `/api/login` and backend socket initialization. No stack
trace or correlated event trace identifies the post-upgrade blocker yet.
The Permissions-Policy warnings are not evidence of the cause.

## Source findings and bounded changes

1. `Packager.newWorker` retained the source generation sampled before
   `Subscribe`. Starting cold capture increments the generation during
   subscription. Bind the worker/transcoder to the opened subscription's
   source, checking identity, codec and a nonzero generation.
2. The capture provider announces initial video dimensions/rate through a
   same-generation discontinuity followed by complete FORMAT. Its cached cold
   demand reason can be `source_restart`; otherwise it reports `format_change`.
   HLS previously restarted the worker set even before initial admission.
   Accept that transition only before format readiness, with the exact expected
   identity/generation/dimensions/rate, and wait for FORMAT. Genuine restarts
   and incompatible formats still restart.
3. Raw x264 output PTS/DTS can include an encoder offset also reflected in its
   segment; AAC timestamps are not in that shifted buffer-time domain. The HLS
   transcoders now map valid output timestamps through each sample's
   `GstSegment` using `gst_segment_to_running_time`. Missing/out-of-segment
   timestamps remain invalid. Capture, WebRTC and WebCodecs use the existing
   raw-timestamp pipeline factory. The
   [GstSegment API](https://gstreamer.freedesktop.org/documentation/gstreamer/gstsegment.html)
   and [GStreamer 1.26.2 encoder source](https://github.com/GStreamer/gstreamer/blob/1.26.2/subprojects/gst-plugins-base/gst-libs/gst/video/gstvideoencoder.c)
   explain the mapping; actual production timestamps have not been measured.
4. HLS `videorate` uses `skip-to-first=true`, preventing duplication of an
   initial gap when capture starts with an already-running room clock. This
   property is documented by
   [GStreamer](https://gstreamer.freedesktop.org/documentation/videorate/index.html#videorate:skip-to-first).
   Opus input caps explicitly specify the fixed stereo mapping family 0;
   a missing mapping family was not established as the operator's failure.
5. The inherited C logger used unbounded `vsprintf` with a 100-byte stack
   buffer. Replace it with `vsnprintf` into a bounded 1,024-byte buffer.
   This closes a confirmed source defect; the supplied sample contains no
   observed crash attributable to it.

The inspected ordinary event path also has unbounded synchronous WebSocket
writes, and legacy loopback HTTP requests have no timeout. A blocked broadcast
can delay session creation while the member login mutex is held. This is an
additional investigation candidate, **not an established cause** of this
incident. These paths are unchanged in this HLS startup repair; obtain evidence
if normal login hangs again during new-image validation.

## Verification and recovery order

The operator executed
[`deploy-hls-media.sh rollback`](../deploy/deploy-hls-media.sh) from the unchanged
93f1fa63 checkout. Restore-Exitcode was 0; configured image
`my-neko/brave:rollback-hls-93f1fa637ae3` started healthy without the HLS overlay.
The operator explicitly confirmed normal login, picture and audio work again.
The helper preserved adaptive/WebCodecs and did not change Caddy. Private
rollback/init/snapshot evidence remains in the existing 93f1fa63 directory;
the image digest was not printed in the supplied console block.

This proves bounded restoration, not the new startup repair. Image, overlay and
process state changed together, so the exact original login blocker remains
unconfirmed. The earlier runtime does not include the latest containment fixes;
it remains an incident baseline rather than final security/device acceptance.

The exact reviewed repair commit
`80020d99477a58318f210b7e14d19cdd92991a6d` passed preparation with exit 0 in
the private `../neko-hls-results-80020d99477a` directory. Later documentation-
only commits
may be fetched but do not need to move this tested application checkout.
Preparation does not replace the working rollback service. Its baseline
snapshot describes that running prior image, not deployed repair-image behavior.
The exact-commit preparation helper now includes:

- focused cold-generation/initial-caps regression cases in the normal Go suite;
- an isolated `hls-packager-checks` image, derived from the newly built
  same-commit server-validation image with codec plugins/tools installed;
- a real x264 test comparing raw versus segment-mapped PTS/DTS, including
  unchanged capture-factory semantics;
- real VP8/Opus fixture encoders feeding the production H.264/AAC transcoders
  and packager, using a shared nonzero 30-second input clock and a cold-caps
  transition; every rendition must publish init and at least three conventional
  parents in one packager generation within the existing readiness window.

Missing required codecs fail the explicit integration gate; they are not
skipped. The validation helper prepares new uniquely tagged images without
replacing the service and writes its success marker only after all gates.
These tests are synthetic codec integration, not production capture-skew,
proxy/authentication, client playback, resource or device acceptance.

Preparation passed: all 13 Go packages, both fuzz jobs, trailing build, the
visible client build and both base/Brave builds completed. Both real-codec tests
passed with GStreamer 1.26.2; four renditions became conventionally ready in
generation 1 (test duration 18.11 s). Audit exit 1 remains open findings. The
client test count is absent from the supplied tail; no new count is claimed.
The working rollback service was not replaced by preparation.

NEXT use the extracted, reviewed operator helper's baseline action to deploy
the prepared image with HLS disabled, retaining the working prior image for
rollback, and verify normal login/picture/audio/control first. The baseline
helper accepts an explicit repository argument and checks the prepared image
ID. It is statically reviewed, NOT EXECUTED IN CODEX and pending target execution.
After that live baseline passes, enable conventional HLS and repeat the
picture/audio/room-event smoke test, valid passive authorization/lifecycle,
mixed-backend isolation, resources and grouped device checks in
[`HLS_LL_HLS_VALIDATION.md`](HLS_LL_HLS_VALIDATION.md). No automatic fallback,
encoder bitrate/profile change, new production dependency or `master` promotion
is included. Existing dependency advisories remain open.
