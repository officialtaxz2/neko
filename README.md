# officialtaxz2/neko

Customized fork of [m1k1o/neko](https://github.com/m1k1o/neko), a self-hosted shared virtual browser/desktop streamed to multiple participants.

This fork keeps Neko's shared multi-user session model and adds substantial client-side work around mobile/touch use, trackpad-style control, playback/reconnect recovery, and UI/UX.

## Status

Comparative architecture assessment (2026-10-07):
[source findings and media-path comparison](docs/STABILITY_REVIEW.md#comparative-fork-and-transport-review--2026-10-07),
[current upstream triage](docs/UPSTREAM_SYNC_AUDIT.md#read-only-upstream-comparison--2026-10-07)
and [prioritized plan with target checks/rollback](docs/WORKPLAN.md#next).
The operator authorized implementation of the improvement/comparison plan:
improve core WebRTC, retain opt-in
conventional HLS, freeze WebCodecs expansion pending comparative benefit, and
defer LL-HLS unless its latency niche is needed. The independent correctness,
event-socket isolation, WebRTC playback/recovery and configuration cleanups are
implemented and statically reviewed. Target regressions/device/resource checks
remain **NOT EXECUTED IN CODEX** and will be run by the operator. No optional
transport was removed, no broad upstream merge or promotion was performed.
The [current implementation/runbook](docs/WORKPLAN.md#implemented-block-and-target-handoff--2026-10-07)
includes a preparation-only helper and a bounded private browser trace for the
reported Hisense VIDAA/Odin event-correlated WebRTC failures; their cause remains
open. Desktop, Smart-TV and iPhone support are required.

The supplied exact-8d preparation **passed with Prepare-Exitcode 0** after the
exact-f03 capture event-order failure. The helper confirms the old defect,
100 repaired subscription race repetitions and backend/server/image gates;
80 client tests/type/build are inherited from unchanged exact-f03. Subsequent
baseline activation **passed with Baseline-Exitcode 0**: healthy exact-8d image,
private snapshot and 2/2 disabled-HLS HTTP probes. The live service is exact-8d
with HLS off; preserve the previously accepted exact-8741 image/evidence. The
[next handoff](docs/WORKPLAN.md#implemented-block-and-target-handoff--2026-10-07)
is same-image HLS enablement/denial probes and parallel PC HLS/WebRTC playback.
The operator reports general PC operation works; the affected VIDAA/Odin TV is
unavailable, so its device/event gate remains unverified. New HLS/resource and
full grouped device acceptance remain open. Keep the target
at its prepared exact-8d commit; no pull, rebuild or repeated preparation is
needed. No local project execution took place in Codex.

Latest HLS checkpoint (2026-10-07): the complete browser trace captures a freeze
at about 24 seconds of media despite continued video delivery. Two source defects
are repaired: rolling playlists now preserve segment discontinuity numbers;
seek-only time jumps no longer hide a playback stall. Exact-8741 preparation
subsequently passed with **Prepare-Exitcode 0**: both old defects reproduced,
66 client tests/type/build, HLS wire/package checks and candidate images passed.
Exact-8741 activation then passed with **Activate-Exitcode 0** and a healthy
`my-neko/brave:hls-8741f7880d9a`. The operator confirmed at least five minutes
of moving HLS picture/audio without Retry/reload while WebRTC continued working.
This bounded PC/Helium playback gate is passed; [remaining grouped/device
acceptance](docs/HLS_PLAYLIST_WINDOW_REPAIR_2026-10-07.md#remaining-acceptance)
stays open. These are supplied target/browser results, **NOT EXECUTED IN CODEX**.
No additional server command or rebuild is needed for this checkpoint.

Historical HLS source review (2026-10-05): fixes paused-time stall accounting,
monitoring before first readiness and mixed HTTP/readiness error budgets, plus
player-event handling. The supplied exact-68 target gate passed **all 60 client
tests, TypeScript and build**. Scoped exact-68 image preparation also passed
(`Image-Prepare-Exitcode: 0`). Exact-68 activation then passed healthy with
`Start-Exitcode: 0`; checkout and running service were then exact `68dbdd4a`.
The operator reports HLS works after Retry. Reliable first start and wider
browser acceptance remain open. Read-only diagnosis then passed: one server
not-ready bootstrap, one successful bootstrap and generation-1 delivery,
without a sampled crash. The saved timing summary then passed: one request
took approximately 24 seconds, the other at most 1 ms. A bounded server-only
candidate raises conventional startup allowance from 24 to 28 seconds inside
the existing 30-second HTTP/client limits. This is not yet a verified fix.
Exact-8f54970f focused target preparation subsequently passed with
`Prepare-Exitcode: 0`: native checks and base/Brave images passed. Server
checkout is exact-8f54970f; the live image was still exact-68 at preparation.
After the requested activation, the operator reports a first picture that
appeared frozen; reload then produced moving HLS video/audio. Reliable first
start remains open. The subsequent read-only diagnosis passed with
`Diagnostic-Exitcode: 0`, matching prepared/live exact-8f image IDs, a healthy
container, two successful bootstraps, all four workers in one generation and
135 successful segment requests; no sampled exit/OOM or fixed error was found.
Browser first-picture state remains unknown. A client-only candidate now
requests autoplay after attachment rather than waiting for canplay; it is not
a confirmed live fix. Exact-7dcc3c5e client/image preparation subsequently
passed with `Prepare-Exitcode: 0`: both images and private evidence recorded,
with the live exact-8f service retained. Earlier client tests/type/build are
covered by final helper success; their counts precede the copied tail.
After the requested exact-7dcc activation, the operator reports moving first
picture after about 20 seconds, followed by a frozen picture requiring reload.
The subsequent read-only progress diagnostic passed, confirming the prepared
exact-7dcc live image and a healthy service. It captured an idle HLS backend:
zero leases/workers/requests; the operator confirmed HLS was closed/logged out.
This does not diagnose the earlier freeze or establish a WebRTC-join cause.
The corrected `4593a6f9` progress helper subsequently passed on the target:
one active lease, all four workers and continuing audio/video publication and
successful HTTP media requests over 12.011 seconds, with no interval restart.
VP8 rows are retained. The operator confirms this sample was taken during frozen
picture/audio loss without reload, followed later by "HLS HTTP connection failed;
retry manually". The freeze is captured despite server flow; its cause and the
later HTTP failure remain unresolved. The supplied browser timing summary shows
one 1,001-ms master with no response metadata, but a full 250-entry buffer omits
the later failure sequence. Status-zero entries with body data are not failures.
The bounded browser trace was subsequently supplied in full; the current repair
and focused verification scope are linked above. Prior results are supplied
target evidence; no speculative timeout change is made.
Reliable first start and sustained playback acceptance remain open.
See the
[focused review and grouped next checkpoint](docs/HLS_CLIENT_STABILITY_REVIEW_2026-10-05.md).

### IMPLEMENTED

Repository implementation is listed here independently of runtime validation. The latest upstream integration is statically reviewed and the operator has confirmed that all applicable target-server build and regression checks passed after correcting the Brave policy mount filename. The opt-in adaptive profile was separately built, tuned and accepted on the target server on 2026-09-10 for the documented three-viewer scenario. These acceptances apply to the tested deployment; they are not universal device-support claims.

- Shared Neko browser/desktop session with multi-user access.
- Existing Neko admin/user/control semantics.
- Fork-specific client redesign.
- Touch/mobile controls and trackpad mode.
- Mobile keyboard/helper integration.
- Autoplay/muted fallback.
- Video stream health/recovery logic.
- ICE disconnect/recovery handling.
- A bounded application-level reconnect after an established peer cannot recover: four serialized attempts at 1/2/5/10-second delays, suppressed for initial-login failure, explicit logout, demo mode and server-directed disconnects.
- Stale WebSocket/peer/data-channel callbacks and old stream timers are isolated from a replacement connection; Safari's central Play fallback remains separate from network recovery.
- Optional server-enforced view-only sharing through the compact `#/<16-character-Base64URL-token>` route: the 96-bit passive bearer remains outside normal HTTP request targets, while the session stays in the same room and receives WebRTC media and every interactive boundary remains denied.
- View-only authorization is represented by a backend-neutral session marker; share sessions are not persisted, and token rotation/removal plus service recreation is the explicit revocation boundary.
- Pure encoded-media descriptors/events, a capture-backed provider with bounded source subscriptions and a central participant-delivery registry now separate shared capture demand from per-session transport delivery.
- The existing WebRTC sender is the first delivery backend behind that boundary. Its signaling, data channels, adaptive selection, effective two-sample drop-new queue, per-session metrics, authorization and deployment defaults remain unchanged by design.
- GStreamer PTS/DTS, format metadata, source generations, keyframe admission and explicit discontinuities are carried through the new provider contract; published media buffers are immutable and slow subscriptions drop locally.
- Generic session watching state is owned by the active participant delivery rather than by WebRTC-named lifecycle code; backend leases expose no login/share credential.
- The default-off `webcodecs-ws` prototype now provides Phase 1 strict version-1 framing, fixtures, PTS-validity propagation, authenticated negotiation and hashed single-use tickets; the Phase 2 credential-free `/api/media/ws` delivery backend; and the Phase 3 client path with a dedicated parser/socket/decoder worker, bounded VP8 canvas and Opus AudioWorklet presentation, common-clock recovery, visible receive-only limits and explicit same-backend retry/WebRTC actions. The client now also persists a manual per-browser WebRTC/WebCodecs default in sidebar settings, safely defaults absent/invalid values to WebRTC, retains exact `?media=webcodecs-ws` as the highest-priority stateless diagnostic override and shows only `WebCodecs` in the compact healthy indicator. Phase 4 adds a separate sanitized enablement overlay, credential-safe fixed-metric/PromQL collection and an interactive grouped-validation record. The route is registered only with the explicit server overlay; new/unset clients retain the unchanged WebRTC path.
- The exact default-off HLS/Low-Latency HLS passive/view-only version-1 contract is fixed in [`docs/HLS_LL_HLS.md`](docs/HLS_LL_HLS.md). Phases 1–3 now implement validated configuration and negotiation, digest-only tickets and sliding path-scoped cookie leases, the shared one-audio/three-video bounded packager, H.264 High 3.1/AAC-LC fMP4 output, deterministic conventional/LL-HLS playlists, conditional authenticated HTTP delivery, central per-session lifecycle and fixed credential-safe metrics. Phase 3 adds the isolated passive native/MSE player with pinned local hls.js, explicitly advertised HLS/LL-HLS choices, private-mode state and bounded lifecycle cleanup. The focused automated target-server gate passed at `e85d8568`, and the default-off base/Brave image deployment started healthy at `741025c3`; HLS-enabled playback/device acceptance remains pending. Phase 4 now supplies a separate opt-in overlay and private target-server validation/rollback assets; the exact Phase 4 automated/image preparation passed at `e55bcd7e`, while enabled checks remain pending and automatic selection remains absent. The exact bounded evidence is recorded in [`docs/WORKPLAN.md`](docs/WORKPLAN.md).
- Demo-mode client infrastructure.
- Per-peer non-blocking media sample delivery so one backpressured WebRTC track does not block dispatch to other tracks.
- Multi-pipeline stream-selection and per-peer bandwidth-estimator infrastructure; the legacy protocol path requests automatic selection when configured.
- An explicitly opt-in Brave Compose overlay with ordered `high`/`medium`/`low` VP8 pipelines and explicit estimator settings; the base Compose deployment remains single-pipeline.
- Peer-local audio/video sample-drop counters plus measured pipeline-bitrate metrics for adaptive-quality diagnosis.
- Stream bitrate accounting in bit/s, matching the Pion estimator rather than comparing its bit/s target with encoded bytes/s.
- Separate downgrade-deficit and upgrade-reserve decisions: a neutral estimate or the mere absence of 15% spare no longer proves that the current tier is congested.
- Current-tier fit uses the greater of measured and nominal video rate plus measured audio and an explicit transport reserve; upgrades use the next tier's nominal delivery requirement plus their separate reserve.
- A downgrade additionally requires fresh peer-local RTCP loss or NACK evidence; the GCC target/trend alone is advisory after a reproduced loss-free false collapse.
- A clean lower tier with an outstanding automatic-downgrade recovery step can run a bounded peer-local one-tier probe when it application-limits GCC below the next-tier requirement. Failed probes return through the same evidence-gated downgrade path and use exponential capped backoff instead of rapidly oscillating.
- Estimator stable, unstable and stalled observation windows now begin together at reader startup, so configured grace periods cannot appear pre-expired on the first qualifying estimate.
- H.265 capture/WebRTC codec support.
- Per-user file download/upload/delete permissions and multi-file deletion.
- Optional server-side “open chat link in app” plugin.
- Clipboard resynchronization on browser-window focus and capture-pointer configuration.

### Integration snapshot

Current upstream integration (2026-09-09):

- pre-sync safety branch `safety-pre-upstream-sync-20260909`: `18e9320c892b4069757a71c9093c9c9b4dd7bd4a`
- upstream `m1k1o/neko:master`: `b0f01cedea68893e85a3fd852c0521238c285695`
- pre-sync merge base: `d74052bb844c43a0cc3c2386d083f7505dc483a2`
- upstream merge commit on `integration/upstream-20260909`: `4e99b8d3ca720d1f184544306820e388716ba23a`
- `master` was fast-forwarded to the reviewed integration history; subsequent `testing` work includes the sanitized Brave deployment reconciliation, the accepted opt-in adaptive-quality unit, bounded iOS recovery, server-enforced view-only sharing and the backend-neutral media-subscription design.

The 97-file upstream delta was reviewed by subsystem. Conflicts in `settings.vue`, `side.vue` and `video.vue` were resolved semantically, preserving the fork's touch/trackpad, UI and cursor/recovery behavior while accepting the upstream permissions, Open-in-App and focus-clipboard changes. A hidden demo-mode payload mismatch caused by the new file-transfer rights fields was also repaired.

The detailed record is in [`docs/UPSTREAM_SYNC_AUDIT.md`](docs/UPSTREAM_SYNC_AUDIT.md).

Local reconciliation (2026-09-09):

- safety branch `safety-pre-local-delta-audit-20260909` preserves the original fork HEAD;
- the complete supplied `MyNekoProjekt` tree was audited;
- no missing application-source delta was found;
- the desired local Brave deployment compose was reconstructed with configurable paths/settings and required external password variables;
- raw credentials, instance policy contents, browser profile and downloads were excluded;
- client lock/type consistency was repaired in `2d89027e` and passed the target-server client checks.

Do not perform a blind upstream overwrite.

The bounded iOS recovery and server-enforced view-only paths are implemented and statically reviewed on `testing`. The full view-only boundary, real inbound-media denial and revocation passed at `80eeca64`; containerized client/server checks, image build/deployment, the HTTP denial probe and the compact `#/<16-character-token>` browser smoke check then passed at exact commit `913a981e`. On 2026-09-11 the operator deliberately closed the grouped checkpoint without the manual iPhone deep test because no Safari Web Inspector/Mac was available and the automated evidence was accepted as sufficient for this deployment decision. This is not evidence that same-peer, replacement-session or bounded-exhaustion recovery works on a real iPhone without reload.

The backend-neutral media-subscription/WebRTC compatibility refactor is implemented and statically reviewed on `testing`. At exact commit `a32027d`, the expanded target-server Go suite and server/plugin build passed, the local base and Brave images built, the adaptive service started healthy, and the current-protocol media lifecycle plus the new credential-free `neko_media_*` metrics behaved as designed. At exact follow-up `e5f55bf9`, the complete suite and build passed again, fresh base/Brave images built, the service remained healthy with zero restarts, and the focused estimator-startup trace stayed on `high` for the full observation window.

The operator closed this combined checkpoint on 2026-09-12 with an explicit limitation: the ordinary/admin/view-only/private-mode/reconnect matrix was not repeated at `e5f55bf9`, and no fresh induced three-viewer `high -> medium -> low -> medium -> high` isolation run was performed. Those end-to-end checks are deferred to final grouped validation; earlier exact-commit view-only and adaptive evidence remains valid but is not represented as a rerun at `e5f55bf9`.

The earlier immediate `high -> medium` startup transition reproduced with the pre-refactor rollback image, proving it was not introduced by the media boundary. Static tracing found and corrected the inherited zero-time defect in the estimator's initial unstable/stalled observation windows. At that checkpoint a separate static motion-quality audit found no encoder change since `bfaca84e` and unchanged encoded payload forwarding into Pion; the later steady-state correction adds estimator-only reserve configuration but still does not alter the encoder. The focused `e5f55bf9` run also reported zero video queue drops. Fast scrolling or high-motion video can nevertheless look softer because the existing `high` tier is fixed-rate VP8 at about 2 Mbit/s with `max-quantizer: 63`; a controlled bitrate/QP A/B remains open before claiming a visual regression or changing the accepted encoder profile.

A later healthy-client trace showed a distinct sustained-selection defect: an active direct-UDP WebRTC session with zero reported loss, NACKs and local video drops still cycled tiers. The first correction at `2d037f39` separated downgrade deficit from upgrade reserve and passed its focused tests/build plus an initial three-minute live gate, but a longer unshaped run rejected it: the same healthy peer changed `high -> medium -> low`, stayed degraded for roughly eight minutes and recovered `low -> medium -> high` without loss, NACKs, video drops or a host shaper. The receiver-evidence revision at `2efcc6b1` then passed target-server tests/build/deployment and kept one healthy viewer on `high` for 20 minutes while its target ranged from about 1.71 to 4.57 Mbit/s. Endpoint-specific shaping moved only C through `medium` to `low`, while H remained visually good on `high`; however, after shaping was removed C remained on `low` for the complete 90-second recovery phase and returned to `high` only later. Exact recovery follow-up `ddf15cee` passed the focused server suite, fuzz/build, fresh image deployment and the shortened two-viewer gate: endpoint-specific 700-kbit/s shaping moved only C `high -> medium -> low`, H stayed on `high` with zero video-drop delta, and after restoring `fq_codel` C probed `low -> medium` at 30 seconds, completed that clean probe, then reached `high` at 60 seconds with no probe failure or restart. The operator observed frozen C playback above `low`, fluid playback on `low`, and both recovery steps visually; this closes the application-limited recovery defect for the tested path without changing the accepted encoder profile.

The first opt-in WebCodecs plus dedicated media-WebSocket receive prototype has an exact version-1 design covering credential-free ticket attachment, `CanWatch` lifecycle, VP8/Opus framing, bounded queues, A/V clocking, security limits, reconnect, rollout and target-server acceptance. Its Phase 1 protocol/ticket boundary, Phase 2 server delivery adapter and Phase 3 isolated client decode/render path were first implemented behind the default-off server flag and exact `?media=webcodecs-ws` selection; the current client additionally supports the persisted manual setting described above while retaining that query as the diagnostic override. Phase 4 repository assets provide the separate opt-in deployment and credential-safe observability/validation workflow. At exact `86893473`, all 17 client tests, TypeScript/build, the full server package sequence, fuzz/build/images/deployment/security and an uninterrupted roughly 15-minute zero-resync stream passed. The measured approximately 25 VP8 units/s match the shared adaptive `high` source's configured 25 fps rather than indicating WebCodecs transport loss; WebCodecs nevertheless appeared subjectively slightly less fluid than WebRTC. Exact `150aac57` then passed 19 client tests, TypeScript/build, fresh-image deployment and security probes; ordinary, admin and view-only receive, private-mode pause/resume, explicit WebRTC return, desktop native fullscreen and the iPhone canvas app-fullscreen fallback were operator-confirmed. During that matrix a visible foreground iPhone still issued three rapid `client_audio_underflow` recoveries, reached `resync_limit` and retried, matching the reported short black/audio interruptions. Commit `95746173` changed this to a canvas-preserving local rebuffer. At documentation commit `6b6cd328`, all 20 client tests, TypeScript/build, the complete server package suite, 849,148 fuzz executions and the trailing build passed; fresh images built, the exact Brave image deployed healthy and every public/direct security probe passed. A fresh foreground-iPhone run then delivered 16,372 VP8 and 32,792 Opus units through one uninterrupted connection for more than ten minutes with no WebCodecs close, retry, server `client_audio_underflow`, progress timeout or restart, and the operator reported picture and audio remained okay. The operator closed this as a bounded checkpoint and deliberately deferred the numeric latency/pacing, induced isolation, resource and remaining live hostile-input matrix, so full prototype acceptance is not claimed. At exact productization commit `12cfe43b`, all 26 client tests, TypeScript and Vite build passed; base/Brave images built, image `sha256:2295d719` deployed healthy with zero restarts and all public/direct security probes passed. The operator confirmed missing/invalid storage stays on WebRTC, stored WebCodecs works without a query and survives reload, exact-query precedence does not mutate the saved default, unrelated query state survives, healthy status is compact, and the disabled backend remains a prominent terminal state with no automatic fallback before explicit return to working WebRTC. Live view-only fragment preservation was not repeated; its focused automated navigation test passed. This focused closure does not expand the earlier full-prototype acceptance claim.

## NEXT

Continue exclusively on `testing` with **HLS/LL-HLS Phase 4 target-server validation**. At exact application `93f1fa63`, automated/image preparation passed; helper `2484a022` then preserved all 11 Caddy hosts, passed reload/synthetic error-log redaction, deployed conventional HLS healthy and passed 17 public plus two cleartext-denial probes. The first actual HLS playback attempt subsequently failed (connection failure/no picture). The supplied read-only diagnostic found the prepared image healthy but no successful HLS readiness/lease; normal login also timed out (tentative WebSocket 101). The saved pre-HLS image was restored successfully (Restore-Exitcode 0); the operator confirmed normal login, picture and audio work again. Exact repair commit 80020d99 passed the target automated/image gate, including both real-codec tests and four-rendition conventional readiness in one generation (test duration 18.11 seconds). Its default-off deployment passed with Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login, picture, audio and control. Same-image HLS activation passed with Enable-Exitcode 0, healthy service and 19/19 HTTP denial probes, but HLS failed and the operator reported all streams stopped afterward. Read-only diagnosis found one ready packager/lease, 23 successful segment requests and two timeline-gap rejections, with no sampled Neko exit/OOM. Default-off restoration passed with Recovery-Exitcode 0 and 2/2 disabled-route probes; fresh browser confirmation is pending. The HLS-only fixed-GOP correction passed the isolated target GOP A/B gate at 97ba4ad9: old code reproduced two timeline gaps and all three repaired codec tests passed, including 30.19 seconds of scene cuts in generation 1. Exact 97ba4ad9 automated/image preparation passed with Repair-Check-Exitcode 0 (47 client tests, 13 Go packages, both fuzz jobs, all three codec tests and base/Brave builds). Default-off deployment of my-neko/brave:hls-97ba4ad9ab3e passed with Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes. The operator reported the requested normal browser check works. The subsequent HLS attempt at the prepared 97ba4ad9 checkpoint failed with "HLS bootstrap failed; retry manually", and the operator confirmed WebRTC also stopped working. The latest enablement CLI/HTTP results have not been supplied. Read-only diagnosis passed with 84 log lines, one not-ready bootstrap, one negotiation rejection and no sampled exit/OOM or generation/lease-open markers. Same-image default-off restoration passed with Recovery-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login/picture/audio work again. The saved-startup summary passed with Saved-Check-Exitcode 0: audio/high/medium subscriptions persisted, low capture reached only its create marker, and the saved matching environment had HLS enabled. Paired startup diagnosis at helper 88f2b25d completed with exit 0: both constructors passed two starts and failed one, with medium losing its initial pre-anchor IDR and reaching only two parents. This narrows the smooth readiness defect independently of registry isolation. The 414639d2 isolated anchor A/B gate passed with Anchor-Check-Exitcode 0: old code reproduced the fixed initial-IDR failure; seven corrected checks each passed three times, with every real-codec fixture staying in generation 1. Full exact-414639d2 preparation FAILED with Repair-Prepare-Exitcode 1: 47 client tests, type/build, 13 Go packages, both fuzz jobs, registry/mapping and anchor lifecycle checks passed, but smooth readiness restarted with worker_failure and failed its generation-1 assertion at 20.03 seconds; scene cuts passed at 30.19 seconds in generation 1. Base/Brave image steps were not reached. Worker diagnosis at helper 53034495 then completed with exit 0: all six checks passed in three fresh processes, each smooth fixture ready in 18.11 seconds in generation 1 with no rejected pushes; the earlier worker failure was not reproduced. The controlled audio-anchor A/B at repair 71a14d21 passed with Audio-Anchor-Exitcode 0: the old AAC hold reproduced both the blocked-drainage unit failure and an audio/anchor/queue_full restart under 256 ms high-input delay; all eight corrected startup checks passed in three fresh processes and scene cuts passed once (25 top-level passes). All seven real-codec fixtures stayed in generation 1; normal readiness was 18.10/18.11/18.11 seconds and delayed-high readiness 18.09 seconds each. This verifies the controlled AAC-overflow mechanism, not the cause of the earlier unobserved worker failure or live all-stream outage. Full exact-71a14d21 preparation then passed with Repair-Prepare-Exitcode 0: the rebuilt GStreamer 1.26.2 image passed all nine selected startup checks, including normal readiness at 18.11 seconds, delayed-high readiness at 18.09 seconds and scene cuts at 30.19 seconds, all in generation 1; base/Brave images were built and the service remained unchanged. The supplied tail starts inside the codec-image build; earlier client/Go/fuzz steps are covered by the script's reported final success, not separately shown in this excerpt. Default-off deployment of my-neko/brave:hls-71a14d2174da then passed with Baseline-Exitcode 0, a healthy service and 2/2 disabled-route probes. The operator confirmed the requested normal browser check works without HLS at 71a14d21. Same-image conventional-HLS activation at 71a14d21 then passed with Enable-Exitcode 0, healthy service, a private enable snapshot and 17/17 public plus 2/2 cleartext-denial probes. The subsequent operator-reported HLS attempt at 71a14d21 went from connecting to failed with "HLS bootstrap failed; retry manually"; the operator reported only WebRTC streaming works. No successful HLS picture/audio or room-event interval is demonstrated. Read-only diagnosis passed with 950 captured lines, one not-ready bootstrap, one started/idle-stopped packager generation, medium/low keyframe-admission drops of 759/564 and cumulative part/segment publication only for audio/high. No sampled Neko exit/OOM was found. Same-image default-off restoration passed with Recovery-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login/picture/audio work again. The isolated source-clock-phase diagnosis at helper 409482b4 passed with Clock-Skew-Exitcode 0: aligned control ready in 18.11 seconds; all three skew runs reproduced not-ready at 24.02 seconds in generation 1 with audio/high ready, medium/low blocked, flowing IDRs and no rejected native pushes. This proves the controlled phase-admission defect, not the exact unmeasured live phases. A common high-source fan-out repair is implemented and statically reviewed: one shared video provider subscription feeds the existing three scaled encoders, plus one audio subscription; provider PTS/DTS are preserved. The isolated common-source repair A/B passed at a7ffb8b1 with Shared-Clock-Exitcode 0: the old defect reproduced once; 46 positive top-level checks passed across three cold processes including one scene-cut check, with all ten real-codec fixtures in generation 1. Full exact-a7ffb8b1 preparation then passed with Repair-Prepare-Exitcode 0: all thirteen selected native/startup checks passed in the rebuilt GStreamer 1.26.2 image, including the four real-codec fixtures in generation 1, and base/Brave images were built without changing the running service. The supplied excerpt starts inside the codec-image build; earlier client/Go/fuzz stages are covered by final script success without separately shown fresh counts. Default-off deployment of my-neko/brave:hls-a7ffb8b13448 then passed with Baseline-Exitcode 0, healthy service, a private baseline snapshot and 2/2 disabled-route probes; the operator reported the requested normal browser check works. Same-image a7ffb8b1 HLS activation passed with Enable-Exitcode 0, healthy service and 19/19 HTTP denial probes. The operator then reported first HLS picture and the compact streaming label, followed after roughly 30 seconds by "HLS playback did not become ready; retry manually"; Retry HLS restored playback and WebRTC continued working. Static inspection found the initial client readiness deadline was never disarmed on canplay/playing. Client-only repair 73d5ff6d cancels it on those current-player events, arms it before attachment and retains the independent startup/stall bounds; Read-only target diagnosis then passed with a healthy a7ff image, two successful HLS bootstraps and 486 successful segment requests, with no sampled process exit/OOM; the packager stopped after idle grace. The isolated target client gate passed with Client-Check-Exitcode 0: the old timer defect reproduced, all 52 repaired client tests plus type/build passed, and checkout/live service remained at a7ff. These are supplied target results, NOT EXECUTED IN CODEX. Scoped exact-73d5ff6d image preparation then passed with Client-Image-Exitcode 0: fresh client bundle index-CrHQRMnq.js, cached unchanged server/runtime layers, both base/Brave images and private snapshot/marker recorded while enabled a7ff stayed running. Exact-73d5ff6d default-off deployment then passed with Baseline-Exitcode 0, healthy my-neko/brave:hls-73d5ff6d2911, a private baseline snapshot and 2/2 disabled-route probes; the operator confirmed the requested normal browser check works without a media override. Same-image conventional-HLS activation then passed with Enable-Exitcode 0, healthy service, a private enable snapshot and 19/19 denial probes. The operator reports PC/Helium HLS playback after an initial Retry and one frozen-picture/page-reload incident; WebRTC kept working and later HLS worked normally. A HLS failed message was confirmed without its detailed error or exact timing, so startup reliability and an uninterrupted room-event interval remain unverified. Checkout/live image is now 73d5ff6d with conventional HLS enabled. Read-only exact-73 diagnosis then passed with Diagnostic-Exitcode 0: healthy service, no sampled exit/OOM or fixed error markers, one active HLS lease/all four workers running, six successful bootstraps and 394 successful segment requests. One not-ready bootstrap supports readiness as a possible initial-Retry explanation without attempt correlation; two startup-labelled generations and one idle stop do not establish a crash loop. The operator cannot confirm the exact uninterrupted interval and mentions possible random reconnects/room actions without correlation. NEXT consolidate startup/frozen-picture/room-event investigation into one bounded later validation step; no further ad-hoc operator check requested at this checkpoint; startup/recovery and grouped acceptance pending; the all-stream symptom's cause remains unconfirmed. [HLS startup source repairs](docs/HLS_STARTUP_REPAIR_2026-10-04.md) and a mandatory real-codec integration gate are implemented; their exact-commit tests/images passed at 80020d99; default-off repair-image deployment and normal browser smoke checks passed; live HLS playback remains pending. The normal-login blocker, working HLS playback, valid authorization/lifecycle and device/resource acceptance remain open. Preserve the earlier evidence. The [audit is classified](docs/DEPENDENCY_AUDIT_2026-10-04.md); dependency maintenance and final security acceptance stay open. Keep WebRTC the default, HLS default-off without its overlay, WebCodecs explicit and fallback manual. Leave `master` pinned until the operator authorizes promotion.

See [`docs/ADAPTIVE_QUALITY.md`](docs/ADAPTIVE_QUALITY.md), [`docs/IOS_RECOVERY.md`](docs/IOS_RECOVERY.md), [`docs/VIEW_ONLY_SHARING.md`](docs/VIEW_ONLY_SHARING.md), [`docs/MEDIA_SUBSCRIPTION_BOUNDARY.md`](docs/MEDIA_SUBSCRIPTION_BOUNDARY.md), [`docs/WEBCODECS_MEDIA_WEBSOCKET.md`](docs/WEBCODECS_MEDIA_WEBSOCKET.md), [`docs/WEBCODECS_MEDIA_WEBSOCKET_OBSERVABILITY.md`](docs/WEBCODECS_MEDIA_WEBSOCKET_OBSERVABILITY.md), [`docs/WEBCODECS_MEDIA_WEBSOCKET_VALIDATION.md`](docs/WEBCODECS_MEDIA_WEBSOCKET_VALIDATION.md), [`docs/HLS_LL_HLS.md`](docs/HLS_LL_HLS.md), `docs/WORKPLAN.md` and `docs/UPSTREAM_SYNC_AUDIT.md`.

## Local Brave deployment

The tracked `docker-compose.yaml` is the sanitized deployment structure from `MyNekoProjekt`. It uses the locally built `my-neko/brave:latest` with pulling disabled, cleans stale Brave singleton locks before startup, and mounts the ignored profile/download directories. The managed policy is mounted at Brave's expected `/etc/brave/policies/managed/policies.json` path.

On the target server, from the repository root:

```bash
./build my-neko/base:latest -y
./build my-neko/brave:latest -y
cp .env.example .env
# Set both passwords, an optional generated view-only token and server-specific values in .env.
docker compose config --quiet
docker compose up -d
```

Use `NEKO_POLICY_FILE=./policy.json` in `.env` only when an instance-specific ignored policy file is desired; otherwise the tracked Brave policy is used.

Adaptive quality is opt-in and requires the second Compose file:

```bash
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml config --quiet
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml up -d --force-recreate
```

Activation, diagnostics, the exact acceptance sequence, the non-destructive evidence collector, the 2026-09-10 target-server result and rollback are documented in [`docs/ADAPTIVE_QUALITY.md`](docs/ADAPTIVE_QUALITY.md). These server/profile changes were not built or runtime-tested in Codex; the documented build and runtime evidence was supplied from the real target server.

## Development environment policy

Codex is used for source editing, repository analysis and static review only. It is **not** the deployment/test environment.

Do not run the application, install dependencies, execute builds/tests/linters, start Docker, or perform WebRTC/device runtime tests inside Codex.

Runtime verification happens separately on the real server. Known server-side commands and the verification matrix are documented in `AGENTS.md` and `docs/WORKPLAN.md`.

For the completed iOS/view-only checkpoint, [`docker-compose.validation.yaml`](docker-compose.validation.yaml) supplied the client checks, focused Go tests/server build, metrics snapshots, token generation and HTTP denial probe in containers. The target host therefore needed no local Node.js/npm, Go or Python installation. The completed View-only procedure and the deliberately deferred manual iPhone phases remain in [`docs/VIEW_ONLY_SHARING.md`](docs/VIEW_ONLY_SHARING.md) and [`docs/IOS_RECOVERY.md`](docs/IOS_RECOVERY.md).

`.env`, `files/`, `downloads/` and `policy.json` remain ignored and must not be committed.

## Product direction

Target outcomes include:

- slow-viewer isolation;
- per-viewer adaptive quality;
- robust mobile/TV behavior;
- target-server validation and future alternate-media support for server-enforced view-only sharing;
- robust reconnect/recovery;
- role/capability-aware media fallback: WebCodecs/WebSocket as an interactive candidate and HLS/LL-HLS as a passive/view-only candidate.

Interactive and passive clients do not need to use the same media backend. MJPEG is only a possible ultra-legacy last resort, not a primary target.

See `docs/PROJECT.md`.

## Structure

```text
client/      Vue 2.7 + TypeScript/Vite client
server/      Go Neko server and plugins
apps/        browser/application image definitions
runtime/     runtime container support
webpage/     inherited Neko documentation site
docs/        fork-specific project/architecture/work knowledge
deploy/      tracked non-secret opt-in deployment configuration
```

## Security / deployment note

The audited `MyNekoProjekt` snapshot, including its instance-specific deployment/runtime material, was deleted at the operator's explicit request on 2026-10-04. Its desired Compose structure remains tracked in sanitized, parameterized form; active server credentials, browser profiles, downloads, cookies and lock/runtime files stay outside Git.

The completed local classification is recorded in `docs/LOCAL_DELTA_AUDIT.md`.
