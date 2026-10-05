# HLS startup repair and unresolved login incident — 2026-10-04

Status: source changes statically reviewed on `testing`; the exact 80020d99
target automated/image gate and both real-codec integration tests passed.
Full preparation at 414639d2 FAILED with a worker_failure restart during
smooth cold readiness; bounded stage/history diagnosis then passed three
fresh starts without reproducing it. The narrower video-only anchor hold at
71a14d21 passed the controlled AAC-overflow A/B: both old-code failures
reproduced, all eight corrected checks passed three fresh processes and scene
cuts passed once. Full exact-71a14d21 preparation subsequently passed with
Repair-Prepare-Exitcode 0: all nine selected startup checks passed in the rebuilt
codec image and base/Brave images were built. The running service was unchanged.
Default-off 71 deployment then passed with Baseline-Exitcode 0, healthy service
and 2/2 disabled-route probes. The operator confirmed normal browser behavior
without HLS. Same-image 71 conventional-HLS activation then passed with
Enable-Exitcode 0, healthy service and 19/19 denial probes. The live HLS attempt
then failed with bootstrap failure; the operator reported WebRTC streaming works.
Read-only diagnosis and same-image default-off restoration subsequently passed;
the operator confirmed normal login/picture/audio. The saved diagnosis found
medium/low admission drops and cumulative part/segment publication only for
audio/high. The isolated clock-phase diagnostic subsequently reproduced the
defect in all three cold runs with its aligned control passing. The common
high-source fan-out correction A/B then passed at a7ffb8b1 with
Shared-Clock-Exitcode 0: the old defect reproduced once and 46 positive top-level
checks passed, including three cold runs and one scene-cut check; all ten codec
fixtures stayed in generation 1. Full exact-a7ffb8b1 preparation then passed
with Repair-Prepare-Exitcode 0: all thirteen selected native/startup checks
passed in the rebuilt GStreamer 1.26.2 image, all four codec fixtures stayed in
generation 1, and base/Brave images were built. Default-off exact-repair
deployment then passed with Baseline-Exitcode 0, healthy service and 2/2 disabled
probes; the operator reported the requested normal browser check works.
Same-image activation then passed with Enable-Exitcode 0 and 19/19 denial
probes. First HLS picture and the compact streaming label were reported,
followed around 30 seconds by an initial-readiness timeout. Retry restored HLS
and WebRTC kept working. Client-only repair 73d5ff6d disarms the obsolete start
timer on readiness. The supplied read-only diagnosis and isolated client gate
then passed with old-fault reproduction and all 52 repaired tests/type/build
(Client-Check-Exitcode 0). Scoped exact-73d5ff6d image preparation then passed
with Client-Image-Exitcode 0: fresh client build, cached unchanged server/runtime
layers, both images and private snapshot/marker. The application checkout is
now 73d5ff6d. Default-off exact-73 deployment then passed healthy with
Baseline-Exitcode 0, a private baseline snapshot and 2/2 disabled-route probes;
the operator confirmed the requested normal browser check works. The live
image is now also 73d5ff6d. Same-image HLS activation subsequently passed
healthy with Enable-Exitcode 0 and 19/19 denial probes. On PC/Helium the operator
reports first playback after Retry, one frozen-picture reload, then working
HLS with WebRTC unaffected. "HLS failed" was confirmed without detailed text;
startup/recovery remains unresolved. Conventional HLS is currently enabled.
NEXT [exact-73 read-only playback diagnosis](HLS_CLIENT_READINESS_REPAIR_2026-10-05.md#next-target-block-exact-73-read-only-playback-diagnosis).
Default-off deployment at 97ba4ad9 and normal browser smoke checks passed; enabled HLS
live acceptance remains pending. Tests/builds/codec
execution are supplied target evidence, **NOT EXECUTED IN CODEX**. No live
sustained-HLS acceptance or unique original-login-cause claim.

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

## GOP-repair failure captured and baseline recovered — 2026-10-05

The supplied read-only diagnostic completed with Diagnostic-Exitcode 0 at exact
application 97ba4ad9ab3e635da936a58c8a7ec795da05ba46, using diagnostic blob
5b4b064926f83eccc28fe8bd596db93ba9c19ff1. The sampled image ID was
sha256:5a7e95f71d927bda838d1ee603e3453fbaaa567bc6760b0b29c5fabb4f2ce3ec;
the container was running/healthy with zero Docker restarts, no OOM and no
Neko process exit in the bounded Supervisor sample. Metrics responded.

Only 84 application log lines were captured. They included one HLS negotiation
rejection; the summary contained no generation-start, ready, lease-open or fixed
codec/timeline error marker. The cumulative bootstrap metrics included
not_ready=1, bad_request=8, too_large=1 and unauthorized=1. The other visible
HTTP counters are denial-shaped probes, not successful delivery. Unlike the
earlier 80020d99 diagnosis, this sample does not demonstrate a ready packager
or any successful segment request. Absence of markers/families is not proof
that a packager was never attempted. The negotiation rejection may belong to
another participant/request; do not identify it as the bootstrap/outage cause.
Runtime gst-inspect remains unavailable, not proof of absent codecs.

Baseline mode then completed with Recovery-Exitcode 0, prepared image
my-neko/brave:hls-97ba4ad9ab3e healthy and 2/2 disabled-route 404 probes passing.
The operator confirmed normal login, picture and audio all work again after
this restoration. The HLS failure is recovered, not resolved. Latest enablement
CLI/HTTP probe output is still not supplied. These are supplied target results,
**NOT EXECUTED IN CODEX**; fresh enabled playback remains failed/pending.

NEXT use [summarize-hls-startup.py](../deploy/summarize-hls-startup.py) against
the latest saved private diagnostic. It reads files only, classifies fixed
capture/start/rejection markers, selects allowlisted capture/delivery metrics,
and optionally compares the closest earlier saved environment record with the
diagnostic image/application. It prints neither original log lines nor
participant IDs, pipeline strings, headers or credentials. Sequences span all
participants and earlier environment evidence does not prove live enablement.
No service request/restart, application checkout/build or HLS attempt is needed.
This helper is statically reviewed only; target execution is pending.

Static inspection places the not-ready result after ticket redemption and live
session authorization, in backend readiness/opening. The central manager does
not hold its global delivery mutex while backend.Open runs. The shared capture
AddListener/keyframe/native-pipeline path and the global GStreamer registry
mutex held during native construction still warrant investigation if the saved
stages stop there; no correlated stack establishes either as this failure's
cause. The real-codec fixtures bypass production capture subscription startup,
so their successful GOP/timestamp checks do not validate that path. Keep HLS
disabled and inspect the saved evidence before choosing a repair or another
enabled checkpoint.

## Saved low-source startup boundary and registry isolation repair — 2026-10-05

The supplied file-only summary from helper 7a51ecbb completed with
Saved-Check-Exitcode 0. It read the saved 97ba4ad9 diagnostic stamped
20261005T101606874367498Z, with 84 application log lines. Its closest earlier
environment record matched the diagnostic image/application and had HLS enabled;
this supports the recorded configuration but is not a new live inspection.

The fixed sequence shows a capabilities/request rejection with not_allowed at
line 44, then audio pipeline creation/first listener at 45/46, high at 47/48,
medium at 49/50, low creation at 51, and idle-stop scheduling at 52. No low
first-listener completion appears. Cumulative central HLS open attempts/error
are both 1. HLS subscriptions for audio/high/medium remain 1 each; no low
subscription series is present. Source generations are 1 for audio/high/medium
and 0 for low. The audio streamsink counters remain zero. This locates the last
observed progress inside low-source startup before generation advancement, not
in playlist decoding. It does not identify the exact native call or correlate
the capabilities rejection with that successful ticket's bootstrap request.

Static inspection confirms a sharing defect: gst.createPipeline holds the
global pipelinesLock across C.gstreamer_pipeline_create/gst_parse_launch.
Every active goHandlePipelineBuffer also acquires that lock. A blocked native
constructor therefore blocks all sample callbacks, including other streams.
Capture CreatePipeline increments its generation only after native construction
and appsink attachment, consistent with low remaining at generation 0. The
archive lacks a native stack, so the actual native blocker and whether this
mutex caused the reported incident remain unconfirmed.

The bounded repair moves the global registry lock to map insertion only. IDs
remain atomic, native contexts/channels are fully initialized before registration,
and samples cannot arrive before the caller attaches/plays the new pipeline.
No native construction runs under the sample-registry mutex; existing callback
lookup and teardown behavior are otherwise retained. Capture adds four fixed
Info markers after parse/attachment and before/after Play; the saved-summary
allowlist understands these markers. They add no participant/pipeline values.
This containment repair does not promise a stuck native constructor will return
or complete the HLS worker set. HLS remains disabled on the working service.

A target-only registry regression deliberately holds the sample-registry mutex
while a known-missing element is parsed. Prior code cannot return until that
mutex is released and fails with the fixed expected marker. Corrected code must
return the parse error independently. It uses no actual room or credentials.
[validate-hls-startup-isolation.sh](../deploy/validate-hls-startup-isolation.sh)
pins the existing codec-validation image ID, checks seven baseline source blobs
against unchanged 97ba4ad9, mounts the new test for the negative control, then
mounts only the corrected gst.go for the positive control plus all three
existing codec checks. Unexpected failure/pass stops the gate. Reports are
private and timestamped; no checkout, live service or prepared-image marker is
changed. The fresh-image codec gate also includes this fourth test.

The supplied isolated A/B gate subsequently failed overall, as recorded below.
Only the registry regression's comparison passed; fixed capture-stage tracing
is not deployed. **NOT EXECUTED IN CODEX; full checks/image and live capture/HLS
acceptance pending**. Preserve the recovered image and old evidence. Never infer
production capture or a resolved WebRTC outage from the codec fixtures.

The [GStreamer parse contract](https://gstreamer.freedesktop.org/documentation/gstreamer/gstparse.html#gst_parse_launch)
describes native element construction separately from the Go registry. Its
[ximagesrc documentation](https://gstreamer.freedesktop.org/documentation/ximagesrc/index.html)
also requires early XInitThreads for threaded capture. No such call is present
in the inspected source; whether deployment initializes it externally remains
unknown. Record that as a separate follow-up, not a second unmeasured change in
this mutex-isolation block or an established cause of the current native stall.

## Isolated registry gate failed; cold readiness diagnostic next — 2026-10-05

The supplied target run used repair/helper commit
48f4acf2685b2247e8606deaedb162b9380afeb9, helper blob
a85c72cb736df6709155c9ca3b1ae6b5e48affe9, unchanged application 97ba4ad9 and
codec-validation image
sha256:06df176c5da073548fb20b423335d10cd1ed92a1db1fb5989dd6dacf9d200d3f.
The helper verified seven baseline source blobs before either fixture run.

- The old constructor reproduced the fixed expected registry-wait failure in
  2.00 seconds (negative-control exit 1).
- With only the registry correction mounted, its regression passed in 0.00 s,
  and encoder running-time mapping passed in 0.01 s.
- Smooth real-codec conventional readiness failed at 24.02 s. The sampled logs
  show generation 1 startup and idle-stop scheduling, with no sample rejection
  or generation restart. They do not identify which rendition was incomplete.
- The scene-cut fixture then became ready in generation 1 and passed at 30.19 s.
  The mediahls package failed overall after 54.430 s; positive-control exit and
  final Isolation-Check-Exitcode were both 1.

This is a failed overall acceptance gate. It confirms the deliberately tested
mutex coupling and correction, not complete cold-start reliability. Earlier
successful smooth tests remain valid earlier observations; this new failure
prevents treating them as a reproducible startup guarantee. The failed gate did
not compare old and corrected smooth startup side by side, so it does not prove
the registry correction introduced this failure. Checkout, live service and
prepared-image markers were unchanged; the working baseline remains without HLS.

Static inspection identifies one possible ordering defect: acceptSample drops
output while the high rendition has not established the common anchor. If
another video's initial aligned IDR arrives first and is discarded, subsequent
two-second IDRs cannot join until the next six-second parent boundary. Producing
three complete parents from that point reaches the 24-second readiness limit,
leaving no startup margin. This is a hypothesis about the smooth fixture;
the supplied failure lacks the per-rendition timestamps needed to establish it.
It does not explain the native low-source stall by itself.

The existing real-codec fixtures now wrap the production factory only with
bounded test observations. They retain input/output counts, first/last PTS/DTS,
the first four output keyframes and the observed anchor, plus init/failure/
part/parent readiness snapshots. They retain no media payload or credentials.
Production codec pipelines, channels, clocks, packager code and readiness
deadlines are unchanged. The observations can perturb scheduling and are not an
atomic admission trace; a keyframe observed before anchor creation is suggestive,
not proof it was discarded, because another worker can establish the anchor
before acceptSample runs.

The completed [diagnose-hls-codec-startup.sh](../deploy/diagnose-hls-codec-startup.sh)
run used unchanged application 97ba4ad9 with its existing private output directory.
It verifies the same seven baseline image sources and pins the image ID. Two
separate network-disabled containers use identical diagnostic fixtures: one
retains the original constructor; one mounts exactly the reviewed 48f4acf2
gst.go correction. Each runs three sequential fresh smooth packagers, with a
120-second process timeout and the unchanged 24-second per-start readiness
deadline. There is no retry-until-pass rule, live room or service restart.
Evidence is private, timestamped and records helper/fixture/repair source blobs.

Diagnostic exit 0 means all six attempts and snapshots were collected, including
any failing test counts. It is not a passing acceptance gate or permission to
enable HLS, and does not erase the earlier failure even if all six attempts pass.
Review per-track output/anchor/parent state before choosing a repair. Keep HLS
disabled. The supplied target results are recorded below. **NOT EXECUTED IN
CODEX; reproducible complete repair gate, full application checks/image and
enabled capture/browser acceptance pending**.

## Paired startup diagnosis complete; initial-output anchor repair — 2026-10-05

The operator supplied the completed diagnostic from helper
88f2b25d02fb6734326e284ca31bd4297cb34bf6, helper blob
208da8cbdc3ce74a434029a3915f7b653fe3ddb7, fixture blob
a13242690ccc35bee78784a2807f859b2f3fe8bb and registry repair blob
35cf5d8402a6c733a46c512e38341872187e3bf2. Application HEAD remained at 97ba4ad9;
the verified codec image remained
sha256:06df176c5da073548fb20b423335d10cd1ed92a1db1fb5989dd6dacf9d200d3f.
Private evidence stamp: codec-startup-diagnostic-20261005T110330529448096Z.
Startup-Diagnostic-Exitcode was 0, meaning complete observations, not acceptance.

| Constructor | Attempt 1 | Attempt 2 | Attempt 3 | Package outcome |
| --- | --- | --- | --- | --- |
| Original 97ba4ad9 | failed 24.04 s | passed 18.10 s | passed 18.09 s | exit 1, 60.456 s |
| Registry-isolated 48f4acf2 | failed 24.02 s | passed 18.09 s | passed 18.09 s | exit 1, 60.418 s |

All six attempts stayed in generation 1 with a common high anchor at 30 s and
zero rejected input pushes. In each failing attempt, medium's first encoded
output was a valid, configured 30 s IDR observed before that anchor. Its next
IDRs were at 32/34/36 s, with the anchor set. It had init, healthy output and
LL readiness, but only parents MSN 2/3 and no conventional readiness at the
deadline. Audio/high/low all retained parents 1/2/3 and were conventionally
ready. Successful starts retained MSN 1/2/3 for every rendition.

Together with the admission code, this localizes the fixture timeout: medium's
initial pre-anchor IDR was not admitted; the next eligible parent boundary was
at 36 s (six seconds after anchor), so its third complete parent required the
24-second boundary. This is independently reproduced with both constructors;
the registry correction did not create this observed admission defect. It is
not proof that production capture has matching clocks, nor an explanation of
the separate native low-source stall or all-stream outage. No service, checkout
or prepared-image marker changed during the diagnosis.

The bounded HLS-only repair now holds each non-high worker's first received
output until high establishes the generation's anchor. High continues to
establish the anchor from its first keyframe without waiting. Only the already
received sample is held; existing eight-sample native handoffs, provider queues,
codec profiles, timestamps and 24-second readiness deadline are unchanged. The
wait releases all publication locks, rechecks the anchor after notifications,
and exits on generation cancellation. A handoff overflow or closed drop channel
still requests the existing worker_failure restart, with the existing fixed
overflow metric. Once anchored, the ordinary admission rules still discard
samples earlier than the anchor and enforce aligned IDRs; this is not an
arbitrary-skew repair.

The ordered regression starts medium's output before high's initial keyframe.
The old pump signals its next select after discarding that initial IDR; the
corrected pump holds it. High is then admitted and the synthetic clock advances
only to 18 seconds. Medium must retain parents 1/2/3. Separate checks cover an
unrelated notification, cancellation without publication/restart, overflow and
closed drop-channel termination. These exercise the production output pump,
not only a helper's return value.

The completed [validate-hls-anchor-startup.sh](../deploy/validate-hls-anchor-startup.sh)
run used unchanged application 97ba4ad9 and its existing private output directory.
The pinned codec image's seven baseline blobs are rechecked. Both comparisons
mount the same registry repair and observed fixtures. Old packager.go must fail
the ordered test with its fixed initial-IDR-loss marker; corrected packager.go
must pass the ordered/lifecycle regressions, registry and encoder mapping tests,
and smooth/scene-cut real-codec fixtures in three repetitions. The positive
process timeout is 180 seconds per package and every required test must have
three explicit passes. There is no image build or live service access.

Keep HLS disabled. Passing this bounded gate would allow preparation of full
exact-commit checks/images; it would not establish live capture, browser/media,
authorization/lifecycle or mixed-backend isolation acceptance. The new repair,
regressions and helper now have the supplied target fixture evidence below;
**NOT EXECUTED IN CODEX; full application checks/image and enabled
capture/browser acceptance pending**.

## Anchor A/B and repeated codec gate passed — 2026-10-05

The supplied target run selected repair/helper commit
414639d2493ad1d2106399337c647d1b007f1e9d. Application HEAD stayed at 97ba4ad9;
the pinned codec image was
sha256:06df176c5da073548fb20b423335d10cd1ed92a1db1fb5989dd6dacf9d200d3f.
Helper blob: 9fb174c398fec2589da9c6f711025b7896e38eef.
Packager repair blob: 1d2d1afcd56602c0be5d58220f698b31b5575d12.
Ordered/lifecycle regression blob: 5488d6cec696b6832f04e056c58abeb32e8865d1.
The registry and observed real-codec fixture blobs were unchanged from the
reviewed comparisons. Private evidence stamp:
anchor-startup-20261005T112357362581762Z.

The old packager reproduced the exact initial-IDR-loss marker and failed the
ordered regression in 0.00 s (negative-control exit 1; package 0.196 s).
The corrected comparison returned positive-control exit 0 and final
Anchor-Check-Exitcode 0. All seven required top-level tests passed three times:
registry independence, encoder timestamp mapping, initial-output retention,
cancellation, overflow/closed drop channels, smooth codec readiness and
sustained scene cuts. The gst package passed in 0.199 s; mediahls in 145.305 s.

| Real-codec fixture | Repetition 1 | Repetition 2 | Repetition 3 |
| --- | --- | --- | --- |
| Smooth conventional readiness | 18.11 s | 18.10 s | 18.09 s |
| Sustained scene cuts | 30.19 s | 30.20 s | 30.09 s |

All six fixtures stayed in generation 1 with zero rejected pushes. Each smooth
fixture retained parents MSN 1/2/3 for audio/high/medium/low; scene cuts advanced
every rendition to retained parents 3/4/5 without a restart. In the first smooth
fixture, medium's initial 30 s IDR was again observed before the anchor, but this
time conventional readiness passed with all three parents. This verifies the
ordered admission repair in the bounded fixture, including its first cold
start, instead of relying on later warm repetitions alone. There is no sampled
HLS timeline rejection or generation restart in the supplied fixture output.

The live service, checkout and preparation markers were unchanged. This gate
does not exercise native X11 capture, real credentials, the public proxy/player
or mixed-backend room-event isolation. The native low-source blocker and the
reported all-stream outage still require live acceptance of the new image.

The next full preparation was attempted using the block in
[HLS_LL_HLS_VALIDATION.md](HLS_LL_HLS_VALIDATION.md) to fast-forward the clean
testing checkout to exact application 414639d2 and prepare its tests/images in
../neko-hls-results-414639d2493a. Retain the current healthy default-off 97ba4ad9
service and its private directory/images. Preparation runs full selected
client/server tests, type/build, two fuzz jobs, exact-source real-codec checks
and new uniquely tagged images; it does not replace that service. Review its
output before default-off image deployment and a fresh normal-browser check.
HLS remains disabled. These are supplied target results, **NOT EXECUTED IN
CODEX; full exact-commit preparation, new-image deployment and enabled
capture/browser/grouped acceptance pending**. Audit/package findings stay open.

## Full exact-414639d2 preparation failed — 2026-10-05

The supplied full-preparation output ended with Repair-Prepare-Exitcode 1.
Its earlier checks passed: all 47 client tests, TypeScript/build, all 13 Go
packages, 30 s WebSocket fuzz with 1,066,192 executions, 30 s HLS request fuzz
with 367,784 executions and the trailing server build. The rebuilt codec image
reported GStreamer 1.26.2. Registry independence and encoder segment mapping
passed in 0.00/0.01 s; the ordered initial-IDR regression, cancellation and
overflow/closed-channel lifecycle checks also passed.

The smooth fixture began generation 1 at 11:39:00 UTC, restarted with
worker_failure into generation 2 at 11:39:02 and became conventionally ready
at 11:39:20. It then FAILED in 20.03 s with `cold fixture caused 2 packager
generations`. All four replacement tracks had three parents (MSN 2/3/4), valid
init and zero rejected input pushes. Those observations describe generation 2:
the original diagnostic factory replaced each trace at restart, so the output
does not identify the failing generation-1 worker/stage. The single pipeline
not-found warning around teardown is not proof of the restart's cause. There
is no sampled HLS timeline-gap/sample-rejection marker in this codec interval.

The scene-cut fixture passed in 30.19 s, stayed in generation 1 and advanced
all four tracks to parents MSN 3/4/5. The codec mediahls package nevertheless
failed in 50.553 s. The earlier passing anchor A/B remains valid bounded
evidence for initial-IDR retention; it does not clear this later fresh-image
worker failure. Do not remove the generation-1 assertion, increase queues or
relax deadlines to turn this output into a pass.

Full preparation stopped before its dependency-audit and new base/Brave image
steps. Its new private validation marker remains PENDING. The application
checkout and mutable codec-validation tag now contain 414, while the running
service remains the previously confirmed default-off 97ba4ad9 deployment.
No deployment/enablement follows from this failed gate. Preserve both output
directories and image IDs; old 97-pinned A/B helpers are not applicable to the
rebuilt mutable tag without their original image.

### Bounded observation follow-up

Worker_failure requests now have a fixed `HLS worker restart requested` marker
with variant, generation, stage and reason. The six call sites distinguish
input push failure, output handoff overflow/closure, anchor-wait overflow/
closure and audio output stall; the request reason/metrics/exit decisions are
unchanged. This adds no sample processing, queue capacity or codec/clock change.
Private live-log summaries retain only fixed marker counts or allowlisted
variant/stage/reason values; no new raw credential fields are exported.

The real-codec fixture retains the first two replaced workers plus the latest
for each track (at most 12 traces total). Each trace has its generation,
closed state, first-output delay/lifetime, input/rejection/output counts,
first/last PTS and at most four keyframe observations; no payload bytes are
retained. Ordinary transcoder channels are still returned directly. These
concurrent observations can perturb scheduling and do not form an atomic
admission trace.

The completed [diagnose-hls-worker-startup.sh](../deploy/diagnose-hls-worker-startup.sh) ran
from the reviewed helper commit while keeping checkout 414 and live HLS off.
It pins the rebuilt codec image ID and checks eight baseline source blobs.
Only the fixed logging and bounded observed fixture are mounted. Three fresh
containers each run six startup checks with -count=1 and a 60 s per-package
timeout; every outcome is retained, including failures. This preserves the
full preparation's two-package startup shape without repeating the already
passing scene-cut fixture or rebuilding images. Intentional anchor overflow
unit cases also produce restart markers, so read each real-codec interval
separately. Diagnostic exit 0 means observations completed, not acceptance.
No live access, checkout mutation or preparation-marker change is performed.

Observation source/helper review is complete; **NOT EXECUTED IN CODEX**.
Supplied target diagnosis is recorded below; a resolved full-preparation gate,
new-image baseline, enabled live capture/browser and grouped acceptance remain
pending.

## Three fresh worker diagnoses passed; original restart not reproduced — 2026-10-05

The supplied helper 5303449509b05aaeaabb6d6d5d2f7b6eb94fbc19 ran against
unchanged application 414639d2 and verified eight baseline source blobs.
Codec image: sha256:d6e7e49a1423ec92994cf3c82559936ae6ff10bd0080be776cc391696355bc3a.
Helper blob: 41b80f45c810462019a25db63a7ca48ffde510e8.
Packager observation blob: 03f727e26284a717751131c3bf406667a22f2b6d.
Observed fixture blob: bd796c1ee9dcb4df5f062d33b0902142991d72c8.
Private evidence stamp: worker-startup-diagnostic-20261005T120101847814028Z.

All six required checks passed once in each of three fresh containers and
processes (18 top-level passes). Every smooth fixture became conventionally
ready in 18.11 s in generation 1; all four tracks retained parents MSN 1/2/3,
valid init, healthy state and zero rejected pushes. mediahls package durations
were 18.423/18.418/18.429 s. Audio first-output delays were 19/19/20 ms; high
96/74/74 ms, medium 91/60/58 ms and low 93/71/68 ms. Audio and medium could
still be observed before high's anchor, without a restart under this schedule.

The six worker restart warnings (medium, anchor, queue_full/drops_closed) are
inside the three intentional overflow/closure unit-test intervals. They do not
show a failed real-codec startup. Pipeline-not-found warnings belong to encoder
mapping teardown. Worker-Diagnostic-Exitcode was 0. No checkout/service or
preparation-marker change occurred. This does not identify the earlier worker
failure or replace the failed full-preparation outcome with a passing gate.

### Narrow audio-drainage repair and controlled comparison

Static inspection identified a bounded hazard in the otherwise useful 414
video-IDR hold: it also holds the first AAC output until high establishes the
anchor. The input pump/native AAC encoder continue producing frames while
that output pump is waiting; the eight-frame Go handoff reports Drops on full.
The wait then requests worker_failure. At the observed 48 kHz/1024-sample AAC
cadence, a sufficiently delayed high frame can fill that small handoff even
when steady streaming would drain it normally. This mechanism follows the
code; it is not established as the cause of the earlier unobserved failure.

The production change narrows the initial-output hold to video: high can still
establish the anchor; medium/low retain their initial IDR. AAC keeps draining
through existing admission, which discards output while the anchor is unset.
After the anchor, ordinary AAC timestamps/geometry/codec validation, aligned
part admission and retention remain unchanged. Early audio is not published
before the video clock exists, and later audio is not rebased to zero.
There are no added buffers, queue/deadline/encoder changes or altered restart
rules. Very late anchors/provider overflow/arbitrary clock skew remain bounded
failure cases, not new support claims.

The new audio unit uses an output-loop re-entry barrier to verify more than
eight AAC frames are drained before high, with no unanchored publication.
It then supplies anchored AAC at a 250 ms common-clock offset and requires
the corresponding 12,000 audio ticks in the first admitted fragment sample.
The old hold must fail with the fixed wait marker. The native fixture delays
generation 1's first high input by twelve AAC periods: 256 ms. Replacement
generations and other inputs are not delayed and no PTS is changed. This
injection is solely a controlled test condition, not a deployment value.

The [validate-hls-audio-anchor.sh](../deploy/validate-hls-audio-anchor.sh) gate
was prepared for the reviewed repair while checkout stayed at 414 and live
HLS remained off. It pins the recorded image ID, verifies the eight 414 source
blobs and requires unchanged native/capture/encoder sources. The same new unit/codec
fixtures are mounted on both sides. Prior behavior is packager.go from 53034495
(414 admission plus fixed logging); the corrected side changes only the audio
wait condition. The negative unit must show its wait failure, and the negative
native fixture must show exactly two generations and audio/anchor/queue_full.
An unexpected negative outcome stops this gate for review.

The corrected side must pass eight checks in three fresh processes, including
normal/delayed-high codec readiness, AAC drainage/common-clock offset, retained
video IDR, cancellation/overflow and registry/encoder mapping. The scene-cut
fixture also runs once in the first process. -count=1 and fixed three attempts
prevent retry-until-pass; the 90 s package timeout does not relax HLS's 24 s
readiness. The future full-image codec gate includes both new regressions.
No service access, image build, checkout or marker mutation occurs in this A/B.

Source/diffs are statically reviewed; **NOT EXECUTED IN CODEX; supplied
controlled A/B and subsequent full exact-repair preparation passed as recorded
below. The new-image default-off baseline and requested normal-browser
checkpoint also passed; enabled live acceptance remains pending**.
A passing controlled A/B does not prove the cause of the
unobserved earlier worker restart, native low-source stall or all-stream outage.

## Controlled audio-anchor A/B passed — 2026-10-05

The supplied run ended with AUDIO ANCHOR A/B GATE PASSED and
Audio-Anchor-Exitcode 0. Application checkout remained
414639d2493ad1d2106399337c647d1b007f1e9d; repair was
71a14d2174dafbc12b1880adde6dc68176bfe9af. Prior behavior was the observation-only
packager from 5303449509b05aaeaabb6d6d5d2f7b6eb94fbc19. The helper verified eight
baseline image-source blobs and unchanged native/capture/encoder sources in
the repair. Both sides mounted the same repaired test fixtures; the production
difference was the AAC wait condition.

Provenance:

- Codec image: sha256:d6e7e49a1423ec92994cf3c82559936ae6ff10bd0080be776cc391696355bc3a.
- Helper blob: a9d9d7f55438776cb8021865f69070ee783996be.
- Codec fixture blob: d7c6a130f779c0aa87fa8bc2695a0bc811488df0.
- Startup unit blob: ae8c9d1945cbd4c5b112fad16bf920b53722b3dd.
- Prior packager blob: 03f727e26284a717751131c3bf406667a22f2b6d.
- Repaired packager blob: 3c3fec7bf4de96b606342d48b9953476170bf7fc.
- Private evidence stamp: audio-anchor-20261005T124423579375451Z under
  /opt/docker/nekoNew/neko-hls-results-414639d2493a.

Both old-code negative controls failed as required with exit 1. The drainage
unit failed at 1.00 s with "audio output waited for the high anchor". Under
the controlled 256 ms high-input delay, the real-codec fixture logged
audio/anchor/queue_full in generation 1, followed by worker_failure and
generation 2; its generation-1 assertion failed at 18.39 s (package 18.948 s).
The first audio worker read one output, accepted 11 inputs with zero rejected
pushes and closed at 207 ms with no anchor. High had zero outputs before that
restart. The replacement generation became ready with all four tracks'
parents MSN 1/2/3. This directly reproduces the AAC hold/overflow mechanism
under the injected schedule.

The corrected side passed eight required checks once in each of three fresh
containers/processes, plus scene cuts once in the first process: 25 top-level
passes. All seven real-codec fixtures remained in generation 1 with zero
rejected pushes and valid init/healthy readiness for all four tracks.

| Corrected cold process | Normal readiness test | Delayed-high readiness test | Scene-cut test | mediahls package |
| --- | --- | --- | --- | --- |
| 1 | 18.10 s, generation 1 | 18.09 s, generation 1 | 30.09 s, generation 1 | 66.600 s |
| 2 | 18.11 s, generation 1 | 18.09 s, generation 1 | Not selected | 36.480 s |
| 3 | 18.11 s, generation 1 | 18.09 s, generation 1 | Not selected | 36.515 s |

Every normal/delayed cold fixture retained parents MSN 1/2/3 in audio, high,
medium and low; the scene-cut fixture advanced to MSN 3/4/5. High's observed
first-output delays were 211/54/55 ms in normal starts and 286/283/286 ms in
delayed starts. The first corrected normal start therefore tolerated a
naturally later high output as well, without establishing what caused the
earlier unobserved failure. The drainage/250 ms timestamp-offset unit, retained
video IDR, cancellation/overflow, native registry independence and encoder
mapping checks all passed in every process. The six medium/anchor warnings
are intentional queue_full/drops_closed lifecycle cases, outside the codec
fixtures. The fixed three attempts used -count=1 and a 90 s package timeout;
the 24 s HLS readiness bound remained unchanged.

No image build, live service access, checkout mutation or preparation-marker
change occurred. The earlier full-414 marker remains PENDING and its failure
is preserved; this isolated comparison does not turn it into a passing image
gate. The live service remains the working default-off 97ba4ad9 deployment.
The native low-source construction stall and reported live all-stream outage
are still unconfirmed; controlled AAC overflow is not retrospective proof of
their cause.

The next step at that checkpoint was the pinned full exact-71a14d21 test/image block in
[validation section 1](HLS_LL_HLS_VALIDATION.md#1-prepare-exact-tests-and-images-without-replacing-the-service),
using /opt/docker/nekoNew/neko-hls-results-71a14d2174da. It subsequently passed
while the working default-off 97 service remained running and HLS disabled.
Its output is reviewed below for a separate default-off deployment
and fresh normal-browser check; enabled HLS picture/audio, authorization,
lifecycle/isolation and grouped device/resource acceptance remain pending.
All execution above is supplied target evidence, **NOT EXECUTED IN CODEX**.

## Full exact-71a14d21 preparation passed — 2026-10-05

The supplied tail ends with AUTOMATED/IMAGE GATE PASSED, running service
unchanged and Repair-Prepare-Exitcode 0. Image tags and the evidence path
identify application 71a14d2174dafbc12b1880adde6dc68176bfe9af. The excerpt starts
inside the codec-image dependency build; it omits earlier client/type/build,
Go/fuzz and server-check output. The known fail-fast preparation script's final
success covers those preceding stages, but this tail does not independently
show their counts or fuzz execution totals. No missing output is invented and
no repeat of passing preparation is required.

The newly built GStreamer 1.26.2 image explicitly passed all nine selected
startup checks once. Native registry independence and encoder running-time
mapping passed; video initial-IDR retention, anchor cancellation/overflow and
AAC drainage/common-clock offset passed. The real-codec fixtures passed:

| Fixture | Test duration | Final generation | All-four-track parent window |
| --- | --- | --- | --- |
| Normal cold readiness | 18.11 s | 1 | MSN 1/2/3 |
| Delayed-high cold readiness | 18.09 s | 1 | MSN 1/2/3 |
| Sustained scene cuts | 30.19 s | 1 | MSN 3/4/5 |

All four tracks had valid init, healthy state and zero rejected pushes.
High first-output delays were 54 ms normal, 284 ms delayed and 53 ms scene-cut.
No real-codec worker restart occurred. The medium/anchor queue_full and
drops_closed warnings are inside the deliberate overflow/closure unit cases.
gst package duration was 0.202 s; mediahls package duration was 66.704 s.

Base and Brave builds completed with tags my-neko/base:hls-71a14d2174da and
my-neko/brave:hls-71a14d2174da. The final private snapshot/success marker path
is /opt/docker/nekoNew/neko-hls-results-71a14d2174da. The deployer will compare
the image ID against that preparation record before replacing the service.
Dependency-audit report exit 1 is recorded for classification/remediation;
no fresh report contents or passing dependency-security claim are supplied.

This closes full exact-71 preparation, without changing the earlier failed
414 checkpoint or proving live playback. The running service remains the
working default-off 97ba4ad9 image until the next operator block. Preserve its
rollback capability and the private failure/A/B evidence. The native low-source
stall and reported all-stream outage still lack a confirmed live cause.

The next step at that checkpoint was the exact-71 default-off deployment block in
[validation section 1](HLS_LL_HLS_VALIDATION.md#1-prepare-exact-tests-and-images-without-replacing-the-service),
which subsequently passed as recorded below. Its normal-browser checkpoint
also subsequently passed. NEXT section 2's same-image HLS enablement/HTTP
denial gate; no rebuild or Caddy change is needed. Enabled HLS picture/audio,
valid authorization/lifecycle, room-event/isolation and grouped device/resource
acceptance remain pending. **NOT EXECUTED IN CODEX; supplied target preparation passed.**

## Default-off exact-71a14d21 deployment passed — 2026-10-05

The supplied operator block ended with Baseline-Exitcode 0. It identified
application 71a14d2174dafbc12b1880adde6dc68176bfe9af and deployer blob
c6f52dc80fdf605ec908f3fe3856ce23e015e494. The unchanged deployer checked the
preparation marker/image record, saved the prior image and deployed
my-neko/brave:hls-71a14d2174da without the HLS overlay. The container became
healthy and a private baseline snapshot was recorded under
/opt/docker/nekoNew/neko-hls-results-71a14d2174da. The two disabled bootstrap/media
HTTP probes returned the expected 404 (2/2 passed).

The running application is now the prepared 71 image, not the earlier 97
baseline. This verifies deployment health and disabled-route behavior only.
At that deployment checkpoint, fresh normal-browser confirmation was still
pending. The operator subsequently confirmed the requested check works
without HLS, as recorded below. Do not repeat baseline deployment.
The prior working image and private failure/A/B evidence remain available.
NEXT the same-image conventional-HLS activation/HTTP denial gate.
Enabled live capture/playback, the previous all-stream symptom's cause and
authorization/lifecycle/isolation/device/resource acceptance remain pending.
**NOT EXECUTED IN CODEX; supplied target baseline passed.**

## Default-off exact-71 browser checkpoint confirmed — 2026-10-05

The operator replied "Ja ohne klappt alles" to the requested fresh
normal-browser login/picture/audio/control check without HLS. This closes the
default-off 71a14d21 browser checkpoint after its healthy deployment and 2/2
disabled-route probes. No additional build or service mutation occurred in
Codex, and no enabled-HLS playback claim follows from this confirmation.

NEXT [validation section 2](HLS_LL_HLS_VALIDATION.md#2-explicit-conventional-hls-deployment-and-invalid-input-checks):
enable conventional HLS in the same prepared 71 image and review the 17 public
plus 2 cleartext-denial probes before one valid admin playback attempt.
Keep application HEAD/evidence at 71, preserve the saved rollback tag and use
the already-reviewed Caddy configuration without another merge/reload.
Enabled live capture/picture/audio, the earlier all-stream symptom's cause and
authorization/lifecycle/isolation/device/resource acceptance remain pending.
**NOT EXECUTED IN CODEX; operator-confirmed default-off browser behavior only.**

## Exact-71 HLS activation/invalid-input checkpoint passed — 2026-10-05

The supplied enable block identified application
71a14d2174dafbc12b1880adde6dc68176bfe9af and unchanged deployer blob
c6f52dc80fdf605ec908f3fe3856ce23e015e494. It redeployed
my-neko/brave:hls-71a14d2174da with conventional HLS enabled; the container
became healthy and the private enable snapshot was recorded under
/opt/docker/nekoNew/neko-hls-results-71a14d2174da. All 17 public invalid-input
probes passed with expected statuses and headers. Both direct loopback
cleartext bootstrap/media probes returned the expected 403. Enable-Exitcode
was 0; no source/image or Caddy change occurred beyond using the HLS overlay.

This closes the exact-71 activation/invalid-input gate only. Synthetic unknown
tickets/cookies do not open a valid playback lease or start/demonstrate live
packager readiness. Previous all-stream failures are still not explained by
this result. Keep the saved rollback tag and all prior evidence.

NEXT [validation section 3](HLS_LL_HLS_VALIDATION.md#first-bounded-pictureaudio-checkpoint-repeat-after-diagnosisrepair):
first verify ordinary WebRTC works with HLS enabled, then make one HLS admin
attempt in a separate private window while WebRTC remains connected. Record
picture/audio, startup impression, exact failure if present and the ordinary
viewer status. Only if both work, extend to five foreground minutes with chat,
control release/take and participant join. A failure must be diagnosed before
same-image default-off restoration; the pinned conditional recovery block is
prepared in the runbook, not executed. Valid authorization/lifecycle and
grouped device/resource acceptance remain pending. **NOT EXECUTED IN CODEX;
supplied exact-71 activation/HTTP passed, enabled playback pending.**

## Exact-71 enabled HLS bootstrap failed; WebRTC reported working — 2026-10-05

The operator reported that only the WebRTC stream works. HLS initially showed
connecting, then failed with "HLS bootstrap failed; retry manually" and the
Retry HLS/Use WebRTC actions. No successful HLS picture/audio or room-event
interval is demonstrated. The earlier activation/19 denial probes remain a
passing boundary gate, not valid playback acceptance.

The client uses this same terminal detail for non-201 bootstrap responses and
request/response-processing exceptions. The message alone does not distinguish
HTTP denial, readiness timeout, network failure or invalid response parsing.
Do not infer a particular cause or mark the prior all-stream symptom resolved
from this report. This attempt's HTTP status/log/metric evidence is not supplied.

NEXT the pinned read-only diagnosis followed by same-image default-off baseline
restoration in [validation section 3](HLS_LL_HLS_VALIDATION.md#first-bounded-pictureaudio-checkpoint-repeat-after-diagnosisrepair).
It preserves raw logs privately and prints a fixed safe summary before the
service restart. Diagnostic failure must not prevent restoration. Both helpers
are unchanged at exact application 71a14d2174dafbc12b1880adde6dc68176bfe9af;
no source fix, new image, checkout movement or Caddy change is required for
this evidence/recovery step. The restart briefly interrupts existing viewers.
Confirm normal behavior after restoration, then analyze the saved evidence
before another HLS attempt or repair. **NOT EXECUTED IN CODEX; supplied live
HLS failure, diagnostic/recovery results pending at that checkpoint.**

## Exact-71 readiness diagnosis and recovery passed — 2026-10-05

The supplied read-only diagnosis returned Diagnostic-Exitcode 0, at exact
application 71a14d2174dafbc12b1880adde6dc68176bfe9af and unchanged helper blob
ddfe001618f8733447f4bdd328a8519db9acef06. It captured 950 application log lines.
The container was running/healthy, with zero restarts and OOM false; no Neko
process exit was found in the bounded supervisor sample. Its inspected image
ID was sha256:9194cbd8fb6b82180f087620fff5cd258db8e60c2a0ab64ee20be870816c60de.
The unavailable `gst-inspect` version is not evidence that codecs are absent.

The fixed summary records one not-ready bootstrap, two negotiation rejections,
one started packager generation and one stop after idle grace. All four worker
starts/generations were recorded once. Medium/low packager keyframe-admission
drops were 759/564; cumulative init publication exists for all four tracks.
Cumulative part/segment bytes exist only for audio/high, with no published
medium/low parts or parents. Zero current object/running gauges are consistent
with the already logged idle teardown. These are cumulative/bounded observations,
not a correlated per-frame trace. No sampled fixed timestamp-gap or worker
restart marker appears. No valid lease/readiness success is demonstrated.

The same-image default-off baseline then passed with Recovery-Exitcode 0,
healthy my-neko/brave:hls-71a14d2174da, a private baseline snapshot and 2/2
disabled bootstrap/media probes returning 404. The operator subsequently
confirmed normal login/picture/audio work again. Keep HLS disabled.

Static inspection narrows a candidate: `acceptSample` discards video before
high's timestamp anchor and requires the first IDR in a common parent bucket.
The fixed-GOP workers start from their own first input timestamps. Every prior
real-codec fixture used identical first PTS for all sources; it did not exercise
independently phased warm/cold capture. The live summary does not contain actual
per-track IDR phases, so the hypothesis is not yet the proven live cause.

NEXT [isolated source-clock-phase diagnosis](HLS_LL_HLS_VALIDATION.md#isolated-source-clock-phase-diagnostic-after-exact-71-recovery).
Its extra-tagged expected-defect check uses artificial high/medium/low phases
of 800/50/100 ms while retaining the production packager, native wrapper,
transcoders, queues, timestamps and 24-second readiness deadline. One aligned
positive control and three fixed cold skew runs use the existing exact-71 codec
image with read-only test mounts. A result of zero means the expected failure
signature reproduced, not acceptance. No live restart, new image, application
checkout movement, preparation-marker change or Caddy edit is involved.
Production repair is deliberately pending target reproduction. **NOT EXECUTED
IN CODEX; supplied diagnosis/recovery passed, controlled reproduction and
enabled live acceptance pending at that checkpoint.**

## Independent source/GOP phase defect reproduced — 2026-10-05

The full supplied clock diagnostic identifies application
71a14d2174dafbc12b1880adde6dc68176bfe9af, helper
409482b47e4ed962a2f9a3fd129114a53563f15a and codec image
sha256:2c885aa463ee9514b120d541e9852f4e9cd5cc334a72e9373cacf6896c963c48.
Its helper/fixture blobs were 5bc3086d2bf12e1a555478e04bd752c94b999c1a,
723975ed19cdbbca68bb9e2e7caf2f066565fe3c and
5f800a1437b2c1bd4b171952873250a7e0c8a2e4 respectively. All four subprocesses
returned zero, and Clock-Skew-Exitcode was 0.

The aligned positive control became conventionally ready at 18.11 seconds,
with all four tracks in generation 1, three parents/MSN 1–3 per track and no
rejected pushes. Each of the three cold skew runs reproduced not-ready at
24.02 seconds without restarting. The high anchor was 30.8 seconds; medium
IDRs were at 30.05/32.05/34.05/36.05 seconds, low at
30.1/32.1/34.1/36.1 seconds. Their initial IDRs precede high's anchor, and
subsequent IDRs occupy odd one-second buckets, so neither track can join the
common six-second parent. Both remained init-ready, not failed and at part
index -1, with no parts/parents despite 479/359 consumed outputs and four
observed IDRs. Audio/high were ready with three parents. All native pushes
were accepted. This confirms the controlled phase-admission defect and shows
why the prior aligned fixture missed it. It is not a measurement of the live
capture phases or proof that every earlier outage had this same cause.

## Common encoded-video input correction — implementation before target A/B

`model.go` now identifies high as the encoded input for all three HLS output
variants. `packager.go` opens one high-video provider subscription and one
audio subscription. The video input pump fans each same immutable unit into
the three existing bounded transcoders before advancing to the next source
event. Scaling/rate conversion and fixed-GOP encoding remain per output.
Input PTS/DTS, duration, generation and media bytes are preserved; no per-track
time shift, relaxed part admission, longer readiness or larger queue is used.
Strict high-source FORMAT/generation and emitted per-rendition caps checks
remain. Failure of any native push still requests a bounded HLS-generation
restart. Shared-source ownership handles successful shutdown and partial
construction without duplicate subscription closes.

This replaces four HLS provider subscriptions with two while retaining four
workers. The three video decoders now consume high-resolution input, so CPU/RSS
costs must be measured anew; no universal resource benefit is claimed. No
capture/WebRTC/WebCodecs implementation or native wrapper was changed.
Capability `source_id` values reflect the common input; output IDs and player
geometry/rate contract are unchanged. The client fixture was updated accordingly.

New unit checks cover unchanged fan-out timestamps/bytes, peer push rejection,
partial construction and normal cancellation/resource ownership. The codec
fixture verifies one audio/high subscription, equal first video PTS and four
aligned IDRs per output. Its new skew case uses the reproduced source phases.
The prior expected-defect check skips under the repaired topology and is used
only from the pinned 409482b4 fixture against the old image. Full preparation's
codec selection includes the new checks; only its whole-suite timeout moves
from 90 to 120 seconds, while per-attempt readiness remains 24 seconds.

The next step at implementation was the [shared-video-clock A/B](HLS_LL_HLS_VALIDATION.md#shared-video-clock-repair-ab-while-the-confirmed-baseline-stays-running):
one pinned old-code signature reproduction, then three predetermined positive
cold runs and one sustained scene-cut run using the existing immutable codec
image and read-only mounts. Keep the confirmed default-off exact-71 live
service/checkout and successful preparation marker unchanged. Full exact-repair
tests/images and live picture/audio/authorization/lifecycle/device/resource
acceptance remain pending. **NOT EXECUTED IN CODEX; supplied diagnosis proves
the controlled old defect, new correction statically reviewed only at that checkpoint.**

## Common encoded-video input repair A/B passed — 2026-10-05

The complete supplied output pins application
`71a14d2174dafbc12b1880adde6dc68176bfe9af`, repair
`a7ffb8b13448a8329c6df24fdcb182ac32ca398c`, diagnostic
`409482b47e4ed962a2f9a3fd129114a53563f15a` and the unchanged codec image
`sha256:2c885aa463ee9514b120d541e9852f4e9cd5cc334a72e9373cacf6896c963c48`.
Helper blob was `b9be81704ccd719afaa15016ee73c253112f12a0`; the five read-only
repair mounts were the committed model, packager, packager tests, new
shared-input tests and real-codec fixture. The old fixture blobs match the
previous phase diagnostic. Negative signature and all three positive cold
processes returned 0; final Shared-Clock-Exitcode was 0.

The old independent-source code reproduced its expected not-ready signature
at 24.02 seconds in generation 1: audio/high ready, medium/low blocked at part
index -1 despite flowing IDRs, no rejected native pushes. This negative PASS
means the known defect was reproduced, not successful HLS playback.

The repaired code passed all fifteen selected checks in each of the three
predetermined fresh processes, plus sustained scene cuts once: **46 positive
top-level passes**, apart from the one old-code expected-defect check.

| Real-codec fixture | Target test duration | Repetitions | Result |
| --- | --- | --- | --- |
| Normal cold readiness | 18.08 seconds each | 3 | All four tracks ready, generation 1 |
| Delayed high input | 18.07 seconds each | 3 | All four tracks ready, generation 1 |
| Artificially skewed source phases | 18.83 seconds each | 3 | All four tracks ready, generation 1 |
| Sustained scene cuts | 30.07 seconds | 1 | Parent advancement continued, generation 1 |

The skew fixtures retained the high anchor at 30.8 seconds; all three video
outputs now begin at 30.8 seconds and share IDRs at 30.8/32.8/34.8/36.8 seconds.
Each track had codec initialization and three retained parents/MSN 1–3, unlike
the old medium/low tracks. Exactly one audio and one high-video subscription,
unchanged input units and no rejected native pushes are enforced by the
positive assertions. All ten real-codec fixtures stayed in generation 1.
No readiness deadline, queue or admission limit was relaxed. The warning
markers for queue_full/drops_closed/push_failed occur within intentionally
induced error-case unit checks; they are not a live failure capture.

Private synthetic evidence remains under
`/opt/docker/nekoNew/neko-hls-results-71a14d2174da/shared-clock-ab-20261005T154825115356001Z`.
The helper did not move the checkout, change preparation markers, build a new
image, access/restart the live service or edit Caddy. The operator-confirmed
default-off 71 baseline remains the current runtime checkpoint.

This verifies the controlled phase-admission repair and its focused lifecycle
regressions. It does not measure the failed live source phases, explain every
earlier all-stream outage, or establish HLS browser/device acceptance. The next step was
[full exact-a7ffb8b1 tests/image preparation](HLS_LL_HLS_VALIDATION.md#exact-a7ffb8b1-preparation-after-the-shared-video-clock-ab-passed)
using a new private output while retaining the working default-off 71 service.
Default-off repair deployment/browser, same-image enablement, valid playback
and grouped authorization/lifecycle/device/resource acceptance follow only
after their respective gates pass. **NOT EXECUTED IN CODEX; supplied isolated
repair A/B passed, full preparation and enabled/live acceptance pending at that checkpoint.**

## Full exact-a7ffb8b1 preparation passed — 2026-10-05

The supplied excerpt begins partway through dependency installation for the
codec-validation image and ends with AUTOMATED/IMAGE GATE PASSED and
Repair-Prepare-Exitcode 0. The built image tags and final private output path
identify application `a7ffb8b13448a8329c6df24fdcb182ac32ca398c` from the pinned
operator preparation block. Earlier client checks/type/build, general Go
checks/fuzz jobs and server-check image output are outside the excerpt. Their
completion is covered by the unchanged fail-fast script's final success;
do not claim fresh individual counts from this tail.

The rebuilt GStreamer 1.26.2 image explicitly passed all thirteen selected
native/startup checks: native registry/mapping, initial-video-IDR retention,
cancellation/overflow, AAC drainage, same-unit fan-out/peer failure,
construction/cleanup, and four real-codec fixtures. The latter passed in one
generation each:

| Real-codec fixture | Target test duration | Recorded result |
| --- | --- | --- |
| Normal cold readiness | 18.08 seconds | All four tracks ready, generation 1 |
| Delayed high input | 18.07 seconds | All four tracks ready, generation 1 |
| Artificially skewed source phases | 18.83 seconds | Shared video IDRs, all four tracks ready, generation 1 |
| Sustained scene cuts | 30.08 seconds | Parents advanced to MSN 3–5, generation 1 |

All four real-codec traces show zero rejected native pushes. The skew fixture
preserved the high anchor at 30.8 seconds and common video IDRs at
30.8/32.8/34.8/36.8 seconds. All tracks retained codec initialization and three
parents. The mediahls package completed in 85.394 seconds; gst in 0.206 seconds.
queue_full/drops_closed/push_failed warnings occur in deliberate failure-case
unit checks, not live stream diagnosis. Dependency-audit report exit 1 remains
open; final preparation success is not a passing security audit.

The client and server image stages completed, followed by both
`my-neko/base:hls-a7ffb8b13448` and `my-neko/brave:hls-a7ffb8b13448` builds.
The final private snapshot and successful exact-commit marker are under
`/opt/docker/nekoNew/neko-hls-results-a7ffb8b13448`. The existing default-off
71 service was unchanged; keep its image/evidence and prior rollback tags.
The deployer must compare the new marker and actual prepared image ID before
touching the service. Build-export digests are not a supplied runtime inspection.

The next step was [exact-a7ffb8b1 default-off deployment/browser confirmation](HLS_LL_HLS_VALIDATION.md#exact-a7ffb8b1-default-off-deployment-and-browser-confirmation-passed),
with the working 71 image saved as the new rollback target. Only after healthy
deployment and ordinary login/picture/audio/control confirmation, proceed to
separate same-image enablement and bounded HLS playback. Valid browser picture/
audio, room-event interval and grouped authorization/lifecycle/device/resource
acceptance remain pending. **NOT EXECUTED IN CODEX; supplied exact-repair
automated/image preparation passed, deployment and enabled/live acceptance pending at that checkpoint.**

## Exact-a7ffb8b1 default-off deployment/browser checkpoint passed — 2026-10-05

The supplied baseline output identifies application
`a7ffb8b13448a8329c6df24fdcb182ac32ca398c` and unchanged deployer blob
`c6f52dc80fdf605ec908f3fe3856ce23e015e494`. The deployer replaced the running
service with `my-neko/brave:hls-a7ffb8b13448` without the HLS overlay. The
container became healthy; the loopback HTTP binding remains 127.0.0.1:8082
and the established UDP ports remain published. Its private baseline snapshot
was recorded under `/opt/docker/nekoNew/neko-hls-results-a7ffb8b13448`.
Both disabled bootstrap/media probes returned the expected 404 (2/2), and
Baseline-Exitcode was 0. The deployer's existing saved prior-image mechanism
retains the working 71 image under `my-neko/brave:rollback-hls-a7ffb8b13448`;
older evidence/rollback tags remain available. No Caddy change was part of this block.

The operator then replied "borwser klappt" to the requested private-window
normal login/picture/audio/control check without a media override. This closes
the bounded reported default-off repair-image browser checkpoint. It is not
an enabled HLS/device/resource or five-minute room-event result.

The next step at that checkpoint was [same-image conventional-HLS activation and denial probes](HLS_LL_HLS_VALIDATION.md#exact-a7ffb8b1-same-image-activation-and-denial-probes-passed),
without pulling documentation commits or rebuilding. Then perform one bounded
admin HLS picture/audio attempt alongside a working WebRTC viewer. If bootstrap
or playback fails, capture read-only diagnosis before restoring this confirmed
same-image default-off baseline; no repeated blind attempts. Valid playback,
room-event isolation, passive authorization/lifecycle and grouped device/resource
acceptance remain pending. **NOT EXECUTED IN CODEX; supplied healthy default-off
deployment and reported browser checkpoint passed, enabled/live acceptance pending.**

## Exact-a7ff activation/first picture and client deadline repair — 2026-10-05

Same-image conventional-HLS activation subsequently passed with Enable-Exitcode
0, healthy `my-neko/brave:hls-a7ffb8b13448`, a private enable snapshot and all
17 public plus two cleartext-denial probes. The application remains exact
`a7ffb8b13448a8329c6df24fdcb182ac32ca398c`, with the unchanged deployer blob
`c6f52dc80fdf605ec908f3fe3856ce23e015e494` and reviewed Caddy configuration.

The operator then reported first HLS picture and the compact healthy indicator.
After approximately 30 seconds the client showed "HLS playback did not become
ready; retry manually". Retry HLS restored playback, and the operator explicitly
confirmed WebRTC continued working. This first-picture result differs from
the previous bootstrap failures; it does not close sustained playback, A/V,
event-triggered device failures or the earlier all-stream outage investigation.

The new client message comes from a startup deadline that static inspection
found was never cancelled after `canplay`/`playing`. A brief buffer underrun
at that deadline could stop already-established playback. Client-only repair
`73d5ff6d29110e3dd06999726a7e88718d09ea23` cancels it on current-player readiness,
arms it before attachment and retains the separate never-ready/progress-stall
bounds. Meaningful native/MSE/lifecycle regressions and a pinned isolated
old/new client/type/build helper are prepared; no server/codec/dependency
change is included. See [the client repair record and next exact block](HLS_CLIENT_READINESS_REPAIR_2026-10-05.md).

That read-only/client gate subsequently passed, as recorded below. The live
service stayed unchanged. The client repair is **NOT EXECUTED IN CODEX**;
scoped image preparation subsequently passed as recorded below; new image
deployment and sustained/grouped acceptance remain pending.

## Client-readiness A/B/type/build and live diagnosis passed — 2026-10-05

The supplied read-only diagnosis at a7ff returned exit 0 with a healthy running
image ID `sha256:e3517c887622e04065a7fec5fae1902c03e2e4e5470946915b9b99457e2992b2`,
zero restarts, OOM false and no sampled Neko exits/fixed error markers in 977
application log lines. Both HLS leases opened and later closed; one generation
became ready, all four tracks published parts/parents, and 486 segment requests
succeeded before the logged idle-grace stop. The capture showed no active
leases/packagers/objects. These cumulative counters span participants/retries;
they do not show a browser buffer state or establish sustained playback.

The isolated client helper reproduced the old startup-deadline assertion and
then passed all 52 repaired tests, TypeScript and the Vite build, with final
Client-Check-Exitcode 0. Evidence is the private a7ff subdirectory
`client-readiness-20261005T171735141870202Z`. The application checkout/images/live
service were unchanged. Native/MSE/lifecycle regression coverage supports the
controlled client correction; no repaired image is deployed yet.

The [scoped client-only image preparation](HLS_CLIENT_READINESS_REPAIR_2026-10-05.md#completed-target-block-client-only-image-preparation)
uses application 73d5ff6d and independent tooling 4a957f3e. Exact client success
and unchanged a7ff backend source evidence are recorded separately; no repeat
Go/fuzz/codec result at 73 is claimed. The running enabled a7ff service remains
during image assembly. That image preparation subsequently passed as recorded
below. New deployment and sustained/grouped acceptance remain pending.
**NOT EXECUTED IN CODEX; supplied diagnosis/client gate passed.**

## Scoped client repair images prepared — 2026-10-05

The complete supplied target output from tooling 4a957f3e ended with CLIENT
REPAIR IMAGE GATE PASSED / Client-Image-Exitcode 0. The checkout advanced to
`73d5ff6d29110e3dd06999726a7e88718d09ea23`, while the enabled a7ff container
and actual image stayed unchanged through both builds and the private snapshot.

Vite 6.4.3 rebuilt the client as `index-CrHQRMnq.js`, matching the passing
isolated client gate. Unchanged server build/plugin/runtime layers were cached;
backend tests/codec/fuzz evidence is inherited separately from a7ff, with no
fresh backend test or compilation claim. Both `my-neko/base:hls-73d5ff6d2911`
and `my-neko/brave:hls-73d5ff6d2911` were prepared. The Brave layer installed
1.96.61. Prepared image IDs, scope records, successful marker and private snapshot
are in `/opt/docker/nekoNew/neko-hls-results-73d5ff6d2911`.

NEXT the [exact-73 read-only playback diagnosis](HLS_CLIENT_READINESS_REPAIR_2026-10-05.md#next-target-block-exact-73-read-only-playback-diagnosis).
The existing deployer validates the preparation/image IDs before stopping Neko
and saves a7ff as the new rollback target. The default-off deployment subsequently
passed as recorded below. Preserve the old evidence/tags. Repaired live playback
and grouped acceptance remain pending. **NOT EXECUTED IN CODEX; supplied scoped
image preparation passed.**

## Exact-73 default-off deployment/browser checkpoint passed — 2026-10-05

The supplied baseline identifies application
`73d5ff6d29110e3dd06999726a7e88718d09ea23`, unchanged deployer blob
`c6f52dc80fdf605ec908f3fe3856ce23e015e494` and healthy image
`my-neko/brave:hls-73d5ff6d2911`. A private baseline snapshot was recorded,
disabled bootstrap/media each returned the expected 404 (2/2), and
Baseline-Exitcode was 0. The operator confirmed the requested normal browser
check works without the HLS override. This is bounded reported browser evidence.

The running image now matches checkout 73d5ff6d with HLS disabled. The prior
a7ff image is saved as `my-neko/brave:rollback-hls-73d5ff6d2911`; keep older
evidence and tags. Same-image enablement/19 denial probes subsequently passed
as recorded below. **NOT EXECUTED IN CODEX; supplied default-off deployment,
disabled probes and reported normal-browser checkpoint passed.**

## Exact-73 activation and startup-recovery report — 2026-10-05

Same-image conventional-HLS activation passed with Enable-Exitcode 0, healthy
`my-neko/brave:hls-73d5ff6d2911`, private enable snapshot and 17 public plus two
cleartext-denial probes. The operator tested on PC/Helium and reports initial
Retry, one frozen picture requiring reload, then functioning HLS while WebRTC
continued working. The follow-up confirms a "HLS failed" message but does not
supply its detailed error or timing; the report describes startup difficulties.

The estimated 20–30-second content lag is not measured latency evidence.
Conventional packaging requires three six-second parents and advertises
HOLD-BACK=18, explaining warm-up/buffered delay without establishing the cause
of the terminal message or frozen picture. The readiness/stall clocks are
separate. No timeout/buffering/GOP change or LL-HLS switch is justified yet.

Checkout/live image remains 73d5ff6d, now with HLS enabled. NEXT [read-only
playback diagnosis](HLS_CLIENT_READINESS_REPAIR_2026-10-05.md#next-target-block-exact-73-read-only-playback-diagnosis)
without changing the currently working service. First-start/recovery, exact
audio/room-event evidence and grouped acceptance remain open. **NOT EXECUTED
IN CODEX; supplied healthy activation/19 probes passed, later playback after
recovery reported.**
