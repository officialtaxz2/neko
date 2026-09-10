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
- Demo-mode client infrastructure.
- Per-peer non-blocking media sample delivery so one backpressured WebRTC track does not block dispatch to other tracks.
- Multi-pipeline stream-selection and per-peer bandwidth-estimator infrastructure; the legacy protocol path requests automatic selection when configured.
- An explicitly opt-in Brave Compose overlay with ordered `high`/`medium`/`low` VP8 pipelines and explicit estimator settings; the base Compose deployment remains single-pipeline.
- Peer-local audio/video sample-drop counters plus measured pipeline-bitrate metrics for adaptive-quality diagnosis.
- Stream bitrate accounting in bit/s, matching the Pion estimator rather than comparing its bit/s target with encoded bytes/s.
- Separate downgrade and upgrade headroom thresholds so approximately 2:1 quality tiers do not reuse an unsafe current-tier upgrade margin; the compatible server default remains unchanged.
- Optional per-pipeline nominal bitrate metadata so an upgrade can be gated against the next tier's capacity requirement instead of a content-dependent current-tier measurement.
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
- `master` was fast-forwarded to the reviewed integration history; subsequent `testing` work includes the sanitized Brave deployment reconciliation, the accepted opt-in adaptive-quality unit and the bounded iOS recovery block pending target-device validation.

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

The bounded iOS recovery path is implemented and statically reviewed on `testing`; its exact no-reload device procedure remains pending for the next grouped target-server checkpoint. No automatic in-place recovery claim is made until that run passes.

## NEXT

Continue exclusively on `testing`: implement server-enforced view-only sharing without weakening the existing control/admin model or coupling authorization to a media transport. Keep the current iOS recovery block on `testing` and validate it with the grouped checkpoint defined in [`docs/IOS_RECOVERY.md`](docs/IOS_RECOVERY.md). `master` remains pinned at the accepted stable baseline until the operator explicitly authorizes a later grouped promotion.

See [`docs/ADAPTIVE_QUALITY.md`](docs/ADAPTIVE_QUALITY.md), [`docs/IOS_RECOVERY.md`](docs/IOS_RECOVERY.md), `docs/WORKPLAN.md` and `docs/UPSTREAM_SYNC_AUDIT.md`.

## Local Brave deployment

The tracked `docker-compose.yaml` is the sanitized deployment structure from `MyNekoProjekt`. It uses the locally built `my-neko/brave:latest` with pulling disabled, cleans stale Brave singleton locks before startup, and mounts the ignored profile/download directories. The managed policy is mounted at Brave's expected `/etc/brave/policies/managed/policies.json` path.

On the target server, from the repository root:

```bash
./build my-neko/base:latest -y
./build my-neko/brave:latest -y
cp .env.example .env
# Set both passwords and any server-specific values in .env.
docker compose config
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

`.env`, `files/`, `downloads/` and `policy.json` remain ignored and must not be committed.

## Product direction

Target outcomes include:

- slow-viewer isolation;
- per-viewer adaptive quality;
- robust mobile/TV behavior;
- server-enforced view-only sharing;
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
