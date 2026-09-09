# AGENTS.md

## Purpose

This repository is a customized fork of [`m1k1o/neko`](https://github.com/m1k1o/neko). It preserves Neko's shared server-side browser/desktop model while carrying fork-specific client work for mobile/touch usability, playback/reconnect recovery, and UI/UX.

The maintenance sequence is:

1. preserve the current fork exactly — **complete**,
2. reconcile the `MyNekoProjekt` working/deployment tree into the fork — **complete; no missing reusable source/config delta was found**,
3. synchronize the completed fork with current upstream without breaking fork behavior — **NEXT**,
4. then continue the product work defined in `docs/PROJECT.md`.

## Authoritative project knowledge

Read these before substantial work:

- [`README.md`](README.md) — compact repository entry and current status.
- [`docs/PROJECT.md`](docs/PROJECT.md) — product goals, requirements, invariants, decisions, and target vs optional work.
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — verified current architecture and fork-specific implementation areas.
- [`docs/WORKPLAN.md`](docs/WORKPLAN.md) — handoff snapshot, `NEXT`, upstream-sync procedure, verification criteria, and open items.
- [`docs/LOCAL_DELTA_AUDIT.md`](docs/LOCAL_DELTA_AUDIT.md) — completed, sanitized classification of the `MyNekoProjekt` filesystem delta.
- `webpage/docs/` — inherited Neko operational/developer documentation. When it conflicts with current code/config, the repository wins.

## Source-of-truth rules

1. Code, configuration, build files, tests and CI determine what is **IMPLEMENTED**.
2. `MyNekoProjekt` remains a local/deployment reference. Its 2026-09-09 snapshot has been audited; any later local changes require a new explicit review before import.
3. Product intent in `docs/PROJECT.md` determines **TARGET**.
4. Upstream issues/PRs are evidence and implementation candidates, not automatically requirements.
5. Never silently drop fork-specific behavior during upstream conflict resolution. Reconcile semantics, not just text.
6. Never copy deployment credentials, browser profiles, downloads, cookies, lock files or other runtime data into this public repository.

## Repository map

- `client/` — Vue 2.7 + TypeScript client built with Vite; most fork-specific changes currently live here.
- `server/` — Go server and plugins.
- `apps/` — browser/application image definitions.
- `runtime/` — runtime image/container support.
- `webpage/` — inherited Neko documentation site.
- `docs/` — fork-specific durable project knowledge and work state.

## Build / lint / verification

Client (CI uses Node 18):

```bash
cd client
npm ci
npm run lint
npm run build
```

Client development server:

```bash
cd client
npm ci
npm run dev
```

Server native build (requires system dependencies documented in `webpage/docs/developer-guide/build.md`):

```bash
cd server
./build
```

Server CI-equivalent container build:

```bash
docker build ./server
```

The repository-root `./build` script owns full image/build tooling; inspect `./build --help` before using options.

## Technical invariants

- A room is one shared remote browser/desktop/session seen by multiple participants.
- At most one participant controls the shared desktop at a time.
- Admin control locking and grant/revoke semantics must remain intact.
- Fork-specific mobile/touch/trackpad, autoplay, playback-recovery, fullscreen and reconnect behavior is regression-sensitive.
- A weak/slow viewer must not degrade healthy viewers in the target architecture.
- Any future viewer/share link must enforce view-only permission server-side; hiding controls in the UI is not authorization.
- WebRTC is the currently implemented primary media path in this fork. Alternative media transports are target/candidate work until code proves otherwise.

## Definition of Done

For a change to be complete:

- relevant client lint/build and/or server build passes;
- existing admin/user/control semantics are preserved;
- touched mobile/touch, playback and reconnect paths are regression-tested per `docs/WORKPLAN.md`;
- no secrets or runtime data are committed;
- durable docs are updated only when project truth, architecture, target state, or work status actually changed.
