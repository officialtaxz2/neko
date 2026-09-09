# Architecture

This document records architecture verified from the current fork and the boundaries a future coding agent must preserve.

## 1. Repository architecture

```text
browser client
    |
    | WebSocket signaling/events
    | WebRTC media + data channel
    v
server/
    |
    +-- session / auth / room APIs
    +-- WebRTC peer/track handling
    +-- capture / GStreamer
    +-- desktop/X11 control
    +-- plugins
            |
            v
shared server-side X11 browser/desktop

apps/       browser/application image definitions
runtime/    common runtime image/container support
client/     participant UI and legacy Neko protocol client
webpage/    inherited documentation site
```

The product is multi-participant around one shared desktop. Client/media work must not change that invariant without an explicit project decision.

## 2. Client

### IMPLEMENTED

Current client package:

- Vue `2.7.13`
- TypeScript `~5.8.2`
- Vite `^6.2.3`
- Vuex / typed-vuex
- WebRTC through browser `RTCPeerConnection`

The reconciled lockfile currently resolves Vue `2.7.14`, TypeScript `5.8.3`, and Vite `6.4.3`. The manifest ranges above remain authoritative for allowed updates.

Client build:

```bash
cd client
npm ci
npm run lint
npm run build
```

CI uses Node 18.

### Connection/media/control

`client/src/neko/base.ts` currently couples:

- WebSocket signaling/session events,
- WebRTC peer setup,
- WebRTC media tracks,
- WebRTC data channel for low-latency input.

Important current behaviors:

- client refuses normal connection when `RTCPeerConnection` support is absent;
- ICE candidates are exchanged over WebSocket;
- if ICE-lite is not used and no STUN URL is configured, the fork injects public Google STUN servers;
- `connectionState=failed`, `iceConnectionState=failed|closed`, or an 8-second unresolved disconnected state leads to disconnect/recovery flow;
- input events are encoded into the WebRTC data channel.

Therefore a true non-WebRTC viewer fallback is **not implemented** by the current client architecture.

## 3. Fork-specific video/touch layer

`client/src/components/video.vue` is a high-conflict/high-value file.

### IMPLEMENTED

It contains:

- custom player styling and fullscreen handling;
- mobile/touch detection;
- trackpad cursor and relative trackpad behavior;
- mobile keyboard helper integration;
- video-content rectangle calculation for letterboxing/pillarboxing;
- autoplay fallback compatible with mobile browser restrictions;
- media-element health monitoring;
- video-track mute/unmute/ended handling;
- bounded recovery by reassigning `srcObject`;
- deliberate avoidance of `video.load()` on WebRTC streams;
- stream/track listener cleanup.

These behaviors must be compared semantically when upstream also modifies `video.vue`.

## 4. Fork-specific client scope

The verified fork diverged from upstream at merge base:

```text
d74052bb844c43a0cc3c2386d083f7505dc483a2
```

The fork's changes from that base to current `master` are concentrated in `client/`, including:

- build migration/configuration,
- global UI/style work,
- connect/chat/clipboard/control/member/settings/sidebar changes,
- added keyboard helper and member sidebar,
- large video component changes,
- changes in `client/src/neko/base.ts` and `client/src/neko/index.ts`,
- store/settings/video adjustments.

This is why the client is the primary conflict/regression surface for upstream sync.

## 5. Server

### IMPLEMENTED

- Go module: `github.com/m1k1o/neko/server`
- Go version declared: `1.25.0`
- Pion WebRTC stack
- GStreamer capture integration
- X11 desktop/input integration
- plugin architecture
- HTTP/WebSocket APIs inherited from Neko

Build:

```bash
cd server
./build
```

The build script compiles the server and plugins. Server CI builds Docker contexts for amd64 and arm64.

### Current media assumption

The fork currently relies on upstream's WebRTC server model. The project does not yet contain the #690 WebCodecs/WebSocket/WebTransport prototype stack as an established backend.

## 6. Configuration / access model

The repository's current `config.yml` includes:

- desktop screen `1920x1080@60`;
- multiuser provider;
- separate admin/user passwords in the example config;
- `merciful_reconnect: true`;
- `implicit_hosting: false`;
- session cookie disabled for legacy API compatibility.

Do not confuse example/default repository config with production secrets.

## 7. Local deployment reference

The local `MyNekoProjekt` tree is a separate deployment/working reference.

The complete 2026-09-09 filesystem audit found that all 724 baseline project files are present there. Of them, 138 are byte-identical, 585 differ only by CRLF/LF representation, and only `docker-compose.yaml` differs semantically. The tree additionally contains 3,928 browser-profile/runtime files and one empty local policy file. No reusable source or repository configuration improvement was missing from the fork.

The excluded deployment composition includes:

- custom Brave image;
- persistent Brave browser profile;
- managed Brave policy mount;
- persistent downloads;
- cleanup container for Chromium/Brave singleton lock files;
- different HTTP/WebRTC port mapping;
- file-transfer environment settings;
- instance credentials.

Only sanitized, reusable changes belong in Git. Credentials and runtime state do not. See [`LOCAL_DELTA_AUDIT.md`](LOCAL_DELTA_AUDIT.md) for the exhaustive classification rules and counts.

## 8. Upstream architecture relevant to target work

Upstream Neko v3 already documents/contains building blocks relevant to this project's target:

- multiple video pipelines / stream selection;
- experimental bandwidth estimation/adaptive quality;
- WebSocket desktop control fallback;
- modularization goals documented in issue #371.

Recent upstream work after this fork's last sync touches client, capture, file transfer, codecs, pipelines and server internals. Do not recreate upstream fixes before reviewing them.

## 9. Architectural target boundaries

### TARGET

- keep authorization/control independent from the media backend;
- make slow-viewer behavior peer-local;
- support per-viewer quality selection;
- make a view-only role server-enforced;
- allow an alternative receive-media path without requiring a separate room/session.

### LATER/OPTIONAL

A clean media-backend interface may eventually support WebRTC, WebCodecs/WebSocket, WebTransport/QUIC, HLS or other transports. That modular abstraction is directionally aligned with upstream #371, but the exact final interface remains OPEN until after baseline sync and prototype evaluation.
