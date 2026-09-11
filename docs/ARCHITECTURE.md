# Architecture

## Current repository architecture

```text
browser client
    |
    | WebSocket signaling/events
    | WebRTC media + data channel
    v
server/
    +-- session / auth / room APIs
    +-- WebRTC peer/track handling
    +-- capture / GStreamer
    +-- desktop/X11 control
    +-- plugins
            |
            v
shared server-side X11 browser/desktop
```

Supporting trees:

- `apps/` — browser/application images
- `runtime/` — common runtime support
- `client/` — participant UI and protocol client
- `webpage/` — inherited documentation

## Client

### IMPLEMENTED

- Vue `2.7.13`
- TypeScript `~5.8.2`
- Vite `^6.2.3`
- Vuex/typed-vuex
- browser `RTCPeerConnection`

The reconciled lockfile resolves Vue `2.7.14`, TypeScript `5.8.3`, and Vite `6.4.3`; the operator confirmed that the target-server installation, lint and production build passed.

`client/src/neko/base.ts` currently couples WebSocket signaling/session events, WebRTC peer setup, WebRTC tracks and the WebRTC data channel for input.

Current behavior includes:

- normal connection requires `RTCPeerConnection`;
- ICE candidates are exchanged over WebSocket;
- public STUN fallback may be injected if no STUN is configured;
- failed/closed or unresolved disconnected ICE states lead to disconnect/recovery flow;
- an ICE `disconnected` state retains the existing peer for an eight-second self-recovery window;
- after an established peer/socket is irrecoverable, one application-owned timer serializes at most four fresh legacy logins after 1/2/5/10 seconds;
- initial-login failures, explicit logout, demo mode and server-directed disconnects cannot enter that automatic retry path;
- old socket/peer/data-channel callbacks are identity-guarded, buffered ICE candidates are cleared during teardown, and the login component does not start a parallel connection;
- input travels over the WebRTC data channel;
- a compact `#/<16-character-Base64URL-token>` fragment creates a view-only login whose 96-bit credential is sent as a WebSocket subprotocol, not an HTTP request path/query string;
- once the passive member marker arrives, the client renders video/playback controls without room, input, clipboard, file, chat or admin controls;
- the legacy protocol path requests automatic video-pipeline selection; it becomes active only when multiple pipelines and the per-peer bandwidth estimator are configured;
- file-transfer capability messages carry separate non-admin download/upload/delete permissions;
- the optional `openinapp` client/store path sends authorized HTTP(S) chat links to the shared application;
- window focus can resynchronize the clipboard for the active controller.

Therefore a true non-WebRTC viewer fallback is not currently implemented.

## Fork-specific video/touch layer

`client/src/components/video.vue` is a high-conflict/high-value file containing:

- custom player/fullscreen behavior;
- touch detection;
- trackpad cursor/interaction;
- mobile keyboard helper;
- aspect-aware coordinate handling;
- autoplay fallback;
- media/track health monitoring;
- bounded `srcObject` recovery;
- stream/track listener cleanup.

These behaviors require semantic preservation through upstream sync.

The video layer remains separate from connection recovery. It can reassign `srcObject` for a bounded media-element stall without calling `video.load()`, but an assignment is not counted as success until playback progress or track unmute. When a new or recovered stream is available and Safari rejects autoplay even while muted, the existing central Play overlay remains the user-activation fallback; this does not trigger another network reconnect.

## Fork-specific scope

Verified fork changes after merge base `d74052bb844c43a0cc3c2386d083f7505dc483a2` are concentrated in `client/`, including UI, connection, controls, settings, video, stores and client protocol behavior.

## Server

### IMPLEMENTED

- Go module `github.com/m1k1o/neko/server`
- Go version declaration `1.25.0`
- Pion WebRTC
- GStreamer capture
- X11 desktop/input integration
- plugins and HTTP/WebSocket APIs

The integrated upstream server additionally contains:

- bounded, non-blocking sample queues per WebRTC track;
- capture-listener dispatch outside the shared listener mutex;
- multi-pipeline selection driven by each peer's bandwidth estimator;
- H.265 codec and software/VA-API/NVENC pipeline support;
- configurable capture-pointer visibility;
- server-enforced file download/upload/delete permissions;
- an optional host-authorized `openinapp` plugin;
- XInput-device keyboard dispatch for Firefox/GDK3 compatibility.

The compatibility refactor on `testing` adds an internal transport-neutral media layer without adding a client-visible transport:

```text
GStreamer appsink
    -> timestamped types.Sample compatibility input
    -> capture-backed EncodedMediaProvider
    -> bounded MediaSubscription event queue
    -> authorized MediaDeliveryManager lease
    -> WebRTC backend adapter
    -> Pion TrackLocalStaticSample
```

`server/pkg/types/media.go` contains only codec/source descriptions, encoded units, lifecycle events, selectors and provider/delivery interfaces; it imports no Pion type. `server/internal/capture/media.go` adapts the existing demand-driven stream sinks, owns keyframe admission and source changes, and publishes format, unit, discontinuity and end events. GStreamer exposes PTS, DTS, duration and caps-derived dimensions/frame rate; each pipeline creation advances a source generation and sequence. The provider aligns valid GStreamer timestamps with one manager-owned timeline and preserves the Go arrival timestamp only as `CapturedAt` for the existing Pion compatibility field.

`server/internal/media/manager.go` is the only participant-facing delivery registry. It checks the current authenticated session and `CanWatch`, intersects requests with registered backend capabilities, issues a backend/session-scoped lease with no credential access, keeps one primary delivery per session and owns replacement, revocation, shutdown and generic watching state. View-only sessions pass the same receive authorization because their normalized profile retains `CanWatch`; no delivery capability grants input, control, plugins or media publication.

WebRTC remains a compatibility backend as well as the current control-data transport. Its private `CreatePeer` path therefore carries the already-authenticated concrete session only in the call context needed to preserve signaling, data-channel authorization and session events. The adapter has no general session-manager dependency, and neither the generic backend request nor its lease exposes a login/share credential.

The migrated WebRTC track has no sample channel of its own. It consumes the provider subscription directly with the existing effective capacity of two encoded units and drop-new overflow. The existing `neko_webrtc_track_dropped_samples_total` callback remains peer-local, while new low-cardinality `neko_media_*` metrics expose subscription demand, queue observations, deliveries, bytes/units, drops, discontinuities and source generation.

The adaptive-quality follow-up adds bit/s stream-rate accounting aligned with the estimator, a per-pipeline bitrate gauge, and peer-local sample-drop counters labeled by session and media kind. The operator confirmed the server-image build and focused bitrate/nominal-rate/estimator tests, then observed plausible pipeline bit/s values and peer-local metrics during three-viewer target-server runs.

The non-blocking queue and unlocked fan-out establish code-level slow-peer isolation: a backpressured peer drops its own samples instead of blocking capture dispatch. The drop path now increments `neko_webrtc_track_dropped_samples_total`, making cross-peer behavior distinguishable without trace logs.

The estimator compares Pion's per-peer target against the current stream bitrate. Static review for the adaptive profile found that the latter had been accumulated as encoded bytes/s even though Pion reports bit/s. The capture path now multiplies sample bytes by eight, publishes the result atomically and exports `neko_capture_streamsink_bitrate`; the target-server measurements confirmed the corrected unit.

Target-server impairment runs then exposed an asymmetric decision requirement. Downgrade needs to ask whether the current tier still fits, while upgrade needs enough capacity for the next tier. Reusing one 0.15 current-tier threshold allowed premature upward oscillation. The server now has a separate `upgrade_diff_threshold`, defaulting to the old 0.15 behavior for compatibility. Measurement-led constrained-tier rates made both shaped phases mostly usable, but one low-to-medium excursion remained because upgrade still referenced the content-dependent measured current rate. Video pipelines can now declare `nominal_bitrate`; before an upgrade the estimator uses the next tier's stable nominal value when present and retains the current measured-rate fallback when absent. The final target-server rerun held the constrained viewer on `low`, then recovered it through `medium` to `high`; the opt-in profile was accepted for that bounded scenario. The stable deployment and existing configurations remain behaviorally unchanged.

A later target-server trace exposed a separate inherited startup-timing edge. `unstableSince` and `stalledSince` began as Go zero times, so `time.Since(...)` made their configured grace periods appear already elapsed on the first qualifying neutral estimate. Current `testing` starts the stable, unstable and stalled observation windows together when the estimator reader starts; upgrade/downgrade backoff timestamps remain unset until a real switch. This changes no profile value, public API or transport. At exact commit `e5f55bf9`, the complete target-server validation service/build passed and one fresh viewer stayed on `high` for the full focused startup trace with no downgrade, upgrade, stall or video queue drop. This does not imply that `high` must be retained when sustained receiver estimates cannot support it.

A static follow-up found no encoder-quality mutation in the media-subscription refactor: the accepted adaptive Compose/profile files are unchanged since `bfaca84e`, encoder construction/configuration is unchanged, and the same encoded GStreamer payload bytes are passed through the provider into Pion. The reported possibility of softer fast-motion output is compatible with the existing roughly 2-Mbit/s, 25-fps VP8 CBR tier and its full `max-quantizer: 63` range, but remains subjective until a controlled bitrate/quantizer A/B. It is tracked as later tuning rather than attributed to the new boundary.

The current fork still relies on Neko's WebRTC server model. WebRTC is merely the first registered adapter behind the new boundary; issue #690 alternative media prototypes are not established backends in this fork. The first `webcodecs-ws` receive prototype is specified, not implemented, in [`WEBCODECS_MEDIA_WEBSOCKET.md`](WEBCODECS_MEDIA_WEBSOCKET.md).

The view-only follow-up adds a transport-independent `MemberProfile.IsViewOnly` marker and a fixed multi-user share profile. The marker is normalized before login-lock evaluation and whenever sessions are created or updated. Server enforcement then applies at the authenticated HTTP routes, current and legacy WebSocket dispatchers, both WebRTC data-channel formats, inbound media tracks, host assignment and plugin managers. Only heartbeat and receive-media signalling cross the passive WebSocket boundary; only data-channel ping crosses the modern passive data boundary. Passive sessions cannot be persisted or restored.

```text
#/<token> fragment
        |
        | Sec-WebSocket-Protocol: neko-view.<token>
        v
multi-user authentication -> normalized is_view_only session
        |                              |
        | receive signalling/media    `-- deny HTTP/input/plugin/media-publish paths
        v
same shared WebRTC room/desktop
```

This is still a WebRTC receive-media implementation. Its authorization identity is deliberately outside the media transport so a later HLS/LL-HLS subscriber can reuse the same room/session boundary.

## Deployment/configuration baseline

The default root `config.yml` is no longer copied into the base image. Server defaults now keep implicit hosting and cookie authentication disabled, matching the removed file's effective defaults. Deployments must supply intentional settings through environment variables or a mounted YAML file.

The repository `docker-compose.yaml` now represents this fork's operator-confirmed deployment baseline: a locally built Brave image with registry pulling disabled, pre-start singleton-lock cleanup, persistent but ignored profile/download paths, optional managed policy, loopback HTTP binding, configurable WebRTC UDP range and enabled file transfer. `.env.example` documents non-secret settings and the optional `NEKO_VIEW_ONLY_TOKEN`; Compose refuses to resolve while either password is empty. Actual `.env`, share credential, profile, downloads and instance policy stay outside Git.

`docker-compose.adaptive.yaml` is a separate opt-in overlay. It mounts `deploy/adaptive-quality.yaml`, whose ordered `high`/`medium`/`low` VP8 definitions activate demand-driven multi-pipeline encoding and the per-peer estimator. Omitting the overlay leaves the validated single-pipeline baseline unchanged. The overlay was accepted on the target server for the documented desktop/iPad/iPhone scenario on 2026-09-10. Activation, metrics, resource implications, bounded evidence and rollback are in [`ADAPTIVE_QUALITY.md`](ADAPTIVE_QUALITY.md).

The first target-server deployment smoke test exposed a filename regression in the sanitized reconstruction: the external policy was mounted as `/etc/brave/policies/managed/policy.json` instead of replacing the image's `/etc/brave/policies/managed/policies.json`. Restoring the original plural destination made both the custom managed policy and persistent profile load as intended. This initially validated the deployment/mount path; the operator subsequently confirmed the rest of the applicable integration regression matrix.

The runtime/browser image tree also includes ARM64 Widevine installation, ARM64 Google Chrome image support, updated Chromium-family policies and NVIDIA encoder fallback selection. These image paths have not been built in Codex.

## Local deployment reference

The complete 2026-09-09 filesystem audit found all 724 baseline files in `MyNekoProjekt`: 138 are byte-identical, 585 differ only by CRLF/LF, and only `docker-compose.yaml` differs substantively. The tree also contains 3,928 browser-profile/runtime files and one empty local policy file. No application-source improvement was missing. The compose structure was initially excluded with its instance data, then deliberately reconstructed after the operator confirmed it as the desired deployment configuration.

The raw local compose combined the custom Brave deployment/profile/policy/download mounts, lock cleanup, port mappings and file-transfer configuration with instance credentials. Only the sanitized structure is tracked.

Only sanitized, reusable deltas belong in Git. See [`LOCAL_DELTA_AUDIT.md`](LOCAL_DELTA_AUDIT.md).

## Target boundaries

### TARGET

- keep authorization/control independent from media transport;
- make weak-viewer behavior peer-local;
- support per-viewer quality selection;
- preserve the server-enforced view-only role across future media backends;
- allow alternative receive-media paths without creating a separate room;
- allow different participants in the same room to use different media backends when role/device/network capability requires it;
- treat interactive low-latency fallback and passive/view-only streaming as separate compatibility problems.

### Target media-backend shape

The intended boundary is a shared encoded-media/subscription layer feeding peer-specific delivery backends:

```text
shared capture / encoder outputs
        |
        +-- WebRTC -------------------- interactive default
        |
        +-- WebCodecs + WebSocket ---- interactive-class receive candidate
        |
        +-- HLS / LL-HLS ------------- passive/view-only candidate
        |
        `-- other backend ------------ only when capability evidence requires it
```

The control/session/auth path must remain independent enough that a receive-only backend does not gain control capability. A passive viewer can therefore use HTTP-streaming media while remaining in the same logical Neko room.

The concrete boundary is implemented in [`MEDIA_SUBSCRIPTION_BOUNDARY.md`](MEDIA_SUBSCRIPTION_BOUNDARY.md). It distinguishes a backend subscription to an encoded source from the authorized delivery attached to a participant: WebRTC subscribes per participant now; the specified media WebSocket will do the same, while a later HLS packager may subscribe once per active variant and issue separate short-lived viewer leases. The central delivery manager checks `CanWatch`; backends never receive login/share credentials or authority over control, plugins or member profiles.

The version-1 WebCodecs/media-WebSocket specialization is fixed in [`WEBCODECS_MEDIA_WEBSOCKET.md`](WEBCODECS_MEDIA_WEBSOCKET.md): an authenticated event-plane exchange creates a short-lived single-use ticket, the dedicated socket carries strict VP8/Opus records, every server/browser queue is bounded, stale generations are rejected, and audio owns the common presentation clock. Both server enablement and client selection are explicit. WebRTC remains the default and no transport fallback is automatic. The design is receive-only because current high-rate input still uses the WebRTC data channel; this is an explicit prototype boundary, not hidden control-path parity.

The implementation closes the planned gaps in the former WebRTC-facing `types.Sample`/`SampleListener` seam: format metadata, real GStreamer PTS/DTS, a manager-owned timeline, generations and discontinuities are explicit; subscriber queues are bounded and non-blocking; and generic session watching state no longer depends on the name `WebRTC`. Legacy stream-sink types remain capture-internal compatibility machinery, while WebRTC depends only on `EncodedMediaProvider`/`MediaSubscription`. No new transport, public API or configuration was added.

This architecture is directionally aligned with upstream issue #371, which explicitly lists `m3u8`/HLS, WebRTC, QUIC and other media backends and proposes selecting them according to user-device, network and server capabilities.

### LATER/OPTIONAL

- WebTransport/QUIC productionization after simpler fallbacks are proven;
- MPEG-DASH where it materially improves passive-client compatibility;
- MJPEG only as an ultra-legacy image-only last resort;
- fully automatic transport/codec selection after explicit capability detection and measured fallback behavior.

The source-subscription and participant-delivery interface semantics are decided. Version-1 WebCodecs framing, VP8/Opus negotiation, concrete queue bounds and explicit rollout are also decided, while their implementation and target-device evidence remain open. HLS packaging parameters, the broader device/codec matrix and automatic selection remain OPEN until their evidence-led prototype blocks.

## Verification boundary

Architecture/runtime claims beyond repository inspection must be verified on the real target server, not in Codex. Codex should prepare server-side validation steps but must not execute the application, builds, tests, Docker or media/device checks.

The semantic upstream merge is recorded in [`UPSTREAM_SYNC_AUDIT.md`](UPSTREAM_SYNC_AUDIT.md). Its applicable target-server build and regression matrix were operator-confirmed on 2026-09-09 after the deployment policy-mount correction. The adaptive overlay, bitrate-unit correction, diagnostics and next-tier nominal upgrade gate were subsequently built, focused-tested and accepted on 2026-09-10 for the documented three-viewer target-server scenario. The view-only boundary, real inbound-media denial and revocation passed through `80eeca64`; exact-commit client/server checks, deployment, HTTP probe and compact-link browser smoke then passed at `913a981e`. The operator closed that grouped checkpoint without executing the manual iPhone deep test, so no no-reload iPhone claim is made. At `a32027d`, the media-subscription/WebRTC compatibility refactor passed the expanded target-server Go suite and server/plugin build, local image builds, healthy startup, a current-protocol lifecycle probe and initial metrics inspection. At exact follow-up `e5f55bf9`, the full server validation/build passed again, fresh images deployed healthy, and the estimator correction passed its focused single-viewer startup trace. The operator closed that checkpoint on 2026-09-12 while explicitly deferring a repeat of the full role/private-mode/reconnect matrix and induced three-viewer adaptive down/up run to final grouped validation; those checks are not claimed at `e5f55bf9`. Codex did not execute any runtime checks.
