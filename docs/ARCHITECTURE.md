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

The reconciled lockfile resolves Vue `2.7.14`, TypeScript `5.8.3`, and Vite `6.4.3`; target-server installation, lint and build remain pending.

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

The first two items establish code-level slow-peer isolation: a backpressured peer drops its own samples instead of blocking the capture fan-out. The product outcome still requires target-server validation with simultaneous healthy and throttled viewers.

The current fork relies on Neko's WebRTC server model. Issue #690 alternative media prototypes are not established backends in this fork.

## Deployment/configuration baseline

The default root `config.yml` is no longer copied into the base image. Server defaults now keep implicit hosting and cookie authentication disabled, matching the removed file's effective defaults. Deployments must supply intentional settings through environment variables or a mounted YAML file; the repository `docker-compose.yaml` is an editable example and its NAT address placeholder must be replaced.

The runtime/browser image tree also includes ARM64 Widevine installation, ARM64 Google Chrome image support, updated Chromium-family policies and NVIDIA encoder fallback selection. These image paths have not been built in Codex.

## Local deployment reference

The complete 2026-09-09 filesystem audit found all 724 baseline files in `MyNekoProjekt`: 138 are byte-identical, 585 differ only by CRLF/LF, and only `docker-compose.yaml` differs substantively. The tree also contains 3,928 browser-profile/runtime files and one empty local policy file. No reusable source or repository configuration improvement was missing from the fork.

The excluded local compose setup contains the custom Brave deployment/profile/policy/download mounts, lock cleanup, port mappings, file-transfer configuration and instance credentials.

Only sanitized, reusable deltas belong in Git. See [`LOCAL_DELTA_AUDIT.md`](LOCAL_DELTA_AUDIT.md).

## Target boundaries

### TARGET

- keep authorization/control independent from media transport;
- make weak-viewer behavior peer-local;
- support per-viewer quality selection;
- make view-only roles server-enforced;
- allow an alternative receive-media path without creating a separate room.

### LATER/OPTIONAL

Possible future backends include WebCodecs/WebSocket, WebTransport/QUIC or other receive-only options. Exact final interfaces remain OPEN until baseline sync and prototype evaluation.

## Verification boundary

Architecture/runtime claims beyond repository inspection must be verified on the real target server, not in Codex. Codex should prepare server-side validation steps but must not execute the application, builds, tests, Docker or media/device checks.

The semantic upstream merge is recorded in [`UPSTREAM_SYNC_AUDIT.md`](UPSTREAM_SYNC_AUDIT.md). Its target-server build and regression matrix remain pending.
