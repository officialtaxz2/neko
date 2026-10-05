# Integrated stability review — 2026-10-04

Scope: `testing` baseline `3f1a62b0d12594f1dfc11a6c0064ae4d69116660` plus the
Phase 4 assets and repairs in this implementation block. The final exact commit
and image IDs must be recorded by the target-server helpers. Source/configuration
inspection only: **NOT EXECUTED IN CODEX** for tests, type/build, containers and
devices. This records items 1–7 of [STABILITY_REVIEW.md](STABILITY_REVIEW.md);
item 8 is now recorded in [the dependency review](DEPENDENCY_AUDIT_2026-10-04.md)
against the supplied exact-e55 report. Its follow-up source repairs passed the
exact target automated/image gate at `93f1fa63`; browser/deployment validation,
remaining package remediation and final security acceptance stay open.

## Confirmed defects and bounded repairs

| Finding | Evidence and repair | Required target check |
| --- | --- | --- |
| Optional chat audio could throw during room events | `chat.newMessage` called `.catch` directly on `Audio.play()` and did not contain synchronous constructor/play failures. `notification-sound.js` guards missing/legacy APIs and catches optional audio failures. Sound remains one element per notification; no encoder or transport change. | New notification test plus real join/chat/control sequence, sound on/off on an affected TV when available |
| Long HLS private pause could terminate a valid player | Both lease authentication and request admission rejected keepalive while paused. The client renews at 15 seconds and treats repeated keepalive errors as terminal; the lease expires after 30 seconds. Only nonblocking keepalive can now renew a valid paused lease and its cookie. Media admission remains denied; expired/revoked leases cannot revive and rate limits remain. | New Go paused-renewal/controller/rate tests and 46-second client pause test; real >=45-second private pause, packager idle, resume and revocation |
| Server option text incorrectly said no client player existed | Phase 3 already supplies the isolated player. The HLS enable flag description now describes experimental shared passive delivery. | Build/config inspection; no behavior change |
| Participant-controlled chat reached Vue's template compiler | Markdown output was concatenated into `template`; HTML escaping does not neutralize Vue expressions. Render fixed-element escaped HTML instead, escape custom attribute values and replace generated emoji directives with native title labels. | New real-parser/Vue-component tests, type/build and browser formatting/emoji/spoiler/open-in-app checks |
| Member names reached an HTML tooltip | v-tooltip 2.1.3 defaults to HTML; `members.vue` supplied participant names without an override. Set `html: false` for that binding. | Verify a harmless markup-shaped name renders literally in another viewer's tooltip |
| Shared Axios retained unneeded XSRF cookie reading | Browser Axios 1.2.6 can automatically read a named cookie for same-origin/credentialed requests. Neko has no use of that mechanism; disable its cookie name when installing the shared instance. | File upload/download and About checks; no automatic XSRF header on external requests; remaining dependency advisories stay open |

MDN documents that older `play()` implementations may return no value:
[HTMLMediaElement.play](https://developer.mozilla.org/en-US/docs/Web/API/HTMLMediaElement/play).
This establishes an API/error-handling defect, **not** the cause of the reported
TV disconnects. Repeated audio allocation remains a device-investigation
candidate. No speculative pooling, bitrate/buffer changes or WebCodecs removal
is included.

## Inspected paths and implications

### Follow-up after the failed enabled HLS playback attempt

The supplied read-only diagnosis from helper `d191b8ea` passed at application
`93f1fa63`: the prepared image was healthy, metrics responded, and the bounded
Supervisor sample showed no Neko exit/OOM. HLS startup/source-restart counters
were present but no readiness/lease-open success was demonstrated. Normal login
also timed out; the operator tentatively reported `/ws` HTTP 101. That would
prove upgrade, not login/session initialization. The post-upgrade blocker remains
unconfirmed; the earlier invalid-input/activation passes do not prove playback.

The [repair record](HLS_STARTUP_REPAIR_2026-10-04.md) separates source defects
from target conclusions. Bounded repository repairs bind workers to the opened
capture generation, tolerate only the initial exact same-generation caps
completion, map HLS encoder PTS/DTS through their output segment, prevent initial
videorate gap filling and replace the unsafe 100-byte `vsprintf` logger with
bounded `vsnprintf`. No crash attributable to that logger was observed. Ordinary
event-socket writes and legacy loopback requests still have unbounded waiting
paths; they are investigation candidates rather than established causes.

The target automated/image gate passed at exact 80020d99 (Check-Exitcode 0):
all 13 selected Go packages, both fuzz jobs, server/plugin build, real-codec
segment mapping and cold VP8/Opus input through the production H.264/AAC
packager succeeded. With GStreamer 1.26.2, all four renditions reached
conventional readiness in one generation after 18.11 seconds. The client build
and both image builds completed; the supplied excerpt begins after the client
test details, so their test count is not repeated here. Audit exit 1 remains
open findings, not a security pass. These are supplied target results,
**NOT EXECUTED IN CODEX**. Synthetic cleanup discarded two final samples after
their pipelines had been removed; the test passed and this is not evidence of
a deployed capture loss. Production capture skew, valid auth/proxy/player,
device/resource and grouped acceptance remain separate gates.

The saved pre-HLS image was restored healthy with Restore-Exitcode 0 and
operator-confirmed normal login/picture/audio. The prepared 80020d99 image then
passed default-off deployment with helper a7669dd1 (Baseline-Exitcode 0,
healthy service, 2/2 disabled-route probes). The operator confirmed normal
login, picture, audio and control work in the requested private browser check.
This is supplied target evidence, **NOT EXECUTED IN CODEX**. Same-image HLS
activation then passed with Enable-Exitcode 0, healthy service and 19/19 HTTP
denial probes. The operator again reported HLS bootstrap failure and all
streams stopping afterward, tentatively recalling a brief initial picture.
No actual process crash trace or cause is supplied. The next step at that
checkpoint was read-only diagnosis before restoring the same known-working
image without HLS. Its subsequent results are recorded below; enabled
HLS/device/resource acceptance remains open.

The subsequent diagnostic/recovery passed (both exit 0): healthy 80020d99
container, no Docker restart/OOM or sampled Neko exit, one successful HLS
bootstrap/lease and 23 successful segment requests, followed by two fixed
timeline-gap rejections. The same image returned healthy to default-off and
passed 2/2 disabled-route probes; fresh browser confirmation is pending. The
bounded HLS-only scenecut=0 correction subsequently passed the isolated target
GOP A/B gate at 97ba4ad9 on 2026-10-05. Old code reproduced two timeline gaps
and failed readiness (expected negative exit 1); all three repaired codec tests
passed, including 30.19 seconds of scene cuts in generation 1, with overall exit
0. This verifies the fixed-boundary defect/correction in the bounded fixture;
its role in the all-stream symptom remains a hypothesis pending live isolation
acceptance. Full exact 97ba4ad9 checks/image preparation subsequently passed:
47 client tests, TypeScript/build, 13 Go packages, both fuzz jobs, all three
codec tests and server/base/Brave builds, with Repair-Check-Exitcode 0. Audit
exit 1 remains open findings. Default-off deployment of hls-97ba4ad9ab3e then
passed with Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes.
The operator reported normal login/picture/audio/control work. The subsequent
HLS attempt failed with `HLS bootstrap failed; retry manually`, and the operator
confirmed that WebRTC also stopped working. Latest activation CLI/HTTP outcomes
are not supplied. Read-only diagnosis then passed with 84 log lines, one
negotiation rejection and one not-ready bootstrap, but no sampled process
exit/OOM or generation/lease-open marker. Same-image restoration returned the
prepared image healthy without HLS, with Recovery-Exitcode 0 and 2/2 disabled
probes; the operator confirmed normal login/picture/audio again. Saved analysis
subsequently passed with exit 0: audio/high/medium subscriptions persisted while
low reached only its capture-create marker and remained at generation 0. Static
inspection confirmed a global registry mutex held during native construction,
also acquired by every media sample callback. Its scope is narrowed to registry
insertion; fixed capture-phase tracing and a target-only regression/A/B helper
are prepared. The isolated gate at 48f4acf2 returned exit 1: the expected old
mutex-wait failure and corrected registry pass were observed, and timestamp
mapping and scene cuts passed, but smooth cold readiness failed at 24.02 s.
The paired diagnosis at helper 88f2b25d then completed with exit 0: both
constructors passed two starts and failed one. Each failure left only medium
with two parents (MSN 2/3), after its first 30 s IDR was observed before the
high anchor; the others had three parents. This supports the pre-anchor sample
loss separately from registry isolation. The bounded first-output anchor wait
then passed the isolated 414639d2 A/B gate: old initial-IDR loss reproduced;
all seven corrected checks passed three times, including cancellation/overflow
and all six real-codec fixtures in generation 1 (Anchor-Check-Exitcode 0).
Full exact-414639d2 preparation then FAILED with Repair-Prepare-Exitcode 1:
47 client tests, type/build, 13 Go packages, both fuzz jobs, registry/mapping
and anchor lifecycle checks passed. Smooth readiness restarted with
worker_failure and failed its generation-1 assertion at 20.03 s; scene cuts
passed at 30.19 s in generation 1. New base/Brave image steps were not reached.
The 53034495 worker diagnosis then passed all six checks in three fresh
processes: each smooth fixture ready at 18.11 s in generation 1, without
rejected pushes or a real-codec restart. Intentional overflow/closure warnings
belonged to their unit cases. The prior worker failure was not reproduced.
Static review found that waiting for high also stops AAC drainage; the narrow
repair keeps AAC flowing through ordinary pre-anchor discard while preserving
the initial-video-IDR hold. The controlled audio-anchor A/B at 71a14d21 passed
with Audio-Anchor-Exitcode 0: old blocked drainage and audio/anchor/queue_full
reproduced under 256 ms injected high-input delay; all eight corrected checks
passed three fresh processes and scene cuts passed once (25 top-level passes).
All seven real-codec fixtures stayed in generation 1, with no rejected pushes.
This verifies that controlled AAC-overflow mechanism, not the earlier
unobserved worker restart or live all-stream outage. No preparation marker or
live service changed in that A/B. Full exact-71a14d21 preparation then passed
with Repair-Prepare-Exitcode 0: the rebuilt GStreamer 1.26.2 image passed all
nine selected startup checks (normal 18.11 s, delayed-high 18.09 s and scene
cuts 30.19 s, all generation 1), and base/Brave images were built. The supplied
tail omits earlier client/Go/fuzz output; the script's reported final success
covers those stages without separately shown counts. The service stayed
unchanged during preparation. Default-off 71 deployment then passed with
Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes. The operator
then confirmed the requested normal-browser check works without HLS. Same-image
conventional-HLS activation then passed with Enable-Exitcode 0, healthy service
and 17/17 public plus 2/2 cleartext-denial probes. The subsequent HLS attempt
failed after connecting with "HLS bootstrap failed; retry manually"; the operator
reported only WebRTC streaming works. Read-only diagnosis passed with 950 lines,
one not-ready bootstrap, medium/low admission drops of 759/564 and cumulative
part/segment publication only for audio/high; no sampled Neko exit/OOM was found.
Same-image default-off restoration passed with a healthy service and 2/2
disabled-route probes; the operator confirmed normal login/picture/audio.
The isolated source-phase diagnosis reproduced not-ready in all three cold
runs at 24.02 seconds in generation 1; aligned control passed at 18.11 seconds.
Common high-source fan-out is implemented and statically reviewed, preserving
provider timestamps and the fixed limits. NEXT its isolated repair A/B with
the confirmed default-off exact-71 service unchanged. No successful HLS
picture/audio or room-event interval is demonstrated, and live IDR phases remain unmeasured.
New repair/gates: NOT EXECUTED IN CODEX; supplied isolated A/B, full preparation
and default-off deployment/browser plus activation/invalid-input gates passed;
enabled HLS failed, diagnosis/recovery and controlled reproduction passed;
repair checks/full preparation and enabled acceptance remain pending.
The actual native blocker and role of the separate capabilities rejection
remain unconfirmed; enabled live playback and isolation/device
acceptance are pending. Dependency-audit report exit 1 is not a security pass.
These are supplied target results, **NOT EXECUTED IN CODEX**.
See [the repair record](HLS_STARTUP_REPAIR_2026-10-04.md). Keep HLS disabled after recovery.

1. **Room events and stores:** traced member list/join/disconnect, room chat and
   control take/release/grant in `client/src/neko/index.ts` and the user/chat/
   remote stores. The inspected ordinary handlers update room/UI state and
   notifications; they do not directly recreate the peer or detach the selected
   receive backend. Socket-generation guards reject stale callbacks. Actual
   old-device CPU/audio/UI interactions remain unmeasured.
2. **Startup/playback:** inspected login/init, selected-backend peer creation,
   WebRTC readiness/recovery and HLS advertised-mode/ticket/bootstrap/readiness.
   HLS uses native Apple playback or the lazy pinned MSE adapter with exact
   codec checks. There is no automatic fallback. Autoplay retains bounded muted
   retry and the Play gesture. Connection readiness is not first-frame evidence;
   measure the reported mobile startup independently.
3. **Recovery and ownership:** inspected event-socket/peer identity guards,
   WebCodecs worker/socket/audio lifecycle, HLS timers/abort controllers/player
   generations and private-mode buffer clearing. The long-pause defect above
   was the concrete repair. Selected tests exercise stale bootstrap, late Play
   rejection, revoke and teardown; native decoder release/background return
   still require actual-device validation.
4. **Authorization:** inspected common `MediaManager.Open`, exact live
   `CanWatch` session rechecks and lifecycle revocation, HLS security/ticket/
   cookie/request boundaries and the event-level view-only separation. Only
   keepalive renewal is allowed during private pause; no media or control right
   is granted. Repeat the full role/denial matrix and one-second revoke gate.
5. **HTTP and proxy:** inspected original peer preservation, prefix stripping
   before CORS/log classification, strict Origin/forwarded-TLS policy, scoped
   cookies, fixed path redaction and bounded HEAD/gzip/range/object writes.
   Caddy is host-managed; the supplied minimal site and version 2.6.2 are reviewed,
   with active path `/etc/caddy/Caddyfile` confirmed; the proposed configuration
   and actual log output still require target validation. The
   new guide deliberately treats access and runtime/error logs separately and
   avoids unconditional negative flush intervals that prevent cancellation.
6. **Bounds/isolation:** inspected the shared four-worker packager, exact-source
   geometry/format/generation admission, 64-unit provider queues, eight-event
   handoff, generation reset on overflow, memory retention, per-lease/process
   HTTP limits and idle release. Viewers share packaged objects; request writes
   use bounded deadlines without holding the publication lock. Those bounds do
   not establish target CPU/RSS, acceptable mux behavior or healthy-viewer
   isolation; all remain measured acceptance gates.
7. **Configuration/deployment:** the explicit HLS overlay leaves base defaults
   unchanged and preserves adaptive/WebCodecs. Reviewed nested origin/proxy
   fallback, exact unique image tags, pre-stop rollback capture, private
   evidence paths, shell-input quoting and secret-safe metadata output. Viper's
   pinned cast dependency splits environment StringSlice on whitespace, so
   both modes use `hls ll-hls`. Dependencies/encoder profiles remain unchanged.

The resulting source/configuration diff was inspected again before commit,
including the paused-media denial and preserved keepalive rate/expiry bounds,
private output paths and pre-stop rollback behavior. Passing repository review is not target compatibility or full security
acceptance. The new parser fuzz job is bounded request-boundary coverage, not
proof of complete mux/player fuzz coverage.

## Open evidence and next checkpoint

- The supplied target-server output at exact `e55bcd7e` passes all 44 client
  tests, TypeScript/build, all 13 selected Go packages, both 30-second fuzz jobs,
  server/plugin build and uniquely tagged base/Brave image builds. It includes
  the notification and long-private-pause regression tests. Real enabled HLS
  playback and live private-mode/revocation/device checks remain pending.
- The earlier server gate passed 42 client tests and relevant Go/build checks
  at `e85d8568`; images/deployment health passed at `741025c3`. The operator
  subsequently reported that things seem to work, without an exact
  backend/device/role matrix. This is bounded smoke evidence only.
- The target run again reports 20 npm findings, including one critical, and
  saves an audit report with exit code 1. The complete report is now supplied
  and classified in [the dependency review](DEPENDENCY_AUDIT_2026-10-04.md).
  The critical entry is the Axios Node multipart dependency, with no deployed
  Node request path found; browser/compiler findings required the follow-up
  source containment above. No dependency versions changed, and no
  dependency-security clearance follows from the image gate. The follow-up
  code passed the follow-up 47-test/type/build/image gate at exact `93f1fa63`;
  browser acceptance and deployment remain pending.
- Exact `93f1fa63` also passed all 13 Go packages, WebSocket fuzz (1,117,452
  executions), HLS request fuzz (318,735 executions), server/plugin build and
  both image builds. The running service was unchanged. Both active/proposed
  Caddyfiles validated successfully, without reload or actual log evidence.
- The default-off public HTTP follow-up passed bootstrap/media 404 checks
  (2/2, exit 0) at exact e55, and supplies both saved image IDs. This does not
  verify the repaired client or any enabled HLS media path.
- The colleague's TV remains unavailable. Exact VIDAA/device compatibility,
  event-triggered failure classification and any benefit of HLS are unknown.
- Numeric latency/pacing, independent fMP4 validation, mixed-backend slow-client
  isolation, resource cleanup, live hostile inputs and the accumulated role/
  recovery matrix remain pending. Use [the runbook](HLS_LL_HLS_VALIDATION.md)
  and commit only a sanitized result with explicit omissions.

WebRTC remains the default; WebCodecs remains explicit; HLS starts default-off
without its overlay. No `master` promotion is authorized.
