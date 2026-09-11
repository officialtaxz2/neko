# Work Plan / Handoff State

Last consolidated: 2026-09-11.

## Environment boundary

Codex is the development/static-review environment only.

Codex must not install project dependencies, execute builds/tests/linters/type-checkers, start the client/server/Docker, execute repository scripts, or perform WebRTC/media/device runtime validation.

All runtime/build/test verification happens later on the real target server.

Codex should still inspect source/config/diffs thoroughly and prepare exact server-side verification steps.

## Verified repository snapshot

Pre-sync fork state:

```text
master and safety-pre-upstream-sync-20260909
18e9320c892b4069757a71c9093c9c9b4dd7bd4a
```

Upstream fetched on 2026-09-09:

```text
m1k1o/neko:master
b0f01cedea68893e85a3fd852c0521238c285695
```

Pre-sync merge base and divergence:

```text
d74052bb844c43a0cc3c2386d083f7505dc483a2
fork-only: 36 commits
upstream-only: 33 commits
```

Current integration state:

```text
integration/upstream-20260909
upstream merge commit: 4e99b8d3ca720d1f184544306820e388716ba23a
relation at merge commit: 37 commits ahead, 0 behind
master: fast-forwarded to the reviewed integration history
testing: deployment reconciliation, accepted opt-in adaptive quality, bounded iOS recovery and server-enforced view-only sharing
master: pinned at d9105ef8 until explicit grouped-promotion authorization
```

## COMPLETED — `MyNekoProjekt` local delta audit

Completed on 2026-09-09 against fork baseline `d9c1afd564ad4c293a0b1b28aecbc103d1d1d9c6`.

- Safety branch `safety-pre-local-delta-audit-20260909` preserves the original fork HEAD.
- A clean baseline and the complete supplied local tree were compared recursively, excluding `.git/`.
- All 724 baseline files exist locally: 138 are byte-identical, 585 differ only by CRLF/LF, and only `docker-compose.yaml` differs substantively.
- The 3,929 local-only files are 3,928 persistent browser-profile/runtime files under `files/**` plus an empty instance `policy.json`; `downloads/` is empty.
- No application-source feature/fix was missing.
- The raw local compose and embedded credentials were initially excluded together. After the operator clarified that its structure is the intended deployment baseline, that structure was reconstructed in sanitized and parameterized form; credential values remain excluded and should be rotated operationally if still active.
- Root ignore rules guard the local reference and runtime/profile paths against accidental staging.
- Static follow-up repaired the stale pre-Vite client lock/type baseline in commit `2d89027e`; its target-server client checks were later operator-confirmed as passed.

Status: **Codex-side implementation complete / target-server verification operator-confirmed as passed on 2026-09-09**.

The exhaustive sanitized classification is in [`LOCAL_DELTA_AUDIT.md`](LOCAL_DELTA_AUDIT.md).

## COMPLETED — semantic upstream synchronization

Completed in Codex on 2026-09-09 on `integration/upstream-20260909`.

- Fetched and pruned the `upstream` remote, then recomputed the merge base and divergence instead of relying on the bootstrap snapshot.
- Created `safety-pre-upstream-sync-20260909` at pre-sync HEAD `18e9320c`.
- Reviewed the 33 upstream-only commits and their 97-file delta by client, server/media, plugins, runtime/browser images, deployment/CI and documentation.
- Merged upstream `master` at `b0f01ced` in merge commit `4e99b8d3`.
- Resolved conflicts only in `client/src/components/settings.vue`, `side.vue` and `video.vue`; both sides' intended behavior was retained.
- Preserved fork touch detection, trackpad/cursor behavior, redesigned side/settings UI, autoplay/playback recovery, fullscreen, reconnect and cursor event cleanup.
- Accepted upstream Open-in-App settings, file-transfer capability gating and clipboard synchronization on window focus.
- Found and repaired a non-conflicting integration defect: the fork demo-mode file list now supplies the new download/upload/delete capability fields.
- Kept `MyNekoProjekt`, credentials, browser profiles, downloads, lock files, runtime state and instance-only policy/raw-compose data out of the upstream merge. The desired compose structure was handled separately afterward.

Static inspection completed:

- no unmerged files or conflict markers;
- `git diff --check` clean after removing one upstream blank line at EOF;
- changed JSON files parse successfully;
- client/server event names, file capability payloads, store registration, H.265 configuration paths and removed `ArrayIn` references cross-checked;
- no staged forbidden runtime paths or high-confidence private-key/token patterns.

Runtime/build/test status: **NOT EXECUTED IN CODEX**.

The detailed merge record is in [`UPSTREAM_SYNC_AUDIT.md`](UPSTREAM_SYNC_AUDIT.md).

## COMPLETED — sanitized deployment-compose reconciliation

Completed after operator clarification on 2026-09-09.

- Replaced Upstream's generic Firefox compose example with the desired local Brave deployment structure from `MyNekoProjekt`.
- Default image is the locally built `my-neko/brave:latest`, overridable through `NEKO_IMAGE`; `pull_policy: never` prevents accidental registry substitution.
- Preserved singleton-lock cleanup, persistent Brave profile/download mounts, managed-policy mount, loopback port `8082`, UDP range `51000-51100`, privileged/capability settings and file transfer.
- Parameterized image, bind address, port, UDP range, screen and host paths.
- Added `.env.example`; both member/admin passwords are required and deliberately empty. Real `.env` remains ignored.
- Kept the browser profile, downloads and instance `policy.json` contents ignored and outside Git.

Status: **statically reviewed in Codex / target-server deployment smoke test passed after the policy-mount correction described below**.

## COMPLETED — target-server validation

Operator-reported target-server result on 2026-09-09:

- The locally built Brave deployment reached an operational state through the tracked Compose structure.
- The first profile/policy check exposed a filename regression in the sanitized Compose reconstruction: it mounted the optional external policy as `/etc/brave/policies/managed/policy.json`, while the audited source compose, Brave image and inherited documentation use `/etc/brave/policies/managed/policies.json`.
- The plural destination was restored in `docker-compose.yaml`.
- After recreating the deployment, the operator confirmed that the custom managed policy and persistent browser profile load as intended.

The operator subsequently confirmed that all build checks and all regression-matrix items applicable to the target deployment were tested successfully. Raw command logs, exact device models and non-applicable architecture variants were not archived in the repository, so this acceptance must not be generalized into claims for unreported devices or architectures.

Status: **target deployment accepted after correction / applicable integrated-baseline validation complete**.

## COMPLETED IN REPOSITORY — opt-in adaptive quality and diagnostics

Implemented and statically reviewed on 2026-09-09:

- Added `docker-compose.adaptive.yaml` as a separate overlay; omitting it preserves the stable single-pipeline Brave deployment.
- Added `deploy/adaptive-quality.yaml` with VP8 tiers ordered `high`, `medium`, `low`, explicit encoder targets and explicit per-peer estimator settings.
- Added `docs/ADAPTIVE_QUALITY.md` with activation, metrics/log diagnostics, resource costs, rollback and an exact two-healthy-plus-one-constrained-viewer acceptance sequence.
- Added `deploy/collect-adaptive-quality.sh` and a result template to capture filtered phase metrics, estimator logs, host resources, commit and running image ID without reading deployment secrets.
- Found a factor-of-eight unit defect while tracing the selection path: capture accumulated encoded bytes/s but compared it directly with Pion's bit/s estimate. Capture now converts sample bytes to bits and publishes the cross-goroutine bitrate atomically.
- Added `neko_capture_streamsink_bitrate` for actual per-pipeline bit/s and `neko_webrtc_track_dropped_samples_total` for peer-local queue drops labeled by session and media kind.
- Added a focused Go unit test for the byte-to-bit conversion.

Static review status: **implementation complete in repository / runtime, build and tests NOT EXECUTED IN CODEX**.

The earlier operator-confirmed integration regression predates the new overlay, metrics and bitrate correction. It remains valid for the tested baseline, but it is not acceptance evidence for this new profile.

## COMPLETED — adaptive-quality target-server validation

Operator-reported evidence on 2026-09-10 from the `testing` branch:

- The server Docker build and focused capture bitrate unit test passed without installing Go on the host; the local base and Brave images built successfully.
- The adaptive Compose model started healthy, loaded all three pipelines, preserved the Brave profile and managed policy, and passed login, audio, video and control smoke checks.
- A clean three-viewer baseline kept both healthy viewers and the cellular constrained-viewer candidate on `high`, with zero peer-local audio/video queue drops and no host resource saturation.
- A container-network PCAP comparison isolated exactly one additional IPv4 WebRTC/UDP endpoint for the cellular viewer. Host shaping counters subsequently proved that only this endpoint entered the constrained class.
- A first `netem rate` attempt with its 1,000-packet queue was rejected as a test setup because it accumulated almost 1 MB of backlog and froze the constrained viewer. The host qdisc was restored completely.
- A bounded HTB plus 40-packet `netem` trial at 1.3 Mbit/s preserved smooth playback for both healthy viewers and kept their peer-local drop counters at zero. The constrained viewer stalled on `high`, changed to `medium` after about 30 seconds, then cascaded to `low` at about 45 seconds. Removing the constraint returned it to `high` without refresh in about 30 seconds; unused pipelines returned inactive.
- That bounded trial rejected the original estimator timing for this target: `stalled_duration: 24s` reacted too slowly and `downgrade_backoff: 10s` did not give the replacement tier enough settling time. The opt-in tuning candidate is now 8 seconds and 30 seconds respectively.
- A rerun with those timings and an exact IPv4/UDP endpoint filter moved the constrained viewer from `high` to `medium` within 20 seconds and recovered `low` to `medium` to `high` within 40 seconds. It still cascaded to `low` during the 1.3 Mbit/s phase and remained visually unusable under both constraints; both healthy viewers stayed smooth and their path and peer-local queues recorded zero drops.
- The shaper delivered the intended approximately 1.30 and 0.72 Mbit/s rates. Pipeline measurements instead showed persistent encoder overshoot: `high` reached 4,394,400 bit/s against a 1,996,800 bit/s target, and `low` ranged from 702,400 to 1,230,040 bit/s against a 499,200 bit/s target. Peak container CPU was 134.04% on an eight-CPU host and memory remained below 1 GiB, so host saturation does not explain the failure.
- Raising each VP8 `max-quantizer` from 20/24/28 to 56 corrected the unconstrained `high` stream to 1,996,120 bit/s against its 1,996,800 bit/s target. All three viewers stayed visually smooth on `high`, peer-local audio/video drops remained zero, and the host remained far from CPU or memory saturation.
- The constrained max-quantizer rerun again isolated the weak viewer: both healthy viewers stayed smooth on `high` and their peer-local drops remained zero. The weak viewer became usable only while a suitable `low` tier held, then oscillated upward despite an unchanged constraint: `medium` to `low` to `medium` to `high` during the 1.3 Mbit/s phase, and mostly `medium` before eventually reaching `low` during the 0.7 Mbit/s phase. Restoring the path made playback smooth immediately.
- Static tracing found a second selection defect behind that oscillation. The original `diff_threshold` was applied to both downgrade and upgrade decisions, while the upgrade compares the estimate with the current tier rather than the next tier. The server now exposes a separate, default-preserving `upgrade_diff_threshold`; its focused target-server test passed.
- A rebuilt-image run with `upgrade_diff_threshold: 1.30` removed upward oscillation: the weak viewer moved `high` to `medium` to `low` under 1.3 Mbit/s, held `low` throughout 0.7 Mbit/s, and later recovered through `medium` to `high`. Both healthy viewers remained smooth on `high` with zero peer-local drops. Recovery to `high` exceeded the 90-second observation window.
- That run still rejected the constrained bitrate budget. At 0.7 Mbit/s, `low` video measured 494,296 bit/s and audio 128,400 bit/s, totaling about 623 kbit/s before transport overhead and retransmission. The shaper delivered its full approximately 0.71 Mbit/s and dropped another 153,969 packets; the weak viewer showed repeated video freezes and audio loss. CPU peaked at 135.66% on eight host CPUs and memory at 920.5 MiB, so resource saturation remained excluded.
- The next opt-in candidate lowered `medium` to 748,800 bit/s and `low` to 332,800 bit/s, permitted `max-quantizer: 63`, and adjusted `upgrade_diff_threshold` to 1.75 for the new maximum adjacent ratio of about 2.67:1. Its target-server run made both constrained phases mostly watchable, kept both healthy viewers smooth on `high` with zero peer-local drops, and recovered the constrained viewer through `medium` to `high` within 45 seconds. Baseline `high` measured 2,057,672 bit/s against its 1,996,800 bit/s target without visible regression; CPU and memory again remained unsaturated.
- The 1.3 Mbit/s phase held `medium` from 30 through 90 seconds and ended at 569,136 bit/s video plus 130,968 bit/s audio. The 0.7 Mbit/s phase normally held `low` at 369,144 bit/s video plus 125,672 bit/s audio, but briefly changed `low` to `medium` at 75 seconds and back to `low` by 90 seconds; the operator correlated that excursion with the remaining larger interruption.
- Static review found why the nominal ratio threshold did not prevent that excursion: upgrade still divided the estimate by the content-dependent measured current rate. The final code candidate adds explicit per-pipeline `nominal_bitrate` metadata and evaluates an upgrade against the next tier's nominal target plus the normal 15% reserve. `Low` to `medium` therefore requires 861,120 bit/s, rejecting the observed 613,076 bit/s estimate. The current measured rate remains the fallback for all existing profiles that omit the field.
- The final candidate at `bfaca84e0bcf` passed the focused capture and WebRTC tests in the repository's server Docker image. The base and Brave images built successfully; the adaptive service started healthy with zero restarts and image ID `sha256:a3efc4573a26965b0b9ccc54a289278fcd570d206c8ddaea58719f5e1a2c69ca`.
- Its final unconstrained baseline placed all three viewers on `high`, kept every peer-local audio/video drop counter at zero, and measured the high pipeline at 1,708,808 bit/s. The previously passing 1.3 Mbit/s evidence remains applicable because the final code change only makes upgrades stricter by using the unchanged next-tier nominal target.
- In the affected 0.7 Mbit/s rerun, the iPhone moved from `medium` to `low` by 45 seconds and held `low` through the 90-second phase; the previous low-to-medium excursion did not recur. Once on `low`, the operator found playback watchable and largely fluid with at most small occasional interruptions. Both healthy viewers remained smooth on `high`, their drop counters stayed at zero, and their unshaped `fq_codel` class recorded no drops.
- Removing impairment returned the iPhone through `medium` to `high` within 45 seconds. The final state had three `high` listeners, inactive zero-listener `medium`/`low` pipelines, a 2,001,544 bit/s high stream, 102.74% container CPU and 903.1 MiB RAM on the eight-CPU host. The host qdisc was restored to its original `fq_codel` configuration.
- Refresh/rejoin created a healthy new iPhone session on `high`. After a 15-second flight-mode interruption, another new iPhone session was already healthy on `high` at the first 10-second sample and stayed connected for the full 90-second observation; the desktop and iPad remained smooth on `high` with zero drop deltas. The operator confirmed playback after reload and an iOS Play tap when required. This accepts the manual recovery fallback and cross-peer isolation, but does not claim automatic in-place iOS recovery without user action.

Status: **accepted on 2026-09-10 as an opt-in profile for the documented target deployment and three-viewer scenario**. The stable base Compose remains single-pipeline, and profiles without `nominal_bitrate` retain their previous estimator behavior. This bounded acceptance must not be generalized to untested devices, architectures or network conditions.

## COMPLETED IN REPOSITORY — bounded iOS transient recovery

Implemented and statically reviewed on `testing` on 2026-09-10:

- Preserved the existing-peer path: ICE `disconnected` still has an eight-second window to return to `connected`/`completed` without replacing the peer or server session.
- Added a bounded application-level path after an established peer/socket becomes irrecoverable. One timer owns four serialized attempts after 1, 2, 5 and 10 seconds, using the same in-memory login values as a manual reconnect.
- Kept initial-login failures manual and suppressed retries for logout, demo mode and every server-directed `system/disconnect`, so recovery cannot fight authentication, kick or other session intent.
- Prevented the remounted login component from starting a parallel automatic login while the client owns recovery; after exhaustion or suppression, the form remains populated for manual use.
- Guarded WebSocket, peer and data-channel callbacks by object identity, cleared buffered ICE candidates during teardown, and stopped stale async offer work from sending through a replacement socket.
- Removed old media-stream listeners and the delayed `removetrack` timer when the store resets or a new stream arrives.
- Corrected the media-element attempt bound: reassigning `srcObject` no longer resets the counter by itself; only actual playback progress, track unmute or a new stream does. `video.load()` remains prohibited.
- Preserved Safari's autoplay behavior as a separate final stage: normal/muted playback is attempted when media is ready, and the central Play overlay remains visible if user activation is required.
- Extracted and covered the reconnect eligibility and delay decisions with dependency-free Node tests in `client/tests/recovery.test.mjs` and an `npm test` script.
- Added [`IOS_RECOVERY.md`](IOS_RECOVERY.md) with exact target-server build commands, same-peer and replacement-session interruption phases, bounded-exhaustion/auth checks, acceptance criteria and rollback.

Static review status: **implementation complete in repository / client test, lint, build and iPhone runtime verification NOT EXECUTED IN CODEX**.

Target-server status: **pending at the next coherent grouped `testing` checkpoint**. The prior 2026-09-10 reload/Play evidence does not validate the new no-reload implementation, and no automatic iOS claim is made yet.

## COMPLETED IN REPOSITORY — server-enforced view-only sharing

Implemented and statically reviewed on `testing` on 2026-09-10:

- Added an optional 96-bit multi-user bearer token encoded as exactly 16 Base64URL characters and the compact fragment share form `#/<token>`. The browser keeps the token out of HTTP paths/query strings and carries it in a validated WebSocket subprotocol; six-character and legacy 64-hex forms are rejected.
- Added the backend-neutral `is_view_only` member/session marker. Login, session creation and profile update normalize it so conflicting admin, host, media-share, clipboard or cursor-send fields cannot restore interaction.
- Kept the viewer in the same logical room and on the existing receive-only WebRTC media path; the authorization marker does not depend on that transport and can be reused by a later passive backend.
- Audited and closed the authenticated HTTP room/API, current WebSocket, legacy protocol, modern/legacy data channel, inbound media track and plugin boundaries. View-only input is denied before core/plugin dispatch, host assignment refuses passive targets, microphone/camera input is stopped, and file capability/list data is withheld.
- Reduced the client to video/playback controls for a connected passive session while preserving server authorization as the actual boundary.
- Defined deployment-bound lifetime and hard revocation: rotate/remove the token and recreate the service. View-only sessions are neither written to `session.file` nor restored from stale serialized data.
- Added focused Go tests for token configuration/authentication, profile normalization, middleware denial, non-persistence and current/legacy WebSocket and WebRTC data-channel allowlists.
- Added [`VIEW_ONLY_SHARING.md`](VIEW_ONLY_SHARING.md) with threat boundary, activation, denial behavior, revocation, rollback and an adversarial ordinary/admin/view-only target-server matrix.

Static review status: **implementation complete in repository / client checks, server tests/build and runtime matrix NOT EXECUTED IN CODEX**.

Target-server status: **the security/role matrix through `80eeca64` and the compact-link follow-up at exact commit `913a981e` passed; the coherent grouped checkpoint remains open until the bounded iOS recovery phases and final shared regressions pass on that deployment**.

Checkpoint evidence begun on 2026-09-11 at `5387f356` and continued through `913a981e`:

- containerized client checks, focused Go tests/server build, the deployment image build and the automated view-only HTTP boundary probe passed;
- the first browser role check reached a connected receive-only WebRTC session, but the passive client remained black/silent and raised `Cannot read properties of undefined (reading 'muted')` during autoplay fallback;
- the same check also confirmed that the view-only shell had hidden the shared playback toolbar together with the interactive room UI;
- after the client repair and image rebuild at `9a449d32`, V received desktop/audio, retained only Play/Pause, Mute/Volume, Fullscreen and PiP, and produced no browser-console error;
- M/A control behavior, refusal to assign V as host, legacy WebSocket/data-channel denial and current WebSocket denial passed without disconnecting V or disturbing M/A;
- the inbound-media phase did not reach the server: forced microphone enablement for V and the normal M baseline both stopped at the client's existing `negotiation is needed (no-op)` handler. This is a pre-existing microphone-passthrough defect rather than evidence for or against the view-only server boundary;
- `80eeca64` restored guarded client-initiated SDP offers after the initial server negotiation, serialized local offer creation, rejected stale peer/socket completion and cleaned up microphone streams that resolve after reconnect;
- all four dependency-free client tests, TypeScript lint and the production build then passed in the target Docker validation service. Rebuilt image `sha256:2dd587c1de59e9d51c1b64abd9c382506949a1a69d3a4645c792f0758096f922` started healthy;
- M delivered a real Opus microphone track to the server and stopped it normally. V delivered the same real track, which the server immediately stopped with `media sharing is disabled for this session`;
- V refresh and a 12-second offline/online cycle recovered as view-only without disturbing M/A. Both active viewers remained on `high` with zero audio/video sample drops;
- token rotation recreated the healthy service, disconnected V, made the old link return `Unauthorized`, and passed the HTTP boundary probe with the new token. The new V link and unchanged M/A logins were operator-confirmed.
- at exact commit `913a981e`, all six dependency-free client tests, TypeScript lint, the Vite production build, focused Go suites and the server/plugin build passed in the Docker validation services;
- the local base and Brave images rebuilt successfully, and `my-neko/brave:latest` became image `sha256:5b5c7747930bc863da05cb3f05d8bb7b97a624386d105bdade2448e5443d36a2` before the adaptive service was recreated with the new 16-character token;
- the complete automated HTTP boundary probe passed against the recreated service. The operator then confirmed that the exact compact `#/<16-character-token>` link automatically logged V in, showed desktop and audio, exposed only playback functions and produced no red browser-console error. Six-character and legacy 64-hex rejection are covered by the passing exact-commit client/server tests.

After the successful original matrix, the operator requested the intentionally incompatible compact `#/<16-character-Base64URL-token>` form. Its fresh exact-commit checks, deployment, HTTP probe and new-link browser smoke test are now complete. The full iOS recovery phases and final shared regressions remain open, so no grouped checkpoint acceptance is claimed yet.

## Target-server verification

This phase is performed by the operator on the real server, not by Codex.

### Client

The grouped iOS/view-only checkpoint does not require Node.js or npm on the target host. From the repository root, the dedicated validation Compose file runs the complete client sequence in a short-lived Node container and keeps `node_modules`/`dist` off the host:

```bash
export NEKO_VALIDATION_COMMIT="$(git rev-parse HEAD)"
docker compose -f docker-compose.validation.yaml pull client-checks
docker compose -f docker-compose.validation.yaml run --rm client-checks
```

The container executes the following equivalent commands:

```bash
cd client
npm ci
npm test
npm run lint
npm run build
```

Required for the integrated baseline: confirm `npm ci`, TypeScript lint and the Vite production build at integration commit `4e99b8d3`. The operator later confirmed these checks passed, closing verification of the lock/type repair in `2d89027e`.

For the accumulated iOS recovery and view-only blocks on `testing`, run all four commands at the exact candidate commit and then follow [`IOS_RECOVERY.md`](IOS_RECOVERY.md) and [`VIEW_ONLY_SHARING.md`](VIEW_ONLY_SHARING.md). These client blocks have not yet been executed on the target server.

### Server and container

The same checkpoint can run the focused Go tests and server/plugin build without Go on the host:

```bash
export NEKO_VALIDATION_COMMIT="$(git rev-parse HEAD)"
docker compose -f docker-compose.validation.yaml build --pull server-checks
docker compose -f docker-compose.validation.yaml run --rm server-checks
```

The container executes the following equivalent checks:

```bash
cd server
go test ./pkg/types ./pkg/auth ./internal/member/multiuser ./internal/session ./internal/http/legacy ./internal/websocket ./internal/webrtc
./build
```

A standalone image build can additionally be run:

```bash
docker build ./server
```

It does not replace the focused tests. The complete copy/paste sequence, safe output capture, metrics helper and manual device/role phases are documented in [`IOS_RECOVERY.md`](IOS_RECOVERY.md) and [`VIEW_ONLY_SHARING.md`](VIEW_ONLY_SHARING.md).

From the repository root, build the exact local base and Brave images referenced by Compose:

```bash
./build my-neko/base:latest -y
./build my-neko/brave:latest -y
```

Then prepare the ignored deployment environment and validate the Compose model without printing expanded credentials:

```bash
cp .env.example .env
# Set both passwords, an optional generated view-only token and any host-specific values in .env.
docker compose config --quiet
docker compose up -d
```

Confirm that `my-neko/brave:latest` is used, the cleanup service completes successfully, the profile/download/policy mounts resolve to the intended server paths, and the HTTP/UDP bindings match the firewall/reverse-proxy setup. The managed policy destination must be `/etc/brave/policies/managed/policies.json`.

The base image no longer embeds root `config.yml`; implicit hosting and cookie authentication now default to disabled. If clients cannot reach the discovered address, set `NEKO_WEBRTC_NAT1TO1` in `.env` and enable the documented Compose entry. Never put real passwords into tracked YAML.

### Manual regression on target server

Access/control:

- admin and regular-user login;
- only one active controller;
- request/grant/revoke;
- lock enforcement;
- admin behavior under lock;
- expected behavior with implicit hosting and cookie authentication explicitly configured or intentionally left disabled.

Slow-viewer isolation and adaptive quality:

- run at least two healthy viewers and one bandwidth-/latency-constrained viewer concurrently;
- confirm the constrained peer may drop/degrade without freezing or adding latency to healthy peers;
- inspect server logs/metrics for peer-local sample drops and absence of global capture blockage;
- with multiple ordered video pipelines and the bandwidth estimator enabled, throttle and restore one viewer;
- confirm only that viewer switches down/up and other viewers remain stable;
- repeat through the legacy protocol path, which now requests automatic selection.

For the tracked opt-in profile, use the exact phases, metrics, acceptance criteria and bounded 2026-09-10 result in [`ADAPTIVE_QUALITY.md`](ADAPTIVE_QUALITY.md).

Desktop:

- Chrome/Chromium and Firefox;
- join/video/audio/control;
- refresh/rejoin;
- transient network reconnect;
- clipboard resynchronization after refocusing the controlling browser window;
- capture-pointer enabled/disabled behavior;
- Firefox keyboard input through the XInput path;
- H.264/VP8 regression plus H.265 only on clients that actually negotiate it.

File transfer and Open-in-App:

- test admin file download/upload/single deletion/bulk deletion;
- test every enabled non-admin download/upload/delete permission and denial when disabled;
- confirm path traversal and unauthenticated requests remain denied;
- if Open-in-App is enabled, confirm only an authorized host opens `http`/`https` links and other schemes/non-host attempts are rejected;
- confirm the feature stays hidden/inert when the plugin is disabled.

Mobile:

- Android Chrome and iOS/iPadOS Safari when available;
- autoplay/play/unmute;
- orientation/fullscreen;
- trackpad/touch;
- keyboard/helper;
- reconnect/recovery.

For the current iOS block, the generic mobile bullets are not sufficient; execute and record the same-peer, replacement-session, bounded-exhaustion and server-directed-disconnect phases in [`IOS_RECOVERY.md`](IOS_RECOVERY.md).

View-only sharing:

- configure a fresh ignored token and join an ordinary member, admin and view-only participant concurrently;
- verify the passive participant receives the same audio/video but cannot obtain control or use mouse, keyboard, touch, clipboard, files, microphone, admin, chat-send or arbitrary plugin input;
- exercise crafted HTTP, current/legacy WebSocket, data-channel and inbound-media attempts, not only hidden UI controls;
- rotate/remove the token and recreate the service to prove old-link and active-session revocation;
- execute the complete matrix and acceptance criteria in [`VIEW_ONLY_SHARING.md`](VIEW_ONLY_SHARING.md).

Smart-TV/embedded:

- join;
- WebRTC establishment;
- playback/audio;
- refresh/reconnect;
- diagnosable failure.

Do not claim device support without device evidence.

Browser/runtime images, when relevant to the deployment:

- build the selected architecture(s);
- verify Brave/Chromium/Chrome/Vivaldi startup and policy loading;
- on ARM64, verify Widevine installation and actual DRM playback;
- on NVIDIA deployments, verify encoder-element selection with the installed driver/GStreamer version.

## NEXT

Continue exclusively on `testing`; do not merge, fast-forward or push changes to `master`. The stable branch remains pinned at `d9105ef8` until the operator explicitly authorizes a later grouped promotion.

Validate the accumulated bounded iOS recovery and server-enforced view-only blocks together at one exact `testing` commit. Run the client checks, focused server tests/build and image build, then execute every phase and record every acceptance item in [`IOS_RECOVERY.md`](IOS_RECOVERY.md) and [`VIEW_ONLY_SHARING.md`](VIEW_ONLY_SHARING.md). Keep the existing ordinary/admin, touch, playback, file-transfer and adaptive-quality behavior in the regression scope. Do not claim no-reload iOS recovery or target-server view-only acceptance before that evidence exists.

After the grouped checkpoint is accepted, design the backend-neutral receive-media/subscription boundary for practical non-WebRTC prototypes. Implementation and any later promotion remain separate decisions; `master` must not move without explicit operator authorization.

## Product priority after stable synced baseline

1. validate the accumulated iOS/view-only client and server behavior at a coherent `testing` checkpoint;
2. design and prototype a practical non-WebRTC viewer-media fallback only after that acceptance;
3. promote accumulated `testing` history only after an explicit operator decision at a coherent validation milestone.

## Fallback prototype sequence

Alternative media work starts only after bounded iOS recovery is target-server validated and the server-enforced view-only boundary is established. It remains on `testing` until the operator explicitly approves a later grouped promotion.

When fallback work begins, separate the two user classes instead of forcing every client through one fallback chain:

1. establish a backend-neutral encoded-media/subscription boundary;
2. prototype **WebCodecs + dedicated WebSocket media** for interactive clients whose WebRTC/ICE path is unusable;
3. prototype **HLS / Low-Latency HLS** for passive/view-only clients such as Smart-TVs and constrained browsers;
4. compare device support, failure behavior, server resource cost, latency and recovery, then define explicit capability-based selection;
5. evaluate WebTransport only afterward if WebSocket's delivery/backpressure characteristics are a demonstrated limitation.

The passive path may trade latency for reliability and compatibility. It must stay in the same logical room and must not gain control authorization. HLS/LL-HLS is a TARGET candidate now, not merely a generic later idea.

## LATER / OPTIONAL

- WebTransport productionization;
- fully automatic transport selection after explicit/manual fallback paths are proven;
- automatic codec selection;
- MPEG-DASH as an additional passive HTTP-streaming backend where useful;
- MJPEG only as an ultra-legacy image-only fallback for a concrete unsupported device class;
- broader v3 client/library rewrite.

## OPEN

- Supported Smart-TV/device matrix, including native HLS, MSE/DASH and WebCodecs capability.
- Whether the target iPhone validates the implemented same-peer and replacement-session paths without reload; a Safari Play gesture remains an explicitly separate, permitted policy fallback.
- Exact WebCodecs/WebSocket framing, queue/drop/backpressure policy and codec set.
- HLS/LL-HLS latency target, segment/part sizing, codec profile and server resource cost.
- Whether DASH adds meaningful compatibility beyond HLS for the actual target devices.
- Exact per-client media-backend capability/selection rules and rollout order.
- Longer-term legacy Vue 2 migration.
