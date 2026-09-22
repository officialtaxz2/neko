# Work Plan / Handoff State

Last consolidated: 2026-09-14.

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
testing: deployment reconciliation, accepted opt-in adaptive quality, bounded iOS recovery, server-enforced view-only sharing, implemented media-subscription/WebRTC compatibility, WebCodecs/media-WebSocket receive path and separate Phase 4 deployment/observability assets
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

Status: **accepted on 2026-09-10 as an opt-in profile for the documented target deployment and three-viewer scenario**. The stable base Compose remains single-pipeline, and profiles without `nominal_bitrate` retain the measured-video fallback. The later correction below deliberately changes downgrade/reference semantics and therefore requires a focused rerun. This bounded acceptance must not be generalized to untested devices, architectures or network conditions.

## COMPLETED IN REPOSITORY — revised steady-state adaptive downgrade correction

Operator-supplied target evidence on 2026-09-19 established a new issue independently of the earlier startup-window bug and encoder tuning:

- one healthy active WebRTC session used direct UDP and recorded zero receiver loss, zero NACKs and zero local video drops;
- automatic selection nevertheless changed `high -> medium -> high -> medium`;
- reload recreated the peer/estimator and only temporarily restored `high`;
- `max-quantizer` can explain softer motion inside a tier but is not the cause of the recorded tier switches.

The first implementation at `2d037f39` passed focused target-server tests/build, deployment and an initial 180-second healthy-client check, but the longer unshaped trace on 2026-09-22 rejected it:

- with no host shaper, direct UDP and zero receiver loss, zero NACKs and zero local video drops, the same peer changed `high -> medium` at `12:11:00Z` and `medium -> low` at `12:11:34Z`;
- it remained degraded for roughly eight minutes, then recovered `low -> medium -> high` around `12:19Z` without a reload;
- the target estimate fell to roughly 0.47 Mbit/s even though the healthy path and qdisc showed no packet loss;
- the first candidate had separated current-tier deficit from upgrade reserve, but still allowed the GCC target/trend alone to prove congestion.

Revised and statically reviewed on `testing`:

- Replaced the old downgrade interpretation of `target / measured_video <= 1 + diff_threshold`. A neutral/application-limited estimate no longer needs upgrade-style spare capacity merely to retain its current tier.
- Current-tier fit now uses `max(measured video, nominal video) + measured active audio`, plus the explicit `transport_reserve` (0.05 in the opt-in profile). `diff_threshold: 0.15` is the tolerated sustained deficit below that complete-delivery reference.
- Downward trend and neutral stall remain signals, but either can downgrade only while the material deficit and fresh peer-local receiver congestion evidence persist through the existing windows. Evidence means a non-zero latest RTCP `FractionLost` or a received NACK no older than two estimator read intervals; clean, missing and stale feedback is non-actionable.
- Insufficient samples and receiver congestion reset the stable-upgrade window. Upgrade remains separately gated by a clean stable interval and `upgrade_diff_threshold` over the next tier's nominal complete-delivery reference, including audio and transport reserve. The gap between current-tier downgrade floor and next-tier upgrade requirement is intentional hysteresis.
- The 12-second stable, 6-second unstable, 8-second stalled, 30-second downgrade-backoff and 5-second upgrade-backoff profile values are unchanged. Estimator state remains per peer, so no selection signal is shared between viewers.
- Added focused tests for a healthy neutral estimate without 15% spare, the reproduced severe loss-free target collapse, fresh/stale loss and NACK evidence, transient evidence, peer isolation, sustained confirmed shortage, stable recovery, the hysteresis deadband, measured/nominal/audio/transport reference construction and exact startup/backoff boundaries.
- Added `neko_webrtc_receiver_report_fraction_lost` and `neko_webrtc_receiver_congestion_evidence`, included receiver loss/NACK series in the evidence collector and enriched estimator logs with feedback freshness and policy classification.
- WebRTC remains the default. No WebCodecs/HLS path, client selection or encoder quantizer/bitrate was changed.

Static status: **revised implementation and review complete in repository / revised builds, tests, containers and live media NOT EXECUTED IN CODEX**. The first exact-commit candidate is explicitly rejected; the revised compact healthy-plus-constrained target-server gate in [`ADAPTIVE_QUALITY.md`](ADAPTIVE_QUALITY.md) is pending.

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

Target-server status: **closed by explicit operator decision on 2026-09-11 with a known manual-device evidence gap**. At exact commit `913a981e`, the dependency-free recovery tests passed with the complete containerized client checks, and the focused server tests/build plus image build/start also passed. No Mac/Safari Web Inspector was available; the operator deliberately omitted phases A–C and accepted the automated evidence for the deployment decision. The prior reload/Play evidence does not validate the new no-reload implementation, and no automatic iOS claim is made.

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

Target-server status: **accepted for the tested deployment**. The security/role matrix through `80eeca64` and the compact-link follow-up at exact commit `913a981e` passed. The separately omitted iPhone deep test remains an explicit limitation, not a failed View-only check.

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

After the successful original matrix, the operator requested the intentionally incompatible compact `#/<16-character-Base64URL-token>` form. Its fresh exact-commit checks, deployment, HTTP probe and new-link browser smoke test are complete.

## CLOSED WITH EXPLICIT LIMITATION — grouped iOS/View-only checkpoint

Closed by operator decision on 2026-09-11:

- View-only is accepted for the tested deployment based on the adversarial role/HTTP/WebSocket/data-channel/inbound-media/revocation evidence through `80eeca64` and the exact compact-format validation at `913a981e`.
- The exact-commit client recovery tests, TypeScript lint, production build, focused Go suites, server/plugin build, local image build and healthy deployment all passed on the target server.
- No Mac/Safari Web Inspector was available. The operator deliberately declined the manual iPhone same-peer, replacement-session, bounded-exhaustion and server-directed-disconnect phases and accepted the automated evidence as sufficient to continue repository work.
- This is an explicit risk acceptance, not fabricated device evidence. Automatic no-reload recovery on a real iPhone remains unverified and the reusable procedure stays in [`IOS_RECOVERY.md`](IOS_RECOVERY.md).
- `master` remains pinned at `d9105ef8`; closing this checkpoint does not authorize promotion.

## COMPLETED — backend-neutral media-subscription boundary design

Designed and statically reviewed on `testing` on 2026-09-11. The complete contract is in [`MEDIA_SUBSCRIPTION_BOUNDARY.md`](MEDIA_SUBSCRIPTION_BOUNDARY.md).

Decisions now fixed:

- shared encoded capture is exposed through a transport-independent source catalog and bounded source subscriptions;
- source subscriptions are separate from participant deliveries, allowing per-viewer WebRTC/WebSocket delivery and shared-per-variant HLS packaging without duplicating authorization or capture;
- one delivery manager validates `CanWatch`, creates scoped leases, owns revocation and updates backend-neutral watching state;
- media codecs/descriptors do not embed Pion types, and format, GStreamer PTS/DTS, timeline generation, keyframes and discontinuities are first-class contract data;
- all subscription queues are bounded/non-blocking and backend overflow must drop/resynchronize/close locally instead of blocking capture;
- WebRTC remains the default and is migrated first with unchanged signaling, data-channel, estimator, queue/drop, metrics and configuration behavior;
- no WebCodecs/WebSocket endpoint, HLS route/packager, automatic fallback or WebTransport implementation belongs to the compatibility-refactor block.

Design-checkpoint status at `04787b55`: **design complete / no alternative media backend implemented / no project code, test, build or runtime check executed in Codex**. The following repository block implements that design without adding an alternative backend.

## COMPLETED IN REPOSITORY — media-subscription/WebRTC compatibility refactor

Implemented and statically reviewed on `testing` on 2026-09-11:

- Added Pion-free encoded-media codecs, sources, units, format/discontinuity/end events, selectors, source-subscription contracts, backend capabilities, participant delivery requests and credential-free leases in `server/pkg/types/media.go`.
- Extended the GStreamer appsink bridge with buffer PTS/DTS validity, duration and caps-derived resolution/frame rate while retaining the existing `types.Sample` capture input.
- Added capture pipeline generation and per-generation sequence metadata. A capture-backed provider now exposes ordered sources, starts/stops them on first/last active subscription, gates video on keyframes, normalizes valid GStreamer timestamps to one provider-owned timeline and publishes immutable encoded data.
- Added bounded manager-owned subscription queues with local non-blocking drop-new overflow, explicit lifecycle transitions and pause/resume/switch/idempotent-close behavior. The WebRTC sender consumes this queue directly; no second queue was stacked in front of it.
- Added a central participant-delivery manager/registry. It validates the current authenticated session and `CanWatch`, intersects receive requests with backend capabilities, creates a backend/session-scoped lease without login/share credentials, keeps one primary delivery per session and owns replacement, revocation, generic watching state and shutdown ordering.
- Migrated WebRTC into the first registered backend without adding a route, endpoint, client protocol or configuration. Existing SDP/ICE signaling, data channels, inbound-media enforcement, audio/video messages, private mode and estimator-driven selection retain their current paths.
- Preserved the effective two-sample WebRTC queue with drop-new behavior and the existing `neko_webrtc_track_dropped_samples_total` meaning. Added low-cardinality `neko_media_*` delivery, subscription, queue, delivered-unit/byte, drop, discontinuity and source-generation metrics; no metric label or lease field carries a credential.
- Added focused tests for source ordering/selection, first/last demand, keyframe admission, switching, pause/resume, idempotent close, timing/generation/format/discontinuity ordering, local overflow isolation, the WebRTC two-unit policy, `CanWatch` denial, view-only receive allowance, replacement, profile/session revocation and shutdown cleanup.
- Expanded `docker-compose.validation.yaml` so the target-server server check includes `./internal/capture` and `./internal/media` and its safe metrics snapshot includes the new `neko_media_*` series.
- Added no WebCodecs/WebSocket endpoint, HLS route/packager, WebTransport or automatic backend selection.

Static review status: **implementation complete in repository / project code, tests, build, Docker image and runtime checks NOT EXECUTED IN CODEX**.

The compatibility implementation subsequently received the bounded target-server checkpoint recorded below. Full exact-commit end-to-end repetition remains deliberately deferred and is not implied by that closure.

The first target-server validation-image build at `637d259f` reached the Go compile step and exposed two stale imports removed from `server/pkg/types/capture.go` even though later capture-pipeline helpers still use `fmt` and `strings`. The follow-up restores those imports only; the complete container check and runtime matrix must be rerun at the resulting exact commit.

At follow-up commit `5d68ef0b`, the validation image and its embedded server/plugin build succeeded. The expanded container test command then passed `pkg/types`, auth, capture, member, session, legacy HTTP, WebSocket and WebRTC, but `internal/media` stopped at compile time because one test retained an unused fake-backend binding. The next follow-up removes only that binding; the complete container command still requires a clean rerun because its trailing `./build` did not execute after the test-command failure.

At exact follow-up commit `a32027d`, the complete validation service passed `pkg/types`, auth, capture, media, member, session, legacy HTTP, WebSocket and WebRTC followed by `./build`. The local base image `sha256:b83d6eabe71885d54d52bf087240c14862bcec44e4b3394e375f03d7ef9f0c61` and Brave image `sha256:cc4d727a93e27d79c900adcb077252512c4cfc8f9b42f33685ddf7fd010b3061` built successfully; the adaptive service then started healthy with zero restarts and the expected profile/download/policy mounts. A short-lived current `/api/ws` probe passed delivery open/close, video disable/enable, exact low/medium/high selection, auto selection and audio disable/enable. The new metrics showed three active WebRTC deliveries, matching open/close arithmetic, active audio/video subscriptions, capacity-two subscription queues, delivered units/bytes and expected discontinuities without credential or delivery-URL labels. Peer-local audio/video drop counters remained zero.

The normal fork client uses the legacy `/ws` bridge, so sending current-only `signal/video` or `signal/audio` messages through its `$client` object is not a valid manual test; the server correctly logged those attempts as unknown legacy events. Real browser windows still exercised normal legacy receive behavior, while the isolated current-protocol probe exercised the new lifecycle controls.

Fresh same-host browser windows on the candidate selected `high` and then switched immediately to `medium` on their first logged neutral estimator reading. A controlled A/B with the accepted pre-refactor Brave rollback image `sha256:5b5c7747930bc863da05cb3f05d8bb7b97a624386d105bdade2448e5443d36a2` reproduced the same immediate behavior. This rules out the media-subscription refactor as its introduction but exposed a separate inherited estimator startup defect described below.

## COMPLETED IN REPOSITORY — estimator startup observation-window correction

Static review of the estimator path found the mechanism matching the immediate-start trace:

- `stableSince` began at estimator-reader startup, but `unstableSince` and `stalledSince` began as Go zero times;
- `time.Since` on those zero values makes the configured observation durations look already expired, so the first qualifying neutral estimate could enter the downgrade path without either intended grace period;
- the trend detector begins `NEUTRAL` and requires eight non-collapsed values before it can report a trend, making the stalled path particularly relevant during startup;
- the correction initializes all three observation windows from the same reader-start timestamp and deliberately leaves only the switch-backoff timestamps zero until an actual upgrade/downgrade;
- a focused WebRTC regression test fixes that initialization contract without changing bitrate targets, thresholds, durations, public APIs, protocols or deployment defaults.

Static review status: **implementation and diff review complete / project tests, build, Docker image and runtime checks NOT EXECUTED IN CODEX**.

The focused target-server correction check passed at exact commit `e5f55bf9` as recorded below. Repository inspection and that bounded trace prove the zero-time behavior is corrected and did not cause an immediate downgrade in the observed run; they do not prove that this defect was the sole cause of every sustained `medium` selection or that a constrained receiver must stay on `high`.

## CLOSED WITH EXPLICIT LIMITATION — media-subscription/estimator checkpoint

The operator closed this combined checkpoint on 2026-09-12 at exact source commit `e5f55bf9` on `testing` with the following bounded evidence:

- the complete validation service passed `pkg/types`, auth, capture, media, member, session, legacy HTTP, WebSocket and WebRTC, including the estimator-startup regression test, followed by `./build`;
- fresh local base and Brave images built successfully, the adaptive Compose service started healthy with zero restarts, and the deployed image was `sha256:4860a706476b111d1c010587058b2a562f29ed696e503b48c4fbbc55a5208349`;
- one fresh viewer selected `high`; its first estimate arrived about two seconds later with `NEUTRAL` trend, and no downgrade, upgrade or stall was logged during the roughly 54-second observation;
- that viewer remained the sole `high` listener, the high pipeline measured about 1.89 Mbit/s, the receiver estimate rose to its 50 Mbit/s cap, and its video queue-drop counter remained zero;
- CPU was about 173% of one core and memory about 953 MiB on the reported eight-CPU/25.43-GiB host, with no indication that the host was saturated in this focused run;
- the earlier `a32027d` current-protocol lifecycle probe, credential-free `neko_media_*` lifecycle/queue metrics, healthy deployment and candidate/rollback A/B remain supporting evidence for the compatibility refactor.

Explicit limitation: the operator chose not to repeat the ordinary/admin/view-only/private-mode/manual-tier/reconnect matrix at `e5f55bf9`, because those boundaries had already been exercised in earlier checkpoints and the follow-up changed only estimator timestamp initialization. A fresh independently constrained three-viewer `high -> medium -> low -> medium -> high` isolation run was also not performed. Both are deferred to final grouped validation. Therefore this closure is sufficient to continue prototype work, but it is not a claim that the omitted matrix passed at `e5f55bf9` and not a universal network/device guarantee.

The reported impression of softer video during fast scrolling or high-motion playback was also reviewed statically. `deploy/adaptive-quality.yaml` and `docker-compose.adaptive.yaml` are unchanged from the accepted `bfaca84e` profile; the subscription refactor adds media metadata/lifecycle but does not alter encoder construction or configuration, and the GStreamer-to-provider-to-Pion path forwards the same encoded payload bytes. Together with zero video queue drops in the focused run, there is no evidence that the new boundary introduced an image-quality regression. The existing `high` tier is still fixed-rate VP8 at 1,996,800 bit/s, 25 fps and `max-quantizer: 63`, so complex motion can be quantized more heavily than a static desktop and appear temporarily softer. This remains a subjective, non-blocking observation until a controlled same-content bitrate/QP A/B records receiver statistics and comparable captures; no accepted profile value is changed in this checkpoint.

## COMPLETED — WebCodecs plus dedicated media-WebSocket contract

Status: **design complete on `testing` on 2026-09-12 / Phases 1–3 subsequently implemented in the repository / deployment enablement, grouped acceptance and automatic fallback not implemented**.

[`WEBCODECS_MEDIA_WEBSOCKET.md`](WEBCODECS_MEDIA_WEBSOCKET.md) now fixes the first default-off `webcodecs-ws` receive prototype:

- an authenticated event-plane negotiation creates a 10-second, 24-byte, single-use ticket without exposing a login/share credential to the backend or a URL;
- the upgrade path rechecks exact Origin, ticket binding, the current live session and `CanWatch` before opening a credential-free delivery lease;
- logout, revocation, private mode, replacement, reconnect and shutdown follow the existing central media-manager lifecycle;
- `neko.media.v1` uses a strict 64-byte binary header for VP8 video and raw Opus audio, including format, per-track delivery generation, sequence, PTS/DTS validity, duration, keyframe and discontinuity semantics;
- browser and server codec support are both checked explicitly; another configured codec is rejected rather than transcoded or silently substituted;
- provider, egress, compressed, decoder and render/audio queues have concrete record/byte/time caps; overflow is local and recovers through a new generation and video keyframe;
- audio is the common presentation clock, stale generations are rejected, and decoder/timing failures use a bounded common resync;
- server and client opt-ins are both required, retry stays on the selected backend, and returning to unchanged WebRTC requires an explicit action;
- origin, size, rate, timeout, logging and metric limits are fixed, with credential-bearing subprotocol/event data excluded from logs;
- target-server gates cover default invariance, authorization, startup, latency, A/V skew, slow-client isolation, reconnect, resources and malformed input;
- implementation is split into protocol/ticket, server adapter, isolated client path and deployment/validation phases.

The design also records a material product boundary: the current client sends high-rate mouse, keyboard and touch input over the WebRTC data channel. Version 1 is therefore an interactive-class receive prototype, not complete non-WebRTC control parity. It adds no replacement input transport. HLS/LL-HLS stays a separate later passive/view-only block.

This repository work was statically reviewed only. Project code, tests, builds, containers, browsers and media paths were **NOT EXECUTED IN CODEX**.

## COMPLETED — WebCodecs/media-WebSocket Phase 1 protocol and ticket boundary

Status: **implemented and statically reviewed on `testing` on 2026-09-12 / target-server tests and build pending / no media endpoint, delivery backend or client path implemented**.

The reviewed Phase 1 block adds only the prerequisites needed before a socket can safely allocate media resources:

- `EncodedMediaUnit.PTSValid` now preserves whether capture supplied PTS or the provider synthesized its normalized fallback; the Pion compatibility path remains otherwise unchanged.
- `server/internal/mediaws` owns the strict 64-byte `neko.media.v1` record codec, bounds and schema validation. Language-neutral byte-exact fixtures cover video and audio FORMAT/UNIT plus DISCONTINUITY and END, with mutation, boundary and fuzz tests prepared for the target server.
- The ticket store retains only SHA-256 digests, issues 24-byte/32-character Base64URL tickets for ten seconds, atomically returns their server-owned session/backend/version/source binding exactly once, replaces an earlier pending ticket and keeps a bounded replay tombstone.
- Current and legacy authenticated event paths now translate `media/capabilities/request`, `media/capabilities`, `media/create` and `media/offer`. View-only sessions may use the two receive-negotiation requests, while all four payloads are excluded from WebSocket payload logging.
- The negotiator requires the exact current connected session object and `CanWatch`, validates only server-owned VP8/raw-Opus sources and exact choices, rate-limits ticket creation and invalidates pending tickets on deletion, final disconnect or watch revocation.
- `media.webcodecs_ws.enabled` defaults to false. When true it registers only the negotiation handler. There is deliberately no `/api/media/ws` route, backend registration, provider subscription, client selection or deployment overlay in Phase 1.

The supplied preliminary implementation was not imported mechanically. Static review corrected five material boundary issues before adoption: ticket redemption now returns the stored binding so a future credential-free route does not need client-supplied binding data; stale session objects are rejected; encoded lengths are rejected before payload copies; version 1 rejects codec-config bytes that VP8/raw Opus do not define; and required zero-value/nullable JSON fields can no longer disappear silently during decoding or legacy translation. Audio fixtures and broader boundary tests were also added. The contract now explicitly handles cold video sources whose dimensions become known only after a bounded Phase 2 subscription produces a complete FORMAT; a preliminary zero-dimension provider format must not cross the wire.

Per `AGENTS.md`, these project checks were **NOT EXECUTED IN CODEX**. Run them on the real target server at the exact Phase 1 commit before using the result as build/test evidence:

```bash
cd server
go test ./pkg/types ./internal/capture ./internal/mediaws ./internal/http/legacy ./internal/websocket
go test ./internal/mediaws -run '^$' -fuzz '^FuzzParseRecord$' -fuzztime 30s
./build
```

No runtime behavior can exercise an alternative media transport at this phase because none is registered.

## COMPLETED — WebCodecs/media-WebSocket Phase 2 server delivery adapter

Status: **implemented and statically reviewed on `testing` on 2026-09-13 / target-server tests and build intentionally deferred to later grouped prototype validation / at this Phase 2 closure no client decode-render path or deployment overlay was implemented**.

The Phase 2 block completes the server-side delivery and security boundary without changing the stable default:

- `webcodecs-ws` is a receive-only `MediaBackend` over the generic encoded provider. It receives only the credential-free lease, normalized exact VP8/Opus source selectors, an initial private-mode state and the upgraded socket attachment; no login/share token enters the backend request.
- `media.webcodecs_ws.enabled=false` remains the default. Only when true does startup share the Phase 1 ticket store across negotiation and attachment, register the backend and expose `GET /api/media/ws`. At the Phase 2 closure, ordinary clients still issued no prototype events and no client/deployment opt-in existed.
- The pre-upgrade controller rejects query parameters, insecure production transport, non-exact Origin, missing/reordered/extra subprotocols and malformed/unknown/expired/replayed tickets before provider allocation. Ticket redemption is atomic, the current connected session and `CanWatch` are re-resolved, concurrent attach verification is capped at 64, invalid attempts at 20/minute per safely resolved address, and sockets at 128 or a lower configured maximum.
- Forwarded address/HTTPS headers are trusted only when the captured original socket peer belongs to configured proxy IP/CIDR ranges. Cleartext direct loopback requires the separate `media.webcodecs_ws.allow_insecure_loopback=true` development option and refuses forwarded headers. The media route's logger emits only its fixed route label, never the full attempted URL; subprotocol/ticket and authorization headers are never logged.
- Per-delivery provider queues are video 4/audio 16 with non-blocking drop-new behavior. A single writer owns the uncompressed socket and gives the four-record lifecycle queue priority over the shared 24-record/16-MiB media queue, with 2-second writes, 10-second ping, 20-second pong, one-second close and a 16-MiB/s five-second outbound cap.
- Strict 4-KiB `ready`, `feedback`, `resync` and `stop` records enforce exact fields, canonical safe integers, queue/time/skew ranges, 10/s burst-20 controls, one feedback/second and five client resyncs/minute. The lease remains opening until all requested complete FORMATs match READY within five seconds.
- Delivery-local generations and sequences, provider PTS/DTS validity, bounded video/Opus duration derivation and VP8 keyframe markers are serialized into the Phase 1 envelope. Provider/egress overflow and provider transitions gate units before lifecycle insertion; video resync clears video, common audio failure clears both, DISCONTINUITY precedes FORMAT, and video reopens only on a keyframe. Three resyncs within 30 seconds close rather than loop.
- Progress tracking records only a bounded sent-unit timestamp history. Client counts cannot claim unsent progress; no feedback for five seconds, no rendered progress for three plus a two-second recovery, or a reconstructable rendered PTS more than 500 ms behind initiates the specified bounded recovery/close behavior.
- Generic delivery close reasons now preserve revoked, replaced and shutdown semantics. Initial private mode is applied inside backend construction before workers can publish; later private mode pauses subscriptions and resume starts a common new generation. Final session disconnect, profile revocation, deletion, explicit stop/socket loss, primary replacement and shutdown release subscriptions, queues, workers, manager state and the controller connection slot.
- Fixed low-cardinality `neko_media_websocket_*` connection, handshake, record, byte, drop, resync, write-time, queue-depth/bytes and client-lag metrics contain no ticket, session, IP, Origin, user-agent, URL or source label.

Focused tests were added for strict control schemas and limits, queue caps/priority/terminal behavior, FORMAT/UNIT duration/validity/keyframe serialization, pre-READY and common transition gating, rendered-PTS lag, END/close reasons, exact pre-upgrade policy/statuses, trusted proxy/loopback behavior, attach/rate caps and central private/revoked/replaced/shutdown lifecycle. The validation Compose command now includes the directly changed config and HTTP packages.

Per `AGENTS.md`, project code, tests and builds were **NOT EXECUTED IN CODEX**. At the operator's direction, Phase 1 and Phase 2 are retained for the later grouped prototype checkpoint. At that exact `testing` commit run:

```bash
cd server
go test ./pkg/types ./pkg/auth ./internal/config ./internal/capture ./internal/media ./internal/mediaws ./internal/member/multiuser ./internal/session ./internal/http ./internal/http/legacy ./internal/websocket ./internal/webrtc
go test ./internal/mediaws -run '^$' -fuzz '^FuzzParseRecord$' -fuzztime 30s
./build
```

This Phase 2 repository completion is not target-server evidence and did not by itself make the alternative path usable; Phase 3 has since supplied the explicit client while retaining the same deferred-validation status.

## COMPLETED — WebCodecs/media-WebSocket Phase 3 isolated client path

Status: **implemented and statically reviewed on `testing` on 2026-09-14 / accumulated client and server tests, builds and target-browser/media acceptance intentionally deferred to Phase 4 grouped validation**.

The Phase 3 block completes the repository receive path without changing the ordinary WebRTC default:

- Exactly one client `media=webcodecs-ws` query selects the prototype. The selected legacy event socket suppresses only its automatic WebRTC `signal/request` and receives its current authenticated session ID in `system/init`; the ordinary signal request and init JSON remain unchanged when selection is absent, wrong or duplicated.
- `client/src/neko/media/protocol.ts` is the strict browser parser for the shared language-neutral golden fixture. It rejects invalid magic/version/type/kind/track/flags/reserved fields, unsafe 64-bit values, inexact lengths, duplicate/unknown metadata fields, schema/value bounds and inconsistent VP8 keyframe markers before decoder admission.
- A dedicated module worker owns generic and exact codec probes, the ticket-bearing media socket, the exact selected subprotocol, FORMAT timeout, VP8/Opus decoders, stale generation/sequence/timestamp rejection and browser feedback. Its compressed/decode queues are capped at video 4/audio 16, decoded video at two outstanding frames and decoded PCM at 200 ms.
- Recovery uses decoder reset plus a new generation and keyframe rather than `flush()`. Common server resync ordering is handled without returning a stale video generation; a video-only server generation clears stale canvas output while retaining the normalized audio master, whereas an audio/common discontinuity clears both presentation paths.
- The controller creates a credential-free media URL, never logs the ticket, requests exact advertised sources and a new one-time ticket per attempt, and schedules the 80-ms audio-master or monotonic video-master clock. A bounded stereo AudioWorklet owns PCM and the canvas renderer closes every frame, drops frames over 80 ms late and holds an early frame for no more than 100 ms.
- Central Play, mute/volume, resize, fullscreen and surface geometry are integrated without changing the ordinary video element. The opt-in UI labels receive-only/no-Picture-in-Picture/no-automatic-fallback limitations, hides input/control/microphone actions and exposes terminal **Retry WebCodecs** and **Use WebRTC** actions.
- Only media close codes 4413 and 4500 schedule the four serialized 1/2/5/10-second same-backend attempts. Every attempt requests a new ticket; logout, event-session loss, policy/protocol/codec failure and explicit stop do not auto-retry or create WebRTC.
- Focused tests are prepared for shared-fixture parsing in the exact browser parser source, malformed records, media retry bounds, exact server query selection and ordinary-versus-selected legacy init serialization; the client performs the same one-value exact comparison directly at initialization. The shared video-key fixture was corrected from inconsistent marker bytes to a minimal marker-valid VP8 key prefix so both server and browser validations describe the same record.

Per `AGENTS.md`, project code, tests, linters and builds were **NOT EXECUTED IN CODEX**. Phase 4 must run the complete exact-commit client sequence plus the accumulated Phase 1–3 server suite/fuzz/build before any runtime claim:

```bash
cd client
npm ci
npm test
npm run lint
npm run build

cd ../server
go test ./pkg/types ./pkg/auth ./internal/config ./internal/capture ./internal/media ./internal/mediaws ./internal/member/multiuser ./internal/session ./internal/http ./internal/http/legacy ./internal/websocket ./internal/webrtc
go test ./internal/mediaws -run '^$' -fuzz '^FuzzParseRecord$' -fuzztime 30s
./build
```

Repository inspection alone is not deployment or browser evidence. The server flag remains false by default and Phase 4 supplies a separate explicit overlay. The operator-provided target-server evidence below supports only the recorded bounded checkpoint; full WebCodecs/media-WebSocket acceptance still requires the deliberately deferred matrix.

## BOUNDED CHECKPOINT CLOSED — WebCodecs/media-WebSocket Phase 4 deployment, observability and grouped validation

Status: **repository assets implemented / bounded target-server checkpoint operator-closed on 2026-09-14 at exact documentation commit `6b6cd328` / exact automated, build, deployment/security and corrected foreground-iPhone evidence passed / numeric latency/pacing, induced slow-client/adaptive isolation, resource comparison and remaining live hostile-input cases explicitly deferred / no full prototype-acceptance claim**.

- Added `docker-compose.webcodecs-ws.yaml` as the only deployment enablement. It sets the existing feature flag and requires an exact public Origin plus an explicitly reviewed immediate trusted-proxy address/CIDR; cleartext loopback remains false and the server's 128-connection default is preserved. `docker-compose.yaml` is unchanged, so omitting the overlay keeps the route absent.
- Added inactive example values to `.env.example`; the overlay itself contains no credentials and refuses activation when its two security-boundary values are empty.
- Extended `docker-compose.validation.yaml` with the accumulated 30-second media-record fuzz target and the fixed WebCodecs, central delivery, capture and Go/process metric families.
- Added `deploy/check-media-websocket-http.sh`, which uses no login credential and only a fixed invalid ticket to verify the disabled route plus query/origin/subprotocol/ticket pre-upgrade statuses without printing response bodies.
- Added `deploy/collect-webcodecs-media.sh` and [`WEBCODECS_MEDIA_WEBSOCKET_OBSERVABILITY.md`](WEBCODECS_MEDIA_WEBSOCKET_OBSERVABILITY.md) for private exact-commit/container/resource records, fixed-label metrics, PromQL and filtered credential-safe logs. Compose consumes ignored `.env` for normal interpolation, but the collector never prints or archives it, a ticket-bearing URL/header or event/control payload.
- Added [`WEBCODECS_MEDIA_WEBSOCKET_RESULTS_TEMPLATE.md`](WEBCODECS_MEDIA_WEBSOCKET_RESULTS_TEMPLATE.md) and [`WEBCODECS_MEDIA_WEBSOCKET_VALIDATION.md`](WEBCODECS_MEDIA_WEBSOCKET_VALIDATION.md) for blockwise operator/Codex execution, the ten-join timing table, role/private-mode/revocation cases, slow-viewer/adaptive isolation, five-minute resource comparison, hostile inputs and state-free rollback.
- Static inspection found no change to the base Compose, WebRTC signaling/data channels, authorization, REST/event protocol, capture encoder configuration or client backend selection. The overlay does not add a client query, fallback or persistent state.

Per `AGENTS.md`, project code, tests, linters, builds, Docker and the helper scripts were **NOT EXECUTED IN CODEX**. All runtime evidence below was supplied by the operator from the target server. The bounded closure does not convert any omitted acceptance item into a pass.

The first target-server client-validation attempt at exact Phase 4 asset commit `b91c6fdd` reached `npm ci` and passed the eight recovery/negotiation/share/media-recovery tests, then stopped before the three media-protocol cases, lint and build because the isolated validation work directory contained only `client/` while that test intentionally reads the shared server golden fixture at `server/internal/mediaws/testdata/neko_media_v1_golden.json`. The repository fixture and test path were correct; the validation container packaging was incomplete. The follow-up copies only that fixture into the matching isolated `/work/server/...` path. The complete client sequence requires a clean rerun at the follow-up commit and no client acceptance is inferred from the partial attempt.

At exact follow-up `33a0ac64`, the complete client sequence passed all 11 tests, TypeScript `tsc --noEmit` and the Vite production build. The server-validation image and its embedded server/plugin build then succeeded, and every listed package except `internal/mediaws` passed. That package exposed an error in `TestControllerAttachAndInvalidAttemptBounds`: after releasing the replaced active connection, the test still expected a second session reservation to exceed a two-connection limit even though only the replacement reservation remained. The production counter correctly allowed that second slot. The follow-up test now admits `other`, verifies a third reservation is rejected with HTTP 429, and releases both reservations explicitly. The media-WebSocket package suite, 30-second fuzz target and trailing validation-service build require a clean rerun; no server acceptance is inferred from the partial attempt.

At exact follow-up `b6a0a5db`, the complete server package sequence, 30-second media-record fuzz target and trailing server/plugin build passed on the target server. The exact local base and Brave images built successfully. The default-off deployment started healthy with zero restarts, the dedicated route returned 404, and the operator reported that ordinary WebRTC remained functional. With the separate overlay enabled, the service again started healthy with zero restarts and the credential-free live HTTP boundary returned every expected 400/401/403/426 status through the public Caddy origin. The client sequence was not repeated at `b6a0a5db`; its most recent complete pass remains `33a0ac64`, whose only later change before `b6a0a5db` was the Go test correction described above.

The first supported-browser media run at `b6a0a5db` did not pass. Three successful attachments delivered 1,166,277 video bytes, 441 audio bytes, 65 VP8 UNITs and 147 Opus UNITs, then each delivery failed after two accepted `client_queue_overflow` resyncs and the bounded third-resync rejection. Capture remained healthy and every subscription/delivery/socket cleaned up. Static tracing identified that `VideoDecoder.decodeQueueSize` may decrease before its output callback runs. The first follow-up at `c4a316d5` reserved render capacity across that gap, kept the socket close handler attached after `END backend_error`, and added the bounded close log. Its exact target-server client sequence passed all 13 tests, TypeScript and the production build; the complete server suite, 30-second fuzz target with 1,348,987 executions and trailing build passed; fresh base/Brave images built; and the enabled deployment started healthy with zero restarts and passed every public/direct security probe.

The `c4a316d5` browser rerun still did not pass: it remained black and cycled through streaming/reconnecting. The fixed close log recorded 22 failed deliveries, all code 4413/reason `resync_limit`; metrics recorded 23 successful attaches, 44 accepted `client_queue_overflow` resyncs, 1,060 Opus and 462 VP8 units, while the service remained healthy and the currently active delivery retained both subscriptions. This also exposed that resetting the retry counter immediately at READY allowed an unbounded reconnect loop despite the four-attempt contract. Revised static tracing found the deeper video flow-control issue: reserving the two renderer slots still stops decode progress while `requestAnimationFrame` holds those frames, allowing the four-unit compressed queue to fill. The `b68f1fbf` follow-up therefore decodes continuously within the decode-queue cap, closes/counts decoded video beyond the two-frame hand-off locally without losing VP8 reference continuity, differentiates bounded audio/video overflow causes in control metrics/logs, and resets the four-attempt recovery budget only after 30 uninterrupted seconds.

At exact `b68f1fbf`, all 13 client tests, TypeScript, the Vite production build, the complete server package sequence, a 30-second fuzz run with 1,435,309 executions and the trailing server build passed. Fresh base and Brave images built; the exact Brave image deployed healthy with zero restarts and passed every public/direct pre-upgrade probe. Its first single-browser run retained one active delivery without reconnecting and delivered 1,304 Opus plus 669 VP8 units/5,568,194 video bytes; only one bounded `client_audio_underflow` resync occurred. The canvas nevertheless remained black. Live browser instrumentation then counted four valid 1280-wide frames with correct 80–112 ms scheduling while `playing`/`playable` were true. More than a minute later, two frames remained queued, the canvas backing store was still its untouched 300×150 default, the real component generation stayed zero and animation handle 6 did not advance even after a manual start. Inspection of vue-class-component 7's constructor data collection explains all observations: class-field arrow callbacks lexically retain its synthetic data instance, so primitive generation/animation updates diverge from the live Vue component.

The `f7a9de4e` correction converted the frame, reset and animation callbacks to Vue-bound prototype methods, released any frame with no EventEmitter listener and added focused source/runtime regression tests. All 15 client tests, TypeScript and Vite build passed; fresh images built, the exact image deployed healthy with zero restarts and every public/direct security probe passed. The operator confirmed visible 1280×720 video, audible approximately synchronized audio and continuous `streaming`; the first snapshot retained one active connection after 2,692 VP8/5,365 Opus units. Short roughly one-second black flashes remained. A later snapshot counted six `client_audio_underflow` resyncs, exactly six audio and six video discontinuities, and 8,479 VP8/17,069 Opus units with no service restart, matching those flashes: one empty 128-frame worklet quantum immediately caused a common reset.

The `f212dafd` follow-up required 100 ms of continuous starvation and added a processor-level transient/sustained case. All 16 client tests, TypeScript and Vite build passed; base image `sha256:32d184f6` and Brave image `sha256:9992699f` built, and the exact Brave image deployed healthy with zero restarts. Its first active attempt requested audio recovery after 12 seconds and then collided with server skew/progress enforcement; the next attempt issued two client skew recoveries 1.4 seconds apart, and both closed with `resync_limit`. A third attempt streamed about 71 seconds before terminal `4429 feedback_rate`. Queue evidence did not show transport pressure: 7,090/7,101 writes were below 1 ms and every observed queue depth was at most four. Static tracing found that both worker and server enforced the same sustained skew and that `setInterval` plus a fatal single inter-arrival check could compress delayed feedback.

The `86893473` correction uses a self-scheduled 1100-ms loop, a server token bucket of one per second with burst two, server-only sustained-skew recovery, fresh post-resync feedback/progress/skew observation windows and explicit resync-processing logs. All 17 client tests, TypeScript and Vite build, the complete server package sequence, 790,415 30-second fuzz executions and the trailing build passed. Base image `sha256:3d6f3a9e` and Brave image `sha256:a5394fa1` built; the exact Brave image deployed healthy with zero restarts and passed every public/direct pre-upgrade security probe. After an initial short setup interval, one uninterrupted active browser window added 46,267 Opus and 23,136 VP8 units over roughly 15 minutes. No WebCodecs resync series, failed WebCodecs close, recovery log or container restart occurred. The operator observed a continuously visible picture without conspicuous black flashes or stutters and audible approximately synchronized audio. The resulting approximately 25 VP8 units/s match `deploy/adaptive-quality.yaml`'s shared `high` source at 25 fps, rather than showing WebCodecs transport loss. WebCodecs nevertheless appeared subjectively slightly less fluid than WebRTC, so a controlled presentation-pacing A/B remains open. `max-quantizer: 63` may affect motion detail but not frame cadence. The remaining grouped matrix is still required; Phase 4 remains open.

The operator then reported that the same fullscreen control worked on desktop but not on a phone. Static inspection found that the player-element Promise rejection was swallowed before the existing surface fallback and that iPhone Safari cannot natively fullscreen the WebCodecs canvas. Commit `3c780664` makes the request helper report synchronous and asynchronous failure, invokes iPhone's video-only native API before awaiting an unsupported container request, converts fullscreen state handling from a synthetic-instance arrow callback to a Vue-bound method, and adds a viewport-filling app fallback with a safe-area-aware explicit exit control and body-scroll cleanup. Two focused client tests were added. At exact documentation follow-up `150aac57`, all 19 client tests, TypeScript and Vite build passed; a fresh image deployed healthy with zero restarts and passed every public/direct pre-upgrade probe. Desktop native enter/exit worked. On iPhone, WebCodecs entered and reliably exited the explicit viewport app mode with browser chrome retained as documented, while WebRTC preserved its separate native video fullscreen.

The same `150aac57` matrix confirmed ordinary, admin and view-only VP8/Opus receive, private-mode pause/resume without reload, and explicit return to working WebRTC audio/video/control. A metrics interval added 14,393 VP8 units over 573 seconds (25.12/s), confirming the shared adaptive 25-fps cadence rather than a WebCodecs-only 25-versus-30 difference. It also isolated one short black/audio interruption to a common `client_audio_underflow` recovery. Later, a still-visible foreground iPhone sent three such recoveries in quick succession, reached the bounded `resync_limit` close and automatically retried; this was a Phase 4 failure, not background throttling. Commit `95746173` keeps 100-ms AudioWorklet starvation local, discards/accounted raced PCM, reanchors the next audio at 160-ms lead within the existing 200-ms cap and retains the canvas. Server rendered-lag/skew, decoder and overflow recovery remain unchanged.

At exact documentation commit `6b6cd328`, the target-server client validation passed all 20 tests, TypeScript and the Vite production build. The complete Go package sequence, 30-second `FuzzParseRecord` run with 849,148 executions and trailing server/plugin build passed. Base image `sha256:2ac7a765` and Brave image `sha256:fe6b70ad` built; the latter deployed healthy with zero restarts, contained the expected hashed client/worker assets and passed every public/direct 400/401/403/426 security probe. In the clean foreground-iPhone correction run, the final snapshot retained one active WebCodecs delivery after 16,372 VP8 and 32,792 Opus units, with one open/success, no WebCodecs close, no server `client_audio_underflow` or `progress_timeout`, no retry and no container restart. The operator kept the iPhone visible for more than ten continuous minutes and reported that picture and audio appeared okay.

The operator then declined the cumbersome induced three-client constraint run and chose an explicit bounded closure. Ten numeric desktop joins/glass-to-glass and A/V-skew measurement, the controlled presentation-pacing A/B, WebCodecs slow-client plus repeated adaptive down/up isolation, five-minute resource/cleanup comparison and the remaining successful/replayed/expired-ticket and live control/binary abuse cases were not executed. Automated parser/fuzz coverage and the live fixed-invalid HTTP boundary did pass, but they do not replace those omitted cases. Phase 4 is therefore sufficient for continued explicit opt-in productization on `testing`, not a full production/backend acceptance or permission to make WebCodecs automatic/default.

The exact URL selector and persistent in-video status were deliberate prototype surfaces. After the bounded Phase 4 closure, the operator selected their productization as the next implementation block: persist an explicit per-client backend selector in sidebar settings, retain the query as a diagnostic override, show a compact current-backend/status indicator and collapse the large overlay during healthy streaming. This remains manual selection with WebRTC as the default; it does not add automatic fallback or expand the receive-only transport boundary.

## IMPLEMENTED IN REPOSITORY — explicit per-client media-backend productization

Status: **implemented and statically reviewed on `testing` on 2026-09-15 / focused target-server checkpoint closed at exact implementation commit `12cfe43b` on 2026-09-19 with the live view-only fragment case explicitly not repeated**.

- Added the pure `client/src/neko/media-selection.js` policy boundary. Only `webrtc` and `webcodecs-ws` are valid persisted values; missing, inaccessible or invalid browser storage resolves to WebRTC.
- Added a persisted **Default Media Backend** selector to the existing sidebar settings. Changing it stores the normalized per-browser choice, removes only the stateless `media` diagnostic parameter and reloads through the existing startup path while preserving unrelated query parameters and the view-only fragment.
- Retained exactly one `?media=webcodecs-ws` value as the highest-priority stateless diagnostic override without mutating the stored default. A present wrong or duplicated selector remains on the safe WebRTC path instead of accidentally activating a stored experimental backend. Settings show the effective backend, override state and an explicit action to remove the override and return to the saved default.
- A stored WebCodecs choice now selects the same isolated receive path without requiring a query. A new/unset client still selects WebRTC and sends no media-prototype event. The separately configured server feature remains default-off.
- Replaced the always-large healthy `WebCodecs receive prototype` panel with a compact localized `WebCodecs` indicator. Idle, negotiation, connection, recovery and terminal states retain the prominent localized panel, the full receive-only/no-Picture-in-Picture/no-automatic-fallback explanation and terminal **Retry WebCodecs** / **Use WebRTC** actions.
- **Use WebRTC** now also persists WebRTC before removing the diagnostic selector and reloading, so a prior stored WebCodecs preference cannot immediately reactivate the receive path.
- Added dependency-free focused coverage for persistence, absent/invalid/default behavior, storage denial, exact URL precedence, wrong/duplicate URL safety, query/fragment-preserving selection navigation, healthy-versus-recovery status treatment and source-level sidebar/action wiring. The complete client test command now includes `tests/media-selection.test.mjs`.
- No server, protocol, authorization, event WebSocket, WebRTC signaling/data-channel/opcode, capture, deployment or Compose source was changed. No automatic fallback, HLS/LL-HLS, WebTransport or replacement input transport was added.

Static review confirms that default WebRTC still uses the existing `undefined` internal backend marker, so `BaseClient.connectSocket` does not append a media selector and continues its unchanged WebRTC offer/signaling path. Only the explicit stored or exact-query WebCodecs result sets the existing `webcodecs-ws` marker and suppresses WebRTC signaling.

At exact `12cfe43b`, the target-server validation container passed all 26 client tests, TypeScript and the Vite production build. Fresh base and Brave images built; Brave image `sha256:2295d7198e020a9ebf365c610188d15cf3e5c3950ceffa41d81d9d5a33b8a71b` deployed healthy with zero restarts. Every recorded public/direct security probe passed, the deliberately disabled route returned 404, and the enabled stack was restored healthy with zero restarts.

The operator confirmed working WebRTC for absent and invalid storage; working stored WebCodecs without a query across reload; explicit return to WebRTC A/V/control; exact-query precedence without saved-default mutation; unrelated-query preservation; and the compact healthy WebCodecs indicator. With the server overlay omitted, the client displayed the prominent terminal detail `The server did not advertise the opt-in backend`, did not switch to WebRTC automatically and returned to working WebRTC through the explicit action. The live compact view-only fragment-preservation case was not repeated; focused automated coverage passed. Credential-safe evidence is outside the worktree in `../neko-productization-results-12cfe43b`. This closes the focused productization checkpoint only and leaves every previously deferred full Phase 4 gate deferred.

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

For the accumulated iOS recovery and view-only blocks on `testing`, the complete containerized client sequence passed at exact commit `913a981e`. The manual View-only matrix and compact-link follow-up are recorded above. The operator deliberately closed the checkpoint without executing the optional remaining iPhone deep-test phases in [`IOS_RECOVERY.md`](IOS_RECOVERY.md).

For the WebCodecs/media-WebSocket implementation through the bounded Phase 4 checkpoint, the earlier exact results are recorded above. The newer per-client selection/status productization passed its complete client sequence and focused deployed-browser matrix at exact `12cfe43b`, with the live view-only fragment case explicitly omitted as recorded above. No productization test/build/browser result was produced in Codex; all runtime evidence came from the target server and operator.

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
go test ./pkg/types ./pkg/auth ./internal/config ./internal/capture ./internal/media ./internal/mediaws ./internal/member/multiuser ./internal/session ./internal/http ./internal/http/legacy ./internal/websocket ./internal/webrtc
go test ./internal/mediaws -run '^$' -fuzz '^FuzzParseRecord$' -fuzztime 30s
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

The same-peer, replacement-session, bounded-exhaustion and server-directed-disconnect phases in [`IOS_RECOVERY.md`](IOS_RECOVERY.md) remain the required procedure for any future claim of automatic no-reload recovery on a real iPhone. They were deliberately not executed for the checkpoint closed on 2026-09-11.

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

## HLS/LL-HLS design block — complete in repository

[`HLS_LL_HLS.md`](HLS_LL_HLS.md) now fixes the default-off version-1 passive delivery contract without implementing the transport. It specializes the existing encoded-provider and central participant-delivery boundary as follows:

- one shared in-memory packager set consumes bounded provider subscriptions; its one AAC audio rendition and at most three H.264 video renditions are shared across viewers and across conventional/low-latency playlist views;
- current VP8/Opus is explicitly not treated as native HLS. The first evidence-led output is separate fMP4 H.264 High 3.1 video plus 48-kHz stereo AAC-LC, with exact initial target variants and a required desktop/Apple/actual-Smart-TV matrix;
- conventional HLS uses six-second parents and an 18-second hold-back. LL-HLS adds one-second parts, three-second part hold-back, blocking reloads, preload hints and rendition reports over the same objects;
- every viewer remains one centrally owned `MediaDelivery`. A ten-second one-time ticket bootstraps a 30-second sliding playback lease using an independent path-scoped Secure/HttpOnly/SameSite cookie; no login/share credential or playback bearer appears in a playlist URL;
- same-origin HTTPS, exact Origin on mutating requests, trusted-proxy CIDRs, uniform invalid-lease responses, no CORS/CDN/shared cache, fixed request/rate/object limits, memory-only 64-MiB retention and credential-free metrics are normative;
- private mode, revocation, replacement and shutdown remain central lifecycle operations. Already downloaded HTTP bytes cannot be revoked retroactively, so the exact visible-window bound and cooperative/server stop deadlines are explicit acceptance evidence;
- repository Phases 1–4 and the exact security, role, device, latency, adaptation, slow-reader, resource and cross-backend isolation gates are fixed.

No HLS route, packager, encoder, player, dependency, Compose overlay, automatic selection, DASH/WebTransport implementation or new control transport was added in this block. Runtime/build/test status for the design follow-up is **NOT EXECUTED IN CODEX**. The small client follow-up in the same repository block reduces the already-compact healthy WebCodecs indicator to only `WebCodecs`; prominent negotiation/recovery/terminal behavior is unchanged.

## NEXT

Continue exclusively on `testing`; do not merge, fast-forward or push changes to `master`. The stable branch remains pinned at `d9105ef8` until the operator explicitly authorizes a later grouped promotion.

Validate the revised receiver-evidence-gated adaptive downgrade correction on one exact clean `testing` commit with the compact two-viewer procedure in [`ADAPTIVE_QUALITY.md`](ADAPTIVE_QUALITY.md): one healthy viewer must remain on `high` through an extended unshaped soak and all constrained phases even if its GCC target collapses while receiver evidence stays clean; one independently shaped viewer must obtain real receiver loss/NACK evidence, step `high -> medium -> low`, and recover `low -> medium -> high` without cross-peer drops or rapid reversal. Run the focused server tests/build and image/deployment checks only on the target server, retain filtered metrics/logs, and restore the host shaper.

After that focused gate passes or its evidence is explicitly dispositioned, implement **Phase 1 of [`HLS_LL_HLS.md`](HLS_LL_HLS.md)** without starting media packaging or adding a player. Its already specified configuration, access, lease, security, deterministic playlist/object and focused-test scope remains unchanged. `master` must not move without explicit operator authorization.

## Product priority after stable synced baseline

1. **bounded checkpoint closed with the documented final-matrix limitation:** media-subscription/WebRTC compatibility refactor plus estimator startup correction;
2. **bounded checkpoint closed with explicit final-matrix limitations:** WebCodecs plus dedicated media WebSocket receive path;
3. **implemented / focused target checkpoint closed with the live-fragment limitation:** persisted explicit per-client selection in sidebar settings with a compact backend/status indicator, WebRTC default and diagnostic URL override;
4. **completed design:** exact default-off HLS/LL-HLS passive/view-only contract in [`HLS_LL_HLS.md`](HLS_LL_HLS.md);
5. **NEXT:** focused exact-commit healthy-plus-constrained target validation of the revised receiver-evidence-gated adaptive downgrade correction;
6. **queued after that gate:** implement HLS/LL-HLS Phase 1 access and deterministic-media foundations without a packager, route or player;
7. promote accumulated `testing` history only after an explicit operator decision at a coherent validation milestone.

## Fallback prototype sequence

Alternative media architecture work began after the operator closed the grouped checkpoint with the explicit iPhone evidence limitation recorded above. This does not convert the omitted no-reload device phases into a pass. All work remains on `testing` until the operator explicitly approves a later grouped promotion.

When fallback work begins, separate the two user classes instead of forcing every client through one fallback chain:

1. **completed design:** establish the backend-neutral encoded-source/subscription and participant-delivery contract in [`MEDIA_SUBSCRIPTION_BOUNDARY.md`](MEDIA_SUBSCRIPTION_BOUNDARY.md);
2. **implemented / bounded target-server checkpoint closed:** the migrated WebRTC compatibility path and estimator startup correction, with the repeated full role/recovery and induced down/up matrix deferred to final grouped validation;
3. **completed design:** the exact protocol, security, queueing, synchronization, rollout and validation contract for **WebCodecs + dedicated WebSocket media** is fixed in [`WEBCODECS_MEDIA_WEBSOCKET.md`](WEBCODECS_MEDIA_WEBSOCKET.md);
4. **implemented Phase 1:** strict framing/fixtures, PTS-validity propagation, authenticated negotiation, one-time tickets and default-off server negotiation;
5. **implemented Phase 2:** credential-free server delivery backend, secure media route, bounded queues, recovery and lifecycle cleanup;
6. **implemented Phase 3:** isolated strict client parser/worker/WebCodecs/AudioWorklet/canvas receive path, explicit UI actions and bounded same-backend recovery;
7. **bounded Phase 4 checkpoint closed:** exact automated/build/security, accumulated functional and corrected foreground-iPhone gates passed; numeric latency/pacing, induced isolation, resource and remaining live hostile-input cases stay deferred;
8. **implemented / focused target checkpoint closed with the live-fragment limitation:** persisted manual per-client `WebRTC`/`WebCodecs` selection and compact healthy status without automatic fallback;
9. **completed design:** exact **HLS / Low-Latency HLS** passive/view-only contract for Smart-TVs and constrained browsers in [`HLS_LL_HLS.md`](HLS_LL_HLS.md);
10. **queued after the revised adaptive gate — HLS Phase 1:** default-off access, lease, security and deterministic playlist/object foundations without a packager or client;
11. later Phases 2–4 add the shared packager/HTTP delivery, isolated passive client and grouped target-server validation in that order;
12. compare device support, failure behavior, server resource cost, latency and recovery before defining any automatic capability-based selection;
13. evaluate WebTransport only afterward if WebSocket's delivery/backpressure characteristics are a demonstrated limitation.

The passive path may trade latency for reliability and compatibility. It must stay in the same logical room and must not gain control authorization. HLS/LL-HLS now has a specified target contract, but no implementation or device evidence yet.

## LATER / OPTIONAL

- WebTransport productionization;
- fully automatic transport selection after explicit/manual fallback paths are proven;
- automatic codec selection;
- MPEG-DASH as an additional passive HTTP-streaming backend where useful;
- MJPEG only as an ultra-legacy image-only fallback for a concrete unsupported device class;
- broader v3 client/library rewrite.

## OPEN

- Final grouped validation must repeat the ordinary/admin/view-only/private-mode/manual-tier/reconnect matrix and the independently constrained three-viewer adaptive down/up isolation run; neither was rerun at `e5f55bf9`.
- The first steady-state downgrade candidate at `2d037f39` is rejected by the captured loss-free `high -> medium -> low -> medium -> high` trace. The revised receiver-evidence-gated correction still needs its compact exact-commit two-viewer target-server gate; revised repository tests/builds were not executed in Codex.
- Determine whether fast-motion softness is acceptable at the current 1,996,800-bit/s VP8 `high` tier through a controlled same-content bitrate/quantizer A/B with receiver statistics and comparable captures; current evidence does not identify a subscription-refactor regression.
- Supported Smart-TV/device matrix, including native HLS, MSE/DASH and WebCodecs capability.
- Whether the target iPhone validates the implemented same-peer and replacement-session paths without reload; a Safari Play gesture remains an explicitly separate, permitted policy fallback.
- Target-device and target-server evidence for the specified VP8/Opus WebCodecs/media-WebSocket contract; framing, queue sizes, synchronization, security limits and rollout behavior are fixed in [`WEBCODECS_MEDIA_WEBSOCKET.md`](WEBCODECS_MEDIA_WEBSOCKET.md).
- Actual HLS/LL-HLS device, latency and resource evidence against the fixed targets in [`HLS_LL_HLS.md`](HLS_LL_HLS.md); no transport is implemented yet.
- Whether DASH adds meaningful compatibility beyond HLS for the actual target devices.
- Eventual automatic per-client media-backend selection rules after the explicitly selected prototypes have measured evidence; version-1 manual selection and rollback are already fixed.
- Live compact view-only fragment preservation for the productized selector was not repeated at `12cfe43b`; focused automated coverage passed, and earlier view-only boundary/browser evidence remains separate.
- Longer-term legacy Vue 2 migration.
