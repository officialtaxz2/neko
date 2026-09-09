# officialtaxz2/neko

Customized fork of [m1k1o/neko](https://github.com/m1k1o/neko), a self-hosted shared virtual browser/desktop streamed to multiple participants.

This fork keeps Neko's multi-user shared-session model and adds substantial client-side work around mobile/touch use, trackpad-style control, playback/reconnect recovery, and a redesigned UI. It is intentionally being brought back into a maintainable relationship with upstream before larger streaming rework continues.

## Status

**IMPLEMENTED**

- Shared Neko browser/desktop session with multi-user access.
- Existing Neko admin/user and control semantics.
- Fork-specific client redesign.
- Touch-device detection and mobile controls.
- Trackpad mode with a client-side cursor.
- Mobile keyboard/keyboard-helper integration.
- Mobile autoplay handling and muted fallback.
- Video stream health/recovery logic for stalled/waiting/ended/muted tracks.
- Client ICE failure/disconnect recovery timeout.
- Demo-mode client infrastructure.

**INTEGRATION STATUS**

Bootstrap snapshot (2026-09-09):

- fork `officialtaxz2/neko:master`: `d9c1afd564ad4c293a0b1b28aecbc103d1d1d9c6`
- upstream `m1k1o/neko:master`: `b0f01cedea68893e85a3fd852c0521238c285695`
- verified merge base: `d74052bb844c43a0cc3c2386d083f7505dc483a2`
- GitHub compare reports the branches as diverged; each side has 33 commits after the merge base.

Local reconciliation (2026-09-09):

- safety branch: `safety-pre-local-delta-audit-20260909` at `d9c1afd564ad4c293a0b1b28aecbc103d1d1d9c6`;
- complete `MyNekoProjekt` filesystem audit finished with no missing reusable source/config delta;
- local deployment credentials, browser profile, downloads, policy file and compose overrides were excluded;
- the client lock/type baseline was repaired in `2d89027e` after the required verification exposed a stale pre-Vite lockfile;
- `npm ci`, `npm run lint` and `npm run build` pass (Node 24/npm 11 in the available environment).

Do **not** perform a blind upstream overwrite. See [`docs/WORKPLAN.md`](docs/WORKPLAN.md).

## NEXT

Create a dedicated integration branch and synchronize current `m1k1o/neko` upstream semantically, preserving the fork-specific client behavior and following the conflict/verification procedure in the work plan. The upstream sync has not started yet.

See [`docs/WORKPLAN.md`](docs/WORKPLAN.md#next).

## Product direction

The target is a robust shared browser experience where:

- multiple people watch the same server-side session;
- only one person controls it at a time;
- admins can fully lock or assign control;
- mobile and Smart-TV clients are reliable;
- a slow viewer cannot degrade healthy viewers;
- quality can adapt per viewer;
- viewers can eventually join through a server-enforced view-only share link;
- WebRTC can eventually have a practical fallback where browser/network compatibility requires it.

The authoritative requirements and state labels are in [`docs/PROJECT.md`](docs/PROJECT.md).

## Development

Client:

```bash
cd client
npm ci
npm run lint
npm run build
```

Development server:

```bash
cd client
npm run dev
```

Server:

```bash
cd server
./build
```

CI also builds the server Docker context for amd64 and arm64.

## Structure

```text
client/      Vue 2.7 + TypeScript/Vite client
server/      Go Neko server and plugins
apps/        browser/application image definitions
runtime/     runtime container support
webpage/     inherited Neko documentation site
docs/        fork-specific product/architecture/work knowledge
```

## Documentation

- [`AGENTS.md`](AGENTS.md) — coding-agent rules and definition of done.
- [`docs/PROJECT.md`](docs/PROJECT.md) — goals, requirements, decisions, status.
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — verified current architecture.
- [`docs/WORKPLAN.md`](docs/WORKPLAN.md) — sync plan, `NEXT`, verification matrix, open work.
- [`docs/LOCAL_DELTA_AUDIT.md`](docs/LOCAL_DELTA_AUDIT.md) — completed local filesystem audit and exclusion decisions.
- `webpage/docs/` — detailed inherited Neko configuration, install and developer docs.
- Upstream: https://github.com/m1k1o/neko

## Security / deployment note

The local `MyNekoProjekt` deployment tree contains instance-specific deployment data and must not be copied wholesale into this public repository. Import only reviewed code/config deltas. Credentials, browser profiles, downloads, cookies, singleton/lock files and similar runtime data stay outside Git.
