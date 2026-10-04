# HLS startup repair and unresolved login incident — 2026-10-04

Status: source changes statically reviewed on `testing`. Tests, builds, codec
execution and deployment of these changes are **NOT EXECUTED IN CODEX** and
**PENDING on the target server**. No working-playback or login-repair claim.

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
if login still hangs after service restoration.

## Verification and recovery order

First use the existing private rollback image and
[`deploy-hls-media.sh rollback`](../deploy/deploy-hls-media.sh) from the unchanged
93f1fa63 checkout. It recreates the service without the HLS overlay, retaining
the existing adaptive/WebCodecs deployment and reviewed Caddy configuration.
Active sessions reconnect. Preserve all earlier private evidence. Record the
configured rollback image separately from the source checkout; this restores
the earlier runtime, not the latest containment repairs. It is an incident
baseline, not a security or final acceptance checkpoint. Rollback and recovered
WebRTC login/playback are **PENDING**, not assumed to work.

Only after reviewing that result, advance the source checkout to the reviewed
repair commit and use a fresh `../neko-hls-results-<12-commit-characters>`
directory. The exact-commit preparation helper now includes:

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

After preparation passes, verify the new image with HLS disabled and working
normal WebRTC login/media first. Then enable conventional HLS and repeat the
picture/audio/room-event smoke test, valid passive authorization/lifecycle,
mixed-backend isolation, resources and grouped device checks in
[`HLS_LL_HLS_VALIDATION.md`](HLS_LL_HLS_VALIDATION.md). No automatic fallback,
encoder bitrate/profile change, new production dependency or `master` promotion
is included. Existing dependency advisories remain open.
