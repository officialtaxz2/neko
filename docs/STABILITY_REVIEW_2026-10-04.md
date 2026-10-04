# Integrated stability review — 2026-10-04

Scope: `testing` baseline `3f1a62b0d12594f1dfc11a6c0064ae4d69116660` plus the
Phase 4 assets and repairs in this implementation block. The final exact commit
and image IDs must be recorded by the target-server helpers. Source/configuration
inspection only: **NOT EXECUTED IN CODEX** for tests, type/build, containers and
devices. This records items 1–7 of [STABILITY_REVIEW.md](STABILITY_REVIEW.md);
item 8, detailed dependency-advisory classification, is **PENDING** until the
target report exists. The mandatory review is not fully closed.

## Confirmed defects and bounded repairs

| Finding | Evidence and repair | Required target check |
| --- | --- | --- |
| Optional chat audio could throw during room events | `chat.newMessage` called `.catch` directly on `Audio.play()` and did not contain synchronous constructor/play failures. `notification-sound.js` guards missing/legacy APIs and catches optional audio failures. Sound remains one element per notification; no encoder or transport change. | New notification test plus real join/chat/control sequence, sound on/off on an affected TV when available |
| Long HLS private pause could terminate a valid player | Both lease authentication and request admission rejected keepalive while paused. The client renews at 15 seconds and treats repeated keepalive errors as terminal; the lease expires after 30 seconds. Only nonblocking keepalive can now renew a valid paused lease and its cookie. Media admission remains denied; expired/revoked leases cannot revive and rate limits remain. | New Go paused-renewal/controller/rate tests and 46-second client pause test; real >=45-second private pause, packager idle, resume and revocation |
| Server option text incorrectly said no client player existed | Phase 3 already supplies the isolated player. The HLS enable flag description now describes experimental shared passive delivery. | Build/config inspection; no behavior change |

MDN documents that older `play()` implementations may return no value:
[HTMLMediaElement.play](https://developer.mozilla.org/en-US/docs/Web/API/HTMLMediaElement/play).
This establishes an API/error-handling defect, **not** the cause of the reported
TV disconnects. Repeated audio allocation remains a device-investigation
candidate. No speculative pooling, bitrate/buffer changes or WebCodecs removal
is included.

## Inspected paths and implications

1. **Room events and stores:** traced member list/join/disconnect, room chat and
   control take/release/grant in `client/src/neko/index.ts` and the user/chat/
   remote stores. The inspected ordinary handlers update room/UI state and
   notifications; they do not directly recreate the peer or detach the selected
   receive backend. Socket-generation guards reject stale callbacks. Actual
   old-device CPU/audio/UI interactions remain unmeasured.
2. **Startup/playback:** inspected login/init, selected-backend peer creation,
   WebRTC readiness/recovery and HLS advertised-mode/ticket/bootstrap/readiness.
   HLS uses native Apple playback or the lazy pinned MSE adapter with exact
   codec checks. There is no automatic fallback. Autoplay retains bounded muted
   retry and the Play gesture. Connection readiness is not first-frame evidence;
   measure the reported mobile startup independently.
3. **Recovery and ownership:** inspected event-socket/peer identity guards,
   WebCodecs worker/socket/audio lifecycle, HLS timers/abort controllers/player
   generations and private-mode buffer clearing. The long-pause defect above
   was the concrete repair. Selected tests exercise stale bootstrap, late Play
   rejection, revoke and teardown; native decoder release/background return
   still require actual-device validation.
4. **Authorization:** inspected common `MediaManager.Open`, exact live
   `CanWatch` session rechecks and lifecycle revocation, HLS security/ticket/
   cookie/request boundaries and the event-level view-only separation. Only
   keepalive renewal is allowed during private pause; no media or control right
   is granted. Repeat the full role/denial matrix and one-second revoke gate.
5. **HTTP and proxy:** inspected original peer preservation, prefix stripping
   before CORS/log classification, strict Origin/forwarded-TLS policy, scoped
   cookies, fixed path redaction and bounded HEAD/gzip/range/object writes.
   Caddy is host-managed; its actual config/version is not yet reviewed. The
   new guide deliberately treats access and runtime/error logs separately and
   avoids unconditional negative flush intervals that prevent cancellation.
6. **Bounds/isolation:** inspected the shared four-worker packager, exact-source
   geometry/format/generation admission, 64-unit provider queues, eight-event
   handoff, generation reset on overflow, memory retention, per-lease/process
   HTTP limits and idle release. Viewers share packaged objects; request writes
   use bounded deadlines without holding the publication lock. Those bounds do
   not establish target CPU/RSS, acceptable mux behavior or healthy-viewer
   isolation; all remain measured acceptance gates.
7. **Configuration/deployment:** the explicit HLS overlay leaves base defaults
   unchanged and preserves adaptive/WebCodecs. Reviewed nested origin/proxy
   fallback, exact unique image tags, pre-stop rollback capture, private
   evidence paths, shell-input quoting and secret-safe metadata output. Viper's
   pinned cast dependency splits environment StringSlice on whitespace, so
   both modes use `hls ll-hls`. Dependencies/encoder profiles remain unchanged.

The resulting source/configuration diff was inspected again before commit,
including the paused-media denial and preserved keepalive rate/expiry bounds,
private output paths and pre-stop rollback behavior. Passing repository review is not target compatibility or full security
acceptance. The new parser fuzz job is bounded request-boundary coverage, not
proof of complete mux/player fuzz coverage.

## Open evidence and next checkpoint

- The supplied target-server output at exact `e55bcd7e` passes all 44 client
  tests, TypeScript/build, all 13 selected Go packages, both 30-second fuzz jobs,
  server/plugin build and uniquely tagged base/Brave image builds. It includes
  the notification and long-private-pause regression tests. Real enabled HLS
  playback and live private-mode/revocation/device checks remain pending.
- The earlier server gate passed 42 client tests and relevant Go/build checks
  at `e85d8568`; images/deployment health passed at `741025c3`. The operator
  subsequently reported that things seem to work, without an exact
  backend/device/role matrix. This is bounded smoke evidence only.
- The target run again reports 20 npm findings, including one critical, and
  saves an audit report with exit code 1. That detailed report has not yet been
  supplied for applicability/repair review; the findings remain unclassified.
  No dependency-security clearance follows from the image gate.
- The colleague's TV remains unavailable. Exact VIDAA/device compatibility,
  event-triggered failure classification and any benefit of HLS are unknown.
- Numeric latency/pacing, independent fMP4 validation, mixed-backend slow-client
  isolation, resource cleanup, live hostile inputs and the accumulated role/
  recovery matrix remain pending. Use [the runbook](HLS_LL_HLS_VALIDATION.md)
  and commit only a sanitized result with explicit omissions.

WebRTC remains the default; WebCodecs remains explicit; HLS starts default-off
without its overlay. No `master` promotion is authorized.
