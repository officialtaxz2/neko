# HLS startup repair and unresolved login incident — 2026-10-04

Status: source changes statically reviewed on `testing`; the exact 80020d99
target automated/image gate and both real-codec integration tests passed.
Default-off deployment and normal browser smoke checks passed; enabled HLS
live acceptance remains pending. Tests/builds/codec
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

The extracted helper from a7669dd184138ba53ea9398d31d0ca706ee0b400 completed
baseline deployment at application 80020d99 with Baseline-Exitcode 0. Its blob
was c6f52dc80fdf605ec908f3fe3856ce23e015e494. The prepared image
my-neko/brave:hls-80020d99477a started healthy without HLS; both disabled-route
probes passed. The operator confirmed normal login, picture, audio and control
work in the requested private browser check. The saved rollback image remains
available. This is supplied target evidence, NOT EXECUTED IN CODEX; neither
enabled HLS playback nor the original timeout's exact cause is established.

Same-image HLS activation then completed with Enable-Exitcode 0: the image
started healthy, all 17 public denial probes and both cleartext-denial probes
passed. The operator again reported HLS fails, with the fixed message
`HLS bootstrap failed; retry manually`. They tentatively recalled a brief
picture before the failure, then reported all streams stopped. Device/browser,
actual process crash versus stalled media, timestamps and cause remain
unconfirmed; this is not a verified successful HLS playback or crash trace.

The next step at that checkpoint was to preserve a read-only diagnostic before restarting, using
the existing helper whose blob is 5b4b064926f83eccc28fe8bd596db93ba9c19ff1.
Its raw application/Supervisor logs stay in a new private report directory;
only its safe summary may be shared. Then invoke the reviewed a7669dd1 helper's
baseline action to restore this same previously working image without HLS.
Diagnostic capture and restoration subsequently passed, as recorded below.
Review that evidence before making further application changes. Later
repeat the picture/audio/room-event smoke test, valid passive authorization/lifecycle,
mixed-backend isolation, resources and grouped device checks in
[`HLS_LL_HLS_VALIDATION.md`](HLS_LL_HLS_VALIDATION.md). No automatic fallback,
encoder bitrate/profile change, new production dependency or `master` promotion
is included. Existing dependency advisories remain open.

## Second diagnosis/recovery and fixed-GOP follow-up

The supplied 80020d99 diagnostic and recovery both ended with exit 0. The
container was running/healthy with image ID
sha256:63d7441365a8ec8628a0abeec8c11f5046ee86962769f9723bf12e276a75c299,
zero Docker restarts and oom_killed=false; no Neko exit appeared in the bounded
Supervisor sample. These are observations, not proof against every historical
hang/crash. The 220-line application sample contains one ready marker, one
opened/changed/closed lease, two sample rejections and two generation starts.
The two rejected samples matched `HLS transcode timeline gap`. Cumulative
metrics show one successful bootstrap, 23 successful segment requests and five
not-ready bootstraps; synthetic denial probes are also included. Three video
keyframe-admission counters were high=201, medium=318 and low=235. This proves
valid server delivery occurred, not uninterrupted visual/audio playback.

The same prepared image was restored healthy without HLS and passed both
disabled-route probes, with Recovery-Exitcode 0. Fresh browser confirmation
after this restoration is pending; earlier default-off browser evidence remains
valid for its tested interval. Runtime gst-inspect remains unavailable, which
does not establish missing codec libraries.

The inspected HLS encoder sets a maximum two-second GOP but permits scene-cut
keyframes. These can shift subsequent GOP boundaries while acceptSample still
requires IDRs at every even one-second part boundary. It can drop delta frames
through a whole part and then reject the later keyframe as a timeline gap. This
is a source contract defect and a matching failure hypothesis, not proof that
it explains the operator's all-stream outage. See the
[GStreamer key-int-max/option-string contract](https://gstreamer.freedesktop.org/documentation/x264/index.html#x264enc:key-int-max)
and [x264's fixed-GOP discussion](https://mailman.videolan.org/pipermail/x264-devel/2017-August/012296.html).

The bounded correction adds scenecut=0 only to the HLS x264 option string. It
preserves encoder rates, geometry, profiles, timestamp mapping, genuine gap
rejection and WebRTC/WebCodecs capture. The target-only real-codec regression
introduces hard black/white cuts every 1.3 input seconds and requires all four
renditions to become ready and advance two more complete parents in generation
1. The prior smooth-ball startup fixture remains separate. Pattern enum values
are documented by [GStreamer](https://gstreamer.freedesktop.org/documentation/videotestsrc/index.html#GstVideoTestSrcPattern).

The next step was [validate-hls-gop-repair.sh](../deploy/validate-hls-gop-repair.sh) against
the existing codec-validation image, with application checkout still 80020d99.
The helper pins its image ID, verifies old transcoder/packager/GStreamer source
blobs, mounts the new test into an isolated no-network container, and requires
the old code to fail with the timeline-gap marker. An unrelated failure or
unexpected pass stops the gate. It then mounts only the corrected transcoder
and requires all three codec integration tests to pass. It keeps new timestamped
private reports without changing the service, checkout or preparation marker.
Its supplied outcome and subsequent full repair checks/build are recorded below.
Deployment and live/device acceptance remain **PENDING; NOT EXECUTED IN CODEX**.
The all-stream symptom and shared capture/room-event isolation still require
explicit target verification before a full acceptance claim.

## Isolated GOP A/B gate passed — 2026-10-05

The supplied comparison ran at unchanged application checkout
80020d99477a58318f210b7e14d19cdd92991a6d, with repair code
97ba4ad9ab3e635da936a58c8a7ec795da05ba46 and helper blob
712b491bd5f9900488555051a1fe9acd705f51dc. The existing codec image was pinned to
sha256:ed572e4ef4cb4dbbf94b56027f4ac19ac47837a74aebe7a4b14860282a5b4a7f;
its four inspected source blobs matched the prior application. Private evidence
is in ../neko-hls-results-80020d99477a/gop-repair-20261005T080229723870874Z.

The negative control returned the expected exit 1: the old encoder reproduced
two low-rendition `HLS transcode timeline gap` rejections, restarted into
generations 2 and 3 and failed conventional readiness after 24.01 seconds.
This is the expected regression failure, not a failed overall gate.

With only the corrected transcoder additionally mounted, the positive control
returned 0. Encoder segment/running-time mapping passed in 0.01 seconds,
smooth startup passed in 18.11 seconds and sustained scene cuts passed in
30.19 seconds. The latter reached readiness in generation 1, retained all four
renditions and advanced two additional complete parents without a generation
restart. The mediahls package completed in 48.505 seconds. GOP-Check-Exitcode
was 0. Cleanup sample-discard warnings followed pipeline removal in the test;
they did not fail the gate or establish deployed capture loss.

This confirms the scene-cut defect and correction in the bounded real-codec
fixture. It does not establish real capture skew, room-event isolation,
production browser playback or the unique cause of the reported all-stream
outage. No new application image was built or deployed by this comparison;
live HLS remains disabled and fresh post-restoration browser confirmation is
still pending. These are supplied target results, **NOT EXECUTED IN CODEX**.

The next step was to fast-forward the clean target checkout to 97ba4ad9 and run
validate-hls-phase4.sh with a new private ../neko-hls-results-97ba4ad9ab3e
directory. Preserve the 80020d99 evidence and rollback image. Review the full
checks/build, three fresh-image integration tests and uniquely tagged images
before default-off deployment/browser confirmation and a separate enabled HLS
checkpoint. That preparation subsequently passed, as recorded below.

## Exact GOP-repair automated/image gate passed — 2026-10-05

The operator confirmed a clean target testing checkout at
97ba4ad9ab3e635da936a58c8a7ec795da05ba46. The supplied validation run then
completed with Repair-Check-Exitcode 0 and AUTOMATED/IMAGE GATE PASSED. Evidence
is in ../neko-hls-results-97ba4ad9ab3e; the success marker names the same full
application commit. This checkpoint did not replace the live service.

All 47 client tests, TypeScript checking and Vite production build passed. The
13 configured Go packages passed, followed by successful 30-second fuzz jobs:
mediaws 1,447,749 executions and mediahls 621,533 executions. The trailing
server/plugin build and codec-validation image build passed. GStreamer 1.26.2
was reported. All three integration tests passed: timestamp mapping 0.01
seconds, smooth readiness 18.11 seconds and sustained scene cuts 30.19 seconds,
with the scene-cut fixture remaining in generation 1. The mediahls integration
package completed in 48.500 seconds. These are bounded fixture results, not
production capture/device or cross-backend isolation acceptance.

Both uniquely tagged images were built successfully:
my-neko/base:hls-97ba4ad9ab3e and my-neko/brave:hls-97ba4ad9ab3e. Cached layers
were used. The helper retained image IDs in private images.txt and wrote the
success marker after its final snapshot. The dependency audit returned 1;
npm ci reported 20 findings (11 low, 3 moderate, 5 high, 1 critical). Findings
remain open; no advisory identity/reachability or remediation claim follows
from these aggregate counts. These are supplied target results,
**NOT EXECUTED IN CODEX**.

The next step was to use the same checkout and new private directory with the reviewed
deploy-hls-media.sh baseline action (blob
c6f52dc80fdf605ec908f3fe3856ce23e015e494), then run both disabled-route probes.
The helper verifies preparation/image identity, saves the currently running
image for rollback and starts the prepared repair image without HLS while
preserving adaptive/WebCodecs overlays. This restarts active sessions; startup
health failure attempts the saved prior image. The supplied deployment result
and browser response are recorded below.

## Exact GOP-repair default-off deployment/browser checkpoint passed — 2026-10-05

At unchanged application 97ba4ad9ab3e635da936a58c8a7ec795da05ba46, the tracked
deployer recorded blob c6f52dc80fdf605ec908f3fe3856ce23e015e494 and completed
baseline mode with Baseline-Exitcode 0. The prepared
my-neko/brave:hls-97ba4ad9ab3e image started healthy; both public disabled-route
probes returned the expected 404 (2/2). The helper retained the new private
directory and recorded its snapshot. These are supplied target results,
**NOT EXECUTED IN CODEX**.

The operator answered the requested fresh normal login/picture/audio/control
check with "ja klappt alles soweit ich denke". This is bounded reported browser
evidence on the new image; no detailed role/recovery/device matrix or enabled
HLS playback is implied. The next step was to enable conventional HLS on this same image and
repeat its 17 public plus two direct cleartext-denial probes, then the bounded
actual picture/audio/room-event checkpoint. Preserve evidence, Caddy and master. Full
enabled HLS, isolation, authorization/lifecycle and device acceptance remain
open; a healthy container and two disabled routes do not establish playback.

## GOP-repair live attempt failed; diagnosis/recovery pending — 2026-10-05

After the confirmed default-off browser checkpoint, the operator reported
`HLS bootstrap failed; retry manually` during the requested HLS attempt. The
operator then confirmed that normal WebRTC also stopped working. The latest
enablement CLI output, HTTP probe outcomes and running image/configuration have
not yet been supplied; do not infer a passed activation gate or a process crash.
The fixed-GOP fixture repair remains verified only within its recorded tests.
It has not resolved the live failure, whose cause remains unconfirmed.

NEXT capture the existing read-only diagnosis before any restart, preserving
private raw logs and sharing only the fixed safe summary. Then use baseline
mode at unchanged application 97ba4ad9 with the existing private result directory
to restore the prepared image without HLS. This is the previously confirmed
same-image baseline; the saved rollback tag instead holds the prior image.
Continue recovery even if diagnosis fails, reporting both exit codes. Review
the disabled-route probes and fresh normal browser behavior before further
HLS attempts. Diagnosis, restoration and browser recovery are pending.
No new build, checkout change or Caddy modification is required for this block.
The operator report is target evidence; runtime work is **NOT EXECUTED IN CODEX**.
