# officialtaxz2/neko

Customized fork of [m1k1o/neko](https://github.com/m1k1o/neko), a self-hosted shared virtual browser/desktop streamed to multiple participants.

This fork keeps Neko's shared multi-user session model and adds substantial client-side work around mobile/touch use, trackpad-style control, playback/reconnect recovery, and UI/UX.

## Status

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
- The exact default-off HLS/Low-Latency HLS passive/view-only version-1 contract is fixed in [`docs/HLS_LL_HLS.md`](docs/HLS_LL_HLS.md). Phase 1 now implements only its access and deterministic-media foundations: validated `media.hls` configuration, authenticated HLS event negotiation, hashed ten-second one-use tickets, path-scoped 30-second playback-cookie leases, exact HTTPS/origin/proxy/path/query/range/rate/redaction helpers, immutable bounded object/generation models and golden conventional/LL-HLS playlists. No HLS HTTP route, packager, encoder, player, deployment overlay or automatic selection exists yet.
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

Continue exclusively on `testing`: implement **HLS/LL-HLS Phase 2** from [`docs/HLS_LL_HLS.md`](docs/HLS_LL_HLS.md): the shared bounded H.264/AAC fMP4 packager, HTTP bootstrap/resource delivery, central per-session delivery attachment, fixed observability and slow-reader/shutdown isolation. Do not add a client player yet. WebRTC remains the default, WebCodecs remains explicit, no fallback is automatic, and `master` remains pinned until the operator authorizes a later grouped promotion.

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

`MyNekoProjekt` contains instance-specific deployment/runtime material. Its desired Compose structure is tracked in sanitized, parameterized form; credentials, browser profiles, downloads, cookies and lock/runtime files stay outside Git.

The completed local classification is recorded in `docs/LOCAL_DELTA_AUDIT.md`.
