# Architecture

The fork runs upstream's Neko server (Go, Pion WebRTC, GStreamer capture, X11 input) and upstream's React
client on the v3 API. This file describes what the fork adds or changes and the rules it must keep. Feature
overview and deployment: [`README.md`](../README.md); client code: [`client/src/fork/README.md`](../client/src/fork/README.md).

## Invariants

- One shared browser and session for all participants, not one per viewer.
- At most one participant controls the desktop; admin lock, grant and revoke keep working.
- View-only is enforced by the server (`MemberProfile.IsViewOnly`); hiding controls is not authorization.
- A slow or lossy viewer may lose its own frames or tier, but must not slow down other viewers.
- Media transport does not decide authorization: WebRTC and HLS viewers share the same room and checks.
- WebRTC is the default; HLS/LL-HLS is chosen manually and is receive-only. Nothing falls back on its own.
- Touch, trackpad, fullscreen, autoplay and playback recovery on phones and tablets are regression-sensitive.

## Event socket

The browser logs in with `POST /api/login` and keeps one WebSocket on `/api/ws` for session events,
WebRTC signalling, chat, control and HLS negotiation. Mouse and keyboard input goes over the WebRTC data
channel. Each socket writes through one queued `utils.WebSocketWriter` (FIFO, at most 128 messages or
16 MiB including the one being written, 5-second write deadline); a client that falls behind is
disconnected instead of blocking room broadcasts. JSON encoding still happens on the sender's goroutine.

## Media delivery

```text
GStreamer pipelines (one per tier, started on demand)
  -> capture provider (server/internal/capture/media.go)
       encoded units with PTS/DTS, format, generation, discontinuities
  -> bounded subscriptions, overflow drops the newest unit for that subscriber only
  -> media delivery manager (server/internal/media/manager.go)
       checks the live session and CanWatch, one primary delivery per session
  -> backend: WebRTC (one subscription per viewer) or HLS (shared, see HLS_LL_HLS.md)
```

- `server/pkg/types/media.go` defines sources, encoded units, events, subscriptions and the backend
  interfaces without Pion types.
- The provider starts an encoder on its first subscriber and stops it after the last. It admits a new
  video subscriber at a keyframe and forces a keyframe when a tier gains a waiting viewer (rate-limited,
  see `ADAPTIVE_QUALITY.md`). Recreating a pipeline or changing its format advances the source generation.
- The manager issues a lease per session and backend. Logout, kick, deletion, loss of `CanWatch`, a new
  primary delivery and shutdown revoke it centrally. Backends never see login or share credentials.
- A WebRTC track consumes its subscription directly with a queue of two units. A full queue drops that
  viewer's unit and counts it in `neko_webrtc_track_dropped_samples_total{session_id,kind}`.
- `neko_media_*` metrics cover deliveries, subscriptions, queue depth, drops, discontinuities and source
  generation, without session labels.

Per-viewer tier selection (`server/internal/webrtc/peer.go`) is described in
[`ADAPTIVE_QUALITY.md`](ADAPTIVE_QUALITY.md), the HLS backend (`server/internal/mediahls/`) in
[`HLS_LL_HLS.md`](HLS_LL_HLS.md).

## Client mechanics

- `signal/request` asks for `video: {auto: true}`, so the server can switch tiers per viewer.
- Public STUN servers are added only when the server lists no ICE servers.
- ICE `disconnected` has an eight-second grace before it counts as failed, matching the server.
- Playback watchdog (`client/core/src/playback.ts`): no new frames or no time progress for eight seconds
  in the foreground re-attaches the same stream, at most three times, then stops. A hidden page, a
  deliberate pause or a running re-attachment is not a stall.
- Reconnect is upstream's: ten attempts with a whoami check, none after kick or logout. A page restored
  from the back/forward cache reconnects.
- The overlay draws no second cursor when `system/init` reports `webrtc.pointer_in_video`.
- Stream choice (`client/src/state/media.ts`): the stored `media_backend` or `?media=hls|ll-hls` for one
  visit, among the modes the server advertises. A change reloads the page. With HLS, `client.receiveOnly`
  turns off control requests, implicit hosting and all overlay input.
- `?debug` shows stream, connection, watchdog and input state over the picture (`client/src/state/debug.ts`).
  `window.__nekoTrace` is a read-only handle for `deploy/trace-*.js`.

## View-only share links

`member.multiuser.view_only_token` (`NEKO_VIEW_ONLY_TOKEN` in `.env`) is a 96-bit bearer token written as
exactly 16 Base64URL characters, different from both passwords. Empty disables the link.

- The link is `https://host/#/<token>`. The client (`client/src/state/viewonly.ts`) sends the token only as
  the password in the body of `POST /api/login`, so it never appears in a request line, proxy log or
  `Referer`. It stays visible in browser history and anything the link is pasted into.
- The provider compares it in constant time and returns `NewViewOnlyMemberProfile`: watch and receive
  chat only. `RestrictViewOnly` removes admin, host, media sharing and clipboard whenever a profile is set.
- Enforcement is on the server: the WebSocket accepts only heartbeat, WebRTC signalling and HLS
  negotiation (`internal/websocket/viewonly.go`), the data channel only ping, and every interactive HTTP
  route, handler and plugin checks `IsInteractive()`. Inbound media tracks and host assignment are refused.
- View-only sessions are not written to `session.file`. The client keeps the session token in
  `sessionStorage` (`neko_view_session`), apart from a member's session, and logs out on leaving the page.
- The token has no expiry. To revoke it, change or clear it in `.env` and recreate the container
  (`bash setup.sh --no-build`); hand out a new link only if sharing should continue.
- `deploy/check-view-only-http.py` checks the HTTP boundary via `docker-compose.validation.yaml`.

## Open items

- An older Hisense Smart TV (VIDAA) of a regular viewer lost WebRTC around joins, chat and control
  changes; model, firmware and cause are unknown. HLS on that TV is untested.
- Recovery on an iPhone after a network interruption without reload has not been tested on a real device.
- Upstream's capture subscription API (#699) is not adopted; see `UPSTREAM_SYNC_AUDIT.md`.
