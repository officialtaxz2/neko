# Project Definition

Last consolidated: 2026-09-11.

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
- A backend-neutral, server-enforced passive/view-only session identity.
- Self-hosted, Docker-oriented Neko deployment.
- WebRTC as the primary implemented media path.

### TARGET

- Weak viewers do not degrade healthy viewers.
- Quality adapts independently per viewer.
- Mobile and Smart-TV/browser robustness.
- Target-server validation and future alternate-media support for view-only guest/share access.
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
- Bounded application-level reconnect for a previously established real session after the existing peer/socket cannot recover: four serialized attempts after 1, 2, 5 and 10 seconds.
- Automatic reconnect suppression for initial-login failure, explicit logout, demo mode and server-directed disconnects, preserving authentication/session intent.
- Identity guards and teardown for stale WebSocket, peer, data-channel, ICE-candidate and media-stream callbacks before a replacement connection becomes authoritative.
- Public STUN fallback injection if no STUN URL is configured.
- Demo mode.
- Cross-browser fullscreen handling.
- Optional 96-bit view-only share credential encoded as exactly 16 Base64URL characters and transported by a compact `#/<token>` fragment and WebSocket subprotocol rather than an HTTP request path/query string.
- Fixed view-only profile/session normalization plus HTTP, current/legacy WebSocket, modern/legacy data-channel, inbound-media, host-assignment and plugin enforcement.
- Non-persisted passive sessions with documented token rotation/removal and service recreation as the hard revocation boundary.

## Integrated upstream implementation

### IMPLEMENTED in `master` (upstream merge `4e99b8d3`)

- Per-peer WebRTC sample queues are bounded and non-blocking; a full peer queue drops that peer's sample instead of blocking capture dispatch.
- Capture listener dispatch no longer holds the shared listener lock while writing samples.
- Existing multi-pipeline selection and per-peer bandwidth-estimator support is now requested automatically by the legacy protocol path when the operator configures multiple pipelines and enables the estimator.
- H.265 is available as a capture/WebRTC codec, including software, VA-API and NVENC pipeline construction.
- File transfer supports separately configured non-admin download, upload and delete permissions plus bulk deletion.
- The optional `openinapp` plugin can open HTTP(S) chat links in the shared application for an authorized host.
- Capture-pointer visibility, clipboard resync on window focus, `?scroll=` sensitivity and several browser/runtime compatibility fixes are present.

These repository implementation claims were first established by static inspection. On 2026-09-09, the operator confirmed that all checks applicable to the target deployment in the documented build and regression matrix passed after the Brave policy-mount correction. No universal support is claimed for devices or architectures not identified in that operator report.

## Adaptive-quality follow-up

### IMPLEMENTED in the repository / target-server accepted for the documented scenario

- The Brave deployment has a separate opt-in adaptive-quality overlay with ordered `high`, `medium` and `low` VP8 pipelines; the stable base Compose file remains single-pipeline.
- Encoded stream rates are measured and compared with the Pion target in bit/s.
- Peer-local audio/video queue drops and per-pipeline bitrates are exported as Prometheus metrics.
- Upgrade headroom is configured separately from the current-tier downgrade margin; the opt-in profile can therefore account for its approximately 2:1 adjacent tiers without changing the compatible server default.
- Optional `nominal_bitrate` metadata gates an upgrade against the next tier's target; configurations without it retain the previous measured-current-tier fallback.
- Activation, diagnostics, resource costs, rollback and the target-server acceptance sequence are documented in [`ADAPTIVE_QUALITY.md`](ADAPTIVE_QUALITY.md).

On 2026-09-10, the operator accepted commit `bfaca84e` on the target server for one desktop, one iPad and one iPhone. The isolated 0.7 Mbit/s rerun held the iPhone on `low` after downgrade, preserved both healthy viewers on `high` with zero peer-local drops, and recovered through `medium` to `high` within 45 seconds after impairment removal. Refresh/rejoin and a transient cellular interruption also preserved both healthy viewers. This is bounded evidence for that deployment and device set, not a universal profile guarantee.

## Highest-priority target outcomes

### TARGET — slow-viewer isolation

The integrated per-peer non-blocking delivery mechanism is the intended code-level isolation: a slow/lossy viewer may drop its own samples without blocking healthy peers. Repeated target-server impairment runs kept both healthy viewers smooth on `high`, with zero new peer-local audio/video drops, while only the constrained viewer changed tiers and stalled. This validates isolation for the reported three-device scenario, not for untested devices or architectures.

### TARGET — per-viewer adaptive quality

The repository contains a reproducible, opt-in three-tier VP8 profile, estimator diagnostics, resource/rollback documentation and an exact healthy-plus-constrained-viewer acceptance procedure. Target-server measurements confirmed the corrected bit/s accounting; focused bitrate, nominal-rate and estimator tests passed. Lower constrained-tier rates, the full VP8 quantizer range and next-tier nominal upgrade gating made both shaped phases acceptably usable for the tested iPhone while preserving the healthy desktop and iPad. The profile was accepted for this bounded target-server scenario on 2026-09-10. The stable single-pipeline deployment remains the default. See [`ADAPTIVE_QUALITY.md`](ADAPTIVE_QUALITY.md).

### TARGET — mobile robustness

Mobile must reliably join, start media under autoplay rules, recover from transient failures, avoid persistent black screens, and retain usable touch/keyboard behavior.

The repository now separates the bounded recovery states: eight seconds for automatic recovery of an existing ICE peer; a four-attempt application-level login when the established peer/socket is unrecoverable; and the existing central Play control when Safari blocks playback after media is available. A successful media-element `srcObject` assignment is no longer treated as recovery until playback progress or track unmute occurs, and old stream timers/listeners are removed when the connection store resets.

The 2026-09-10 iPhone check proved successful rejoin after reload and, when required by iOS autoplay policy, a central Play tap; it also proved that this recovery did not disturb the other viewers. The new no-reload path is implemented and statically reviewed on `testing`, and its dependency-free recovery tests passed in the target-server validation container at `913a981e`. On 2026-09-11 the operator deliberately closed the checkpoint without the manual same-peer, replacement-session and bounded-exhaustion iPhone phases because no Safari Web Inspector/Mac was available and accepted the automated evidence for the deployment decision. Fully automatic in-place recovery therefore remains an unverified device claim; [`IOS_RECOVERY.md`](IOS_RECOVERY.md) retains the deferred procedure.

### TARGET — Smart-TV / constrained browser compatibility

A viewer should not be permanently excluded solely because WebRTC/ICE/media support is unreliable. Non-WebRTC receive-media paths are a target, and they do not need to use the same transport as interactive clients.

For passive/view-only devices such as Smart-TVs or constrained/older browsers, higher media latency is acceptable when it materially improves compatibility and stability. A conventional HTTP live-streaming path (preferably HLS/Low-Latency HLS as the first candidate, with DASH as an alternative to evaluate) should therefore be considered separately from low-latency interactive fallbacks.

### IMPLEMENTED IN REPOSITORY — server-enforced view-only share link

The optional `#/<16-character-Base64URL-token>` link permits passive WebRTC viewing in the same room without mouse, keyboard, touch, clipboard, file, microphone/media-share, control-request or admin capabilities. The token remains in the URL fragment and is carried in a WebSocket subprotocol; it is never placed in the normal share-link HTTP request path or query string. Six-character and legacy 64-hex credentials are rejected.

Authorization is enforced by a normalized backend-neutral `is_view_only` profile and allow/deny checks across every current ingress path, not by hidden controls or by WebRTC. This leaves the same identity usable by a future different receive-only media backend without creating a separate desktop/session or weakening authorization. Token lifetime, revocation, denial behavior, deployment, rollback and the three-role target-server matrix are specified in [`VIEW_ONLY_SHARING.md`](VIEW_ONLY_SHARING.md).

Runtime status: **the full view-only boundary, real inbound-media denial and revocation passed on the target server through `80eeca64`. At exact commit `913a981e`, fresh client/server checks, image deployment, the HTTP boundary probe and the compact-link browser smoke test also passed. View-only is accepted for the tested deployment; the separately omitted iPhone recovery deep test is not evidence against or for its no-reload behavior**.

### TARGET — robust recovery

Refresh, reconnect and network transitions must not leave peers permanently black or stuck.

## Alternative media direction

### DESIGNED — backend-neutral encoded-media subscription boundary

The source-subscription and participant-delivery contract is fixed in [`MEDIA_SUBSCRIPTION_BOUNDARY.md`](MEDIA_SUBSCRIPTION_BOUNDARY.md). It separates demand on shared encoded sources from per-session delivery, centralizes `CanWatch` authorization above all backends, requires bounded non-blocking queues and makes format, GStreamer presentation timing, source generations and discontinuities explicit.

This is an architecture design, not a source implementation. The existing WebRTC sender has not yet been migrated and no alternative endpoint, packager or client exists.

The project distinguishes **interactive** and **passive/view-only** media fallback needs. There is not one mandatory fallback chain for every client.

This direction is consistent with upstream issue #371, which proposes protocol-independent media backends including HLS (`m3u8`), WebRTC and QUIC, with backend selection based on device capabilities, network conditions and server capabilities.

### TARGET candidate — interactive fallback

Upstream issue #690 proposes staged work around:

1. media-subscription abstraction;
2. WebCodecs + dedicated WebSocket media;
3. WebTransport.

For an interactive participant, WebCodecs + WebSocket remains the first practical non-WebRTC candidate to evaluate after the media abstraction. Its purpose is primarily to bypass WebRTC/ICE/browser compatibility failures while retaining a low-latency path.

Do not assume WebSocket is automatically better on a poor or lossy connection: the standard `WebSocket` API has no built-in backpressure, and reliable ordered delivery can accumulate latency if the receiver cannot keep up. Queueing/drop policy, codec support and actual device behavior must be measured.

### TARGET candidate — passive/view-only HTTP streaming

For passive viewers, especially Smart-TVs, old/constrained browsers, and share-link clients, evaluate a conventional adaptive HTTP livestream independently of the interactive path:

- **HLS / Low-Latency HLS** is the primary candidate because it is HTTP-based, supports live audio/video and multiple bitrate variants, and is designed to adapt playback to changing network conditions.
- **MPEG-DASH** is a secondary candidate where client/platform support makes it useful.
- Higher latency is acceptable for this role because these clients are not controlling the desktop.
- A passive HTTP-stream client must remain part of the same room/session and receive no implicit control rights.
- Prefer using the same underlying capture/encoded-media abstraction where practical rather than introducing a second independent desktop capture.

This is conceptually similar to a conventional Twitch/YouTube-style viewer delivery path, not a requirement to reproduce either service's exact protocol stack.

### LATER/OPTIONAL — ultra-legacy image fallback

MJPEG may be evaluated only as a last-resort, ultra-legacy image-only fallback if a concrete device class cannot use the other media paths. It is not a primary architecture target because it lacks an integrated audio/adaptive-streaming model and is bandwidth-inefficient for continuous desktop video.

### Selection principle

The eventual media backend may differ per participant based on role and capability:

```text
interactive capable client  -> WebRTC
interactive WebRTC failure  -> WebCodecs/WebSocket candidate
passive/view-only client     -> HLS/LL-HLS candidate
other constrained client     -> capability-tested alternative
```

Automatic selection is a TARGET direction, not currently IMPLEMENTED. Manual/explicit selection may be used first for diagnosis and rollout.

None of these alternative-media candidates are IMPLEMENTED here unless repository code proves otherwise. The decided boundary design does not change that status.

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

The supplied `MyNekoProjekt` snapshot was completely compared with the fork baseline on 2026-09-09. It contained no missing application-source improvement. After the operator clarified that the differing compose defines the fork's intended deployment, its reusable structure was reconstructed in tracked, parameterized form. Credential values, the empty instance policy, browser profile and downloads remain outside Git. See [`LOCAL_DELTA_AUDIT.md`](LOCAL_DELTA_AUDIT.md).

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
