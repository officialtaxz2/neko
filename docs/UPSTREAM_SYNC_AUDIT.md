# Upstream Synchronization Audit

## Authorized focused adaptations — 2026-10-07

After the read-only assessment below, the operator approved the improvement
plan. The working source now adapts the reviewed #697 PLI exit, #711 missing-ID
and empty-selection behavior (preserving explicit fork ladder order and closing
partially initialized peers), #701 pipeline shorthand, #700 native allocation
ownership and #709 Safari clipboard fallback. A small common bitrate mutex
addresses the reviewed bucket race without migrating the #699 subscription API.
These are source adaptations, not a merge/cherry-pick or changed upstream ref.
Target checks/builds/race/native-drop/Safari acceptance are pending,
**NOT EXECUTED IN CODEX**. No upstream performance benefit is claimed.
Full #699 convergence, auth product changes, dependency upgrades and optional
transport disposition retain their distinct conditions in [Workplan NEXT](WORKPLAN.md#next).

## Read-only upstream comparison — 2026-10-07

This is a new source assessment, **not a second synchronization**. Initial fork
HEAD was `testing` at `a91d9388d75f3d042f212398a1dcff1af198992d`, with a clean
working tree; application HEAD is `8741f7880d9a709e6dc17924d2af0564a949ccd9`.
`master` remains `d9105ef8`. The local `upstream/master` is still
`b0f01cedea68893e85a3fd852c0521238c285695`, the previously integrated base.

The GitHub repository/compare API was read independently of that stale local
reference. Remote master was
`a2cb38dd10e20774b5f3ecf8b8c5d0b07a2051f5`, dated 2026-10-04, with 18 commits
and 71 changed files after `b0f01ced`; the API relation to that base is ahead 18,
behind zero. The fork has 158 commits after the locally integrated upstream
base, including implementation, deployment helpers and evidence/documentation.
That count does not quantify useful features or maintenance quality.

Primary comparison:
[b0f01ced…a2cb38dd](https://github.com/m1k1o/neko/compare/b0f01cedea68893e85a3fd852c0521238c285695...a2cb38dd10e20774b5f3ecf8b8c5d0b07a2051f5).
All 18 commit subjects and the relevant capture, WebRTC, signaling/configuration,
clipboard, Markdown and authentication patches were examined against current
fork source. No fetch, merge, cherry-pick, ref update, code execution or
dependency install was performed. This is a semantic review of relevant
subsystems, not a fresh runtime verdict or exhaustive audit of unrelated images.

### Triage against the current fork

| Upstream change | Current fork and significance | Proposed disposition, after operator decision |
|---|---|---|
| [`371f5c2c`, RTCP PLI exit signal #697](https://github.com/m1k1o/neko/commit/371f5c2c7bdd92d556d29b42b442a3a8eac28871) | The fork still ranges a stopped ticker's channel without a separate exit. Admitted inbound microphone/camera tracks can leave a waiting goroutine; passive watching alone does not enter this path. No magnitude was measured. | Small correctness patch first; verify repeated incoming-track teardown on the target. Preserve inbound-media/view-only permissions. |
| [`a2cb38dd`, pipeline IDs/empty guard #711](https://github.com/m1k1o/neko/commit/a2cb38dd10e20774b5f3ecf8b8c5d0b07a2051f5) | Pipeline definitions without IDs can be absent from the fork provider's source enumeration. Signaling has an unguarded fallback index, but current `openPeer` already rejects empty provider sources first; an actual panic is not established. | Derive missing IDs and fail closed on empty selection, checking peer cleanup. Preserve explicit `high, medium, low` order; alphabetical `high, low, medium` must not become an adaptive quality order. |
| [`f4d632df`, capture subscription API #699](https://github.com/m1k1o/neko/commit/f4d632df2cb0a6109229c363c498ec304e0c912d) | Explicit subscriptions replace reflected listener identity; immutable listener snapshots avoid rebuilding a slice per sample. Idempotent start and done/once shutdown address lifecycle races. Bitrate buckets get a common mutex; GStreamer avoids the intermediate duplicate buffer extraction, still copying into Go ownership. These overlap capture internals that remain under the fork provider. | Converge the low-level seam in a separate validated block. Keep the fork's format/caps generations, PTS/DTS validity, running-time contract, authorization and bounded provider queues. Do not replace corrected bits/s accounting with upstream byte-count arithmetic or build a second fan-out layer. Performance gains require measurement. |
| [`ba098db6`, rendered Markdown containment](https://github.com/m1k1o/neko/commit/ba098db66396df6b6f4201dbee9920253cfc3388) | The fork independently removed runtime Vue compilation and strengthened attribute/URL handling in the 93f1fa63 stability repairs. Tooltip and Axios containment extend beyond this upstream patch. | Compare security semantics and hostile fixtures rather than reapply a duplicate renderer or overwrite fork emoji/Open-in-App behavior. A package advisory is not closed merely by containment. |
| [`65fba485`, Safari clipboard fallback #709](https://github.com/m1k1o/neko/commit/65fba485de2987998c7028b04240b21669d674e0) | The current client fallback condition handles Firefox but misses Safari. This is a small browser compatibility gap, distinct from media transport. | Candidate focused adoption with actual controller/Safari permission/clipboard checks and unchanged view-only denial. |
| [`73b9d71b`, pipeline-string shorthand #701](https://github.com/m1k1o/neko/commit/73b9d71b0d94129cd3d0421f17bcccbc04dee168) | The capture configuration lacks this decode hook. Current explicit full pipeline objects need no change. | Adopt only with configuration validation coverage; this is compatibility/convergence, not a playback-performance optimization. |
| [`347418c3`, native drag/drop allocation cleanup #700](https://github.com/m1k1o/neko/commit/347418c3e7c1bc04b9a088d477025ea57daf05ae) | Unrelated to transport selection but a useful native ownership correction in the changed upstream surface. No target leak measurement is supplied. | Review the fork's X11/drop ownership and apply as a small separate fix if the allocation path matches; exercise repeated URI drops on the target. |
| [OIDC-only authentication `dae23a90` #686](https://github.com/m1k1o/neko/commit/dae23a90b52bb31d8cdf047f41ae01b9bf41defa) and its provider/email/scope refactors | New product/authentication scope, not a prerequisite for media subscription convergence. It crosses legacy login, view-only tokens and session capability assumptions. | Defer unless required. If adopted later, include browser-bound OAuth state from [`1725bfa5` #708](https://github.com/m1k1o/neko/commit/1725bfa50b91c218f378a15a7a0529222ef8029a) and revalidate the full role/token matrix; do not transplant only an isolated auth fragment. |
| [`bdd0428a`, internal originless dial #710](https://github.com/m1k1o/neko/commit/bdd0428a607ccfd22eebabc74b85ae17a4cb42d5) | Relevant to the internal event-WebSocket/legacy adapter. It is not evidence that a media endpoint should accept absent Origin. | Review only if changing that adapter or adopting dependent auth changes. Preserve exact external media Origin/TLS/proxy trust checks and distinguish trusted loopback from public routes. |
| Compose examples, TOR download URI, release notes | Deployment/example/release maintenance rather than fixes to the selected media paths. | Review when their deployed component is used; do not replace the sanitized operator-specific compose, Brave policy mount or opt-in overlays wholesale. |

### What the fork adds and what it should preserve

The earlier integration already brought bounded per-peer media channels,
nonblocking fan-out, asynchronous/leaky GStreamer output, codec/hardware options
and the initial multi-pipeline/estimator machinery. Credit those to upstream;
do not present them as new fork inventions or remove them as unused solely
because the current deployment selects VP8.

The fork adds or materially revises mobile/touch UI and playback/reconnect
behavior, server-enforced passive sharing, validated bitrate units/startup and
receiver-evidence/probe quality logic, the richer encoded-media/provider and
participant-delivery boundary, explicit per-client selection, WebCodecs delivery,
HLS transcoding/packaging, separate deployment/observability and security
containment. These additions have different evidence strengths. The accepted
adaptive corrections and role/permission boundary have stronger justification
than unmeasured optional-transport performance claims.

Capture convergence is the clearest opportunity to reduce divergent plumbing.
Upstream #699 has sample timestamps and subscriptions, but not the fork's whole
authorized delivery/format-generation contract. The correct direction is one
upstream-aligned capture ownership model **behind** the existing provider, with
unchanged behavior for WebRTC, WebCodecs and HLS. It is not a wholesale replacement
of the provider, or an upstream merge that silently discards fork semantics.
Fresh target codec/lifecycle/race checks are justified only once that code changes.

The [comparative stability review](STABILITY_REVIEW.md#comparative-fork-and-transport-review--2026-10-07)
records overlapping recovery, common event-plane waits, HLS format/conversion
costs, stale configuration and evidence limits. The
[Workplan NEXT](WORKPLAN.md#next) orders adoption and measurement. No broad merge,
dependency migration, transport removal or promotion follows automatically.

## Historical completed synchronization — 2026-09-09

Completed in Codex: 2026-09-09.

This document records the semantic synchronization of the preserved/customized fork with the then-current `m1k1o/neko:master`. It distinguishes repository integration from runtime validation.

## Git scope

| Item | Reference |
|---|---|
| Pre-sync fork HEAD | `18e9320c892b4069757a71c9093c9c9b4dd7bd4a` |
| Safety branch | `safety-pre-upstream-sync-20260909` |
| Upstream HEAD fetched with prune | `b0f01cedea68893e85a3fd852c0521238c285695` |
| Pre-sync merge base | `d74052bb844c43a0cc3c2386d083f7505dc483a2` |
| Pre-sync divergence | 36 fork-only / 33 upstream-only commits |
| Integration branch | `integration/upstream-20260909` |
| Merge commit | `4e99b8d3ca720d1f184544306820e388716ba23a` |
| Post-merge relation to fetched upstream | 37 ahead / 0 behind |
| Promotion | reviewed integration history fast-forwarded to `master` |

The upstream-only delta covered 97 files with 2,559 insertions and 192 deletions. It was reviewed by subsystem before merging; no blanket `ours` or `theirs` resolution was used.

## Accepted subsystem changes

### Client and protocol

- Clipboard synchronization when the active controller refocuses the browser window.
- `?scroll=` query parameter for initial scroll sensitivity.
- Separate file download/upload/delete capability fields, UI gating, single deletion and bulk deletion.
- Optional Open-in-App state, chat actions and protocol events.
- Automatic video-pipeline selection request in the legacy protocol handler.
- Matching locale additions for file deletion/selection.

### Server, media and input

- Per-track sample channels are bounded; a full channel drops that peer's sample instead of blocking the producer.
- Capture fan-out copies listeners and releases the listener mutex before dispatch.
- GStreamer appsinks are asynchronous and audio queues are bounded/leaky downstream to reduce accumulated latency.
- H.265 codec parsing and GStreamer pipelines for software, VA-API and NVENC.
- Current VA-API element names and NVIDIA automatic encoder fallback.
- Configured V3 capture pipelines are no longer overwritten by legacy video settings.
- Capture-pointer visibility configuration, including broadcast/screencast.
- XInput/XTEST keyboard dispatch for Firefox/GDK3.
- Nil-safe delayed WebRTC peer teardown and corrected legacy host authorization logic.
- Standard-library `slices.Contains` replaces the deleted local `ArrayIn` helper.

### Plugins and permissions

- Server-enforced non-admin file download/upload/delete permissions.
- File deletion validates authentication, feature/role permission, filename and root confinement before removal.
- Optional Open-in-App plugin validates authentication/host capability and permits only `http`/`https` URLs before invoking its configured executable without a shell.

### Images, deployment, CI and inherited documentation

- ARM64 Widevine runtime support and ARM64 Google Chrome image builds.
- Chromium-family policy updates and browser image fixes.
- Branch-test image workflow and expanded image architecture matrix.
- Removal of the baked root `config.yml`; equivalent security-sensitive defaults moved into server flags, while deployments remain responsible for explicit environment/mounted configuration.
- Updated example compose and inherited operational/developer documentation.

## Conflict resolution

Only three files conflicted textually:

- `client/src/components/settings.vue`: retained force-touch and trackpad settings/detection; added Open-in-App preference accessors and UI.
- `client/src/components/side.vue`: retained the fork's members tab, context integration and animated tab layout; added upstream file capability checks.
- `client/src/components/video.vue`: retained cursor-position subscription and all fork media/touch/recovery teardown; added the stable focus listener and matching removal for clipboard sync.

The high-risk `client/src/neko/base.ts` had no upstream delta. `client/src/neko/index.ts`, settings/store modules and the UI files touched by both histories were compared after the automatic merge to ensure the fork paths remained present.

## Additional integration repair

The automatic merge was textually clean in `client/src/neko/index.ts` but not semantically complete: upstream made `user_download`, `user_upload` and `user_delete` required in `FileTransferListPayload`, while the fork's demo-mode synthetic file-list message still emitted the old shape. The demo payload now supplies all three fields, preventing a likely TypeScript error and keeping demo file controls coherent.

One upstream-added blank line at the end of `runtime/widevine-installer/widevine_fixup.py` was removed so `git diff --check` is clean.

## Exclusions and security review

No content was imported from instance-only `MyNekoProjekt` deployment/runtime data during the upstream merge. The merge contains no browser profile, download, cookie, lock file, local policy contents, deployment credential or raw local compose override. The operator-confirmed compose structure was reconstructed separately afterward with secrets and runtime payloads excluded.

The staged path list and index were checked for forbidden runtime paths and high-confidence private-key/token patterns. Only repository examples and secret-variable references already intended by upstream remain; no credential value was added.

## Static verification completed

- upstream ancestry, pre/post divergence and both merge parents verified;
- no unmerged paths or conflict markers;
- `git diff --check` clean;
- all changed JSON files parsed successfully;
- client/server Open-in-App event names and file capability fields cross-checked;
- Open-in-App store registration and settings accessors cross-checked;
- H.265 parser/config/pipeline references cross-checked;
- no remaining `ArrayIn` call sites after helper deletion;
- fork touch/trackpad/cursor, autoplay, playback recovery, fullscreen and reconnect paths confirmed present by source inspection.

No dependency installation, project script, test, linter, type-checker, build, Docker image, server, browser, WebRTC session or device check was executed in Codex. Those checks remained target-server work and were later operator-confirmed as passed for the applicable deployment matrix on 2026-09-09.

## Outcome

The upstream synchronization is Codex-side complete and its reviewed history is on `master`. Runtime/build status remains **NOT EXECUTED IN CODEX** by policy; the operator later confirmed that the applicable target-server build and regression matrix passed after correcting the Brave policy mount filename. The next bounded product unit is defined in [`WORKPLAN.md`](WORKPLAN.md#next).
