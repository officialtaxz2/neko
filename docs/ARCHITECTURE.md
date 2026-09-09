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
- input travels over the WebRTC data channel.

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

The current fork relies on Neko's WebRTC server model. Issue #690 alternative media prototypes are not established backends in this fork.

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
