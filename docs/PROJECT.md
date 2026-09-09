# Project Definition

Last consolidated: 2026-09-09.

State labels:

- **IMPLEMENTED** — verified in repository code/config or inherited behavior currently present.
- **TARGET** — desired and decided, but not fully implemented.
- **LATER/OPTIONAL** — not required for the current implementation sequence.
- **OPEN** — unresolved.

## Purpose

This fork keeps Neko's shared server-side browser/desktop model while improving reliability across heterogeneous viewers.

The product remains one shared session, not independent per-user browser sessions.

## Core invariants

### IMPLEMENTED / preserve

- One shared browser/desktop/session for multiple participants.
- Admin and regular-user access.
- At most one controller at a time.
- Admin lock/grant/revoke semantics.
- Self-hosted, Docker-oriented Neko deployment.
- WebRTC as the primary implemented media path.

### TARGET

- Weak viewers do not degrade healthy viewers.
- Quality adapts independently per viewer.
- Mobile and Smart-TV/browser robustness.
- Server-enforced view-only guest/share access.
- Bounded, observable reconnect/recovery.
- Media/control/auth remain separable enough to support alternative media paths.

## Verified fork-specific implementation

### IMPLEMENTED

- Vue 2.7 + TypeScript client built with Vite.
- Major client/UI redesign.
- Touch-device detection.
- Trackpad mode and optional trackpad cursor.
- Mobile keyboard and keyboard-helper integration.
- Letterbox/pillarbox-aware touch coordinate handling.
- Mobile autoplay muted fallback.
- Media-element stall/waiting/timeupdate monitoring.
- Track mute/unmute/ended monitoring.
- Bounded `srcObject`-based stream recovery.
- Explicit avoidance of `video.load()` for WebRTC recovery.
- ICE failure/disconnect timeout handling.
- Public STUN fallback injection if no STUN URL is configured.
- Demo mode.
- Cross-browser fullscreen handling.

## Integrated upstream implementation

### IMPLEMENTED in `integration/upstream-20260909`

- Per-peer WebRTC sample queues are bounded and non-blocking; a full peer queue drops that peer's sample instead of blocking capture dispatch.
- Capture listener dispatch no longer holds the shared listener lock while writing samples.
- Existing multi-pipeline selection and per-peer bandwidth-estimator support is now requested automatically by the legacy protocol path when the operator configures multiple pipelines and enables the estimator.
- H.265 is available as a capture/WebRTC codec, including software, VA-API and NVENC pipeline construction.
- File transfer supports separately configured non-admin download, upload and delete permissions plus bulk deletion.
- The optional `openinapp` plugin can open HTTP(S) chat links in the shared application for an authorized host.
- Capture-pointer visibility, clipboard resync on window focus, `?scroll=` sensitivity and several browser/runtime compatibility fixes are present.

These are repository implementation claims based on static inspection. Build, runtime, multi-user, media and device verification for the integration commit is still pending on the target server.

## Highest-priority target outcomes

### TARGET — slow-viewer isolation

The integrated per-peer non-blocking delivery mechanism is the intended code-level isolation: a slow/lossy viewer may drop its own samples without blocking healthy peers. The outcome remains a target acceptance criterion until it is demonstrated with simultaneous healthy and throttled viewers on the target server.

### TARGET — per-viewer adaptive quality

Configure and evaluate the integrated multi-pipeline and per-peer bandwidth-estimation path before inventing a parallel ABR architecture. The estimator is not sufficient merely by being present in code; automatic peer-local switching must be verified under controlled bandwidth changes.

### TARGET — mobile robustness

Mobile must reliably join, start media under autoplay rules, recover from transient failures, avoid persistent black screens, and retain usable touch/keyboard behavior.

### TARGET — Smart-TV / constrained browser compatibility

A viewer should not be permanently excluded solely because WebRTC/ICE/media support is unreliable. A non-WebRTC media fallback is a target; exact protocol remains open.

### TARGET — server-enforced view-only share link

A future `/watch/<token>`-style path or equivalent should permit passive viewing without mouse/keyboard/touch/control/admin capabilities.

Authorization must be enforced server-side.

### TARGET — robust recovery

Refresh, reconnect and network transitions must not leave peers permanently black or stuck.

## Alternative media direction

Upstream issue #690 proposes staged work:

1. media-subscription abstraction;
2. WebCodecs + WebSocket;
3. WebTransport.

Project direction:

- evaluate only after local reconciliation and upstream sync;
- prefer abstraction first;
- treat WebSocket + WebCodecs as the first practical fallback candidate;
- keep WebTransport later/optional until justified.

These prototypes are not IMPLEMENTED here unless repository code proves otherwise.

## Development / verification environment constraint

### Workflow invariant

Codex is an editing/static-review environment only and is not representative of the deployment server.

Codex may inspect source, configuration, history and diffs, but must not install dependencies, run builds/tests/linters, start the application, invoke Docker/project scripts, or perform WebRTC/device runtime tests.

Runtime/build/test verification occurs separately on the real server. Until server results exist, changes may be described as statically reviewed but not runtime-verified.

## Security / deployment constraints

- authorization remains server-side;
- do not expose control/admin capability through a view-only token;
- do not commit credentials or persistent browser profiles;
- downloads/browser profiles/cookies are runtime/user data;
- local deployment differences must be sanitized before any reusable example is committed.

### AUDITED local snapshot

The supplied `MyNekoProjekt` snapshot was completely compared with the fork baseline on 2026-09-09. It contained no missing reusable source/config improvement. Its compose override, empty policy file, browser profile and downloads remain outside Git because they are instance-specific and include credentials/runtime state. Existing repository material already covers the reusable Brave and file-transfer mechanisms. See [`LOCAL_DELTA_AUDIT.md`](LOCAL_DELTA_AUDIT.md).

## Non-goals

- one session per viewer;
- discarding fork UX/mobile behavior merely to simplify upstream merging;
- implementing every upstream v3 idea before restoring a stable baseline;
- treating UI-hidden controls as authorization;
- copying the local deployment tree wholesale into the public repository.

## External references

- upstream: https://github.com/m1k1o/neko
- v3 architecture direction: https://github.com/m1k1o/neko/issues/371
- alternative media proposal: https://github.com/m1k1o/neko/issues/690
- mobile trackpad issue: https://github.com/m1k1o/neko/issues/640
