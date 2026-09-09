# officialtaxz2/neko

Customized fork of [m1k1o/neko](https://github.com/m1k1o/neko), a self-hosted shared virtual browser/desktop streamed to multiple participants.

This fork keeps Neko's shared multi-user session model and adds substantial client-side work around mobile/touch use, trackpad-style control, playback/reconnect recovery, and UI/UX.

## Status

### IMPLEMENTED

Repository implementation is listed here independently of runtime validation. The latest upstream integration is statically reviewed; its target-server build and regression matrix are still pending.

- Shared Neko browser/desktop session with multi-user access.
- Existing Neko admin/user/control semantics.
- Fork-specific client redesign.
- Touch/mobile controls and trackpad mode.
- Mobile keyboard/helper integration.
- Autoplay/muted fallback.
- Video stream health/recovery logic.
- ICE disconnect/recovery handling.
- Demo-mode client infrastructure.
- Per-peer non-blocking media sample delivery so one backpressured WebRTC track does not block dispatch to other tracks.
- Multi-pipeline stream-selection and per-peer bandwidth-estimator infrastructure; the legacy protocol path requests automatic selection when configured.
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
- `master` was fast-forwarded to the reviewed integration history; commits after `4e99b8d3` only update repository knowledge/status.

The 97-file upstream delta was reviewed by subsystem. Conflicts in `settings.vue`, `side.vue` and `video.vue` were resolved semantically, preserving the fork's touch/trackpad, UI and cursor/recovery behavior while accepting the upstream permissions, Open-in-App and focus-clipboard changes. A hidden demo-mode payload mismatch caused by the new file-transfer rights fields was also repaired.

The detailed record is in [`docs/UPSTREAM_SYNC_AUDIT.md`](docs/UPSTREAM_SYNC_AUDIT.md).

Local reconciliation (2026-09-09):

- safety branch `safety-pre-local-delta-audit-20260909` preserves the original fork HEAD;
- the complete supplied `MyNekoProjekt` tree was audited;
- no missing reusable source/config delta was found;
- credentials, deployment-only compose/policy data, browser profile and downloads were excluded;
- client lock/type consistency was repaired in `2d89027e` and awaits target-server verification.

Do not perform a blind upstream overwrite.

## NEXT

Build and regression-test the integrated `master` on the real target server. After acceptance, record the results and begin the slow-client/adaptive-quality product phase.

See `docs/WORKPLAN.md` and `docs/UPSTREAM_SYNC_AUDIT.md`.

## Development environment policy

Codex is used for source editing, repository analysis and static review only. It is **not** the deployment/test environment.

Do not run the application, install dependencies, execute builds/tests/linters, start Docker, or perform WebRTC/device runtime tests inside Codex.

Runtime verification happens separately on the real server. Known server-side commands and the verification matrix are documented in `AGENTS.md` and `docs/WORKPLAN.md`.

## Product direction

Target outcomes include:

- slow-viewer isolation;
- per-viewer adaptive quality;
- robust mobile/TV behavior;
- server-enforced view-only sharing;
- robust reconnect/recovery;
- a practical non-WebRTC viewer-media fallback.

See `docs/PROJECT.md`.

## Structure

```text
client/      Vue 2.7 + TypeScript/Vite client
server/      Go Neko server and plugins
apps/        browser/application image definitions
runtime/     runtime container support
webpage/     inherited Neko documentation site
docs/        fork-specific project/architecture/work knowledge
```

## Security / deployment note

`MyNekoProjekt` contains instance-specific deployment/runtime material. Import only reviewed and sanitized code/config deltas. Credentials, browser profiles, downloads, cookies and lock/runtime files stay outside Git.

The completed local classification is recorded in `docs/LOCAL_DELTA_AUDIT.md`.
