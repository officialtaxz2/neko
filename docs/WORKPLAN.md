# Work Plan / Handoff State

Last consolidated: 2026-09-09.

This file owns current work status. Do not turn `AGENTS.md` into a roadmap.

## ACTIVE

The repository handoff and complete `MyNekoProjekt` reconciliation are finished. The next bounded unit is semantic synchronization with current upstream; it has not started.

## Verified repository snapshot

Pre-audit fork baseline:

```text
repository: officialtaxz2/neko
branch:     master
HEAD:       d9c1afd564ad4c293a0b1b28aecbc103d1d1d9c6
```

Current upstream checked during bootstrap:

```text
repository: m1k1o/neko
branch:     master
HEAD:       b0f01cedea68893e85a3fd852c0521238c285695
```

Verified merge base:

```text
d74052bb844c43a0cc3c2386d083f7505dc483a2
```

GitHub comparison at bootstrap reported the branches as diverged, with 33 commits on the fork side and 33 on the upstream side after the merge base.

The fork did previously merge upstream at commit `8fcb855a` on 2026-07-05; do not assume it has never been synchronized.

## COMPLETED — `MyNekoProjekt` local delta audit

Completed on 2026-09-09 against the pre-audit baseline above.

- Safety branch `safety-pre-local-delta-audit-20260909` preserves the original fork HEAD.
- A clean filesystem baseline and the complete supplied local tree were compared recursively, excluding `.git/`.
- All 724 baseline files exist locally: 138 are byte-identical, 585 differ only by CRLF/LF, and only `docker-compose.yaml` differs substantively.
- The 3,929 local-only files are 3,928 persistent browser-profile/runtime files under `files/**` plus an empty instance `policy.json`; `downloads/` is empty.
- No source feature/fix or reusable repository config was missing, so no local source/config content was imported.
- The local compose override and its embedded credentials were excluded. Active credentials from that artifact should be rotated operationally.
- Root ignore rules now guard the local reference and its runtime/profile paths against accidental staging.
- Verification exposed and repaired the stale pre-Vite client lockfile/type-check baseline in commit `2d89027e`.

The exhaustive classification and sanitized decision record is in [`LOCAL_DELTA_AUDIT.md`](LOCAL_DELTA_AUDIT.md).

## Verification matrix

### Automated / build

Client:

```bash
cd client
npm ci
npm run lint
npm run build
```

2026-09-09 result: all three commands passed after commit `2d89027e` repaired the stale lock/type baseline. Available environment: Node `24.19.0`, npm `11.17.0`. The build had only Vite's non-failing large-chunk warning.

Server if relevant:

```bash
cd server
./build
```

or CI-equivalent:

```bash
docker build ./server
```

Not run for the local reconciliation: no server, root build, runtime configuration, or application-image source changed.

### Manual regression — minimum baseline

No live application/server or device test was run during the local reconciliation. The audit imported no runtime source/config delta, and the corrective client changes are dependency-lock/build typing changes, but this is not evidence of WebRTC, multi-user, mobile, or Smart-TV behavior. Run the applicable matrix against the user's deployment during the upstream-integration verification phase.

#### Access/control

- admin login works;
- regular-user login works;
- only one active controller at a time;
- control request/grant/revoke remains correct;
- locked controls prevent normal-user control;
- admin behavior remains correct under lock.

#### Desktop browsers

At minimum current Chrome/Chromium and Firefox:

- join;
- video;
- audio;
- control;
- reconnect after transient network loss;
- refresh and rejoin.

#### Mobile

At minimum Android Chrome and iOS/iPadOS Safari when hardware is available:

- join;
- autoplay behavior does not result in silent black screen;
- manual play/unmute overlay works;
- audio can be enabled by user gesture;
- orientation/fullscreen changes;
- trackpad mode;
- touch control;
- mobile keyboard/helper;
- reconnect/recovery.

#### Smart-TV / embedded browser

For each actually supported target device:

- join;
- WebRTC establishment;
- playback;
- audio;
- refresh/reconnect;
- failure is diagnosable, not an indefinite spinner/black screen.

Do not claim Smart-TV support without device/browser evidence.

## NEXT

### Upstream synchronization

The local reconciliation prerequisite is complete. For the next bounded unit:

1. add/fetch upstream:
   ```bash
   git remote add upstream https://github.com/m1k1o/neko.git
   git fetch upstream --tags
   ```
2. create a dedicated integration branch from the completed fork baseline;
3. recompute current merge base and compare at that time;
4. review upstream changes by subsystem before merging;
5. merge upstream `master` into the integration branch;
6. resolve conflicts semantically.

Never globally resolve conflicts with blanket `--ours`/`--theirs`.

#### High-risk conflict areas

- `client/src/components/video.vue`
- `client/src/neko/base.ts`
- `client/src/neko/index.ts`
- touch/mobile controls and settings
- connect/autoplay behavior
- UI styles/components

Upstream also changed client files after the fork diverged, so passing a textual merge is not sufficient.

#### Upstream changes to evaluate rather than reimplement

At bootstrap, current upstream includes work in areas directly relevant to this fork, including:

- H.265 support;
- audio jitter/backpressure improvements;
- file-transfer permissions/deletion/bulk operations;
- automatic quality selection for legacy clients when multiple pipelines and bandwidth estimation are enabled;
- capture/pipeline configuration improvements;
- additional browser/runtime changes;
- WebRTC/session fixes.

Reassess upstream at merge time; this list is not frozen.

## Product development after stable synced baseline

### NEXT priority order (TARGET)

1. prove or fix slow-client isolation;
2. configure/test existing upstream multi-pipeline + bandwidth estimator for per-viewer quality;
3. stabilize mobile/reconnect behavior after sync;
4. implement server-enforced view-only share role/link;
5. prototype practical non-WebRTC viewer media fallback.

This order intentionally evaluates existing upstream adaptive-quality work before building a separate ABR subsystem.

## Fallback prototype sequence

If alternative media work begins, evaluate upstream issue #690 in stages:

1. media-subscription abstraction;
2. WebCodecs + dedicated WebSocket media;
3. only then WebTransport.

Do not import all prototype branches blindly. Rebase/reconcile them against the then-current fork/upstream baseline and test VP8/H.264/browser support.

## LATER / OPTIONAL

- WebTransport/QUIC productionization;
- automatic transport selection;
- automatic codec selection;
- HLS/DASH fallback if justified;
- broader v3 modular client/library rewrite.

## OPEN

- Exact supported Smart-TV browser/device matrix.
- Whether existing upstream adaptive quality is sufficient for the desired per-viewer experience.
- Which first non-WebRTC media backend best satisfies latency + compatibility after measurement.
- Token format/lifetime/revocation model for the future view-only share link.
- Whether the fork should ultimately migrate away from the legacy Vue 2 client as part of upstream's longer-term v3 direction.

## Evidence / upstream references

- upstream repository: https://github.com/m1k1o/neko
- v3 modular architecture: https://github.com/m1k1o/neko/issues/371
- alternative media proposal: https://github.com/m1k1o/neko/issues/690
- mobile trackpad issue: https://github.com/m1k1o/neko/issues/640
- releases: https://github.com/m1k1o/neko/releases
