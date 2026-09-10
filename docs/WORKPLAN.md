# Work Plan / Handoff State

Last consolidated: 2026-09-10.

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
following work: deployment reconciliation, repository knowledge and the opt-in adaptive-quality unit
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

## IN PROGRESS — adaptive-quality target-server validation

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
- The next opt-in candidate raises each VP8 `max-quantizer` from 20/24/28 to 56 so CBR can trade image quality for the configured rate. Target bitrates, resolutions, frame rates and all stable base-Compose behavior remain unchanged.

Pending before acceptance: deploy the encoder candidate, rerun the affected 1.3 Mbit/s and 0.7 Mbit/s phases plus final recovery, capture resources/logs, and complete refresh/rejoin plus transient-interruption checks. The stable base Compose remains unchanged.

## Target-server verification

This phase is performed by the operator on the real server, not by Codex.

### Client

```bash
cd client
npm ci
npm run lint
npm run build
```

Required for the integrated baseline: confirm `npm ci`, TypeScript lint and the Vite production build at integration commit `4e99b8d3`. The operator later confirmed these checks passed, closing verification of the lock/type repair in `2d89027e`.

### Server and container

```bash
cd server
./build
```

or:

```bash
docker build ./server
```

From the repository root, build the exact local base and Brave images referenced by Compose:

```bash
./build my-neko/base:latest -y
./build my-neko/brave:latest -y
```

Then prepare the ignored deployment environment and inspect the fully resolved Compose before startup:

```bash
cp .env.example .env
# Set both passwords and any host-specific values in .env.
docker compose config
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

For the new tracked opt-in profile, use the exact phases, metrics and acceptance criteria in [`ADAPTIVE_QUALITY.md`](ADAPTIVE_QUALITY.md). That procedure supersedes the abbreviated adaptive-quality bullets above for the current `NEXT`.

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

Build the current Brave image on the real target server, activate `docker-compose.adaptive.yaml`, and execute the complete acceptance procedure in [`ADAPTIVE_QUALITY.md`](ADAPTIVE_QUALITY.md). Archive the four-phase metric/result table, relevant estimator logs, host resource measurements, commit and image ID.

Accept or reject the profile from that evidence. If thresholds, bitrates, frame rates or encoder thread counts need adjustment, change only the opt-in YAML and repeat the affected phases plus final recovery. Do not change the stable base Compose default, and do not call the adaptive tuning validated before this target-server run passes.

## Product priority after stable synced baseline

1. validate the reproducible slow-client isolation and multi-pipeline estimator profile on the target server;
2. tune per-viewer quality using the recorded target-server measurements;
3. stabilize mobile/reconnect after sync;
4. server-enforced view-only sharing;
5. practical non-WebRTC viewer-media fallback.

## Fallback prototype sequence

Alternative media work starts only after the current adaptive-quality `NEXT` is accepted or explicitly rejected from target-server evidence.

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
- Whether upstream adaptive quality meets desired per-viewer behavior.
- Exact WebCodecs/WebSocket framing, queue/drop/backpressure policy and codec set.
- HLS/LL-HLS latency target, segment/part sizing, codec profile and server resource cost.
- Whether DASH adds meaningful compatibility beyond HLS for the actual target devices.
- Exact per-client media-backend capability/selection rules and rollout order.
- View-only token format/lifetime/revocation.
- Longer-term legacy Vue 2 migration.
