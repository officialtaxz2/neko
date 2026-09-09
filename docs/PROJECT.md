# Project Definition

Last consolidated: 2026-09-09.

This document is the authoritative product/requirements source for this fork. State labels are normative:

- **IMPLEMENTED** — verified in repository code/config or inherited Neko behavior currently present.
- **TARGET** — desired and decided, but not yet fully implemented.
- **LATER/OPTIONAL** — useful direction, not required for the current implementation sequence.
- **OPEN** — unresolved; do not silently choose an answer.

## 1. Purpose

This fork exists because Neko's shared-session interaction model fits the intended use case well, but the client/media path needs to be substantially more robust across heterogeneous viewers.

The product remains a self-hosted shared server-side browser/desktop, not a collection of independent per-user browser sessions.

## 2. Core product invariants

### IMPLEMENTED / must remain true

- One shared server-side browser/desktop/session is visible to multiple participants.
- Admin and regular-user access exist.
- Shared control is exclusive: at most one participant controls the remote desktop at a time.
- Admins can lock controls for users and retain the ability to administer control.
- Neko remains self-hosted and Docker-oriented.
- WebRTC is the currently implemented primary media path.

### TARGET / must remain true after future rework

- Admin can completely lock control such that no normal user controls the session.
- Admin can grant/revoke control predictably.
- New streaming work must not weaken server-side authorization.
- Mobile/TV compatibility work must not fork the room into separate unsynchronized desktops.
- Healthy viewers must be isolated from weak viewers.
- Viewer quality decisions must be per viewer, not a single globally degraded quality for everyone.

## 3. Current fork-specific implementation

### IMPLEMENTED

Verified in the current fork client:

- Vue 2.7 + TypeScript client, built with Vite.
- Major visual redesign of the legacy client.
- Touch-device detection including coarse pointers and mobile UA cases.
- Trackpad mode and optional trackpad cursor.
- Mobile keyboard button and keyboard-helper component.
- Touch coordinate/aspect-ratio handling around letterbox/pillarbox regions.
- Mobile-safe autoplay logic: unmuted play is attempted, then muted fallback with user-visible unmute/play behavior.
- Stream-health monitoring using media events (`stalled`, `waiting`, `timeupdate`) and track lifecycle events.
- Bounded client-side stream recovery using `srcObject` reassignment; `video.load()` is explicitly avoided for WebRTC recovery.
- Client ICE disconnect timeout and fail/closed handling.
- Public STUN fallback injection in the client when no STUN server is configured.
- Client demo mode for UI/demo behavior.
- Fullscreen handling across standard/WebKit/Mozilla event variants.
- Local settings for trackpad mode, trackpad cursor visibility, forced touch detection, autoplay and scroll behavior.

These are regression-sensitive when synchronizing upstream.

## 4. Target outcomes

### TARGET — highest priority

#### 4.1 Slow-viewer isolation

A slow or lossy viewer must not stall the shared capture/encoder path or degrade other peers.

Expected behavior for a weak viewer:

1. that viewer may drop frames;
2. that viewer may be moved to a lower profile;
3. that viewer may reconnect independently if necessary;
4. healthy viewers remain fluid.

Acceptance requires a controlled multi-client test, not just code inspection.

#### 4.2 Per-viewer adaptive quality

Quality should be independently selectable/adaptive per peer. Desired profile concept:

```text
high-quality viewer      -> 1080p60 / high bitrate
normal viewer            -> 1080p30 or 720p30
constrained viewer       -> 480p/720p lower bitrate
```

Use connection evidence such as bandwidth estimation, loss, RTT/jitter and client capability where available.

Important constraint: upstream already contains multi-pipeline selection and experimental bandwidth-estimation/adaptive-quality work. Stabilize/evaluate that before inventing a parallel ABR architecture.

#### 4.3 Mobile robustness

Mobile must reliably:

- join,
- start video/audio under browser autoplay rules,
- recover from transient media/ICE failures,
- avoid black-screen/reconnect loops,
- provide usable touch/trackpad and keyboard interaction,
- behave predictably across orientation/fullscreen changes.

Existing fork fixes are a starting point, not proof that mobile is fully solved.

#### 4.4 Smart-TV / constrained-browser compatibility

A viewer should not be permanently excluded solely because its WebRTC/ICE/media implementation is unreliable.

A non-WebRTC media fallback is a product target, but its exact protocol is not yet a mandatory implementation choice.

#### 4.5 Server-enforced view-only share link

Target UX:

```text
/watch/<token>
```

or equivalent.

A viewer joining through the view-only path should receive video/audio and optionally permitted passive features, but no mouse, keyboard, touch control, control request, or privileged API ability.

Security invariant: permission must be enforced by the server. `cast`, `embed`, hidden buttons or other client-only UI modes are not sufficient.

#### 4.6 Robust connection recovery

Refresh/reconnect/network transitions must not leave a peer permanently black or stuck in reconnecting state. Recovery must remain bounded and observable.

## 5. Architecture direction

### TARGET

Keep connection/control/media separable enough that alternative media transports can be added without rewriting room authorization or control semantics.

This aligns with upstream issue #371, which explicitly proposes protocol-independent connection/media/control interfaces and multiple media backends.

### LATER/OPTIONAL

- automatic transport selection by device/network/server capability;
- automatic codec selection (for example H.264 vs H.265/VPx based on actual support);
- WebTransport/QUIC media;
- HLS/DASH or other additional receive-only backends;
- broader plugin/general-device abstractions from upstream's long-term v3 concept.

Do not block the near-term sync and robustness work on these.

## 6. Candidate WebRTC fallback work

### TARGET direction, implementation OPEN

Upstream issue #690 proposes a concrete staged prototype:

1. extract encoded-media subscriptions from WebRTC-specific plumbing;
2. add WebCodecs + dedicated WebSocket media;
3. add WebTransport media on top.

The upstream maintainer has publicly stated interest in having an alternative because WebRTC causes issues.

Project decision:

- evaluate the staged work after the fork is fully reconciled and synced;
- prefer the media-abstraction step before importing a transport;
- treat WebSocket + WebCodecs as the first practical fallback candidate;
- treat WebTransport as later/optional until deployment/browser complexity is justified.

None of those #690 prototype PRs are considered IMPLEMENTED in this fork unless merged code proves otherwise.

## 7. UX requirements

### IMPLEMENTED / preserve

- redesigned client appearance and overlays;
- touch-visible controls;
- mobile keyboard helper;
- trackpad-style interaction;
- play/unmute overlays compatible with browser autoplay policy.

### TARGET

- view-only entry should be nearly frictionless for guests;
- recovery and fallback should fail visibly and recoverably, not as a silent black screen;
- quality adaptation should normally be automatic; manual diagnostics/override may exist but is not the primary UX.

## 8. Security and privacy constraints

### TARGET / invariant

- authorization is server-side;
- do not expose admin/control capability through a view-only token;
- do not commit deployment credentials or persistent browser profiles;
- browser profile storage can contain cookies/session secrets and is runtime data;
- local download directories are runtime/user data, not source;
- external STUN/TURN behavior must remain explicit in deployment documentation when changed.

## 9. Deployment

### IMPLEMENTED

The public repository ships the inherited/example Docker-oriented Neko configuration.

A separate local `MyNekoProjekt` deployment tree uses a custom Brave image, persistent Brave profile, custom policy mount, download mount, cleanup of browser singleton locks, non-default local HTTP/UDP ports and file-transfer settings.

### AUDITED boundary

The complete 2026-09-09 audit found no missing reusable source/config improvement. The local compose file, empty policy file, browser profile and downloads remain excluded because they form an instance-specific deployment and include credentials/runtime state. Existing repository documentation already covers the reusable Brave and file-transfer capabilities. See [`LOCAL_DELTA_AUDIT.md`](LOCAL_DELTA_AUDIT.md).

## 10. Non-goals

### Current non-goals

- replacing the shared session with one session per viewer;
- sacrificing existing fork mobile/UX behavior merely to make an upstream merge easy;
- implementing every idea from upstream v3 rewrite before restoring a stable baseline;
- treating a UI-hidden control as authorization;
- copying the local deployment folder wholesale into the public repository.

## 11. External references

Use these as evidence/candidates, not as project truth:

- Upstream repository: https://github.com/m1k1o/neko
- Upstream v3 rewrite / modular backends: https://github.com/m1k1o/neko/issues/371
- WebCodecs/WebSocket/WebTransport proposal: https://github.com/m1k1o/neko/issues/690
- Mobile trackpad issue: https://github.com/m1k1o/neko/issues/640
- Upstream releases: https://github.com/m1k1o/neko/releases
