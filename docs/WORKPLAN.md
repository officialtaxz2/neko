# Work Plan / Handoff State

Last consolidated: 2026-09-09.

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
following change: repository-knowledge update only
```

## COMPLETED — `MyNekoProjekt` local delta audit

Completed on 2026-09-09 against fork baseline `d9c1afd564ad4c293a0b1b28aecbc103d1d1d9c6`.

- Safety branch `safety-pre-local-delta-audit-20260909` preserves the original fork HEAD.
- A clean baseline and the complete supplied local tree were compared recursively, excluding `.git/`.
- All 724 baseline files exist locally: 138 are byte-identical, 585 differ only by CRLF/LF, and only `docker-compose.yaml` differs substantively.
- The 3,929 local-only files are 3,928 persistent browser-profile/runtime files under `files/**` plus an empty instance `policy.json`; `downloads/` is empty.
- No source feature/fix or reusable repository config was missing, so no local source/config content was imported.
- The local compose override and embedded credentials were excluded. Active credentials from that artifact should be rotated operationally.
- Root ignore rules guard the local reference and runtime/profile paths against accidental staging.
- Static follow-up repaired the stale pre-Vite client lock/type baseline in commit `2d89027e`; target-server verification is pending.

Status: **Codex-side implementation complete / target-server verification pending**.

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
- Kept `MyNekoProjekt`, credentials, browser profiles, downloads, lock files, runtime state and instance-only policy/compose data out of the merge.

Static inspection completed:

- no unmerged files or conflict markers;
- `git diff --check` clean after removing one upstream blank line at EOF;
- changed JSON files parse successfully;
- client/server event names, file capability payloads, store registration, H.265 configuration paths and removed `ArrayIn` references cross-checked;
- no staged forbidden runtime paths or high-confidence private-key/token patterns.

Runtime/build/test status: **NOT EXECUTED IN CODEX**.

The detailed merge record is in [`UPSTREAM_SYNC_AUDIT.md`](UPSTREAM_SYNC_AUDIT.md).

## Target-server verification

This phase is performed by the operator on the real server, not by Codex.

### Client

```bash
cd client
npm ci
npm run lint
npm run build
```

Required for the integrated baseline: confirm `npm ci`, TypeScript lint and the Vite production build at integration commit `4e99b8d3`. This also closes the still-pending verification of the lock/type repair in `2d89027e`.

### Server and container

```bash
cd server
./build
```

or:

```bash
docker build ./server
```

For the deployed browser image, also build the actual root/application-image path used in production. The upstream merge changes the base image, runtime dependencies and several browser Dockerfiles; a server-only binary build does not verify those layers.

Before starting the deployment, confirm that intentional configuration is supplied by environment variables or a mounted YAML file. The base image no longer embeds root `config.yml`; implicit hosting and cookie authentication now default to disabled. Replace the example compose file's `NEKO_WEBRTC_NAT1TO1` placeholder and never reuse its example passwords in production.

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

The semantic upstream synchronization is complete in Codex. The next bounded unit is target-server validation of the source tree introduced by `4e99b8d3` on `integration/upstream-20260909` using the commands and regression matrix above. The following commit on that branch updates only repository knowledge.

If the target-server checks pass, promote the exact reviewed integration commit without rewriting it:

```bash
git switch master
git merge --ff-only integration/upstream-20260909
```

Record the target-server results here before starting product work. If validation fails, fix and re-review on the integration branch; do not promote a partially verified tree.

## Product priority after stable synced baseline

1. validate and tune the integrated slow-client isolation path;
2. evaluate the integrated multi-pipeline + bandwidth estimator for per-viewer quality;
3. stabilize mobile/reconnect after sync;
4. server-enforced view-only sharing;
5. practical non-WebRTC viewer-media fallback.

## Fallback prototype sequence

If alternative media work begins:

1. media-subscription abstraction;
2. WebCodecs + dedicated WebSocket media;
3. WebTransport only later.

## LATER / OPTIONAL

- WebTransport productionization;
- automatic transport selection;
- automatic codec selection;
- HLS/DASH fallback if justified;
- broader v3 client/library rewrite.

## OPEN

- Supported Smart-TV/device matrix.
- Whether upstream adaptive quality meets desired per-viewer behavior.
- Best first non-WebRTC media backend after measurement.
- View-only token format/lifetime/revocation.
- Longer-term legacy Vue 2 migration.
