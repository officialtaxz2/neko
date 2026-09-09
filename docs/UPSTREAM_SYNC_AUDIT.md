# Upstream Synchronization Audit

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

No content was imported from instance-only `MyNekoProjekt` deployment/runtime data during this sync. The merge contains no browser profile, download, cookie, lock file, local policy mount, deployment credential or local compose override.

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

No dependency installation, project script, test, linter, type-checker, build, Docker image, server, browser, WebRTC session or device check was executed in Codex.

## Outcome

The upstream synchronization is Codex-side complete on `integration/upstream-20260909`. Runtime/build status is **NOT EXECUTED IN CODEX**. The next bounded unit is the target-server verification and promotion procedure in [`WORKPLAN.md`](WORKPLAN.md#next).
