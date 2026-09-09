# Work Plan / Handoff State

Last consolidated: 2026-09-09.

## Environment boundary

Codex is the development/static-review environment only.

Codex must not install project dependencies, execute builds/tests/linters/type-checkers, start the client/server/Docker, execute repository scripts, or perform WebRTC/media/device runtime validation.

All runtime/build/test verification happens later on the real target server.

Codex should still inspect source/config/diffs thoroughly and prepare exact server-side verification steps.

## Verified repository snapshot

Fork:

```text
officialtaxz2/neko
master
d9c1afd564ad4c293a0b1b28aecbc103d1d1d9c6
```

Upstream checked during bootstrap:

```text
m1k1o/neko
master
b0f01cedea68893e85a3fd852c0521238c285695
```

Merge base:

```text
d74052bb844c43a0cc3c2386d083f7505dc483a2
```

At bootstrap GitHub reported both sides as diverged, with 33 commits on each side after the merge base.

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

## Target-server verification

This phase is performed by the operator on the real server, not by Codex.

### Client

```bash
cd client
npm ci
npm run lint
npm run build
```

Required for commit `2d89027e`: confirm `npm ci`, TypeScript lint and the Vite production build on the target server. Codex-side manifest/lock inspection confirms that root dependency and dev-dependency declarations match, but this is not build evidence.

### Server if relevant

```bash
cd server
./build
```

or:

```bash
docker build ./server
```

### Manual regression on target server

Access/control:

- admin and regular-user login;
- only one active controller;
- request/grant/revoke;
- lock enforcement;
- admin behavior under lock.

Desktop:

- Chrome/Chromium and Firefox;
- join/video/audio/control;
- refresh/rejoin;
- transient network reconnect.

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

## NEXT

The local reconciliation prerequisite is complete. Synchronize with current upstream:

1. fetch current upstream;
2. recompute merge base/divergence;
3. use an integration working state;
4. review changes by subsystem;
5. merge upstream `master`;
6. resolve conflicts semantically.

Never globally resolve with blanket `ours`/`theirs`.

High-risk areas:

- `client/src/components/video.vue`
- `client/src/neko/base.ts`
- `client/src/neko/index.ts`
- mobile/touch/settings
- connect/autoplay
- UI styles/components

Reassess upstream at merge time instead of relying on the bootstrap snapshot.

## Product priority after stable synced baseline

1. slow-client isolation;
2. evaluate existing upstream multi-pipeline + bandwidth estimator for per-viewer quality;
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
