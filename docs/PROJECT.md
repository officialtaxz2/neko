# Project Definition

Last consolidated: 2026-09-14.

State labels:

- **IMPLEMENTED** — verified in repository code/config or inherited behavior currently present.
- **TARGET** — desired and decided, but not fully implemented.
- **LATER/OPTIONAL** — not required for the current implementation sequence.
- **OPEN** — unresolved.

## Purpose

This fork keeps Neko's shared server-side browser/desktop model while improving reliability across heterogeneous viewers.

The product remains one shared session, not independent per-user browser sessions.

## Core invariants

### IMPLEMENTED / preserve

- One shared browser/desktop/session for multiple participants.
- Admin and regular-user access.
- At most one controller at a time.
- Admin lock/grant/revoke semantics.
- A backend-neutral, server-enforced passive/view-only session identity.
- Self-hosted, Docker-oriented Neko deployment.
- WebRTC as the primary implemented media path.

### TARGET

- Weak viewers do not degrade healthy viewers.
- Quality adapts independently per viewer.
- Mobile and Smart-TV/browser robustness.
- Target-server validation and future alternate-media support for view-only guest/share access.
- Bounded, observable reconnect/recovery.
- Media/control/auth remain separable enough to support alternative media paths.

## Verified fork-specific implementation

### IMPLEMENTED

- Vue 2.7 + TypeScript client built with Vite.
- Major client/UI redesign.
- Touch-device detection.
- Trackpad mode and optional trackpad cursor.
- Mobile keyboard and keyboard-helper integration.
- Letterbox/pillarbox-aware touch coordinate handling.
- Mobile autoplay muted fallback.
- Media-element stall/waiting/timeupdate monitoring.
- Track mute/unmute/ended monitoring.
- Bounded `srcObject`-based stream recovery.
- Explicit avoidance of `video.load()` for WebRTC recovery.
- ICE failure/disconnect timeout handling.
- Bounded application-level reconnect for a previously established real session after the existing peer/socket cannot recover: four serialized attempts after 1, 2, 5 and 10 seconds.
- Automatic reconnect suppression for initial-login failure, explicit logout, demo mode and server-directed disconnects, preserving authentication/session intent.
- Identity guards and teardown for stale WebSocket, peer, data-channel, ICE-candidate and media-stream callbacks before a replacement connection becomes authoritative.
- Public STUN fallback injection if no STUN URL is configured.
- Demo mode.
- Cross-browser fullscreen handling.
- Optional 96-bit view-only share credential encoded as exactly 16 Base64URL characters and transported by a compact `#/<token>` fragment and WebSocket subprotocol rather than an HTTP request path/query string.
- Fixed view-only profile/session normalization plus HTTP, current/legacy WebSocket, modern/legacy data-channel, inbound-media, host-assignment and plugin enforcement.
- Non-persisted passive sessions with documented token rotation/removal and service recreation as the hard revocation boundary.

## Integrated upstream implementation

### IMPLEMENTED in `master` (upstream merge `4e99b8d3`)

- Per-peer WebRTC sample queues are bounded and non-blocking; a full peer queue drops that peer's sample instead of blocking capture dispatch.
- Capture listener dispatch no longer holds the shared listener lock while writing samples.
- Existing multi-pipeline selection and per-peer bandwidth-estimator support is now requested automatically by the legacy protocol path when the operator configures multiple pipelines and enables the estimator.
- H.265 is available as a capture/WebRTC codec, including software, VA-API and NVENC pipeline construction.
- File transfer supports separately configured non-admin download, upload and delete permissions plus bulk deletion.
- The optional `openinapp` plugin can open HTTP(S) chat links in the shared application for an authorized host.
- Capture-pointer visibility, clipboard resync on window focus, `?scroll=` sensitivity and several browser/runtime compatibility fixes are present.

These repository implementation claims were first established by static inspection. On 2026-09-09, the operator confirmed that all checks applicable to the target deployment in the documented build and regression matrix passed after the Brave policy-mount correction. No universal support is claimed for devices or architectures not identified in that operator report.

## Adaptive-quality follow-up

### IMPLEMENTED in the repository / target-server accepted for the documented scenario

- The Brave deployment has a separate opt-in adaptive-quality overlay with ordered `high`, `medium` and `low` VP8 pipelines; the stable base Compose file remains single-pipeline.
- Encoded stream rates are measured and compared with the Pion target in bit/s.
- Peer-local audio/video queue drops and per-pipeline bitrates are exported as Prometheus metrics.
- Downgrade deficit and upgrade reserve are separate decisions. Neutral estimation and the mere absence of upgrade-style headroom do not by themselves make the current tier congested.
- Current-tier fit uses the greater of measured and nominal video rate plus the active measured audio rate and a configurable transport reserve. Upgrade gating applies its separate reserve to the next tier's nominal complete-delivery reference, with measured video retained as the fallback where nominal metadata is absent.
- The GCC target/trend cannot authorize a downgrade without fresh peer-local RTCP interval loss or NACK evidence; stale, missing or clean feedback keeps the current tier.
- A peer with an outstanding automatic-downgrade recovery step may test exactly one higher tier when its clean application-limited target cannot reach the next-tier nominal gate, but only after a separate bounded interval and current-tier reserve; failed probes return through the normal evidence-gated downgrade and exponential capped peer-local backoff.
- Stable, unstable and stalled estimator observation windows start together instead of leaving the latter two at Go zero time; switch-backoff timestamps remain unset until a real switch.
- Activation, diagnostics, resource costs, rollback and the target-server acceptance sequence are documented in [`ADAPTIVE_QUALITY.md`](ADAPTIVE_QUALITY.md).

On 2026-09-10, the operator accepted commit `bfaca84e` on the target server for one desktop, one iPad and one iPhone. The isolated 0.7 Mbit/s rerun held the iPhone on `low` after downgrade, preserved both healthy viewers on `high` with zero peer-local drops, and recovered through `medium` to `high` within 45 seconds after impairment removal. Refresh/rejoin and a transient cellular interruption also preserved both healthy viewers. This is bounded evidence for that deployment and device set, not a universal profile guarantee. The later zero-time startup correction passed its focused target-server trace at `e5f55bf9`; the operator closed that follow-up while explicitly deferring a repeat of the full role/recovery and induced three-viewer adaptive matrix to final grouped validation.

On 2026-09-19, operator-supplied evidence confirmed a separate false-downgrade case: one healthy direct-UDP WebRTC session had zero reported packet loss, NACKs and local video drops but repeatedly switched tiers. The first repository candidate at `2d037f39` passed focused tests/build and a short live gate, then failed an extended unshaped soak on 2026-09-22 with the exact sequence `high -> medium -> low -> medium -> high` while loss, NACKs and video drops remained zero. The receiver-evidence revision at `2efcc6b1` then passed target-server tests/build/deployment, held H on `high` for 20 minutes and isolated real C downgrades, but C remained application-limited on `low` for the complete 90-second recovery phase before eventually returning to `high`. Follow-up `ddf15cee` retained that validated downgrade gate and added only the bounded peer-local one-tier recovery probe described above. Its exact-commit target suite/build/deployment and shortened two-viewer gate passed: H remained `high` with zero video-drop delta, shaped C reached `low`, and clean C recovered in order to `medium` and `high` within 60 seconds. The original startup, stable, unstable, stalled and switch-backoff durations remain unchanged.

## Highest-priority target outcomes

### TARGET — slow-viewer isolation

The integrated per-peer non-blocking delivery mechanism is the intended code-level isolation: a slow/lossy viewer may drop its own samples without blocking healthy peers. Repeated target-server impairment runs kept both healthy viewers smooth on `high`, with zero new peer-local audio/video drops, while only the constrained viewer changed tiers and stalled. This validates isolation for the reported three-device scenario, not for untested devices or architectures.

### TARGET — per-viewer adaptive quality

The repository contains a reproducible, opt-in three-tier VP8 profile, estimator diagnostics, resource/rollback documentation and an exact healthy-plus-constrained-viewer acceptance procedure. Target-server measurements confirmed the corrected bit/s accounting; lower constrained-tier rates, the full VP8 quantizer range and next-tier nominal upgrade gating made both shaped phases acceptably usable for the tested iPhone while preserving the healthy desktop and iPad. The revised decision policy requires sustained shortage plus recent peer-local receiver loss/NACK evidence. Ordinary upgrades retain the higher next-tier threshold, while a previously downgraded clean peer may use a bounded one-tier probe to escape an application-limited estimate; failed probes back off. This prevents a loss-free GCC target collapse from changing tiers without forcing `high` or disabling adaptive selection for documented lossy constrained phases. The receiver-evidence gate was target-validated at `2efcc6b1`, and its bounded recovery follow-up passed the focused exact-commit target gate at `ddf15cee`. The stable single-pipeline deployment remains the default. See [`ADAPTIVE_QUALITY.md`](ADAPTIVE_QUALITY.md).

### TARGET — mobile robustness

Mobile must reliably join, start media under autoplay rules, recover from transient failures, avoid persistent black screens, and retain usable touch/keyboard behavior.

The repository now separates the bounded recovery states: eight seconds for automatic recovery of an existing ICE peer; a four-attempt application-level login when the established peer/socket is unrecoverable; and the existing central Play control when Safari blocks playback after media is available. A successful media-element `srcObject` assignment is no longer treated as recovery until playback progress or track unmute occurs, and old stream timers/listeners are removed when the connection store resets.

The 2026-09-10 iPhone check proved successful rejoin after reload and, when required by iOS autoplay policy, a central Play tap; it also proved that this recovery did not disturb the other viewers. The new no-reload path is implemented and statically reviewed on `testing`, and its dependency-free recovery tests passed in the target-server validation container at `913a981e`. On 2026-09-11 the operator deliberately closed the checkpoint without the manual same-peer, replacement-session and bounded-exhaustion iPhone phases because no Safari Web Inspector/Mac was available and accepted the automated evidence for the deployment decision. Fully automatic in-place recovery therefore remains an unverified device claim; [`IOS_RECOVERY.md`](IOS_RECOVERY.md) retains the deferred procedure.

### TARGET — Smart-TV / constrained browser compatibility

A viewer should not be permanently excluded solely because WebRTC/ICE/media support is unreliable. Non-WebRTC receive-media paths are a target, and they do not need to use the same transport as interactive clients.

For passive/view-only devices such as Smart-TVs or constrained/older browsers, higher media latency is acceptable when it materially improves compatibility and stability. A conventional HTTP live-streaming path (preferably HLS/Low-Latency HLS as the first candidate, with DASH as an alternative to evaluate) should therefore be considered separately from low-latency interactive fallbacks.

### IMPLEMENTED IN REPOSITORY — server-enforced view-only share link

The optional `#/<16-character-Base64URL-token>` link permits passive WebRTC viewing in the same room without mouse, keyboard, touch, clipboard, file, microphone/media-share, control-request or admin capabilities. The token remains in the URL fragment and is carried in a WebSocket subprotocol; it is never placed in the normal share-link HTTP request path or query string. Six-character and legacy 64-hex credentials are rejected.

Authorization is enforced by a normalized backend-neutral `is_view_only` profile and allow/deny checks across every current ingress path, not by hidden controls or by WebRTC. This leaves the same identity usable by a future different receive-only media backend without creating a separate desktop/session or weakening authorization. Token lifetime, revocation, denial behavior, deployment, rollback and the three-role target-server matrix are specified in [`VIEW_ONLY_SHARING.md`](VIEW_ONLY_SHARING.md).

Runtime status: **the full view-only boundary, real inbound-media denial and revocation passed on the target server through `80eeca64`. At exact commit `913a981e`, fresh client/server checks, image deployment, the HTTP boundary probe and the compact-link browser smoke test also passed. View-only is accepted for the tested deployment; the separately omitted iPhone recovery deep test is not evidence against or for its no-reload behavior**.

### TARGET — event-triggered viewer stability and measured startup

The operator's 2026-10-04 report distinguishes older-TV WebRTC failures around joins/chat/control transitions from longer mobile time-to-first-picture. VIDAA is suspected but exact devices are unknown; the affected colleague's TV is unavailable and the existing WebCodecs path is untested there. HLS remains a compatibility candidate rather than a confirmed repair. Continue its separate Phase 4 target validation while retaining WebCodecs and keeping the TV acceptance explicitly open.

[`STABILITY_REVIEW.md`](STABILITY_REVIEW.md) fixes the reported evidence, the unconfirmed chat-sound candidate and the mandatory final static stability/cleanup/authorization review before grouped target-server validation and any promotion. Available-device checks can proceed independently; actual stability, quality and startup improvements require comparable target evidence.

### TARGET — robust recovery

Refresh, reconnect and network transitions must not leave peers permanently black or stuck.

## Alternative media direction

### IMPLEMENTED IN REPOSITORY — backend-neutral encoded-media subscription boundary and WebRTC compatibility adapter

The source-subscription and participant-delivery contract is fixed and implemented in [`MEDIA_SUBSCRIPTION_BOUNDARY.md`](MEDIA_SUBSCRIPTION_BOUNDARY.md). It separates demand on shared encoded sources from per-session delivery, centralizes `CanWatch` authorization above all backends, requires bounded non-blocking queues and makes format, GStreamer presentation timing, source generations and discontinuities explicit.

The capture-backed provider now owns source discovery, keyframe-gated subscriptions, switch/pause/resume/close, generation/format/discontinuity ordering and immutable event publication. The central delivery manager registers backend capabilities, issues credential-free session leases, allows `CanWatch` view-only receivers, rejects non-watchers and owns replacement, revocation, shutdown and generic `IsWatching` state. The existing WebRTC sender consumes the provider queue directly with capacity two and drop-new behavior, so no second queue was added. Its SDP/signaling, data channel, inbound-media authorization, adaptive tier selection and existing per-session metrics remain on the same protocol and configuration.

GStreamer now exports encoded-buffer PTS, DTS, duration and caps-derived resolution/frame rate. Provider normalization maps those timestamps onto a manager-owned timeline, while `CapturedAt` is retained only for the compatibility Pion sample field and is not treated as a cross-backend PTS. Pipeline recreation and format change advance/propagate explicit generations and force video readmission at a keyframe.

Focused repository tests cover ordered selection, demand lifecycle, keyframe admission, switch/pause/resume/idempotent close, timing/generation/format/discontinuity, local non-blocking overflow, the unchanged WebRTC two-unit policy, `CanWatch` denial, view-only receive allowance, replacement/revocation and shutdown. These tests were **NOT EXECUTED IN CODEX**. At `a32027d`, they and the server/plugin build passed in the target-server validation container; local images also built and the deployed compatibility path produced the expected lifecycle metrics. At exact follow-up `e5f55bf9`, the complete suite/build passed again, fresh images deployed healthy, and one fresh viewer remained on `high` throughout the focused estimator-startup trace with no video queue drop. The operator accepted this bounded checkpoint on 2026-09-12 for continued prototype work, while explicitly deferring repetition of the ordinary/admin/view-only/private-mode/reconnect matrix and an induced three-viewer adaptive down/up run to final grouped validation.

Static comparison found no subscription-refactor change to the accepted adaptive configuration, encoder construction or encoded payload bytes. A possible softer image during fast movement therefore remains an unproven quality observation consistent with the existing fixed-rate roughly 2-Mbit/s VP8 `high` tier and its full quantizer range. A controlled bitrate/quantizer A/B is still required before changing the profile or claiming a regression.

The exact first `webcodecs-ws` receive-prototype contract is fixed in [`WEBCODECS_MEDIA_WEBSOCKET.md`](WEBCODECS_MEDIA_WEBSOCKET.md). Its Phase 1 protocol/ticket/negotiation primitives, Phase 2 credential-free server delivery endpoint and Phase 3 isolated browser receive path are implemented. The endpoint owns strict pre-upgrade policy, bounded provider/egress/lifecycle queues, exact VP8/Opus serialization, READY/progress/resync state, private-mode and central delivery cleanup, and credential-safe metrics. The client now resolves a persisted per-browser WebRTC/WebCodecs default before connecting; absent, inaccessible or invalid storage stays on WebRTC, while exact `?media=webcodecs-ws` remains a stateless highest-priority diagnostic override. A stored WebCodecs choice starts the same dedicated worker without requiring the query. Selection changes persist the normalized choice, remove only the diagnostic parameter and re-enter normal startup; there is still no automatic fallback. Healthy WebCodecs streaming displays only the compact `WebCodecs` label, while negotiation, recovery and terminal failure retain the prominent receive-only/Picture-in-Picture limitation and explicit actions. Phase 4 repository assets add a separate sanitized overlay, fixed-metric PromQL/evidence collection and an interactive target-server runbook without altering base Compose. Exact `86893473` passed the full automated/image/deployment/security sequence and an uninterrupted roughly 15-minute zero-resync stream; its approximately 25 VP8 units/s are the configured cadence of the shared adaptive `high` source, not evidence of transport loss. Exact `150aac57` subsequently passed 19 client tests, TypeScript/build, fresh-image deployment and security probes. The operator confirmed ordinary/admin/view-only receive, private-mode pause/resume, explicit WebRTC return, desktop native fullscreen and iPhone viewport app fullscreen, including its documented browser-chrome limitation. The wider run nevertheless reproduced short black/audio interruptions and finally three rapid `client_audio_underflow` resyncs plus `resync_limit` on a visible foreground iPhone. Commit `95746173` changed transient AudioWorklet starvation to a local 160-ms rebuffer within the fixed 200-ms cap while preserving the canvas. At exact documentation commit `6b6cd328`, all 20 client tests, TypeScript/build, the complete server suite, 849,148 fuzz executions, trailing build, images, deployment and security probes passed. A fresh foreground-iPhone connection delivered 16,372 VP8 and 32,792 Opus units for more than ten minutes with no WebCodecs close, retry, server `client_audio_underflow`, progress timeout or restart; the operator reported picture and audio remained okay. This is a bounded target-browser checkpoint, not full prototype acceptance: presentation pacing, numeric latency, induced isolation/resource and remaining live hostile-input gates were explicitly deferred. At exact `12cfe43b`, all 26 client tests, TypeScript/build, fresh images, deployment and public/direct security probes passed. The focused browser matrix confirmed the persisted/default/override and compact-versus-terminal behavior, including no automatic fallback and explicit recovery to WebRTC; live view-only fragment preservation was not repeated, while the focused automated navigation test passed. HLS/LL-HLS Phases 1–3 and the separate Phase 4 deployment/observability assets are implemented; enabled target acceptance remains pending, and automatic backend selection and WebTransport remain later candidates.

The project distinguishes **interactive** and **passive/view-only** media fallback needs. There is not one mandatory fallback chain for every client. The passive HLS/LL-HLS version-1 behavior is specified normatively in [`HLS_LL_HLS.md`](HLS_LL_HLS.md). Its default-off negotiation, ticket/lease and security foundations plus the shared bounded packager and authenticated HTTP delivery are implemented in the repository. Phase 3 adds the isolated native/MSE passive client and advertised-only manual selection. Phase 4 adds separate opt-in deployment/observability and exact-image validation/rollback assets. Exact application 93f1fa63 passed automated/image preparation and, using helper 2484a022, the preserving Caddy/logging/healthy-HLS-activation and 19-denial-probe gate. The first actual playback attempt failed and normal login subsequently timed out, with a tentative /ws 101 report. The supplied read-only diagnostic found the prepared image healthy and HLS startup/source-restart counters without demonstrated readiness. Cold-start source/caps handling, HLS-only encoder timestamp mapping and bounded C logging are repaired in the repository, with a required real-codec integration gate; the new exact-commit automated/image gate passed at 80020d99, including both real-codec tests and four-rendition conventional readiness (test duration 18.11 seconds); live acceptance remains pending. The saved pre-HLS runtime was restored healthy with exit 0 and operator-confirmed normal login/picture/audio. The prepared 80020d99 image passed default-off deployment (Baseline-Exitcode 0, healthy, 2/2 disabled-route probes) and operator-confirmed normal login/picture/audio/control. Same-image HLS activation passed with Enable-Exitcode 0, healthy service and 19/19 HTTP denial probes, but HLS failed and the operator reported all streams stopped afterward. Read-only diagnosis found one ready packager/lease, 23 successful segment requests and two timeline-gap rejections, with no sampled Neko exit/OOM; default-off restoration passed with Recovery-Exitcode 0 and 2/2 disabled-route probes. Fresh browser confirmation is pending. HLS-only scene-cut suppression passed the isolated target GOP A/B gate at 97ba4ad9: the old code reproduced two timeline gaps, all three repaired codec tests passed and the scene-cut fixture stayed in generation 1 for 30.19 seconds. Exact 97ba4ad9 automated/image preparation passed with Repair-Check-Exitcode 0 (47 client tests, 13 Go packages, both fuzz jobs, all three codec tests and base/Brave builds). Default-off deployment of my-neko/brave:hls-97ba4ad9ab3e passed with Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes. The operator reported the requested normal browser check works. The subsequent HLS attempt at the prepared 97ba4ad9 checkpoint failed with "HLS bootstrap failed; retry manually", and the operator confirmed WebRTC also stopped working. The latest enablement CLI/HTTP results have not been supplied. Read-only diagnosis passed with 84 log lines, one not-ready bootstrap, one negotiation rejection and no sampled exit/OOM or generation/lease-open markers. Same-image default-off restoration passed with Recovery-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login/picture/audio work again. The saved-startup summary passed with Saved-Check-Exitcode 0: audio/high/medium subscriptions persisted, low capture reached only its create marker, and the saved matching environment had HLS enabled. Paired startup diagnosis at helper 88f2b25d completed with exit 0: both constructors passed two starts and failed one, with medium losing its initial pre-anchor IDR and reaching only two parents. This narrows the smooth readiness defect independently of registry isolation. The 414639d2 isolated anchor A/B gate passed with Anchor-Check-Exitcode 0: old code reproduced the fixed initial-IDR failure; seven corrected checks each passed three times, with every real-codec fixture staying in generation 1. Full exact-414639d2 preparation FAILED with Repair-Prepare-Exitcode 1: 47 client tests, type/build, 13 Go packages, both fuzz jobs, registry/mapping and anchor lifecycle checks passed, but smooth readiness restarted with worker_failure and failed its generation-1 assertion at 20.03 seconds; scene cuts passed at 30.19 seconds in generation 1. Base/Brave image steps were not reached. Worker diagnosis at helper 53034495 then completed with exit 0: all six checks passed in three fresh processes, each smooth fixture ready in 18.11 seconds in generation 1 with no rejected pushes; the earlier worker failure was not reproduced. The controlled audio-anchor A/B at repair 71a14d21 passed with Audio-Anchor-Exitcode 0: the old AAC hold reproduced both the blocked-drainage unit failure and an audio/anchor/queue_full restart under 256 ms high-input delay; all eight corrected startup checks passed in three fresh processes and scene cuts passed once (25 top-level passes). All seven real-codec fixtures stayed in generation 1; normal readiness was 18.10/18.11/18.11 seconds and delayed-high readiness 18.09 seconds each. This verifies the controlled AAC-overflow mechanism, not the cause of the earlier unobserved worker failure or live all-stream outage. Full exact-71a14d21 preparation then passed with Repair-Prepare-Exitcode 0: the rebuilt GStreamer 1.26.2 image passed all nine selected startup checks, including normal readiness at 18.11 seconds, delayed-high readiness at 18.09 seconds and scene cuts at 30.19 seconds, all in generation 1; base/Brave images were built and the service remained unchanged. The supplied tail starts inside the codec-image build; earlier client/Go/fuzz steps are covered by the script's reported final success, not separately shown in this excerpt. Default-off deployment of my-neko/brave:hls-71a14d2174da then passed with Baseline-Exitcode 0, a healthy service and 2/2 disabled-route probes. The operator confirmed the requested normal browser check works without HLS at 71a14d21. Same-image conventional-HLS activation at 71a14d21 then passed with Enable-Exitcode 0, healthy service, a private enable snapshot and 17/17 public plus 2/2 cleartext-denial probes. The subsequent operator-reported HLS attempt at 71a14d21 went from connecting to failed with "HLS bootstrap failed; retry manually"; the operator reported only WebRTC streaming works. No successful HLS picture/audio or room-event interval is demonstrated. Read-only diagnosis passed with 950 captured lines, one not-ready bootstrap, one started/idle-stopped packager generation, medium/low keyframe-admission drops of 759/564 and cumulative part/segment publication only for audio/high. No sampled Neko exit/OOM was found. Same-image default-off restoration passed with Recovery-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login/picture/audio work again. The isolated source-clock-phase diagnosis at helper 409482b4 passed with Clock-Skew-Exitcode 0: aligned control ready in 18.11 seconds; all three skew runs reproduced not-ready at 24.02 seconds in generation 1 with audio/high ready, medium/low blocked, flowing IDRs and no rejected native pushes. This proves the controlled phase-admission defect, not the exact unmeasured live phases. A common high-source fan-out repair is implemented and statically reviewed: one shared video provider subscription feeds the existing three scaled encoders, plus one audio subscription; provider PTS/DTS are preserved. The isolated common-source repair A/B passed at a7ffb8b1 with Shared-Clock-Exitcode 0: the old defect reproduced once; 46 positive top-level checks passed across three cold processes including one scene-cut check, with all ten real-codec fixtures in generation 1. Full exact-a7ffb8b1 preparation then passed with Repair-Prepare-Exitcode 0: all thirteen selected native/startup checks passed in the rebuilt GStreamer 1.26.2 image, including the four real-codec fixtures in generation 1, and base/Brave images were built without changing the running service. The supplied excerpt starts inside the codec-image build; earlier client/Go/fuzz stages are covered by final script success without separately shown fresh counts. NEXT exact-a7ffb8b1 default-off deployment and normal browser confirmation; same-image enablement and valid/live acceptance pending; the all-stream symptom's cause remains unconfirmed. The original login blocker is not established. See [the repair record](HLS_STARTUP_REPAIR_2026-10-04.md). Working playback, passive authorization/lifecycle and grouped device/resource acceptance remain pending; the host Caddy review is in HLS_LL_HLS_CADDY.md.

This direction is consistent with upstream issue #371, which proposes protocol-independent media backends including HLS (`m3u8`), WebRTC and QUIC, with backend selection based on device capabilities, network conditions and server capabilities.

### PARTIALLY IMPLEMENTED TARGET candidate — interactive receive fallback

Upstream issue #690 proposes staged work around:

1. media-subscription abstraction;
2. WebCodecs + dedicated WebSocket media;
3. WebTransport.

For an interactive participant, WebCodecs + WebSocket remains the first practical non-WebRTC receive candidate to evaluate after the media abstraction. Its exact version-1 design uses an explicitly selected, default-off dedicated media socket, a credential-free one-time attachment ticket, server-owned `CanWatch` lifecycle, VP8 plus raw Opus capability probes, a strict binary envelope, bounded non-blocking queues, drop-to-keyframe recovery and a shared A/V clock. It does not silently fall back to or from WebRTC.

This first prototype is intentionally receive-only. The current client sends high-rate mouse, keyboard and touch input through the WebRTC data channel; a replacement control transport is a separate later decision. Until that exists, the prototype must not be represented as complete non-WebRTC interactive parity. The normative contract, security limits, implementation phases, rollback and target-server gates are in [`WEBCODECS_MEDIA_WEBSOCKET.md`](WEBCODECS_MEDIA_WEBSOCKET.md).

Do not assume WebSocket is automatically better on a poor or lossy connection: the standard `WebSocket` API has no built-in backpressure, and reliable ordered delivery can accumulate latency if the receiver cannot keep up. Queueing/drop policy, codec support and actual device behavior must be measured.

### PARTIALLY IMPLEMENTED TARGET candidate — passive/view-only HTTP streaming

For passive viewers, especially Smart-TVs, old/constrained browsers, and share-link clients, evaluate a conventional adaptive HTTP livestream independently of the interactive path:

- **HLS / Low-Latency HLS** is the primary candidate because it is HTTP-based, supports live audio/video and multiple bitrate variants, and is designed to adapt playback to changing network conditions.
- **MPEG-DASH** is a secondary candidate where client/platform support makes it useful.
- Higher latency is acceptable for this role because these clients are not controlling the desktop.
- A passive HTTP-stream client must remain part of the same room/session and receive no implicit control rights.
- Prefer using the same underlying capture/encoded-media abstraction where practical rather than introducing a second independent desktop capture.

This is conceptually similar to a conventional Twitch/YouTube-style viewer delivery path, not a requirement to reproduce either service's exact protocol stack.

The first contract fixes H.264 High 3.1 plus stereo AAC-LC in separate fMP4 renditions, six-second parent segments, one-second LL-HLS parts, one shared bounded packager set, per-viewer central deliveries and short-lived path-scoped HttpOnly cookie leases whose bearer never appears in a playlist URL. Same-origin HTTPS, no shared cache/CDN, bounded memory-only retention, private-mode/revocation behavior, explicit rollout phases and target-device/latency/resource/isolation gates are fixed in [`HLS_LL_HLS.md`](HLS_LL_HLS.md). Phases 1–3 implement server access/packaging/HTTP delivery and the isolated explicitly selected native/MSE passive client. Focused automated target-server checks passed at `e85d8568`, and the default-off image deployment started healthy at `741025c3`; HLS-enabled runtime and device gates remain pending. Repository implementation is not a device-support claim.

### LATER/OPTIONAL — ultra-legacy image fallback

MJPEG may be evaluated only as a last-resort, ultra-legacy image-only fallback if a concrete device class cannot use the other media paths. It is not a primary architecture target because it lacks an integrated audio/adaptive-streaming model and is bandwidth-inefficient for continuous desktop video.

### Selection principle

The eventual media backend may differ per participant based on role and capability:

```text
interactive capable client  -> WebRTC
interactive WebRTC failure  -> WebCodecs/WebSocket candidate
passive/view-only client     -> HLS/LL-HLS candidate
other constrained client     -> capability-tested alternative
```

Automatic selection is a TARGET direction, not currently IMPLEMENTED. Manual selection is implemented in the repository: the existing sidebar persists a per-client `WebRTC`/`WebCodecs` default, WebRTC is the absent/invalid value, and the exact URL parameter remains a stateless highest-priority diagnostic override. The active backend is reported in settings; healthy WebCodecs streaming uses a compact `WebCodecs` in-player indicator, while the large status and explicit actions remain for negotiation, recovery or terminal failure. The focused exact-commit target-server checkpoint passed at `12cfe43b` with the live view-only fragment case explicitly omitted; this does not authorize automatic fallback or broaden the earlier prototype-acceptance claim. HLS/LL-HLS Phase 3 now adds advertised-only manual choices for passive viewers/admin diagnostics plus exact diagnostic overrides, with WebRTC default and no automatic fallback. Focused automated target checks passed at `e85d8568`; HLS-enabled runtime/device acceptance remains pending.

The repository implements the WebCodecs/media-WebSocket receive path through Phase 3, the separate Phase 4 deployment/observability/validation assets and the manual per-client selection/status productization. The server feature remains absent unless `docker-compose.webcodecs-ws.yaml` is explicitly included. A new/unset client sends no prototype request and starts no prototype worker; a client that deliberately stores WebCodecs can start it without a query, and the exact query can still force it for a single diagnostic URL. The bounded exact `6b6cd328` target checkpoint passed automated, build, deployment/security, accumulated functional and corrected foreground-iPhone evidence; exact `12cfe43b` then passed the focused productized selection/status checkpoint with the documented live-fragment limitation. The controlled presentation-pacing A/B and wider numeric latency/isolation/resource/live-hostile-input matrix remain explicitly unexecuted. Default WebRTC behavior remains unchanged, and the result is not represented as full prototype acceptance.

## Development / verification environment constraint

### Workflow invariant

Codex is an editing/static-review environment only and is not representative of the deployment server.

Codex may inspect source, configuration, history and diffs, but must not install dependencies, run builds/tests/linters, start the application, invoke Docker/project scripts, or perform WebRTC/device runtime tests.

Runtime/build/test verification occurs separately on the real server. Until server results exist, changes may be described as statically reviewed but not runtime-verified.

## Security / deployment constraints

- authorization remains server-side;
- do not expose control/admin capability through a view-only token;
- do not commit credentials or persistent browser profiles;
- downloads/browser profiles/cookies are runtime/user data;
- local deployment differences must be sanitized before any reusable example is committed.

### AUDITED local snapshot

The supplied `MyNekoProjekt` snapshot was completely compared with the fork baseline on 2026-09-09. It contained no missing application-source improvement. After the operator clarified that the differing compose defines the fork's intended deployment, its reusable structure was reconstructed in tracked, parameterized form. Credential values, the empty instance policy, browser profile and downloads remain outside Git. See [`LOCAL_DELTA_AUDIT.md`](LOCAL_DELTA_AUDIT.md).

## Non-goals

- one session per viewer;
- discarding fork UX/mobile behavior merely to simplify upstream merging;
- implementing every upstream v3 idea before restoring a stable baseline;
- treating UI-hidden controls as authorization;
- copying the local deployment tree wholesale into the public repository.

## External references

- upstream: https://github.com/m1k1o/neko
- v3 architecture direction: https://github.com/m1k1o/neko/issues/371
- alternative media proposal: https://github.com/m1k1o/neko/issues/690
- mobile trackpad issue: https://github.com/m1k1o/neko/issues/640
