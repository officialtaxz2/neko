# Work Plan / Handoff State

Last consolidated: 2026-09-23.

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
testing: deployment reconciliation, accepted opt-in adaptive quality plus validated recovery probe, bounded iOS recovery, server-enforced view-only sharing, implemented media-subscription/WebRTC compatibility, WebCodecs/media-WebSocket receive path and separate Phase 4 assets, and HLS/LL-HLS Phases 1–2 server foundations/shared packaging/authenticated HTTP delivery
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

## COMPLETED IN REPOSITORY — adaptive downgrade correction and bounded recovery follow-up

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

Exact `2efcc6b1` target evidence subsequently closed three parts of the revised gate:

- focused capture/WebRTC tests, server/build images and the deployed adaptive stack passed and remained healthy with zero restarts;
- one unshaped H stayed on `high` for 20 minutes while its estimator target ranged from about 1.71 to 4.57 Mbit/s, with zero loss, zero local video drops, no confirmed-congestion log entry and no switch; ten isolated NACK entries never persisted as downgrade evidence;
- endpoint-specific shaping moved only C through `medium` to `low`, while H stayed visually good on `high` without a drop or tier change.

The recovery part did not pass its 90-second bound. After restoring `fq_codel`, C remained clean on `low` with an application-limited target around 0.49–0.65 Mbit/s, below the approximately 1.06-Mbit/s nominal medium gate. It eventually returned to `high` later at about 3.76 Mbit/s, proving delayed recovery rather than a permanent lock. The current follow-up therefore:

- retains the ordinary next-tier nominal reserve gate;
- records one peer-local recovery step per successful automatic downgrade and permits exactly one higher-tier probe only while such a step remains, after a clean stable window, configured reserve over the current complete-delivery reference and a separate 30-second profile interval;
- validates the probe-selected tier for a complete clean stable window;
- returns a genuinely insufficient probe through the existing receiver-evidence downgrade and applies exponential failed-probe backoff capped at two minutes in the profile;
- exports per-session active/attempt/success/failure probe metrics and preserves the evidence in the collector;
- adds focused tests for application-limited probe entry, clean completion, failed-probe backoff/cap and peer isolation while retaining the exact existing startup/downgrade/upgrade boundary tests.

Target status: **receiver-evidence downgrade and bounded recovery complete for the focused target path**. The original false downgrade was validated at `2efcc6b1`. Exact follow-up `ddf15cee` then passed the configured server packages, 30-second media-WebSocket fuzz run with 1,357,656 executions, server/plugin and fresh base/Brave builds, adaptive deployment preflight, healthy zero-restart service and the compact two-viewer gate. A 700-kbit/s endpoint-specific constraint moved only C `high -> medium -> low` with fresh receiver evidence while H remained `high` with zero video-drop delta. Once `fq_codel` was restored, C started one clean probe at 30 seconds, completed it successfully, and reached `high` at 60 seconds; final probe metrics were active/attempt/success/failure `0/1/1/0`. The operator observed C frozen above `low`, fluid on `low`, and both recovery steps. The ad-hoc wrapper's final log-file counters produced a false exit `1`, but its CSV/metrics, direct container log follow-up, final queue/container state and operator observation establish the functional pass. Evidence remains outside the worktree in `../neko-adaptive-recovery-ddf15cee-20260923T063447Z`.

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
go test ./pkg/types ./pkg/auth ./internal/config ./internal/capture ./internal/media ./internal/mediahls ./internal/mediaws ./internal/member/multiuser ./internal/session ./internal/http ./internal/http/legacy ./internal/websocket ./internal/webrtc
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

## COMPLETED IN REPOSITORY — HLS/LL-HLS Phase 1 foundations

Implemented and statically reviewed on `testing` on 2026-09-23:

- added default-off `media.hls` configuration with explicit exact HTTPS origins, trusted proxy CIDRs, allowed `hls`/`ll-hls` modes and bounded global lease/request ceilings;
- added strict current and legacy HLS capability/create/offer events, passive-session negotiation, payload-redacted event logging and live connected-`CanWatch` re-resolution;
- added peer-local create-rate limiting and digest-only ten-second, 24-byte-entropy, single-use bootstrap tickets bound to session/backend/version/mode/request; replacement, replay, expiry, disconnect, deletion and permission-loss invalidation are explicit;
- added strict 2-KiB bootstrap JSON parsing and a separate playback-lease store with 16-byte public IDs, digest-only 24-byte secrets, per-session replacement, 30-second sliding activity only for playlist/keepalive, pause/revocation, path-scoped `Secure`/`HttpOnly`/`SameSite=Strict` cookies and fixed per-lease/global request, blocking and rate limits;
- added exact HTTPS/origin/trusted-proxy checks, canonical fixed path parsing, bounded LL-HLS query directives, single byte ranges, no-store/no-referrer headers and credential-free normalized access-log paths;
- added immutable generation/media-object models with fixed per-object, count and 64-MiB retention ceilings plus deterministic master, conventional HLS and LL-HLS playlist rendering with relative credential-free URIs and golden fixtures;
- added focused tests for config, strict payloads, ticket replay/expiry/binding/replacement, lease cookie/scope/lifetime/rates/concurrency/revocation, security/path/query/range/redaction, playlist goldens and object/generation invariants.

The bootstrap and resource paths are deliberately **not registered**. Phase 1 adds no packager, H.264/AAC conversion, encoded-provider subscription, HTTP media delivery, player, deployment overlay or client/backend selection. WebRTC remains the default and WebCodecs remains explicit.

Runtime/build/test status: **NOT EXECUTED IN CODEX**. The focused target-server gate for this block is pending:

```bash
cd server
go test ./internal/config ./internal/mediahls ./internal/http/legacy ./internal/websocket ./pkg/types/...
go test ./...
./build
```

## COMPLETED IN REPOSITORY — HLS/LL-HLS Phase 2 shared packager and HTTP delivery

Implemented and statically reviewed on `testing` on 2026-09-23:

- added one process-wide packager set shared by `hls` and `ll-hls` leases, with one exact Opus audio subscription and exact `high`/`medium`/`low` VP8 subscriptions, provider queues fixed at 64, observable input/output worker hand-offs fixed at 8 and 15-second idle teardown after the last unpaused lease;
- required the first complete provider `FORMAT`, its nonzero generation and an exact advertised dimension/rate match before media admission; added normalized-timestamp-preserving GStreamer appsrc input and codec-configuration extraction, then bounded VP8-to-H.264 High 3.1 and Opus-to-48-kHz stereo AAC-LC conversion at the fixed rendition rates, dimensions and frame rates;
- added deterministic single-track fMP4 init/fragment generation, aligned one-second parts/two-second IDRs/six-second parents, shared conventional/LL playlist publication, immutable object serving, bounded count/size/64-MiB retention and generation/discontinuity restart handling;
- registered the central `hls` backend and bootstrap/keepalive/resource routes only when `media.hls.enabled=true`; omitted configuration still exposes no HLS route or backend and does not change WebRTC or explicit WebCodecs selection;
- connected every playback lease to the central per-session delivery owner. A successfully served media playlist activates watching; an initially private session attaches in the paused state without starting workers, Private Mode wakes requests, pauses the digest-only lease and releases packager demand, and resume keeps HTTP paused until the shared set is ready; replacement, expiry, permission loss and shutdown close centrally;
- implemented exact cookie renewal on master/media-playlist and keepalive responses, uniform invalid-lease handling, per-request read/write deadlines, `HEAD`, one-range MP4 delivery, negotiated playlist gzip, fixed HTTP priorities, batch/CORS exclusion and path/query redaction without applying a global timeout to existing WebSockets;
- kept slow readers outside provider/packager locks by serving immutable object references, bounded every request/blocker/rate/object path, evicted only retention beyond the currently advertised init/part/parent floors when the 64-MiB aggregate ceiling requires space, and made GStreamer callback shutdown cancellation-aware so worker teardown cannot wait forever on a full Go hand-off;
- added all fixed `neko_media_hls_*` metrics and credential-safe lifecycle/failure logs plus focused fMP4, shared-worker, pause/revocation, cookie-renewal, compressed-HEAD and range tests.

At this Phase 2 closure there was no HLS/LL-HLS browser player, per-client HLS selector, deployment Compose overlay or automatic fallback. Phase 3 below has since added the isolated player and manual selector. No HLS route exists in the normal deployment while the default-off server flag is omitted.

Runtime/build/test status: **NOT EXECUTED IN CODEX**. Before Phase 2 is treated as target-server verified, run this exact-commit gate on the target server:

```bash
export NEKO_VALIDATION_COMMIT="$(git rev-parse HEAD)"
docker compose -f docker-compose.validation.yaml build server-checks
docker compose -f docker-compose.validation.yaml run --rm server-checks
docker build -t my-neko/server:hls-phase2 ./server
```

The runtime gate must additionally inspect generated H.264/AAC init/fragments and both playlist modes with an independent HLS/fMP4 parser, verify the disabled route is `404`, and exercise the authenticated HTTPS cookie/range/HEAD/gzip/private-mode/revocation/slow-reader/shutdown boundaries before any device-support or latency claim. The Phase 4 overlay and evidence collector are intentionally not part of this block.

## COMPLETED IN REPOSITORY — HLS/LL-HLS Phase 3 passive client

Status on 2026-10-04: **implemented and statically reviewed on `testing`; the focused automated target-server gate passed at `e85d8568`, and the default-off base/Brave image deployment started healthy at `741025c3`; browser/media smoke checks and HLS-enabled runtime/device acceptance remain pending**.

The isolated HLS controller and pinned local hls.js 1.7.3 MSE/worker path now integrate with the native video/playback UI. Advertised-only manual HLS/LL-HLS choices appear for passive viewers and admin diagnostics; exact diagnostic queries and saved choices preserve unrelated navigation state. WebRTC remains the absent/invalid default, WebCodecs remains supported, disabled HLS terminates visibly, and no automatic fallback or deployment overlay was added. The authenticated event session skips unused WebRTC signaling for exact HLS selections. Private-mode state now crosses the legacy bridge, and native/MSE cleanup owns buffers, listeners, requests and stale callbacks. Prefix-scoped cookies/CORS/logging and six-part blocking-reload rollover were corrected during static integration. See [`HLS_LL_HLS.md`](HLS_LL_HLS.md) for the full implementation boundaries.

The pre-push static review supplied the missing standard `URL` constructor to the isolated MSE-loader test VM, so valid scoped requests can reach the intended loader assertions. These tests passed in the target-server gate recorded below.

Runtime/build/test status: **NOT EXECUTED IN CODEX**. On the real target server at the eventual exact reviewed commit, run the accumulated client and server gate:

```bash
export NEKO_VALIDATION_COMMIT="$(git rev-parse HEAD)"
docker compose -f docker-compose.validation.yaml run --rm client-checks
docker compose -f docker-compose.validation.yaml build server-checks
docker compose -f docker-compose.validation.yaml run --rm server-checks
```

The client container runs `npm ci`, `npm test`, `npm run lint` and `npm run build`, including HLS controller/protocol tests and server playlist fixtures. The server container includes HLS, common media, legacy/current events, HTTP, auth/session and existing WebRTC/WebCodecs checks plus the build. In Phase 4, validate exact images, disabled/enabled HTTPS delivery, root/prefixed cookie renewal/redaction/CORS, both playlist modes over repeated parent rollovers, ordinary/admin/view-only/private/revoked roles, autoplay/gesture/fullscreen/PiP and induced mixed-backend isolation/resources. Measure startup and latency separately. None of those runtime or device results is claimed here.

### Target-server automated gate supplied on 2026-10-04

The operator first saved the mixed uploaded working tree, fast-forwarded `testing` to exact `e85d8568` and supplied a clean Git status. The subsequent containerized validation log reported exit code `0`: `npm ci`, all 42 client tests, TypeScript and Vite build passed; the validation image built; all 13 focused Go packages passed, including HLS, HTTP/legacy, auth/session and WebRTC; the WebSocket parser fuzz test passed with 1,159,263 executions; and the trailing server/plugin build passed. No project code or runtime checks were executed in Codex.

The running older deployment reported the adaptive-quality configuration and WebCodecs/media-WebSocket enabled. Both explicit overlays must be preserved during image deployment. At this automated gate, base/Brave image builds, actual service replacement, health and browser/media/device acceptance were still pending; subsequent image/deployment evidence is recorded below. This automated gate is not full HLS acceptance.

`npm ci` also reported 20 audit findings (11 low, 3 moderate, 5 high and 1 critical). No detailed advisory report was supplied, so affected packages, production relevance and repairs remain unclassified. Track this in the final stability/security review; do not treat passing tests as a dependency-security clearance or apply a blind breaking dependency update.

### Target-server deployment build interruption on 2026-10-04

The subsequent deployment attempt at `e85d8568` passed Compose configuration validation but failed while building the base image: after `npm install` and `COPY client .`, `npm run build` reported `vite: Permission denied` (exit `126`); the outer deployment reported exit `1`. The script stopped before the Brave image build and before stopping or replacing the running service. This log supplies no new deployment-health or playback acceptance.

Static inspection found no Docker ignore file: the generated repository-root build context could copy host `client/node_modules` over the Linux dependencies installed in the image. Dependencies left by the earlier Windows file upload are a plausible explanation; the target host's actual dependency files were not inspected. The correction excludes host dependencies at every depth in the root and standalone client contexts, excludes local credentials/profile/download state from the root context, and uses `npm ci` against the same lockfile as the passing validation gate. The root ignore file deliberately retains `dist` directories for the existing `CLIENT_DIST` build option. No dependency versions or application source changed.

Correction status: statically reviewed and the target-server image-build/deployment follow-up passed at `741025c3`, as recorded below; **NOT EXECUTED IN CODEX**. The successful automated gate above remains evidence for `e85d8568`; the full test suite was not repeated at the Docker-context-only correction.

### Target-server image/deployment checkpoint supplied on 2026-10-04

The operator supplied the follow-up log after fast-forwarding to exact `741025c3`. The base-image client stage ran `npm ci` and Vite 6.4.3 successfully, including the bundled HLS worker; the server build passed, and both `my-neko/base:test` and `my-neko/brave:test` images built successfully. The earlier `vite: Permission denied` failure did not recur. The build log recorded these manifest-list digests:

- base: `sha256:e55f0825a67032c621dc050b715f545c82350b17a2ecf7faf9329dbbf47f4f5f`;
- Brave: `sha256:756540cce0345b650861a7fbb7b1370f84bb49df89d3d97b2bea5c7931b6e1d6`.

Compose used the base, adaptive-quality and WebCodecs/media-WebSocket files, stopped the previous service, completed the ordered profile-lock cleanup, recreated `neko-neko-1` with `my-neko/brave:test` and reported `Healthy`. Its final status was `Up 14 seconds (healthy)` and the deployment exit code was `0`; the rollback branch was not entered. This confirms initial image deployment and health, not a long soak, zero restarts or successful browser/media playback.

HLS stayed default-off: no HLS enablement overlay was supplied and no packager/player runtime acceptance follows from the HLS assets being built. Next, check available-device WebRTC and explicitly selected WebCodecs playback while another participant joins, chats and takes/releases control. HLS Phase 4 enablement, the final stability review, the grouped security/role/isolation/resource matrix and the unavailable television's device evidence remain open. The npm audit findings above also remain unclassified. No target-server commands or project code were executed in Codex.

## Operator direction — event-triggered TV failures and final stability review

On 2026-10-04 the operator reported occasional older-TV WebRTC failures specifically around room join, chat and control take/release, with quiet viewing working; VIDAA is suspected but exact devices are unknown and the colleague's TV is currently unavailable. WebCodecs/media-WebSocket has not been tested on those devices. Mobile time-to-first-picture is a separate observation. The shared chat-sound path is a source-inspection candidate, not a diagnosis.

Phase 3 was authorized and implemented without waiting for that television; its device gate remains open through Phase 4. Preserve WebCodecs and treat HLS as a compatibility prototype with pending device evidence. The detailed evidence limits, one-variable comparison and **mandatory final static stability review before grouped target-server validation and promotion** are in [`STABILITY_REVIEW.md`](STABILITY_REVIEW.md). Do not claim that HLS fixes room-event failures or starts faster than WebRTC.

## HLS Phase 4 repository assets and review — 2026-10-04

After the healthy `741025c3` deployment, the operator reported that things seem to work and authorized Phase 4. This is bounded browser smoke evidence without an exact device/backend/role matrix. The operator confirmed Caddy runs as a host system service; its version/configuration and safe access/runtime logs still need target review.

Phase 4 now supplies the separate conventional-first HLS overlay, exact HTTPS/proxy fallback to the reviewed WebCodecs settings, private evidence/result collection, synthetic credential-free public/direct HTTP probes, target-only automated/image preparation and captured-image enable/rollback helpers. Base Compose and `master` remain unchanged. Adaptive and WebCodecs overlays are preserved during both enablement and rollback. LL-HLS remains an explicit later enablement after its actual public-protocol/RTT gate.

The integrated [review record](STABILITY_REVIEW_2026-10-04.md) identifies and repairs legacy optional-chat-audio failures and long private-pause lease loss. Only authenticated bounded keepalive can renew during pause; media bytes remain denied, and expiry/revocation/rate bounds remain intact. Focused notification, long-pause client/Go and HTTP-controller tests plus a bounded HLS request-parser fuzz target are added. No speculative encoding, transport removal or dependency upgrade is included.

All new runtime/build/test status is **NOT EXECUTED IN CODEX**. Source/configuration review items 1–7 are recorded; detailed dependency-advisory classification (item 8), exact new-block checks and the complete enabled acceptance matrix are pending. The earlier 20 findings are not cleared. [HLS_LL_HLS_VALIDATION.md](HLS_LL_HLS_VALIDATION.md), [HLS_LL_HLS_CADDY.md](HLS_LL_HLS_CADDY.md) and [HLS_LL_HLS_OBSERVABILITY.md](HLS_LL_HLS_OBSERVABILITY.md) define the target checkpoint and explicit evidence limits.

## HLS Phase 4 automated/image checkpoint — exact `e55bcd7e`

The operator supplied the complete successful preparation output on 2026-10-04
for `e55bcd7e256c05ca0053d705879ca3adfd8c4e65`, a clean `testing` checkout.
Caddy printed `2.6.2` and its service state was `active`. The operator separately
supplied `neko.taxzvps.de { reverse_proxy 127.0.0.1:8082 }` with only commented
old-port/upload examples, no access logger/imports/global options. The existing
route is sufficient; runtime/error-log filtering and target proof remain open.
Quiet Compose and
target shell syntax checks succeeded; the HTTP checker's help invocation
succeeded, without executing any HTTP security probe.

All **44 client tests** passed, including legacy notification handling and the
46-second private-pause controller test, followed by `tsc --noEmit` and Vite
production build. The HLS worker/player bundles were emitted. All **13 selected
Go packages** passed, including the new paused-renewal/rate/media-denial tests.
The WebSocket fuzz job passed **963,307 executions** and the HLS request-boundary
fuzz job passed **291,779 executions**, followed by the server/plugin build.

Both exact images `my-neko/base:hls-e55bcd7e256c` and
`my-neko/brave:hls-e55bcd7e256c` built successfully. The log records their build
manifest-list digests as `sha256:33ff896c2738768c0c4ad9ebb4f530575e37e826a4331ccc82186205bd0b165e`
and `sha256:1ae894e8ef11963a28d68e58cf2a43a62cfa7e8e1587464b591cd54bf62d8efb`;
the follow-up supplied the same IDs from private server `images.txt`.
Evidence was saved to `/opt/docker/nekoNew/neko-hls-results`.
The helper printed `AUTOMATED/IMAGE GATE PASSED; running service unchanged` and
`Check-Exitcode: 0`. It did not deploy either new image or enable HLS.

The dependency audit returned **1**, again with 20 findings (11 low, 3 moderate,
5 high and 1 critical). Its detailed private report was subsequently supplied
and classified as recorded below; the image gate is not dependency-security
acceptance. Chunk-size and upstream deprecation notices did not fail the build.
All enabled security, role, playback/device, latency/isolation/resource and
rollback gates remain open. Codex executed no project checks or runtime code.

## HLS audit/default-off follow-up and client containment — 2026-10-04

The operator supplied the complete audit v2 report and saved image IDs at exact
`e55bcd7e256c05ca0053d705879ca3adfd8c4e65`. The credential-free public HTTP
follow-up passed bootstrap/media 404 checks (2/2, exit 0). Saved image IDs:

- `my-neko/base:hls-e55bcd7e256c`: `sha256:33ff896c2738768c0c4ad9ebb4f530575e37e826a4331ccc82186205bd0b165e`;
- `my-neko/brave:hls-e55bcd7e256c`: `sha256:1ae894e8ef11963a28d68e58cf2a43a62cfa7e8e1587464b591cd54bf62d8efb`.

The [dependency review](DEPENDENCY_AUDIT_2026-10-04.md) classifies all 20 package
entries against the lock and source paths. The critical `form-data` finding is
an Axios Node dependency; the inspected browser upload uses native FormData
and no deployed Node request process is configured. This is a reachability
assessment, not security acceptance. Browser/compiler review established three
bounded client repairs: remove participant-controlled Vue-template compilation
and escape generated attributes, render member-name tooltips as literal text,
and disable unused automatic Axios XSRF-cookie reading. Emoji names now use
native title tooltips. Media transport/quality and dependency versions are
unchanged. Real-parser/Vue-component security/formatting regressions are added.

These follow-up changes are statically reviewed only: **NOT EXECUTED IN CODEX**.
The e55 images and automated pass do not contain them. Prepare a new exact
test/image checkpoint in its own private output directory before deployment.
Retain the e55 evidence. Remaining dependency updates/Vue 2 maintenance and
final security acceptance stay open; no blind major update/downgrade is applied.

The supplied Caddy 2.6.2 site is the minimal `127.0.0.1:8082` proxy, and the
operator confirmed `/etc/caddy/Caddyfile` as the active file. The tracked logging/
proxy proposal is not applied. Validate that proposal and the active file on
the target, then complete actual runtime/access-log checks before valid leases.
No HLS-enabled service, playback or device evidence follows from these results.

## HLS repair-image and Caddy validation checkpoint — exact `93f1fa63`

The operator supplied the complete recheck output on 2026-10-04 at exact
`93f1fa637ae3f14ba41d1bfef39e860d43993ff1`, clean `testing`. It records:

- all **47 client tests passed**, including the three real-parser/Vue-component
  chat security/formatting regressions; TypeScript and Vite 6.4.3 build passed;
- all **13 selected Go packages passed**; both requested 30-second fuzz jobs
  passed (WebSocket 1,117,452 executions; HLS request boundary 318,735);
- trailing server/plugin build and uniquely tagged base/Brave images passed;
- final `AUTOMATED/IMAGE GATE PASSED; running service unchanged`;
- active `/etc/caddy/Caddyfile` and private `Caddyfile.proposed` both printed
  `Valid configuration`, with final **Recheck-Exitcode 0**. The proposal only
  produced a formatting warning; it was not applied or reloaded.

Saved target image IDs are:

- `my-neko/base:hls-93f1fa637ae3`: `sha256:18414b80e56fc7f474f34e09d983497bb9881d35d9bde56d18d75f07d3216a53`;
- `my-neko/brave:hls-93f1fa637ae3`: `sha256:d217eccd941810f2436c4e3372ecd796b9e77b7758ecbaf5ed9e0c2a2a7fc4a7`.

Evidence is private at `/opt/docker/nekoNew/neko-hls-results-93f1fa637ae3`.
The audit still reports 20 entries with exit 1; no dependency versions changed
and no dependency-security clearance is implied. Browser/role/device checks,
HLS-enabled runtime, actual proxy-error logging, resources and rollback remain
pending. These are supplied target results; **NOT EXECUTED IN CODEX**.

A separate statically reviewed operator helper, `deploy/activate-hls-phase4.sh`,
was prepared with the supplied Neko snippet assumed to be the complete file.
It refuses unrelated options,
backs up/validates before reload, captures the prior Neko image, uses the
deployment stop to provoke one synthetic public 502, checks the private real
error journal for path/header redaction, then invokes the existing enable and
invalid-input checks. Failure attempts restoration of the old image/config.
It had not run at that preparation checkpoint; the subsequent guard result is
recorded below. No valid media credential is used by that gate.

Keep the application checkout at **93f1fa63** and its preparation marker.
Fetch the exact reviewed helper commit and use `git show` to extract only the
operator script into that private result directory. Record its commit; the
script records its blob hash and application hash. This docs/tooling follow-up
does not change application sources, dependencies or existing tested overlays,
and does not require rebuilding that already tested application image. A later
application change still requires its own exact gate.

## HLS activation guard checkpoint — application `93f1fa63`, helper `58ca75e9`

The operator executed the extracted helper on 2026-10-04. Output confirms
application `93f1fa637ae3f14ba41d1bfef39e860d43993ff1`, helper blob
`e4758df19978d1d8e8bca0398233be7aef14b8a3`, then:
`active Caddyfile differs from the supplied minimal site; no change applied`,
with **Activate-Exitcode 1**. Initialization and private activation metadata
were written; the helper stopped before backup, Caddy reload, image capture,
Neko stop or deployment. The previously running default-off deployment remains
unchanged by that attempt.

The supplied Neko snippet did not establish the contents of the complete active
Caddyfile. The exact difference is unknown; do not infer particular additional
sites/imports/global options without inspection. A read-only operator helper,
`deploy/inspect-hls-caddy.sh`, now adapts the actual configuration in memory and
reports only fixed Neko-proxy and logging counts/booleans. It is statically
reviewed, **NOT EXECUTED IN CODEX**, and awaits target output. Review that
structure before preparing a complete candidate preserving unrelated settings.
The sole-site activation guard is retained. No tests/images need repetition
for this docs/inspection-only follow-up; keep the tested application at 93f1fa63.

## Read-only Caddy inspection and preserving merge — application `93f1fa63`

The operator ran inspector helper `68a7f9047f81ecce5eb89204b15fc0b5acf174e8`
on 2026-10-04 with **Inspect-Exitcode 0**. The adapted configuration reported
one HTTP server, 11 distinct explicit hosts, one exact Neko route and one proxy,
whose sole upstream is `127.0.0.1:8082`. Forwarded-header removal, explicit
upstream compression disablement and explicit request/response buffering were
absent. Runtime logging contains the default plus one other logger; the default
encoder was `unset-or-other`, with zero filter fields. No exact Neko site logger
association or credential logging was reported. Other hosts, raw configuration
and credential values were not printed. This was read-only; HLS remains off.

The activation helper now has a matching offline `deploy/merge-hls-caddy.py`.
It supports the inspected bare Neko proxy/explicit site and an encoder-free
default logger, retains unrelated source text, and uses Caddy's actual adapted
JSON to require exact preservation of all other settings. Unknown proxy,
source/import/alias, debug, encoder, Neko access/fallback or extra runtime-error
logging structures are rejected before reload. Other site access loggers and
default writer/level/include/exclude options are preserved; the new default
filter also removes headers from other runtime error records. The temporary
candidate stays in `/etc/caddy` for relative imports and is cleaned up; raw
JSON/validation output stays private. Source/adaptation rechecks precede backup,
the existing synthetic-error log gate, tested-image activation and restoration
logic. The complete new helper is statically reviewed, **NOT EXECUTED IN CODEX**;
its merge/reload/log/enable gates await target output. It changes no application,
dependency or tested Compose source, so keep application/images at `93f1fa63`.

## HLS activation and invalid-input checkpoint — application `93f1fa63`

The operator supplied successful activation output on 2026-10-04 at application
`93f1fa637ae3f14ba41d1bfef39e860d43993ff1`, helper commit
`2484a022f1b6a140712c221d9d471832fef6f3b5`, shell blob
`d8fbb7ef9e2b9da13f85d0f3ac6aa24e3fa75cf3` and merger blob
`d653619a7ede7b2ecf37966f24bd07ef2d5b01b0`.

- Source preparation and complete adapted-configuration equality passed:
  **11 hosts retained; all other configuration equal**.
- Merged Caddy validation and reload passed. A real stopped-upstream synthetic
  502 error was observed with normalized HLS URI, absent request/response header
  fields and absent synthetic marker. Raw journal/configuration stays private.
- Conventional HLS enabled using the already-tested image tag
  `my-neko/brave:hls-93f1fa637ae3`; the container was **healthy**, HTTP bound to
  `127.0.0.1:8082`. The adaptive/WebCodecs overlays were retained.
- **17/17** public HTTPS invalid-input/header checks and **2/2** direct cleartext
  bootstrap/media-denial probes passed. **Activate-Exitcode 0** and the helper's
  activation/invalid-input gate passed.

Evidence is in the private `../neko-hls-results-93f1fa637ae3` directory; no
secrets/raw configuration/logs were imported. These supplied target results are
**NOT EXECUTED IN CODEX**. The output proves activation and these denial cases,
not valid credentials, packager output, actual browser playback, spoofing from
an untrusted peer, view-only boundaries, lifecycle, device/numeric/resource/
isolation, remaining logging or rollback execution. Those gates remain open;
dependency maintenance is still pending and full prototype acceptance is not
claimed. LL-HLS remains unadvertised pending its public protocol/RTT gate.

The operator subsequently reported HLS connection failure/no picture on the
first playback attempt. At this initial report, exact UI status, device/browser, login role and server
diagnostics remain unknown; no playback pass or confirmed cause is claimed.
The initial next step was read-only diagnosis from HLS_LL_HLS_VALIDATION.md, then repair and a
repeat of the bounded first five-minute admin-HLS picture/audio and room-event
checkpoint in `HLS_LL_HLS_VALIDATION.md`, followed by its separate full passive
authorization/lifecycle and device/resource matrix. Keep the app checkout and
images at 93f1fa63; documentation-only updates do not require reactivation/build.

### Failed playback/login diagnosis and startup repair — 2026-10-04

The helper from `d191b8ea` completed with Diagnose-Exitcode 0 at application
`93f1fa63` and prepared image `sha256:d217eccd9418`. The supplied summary found
healthy/running, zero Docker restarts, no OOM or sampled Neko process exit,
working loopback metrics, HLS bootstrap `not_ready`/`backend_error` and two
worker-set starts with a `source_restart` transition. No successful readiness,
lease-open or publication was demonstrated. The operator reported normal login
timeouts too, with tentative `/ws` status 101; the post-upgrade blocker is unknown.

[The repair record](HLS_STARTUP_REPAIR_2026-10-04.md) documents cold-generation/
initial-caps handling, HLS-only encoder timestamp mapping, initial-gap prevention,
bounded C logging and the new target-only real-codec integration job. These are
repository changes, not a target pass. The subsequent rollback completed with
Restore-Exitcode 0: configured image my-neko/brave:rollback-hls-93f1fa637ae3
started healthy, and the operator confirmed normal login/picture/audio work again.
The rollback changed image, overlay and process state together; it does not
identify the original blocker or validate the new HLS fixes. The subsequent
exact 80020d99 automated/image gate and default-off repair-image/browser
checkpoint passed as recorded below.

### Exact startup-repair automated/image gate passed — 2026-10-04

The supplied tail of the target run at application
`80020d99477a58318f210b7e14d19cdd92991a6d` ended with
AUTOMATED/IMAGE GATE PASSED and Check-Exitcode 0. The client build completed;
the excerpt begins after the client test details, so their count is not
repeated. All 13 selected Go packages, both 30-second fuzz jobs (983,744
WebSocket executions and 380,301 HLS request executions), and the trailing
server/plugin build passed. With validation-image GStreamer 1.26.2:

- TestEncoderSegmentMapsToRunningTimeWithoutChangingCapture passed (0.02 s);
- TestRealCodecsReachConventionalPackagerReadiness passed (18.11 s), with
  packager startup and conventional readiness both in generation 1;
- every rendition's init and three conventional parents were asserted by
  the real-codec integration test. This is synthetic-clock/fixture evidence,
  not actual production capture, public valid auth or browser playback.

Both uniquely tagged base/Brave image builds completed. The exported config
digests were base sha256:0f94dbd98253ddac78888661f3cfcfdd9727fabba6228765bcee24017995f2f3
and Brave sha256:e633206d1bdae3ff95709fc5a5cc6c0854f13792a9b838789e9d916ca0889c44;
these are build metadata, not newly deployed container observations. The private
images.txt and preparation marker are in ../neko-hls-results-80020d99477a.
The baseline snapshot still describes the running prior rollback image.
Audit exit code 1 remains open dependency findings, not a security pass.
All these results are supplied target evidence, **NOT EXECUTED IN CODEX**.

At that checkpoint the next step was to deploy the prepared repair image with
HLS disabled, preserve the working image under a rollback tag before stopping
the service, and confirm normal login/picture/audio/control. The helper's baseline action
and optional explicit repository argument allow extraction from a reviewed
separate tooling commit without moving the 80020d99 application checkout. It
checks the exact preparation marker and saved image ID; health failure restores
the prior image. Its subsequent target execution and bounded browser smoke
passed as recorded below. No application/image change or repeat build was
required for it. Enabled HLS valid delivery remains the next gate.

### Exact startup-repair default-off deployment/browser smoke passed — 2026-10-04

At unchanged application 80020d99477a58318f210b7e14d19cdd92991a6d, the operator
ran baseline deployment with helper a7669dd184138ba53ea9398d31d0ca706ee0b400
(blob c6f52dc80fdf605ec908f3fe3856ce23e015e494). Baseline-Exitcode was 0;
my-neko/brave:hls-80020d99477a started healthy, and disabled bootstrap/media
returned the expected 404 (2/2 probes). Private evidence and the saved rollback
tag are in ../neko-hls-results-80020d99477a. The helper checked the prepared
image ID before stopping the service; no new runtime image ID was printed.

The operator answered that normal login, picture, audio and control all work in
the requested private browser check. This is bounded target browser evidence,
**NOT EXECUTED IN CODEX**; no wider room-event/device matrix or enabled HLS
playback is claimed. The next step at that checkpoint was same-image HLS
activation and the bounded picture/audio/room-event check; its result is below.
The completed Caddy source merge and application HEAD were retained.
The original normal-login timeout's exact cause remains unconfirmed.

### Exact startup-repair enabled HTTP gate passed, playback failed again — 2026-10-04

At unchanged application 80020d99477a58318f210b7e14d19cdd92991a6d, the reviewed
helper c6f52dc80fdf605ec908f3fe3856ce23e015e494 enabled the same prepared image.
Enable-Exitcode was 0, my-neko/brave:hls-80020d99477a started healthy, all 17
public denial probes and both cleartext-denial probes passed. These checks do
not demonstrate valid bootstrap or playback.

The operator again reported HLS fails with `HLS bootstrap failed; retry
manually`, tentatively recalled a brief initial picture, then reported all
streams stopped afterward. The browser/device, actual process crash versus
stalled media and exact failure cause are unconfirmed. The next step at that
checkpoint was read-only diagnosis before same-image default-off restoration;
both subsequently passed, as recorded below. Application HEAD, Caddy and prior
private evidence were retained without rebuilding or retrying playback.
Review that evidence before further application repairs.

### Second diagnostic/default-off recovery passed; fixed-GOP comparison prepared — 2026-10-04

The supplied 80020d99 diagnostic had exit 0, a running/healthy image
sha256:63d7441365a8ec8628a0abeec8c11f5046ee86962769f9723bf12e276a75c299,
zero Docker restarts/OOM and no Neko exit in the bounded Supervisor sample.
Its 220 application lines included one ready/opened/changed/closed lease,
two sample rejections and two generation starts. Cumulative metrics included
one successful bootstrap and 23 successful segment requests, plus five
not-ready bootstraps and high/medium/low keyframe-admission drops of 201/318/235.
Both fixed rejections were HLS transcode timeline gap. This demonstrates prior
valid server delivery, not full visual/audio acceptance or the all-stream
failure's unique cause. Synthetic denial probes remain part of the counters.

Same-image default-off recovery passed with exit 0, healthy service and 2/2
disabled-route probes; fresh browser confirmation is pending. The next step at
that checkpoint was the isolated old-code/repaired-code comparison against the
pinned existing codec image, keeping the restored application at 80020d99 and
live HLS disabled. It subsequently passed, as recorded below. See
[the repair record](HLS_STARTUP_REPAIR_2026-10-04.md). Full application
checks/build subsequently passed at 97ba4ad9 below; its deployment and live
mixed-backend isolation remain pending.

### Isolated fixed-GOP target comparison passed — 2026-10-05

At unchanged application 80020d99477a58318f210b7e14d19cdd92991a6d, helper
712b491bd5f9900488555051a1fe9acd705f51dc compared repair
97ba4ad9ab3e635da936a58c8a7ec795da05ba46 against pinned codec image
sha256:ed572e4ef4cb4dbbf94b56027f4ac19ac47837a74aebe7a4b14860282a5b4a7f.
The expected negative control returned 1, reproduced two low-rendition timeline
gaps, restarted generations 2/3 and failed readiness after 24.01 seconds.
The positive control returned 0: timestamp mapping passed in 0.01 seconds,
smooth startup in 18.11 seconds and sustained scene cuts in 30.19 seconds,
with all four renditions advancing two more complete parents in generation 1.
GOP-Check-Exitcode was 0. This establishes the scene-cut defect/correction in
the fixture, not production playback or the all-stream failure's unique cause.

These are supplied target results, **NOT EXECUTED IN CODEX**. The live service
was not replaced; HLS remains disabled. Preserve old private reports and the
rollback image. The next step was full checks and uniquely tagged images at
exact 97ba4ad9 in ../neko-hls-results-97ba4ad9ab3e using
validate-hls-phase4.sh. Its full application gate, default-off image deployment
and browser confirmation precede any separate enabled playback block. The
automated/image result is recorded below.

### Exact GOP-repair automated/image preparation passed — 2026-10-05

At clean application 97ba4ad9ab3e635da936a58c8a7ec795da05ba46, the supplied
validation run returned Repair-Check-Exitcode 0 and AUTOMATED/IMAGE GATE PASSED.
All 47 client tests, TypeScript/build, 13 configured Go packages, both 30-second
fuzz jobs and the trailing server/plugin build passed. The fresh codec-image
gate reported GStreamer 1.26.2; all three tests passed, including readiness in
18.11 seconds and scene cuts in generation 1 for 30.19 seconds. Base/Brave
images with tag hls-97ba4ad9ab3e were built successfully, using cached layers.
The running service was not replaced. Private image IDs, marker and evidence
remain in ../neko-hls-results-97ba4ad9ab3e.

Audit exit 1 remains open findings; the client install reported 20 findings,
including one critical. No new reachability/remediation or enabled playback
claim is made. These are supplied target results, **NOT EXECUTED IN CODEX**.
See [the repair record](HLS_STARTUP_REPAIR_2026-10-04.md) for exact provenance.

The next step was tracked deploy-hls-media.sh baseline with the new directory,
then the 2/2 disabled-route probes. Keep exact application 97ba4ad9 and reviewed
Caddy settings. The helper verifies prepared image identity, saves the current
image for rollback and restarts with the new image without HLS, preserving
adaptive/WebCodecs. The deployment subsequently passed, as recorded below;
the operator's subsequent normal browser response is recorded below.

### Exact GOP-repair default-off deployment/browser checkpoint passed — 2026-10-05

At unchanged 97ba4ad9ab3e635da936a58c8a7ec795da05ba46, tracked deployer blob
c6f52dc80fdf605ec908f3fe3856ce23e015e494 completed baseline mode with exit 0.
The prepared my-neko/brave:hls-97ba4ad9ab3e image started healthy and passed
both public disabled-route 404 probes (2/2). Private snapshots/rollback evidence
remain in ../neko-hls-results-97ba4ad9ab3e. These are supplied target results,
**NOT EXECUTED IN CODEX**.

The operator reported the requested fresh normal login/picture/audio/control
check works (qualified "soweit ich denke"). This closes the bounded baseline
browser checkpoint, not the wider role/recovery/device or HLS playback matrix.
The next step was to enable conventional HLS on the same image and repeat its 19 HTTP denial
probes, followed by actual picture/audio and room-event isolation. Enabled HLS
and grouped isolation/lifecycle/device acceptance remain pending. Keep exact
97ba4ad9, the existing Caddy configuration and master.

## NEXT

Continue exclusively on `testing`; do not merge, fast-forward or push changes to `master`. The stable branch remains pinned at `d9105ef8` until the operator explicitly authorizes a later grouped promotion.

**Latest exact-68 client checkpoint — 2026-10-05:** the requested
[HLS client stability review](HLS_CLIENT_STABILITY_REVIEW_2026-10-05.md)
corrected paused-time stall accounting, progress monitoring before first
readiness and mixed HTTP/readiness failure counts, with defensive player-event
handling and fixed bootstrap availability detail. The supplied isolated
target gate passed all 60 client tests, TypeScript and build with
Client-Check-Exitcode 0, fresh index-urce48rY.js and private report
`/opt/docker/nekoNew/neko-hls-client-check-68dbdd4a-lJbw35xa`.
Scoped exact-68 image preparation subsequently passed with
Image-Prepare-Exitcode 0 using helper `28d081a4`: fresh client build, base/Brave
images and private snapshot/marker recorded, with identical backend evidence
inherited. The checkout was then exact-68; the working exact-73 HLS/WebRTC
container remained unchanged during preparation. Activation subsequently
passed with Start-Exitcode 0, healthy `my-neko/brave:hls-68dbdd4a8dd7` and a
private enable snapshot. Checkout/live were then exact application
`68dbdd4a8dd798886302b235c1f8f208452e0c6e`. The operator reports HLS needed
Retry, then worked without problems; the initial error, five-minute/event
interval and concurrent WebRTC result are not separately confirmed.
**NOT EXECUTED IN CODEX; reliable first start and wider acceptance remain open.**
The read-only diagnostic subsequently passed with Diagnostic-Exitcode 0:
healthy service, no sampled exit/OOM, one not-ready and one successful
bootstrap, one startup generation per track and 17 successful segments.
The lease closed and packaging stopped after idle grace; zero current objects
are consistent with that cleanup. The saved histogram summary then passed
with Timing-Exitcode 0 using helper `833cff60`: two requests total
24.001096458 seconds, one at most 1 ms and the other approximately 24 seconds.
This fits the 24-second readiness deadline followed by a warmed Retry, without
per-attempt correlation or a measurement of first media readiness.
The server-only candidate raises conventional startup allowance to 28 seconds
within the existing 30-second HTTP/client limits; LL-HLS, codecs, playlists,
client behavior and the p95 24-second acceptance target stay unchanged.
Exact candidate `8f54970f025e3491a123540cc94870508b66f119` preparation subsequently
passed with Prepare-Exitcode 0. All 13 selected native checks passed, with four
real-codec fixtures in generation 1 (18.09/18.07/18.84/30.07 seconds); candidate
base/Brave images and private snapshot/validation marker were recorded in
`/opt/docker/nekoNew/neko-hls-results-8f54970f025e`. Earlier HLS package/server-
image checks precede the supplied excerpt and are covered by final helper
success without separately shown results. Unchanged client build layers were
cached; the exact-68 client gate is inherited. Target checkout is exact-8f;
the live image was exact-68 at preparation. After the requested activation,
the operator reports connecting then an apparently frozen first picture;
page reload produced moving video/audio while HLS stayed selected. Activation
CLI, initial player state/timing and concurrent WebRTC are not supplied.
Read-only diagnosis subsequently passed with Diagnostic-Exitcode 0: matching
prepared/live exact-8f image IDs, healthy container, two successful bootstraps,
one startup generation per track, all four workers running, one active lease
and 135 successful segments; no sampled exit/OOM or fixed error. These counters
do not isolate the first browser's frozen picture. A client-only start-order
candidate requests existing autoplay immediately after player attachment,
without waiting for canplay. It retains the independent readiness/progress
deadlines and user/private Pause plus muted/manual Play. Exact candidate
`7dcc3c5e4ba0279f35727eb4bed2e86bf721ad3b` preparation subsequently passed with
Prepare-Exitcode 0 using `deploy/prepare-hls-client-start.sh`: both images,
private snapshot and validation marker recorded in
`/opt/docker/nekoNew/neko-hls-results-7dcc3c5e4ba0`, with the live exact-8f
container retained. The supplied tail starts inside the base-image build;
earlier isolated old/new client tests/type/build are covered by final helper
success, without separately visible counts. Backend evidence is inherited
from unchanged exact-8f sources, not freshly rerun. Target checkout is now
exact-7dcc3c5e. After the requested activation, the operator reports moving
first picture after approximately 20 seconds, then a frozen picture requiring
reload. The subsequent progress diagnostic passed with Diagnostic-Exitcode 0,
confirming the prepared/live exact-7dcc image ID, healthy service and valid
12.012-second deltas. HLS was idle in both samples: zero leases/subscriptions/
workers and no new publication/HTTP/generation events. The operator confirmed
HLS was closed/logged out; this sample does not capture the freeze. Earlier
lease closures/idle stop fit cleanup. A WebRTC join/video start remains an
uncorrelated hypothesis.
Static review found no direct join-triggered capture rebuild/HLS teardown while
the HLS listeners remain attached, or client join-triggered source overwrite.
The corrected `4593a6f9` progress helper subsequently passed with Diagnostic-
Exitcode 0 and valid 12.011-second deltas at the same healthy exact-7dcc image.
One active lease, audio/high subscriptions and all four workers persisted;
601 audio/300 video units flowed, with parts/segments published for every track
and successful HTTP media requests. No interval generation/discontinuity/drop/
capture-creation/WebRTC-open increment; VP8 capture rows are now retained.
These are supplied target results, NOT EXECUTED IN CODEX; the corrected
saved-startup helper has no fresh execution. The operator confirmed this sample
was taken during frozen picture/audio loss without reload; the page later failed
with "HLS HTTP connection failed; retry manually". Server flow during the freeze
does not identify its cause, and the later HTTP failure was outside this sample.
The supplied retained browser timing summary shows one 1,001-ms master without
response metadata, fitting the 1-second deadline. Its full 250-entry buffer ends
about nine minutes before the query, omitting the later failure sequence. Other
status-zero entries have body data; terminal cleanup already cleared the video.
Neither three timed-out probes nor the freeze cause is proved.
The full bounded five-minute trace is now supplied: playback stops at media time
23.948 seconds, frames remain 513 and the buffer stays at 12.008–24.019 seconds,
despite continued medium-video loading. Audio loads four segments; later six-second
seek jumps hide the stall from the old watchdog. No new terminal HTTP failure
appears after Retry. This is captured failure, not successful five-minute playback.
Exact-8741 repair preparation then passed with Prepare-Exitcode 0: both old
defects reproduced in isolated copies; all 66 client tests/type/build, repaired
all-track/both-mode wire check, HLS package and server/base/Brave builds passed.
Target checkout/live deployment is now
`8741f7880d9a709e6dc17924d2af0564a949ccd9`: activation passed with
Activate-Exitcode 0, a healthy candidate image and private enable snapshot.
Evidence: `/opt/docker/nekoNew/neko-hls-results-8741f7880d9a`.
The operator confirmed at least five minutes of moving HLS picture/audio without
Retry/reload while WebRTC continued working. The bounded PC/Helium sustained
playback gate is PASSED. NEXT consolidate [remaining grouped acceptance](HLS_PLAYLIST_WINDOW_REPAIR_2026-10-07.md#remaining-acceptance);
no additional server command or ad-hoc check is requested at this checkpoint.
No rebuild, later-doc pull, old-image trace, repeated five-minute gate or timeout
tuning is needed. These are supplied target/browser results, NOT EXECUTED IN CODEX.
The uncaptured hls.js parser error as the old audio-stop cause remains an inference
despite controlled reproduction and the positive repaired run. The earlier
WebRTC-join hypothesis remains uncorrelated. Repeated cold starts, room actions,
authorization/lifecycle and grouped device/resource/dependency acceptance stay open.
No repeated unchanged native/fuzz or HTTP-denial gate is required for this
bounded playlist/watchdog change. Prior client/image checks apply to their commits;
the new production delta's focused exact-8741 preparation has now passed.
Do not repeat passed client/image/HTTP-denial gates without a new reason.
Historical results below apply only to their recorded commits.

Continue **Phase 4 target-server validation** by [consolidating the remaining startup/frozen-picture/room-event investigation](HLS_CLIENT_READINESS_REPAIR_2026-10-05.md#next-consolidate-startup-frozen-picture-and-room-event-investigation). The exact-73 read-only diagnosis passed with Diagnostic-Exitcode 0, a healthy active HLS service, six bootstrap successes, 394 successful segments and one not-ready bootstrap; the browser failure cause remains uncorrelated. The operator could not confirm the requested interval and reports possible random reconnects/room actions. No additional ad-hoc operator check is requested now. The isolated client gate passed with old timer-fault reproduction and all 52 tests/type/build. Scoped exact-73d5ff6d image preparation then passed with Client-Image-Exitcode 0: the client rebuilt as index-CrHQRMnq.js, unchanged server/runtime layers were cached, and both images/private snapshot/marker were recorded using separately inherited a7ff backend evidence. Exact-73 default-off deployment then passed with Baseline-Exitcode 0, a healthy service, private baseline snapshot and 2/2 disabled-route probes; the operator confirmed the requested normal browser check works. Same-image activation then passed healthy with Enable-Exitcode 0, a private enable snapshot and 19/19 denial probes. On PC/Helium the operator reports an initial Retry, one frozen picture requiring reload, then working HLS with WebRTC unaffected; a HLS failed message was confirmed without its detail. Checkout/live image is now 73d5ff6d with conventional HLS enabled. Preserve the working service and supplied diagnostic evidence; first-start/recovery, a documented uninterrupted room-event interval and grouped acceptance remain open. Consolidate further checks into one bounded later step with a clear action/result rather than additional ad-hoc operator requests. Preserve the working a7ff/71 images and old evidence directories. Exact repair-image preparation and subsequent helper `2484a022` activation passed at application `93f1fa63`: all 11 Caddy hosts preserved, merged validation/reload and synthetic runtime-error redaction passed, HLS image healthy, 17 public plus two cleartext-denial probes passed. The first actual playback attempt failed, followed by normal-login timeouts (tentative /ws 101). Read-only diagnosis from helper d191b8ea passed: prepared image healthy, no sampled process exit/OOM, HLS startup/source-restart but no demonstrated readiness. [Startup/generation/timestamp/log-bound repairs](HLS_STARTUP_REPAIR_2026-10-04.md) and a mandatory real-codec integration gate are implemented; their target automated/image gate passed at 80020d99, including both real-codec tests and all-four-rendition conventional readiness in one generation (test duration 18.11 seconds); the original login blocker remains unconfirmed. The saved pre-HLS image was restored healthy with Restore-Exitcode 0 and operator-confirmed normal login/picture/audio. Preserve evidence. Default-off deployment passed with Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login/picture/audio/control. Same-image HLS activation passed with Enable-Exitcode 0, healthy service and 19/19 HTTP denial probes, but HLS failed and the operator reported all streams stopped afterward. Read-only diagnosis found one ready packager/lease, 23 successful segment requests and two timeline-gap rejections, with no sampled Neko exit/OOM. Same-image default-off restoration passed with Recovery-Exitcode 0 and 2/2 disabled-route probes; fresh browser confirmation is pending. The HLS-only fixed-GOP correction passed the isolated target GOP A/B gate at 97ba4ad9: old code reproduced two timeline gaps and all three repaired codec tests passed, including 30.19 seconds of scene cuts in generation 1. Exact 97ba4ad9 automated/image preparation passed with Repair-Check-Exitcode 0 (47 client tests, 13 Go packages, both fuzz jobs, all three codec tests and base/Brave builds). Default-off deployment of my-neko/brave:hls-97ba4ad9ab3e passed with Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes. The operator reported the requested normal browser check works. The subsequent HLS attempt at the prepared 97ba4ad9 checkpoint failed with "HLS bootstrap failed; retry manually", and the operator confirmed WebRTC also stopped working. The latest enablement CLI/HTTP results have not been supplied. Read-only diagnosis passed with 84 log lines, one not-ready bootstrap, one negotiation rejection and no sampled exit/OOM or generation/lease-open markers. Same-image default-off restoration passed with Recovery-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login/picture/audio work again. The saved-startup summary passed with Saved-Check-Exitcode 0: audio/high/medium subscriptions persisted, low capture reached only its create marker, and the saved matching environment had HLS enabled. Paired startup diagnosis at helper 88f2b25d completed with exit 0: both constructors passed two starts and failed one, with medium losing its initial pre-anchor IDR and reaching only two parents. This narrows the smooth readiness defect independently of registry isolation. The 414639d2 isolated anchor A/B gate passed with Anchor-Check-Exitcode 0: old code reproduced the fixed initial-IDR failure; seven corrected checks each passed three times, with every real-codec fixture staying in generation 1. Full exact-414639d2 preparation FAILED with Repair-Prepare-Exitcode 1: 47 client tests, type/build, 13 Go packages, both fuzz jobs, registry/mapping and anchor lifecycle checks passed, but smooth readiness restarted with worker_failure and failed its generation-1 assertion at 20.03 seconds; scene cuts passed at 30.19 seconds in generation 1. Base/Brave image steps were not reached. Worker diagnosis at helper 53034495 then completed with exit 0: all six checks passed in three fresh processes, each smooth fixture ready in 18.11 seconds in generation 1 with no rejected pushes; the earlier worker failure was not reproduced. The controlled audio-anchor A/B at repair 71a14d21 passed with Audio-Anchor-Exitcode 0: the old AAC hold reproduced both the blocked-drainage unit failure and an audio/anchor/queue_full restart under 256 ms high-input delay; all eight corrected startup checks passed in three fresh processes and scene cuts passed once (25 top-level passes). All seven real-codec fixtures stayed in generation 1; normal readiness was 18.10/18.11/18.11 seconds and delayed-high readiness 18.09 seconds each. This verifies the controlled AAC-overflow mechanism, not the cause of the earlier unobserved worker failure or live all-stream outage. Full exact-71a14d21 preparation then passed with Repair-Prepare-Exitcode 0: the rebuilt GStreamer 1.26.2 image passed all nine selected startup checks, including normal readiness at 18.11 seconds, delayed-high readiness at 18.09 seconds and scene cuts at 30.19 seconds, all in generation 1; base/Brave images were built and the service remained unchanged. The supplied tail starts inside the codec-image build; earlier client/Go/fuzz steps are covered by the script's reported final success, not separately shown in this excerpt. Default-off deployment of my-neko/brave:hls-71a14d2174da then passed with Baseline-Exitcode 0, a healthy service and 2/2 disabled-route probes. The operator confirmed the requested normal browser check works without HLS at 71a14d21. Same-image conventional-HLS activation at 71a14d21 then passed with Enable-Exitcode 0, healthy service, a private enable snapshot and 17/17 public plus 2/2 cleartext-denial probes. The subsequent operator-reported HLS attempt at 71a14d21 went from connecting to failed with "HLS bootstrap failed; retry manually"; the operator reported only WebRTC streaming works. No successful HLS picture/audio or room-event interval is demonstrated. Read-only diagnosis passed with 950 captured lines, one not-ready bootstrap, one started/idle-stopped packager generation, medium/low keyframe-admission drops of 759/564 and cumulative part/segment publication only for audio/high. No sampled Neko exit/OOM was found. Same-image default-off restoration passed with Recovery-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login/picture/audio work again. The isolated source-clock-phase diagnosis at helper 409482b4 passed with Clock-Skew-Exitcode 0: aligned control ready in 18.11 seconds; all three skew runs reproduced not-ready at 24.02 seconds in generation 1 with audio/high ready, medium/low blocked, flowing IDRs and no rejected native pushes. This proves the controlled phase-admission defect, not the exact unmeasured live phases. A common high-source fan-out repair is implemented and statically reviewed: one shared video provider subscription feeds the existing three scaled encoders, plus one audio subscription; provider PTS/DTS are preserved. The isolated common-source repair A/B passed at a7ffb8b1 with Shared-Clock-Exitcode 0: the old defect reproduced once; 46 positive top-level checks passed across three cold processes including one scene-cut check, with all ten real-codec fixtures in generation 1. Full exact-a7ffb8b1 preparation then passed with Repair-Prepare-Exitcode 0: all thirteen selected native/startup checks passed in the rebuilt GStreamer 1.26.2 image, including the four real-codec fixtures in generation 1, and base/Brave images were built without changing the running service. The supplied excerpt starts inside the codec-image build; earlier client/Go/fuzz stages are covered by final script success without separately shown fresh counts. Default-off deployment of my-neko/brave:hls-a7ffb8b13448 then passed with Baseline-Exitcode 0, healthy service, a private baseline snapshot and 2/2 disabled-route probes; the operator reported the requested normal browser check works. Same-image a7ffb8b1 HLS activation passed with Enable-Exitcode 0, healthy service and 19/19 HTTP denial probes. The operator then reported first HLS picture and the compact streaming label, followed after roughly 30 seconds by "HLS playback did not become ready; retry manually"; Retry HLS restored playback and WebRTC continued working. Static inspection found the initial client readiness deadline was never disarmed on canplay/playing. Client-only repair 73d5ff6d cancels it on those current-player events, arms it before attachment and retains the independent startup/stall bounds; Read-only target diagnosis then passed with a healthy a7ff image, two successful HLS bootstraps and 486 successful segment requests, with no sampled process exit/OOM; the packager stopped after idle grace. The isolated target client gate passed with Client-Check-Exitcode 0: the old timer defect reproduced, all 52 repaired client tests plus type/build passed, and checkout/live service remained at a7ff. These are supplied target results, NOT EXECUTED IN CODEX. Scoped exact-73d5ff6d image preparation then passed with Client-Image-Exitcode 0: fresh client bundle index-CrHQRMnq.js, cached unchanged server/runtime layers, both base/Brave images and private snapshot/marker recorded while enabled a7ff stayed running. Exact-73d5ff6d default-off deployment then passed with Baseline-Exitcode 0, healthy my-neko/brave:hls-73d5ff6d2911, a private baseline snapshot and 2/2 disabled-route probes; the operator confirmed the requested normal browser check works without a media override. Same-image conventional-HLS activation then passed with Enable-Exitcode 0, healthy service, a private enable snapshot and 19/19 denial probes. The operator reports PC/Helium HLS playback after an initial Retry and one frozen-picture/page-reload incident; WebRTC kept working and later HLS worked normally. A HLS failed message was confirmed without its detailed error or exact timing, so startup reliability and an uninterrupted room-event interval remain unverified. Checkout/live image is now 73d5ff6d with conventional HLS enabled. Read-only exact-73 diagnosis then passed with Diagnostic-Exitcode 0: healthy service, no sampled exit/OOM or fixed error markers, one active HLS lease/all four workers running, six successful bootstraps and 394 successful segment requests. One not-ready bootstrap supports readiness as a possible initial-Retry explanation without attempt correlation; two startup-labelled generations and one idle stop do not establish a crash loop. The operator cannot confirm the exact uninterrupted interval and mentions possible random reconnects/room actions without correlation. NEXT consolidate startup/frozen-picture/room-event investigation into one bounded later validation step; no further ad-hoc operator check requested at this checkpoint; startup/recovery and grouped acceptance pending; picture/audio/room-event, passive authorization/lifecycle and grouped device/resource acceptance remain open. The original blocker remains unconfirmed; rollback recovery does not validate the new repair or HLS playback. The [audit](DEPENDENCY_AUDIT_2026-10-04.md) and [static review](STABILITY_REVIEW_2026-10-04.md) record the repairs and limits; package remediation and final security/live acceptance stay open. Keep HLS default-off without its overlay, WebRTC the default, WebCodecs explicit and fallback manual. The unavailable colleague's television remains an open device gate. `master` must not move without explicit operator authorization.

## Product priority after stable synced baseline

1. **bounded checkpoint closed with the documented final-matrix limitation:** media-subscription/WebRTC compatibility refactor plus estimator startup correction;
2. **bounded checkpoint closed with explicit final-matrix limitations:** WebCodecs plus dedicated media WebSocket receive path;
3. **implemented / focused target checkpoint closed with the live-fragment limitation:** persisted explicit per-client selection in sidebar settings with a compact backend/status indicator, WebRTC default and diagnostic URL override;
4. **completed design:** exact default-off HLS/LL-HLS passive/view-only contract in [`HLS_LL_HLS.md`](HLS_LL_HLS.md);
5. **focused target checkpoint closed:** exact `ddf15cee` tests/build/deployment plus two-viewer bounded application-limited recovery, building on the healthy hold, real downgrade and isolation evidence from `2efcc6b1`;
6. **implemented / automated/image and activation/invalid-input gates passed:** HLS/LL-HLS Phases 1–3 plus Phase 4 assets; exact application `93f1fa63` with helper `2484a022` deployed healthy after preserving Caddy/logging gates and passed 19 HTTP probes; valid playback/device acceptance remains pending;
7. **isolated GOP A/B, exact 97ba4ad9 preparation/default-off checkpoint passed / live attempt recovered / diagnosis complete / anchor repair A/B passed:** helper 88f2b25d localized medium's initial pre-anchor IDR loss in both constructors. At 414639d2 the old defect reproduced and all seven corrected checks passed three repetitions, including all six real-codec fixtures in generation 1 (Anchor-Check-Exitcode 0). Full exact-414639d2 preparation then FAILED: smooth readiness restarted into generation 2 (worker_failure), while scene cuts passed in generation 1. The bounded worker diagnosis then passed all six checks in three fresh processes (18.11 s, generation 1 each) without reproducing the failure. The controlled audio-anchor A/B at 71a14d21 then passed: both old-code failures reproduced, all eight corrected startup checks passed three fresh processes and scene cuts passed once (25 top-level passes; all seven codec fixtures in generation 1). Full exact-71a14d21 preparation then passed with Repair-Prepare-Exitcode 0: all nine selected startup checks passed in the rebuilt codec image and base/Brave images were built; the service stayed unchanged. Default-off 71 deployment then passed with Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes. The operator confirmed normal browser checks work without HLS. Same-image HLS activation then passed with Enable-Exitcode 0, healthy service and 19/19 denial probes. The subsequent HLS attempt failed at bootstrap; the operator reports WebRTC streaming works. Read-only diagnosis and same-image default-off restoration passed; the operator confirmed normal login/picture/audio. The isolated source-phase diagnosis reproduced the defect three times with its aligned control passing. The common high-source fan-out repair A/B passed at a7ffb8b1 with Shared-Clock-Exitcode 0: one old defect reproduction and 46 positive top-level checks, including three skew fixtures and one scene-cut check, all ten codec fixtures in generation 1. Full exact-a7ffb8b1 tests/image preparation then passed with thirteen selected native/startup checks and base/Brave builds. Default-off exact-a7ffb8b1 deployment then passed healthy with Baseline-Exitcode 0 and 2/2 disabled probes; the operator reported the requested normal browser check works. Same-image a7ff activation then passed with Enable-Exitcode 0 and 19/19 denial probes. First HLS picture/streaming was reported, followed around 30 seconds by the initial-readiness error; Retry HLS restored playback and WebRTC stayed working. Client-only repair 73d5ff6d disarms that startup deadline on readiness. Read-only diagnosis and the isolated client gate then passed with Client-Check-Exitcode 0, old-fault reproduction and all 52 tests/type/build. Scoped exact-73d5ff6d image preparation then passed with Client-Image-Exitcode 0, fresh client build and cached unchanged server/runtime layers, leaving enabled a7ff running. Default-off exact-73 deployment then passed healthy with Baseline-Exitcode 0, 2/2 disabled probes and the reported normal-browser checkpoint. Same-image activation then passed healthy with Enable-Exitcode 0 and 19/19 denial probes; PC/Helium playback was reported after Retry and one frozen-picture reload, with WebRTC unaffected. Read-only diagnosis then passed with a healthy active service, six bootstrap successes, 394 successful segments and one not-ready bootstrap. NEXT consolidate remaining startup/frozen-picture/room-event investigation while retaining the working images/evidence; no further ad-hoc operator check requested now. Sustained playback, lifecycle/grouped acceptance and package remediation remain open;
8. promote accumulated `testing` history only after an explicit operator decision at a coherent validation milestone.

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
10. **implemented / focused automated target gate passed — HLS Phases 1–2:** default-off access, leases, security, deterministic models, shared H.264/AAC packaging, authenticated HTTP delivery and observability; enabled runtime acceptance pending;
11. **implemented HLS Phase 3 / focused automation and default-off image deployment passed:** isolated passive client, pinned player support and advertised-only manual HLS/LL-HLS selection; enabled playback/device acceptance pending;
12. **repository assets implemented / exact-71a14d21 preparation passed / default-off deployment/browser checkpoint passed / same-image activation/19 denial probes passed / live HLS bootstrap failed / diagnosis and default-off recovery passed / source-phase defect reproduced / common-source repair A/B, full exact-a7ffb8b1 preparation/default-off checkpoint and activation passed / first picture then client deadline failure / client gate and scoped repair images passed / exact-73 default-off deployment/browser passed / same-image activation/19 probes passed / playback after Retry/reload reported / read-only diagnosis passed / NEXT consolidated startup/recovery investigation — HLS Phase 4:** prior live failures were restored; the initial-video-IDR correction passed its isolated gate, but full 414 preparation failed after a worker restart. The narrower video-only hold at 71a14d21 passed the controlled AAC-overflow A/B with both old-code failures reproduced and 25 positive top-level passes across three fresh processes. Full exact-71a14d21 preparation passed with nine selected startup checks and base/Brave builds while the default-off 97ba4ad9 service stayed running. Default-off 71 deployment passed with Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes. The operator confirmed the requested normal browser check works. Conventional HLS activation in that same image passed with Enable-Exitcode 0, healthy service and 19/19 denial probes. The live HLS attempt then failed with bootstrap failure; the operator reports WebRTC works. The read-only diagnosis found medium/low admission drops with publication only for audio/high; same-image default-off restoration and normal browser behavior passed. The source-phase diagnosis reproduced the defect in all three runs while its aligned control passed. The common-source repair A/B passed with 46 positive top-level checks and the old defect reproduced once. Full exact-a7ffb8b1 tests/images then passed with thirteen selected native/startup checks and base/Brave builds, leaving the default-off 71 service unchanged. Default-off exact-repair deployment then passed healthy with Baseline-Exitcode 0 and 2/2 disabled probes; the operator reported the requested normal browser check works. Same-image activation subsequently passed with Enable-Exitcode 0 and 19/19 denial probes. First HLS picture/compact streaming was reported, then the initial-readiness timeout around 30 seconds; Retry restored HLS and WebRTC kept working. Client-only repair 73d5ff6d cancels the obsolete startup deadline on readiness. The supplied read-only diagnosis and isolated client gate then passed with old-fault reproduction and all 52 repaired tests/type/build (Client-Check-Exitcode 0). Scoped exact-73d5ff6d image preparation then passed with Client-Image-Exitcode 0: the client rebuilt, unchanged server/runtime layers were cached, both images/private snapshot/marker were recorded, and enabled a7ff stayed running. Default-off exact-73 deployment then passed with Baseline-Exitcode 0, healthy service, 2/2 disabled probes and the reported normal-browser checkpoint. Same-image activation then passed with Enable-Exitcode 0, healthy service and 19/19 denial probes. PC/Helium HLS worked after Retry and one frozen-picture reload while WebRTC kept working; the detailed HLS failed error is missing. Read-only exact-73 diagnosis then passed healthy with one active HLS lease/four running workers, six bootstrap successes, 394 successful segments and one not-ready bootstrap. The operator cannot confirm the requested interval and mentions possible random reconnects/room actions without correlation. Consolidate the remaining investigation into one bounded later step, retaining the working service/images/evidence; no more ad-hoc operator checks now. Startup/recovery reliability and the uninterrupted room-event gate remain open. Sustained repaired playback and grouped live validation remain pending;
13. compare device support, failure behavior, server resource cost, latency and recovery before defining any automatic capability-based selection;
14. evaluate WebTransport only afterward if WebSocket's delivery/backpressure characteristics are a demonstrated limitation.

The passive path may trade latency for reliability and compatibility. It must stay in the same logical room and must not gain control authorization. HLS/LL-HLS now has the specified contract, server/passive-client delivery through Phase 3 and separate Phase 4 repository assets. Earlier focused automation and default-off image deployment passed; the exact startup-repair automated/image and real-codec gates passed at 80020d99. Default-off repair-image deployment and normal login/picture/audio/control passed. Enabled HLS runtime validation and device evidence remain pending.

## LATER / OPTIONAL

- WebTransport productionization;
- fully automatic transport selection after explicit/manual fallback paths are proven;
- automatic codec selection;
- MPEG-DASH as an additional passive HTTP-streaming backend where useful;
- MJPEG only as an ultra-legacy image-only fallback for a concrete unsupported device class;
- broader v3 client/library rewrite.

## OPEN

- Final grouped validation must repeat the ordinary/admin/view-only/private-mode/manual-tier/reconnect matrix and the independently constrained three-viewer adaptive down/up isolation run; neither was rerun at `e5f55bf9`.
- The first steady-state candidate at `2d037f39` is rejected. The receiver-evidence downgrade revision at `2efcc6b1` passed tests/build/deployment, the 20-minute healthy hold, real constrained downgrade and peer isolation, but its application-limited C recovery exceeded 90 seconds. Bounded recovery follow-up `ddf15cee` passed its exact-commit server tests/build/deployment and compact two-viewer gate, returning C `low -> medium -> high` in 60 seconds while H remained `high` with zero video-drop delta.
- Determine whether fast-motion softness is acceptable at the current 1,996,800-bit/s VP8 `high` tier through a controlled same-content bitrate/quantizer A/B with receiver statistics and comparable captures; current evidence does not identify a subscription-refactor regression.
- Supported Smart-TV/device matrix, including native HLS, MSE/DASH and WebCodecs capability.
- Whether the target iPhone validates the implemented same-peer and replacement-session paths without reload; a Safari Play gesture remains an explicitly separate, permitted policy fallback.
- Target-device and target-server evidence for the specified VP8/Opus WebCodecs/media-WebSocket contract; framing, queue sizes, synchronization, security limits and rollout behavior are fixed in [`WEBCODECS_MEDIA_WEBSOCKET.md`](WEBCODECS_MEDIA_WEBSOCKET.md).
- Final-commit grouped tests/build plus independent generated playlist/fMP4 inspection for the HLS/LL-HLS Phase 1–3 path; focused automation passed at `e85d8568` and the default-off image deployment started healthy at `741025c3`.
- Actual HLS/LL-HLS device, latency and resource evidence against the fixed targets in [`HLS_LL_HLS.md`](HLS_LL_HLS.md); the HTTP server and passive client are implemented; deployment assets and runtime/device acceptance remain pending.
- Whether DASH adds meaningful compatibility beyond HLS for the actual target devices.
- Eventual automatic per-client media-backend selection rules after the explicitly selected prototypes have measured evidence; version-1 manual selection and rollback are already fixed.
- Live compact view-only fragment preservation for the productized selector was not repeated at `12cfe43b`; focused automated coverage passed, and earlier view-only boundary/browser evidence remains separate.
- Longer-term legacy Vue 2 migration.
