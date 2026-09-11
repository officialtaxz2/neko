# WebCodecs plus dedicated media WebSocket contract

Status: **design complete on `testing` on 2026-09-12; no media-WebSocket endpoint, backend registration, client decoder, deployment enablement or automatic fallback is implemented**.

This document fixes the version-1 contract and the bounded implementation and acceptance plan for Neko's first non-WebRTC receive-media prototype. It specializes the backend-neutral boundary in [`MEDIA_SUBSCRIPTION_BOUNDARY.md`](MEDIA_SUBSCRIPTION_BOUNDARY.md) without changing that boundary or the current WebRTC implementation.

The prototype name is `webcodecs-ws`. The wire-protocol and WebSocket subprotocol name is `neko.media.v1`.

## Scope and non-goals

Version 1 is an explicit, opt-in, low-latency **receive-media** experiment:

- one dedicated authenticated WebSocket carries encoded audio and video from server to browser;
- WebCodecs decodes VP8 video and Opus audio in a dedicated worker;
- the existing authenticated event WebSocket remains the session, authorization, chat and control plane;
- the existing WebRTC path, signaling and data channels remain the default and remain unchanged;
- no backend is chosen automatically and no failure automatically changes transports;
- HLS/LL-HLS remains a separate later passive/view-only prototype.

The current client sends high-rate keyboard, pointer and touch input through the WebRTC data channel. Version 1 deliberately does not add a replacement control transport and therefore must not be described as complete non-WebRTC interactive parity. It proves an interactive-class receive path. A later, independently reviewed block must decide how a controlling participant sends high-rate input when no WebRTC peer connection exists.

This design block adds documentation only. It does not add a route, event, configuration key, backend, decoder, worker, audio worklet, dependency or deployment overlay.

## Invariants

An implementation conforming to this contract must preserve all of the following:

1. `CanWatch` is checked by the server against the current live session object. A browser capability claim or possession of an old ticket is not authorization.
2. The media backend never receives a password, login token, view-only token, session cookie, bearer value or the temporary attach ticket.
3. At most one primary media delivery is active for a session. A successful replacement closes the previous delivery through the existing media manager.
4. Media delivery does not grant `CanControl`, `CanHost`, `CanAdmin` or clipboard/file privileges. Existing input authorization remains authoritative.
5. Capture and other viewers never block on a slow WebSocket or slow decoder.
6. Frames from an obsolete generation are never decoded or rendered.
7. The stable deployment does not expose or advertise the prototype. WebRTC remains the default when the opt-in flag or explicit client selection is absent.
8. Existing `/api/ws`, WebRTC signaling, WebRTC data-channel opcodes, REST APIs, configuration keys and base Compose files keep their current behavior.

## Explicit selection and lifecycle

The implementation block may introduce both of these opt-ins, with the names fixed here:

- server configuration `media.webcodecs_ws.enabled`, default `false`;
- client URL selection `?media=webcodecs-ws`.

Both must be present. A selected client must not also start the normal `signal/request` flow. An unselected client must not send prototype capability or ticket events. Disabling the server flag must leave the new route unregistered and the backend unadvertised.

An explicitly selected client that cannot use the prototype shows a terminal explanation plus two deliberate actions:

- **Retry WebCodecs**, which retries the same selected backend; and
- **Use WebRTC**, which stops any prototype delivery, removes the selection for that navigation and starts or reloads into the unchanged WebRTC path.

There is no automatic WebRTC-to-WebSocket or WebSocket-to-WebRTC fallback. In particular, an ICE failure must not silently select this backend, and a WebCodecs or media-WebSocket failure must not silently create a peer connection.

## Authenticated delivery creation

### Control-plane negotiation

Ticket creation uses the already authenticated event WebSocket. It does not put a durable credential in JavaScript solely for the media backend.

1. The explicitly selected browser sends `media/capabilities/request` on its authenticated event WebSocket.
2. The server returns `media/capabilities` with protocol version `1`, the enabled backend name, the available VP8 and Opus source formats, and the allowed source selectors.
3. The browser probes the exact proposed audio and video configurations with `VideoDecoder.isConfigSupported()` and `AudioDecoder.isConfigSupported()`.
4. The browser sends `media/create` with the exact audio and video choices it successfully probed. Audio may be explicitly disabled; the server must not silently remove a requested kind.
5. The session/controller layer re-resolves the current session, requires `CanWatch`, validates the choices against server-owned capabilities and creates one pending attach ticket.
6. The server returns `media/offer` over the authenticated event WebSocket. It contains the protocol name, relative path `/api/media/ws`, ticket and expiry.

All four new event types must be excluded from payload logging in both current and legacy WebSocket adapters. Unknown fields are rejected. Codec names, source IDs and selectors are accepted only from server-owned allowlists.

### Attach ticket

The attach ticket has this exact contract:

- 24 cryptographically random bytes encoded as 32 unpadded Base64URL characters;
- 10-second lifetime measured with a monotonic deadline where available;
- stored server-side only as a SHA-256 digest;
- atomically single-use;
- bound to the session ID, backend name, protocol version and normalized delivery request;
- at most one pending ticket per session; creating another invalidates the former ticket;
- not bound to a remote IP address, because a mobile client may legitimately change its network path;
- never written to a URL, query string, cookie, log, metric or backend request.

The browser opens `/api/media/ws` with these requested WebSocket subprotocol values, in order:

```text
neko.media.v1
neko.media.ticket.<32-character-ticket>
```

The upgrade handler selects and echoes only `neko.media.v1`. It must never echo the ticket-bearing value. The route accepts no query parameters.

Before upgrading, the server must validate the exact allowed `Origin`, both subprotocol values, ticket syntax, digest, expiry, unused state, binding, the current live session object and current `CanWatch`. Ticket redemption is atomic. Only then may the controller call `MediaDeliveryManager.Open` with the normalized request.

The manager supplies the backend only the existing credential-free `MediaLease` and normalized media request. Ticket creation alone does not replace an active delivery. A successfully attached delivery follows the manager's existing single-primary replacement rule.

### Ready transition

After attachment, the server sends the requested `FORMAT` records. The browser performs exact capability probes once more, configures both requested decoders, and sends one `ready` control record within 5 seconds. Provider subscriptions start only after this transition, and `IsWatching` becomes true only when the lease is active.

If format validation, decoder configuration or the deadline fails, the new delivery closes without affecting capture or another session. The implementation must not hold two primary deliveries open to manufacture fallback.

### Revocation and replacement

The ticket and active delivery are re-evaluated at separate boundaries. Any of these events invalidates pending tickets and closes the active `webcodecs-ws` delivery:

- logout or session deletion;
- `CanWatch` changing to false;
- replacement by another successful primary media delivery;
- explicit media stop;
- media socket close;
- server shutdown;
- final loss of the owning event-WebSocket session after its existing reconnect grace period.

View-token rotation is enforced by the existing service/session replacement behavior. Private mode pauses the provider subscriptions through the generic delivery interface and resumes with a discontinuity and new generation. A view-only session may receive media but remains unable to send input under the existing server checks.

## Codec and capability contract

Version 1 intentionally has one narrow interoperable codec set:

| Kind | WebCodecs codec string | Encoded payload | Required configuration |
| --- | --- | --- | --- |
| video | `vp8` | exactly one VP8 frame | `codedWidth`, `codedHeight`, `optimizeForLatency: true`; no description |
| audio | `opus` | exactly one raw Opus packet | `sampleRate: 48000`, `numberOfChannels: 2`; no description |

The server advertises the actual dimensions, frame-rate ratio, channels and clock rate from the selected `MediaSource`. The browser must probe the exact format record, not only the generic codec name. A positive browser report is advisory; the server still validates the request.

If the target deployment is configured for another capture codec, `webcodecs-ws` is unavailable for that kind and the explicit selection fails cleanly. It must not transcode, silently substitute a codec or change the deployment's capture configuration. H.264 is deferred until Annex-B versus AVC framing, decoder-description bytes and profile/level negotiation have an equally exact contract. VP9, AV1, HEVC, G.722, PCMU and PCMA are outside version 1.

The first prototype uses an exact source/manual selector. It does not run the WebRTC bandwidth estimator or automatically change quality. A later source switch requested through the authenticated event plane must use the existing provider-switch semantics and produce a discontinuity and new delivery generation.

## Version-1 binary envelope

Each WebSocket binary message contains exactly one protocol record. WebSocket fragmentation is transport-internal; the browser receives and validates the complete message. All multibyte integers use network byte order. The fixed header is 64 bytes:

| Offset | Size | Field | Version-1 value or meaning |
| ---: | ---: | --- | --- |
| 0 | 4 | magic | ASCII `NEKO` |
| 4 | 1 | version | `1` |
| 5 | 1 | record type | `1` FORMAT, `2` UNIT, `3` DISCONTINUITY, `4` END |
| 6 | 1 | kind | `0` none, `1` audio, `2` video |
| 7 | 1 | flags | bit 0 keyframe, bit 1 PTS valid, bit 2 DTS valid, bit 3 config present; bits 4-7 zero |
| 8 | 2 | header length | `64` |
| 10 | 2 | reserved | zero |
| 12 | 4 | metadata length | bytes after header and before payload |
| 16 | 4 | payload length | encoded/config payload bytes |
| 20 | 4 | track ID | `1` audio, `2` video in version 1 |
| 24 | 8 | delivery generation | per-track media-socket epoch |
| 32 | 8 | sequence | UNIT sequence within that generation |
| 40 | 8 | PTS | signed microseconds |
| 48 | 8 | DTS | signed microseconds |
| 56 | 8 | duration | unsigned microseconds |

Before allocating or slicing, a receiver must prove that the message length is exactly `header length + metadata length + payload length` without integer overflow. Version 1 rejects an unknown version, type, kind, track, flag bit, nonzero reserved field or header length other than 64.

The per-record field matrix is strict:

| Record | Kind/track | Flags | Generation | Sequence/timing |
| --- | --- | --- | --- | --- |
| FORMAT | audio/1 or video/2 | config-present only when payload is nonempty | new/current value, at least 1 | sequence, PTS, DTS and duration zero |
| UNIT | audio/1 or video/2 | PTS-valid/DTS-valid as sourced; keyframe allowed only for video | current value, at least 1 | sequence starts at 0; PTS nonnegative; invalid DTS is zero; duration is positive |
| DISCONTINUITY | audio/1 or video/2 | zero | new value, at least 1 | sequence, PTS, DTS and duration zero |
| END | none/0 | zero | zero | sequence, PTS, DTS and duration zero |

A common A/V discontinuity is represented by one DISCONTINUITY per requested track, followed by each track's FORMAT. Version-1 generation, sequence, PTS, DTS and duration values must also fit JavaScript's exact nonnegative integer range, at most `2^53 - 1`; valid duration is at most 10 seconds. A negative timestamp, zero/over-limit duration on UNIT or an invalid-DTS flag paired with nonzero DTS is a protocol error.

Timestamps and duration use microseconds because that is the WebCodecs chunk timebase. Conversion from Go `time.Duration` must be checked for range loss. PTS is always a backend-normalized, nonnegative value suitable for scheduling. The PTS-valid flag records whether the capture source supplied a valid PTS or the provider synthesized one. DTS is zero with its flag clear when unavailable. Supporting this distinction requires the later implementation to add `PTSValid` to `EncodedMediaUnit`; it is not a change made by this design block.

UNIT duration must be positive. If capture does not supply one, the adapter may derive video duration from the advertised frame-rate ratio and audio duration from the validated Opus packet; inability to derive a bounded duration is a format/backend error, not permission to send zero.

`delivery generation` starts at 1 separately for each track and socket. It increments on initial start, provider/source generation change, source switch, format change and any resynchronization. It is generated by the backend rather than copying `MediaSource.Generation`, because independent sources can use colliding provider-generation values.

UNIT sequence starts at zero and increases by one in sending order within a delivery generation. Per-track ordering is guaranteed; relative audio/video arrival order is not. A receiver rejects a record from an older generation, a sequence at or below the last accepted value, a generation jump without FORMAT, or video in a new generation before its first keyframe.

### FORMAT

FORMAT uses UTF-8 JSON metadata of at most 4 KiB. Its schema is `neko.media.format/1` and it contains only:

```json
{
  "schema": "neko.media.format/1",
  "source_id": "string",
  "source_generation": 1,
  "codec": "vp8",
  "mime_type": "video/VP8",
  "clock_rate": 90000,
  "channels": 0,
  "coded_width": 1920,
  "coded_height": 1080,
  "display_width": 1920,
  "display_height": 1080,
  "frame_rate_numerator": 25,
  "frame_rate_denominator": 1,
  "nominal_bitrate": 2000000
}
```

Fields that do not apply to the kind are zero, not omitted. The config-present flag requires a config payload; a clear flag requires an empty config payload. VP8 and raw Opus require the latter in version 1. A future codec cannot reuse version 1 until its config bytes are specified.

### UNIT

UNIT has no metadata. Its payload is one encoded VP8 frame or one raw Opus packet. Keyframe is valid only for video and must agree with the encoded VP8 frame. Audio UNIT records are treated as key chunks by WebCodecs but carry no video keyframe flag.

The maximum video payload is 8 MiB; the maximum audio payload is 64 KiB. Zero-length media payloads are invalid.

### DISCONTINUITY and END

DISCONTINUITY has no media payload and carries JSON metadata of at most 4 KiB:

```json
{"schema":"neko.media.discontinuity/1","reason":"source_restart"}
```

Allowed reasons are `source_switch`, `source_restart`, `format_change`, `timestamp_reset`, `server_overflow`, `browser_resync` and `resumed`. It starts a new generation and is followed by FORMAT; video then waits for a keyframe.

END has no media payload and uses schema `neko.media.end/1` with one of `normal`, `revoked`, `replaced`, `backend_error` or `shutdown`. FORMAT, DISCONTINUITY and END are lifecycle records. They are never silently dropped: the server clears queued media to make room or closes the local delivery.

## Browser-to-server control records

The media socket accepts only UTF-8 JSON text records of at most 4 KiB:

- `ready`: `{"type":"ready","version":1,"audio":{"generation":"1","codec":"opus","sample_rate":48000,"channels":2},"video":{"generation":"1","codec":"vp8","coded_width":1920,"coded_height":1080}}`; a deliberately disabled kind is `null`;
- `feedback`: `{"type":"feedback","audio":{"generation":"1","received":"10","decoded":"10","rendered":"9","compressed_queue":0,"decode_queue":0,"buffered_ms":80,"drops":"0"},"video":{"generation":"1","received":"10","decoded":"9","rendered":"8","compressed_queue":1,"decode_queue":1,"buffered_ms":0,"drops":"1"},"av_skew_ms":12}`; a disabled kind is `null`;
- `resync`: `{"type":"resync","kind":"all","generation":"1","reason":"decoder_error"}`; kind is `all`, `audio` or `video`, and reason is `queue_overflow`, `decoder_error`, `timestamp`, `audio_underflow` or `av_skew`;
- `stop`: exactly `{"type":"stop"}` for deliberate local shutdown.

Generation and sequence use canonical unsigned decimal strings because JSON numbers cannot represent every permitted header integer exactly. Queue depths must fit their declared caps; buffered time is 0-200 ms; reported absolute skew is at most 10,000 ms; cumulative drops are safe unsigned decimal strings. A requested kind may be null only when it was deliberately disabled. Unknown or missing fields and duplicate JSON object keys are rejected.

Feedback is sent at most once per second and is untrusted telemetry. The server applies syntax, enum, range and rate checks before use. The client must never send encoded media, input events or arbitrary event-plane messages on this socket. A client binary message or unknown text record is a protocol error.

## Bounded queues and slow-client isolation

All queues are per delivery. No operation in the capture callback waits for the WebSocket, decoder, renderer or audio device.

| Stage | Video bound | Audio bound | Additional bound |
| --- | ---: | ---: | --- |
| provider subscription | 4 units | 16 units | existing non-blocking `drop_newest` observer semantics |
| server egress media queue | shared | shared | 24 records and 16 MiB payload, whichever is reached first |
| server lifecycle queue | shared | shared | 4 records, lifecycle priority |
| browser compressed queue | 4 units | 16 units | stale generation rejected before enqueue |
| WebCodecs decode queue | 4 | 16 | checked before every `decode()` |
| decoded/render queue | 2 frames | 200 ms PCM | every discarded `VideoFrame` is closed |

One goroutine owns server socket writes and uses a 2-second deadline per record. Per-message compression is disabled. A missed write deadline or an egress queue that cannot return below its limit closes only that delivery with backpressure status; it never blocks capture or another viewer.

On provider or egress video overflow, the backend clears queued video, increments the video delivery generation, emits `server_overflow` plus FORMAT, and drops delta frames until a keyframe. The first implementation may obtain that keyframe with a bounded provider unsubscribe/resubscribe or pause/resume sequence; it must not add a blocking capture callback.

If audio overflow makes the common timeline unreliable, the backend performs a common audio/video resync: clear both media queues, increment both delivery generations, emit lifecycle records, reset audio and wait for a video keyframe. More than five resync requests in one minute closes the delivery rather than looping indefinitely.

The browser socket, parser and decoders live in a dedicated worker so rendering work does not deliberately stall socket consumption. Before calling `decode()`, it enforces the compressed and `decodeQueueSize` limits. Overflow, decoder failure or a hard timing error clears local queues, calls `reset()`, reconfigures after FORMAT and requests a common resync. It does not use `flush()` for loss recovery and never renders an obsolete generation.

The browser WebSocket API does not expose receive-side transport backpressure. Therefore the client sends progress feedback, and the server treats either of these as a hidden-backlog failure:

- rendered progress absent for 3 seconds while media is being sent; or
- reported rendered PTS more than 500 ms behind the active clock.

The server initiates one resync. Failure to resume progress within 2 seconds, or three resyncs within 30 seconds, closes that delivery.

## A/V clock and resynchronization

Both tracks carry provider-normalized microsecond PTS. When audio is enabled and its `AudioContext` is running, the audio clock is master:

1. after FORMAT/resync, the first decoded audio timestamp is anchored to `AudioContext.currentTime + 80 ms`;
2. subsequent audio is scheduled relative to that PTS and transferred to an `AudioWorklet` with bounded buffers;
3. video presentation time is derived from the same anchor;
4. the renderer chooses the closest frame, drops video over 80 ms late and holds an early frame for at most 100 ms.

The design does not require `SharedArrayBuffer` or cross-origin isolation. Muting uses the audio graph so decoding and the master clock continue. When audio is deliberately disabled, a monotonic `performance.now()` clock with the same 80 ms initial lead becomes master.

The soft target is absolute A/V skew at or below 80 ms. Any of these triggers a common resync:

- absolute skew over 200 ms for at least 1 second;
- timestamp regression or duration overflow;
- audio underrun/overflow;
- decoder error;
- new generation or DISCONTINUITY that invalidates the common anchor.

WebCodecs `reset()` removes decoder configuration and makes video require a new key chunk. The recovery sequence is therefore reset, discard, request resync, receive FORMAT, reconfigure, then accept a keyframe. The client must not draw or play residual output from the old generation.

The existing central Play fallback remains the user gesture that resumes a suspended audio context and starts the selected media surface. This contract does not change Safari/WebRTC autoplay handling. A canvas-based first prototype may not provide video-element Picture-in-Picture; that limitation must be visible in the opt-in UI and must not alter the default surface, touch/trackpad coordinates or fullscreen behavior.

## Reconnect and status handling

A selected client may reconnect only the media delivery while its authenticated event WebSocket and live session remain valid. It performs at most four serialized attempts after 1, 2, 5 and 10 seconds. Every attempt obtains a new one-time ticket; attempts never overlap and no old generation is retained.

There is no automatic retry after logout, explicit stop, authorization failure, unsupported codec/protocol or server policy rejection. After exhaustion the UI offers the two explicit actions described above. Existing event-WebSocket reconnect ownership and timing remain unchanged.

Pre-upgrade HTTP results are fixed as follows:

| Status | Meaning |
| ---: | --- |
| 400 | malformed path, query, header or ticket syntax |
| 401 | ticket unknown or binding invalid |
| 403 | origin rejected or current session lacks `CanWatch` |
| 409 | conflicting attach state |
| 410 | ticket expired or already redeemed |
| 426 | required version/subprotocol unavailable |
| 429 | creation or attachment rate exceeded |

After upgrade, private application close codes are:

| Code | Meaning | Automatic same-backend retry |
| ---: | --- | --- |
| 4400 | malformed or inconsistent protocol record | no |
| 4401 | attachment/session no longer valid | no |
| 4403 | authorization revoked | no |
| 4406 | codec or format unsupported | no |
| 4408 | primary delivery replaced | no |
| 4413 | bounded backpressure/overload | yes |
| 4429 | control or reconnect rate exceeded | no |
| 4500 | local backend failure | yes |

Retryable codes still retry only `webcodecs-ws`; they never select WebRTC.

## Security and resource limits

The first implementation must enforce these defaults before target-server testing:

- `wss:` in production; cleartext `ws:` only for an explicit loopback development environment;
- exact configured Origin comparison before upgrade;
- no query parameters and no ambient login/share credential at the media route;
- one pending ticket and one active media delivery per session;
- ticket creation: six per minute per session with burst two;
- invalid attachment: twenty per minute per resolved remote address plus a 64-slot global verifier semaphore; forwarded addresses count only from explicitly trusted proxies;
- maximum 128 concurrent prototype sockets unless the operator explicitly lowers it;
- client control records: 4 KiB, ten per second with burst twenty;
- metadata: 4 KiB; codec config: 64 KiB; audio UNIT: 64 KiB; video UNIT: 8 MiB;
- server egress: 24 media records and 16 MiB; lifecycle queue: four records;
- READY timeout 5 seconds; feedback/progress timeout 5 seconds;
- ping every 10 seconds, pong timeout 20 seconds, write deadline 2 seconds, close deadline 1 second;
- maximum sustained outbound application payload 16 MiB/s per socket over a 5-second window;
- maximum five client resync requests per minute;
- no WebSocket per-message compression.

Length arithmetic is checked before allocation. JSON has fixed schemas and bounds. All protocol errors close only the offending delivery. The handler must release subscriptions, decoder-facing buffers, ticket state and manager state on every exit path.

Logs must never contain the ticket-bearing `Sec-WebSocket-Protocol` header, full request URL, cookie, authorization header, login/share token or new event payload. Normal operational logs may use the existing session ID and backend name. Errors use enum reason codes rather than reflecting hostile input.

Metrics have low-cardinality labels only and do not label tickets, sessions, IPs, origins, user agents, source IDs or URLs. The implementation should add:

- `neko_media_websocket_connections{state}`;
- `neko_media_websocket_handshakes_total{result}`;
- `neko_media_websocket_records_total{kind,type}`;
- `neko_media_websocket_bytes_total{kind}`;
- `neko_media_websocket_drops_total{kind,stage,reason}`;
- `neko_media_websocket_resyncs_total{reason}`;
- bounded histograms for write time, queue depth, queue bytes and client-reported lag.

The generic `neko_media_*` backend and delivery metrics remain authoritative for manager/provider lifecycle.

## Bounded implementation plan

Implementation is a later block on `testing`, reviewed and committed separately.

### Phase 1: protocol and ticket boundary

1. Add `PTSValid` to `EncodedMediaUnit` and preserve it through the capture provider; keep the Pion adapter behavior unchanged.
2. Add an isolated `server/internal/mediaws` package for the envelope encoder, strict parser, ticket store and limit constants.
3. Add byte-exact Go golden fixtures for every record type, cross-language parser fixtures, truncation/overflow cases and fuzz targets.
4. Add the four event-plane negotiation messages to current and legacy adapters and mark their payloads non-loggable.
5. Add default-off configuration without renaming or changing existing keys.

### Phase 2: server delivery adapter

1. Implement one `MediaDeliveryBackend` using the generic provider subscriptions and credential-free lease.
2. Register the backend and `/api/media/ws` route only while the feature is enabled.
3. Enforce pre-upgrade authorization, ticket, origin and rate checks before allocating delivery resources.
4. Implement the single writer, lifecycle priority, queue caps, drop-to-keyframe state machine, progress timeout and complete cleanup.
5. Preserve manager replacement, `CanWatch` revocation, private-mode pause and shutdown behavior with focused tests.

### Phase 3: isolated client path

1. Add a strict TypeScript envelope parser using the shared golden fixtures and checked 64-bit handling.
2. Add a dedicated worker that owns the socket, support probes, decoder queues and stale-generation rejection.
3. Add an `AudioWorklet` and bounded video renderer implementing the fixed common clock and reset sequence.
4. Put the new surface behind both opt-ins. Do not modify the default WebRTC service or data-channel opcodes.
5. Integrate central Play, mute, resize, fullscreen and touch-coordinate geometry; expose prototype limitations such as Picture-in-Picture.
6. Implement only the four bounded same-backend reconnect attempts and the two explicit operator/user actions.

### Phase 4: deployment and validation assets

1. Add a separate sanitized opt-in Compose overlay; leave stable base Compose unchanged.
2. Add credential-safe dashboards/queries for the fixed queue, drop, resync, connection and resource metrics.
3. Complete static security review and the target-server matrix below before calling the prototype validated.

No phase includes HLS, LL-HLS, WebTransport, automatic backend choice or a new control transport.

## Target-server acceptance matrix

These checks are for the real target server only. They are **NOT EXECUTED IN CODEX** as part of this design.

Record the exact commit, image digests, Compose overlays, browser/OS/device versions, capture codec/profile, source dimensions/frame rate, network shaping, participant roles and measurement method.

### Default invariance

- With the feature absent or disabled, `/api/media/ws` is unregistered, capability responses do not advertise the backend, and an ordinary client sends no prototype events.
- The existing WebRTC signaling, A/V, data-channel control, reconnect, view-only and iOS behavior passes its applicable established matrix.
- Stable Compose files and existing configuration values produce no new listener, pipeline or persistent state.

### Functional and authorization matrix

- VP8 plus 48 kHz stereo Opus succeeds for ordinary, admin and `CanWatch` view-only sessions.
- A session without `CanWatch` is denied before upgrade; revocation, logout and replacement stop output within 2 seconds.
- View-only input remains denied; private mode pauses and resumes with a clean generation.
- Unsupported audio or video rejects the requested delivery rather than silently changing codecs or omitting a requested kind.
- Explicit **Use WebRTC** restores the unchanged WebRTC A/V and control path.

### Startup, latency and sync

Across ten clean opt-in joins on the supported desktop Chromium/Brave target:

- p95 first rendered video is at most 2 seconds after media-socket open;
- p95 audible synchronized A/V is at most 3 seconds after media-socket open;
- p95 glass-to-glass video latency is at most 500 ms and no more than 250 ms worse than the same-device WebRTC baseline under the same source and network;
- p95 absolute A/V skew is at most 80 ms, with no excursion over 200 ms lasting 1 second.

Browser support results are evidence, not assumption. An unsupported browser must fail the explicit support probe without disturbing its session and must move to WebRTC only after the explicit action.

### Slow-client isolation and reconnect

Run two healthy WebRTC viewers and one `webcodecs-ws` viewer on the same source. Stall or constrain only the WebSocket viewer:

- provider and egress queues never exceed their fixed record/byte caps;
- capture callbacks and healthy viewers do not stall or switch quality because of that viewer;
- only the affected delivery resynchronizes or closes;
- after the constraint is removed, current-generation video resumes from a keyframe within 2 seconds of a successful reconnect;
- no obsolete frame is rendered and no parallel reconnect attempts occur;
- a media-only disconnect uses new tickets and at most the 1/2/5/10-second attempts;
- logout, kick or `CanWatch` revocation suppresses retries.

Repeat one induced down/up cycle with the accepted adaptive three-viewer profile during final grouped validation; it remains deferred from the earlier compatibility checkpoint for exactly that grouped run.

### Resource cost

Compare stable WebRTC with the opt-in viewer under the same source and five-minute workload:

- adding the viewer creates no extra capture encoder or GStreamer pipeline;
- average server CPU increase is at most 10 percentage points of one CPU core per opt-in viewer;
- server RSS increase is at most 128 MiB per opt-in viewer and returns to within 32 MiB of baseline 60 seconds after disconnect;
- application payload bytes stay within 10% of the selected encoded-source bytes plus the specified envelope/JSON overhead;
- no queue, goroutine, ticket or manager-delivery count grows after repeated connect/disconnect cycles.

### Malformed and hostile input

Exercise wrong/missing Origin, malformed/replayed/expired tickets, wrong subprotocol/version, every unknown flag/type, nonzero reserved fields, truncated/overflowing lengths, over-limit metadata/config/audio/video, invalid JSON, binary client messages and control-rate abuse.

Each case must produce the specified HTTP status or private close code, affect only the offending client, disclose no credential or hostile payload in logs, cause no server panic/restart, and return memory and delivery counts to the bounds above.

## Rollback

Rollback is intentionally state-free:

1. remove `?media=webcodecs-ws` from clients;
2. disable or remove the opt-in deployment overlay;
3. recreate only the affected service if configuration changed;
4. verify normal WebRTC A/V and data-channel control.

There is no database migration, volume change, capture-format mutation or automatic preference to undo. A failed prototype therefore rolls back to the existing WebRTC path without changing the stable deployment contract.

## Standards and project evidence

- [WebCodecs](https://www.w3.org/TR/webcodecs/) defines decoder support probes, timestamp units, queue visibility and reset/key-chunk behavior.
- [WebCodecs Codec Registry](https://www.w3.org/TR/webcodecs-codec-registry/), [VP8 registration](https://www.w3.org/TR/webcodecs-vp8-codec-registration/) and [Opus registration](https://www.w3.org/TR/webcodecs-opus-codec-registration/) define the codec strings and chunk framing used here.
- [WebSockets](https://websockets.spec.whatwg.org/) and [RFC 6455](https://www.rfc-editor.org/rfc/rfc6455) motivate explicit origin, authentication, message and buffering limits; the browser API's `bufferedAmount` is send-side only, so receive progress is an application concern.
- Upstream [alternative-media issue #690](https://github.com/m1k1o/neko/issues/690) is directional evidence for subscription-first evolution and explicitly notes the lack of free WebSocket backpressure and the separate control-path problem.
