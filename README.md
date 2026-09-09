# officialtaxz2/neko

Customized fork of [m1k1o/neko](https://github.com/m1k1o/neko), a self-hosted shared virtual browser/desktop streamed to multiple participants.

This fork keeps Neko's shared multi-user session model and adds substantial client-side work around mobile/touch use, trackpad-style control, playback/reconnect recovery, and UI/UX.

## Status

### IMPLEMENTED

- Shared Neko browser/desktop session with multi-user access.
- Existing Neko admin/user/control semantics.
- Fork-specific client redesign.
- Touch/mobile controls and trackpad mode.
- Mobile keyboard/helper integration.
- Autoplay/muted fallback.
- Video stream health/recovery logic.
- ICE disconnect/recovery handling.
- Demo-mode client infrastructure.

### Integration snapshot

Bootstrap snapshot (2026-09-09):

- fork `officialtaxz2/neko:master`: `d9c1afd564ad4c293a0b1b28aecbc103d1d1d9c6`
- upstream `m1k1o/neko:master`: `b0f01cedea68893e85a3fd852c0521238c285695`
- merge base: `d74052bb844c43a0cc3c2386d083f7505dc483a2`
- GitHub comparison: diverged, with 33 commits on each side after the merge base.

Local reconciliation (2026-09-09):

- safety branch `safety-pre-local-delta-audit-20260909` preserves the original fork HEAD;
- the complete supplied `MyNekoProjekt` tree was audited;
- no missing reusable source/config delta was found;
- credentials, deployment-only compose/policy data, browser profile and downloads were excluded;
- client lock/type consistency was repaired in `2d89027e` and awaits target-server verification.

Do not perform a blind upstream overwrite.

## NEXT

Synchronize current `m1k1o/neko` upstream on a dedicated integration branch, preserving fork-specific behavior through semantic conflict resolution. The upstream sync has not started.

See `docs/WORKPLAN.md`.

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
