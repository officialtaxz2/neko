# AGENTS.md

## Purpose

This repository is a customized fork of [`m1k1o/neko`](https://github.com/m1k1o/neko). It preserves Neko's shared server-side browser/desktop model while carrying fork-specific client work for mobile/touch usability, playback/reconnect recovery, and UI/UX.

Immediate sequence:

1. preserve the current fork — **complete**,
2. reconcile desired local changes from `MyNekoProjekt` — **complete; no missing reusable source/config delta was found**,
3. synchronize the completed fork with current upstream without breaking fork behavior — **complete on `integration/upstream-20260909`; target-server verification pending**,
4. validate and promote the integration branch — **NEXT**,
5. continue the product work defined in `docs/PROJECT.md`.

## Authoritative knowledge

Read before substantial work:

- `README.md` — compact entry and current status.
- `docs/PROJECT.md` — goals, requirements, invariants and target state.
- `docs/ARCHITECTURE.md` — verified current architecture.
- `docs/WORKPLAN.md` — current `NEXT`, sync procedure, server-side verification and open items.
- `docs/LOCAL_DELTA_AUDIT.md` — completed, sanitized classification of the supplied `MyNekoProjekt` snapshot.
- `docs/UPSTREAM_SYNC_AUDIT.md` — completed semantic review and merge record for the 2026-09-09 upstream synchronization.
- `webpage/docs/` — inherited Neko documentation. Current repository code/config wins on conflicts.

## Truth rules

1. Code, configuration, Git history and other real artifacts determine **IMPLEMENTED**.
2. The supplied `MyNekoProjekt` snapshot was audited on 2026-09-09. Any later local delta must be reviewed explicitly before import.
3. `docs/PROJECT.md` determines **TARGET**.
4. Upstream issues/PRs are evidence or candidates, not automatically requirements.
5. Never silently drop fork-specific behavior during upstream conflict resolution.
6. Never commit deployment credentials, browser profiles, downloads, cookies, lock files or other runtime data.

## Codex execution policy

**Codex is an edit and static-review environment only. It is not the target runtime environment.**

Do not execute project code or runtime verification in Codex.

Do **not**:

- install project dependencies (`npm install`, `npm ci`, `go get`, etc.);
- start the client, server, browser, Vite, Docker containers or images;
- run tests, linters, type-checkers, builds or package scripts;
- run repository build/start scripts;
- perform WebRTC/media/network/device runtime tests;
- treat Codex-environment execution as evidence of target-server compatibility.

Static repository work is allowed and expected:

- read and compare files;
- inspect Git history, status and diffs;
- inspect manifests, configuration and source;
- reason about syntax/type/build/runtime implications;
- identify likely regressions by inspection;
- prepare exact verification commands/checklists for the real server.

Runtime/build/test status must be reported as **NOT EXECUTED IN CODEX** unless results are supplied from the target server.

## Repository map

- `client/` — Vue 2.7 + TypeScript/Vite client; most fork-specific work currently lives here.
- `server/` — Go server and plugins.
- `apps/` — browser/application image definitions.
- `runtime/` — runtime image/container support.
- `webpage/` — inherited Neko documentation site.
- `docs/` — fork-specific durable project knowledge.

## Server-side verification reference

These commands are for the **real target server/environment only**. Codex must not run them.

Client:

```bash
cd client
npm ci
npm run lint
npm run build
```

Server:

```bash
cd server
./build
```

Container build:

```bash
docker build ./server
```

Use only the checks relevant to the changed areas, with broader verification after major integrations.

## Technical invariants

- One shared remote browser/desktop/session is seen by multiple participants.
- At most one participant controls the shared desktop at a time.
- Admin lock and grant/revoke behavior must remain intact.
- Fork mobile/touch/trackpad, autoplay, playback-recovery, fullscreen and reconnect behavior is regression-sensitive.
- A weak viewer must not degrade healthy viewers in the target architecture.
- Future view-only sharing must be enforced server-side; hiding controls in the UI is not authorization.
- WebRTC is the currently implemented primary media path.

## Definition of Done for Codex work

A Codex task is complete when:

- the requested repository change is implemented and reviewed statically;
- relevant diffs and surrounding code are inspected for regressions;
- no secrets/runtime data are committed;
- required target-server verification is explicitly listed as pending;
- durable docs are updated when project truth/status changed.

Do not claim runtime/build/test success without results from the real target server.
