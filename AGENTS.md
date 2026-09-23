# AGENTS.md

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
24. implement HLS/LL-HLS Phase 1: default-off configuration, authenticated bootstrap/playback-lease foundations, deterministic playlist/object models and golden fixtures without starting a packager or adding a client player — **complete in the repository on `testing`; target-server tests/build pending**,
25. implement HLS/LL-HLS Phase 2: shared bounded packager, H.264/AAC fMP4 generation, authenticated HTTP delivery, central per-session attachment and fixed observability without adding a client player — **complete in the repository on `testing`; target-server tests/build/runtime validation pending**,
26. implement HLS/LL-HLS Phase 3: isolated passive client, pinned player support, explicit advertised-only HLS/LL-HLS selection and bounded lifecycle cleanup without automatic fallback — **NEXT**,
27. keep accumulating reviewed implementation blocks on `testing`; promote to `master` only after the operator explicitly authorizes the final grouped promotion — **pending**.

## Authoritative knowledge

Read before substantial work:

- `README.md` — compact entry and current status.
- `docs/PROJECT.md` — goals, requirements, invariants and target state.
- `docs/ARCHITECTURE.md` — verified current architecture.
- `docs/WORKPLAN.md` — current `NEXT`, sync procedure, server-side verification and open items.
- `docs/LOCAL_DELTA_AUDIT.md` — completed, sanitized classification of the supplied `MyNekoProjekt` snapshot.
- `docs/UPSTREAM_SYNC_AUDIT.md` — completed semantic review and merge record for the 2026-09-09 upstream synchronization.
- `docs/ADAPTIVE_QUALITY.md` — opt-in profile, diagnostics, resource costs, exact target-server acceptance procedure and rollback.
- `docs/IOS_RECOVERY.md` — implemented bounded reconnect states, exact target-server iPhone procedure, acceptance criteria and rollback.
- `docs/VIEW_ONLY_SHARING.md` — implemented passive-session boundary, token lifetime/revocation, denial behavior, exact three-role target-server matrix and rollback.
- `docs/MEDIA_SUBSCRIPTION_BOUNDARY.md` — decided backend-neutral encoded-media/source-subscription and participant-delivery contract, migration order, security boundary and prototype gates.
- `docs/HLS_LL_HLS.md` — exact version-1 passive HLS/LL-HLS codec, packaging, authorization, HTTP security, retention, rollout and acceptance contract.
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
- WebCodecs/WebSocket Phases 1–3 are IMPLEMENTED: protocol/ticket/negotiation, the credential-free server route/delivery backend and the receive-only client parser/worker/WebCodecs/AudioWorklet/canvas path. Manual client productization persists a per-browser WebRTC/WebCodecs default, resolves absent/invalid storage to WebRTC and retains exact `?media=webcodecs-ws` as a stateless highest-priority diagnostic override. Healthy WebCodecs streaming uses only the compact `WebCodecs` label; negotiation, recovery and terminal actions remain prominent. The focused exact-commit target checkpoint for that productization passed at `12cfe43b`, except that live view-only fragment preservation was not repeated and remains backed only by focused automated coverage. Phase 4 repository assets add only a separate explicit deployment overlay, credential-safe observability and an interactive acceptance runbook. Its wider bounded target-server checkpoint passed the recorded exact automated/build/security, functional and corrected foreground-iPhone gates at `6b6cd328`, but the deliberately omitted numeric latency, induced isolation, resource and remaining hostile-input matrix prevents a full prototype-acceptance claim. The server remains default-off when that overlay is omitted, and new/unset clients remain on WebRTC. HLS/LL-HLS Phases 1–2 implement default-off negotiation, tickets/leases, strict HTTP security, deterministic models, a shared bounded H.264/AAC fMP4 packager and authenticated server delivery; there is still no HLS client player, deployment overlay, client selection or runtime acceptance. Do not assume WebSocket is inherently better for poor networks, and do not promote MJPEG beyond an optional ultra-legacy fallback without device evidence.

## Definition of Done for Codex work

A Codex task is complete when:

- the requested repository change is implemented and reviewed statically;
- relevant diffs and surrounding code are inspected for regressions;
- no secrets/runtime data are committed;
- required target-server verification is explicitly listed as pending;
- durable docs are updated when project truth/status changed.

Do not claim runtime/build/test success without results from the real target server.
