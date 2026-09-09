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
- input travels over the WebRTC data channel;
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

The current adaptive-quality follow-up adds bit/s stream-rate accounting aligned with the estimator, a per-pipeline bitrate gauge, and peer-local sample-drop counters labeled by session and media kind. These changes are implemented in the repository but not yet target-server verified.

The non-blocking queue and unlocked fan-out establish code-level slow-peer isolation: a backpressured peer drops its own samples instead of blocking capture dispatch. The drop path now increments `neko_webrtc_track_dropped_samples_total`, making cross-peer behavior distinguishable without trace logs.

The estimator compares Pion's per-peer target against the current stream bitrate. Static review for the adaptive profile found that the latter had been accumulated as encoded bytes/s even though Pion reports bit/s. The capture path now multiplies sample bytes by eight, publishes the result atomically and exports `neko_capture_streamsink_bitrate`. This corrected selection path has not yet been runtime-verified.

The current fork relies on Neko's WebRTC server model. Issue #690 alternative media prototypes are not established backends in this fork.

## Deployment/configuration baseline

The default root `config.yml` is no longer copied into the base image. Server defaults now keep implicit hosting and cookie authentication disabled, matching the removed file's effective defaults. Deployments must supply intentional settings through environment variables or a mounted YAML file.

The repository `docker-compose.yaml` now represents this fork's operator-confirmed deployment baseline: a locally built Brave image with registry pulling disabled, pre-start singleton-lock cleanup, persistent but ignored profile/download paths, optional managed policy, loopback HTTP binding, configurable WebRTC UDP range and enabled file transfer. `.env.example` documents non-secret settings; Compose refuses to resolve while either password is empty. Actual `.env`, profile, downloads and instance policy stay outside Git.

`docker-compose.adaptive.yaml` is a separate opt-in overlay. It mounts `deploy/adaptive-quality.yaml`, whose ordered `high`/`medium`/`low` VP8 definitions activate demand-driven multi-pipeline encoding and the per-peer estimator. Omitting the overlay leaves the validated single-pipeline baseline unchanged. Activation, metrics, resource implications and rollback are in [`ADAPTIVE_QUALITY.md`](ADAPTIVE_QUALITY.md).

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
- make view-only roles server-enforced;
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
        +-- WebCodecs + WebSocket ---- interactive fallback candidate
        |
        +-- HLS / LL-HLS ------------- passive/view-only candidate
        |
        `-- other backend ------------ only when capability evidence requires it
```

The control/session/auth path must remain independent enough that a receive-only backend does not gain control capability. A passive viewer can therefore use HTTP-streaming media while remaining in the same logical Neko room.

This architecture is directionally aligned with upstream issue #371, which explicitly lists `m3u8`/HLS, WebRTC, QUIC and other media backends and proposes selecting them according to user-device, network and server capabilities.

### LATER/OPTIONAL

- WebTransport/QUIC productionization after simpler fallbacks are proven;
- MPEG-DASH where it materially improves passive-client compatibility;
- MJPEG only as an ultra-legacy image-only last resort;
- fully automatic transport/codec selection after explicit capability detection and measured fallback behavior.

Exact final interfaces remain OPEN until the relevant prototype work is designed against the current synchronized baseline.

## Verification boundary

Architecture/runtime claims beyond repository inspection must be verified on the real target server, not in Codex. Codex should prepare server-side validation steps but must not execute the application, builds, tests, Docker or media/device checks.

The semantic upstream merge is recorded in [`UPSTREAM_SYNC_AUDIT.md`](UPSTREAM_SYNC_AUDIT.md). Its applicable target-server build and regression matrix were operator-confirmed on 2026-09-09 after the deployment policy-mount correction. The later adaptive overlay, bitrate-unit correction and new metrics are statically reviewed repository changes whose target-server verification is still pending.
