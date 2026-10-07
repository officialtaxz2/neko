# AGENTS.md

Latest implementation checkpoint (2026-10-07): supplied exact
`8d8126c92903dab58056fa316b626b38dcffb376` preparation PASSED with
Prepare-Exitcode 0. The helper reports the old deterministic ordering defect,
100 repaired subscription race repetitions and backend/race/server/Base/Brave
checks passed. The supplied excerpt starts inside the base-image build; earlier
gate details are covered by helper success, not separately visible test output.
Client evidence is inherited only from unchanged exact-f03 (80 tests/type/build).
Private evidence: `/opt/docker/nekoNew/neko-stability-8d8126c92903-20261007T171646Z`.
Images `my-neko/base:hls-8d8126c92903` and `my-neko/brave:hls-8d8126c92903`
were built; the retained-live assertion and private snapshot passed. Subsequent
baseline activation PASSED with Baseline-Exitcode 0: healthy exact-8d Brave image,
private baseline snapshot and 2/2 disabled bootstrap/media 404 probes. HLS was
disabled at that baseline; adaptive/WebCodecs overlays are retained. The operator subsequently
reports general PC operation works. This is bounded desktop evidence; no exact
event/outage/Pause matrix, duration or separate per-action result is supplied.
The affected Hisense VIDAA/Odin TV is currently unavailable (belongs to a friend),
so TV causality/acceptance is deferred, not failed or passed. Keep the target
checkout pinned at exact-8d; no documentation pull, rebuild or preparation repeat
is needed. Subsequent same-image conventional-HLS activation PASSED with
Enable-Exitcode 0: healthy exact-8d image, private enable snapshot, 17/17 public
and 2/2 direct-cleartext denial probes (including required headers). HLS is now
enabled in conventional-only mode. The operator explicitly confirmed the
requested at-least-ten-minute PC interval: one HLS tab with moving picture/audio
and one concurrent working WebRTC tab, no observed freeze, Retry or reload.
This bounded sustained-playback/coexistence gate PASSED; it is operator browser
evidence, not an automated trace, a scripted event/outage matrix, numerical
latency/resource comparison, iPhone or TV acceptance. Do not repeat it or the
passed deployment/HTTP gates. NEXT B4 read-only resource baseline and comparable
per-path intervals using the existing collector; first no-viewer phase after
closing viewer tabs and settling. Keep exact-8d and all overlays unchanged;
measurements remain pending. iPhone/grouped checks and unavailable-TV gate stay open.
Preserve accepted exact-8741 images/
configuration/evidence. These are supplied target results, NOT EXECUTED IN CODEX.
VIDAA causality, grouped acceptance, comparative costs and B5/B7/B8 decisions
remain open; do not remove paths or promote master.

Historical failed implementation preparation (2026-10-07): authorized block
`f03bc4bcf68be81a76e65f9195580daad76df7b2` was published on origin/testing and
transferred to the target. Supplied preparation FAILED with Prepare-Exitcode 1
in capture `TestMediaSubscriptionTimingGenerationAndFormatOrdering`: a format
arrived instead of generation-2 discontinuity. All 80 client tests/type/build,
the selected normal Go suite and six other race packages passed; no DATA RACE
report was supplied. Base/Brave builds and final live-retention verification
were not reached. Evidence: `/opt/docker/nekoNew/neko-stability-f03bc4bcf68b-20261007T164845Z`.
No activation was requested; exact-8741 remains the last accepted live image.
An older capture format-publication ordering defect is now corrected and
statically reviewed: commit the immutable in-flight format barrier under the
dequeue mutex before its unbuffered handoff, retaining pre-selection coalescing.
Do not attribute this finding to VIDAA or the earlier HLS freeze. Follow current
`docs/WORKPLAN.md` NEXT: prepare the narrow repair with the scope-guarded
`deploy/validate-media-stability.sh OUTPUT --capture-ordering-repair`; require
the old deterministic failure and 100 repaired subscription race repetitions
plus repeated backend/race/server/image gates. Only unchanged exact-f03 client
evidence is inherited. At this earlier checkpoint repair gates were pending,
NOT EXECUTED IN CODEX. Do not
activate failed f03, repeat its failing preparation unchanged, remove paths,
or promote master. Preserve failed evidence and accepted images/configuration.

Current authorized implementation checkpoint (2026-10-07): the comparative
assessment was approved for implementation. Independent B1–B3/B6 corrections
and B4/B9 target handoff assets are now implemented and statically reviewed in
the working source; target tests/builds/runtime gates are NOT EXECUTED IN CODEX.
The operator performs target checks. Desktop, Smart-TV and iPhone are required;
Hisense VIDAA/Odin is the reported event-correlated WebRTC failure platform,
with cause/model/firmware still open. Follow current `docs/WORKPLAN.md` NEXT,
including the preparation-only `deploy/validate-media-stability.sh` and private
`deploy/trace-media-browser.js`; keep all paths until conditional B5/B7/B8
usage/resource decisions. Preserve exact-8741 live image/evidence and master
d9105ef8. Older 'no new server command' HLS checkpoint wording below applies to
unchanged exact-8741 evidence, not the newly changed common-event/WebRTC source.

Latest checkpoint (2026-10-07): the full supplied five-minute exact-7dcc browser
trace captures moving video followed by a freeze at media time 23.948 seconds,
513 frames and a fixed 12.008–24.019-second buffered range. Audio loads only four
segments while medium video continues; repeated six-second seeks hide the stall
from the old watchdog. No new terminal HTTP error was captured after Retry.
The [rolling-playlist repair](docs/HLS_PLAYLIST_WINDOW_REPAIR_2026-10-07.md)
increments track-local discontinuity bases when an earlier tag is evicted and
excludes seek-only/metadata-only time changes from playback progress. Both defects
are confirmed by static inspection; attribution of the live audio stop to the
uncaptured hls.js parser error remains an inference. Exact-8741 target preparation
subsequently PASSED with Prepare-Exitcode 0: both old defects reproduced in
isolated copies; all 66 client tests, TypeScript/build, repaired all-track/both-mode
rolling wire check, HLS package and server/base/Brave builds passed. Private
evidence: `/opt/docker/nekoNew/neko-hls-results-8741f7880d9a`.
Target checkout/live deployment is now
`8741f7880d9a709e6dc17924d2af0564a949ccd9`: activation PASSED with
Activate-Exitcode 0, healthy `my-neko/brave:hls-8741f7880d9a` and a private
enable snapshot. The operator initially reported no frozen pictures, then
explicitly confirmed at least five minutes of moving HLS picture/audio without
Retry/reload while WebRTC continued working. This closes the bounded PC/Helium
conventional-HLS sustained-playback gate, not full Phase 4 acceptance.
These are supplied target/browser results, NOT EXECUTED IN CODEX; no fresh
native/fuzz gate or causal parser trace is claimed. Preserve prior images/evidence.
NEXT consolidate remaining grouped acceptance; no new server command, repeated
five-minute check, later-doc pull or rebuild is needed at this checkpoint.
No further old-image browser trace or speculative timeout/buffer tuning is needed.
Keep cold-start/grouped/device/dependency acceptance open; master remains d9105ef8.

Historical client checkpoint (2026-10-05): the requested
[HLS client stability review](docs/HLS_CLIENT_STABILITY_REVIEW_2026-10-05.md)
corrected pause/startup progress monitoring and independent HTTP/readiness
failure budgets, with defensive player-event handling and fixed bootstrap
availability detail. The supplied exact-68 isolated target gate passed all
60 client tests, TypeScript and build (Client-Check-Exitcode 0); report
`/opt/docker/nekoNew/neko-hls-client-check-68dbdd4a-lJbw35xa`. These are supplied
server results, **NOT EXECUTED IN CODEX**. Scoped exact-68 image preparation
also passed with Image-Prepare-Exitcode 0 using helper `28d081a4`: fresh client
build, base/Brave images and private preparation evidence recorded. Backend
evidence is inherited through identical exact-73 sources; no fresh Go/codec/
fuzz result is claimed. Exact-68 activation subsequently passed with
Start-Exitcode 0, a healthy `my-neko/brave:hls-68dbdd4a8dd7` and private enable
snapshot. The operator reports HLS needed Retry, then worked without problems;
the detailed initial error, uninterrupted interval and concurrent WebRTC
result are not separately supplied. Reliable first-start acceptance stays open.
At that checkpoint checkout/live deployment was exact `68dbdd4a`, conventional
HLS enabled. Its read-only diagnostic subsequently passed with
Diagnostic-Exitcode 0: healthy image, no sampled exit/OOM, one not-ready and one
successful bootstrap, one startup generation per track and 17 successful
segment requests; the lease closed and the packager stopped after idle grace.
The saved timing summary from helper `833cff60` then passed with
Timing-Exitcode 0: two requests total 24.001096458 seconds, one at most 1 ms,
the other approximately 24 seconds. This fits a readiness deadline followed
by an immediate warmed Retry, without exact attempt/readiness correlation.
The new server-only candidate raises ConventionalReadyWindow from 24 to 28
seconds within the unchanged 30-second HTTP/client limits; LL-HLS stays six
seconds. This is a bounded candidate, not a confirmed first-start fix; the
p95 24-second acceptance target remains unchanged. Exact candidate
`8f54970f025e3491a123540cc94870508b66f119` preparation subsequently passed with
Prepare-Exitcode 0: all 13 selected native checks passed, including the four
real-codec fixtures in generation 1; base/Brave images and private snapshot/
validation marker were recorded. The supplied excerpt starts during the native
image build; earlier HLS package/server-image checks are covered by final helper
success, without separately visible results. Client build layers were cached
from unchanged sources; no fresh client gate is claimed. Target checkout is
exact-8f54970f; the live image was still exact-68 at preparation. These are
supplied target results, NOT EXECUTED IN CODEX. After the requested activation,
the operator reports connecting, then an apparently frozen first picture;
page reload produced moving video/audio while HLS remained selected. Activation
CLI evidence, the initial player/error/timing state and concurrent WebRTC
behavior are not supplied. The latest read-only target diagnosis passed with
Diagnostic-Exitcode 0: prepared/live image IDs match exact-8f, healthy container,
two successful bootstraps, one startup generation per track, all four workers
running, one active lease and 135 successful segment requests. No sampled
exit/OOM or fixed error was recorded. These cumulative counters do not prove
the first browser rendered moving video/audio or isolate the frozen picture.
A client-only start-order candidate now requests existing autoplay after
source/player attachment, without waiting for canplay. Readiness, Pause,
manual Play/muted fallback and the existing deadlines remain independent.
This is a bounded candidate, not a confirmed live cause/fix. Exact candidate
`7dcc3c5e4ba0279f35727eb4bed2e86bf721ad3b` preparation subsequently passed with
Prepare-Exitcode 0 using `deploy/prepare-hls-client-start.sh`: both candidate
images and private snapshot/validation marker were recorded in
`/opt/docker/nekoNew/neko-hls-results-7dcc3c5e4ba0`, with the live exact-8f
container retained. The supplied tail starts inside the base-image build;
earlier isolated old/new client tests/type/build are covered by final helper
success, without separately visible counts. Backend evidence is inherited
from unchanged exact-8f sources, not freshly rerun. Target checkout is now
exact-7dcc3c5e. After the requested activation, the operator reports moving
first picture after approximately 20 seconds, then a later frozen picture
requiring reload. The subsequent exact-7dcc progress diagnostic passed with
Diagnostic-Exitcode 0, confirming prepared/live image ID
`sha256:c445e6541db2eb4ddf821874480a37be86542f749d5bb83e0cc67ce9640bac25`,
a healthy container, no sampled exit/OOM and valid 12.012-second deltas.
All HLS leases/subscriptions/workers were zero in both samples, with no new
publication/requests/generations; two earlier lease closures and idle stop
were recorded. The operator confirmed HLS was closed/logged out during sampling.
This is expected idle cleanup, not a captured freeze. One cumulative high-source
discontinuity and seven negotiation rejections are not correlated to the freeze.
A WebRTC participant joining/starting video is an uncorrelated hypothesis.
Static review found a shared-capture keyframe request on new listeners, with
existing HLS listeners retained; no direct join-triggered HLS teardown or
video-source overwrite was found. Static inspection corrected a diagnostic-only
label parser defect: `vp8` had been omitted because digits were rejected despite
its allowlist entry. The progress and saved-startup helper now accept digits
before checking the unchanged fixed allowlists. Prior accepted HLS/lease/source
deltas are unaffected. The corrected progress helper `4593a6f9` subsequently
passed on the target with Diagnostic-Exitcode 0, blob
`60cf47ddf8489c2397a5770a4c9523ae09a50b56`, valid 12.011-second deltas and the
same healthy exact-7dcc image. One active lease, both HLS source subscriptions
and all four workers persisted. 601 audio/300 video units were consumed; all
four tracks published parts/segments and HTTP returned success (12 masters,
2 playlists, 2 segments, 1 keepalive). No interval generation/discontinuity/drop/
capture-creation/WebRTC-open increment or sampled exit/OOM/fixed error.
VP8 rows are now visible; idle low/medium capture counters are expected because
HLS fans out the high source. Supplied target results, NOT EXECUTED IN CODEX;
the corrected saved-startup helper has no fresh target result. The operator then
confirmed this sample was taken during a frozen HLS picture with audio stopped,
after initial moving picture/audio and without reload. Later the page reported
"HLS HTTP connection failed; retry manually". This captures a browser failure
despite interval server flow; the later HTTP failure was not sampled on the
server. Inspection shows serial master/keepalive checks use a 1-second full-body
deadline and fail after three HTTP/transport/body-validation errors; that is
a hypothesis, not proof of timeout or the freeze cause. The supplied browser
timing summary now shows 250 retained entries and a last master lasting 1,001 ms
without response metadata; other status-zero entries contain complete body-size
metadata. The last retained request ended about nine minutes before the query,
so subsequent failures are missing. Current video was already cleared by terminal
cleanup. This supports one timeout-like attempt, not the entire failure sequence
or freeze cause. The requested bounded trace has since been supplied in full;
the current repair and focused target preparation are recorded at the top of
this file. Do not repeat the old-image browser trace or unchanged native/fuzz/
HTTP-denial gates. The WebRTC-join
cause and reliable startup/sustained/grouped acceptance remain open.
Reliable first-start
and wider acceptance remain open; do not repeat passed preparation gates.
Preserve old images/evidence and keep dependency maintenance open.
Do not repeat passed client/image/HTTP-denial
gates without a new reason. Prior results below apply only to their commits.
`master` remains pinned at `d9105ef8`.

## Purpose

This repository is a customized fork of [`m1k1o/neko`](https://github.com/m1k1o/neko). It preserves Neko's shared server-side browser/desktop model while carrying fork-specific client work for mobile/touch usability, playback/reconnect recovery, and UI/UX.

Immediate sequence:

1. preserve the current fork — **complete**,
2. reconcile desired local changes from `MyNekoProjekt` — **complete; no source delta was missing and the operator-confirmed deployment compose was imported in sanitized form**,
3. synchronize the completed fork with current upstream without breaking fork behavior — **complete; merge commit `4e99b8d3`**,
4. fast-forward the reviewed integration history to `master` — **complete**,
5. validate the integrated `master` on the target server — **complete; operator-confirmed on 2026-09-09 after correcting the Brave policy mount filename**,
6. make the multi-pipeline/bandwidth-estimator path reproducible and observable without changing the stable single-pipeline default — **complete**,
7. validate and measurement-tune the opt-in adaptive-quality profile on the target server — **complete; operator-accepted on 2026-09-10 for the documented three-viewer scenario at `bfaca84e`**,
8. implement bounded iOS transient recovery without requiring a page reload while preserving Safari's Play fallback — **complete in the repository on `testing`; exact-commit automated checks passed, but the operator deliberately closed the checkpoint without the manual iPhone deep test, so no no-reload device claim is made**,
9. implement server-enforced view-only sharing on `testing` — **complete in the repository and target-server verified through the full boundary matrix plus the compact-link follow-up at `913a981e`**,
10. validate bounded iOS recovery and server-enforced view-only sharing together at one exact `testing` commit — **closed on 2026-09-11 with the explicit iPhone evidence limitation above**,
11. design the backend-neutral encoded-media subscription boundary for practical non-WebRTC prototypes — **complete on `testing`; design only, no alternative backend implemented**,
12. implement the no-new-transport compatibility refactor defined in `docs/MEDIA_SUBSCRIPTION_BOUNDARY.md` — **complete in the repository and bounded target-server checkpoint closed on `testing` at `e5f55bf9`; the repeated full role/recovery matrix and induced three-viewer down/up isolation run are explicitly deferred to final grouped validation**,
13. correct estimator startup observation-window initialization so configured unstable/stalled delays cannot be bypassed by Go zero-time values — **complete and focused target-server validated at `e5f55bf9`; one fresh viewer remained on `high` throughout the recorded startup window**,
14. specify the exact opt-in WebCodecs plus dedicated media-WebSocket prototype contract without adding a transport yet — **complete on `testing`; design only, no endpoint/backend/client transport implemented**,
15. implement Phase 1 of the default-off WebCodecs plus dedicated media-WebSocket receive prototype: protocol, fixtures, PTS-validity propagation, authenticated negotiation and one-time tickets — **complete in the repository on `testing`; target-server tests/build intentionally deferred to grouped prototype validation**,
16. implement Phase 2: the credential-free server delivery backend, pre-upgrade security boundary, dedicated media route, bounded queues and lifecycle cleanup — **complete in the repository on `testing`; statically reviewed, target-server tests/build intentionally deferred to grouped prototype validation**,
17. implement the isolated opt-in client decode/render path — **complete in the repository on `testing`; statically reviewed, target-server client/server tests, builds and browser/media acceptance intentionally deferred to grouped prototype validation**,
18. add separate deployment/observability assets and validate the completed prototype only in later blocks, while retaining WebRTC as the default — **repository assets complete and a bounded target-server checkpoint closed at `6b6cd328`; exact automated/build/security checks, the earlier role/fullscreen/private-mode matrix and the corrected foreground-iPhone interval passed, while numeric latency, induced slow-client/adaptive isolation, resource comparison and remaining live hostile-input cases were explicitly deferred without a full prototype-acceptance claim**,
19. productize the still-explicit client choice without adding automatic fallback — **complete on `testing`; the per-client WebRTC/WebCodecs preference, URL override precedence and compact healthy status are implemented and focused target-server validated at `12cfe43b`**,
20. validate the per-client media-backend productization on one exact `testing` commit while preserving the bounded Phase 4 limitations — **closed on 2026-09-19 at `12cfe43b`; 26 client tests, type/build, image/deployment/security, preference/override/status and disabled-backend terminal gates passed; live view-only fragment preservation was not repeated, while its focused automated test passed**,
21. specify the exact default-off HLS/LL-HLS passive/view-only prototype contract without implementing a transport yet — **complete on `testing`; design only, no HLS endpoint, packager or player implemented**,
22. correct the confirmed steady-state WebRTC estimator downgrade defect so neutral/application-limited estimates do not need upgrade-style spare capacity and a loss-free GCC target collapse cannot change tiers by itself — **receiver-evidence revision target-validated at `2efcc6b1`: exact tests/build/deployment, 20-minute healthy hold, real constrained downgrade and peer isolation passed**,
23. break lower-tier application-limited recovery deadlock with a clean peer-local one-tier probe and exponential failed-probe backoff — **complete on `testing`; focused exact-commit tests/build/deployment and the two-viewer constrained recovery gate passed at `ddf15cee`**,
24. implement HLS/LL-HLS Phase 1: default-off configuration, authenticated bootstrap/playback-lease foundations, deterministic playlist/object models and golden fixtures without starting a packager or adding a client player — **complete on `testing`; focused automated target checks passed at `e85d8568`, default-off image deployment healthy at `741025c3`; HLS runtime acceptance pending**,
25. implement HLS/LL-HLS Phase 2: shared bounded packager, H.264/AAC fMP4 generation, authenticated HTTP delivery, central per-session attachment and fixed observability without adding a client player — **complete on `testing`; focused automated target checks passed at `e85d8568`, default-off image deployment healthy at `741025c3`; HLS-enabled runtime validation pending**,
26. implement HLS/LL-HLS Phase 3: isolated passive client, pinned player support, explicit advertised-only HLS/LL-HLS selection and bounded lifecycle cleanup without automatic fallback — **complete on `testing`; focused automated target checks passed at `e85d8568`, default-off image deployment healthy at `741025c3`; HLS-enabled runtime/device validation pending**,
27. implement HLS/LL-HLS Phase 4 separate opt-in deployment/observability assets and grouped target-server acceptance — **exact repair-commit automated/image preparation passed at `93f1fa63`; helper `2484a022` then passed the 11-host preserving Caddy merge/reload, synthetic runtime-error redaction, healthy conventional-HLS image deployment and 17 public plus 2 cleartext-denial probes; first enabled playback attempt failed per operator report (connection failure/no picture); read-only target diagnosis passed with healthy image but no demonstrated HLS readiness; normal login also timed out (tentative /ws 101); cold-start/generation/timestamp/log-bound repairs passed the target automated/image gate at 80020d99, including real-codec segment mapping and all-four-rendition conventional readiness in one generation (test duration 18.11 seconds); pre-HLS rollback passed with healthy service and operator-confirmed normal login/picture/audio; default-off repair-image deployment passed with Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes; operator-confirmed normal login/picture/audio/control; same-image HLS activation passed with Enable-Exitcode 0, healthy service and 19/19 HTTP denial probes, but HLS failed and the operator reported all streams stopped afterward; read-only diagnosis passed with one ready packager/lease, 23 successful segment requests and two timeline-gap rejections, no sampled Neko exit/OOM; same-image default-off restoration passed with Recovery-Exitcode 0 and 2/2 disabled-route probes; HLS-only fixed-GOP correction passed the isolated target GOP A/B gate at 97ba4ad9: old code reproduced two timeline gaps and repaired code passed all three codec tests, including sustained scene cuts in generation 1 (30.19 seconds); exact 97ba4ad9 automated/image preparation passed with Repair-Check-Exitcode 0 (47 client tests, 13 Go packages, both fuzz jobs, all three codec tests and base/Brave builds); Default-off deployment of my-neko/brave:hls-97ba4ad9ab3e passed with Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes. The operator reported the requested normal browser check works. The subsequent HLS attempt at the prepared 97ba4ad9 checkpoint failed with "HLS bootstrap failed; retry manually", and the operator confirmed WebRTC also stopped working. The latest enablement CLI/HTTP results have not been supplied. Read-only diagnosis passed with 84 log lines, one not-ready bootstrap, one negotiation rejection and no sampled exit/OOM or generation/lease-open markers. Same-image default-off restoration passed with Recovery-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login/picture/audio work again. The saved-startup summary passed with Saved-Check-Exitcode 0: audio/high/medium subscriptions persisted, low capture reached only its create marker, and the saved matching environment had HLS enabled. Paired startup diagnosis at helper 88f2b25d completed with exit 0: both constructors passed two starts and failed one, with medium losing its initial pre-anchor IDR and reaching only two parents. This narrows the smooth readiness defect independently of registry isolation. The 414639d2 isolated anchor A/B gate passed with Anchor-Check-Exitcode 0: old code reproduced the fixed initial-IDR failure; seven corrected checks each passed three times, with every real-codec fixture staying in generation 1. Full exact-414639d2 preparation FAILED with Repair-Prepare-Exitcode 1: 47 client tests, type/build, 13 Go packages, both fuzz jobs, registry/mapping and anchor lifecycle checks passed, but smooth readiness restarted with worker_failure and failed its generation-1 assertion at 20.03 seconds; scene cuts passed at 30.19 seconds in generation 1. Base/Brave image steps were not reached. Worker diagnosis at helper 53034495 then completed with exit 0: all six checks passed in three fresh processes, each smooth fixture ready in 18.11 seconds in generation 1 with no rejected pushes; the earlier worker failure was not reproduced. The controlled audio-anchor A/B at repair 71a14d21 passed with Audio-Anchor-Exitcode 0: the old AAC hold reproduced both the blocked-drainage unit failure and an audio/anchor/queue_full restart under 256 ms high-input delay; all eight corrected startup checks passed in three fresh processes and scene cuts passed once (25 top-level passes). All seven real-codec fixtures stayed in generation 1; normal readiness was 18.10/18.11/18.11 seconds and delayed-high readiness 18.09 seconds each. This verifies the controlled AAC-overflow mechanism, not the cause of the earlier unobserved worker failure or live all-stream outage. Full exact-71a14d21 preparation then passed with Repair-Prepare-Exitcode 0: the rebuilt GStreamer 1.26.2 image passed all nine selected startup checks, including normal readiness at 18.11 seconds, delayed-high readiness at 18.09 seconds and scene cuts at 30.19 seconds, all in generation 1; base/Brave images were built and the service remained unchanged. The supplied tail starts inside the codec-image build; earlier client/Go/fuzz steps are covered by the script's reported final success, not separately shown in this excerpt. Default-off deployment of my-neko/brave:hls-71a14d2174da then passed with Baseline-Exitcode 0, a healthy service and 2/2 disabled-route probes. The operator confirmed the requested normal browser check works without HLS at 71a14d21. Same-image conventional-HLS activation at 71a14d21 then passed with Enable-Exitcode 0, healthy service, a private enable snapshot and 17/17 public plus 2/2 cleartext-denial probes. The subsequent operator-reported HLS attempt at 71a14d21 went from connecting to failed with "HLS bootstrap failed; retry manually"; the operator reported only WebRTC streaming works. No successful HLS picture/audio or room-event interval is demonstrated. Read-only diagnosis passed with 950 captured lines, one not-ready bootstrap, one started/idle-stopped packager generation, medium/low keyframe-admission drops of 759/564 and cumulative part/segment publication only for audio/high. No sampled Neko exit/OOM was found. Same-image default-off restoration passed with Recovery-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login/picture/audio work again. The isolated source-clock-phase diagnosis at helper 409482b4 passed with Clock-Skew-Exitcode 0: aligned control ready in 18.11 seconds; all three skew runs reproduced not-ready at 24.02 seconds in generation 1 with audio/high ready, medium/low blocked, flowing IDRs and no rejected native pushes. This proves the controlled phase-admission defect, not the exact unmeasured live phases. A common high-source fan-out repair is implemented and statically reviewed: one shared video provider subscription feeds the existing three scaled encoders, plus one audio subscription; provider PTS/DTS are preserved. The isolated common-source repair A/B passed at a7ffb8b1 with Shared-Clock-Exitcode 0: the old defect reproduced once; 46 positive top-level checks passed across three cold processes including one scene-cut check, with all ten real-codec fixtures in generation 1. Full exact-a7ffb8b1 preparation then passed with Repair-Prepare-Exitcode 0: all thirteen selected native/startup checks passed in the rebuilt GStreamer 1.26.2 image, including the four real-codec fixtures in generation 1, and base/Brave images were built without changing the running service. The supplied excerpt starts inside the codec-image build; earlier client/Go/fuzz stages are covered by final script success without separately shown fresh counts. Default-off deployment of my-neko/brave:hls-a7ffb8b13448 then passed with Baseline-Exitcode 0, healthy service, a private baseline snapshot and 2/2 disabled-route probes; the operator reported the requested normal browser check works. Same-image a7ffb8b1 HLS activation passed with Enable-Exitcode 0, healthy service and 19/19 HTTP denial probes. The operator then reported first HLS picture and the compact streaming label, followed after roughly 30 seconds by "HLS playback did not become ready; retry manually"; Retry HLS restored playback and WebRTC continued working. Static inspection found the initial client readiness deadline was never disarmed on canplay/playing. Client-only repair 73d5ff6d cancels it on those current-player events, arms it before attachment and retains the independent startup/stall bounds; Read-only target diagnosis then passed with a healthy a7ff image, two successful HLS bootstraps and 486 successful segment requests, with no sampled process exit/OOM; the packager stopped after idle grace. The isolated target client gate passed with Client-Check-Exitcode 0: the old timer defect reproduced, all 52 repaired client tests plus type/build passed, and checkout/live service remained at a7ff. These are supplied target results, NOT EXECUTED IN CODEX. Scoped exact-73d5ff6d image preparation then passed with Client-Image-Exitcode 0: fresh client bundle index-CrHQRMnq.js, cached unchanged server/runtime layers, both base/Brave images and private snapshot/marker recorded while enabled a7ff stayed running. Exact-73d5ff6d default-off deployment then passed with Baseline-Exitcode 0, healthy my-neko/brave:hls-73d5ff6d2911, a private baseline snapshot and 2/2 disabled-route probes; the operator confirmed the requested normal browser check works without a media override. Same-image conventional-HLS activation then passed with Enable-Exitcode 0, healthy service, a private enable snapshot and 19/19 denial probes. The operator reports PC/Helium HLS playback after an initial Retry and one frozen-picture/page-reload incident; WebRTC kept working and later HLS worked normally. A HLS failed message was confirmed without its detailed error or exact timing, so startup reliability and an uninterrupted room-event interval remain unverified. Checkout/live image is now 73d5ff6d with conventional HLS enabled. Read-only exact-73 diagnosis then passed with Diagnostic-Exitcode 0: healthy service, no sampled exit/OOM or fixed error markers, one active HLS lease/all four workers running, six successful bootstraps and 394 successful segment requests. One not-ready bootstrap supports readiness as a possible initial-Retry explanation without attempt correlation; two startup-labelled generations and one idle stop do not establish a crash loop. The operator cannot confirm the exact uninterrupted interval and mentions possible random reconnects/room actions without correlation. NEXT consolidate startup/frozen-picture/room-event investigation into one bounded later validation step; no further ad-hoc operator check requested at this checkpoint; startup/recovery and grouped acceptance pending; authorization/lifecycle and grouped acceptance pending**,
28. record the final integrated static stability review from `docs/STABILITY_REVIEW.md` before final grouped validation and any promotion — **source/configuration and detailed audit classification recorded; notification, paused HLS renewal and additional chat/tooltip/Axios containment repairs passed automation at `93f1fa63`; browser/deployment checks, package remediation and final security/live acceptance remain open**,
29. keep accumulating reviewed implementation blocks on `testing`; promote to `master` only after the operator explicitly authorizes the final grouped promotion — **pending**.

## Authoritative knowledge

Read before substantial work:

- `README.md` — compact entry and current status.
- `docs/PROJECT.md` — goals, requirements, invariants and target state.
- `docs/ARCHITECTURE.md` — verified current architecture.
- `docs/WORKPLAN.md` — current `NEXT`, sync procedure, server-side verification and open items.
- `docs/STABILITY_REVIEW.md` — reported TV/event failures, evidence limits, mandatory final stability review and device/startup comparisons.
- `docs/LOCAL_DELTA_AUDIT.md` — completed, sanitized classification of the supplied `MyNekoProjekt` snapshot.
- `docs/UPSTREAM_SYNC_AUDIT.md` — completed semantic review and merge record for the 2026-09-09 upstream synchronization.
- `docs/ADAPTIVE_QUALITY.md` — opt-in profile, diagnostics, resource costs, exact target-server acceptance procedure and rollback.
- `docs/IOS_RECOVERY.md` — implemented bounded reconnect states, exact target-server iPhone procedure, acceptance criteria and rollback.
- `docs/VIEW_ONLY_SHARING.md` — implemented passive-session boundary, token lifetime/revocation, denial behavior, exact three-role target-server matrix and rollback.
- `docs/MEDIA_SUBSCRIPTION_BOUNDARY.md` — decided backend-neutral encoded-media/source-subscription and participant-delivery contract, migration order, security boundary and prototype gates.
- `docs/HLS_LL_HLS.md` — exact version-1 passive HLS/LL-HLS codec, packaging, authorization, HTTP security, retention, rollout and acceptance contract.
- `docs/DEPENDENCY_AUDIT_2026-10-04.md` — supplied exact-e55 audit classification, client containment repairs, target revalidation and open dependency-maintenance decisions.
- `docs/HLS_LL_HLS_VALIDATION.md` — exact Phase 4 target-server preparation, opt-in enablement, security/role/device/resource gates and rollback.
- `docs/HLS_CLIENT_READINESS_REPAIR_2026-10-05.md` — first live HLS picture, premature startup deadline, passed client-only target gate and scoped image preparation.
- `docs/HLS_LL_HLS_CADDY.md` — host-service proxy trust, streaming and access/runtime-log review before valid HLS media tests.
- `docs/HLS_LL_HLS_OBSERVABILITY.md` — private evidence collector, fixed HLS metrics and resource interpretation.
- `docs/STABILITY_REVIEW_2026-10-04.md` — integrated source/configuration findings, bounded repairs and outstanding advisory/device evidence.
- `docs/WEBCODECS_MEDIA_WEBSOCKET.md` — exact version-1 authentication, wire-format, codec, queueing, A/V synchronization, recovery, security, rollout and acceptance contract for the default-off receive prototype.
- `docs/WEBCODECS_MEDIA_WEBSOCKET_OBSERVABILITY.md` — credential-safe fixed metrics, PromQL and evidence collection for the opt-in prototype.
- `docs/WEBCODECS_MEDIA_WEBSOCKET_VALIDATION.md` — exact interactive target-server Phase 4 procedure and rollback.
- `webpage/docs/` — inherited Neko documentation. Current repository code/config wins on conflicts.

## Truth rules

1. Code, configuration, Git history and other real artifacts determine **IMPLEMENTED**.
2. The supplied `MyNekoProjekt` snapshot was audited on 2026-09-09. Any later local delta must be reviewed explicitly before import.
3. `docs/PROJECT.md` determines **TARGET**.
4. Upstream issues/PRs are evidence or candidates, not automatically requirements.
5. Never silently drop fork-specific behavior during upstream conflict resolution.
6. Never commit deployment credentials, browser profiles, downloads, cookies, lock files or other runtime data.

## Codex execution policy

**Codex is an edit and static-review environment only. It is not the target runtime environment.**

Do not execute project code or runtime verification in Codex.

Do **not**:

- install project dependencies (`npm install`, `npm ci`, `go get`, etc.);
- start the client, server, browser, Vite, Docker containers or images;
- run tests, linters, type-checkers, builds or package scripts;
- run repository build/start scripts;
- perform WebRTC/media/network/device runtime tests;
- treat Codex-environment execution as evidence of target-server compatibility.

Static repository work is allowed and expected:

- read and compare files;
- inspect Git history, status and diffs;
- inspect manifests, configuration and source;
- reason about syntax/type/build/runtime implications;
- identify likely regressions by inspection;
- prepare exact verification commands/checklists for the real server.

Runtime/build/test status must be reported as **NOT EXECUTED IN CODEX** unless results are supplied from the target server.

## Repository map

- `client/` — Vue 2.7 + TypeScript/Vite client; most fork-specific work currently lives here.
- `server/` — Go server and plugins.
- `apps/` — browser/application image definitions.
- `runtime/` — runtime image/container support.
- `webpage/` — inherited Neko documentation site.
- `docs/` — fork-specific durable project knowledge.
- `deploy/` — tracked, non-secret opt-in deployment configuration overlays.

## Server-side verification reference

These commands are for the **real target server/environment only**. Codex must not run them.

Client:

```bash
cd client
npm ci
npm run lint
npm run build
```

Server:

```bash
cd server
./build
```

Container build:

```bash
docker build ./server
```

Use only the checks relevant to the changed areas, with broader verification after major integrations.

## Technical invariants

- One shared remote browser/desktop/session is seen by multiple participants.
- At most one participant controls the shared desktop at a time.
- Admin lock and grant/revoke behavior must remain intact.
- Fork mobile/touch/trackpad, autoplay, playback-recovery, fullscreen and reconnect behavior is regression-sensitive.
- A weak viewer must not degrade healthy viewers in the target architecture.
- Adaptive downgrade requires sustained material insufficiency against the peer-local complete-delivery requirement plus fresh receiver loss/NACK evidence; GCC target/trend alone is advisory, ordinary upgrade reserve is a separate next-tier gate, and neither decision may couple viewers. A clean peer with an outstanding automatic-downgrade recovery step may probe only one higher tier after its separate interval/current-tier reserve; failed probes use evidence-gated fallback and capped exponential peer-local backoff.
- View-only sharing is enforced server-side; hiding controls in the UI is not authorization.
- WebRTC is the currently implemented primary media path.
- Future interactive and passive/view-only clients may use different media backends in the same logical room; media transport must not determine authorization.
- WebCodecs/WebSocket Phases 1–3 are IMPLEMENTED: protocol/ticket/negotiation, the credential-free server route/delivery backend and the receive-only client parser/worker/WebCodecs/AudioWorklet/canvas path. Manual client productization persists a per-browser WebRTC/WebCodecs default, resolves absent/invalid storage to WebRTC and retains exact `?media=webcodecs-ws` as a stateless highest-priority diagnostic override. Healthy WebCodecs streaming uses only the compact `WebCodecs` label; negotiation, recovery and terminal actions remain prominent. The focused exact-commit target checkpoint for that productization passed at `12cfe43b`, except that live view-only fragment preservation was not repeated and remains backed only by focused automated coverage. Phase 4 repository assets add only a separate explicit deployment overlay, credential-safe observability and an interactive acceptance runbook. Its wider bounded target-server checkpoint passed the recorded exact automated/build/security, functional and corrected foreground-iPhone gates at `6b6cd328`, but the deliberately omitted numeric latency, induced isolation, resource and remaining hostile-input matrix prevents a full prototype-acceptance claim. The server remains default-off when that overlay is omitted, and new/unset clients remain on WebRTC. HLS/LL-HLS Phases 1–3 implement default-off negotiation, tickets/leases, strict HTTP security, deterministic models, a shared bounded H.264/AAC fMP4 packager and authenticated server delivery; Phase 3 adds an isolated explicitly selected native/MSE client with pinned local hls.js, advertised-only passive/admin choices and bounded pause/revoke cleanup. Phase 4 provides a separate opt-in HLS overlay, private evidence/HTTP-check/image/rollback helpers and the host-Caddy runbook; exact application 93f1fa63 passed automated/image preparation and helper 2484a022 passed the preserving Caddy merge/reload, synthetic runtime-error redaction, healthy HLS activation and 19 denial probes. The first playback attempt and subsequent normal login failed. The supplied read-only diagnosis found a healthy container and HLS startup/source-restart evidence without demonstrated readiness. HLS startup/generation/timestamp/log-bound repairs and real-codec integration assets are implemented; their target automated/image gate passed at 80020d99, including both real-codec tests; live repair-image acceptance remains pending. The recorded pre-HLS runtime was restored healthy with Restore-Exitcode 0 and operator-confirmed normal login/picture/audio. The prepared 80020d99 image passed default-off deployment (Baseline-Exitcode 0, healthy, 2/2 disabled-route probes) and operator-confirmed normal login/picture/audio/control. Same-image HLS activation passed with Enable-Exitcode 0, healthy service and 19/19 HTTP denial probes, but HLS failed and the operator reported all streams stopped afterward. Read-only diagnosis passed with one ready packager/lease, 23 successful segment requests and two timeline-gap rejections; no sampled Neko exit/OOM was found. Same-image default-off restoration passed with Recovery-Exitcode 0 and 2/2 disabled-route probes; its fresh browser confirmation is pending. HLS-only scene-cut suppression passed the isolated target GOP A/B gate at 97ba4ad9: old code reproduced two timeline gaps, all three repaired codec tests passed and the scene-cut fixture stayed in generation 1 for 30.19 seconds. Exact 97ba4ad9 automated/image preparation passed with Repair-Check-Exitcode 0 (47 client tests, 13 Go packages, both fuzz jobs, all three codec tests and base/Brave builds); Default-off deployment of my-neko/brave:hls-97ba4ad9ab3e passed with Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes. The operator reported the requested normal browser check works. The subsequent HLS attempt at the prepared 97ba4ad9 checkpoint failed with "HLS bootstrap failed; retry manually", and the operator confirmed WebRTC also stopped working. The latest enablement CLI/HTTP results have not been supplied. Read-only diagnosis passed with 84 log lines, one not-ready bootstrap, one negotiation rejection and no sampled exit/OOM or generation/lease-open markers. Same-image default-off restoration passed with Recovery-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login/picture/audio work again. The saved-startup summary passed with Saved-Check-Exitcode 0: audio/high/medium subscriptions persisted, low capture reached only its create marker, and the saved matching environment had HLS enabled. Paired startup diagnosis at helper 88f2b25d completed with exit 0: both constructors passed two starts and failed one, with medium losing its initial pre-anchor IDR and reaching only two parents. This narrows the smooth readiness defect independently of registry isolation. The 414639d2 isolated anchor A/B gate passed with Anchor-Check-Exitcode 0: old code reproduced the fixed initial-IDR failure; seven corrected checks each passed three times, with every real-codec fixture staying in generation 1. Full exact-414639d2 preparation FAILED with Repair-Prepare-Exitcode 1: 47 client tests, type/build, 13 Go packages, both fuzz jobs, registry/mapping and anchor lifecycle checks passed, but smooth readiness restarted with worker_failure and failed its generation-1 assertion at 20.03 seconds; scene cuts passed at 30.19 seconds in generation 1. Base/Brave image steps were not reached. Worker diagnosis at helper 53034495 then completed with exit 0: all six checks passed in three fresh processes, each smooth fixture ready in 18.11 seconds in generation 1 with no rejected pushes; the earlier worker failure was not reproduced. The controlled audio-anchor A/B at repair 71a14d21 passed with Audio-Anchor-Exitcode 0: the old AAC hold reproduced both the blocked-drainage unit failure and an audio/anchor/queue_full restart under 256 ms high-input delay; all eight corrected startup checks passed in three fresh processes and scene cuts passed once (25 top-level passes). All seven real-codec fixtures stayed in generation 1; normal readiness was 18.10/18.11/18.11 seconds and delayed-high readiness 18.09 seconds each. This verifies the controlled AAC-overflow mechanism, not the cause of the earlier unobserved worker failure or live all-stream outage. Full exact-71a14d21 preparation then passed with Repair-Prepare-Exitcode 0: the rebuilt GStreamer 1.26.2 image passed all nine selected startup checks, including normal readiness at 18.11 seconds, delayed-high readiness at 18.09 seconds and scene cuts at 30.19 seconds, all in generation 1; base/Brave images were built and the service remained unchanged. The supplied tail starts inside the codec-image build; earlier client/Go/fuzz steps are covered by the script's reported final success, not separately shown in this excerpt. Default-off deployment of my-neko/brave:hls-71a14d2174da then passed with Baseline-Exitcode 0, a healthy service and 2/2 disabled-route probes. The operator confirmed the requested normal browser check works without HLS at 71a14d21. Same-image conventional-HLS activation at 71a14d21 then passed with Enable-Exitcode 0, healthy service, a private enable snapshot and 17/17 public plus 2/2 cleartext-denial probes. The subsequent operator-reported HLS attempt at 71a14d21 went from connecting to failed with "HLS bootstrap failed; retry manually"; the operator reported only WebRTC streaming works. No successful HLS picture/audio or room-event interval is demonstrated. Read-only diagnosis passed with 950 captured lines, one not-ready bootstrap, one started/idle-stopped packager generation, medium/low keyframe-admission drops of 759/564 and cumulative part/segment publication only for audio/high. No sampled Neko exit/OOM was found. Same-image default-off restoration passed with Recovery-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login/picture/audio work again. The isolated source-clock-phase diagnosis at helper 409482b4 passed with Clock-Skew-Exitcode 0: aligned control ready in 18.11 seconds; all three skew runs reproduced not-ready at 24.02 seconds in generation 1 with audio/high ready, medium/low blocked, flowing IDRs and no rejected native pushes. This proves the controlled phase-admission defect, not the exact unmeasured live phases. A common high-source fan-out repair is implemented and statically reviewed: one shared video provider subscription feeds the existing three scaled encoders, plus one audio subscription; provider PTS/DTS are preserved. The isolated common-source repair A/B passed at a7ffb8b1 with Shared-Clock-Exitcode 0: the old defect reproduced once; 46 positive top-level checks passed across three cold processes including one scene-cut check, with all ten real-codec fixtures in generation 1. Full exact-a7ffb8b1 preparation then passed with Repair-Prepare-Exitcode 0: all thirteen selected native/startup checks passed in the rebuilt GStreamer 1.26.2 image, including the four real-codec fixtures in generation 1, and base/Brave images were built without changing the running service. The supplied excerpt starts inside the codec-image build; earlier client/Go/fuzz stages are covered by final script success without separately shown fresh counts. Default-off deployment of my-neko/brave:hls-a7ffb8b13448 then passed with Baseline-Exitcode 0, healthy service, a private baseline snapshot and 2/2 disabled-route probes; the operator reported the requested normal browser check works. Same-image a7ffb8b1 HLS activation passed with Enable-Exitcode 0, healthy service and 19/19 HTTP denial probes. The operator then reported first HLS picture and the compact streaming label, followed after roughly 30 seconds by "HLS playback did not become ready; retry manually"; Retry HLS restored playback and WebRTC continued working. Static inspection found the initial client readiness deadline was never disarmed on canplay/playing. Client-only repair 73d5ff6d cancels it on those current-player events, arms it before attachment and retains the independent startup/stall bounds; Read-only target diagnosis then passed with a healthy a7ff image, two successful HLS bootstraps and 486 successful segment requests, with no sampled process exit/OOM; the packager stopped after idle grace. The isolated target client gate passed with Client-Check-Exitcode 0: the old timer defect reproduced, all 52 repaired client tests plus type/build passed, and checkout/live service remained at a7ff. These are supplied target results, NOT EXECUTED IN CODEX. Scoped exact-73d5ff6d image preparation then passed with Client-Image-Exitcode 0: fresh client bundle index-CrHQRMnq.js, cached unchanged server/runtime layers, both base/Brave images and private snapshot/marker recorded while enabled a7ff stayed running. Exact-73d5ff6d default-off deployment then passed with Baseline-Exitcode 0, healthy my-neko/brave:hls-73d5ff6d2911, a private baseline snapshot and 2/2 disabled-route probes; the operator confirmed the requested normal browser check works without a media override. Same-image conventional-HLS activation then passed with Enable-Exitcode 0, healthy service, a private enable snapshot and 19/19 denial probes. The operator reports PC/Helium HLS playback after an initial Retry and one frozen-picture/page-reload incident; WebRTC kept working and later HLS worked normally. A HLS failed message was confirmed without its detailed error or exact timing, so startup reliability and an uninterrupted room-event interval remain unverified. Checkout/live image is now 73d5ff6d with conventional HLS enabled. Read-only exact-73 diagnosis then passed with Diagnostic-Exitcode 0: healthy service, no sampled exit/OOM or fixed error markers, one active HLS lease/all four workers running, six successful bootstraps and 394 successful segment requests. One not-ready bootstrap supports readiness as a possible initial-Retry explanation without attempt correlation; two startup-labelled generations and one idle stop do not establish a crash loop. The operator cannot confirm the exact uninterrupted interval and mentions possible random reconnects/room actions without correlation. NEXT consolidate startup/frozen-picture/room-event investigation into one bounded later validation step; no further ad-hoc operator check requested at this checkpoint; startup/recovery and grouped acceptance pending; the reported all-stream failure's cause remains unconfirmed. Valid picture/audio, authorization/lifecycle and grouped acceptance remain pending. See docs/HLS_STARTUP_REPAIR_2026-10-04.md. The required final stability review and device evidence limits are tracked in `docs/STABILITY_REVIEW.md`. Do not assume WebSocket is inherently better for poor networks, and do not promote MJPEG beyond an optional ultra-legacy fallback without device evidence.

## Definition of Done for Codex work

A Codex task is complete when:

- the requested repository change is implemented and reviewed statically;
- relevant diffs and surrounding code are inspected for regressions;
- no secrets/runtime data are committed;
- required target-server verification is explicitly listed as pending;
- durable docs are updated when project truth/status changed.

Do not claim runtime/build/test success without results from the real target server.
