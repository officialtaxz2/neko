# HLS / Low-Latency HLS passive delivery contract

Status: **version-1 design and Phase 1 foundations complete in the repository on `testing`; no HLS HTTP endpoint, packager, encoder, player, deployment overlay or automatic selection is implemented, and target-server tests/build remain pending**.

This document is the normative contract for the first default-off passive/view-only HTTP-streaming prototype. It specializes the encoded-source/subscription and participant-delivery boundary in [`MEDIA_SUBSCRIPTION_BOUNDARY.md`](MEDIA_SUBSCRIPTION_BOUNDARY.md). It does not authorize a second desktop capture, a stable-deployment change or a claim that any untested device supports the proposed path.

The key words **MUST**, **MUST NOT**, **SHOULD** and **MAY** describe the intended implementation contract.

## Scope and non-goals

Version 1 provides receive-only live audio/video for an already authenticated participant in the existing logical Neko room. It has two explicit modes:

- `hls`: conventional live HLS for the broadest passive-device compatibility;
- `ll-hls`: the same media objects plus Low-Latency HLS playlists, partial segments and blocking reloads.

Both modes are absent unless a future dedicated server flag and deployment overlay enable them. WebRTC remains the default interactive backend. `webcodecs-ws` remains an explicit independent receive prototype. There is no automatic fallback between any backend.

This contract does not add:

- a control, keyboard, pointer, touch, clipboard, file, microphone or camera transport;
- a new room, identity, member profile or authorization system;
- DASH, WebTransport, DRM, recording, CDN distribution or public anonymous playlists;
- a second X11/PulseAudio desktop capture;
- a client dependency, endpoint or packager in this design-only block.

The HLS backend itself is always receive-only. A normal/admin profile retains only the rights it already has on the existing event plane; selecting HLS grants none. The first product UI SHOULD expose HLS to server-enforced view-only sessions and an explicit administrator diagnostic path. Only `MemberProfile.IsViewOnly` provides the hard no-input guarantee.

## Evidence baseline and compatibility choice

The current adaptive deployment publishes VP8 video and raw Opus through the encoded-media provider. Those codecs MUST NOT be advertised as native HLS compatibility. Version 1 instead produces:

- fragmented MP4 (`fMP4`) video renditions using H.264/AVC High Profile, Level 3.1, 8-bit 4:2:0, `avc1` sample entries and no B-frames;
- one separate fMP4 stereo AAC-LC audio rendition at 48 kHz and 128 kbit/s, advertised as `mp4a.40.2`;
- one multivariant playlist with an AAC audio group and separate video renditions.

The exact H.264 `CODECS` value MUST be derived from the emitted AVC configuration record rather than copied from configuration. The intended first value is `avc1.64001f`; a mismatch is a packaging error. HEVC, AV1, Opus-in-fMP4, MPEG-TS and alternate audio codecs are outside version 1.

This choice follows the current Apple authoring requirements: H.264 in fMP4 is supported, High Profile is preferred, IDR frames are recommended every two seconds, and stereo AAC must be available. It also gives an ISO BMFF input to MSE-based players. It is still a hypothesis until the device matrix below passes.

### First target-device matrix

| Client class | Version-1 playback path | Required first result |
| --- | --- | --- |
| Desktop Brave/Chrome/Edge | pinned, locally bundled `hls.js` over MSE/fMP4 | `hls` and `ll-hls` |
| Desktop Firefox | same pinned `hls.js` path | `hls`; `ll-hls` measured separately |
| macOS Safari | native `<video>` HLS first | `hls` and `ll-hls` |
| iPhone Safari | native `<video playsinline>` HLS | `hls` and `ll-hls`, audio, orientation and native fullscreen |
| iPad Safari | native `<video playsinline>` HLS | `hls` and `ll-hls`, audio and fullscreen/PiP |
| Actual Smart-TV browser | native HLS when its documented/runtime probe supports it | conventional `hls` is mandatory; LL-HLS is optional evidence |
| Constrained/older browser | native HLS, otherwise the pinned MSE path only when supported | conventional `hls`; otherwise a diagnosable unsupported state |

Apple clients use native HLS deliberately so their native media/fullscreen behavior is preserved. Other clients use feature detection and a repository-pinned `hls.js` version; they MUST NOT load a floating CDN script. A non-empty `canPlayType()` result alone is not sufficient evidence for non-Apple browsers. If neither the exact native codec/container path nor MSE playback is supported, the client reports an unsupported terminal state and does not switch backend automatically.

Every real Smart-TV result MUST record manufacturer, model, OS/firmware, browser/runtime version, native-versus-MSE path and exact mode. “Smart-TV supported” is not an acceptable aggregate claim.

## Topology and ownership

```text
existing shared desktop capture / encoded provider
        |
        | one bounded source subscription per active video variant
        | one bounded source subscription for shared audio
        v
default-off HLS packager set
        +-- AAC audio rendition --------------------+
        +-- low H.264 rendition --------------------+-- immutable fMP4 objects
        +-- medium H.264 rendition -----------------+   and playlist snapshots
        `-- high H.264 rendition -------------------+
                                                         |
                    per-viewer MediaDelivery + HTTP lease|
        session A ---------------------------------------+
        session B ---------------------------------------+
```

There is one process-wide packager set for the single active Neko room, not one encoder or muxer per viewer. One audio worker is shared by all variants. Each active video variant has exactly one decode/encode/mux worker and one provider subscription, shared across both `hls` and `ll-hls` viewers. Regular and low-latency modes render different playlist views over the same retained objects.

The first authorized lease starts all three configured variants so the multivariant playlist is internally consistent and can adapt immediately. The hard version-1 maximum is three video variants plus one audio rendition. When the last unpaused lease closes, the set enters a 15-second idle grace and then closes subscriptions, workers and retained objects. A new viewer after teardown starts a new packager generation.

The packager consumes the existing encoded provider. On the current VP8/Opus deployment it decodes and re-encodes once per active rendition. This is intentionally less efficient than a future shared raw-frame tee, but it does not create another desktop capture and keeps the first prototype behind the existing source contract. A later optimization may add HLS-compatible provider sources only if it preserves the same capture, authorization and isolation invariants.

Per-viewer ownership remains in `MediaDeliveryManager`:

1. the authenticated current session is resolved;
2. current `CanWatch` is required;
3. one credential-free `MediaLease` and HLS delivery are prepared, and backend open waits for the mode-specific ready generation within the fixed startup deadline;
4. that delivery replaces any prior primary delivery for the session;
5. the delivery holds a reference to the shared packager set but never owns its capture;
6. logout, deletion, disconnect after the existing grace, `CanWatch == false`, replacement or shutdown revokes it centrally.

`IsConnected` continues to describe the event connection. `IsWatching` becomes true only after an authenticated media playlist has been served from a ready generation, and false when that delivery is paused, failed or closed. Individual segment GETs are not participants.

## Renditions, timestamps and keyframes

The initial target-server profile is:

| ID | Source geometry/rate on the current profile | H.264 video target | Initial `AVERAGE-BANDWIDTH` including audio | Initial peak `BANDWIDTH` |
| --- | --- | ---: | ---: | ---: |
| `high` | 1280x720 at 25 fps | 3,000 kbit/s | 3,128,000 | 4,000,000 |
| `medium` | approximately 854x480 at 20 fps | 1,100 kbit/s | 1,228,000 | 1,500,000 |
| `low` | 640x360 at 15 fps | 365 kbit/s | 493,000 | 650,000 |

The packager MUST take actual width, height and frame rate from the first complete provider `FORMAT`; it MUST NOT publish zero or guessed dimensions. The table is the first deployment target, not permission to mislabel another configuration. `AVERAGE-BANDWIDTH` and `BANDWIDTH` MUST be replaced with measured values before acceptance and must satisfy the Apple live-stream guidance recorded in the sources below.

All renditions use a closed two-second GOP. With the current rates this means 50, 40 and 30 frames respectively. IDRs align to the same two-second presentation boundaries across variants. A six-second parent segment begins on an aligned IDR; the parts at 0, 2 and 4 seconds are marked `INDEPENDENT=YES`. Other parts are not advertised as independent.

Provider `PTS`, `DTS`, validity, duration, generation and discontinuity are authoritative inputs. The packager maps them onto one zero-based presentation timeline per packager generation, uses a 90 kHz video timescale and 48 kHz audio timescale, and never substitutes wall-clock arrival as media time. One wall-clock/monotonic anchor per generation produces `EXT-X-PROGRAM-DATE-TIME`; it does not alter media PTS.

A provider generation change, format/config change, source switch, timestamp regression, input overflow, encoder restart or mux failure ends the current packaging generation. The next published media playlist MUST:

- discard any incomplete parent segment and parts;
- wait for fresh complete formats and an aligned video IDR;
- publish a new init segment through `EXT-X-MAP`;
- increment `EXT-X-DISCONTINUITY-SEQUENCE` as needed and place `EXT-X-DISCONTINUITY` before the first new-generation media;
- keep media sequence numbers strictly increasing and never reuse an object URI;
- update every rendition within one part target for LL-HLS or fail the lagging rendition out of the master playlist.

No object crosses a generation boundary. The master playlist advertises only renditions with a complete init section and a playable aligned boundary.

## Conventional HLS and LL-HLS parameters

Both modes use individual fMP4 resources, relative URIs and live playlists without `EXT-X-ENDLIST`.

| Parameter | `hls` | `ll-hls` |
| --- | --- | --- |
| Playlist protocol version | 7 | 9 |
| Parent target duration | 6 seconds | 6 seconds |
| Part target | none | 1 second |
| GOP / independent boundary | 2 seconds | 2 seconds |
| `HOLD-BACK` | 18 seconds | 18 seconds |
| `PART-HOLD-BACK` | none | 3 seconds |
| Playlist-visible completed parents | 3 | 3 plus the current partial parent |
| Internally retained completed parents | 6 | 6 plus current parts |
| Blocking reload | no | yes, `CAN-BLOCK-RELOAD=YES` |
| Preload hint | no | required for next part |
| Rendition reports | no | required for every other rendition |
| Delta updates / `_HLS_skip` | no | no in version 1 |

One-second parts are deliberate: Apple recommends that value and requires the target to cover expected RTT; the prototype does not assume a 200 ms part is reliable. `PART-HOLD-BACK=3` is three part targets. The conventional 18-second hold-back is three target durations. The short three-segment visible window is still spec-valid while bounding retained/replayable live history.

An LL-HLS playlist accepts only decimal `_HLS_msn` and `_HLS_part` directives; `_HLS_part` without `_HLS_msn` is `400`. The server blocks for at most seven seconds for a valid near-future request, wakes on publication/revocation/pause/shutdown, and returns `503` with `Retry-After: 1` when the requested update is still unavailable. Requests more than two parent sequences or three parts ahead are `400`. `_HLS_skip` and all unknown query fields are rejected in version 1.

### Required playlist and object shape

The multivariant playlist contains `EXTM3U`, `EXT-X-VERSION:7`, `EXT-X-INDEPENDENT-SEGMENTS`, one `EXT-X-MEDIA:TYPE=AUDIO` entry, and one `EXT-X-STREAM-INF` per ready video rendition. Every stream entry includes measured `BANDWIDTH`, `AVERAGE-BANDWIDTH`, exact `RESOLUTION`, `FRAME-RATE`, the emitted H.264 plus AAC `CODECS` values, and the shared audio group. It contains no absolute URI, credential, session ID or unavailable rendition.

A conventional media playlist contains at least `EXTM3U`, `EXT-X-VERSION:7`, `EXT-X-TARGETDURATION:6`, monotonically increasing `EXT-X-MEDIA-SEQUENCE`, the current `EXT-X-DISCONTINUITY-SEQUENCE`, `EXT-X-SERVER-CONTROL:HOLD-BACK=18`, `EXT-X-MAP`, `EXT-X-PROGRAM-DATE-TIME` and exactly three completed `EXTINF` entries once warm. Conventional backend open becomes ready only after the audio and every advertised video rendition have an init section plus three complete aligned parents. It waits for at most 24 seconds; expiry fails bootstrap with `503` and leaves the prior primary delivery unchanged.

An LL-HLS media playlist uses version 9 and adds `EXT-X-PART-INF:PART-TARGET=1`, `EXT-X-SERVER-CONTROL:CAN-BLOCK-RELOAD=YES,HOLD-BACK=18,PART-HOLD-BACK=3`, current `EXT-X-PART` entries, one next-part `EXT-X-PRELOAD-HINT`, and `EXT-X-RENDITION-REPORT` for every other ready rendition. Audio parts are independently decodable. Video parts beginning at the aligned 0/2/4-second IDRs carry `INDEPENDENT=YES`; other video parts do not. Normal parts are one second and satisfy the specification's 85%-of-target rule; only the permitted final/independent/gap exceptions may be shorter. LL-HLS backend open becomes ready after the audio and every advertised video rendition have an init section plus three complete aligned parts, including an independent video start. It waits for at most six seconds; expiry has the same fail-without-replacement behavior. Completed parents accumulate into the fixed three-parent window after startup.

Every init object is one valid ISO BMFF initialization section containing `ftyp` and `moov`. Every part/parent object is a self-contained `moof` plus `mdat` fragment whose decode times and sample durations are monotonic and reference the current init section. Parent segments are addressable complete objects, not concatenation instructions over credential-bearing byteranges. Version 1 claims HLS-compatible fragmented MP4, not broader CMAF conformance, until a CMAF validator is part of acceptance.

The acceptance targets, measured from a cold explicit selection on the actual remote path, are:

- conventional HLS: p95 first moving video with audio at or below 24 seconds and p95 glass-to-glass latency at or below 24 seconds;
- LL-HLS: p95 first moving video with audio at or below 6 seconds and p95 glass-to-glass latency at or below 6 seconds;
- both: no unrequested backend change, no A/V drift over 100 ms for more than two seconds, and no playback stall longer than one second during a ten-minute unshaped run after startup.

These are prototype gates, not claims. LL-HLS with a one-second part target additionally requires measured path P95 RTT at or below 333 ms so the part target is at least three times that RTT; otherwise LL-HLS remains unavailable on that path while conventional HLS can still be tested. A synchronized visible clock/QR frame in the shared desktop and synchronized client wall clock MUST be used for at least 30 samples per mode/device; playlist timestamps or request duration alone are not end-to-end latency evidence.

## Authenticated creation and credential transport

### Event-plane negotiation

The future client negotiates through its already authenticated event WebSocket:

1. `media/hls/capabilities/request` asks for `hls` or `ll-hls`.
2. `media/hls/capabilities` reports enabled modes, source IDs, codec/container contract and limits.
3. `media/hls/create` requests one exact mode and the server-owned variant set.
4. The controller re-resolves the exact live session and `CanWatch`, then creates one pending bootstrap ticket.
5. `media/hls/offer` returns only the HTTPS relative bootstrap path, one-time ticket and expiry.

All four payloads reject unknown fields and are redacted from event payload logs. The bootstrap ticket reuses the established media-ticket shape: 24 random bytes, 32 unpadded Base64URL characters, ten-second lifetime, SHA-256 digest at rest, atomic single use, one pending ticket per session, and binding to session/backend/mode/request. It is sent only in the JSON body of the bootstrap POST, never in a URL, cookie, log or metric.

### Bootstrap and playback lease

The exact bootstrap is:

```text
POST /api/media/hls/session
Content-Type: application/json
Origin: https://the-exact-allowed-origin

{"ticket":"<32-character-one-time-ticket>"}
```

After all pre-body and ticket checks pass, the server prepares the central delivery and waits for the mode-specific ready condition above. Only a successful backend open atomically replaces the prior primary delivery and returns `201` with:

```json
{
  "mode": "ll-hls",
  "master": "/api/media/hls/<22-character-public-id>/master.m3u8",
  "idle_expires_in_ms": 30000
}
```

It also sets:

```text
Set-Cookie: __Secure-neko-hls=<32-character-secret>; Path=/api/media/hls/<public-id>/; Max-Age=30; Secure; HttpOnly; SameSite=Strict
```

The public ID is 16 random bytes encoded as 22 unpadded Base64URL characters. It is an opaque routing/correlation handle, not sufficient authorization. The cookie secret is an independent 24 random bytes encoded as 32 Base64URL characters and is stored only as a SHA-256 digest. A unique cookie path permits independent HLS tabs without putting a bearer in their playlist URLs.

The lease has a 30-second sliding idle deadline. A valid master/media-playlist response or same-origin `POST .../<public-id>/keepalive` extends the server deadline and repeats the same cookie with `Max-Age=30`; the intended client keepalive interval is 15 seconds. Segment/init/part requests do not extend it. The event session and central `MediaLease` must still be current. There is no IP binding because mobile network changes are legitimate and no durable or cross-service credential is created.

Opening another primary delivery for that session invalidates the public ID and cookie immediately. The secret is never accepted by room, login, event, control, plugin, upload or publish APIs. It is erased on close and never persisted across a process restart.

## HTTP resource model

All playlist URIs are relative and remain under the lease path:

```text
/api/media/hls/<public-id>/master.m3u8
/api/media/hls/<public-id>/audio/index.m3u8
/api/media/hls/<public-id>/audio/init-<generation>.mp4
/api/media/hls/<public-id>/audio/seg-<sequence>.m4s
/api/media/hls/<public-id>/audio/part-<sequence>-<part>.m4s
/api/media/hls/<public-id>/<variant>/index.m3u8
/api/media/hls/<public-id>/<variant>/init-<generation>.mp4
/api/media/hls/<public-id>/<variant>/seg-<sequence>.m4s
/api/media/hls/<public-id>/<variant>/part-<sequence>-<part>.m4s
```

The lease handler authenticates every `GET` and `HEAD` with both exact public ID and cookie digest before resolving an object. Unknown, expired, revoked, mismatched and syntactically valid nonexistent leases all return the same `404`; they must not be distinguishable. Aged-out or unknown objects also return `404`. The one exception is the exact next part named by the current `EXT-X-PRELOAD-HINT`: its request is admitted as a bounded blocking resource request, sends no partial bytes, and wakes only when the complete immutable part can be written at connection speed or the seven-second/revocation deadline ends. `HEAD` returns the same authorization and headers without a body. One syntactically valid byte range is supported for completed MP4 objects; multiple/invalid ranges return `416`.

The implementation uses immutable byte slices and releases the store lock before writing to a client. A slow or disconnected HTTP response can retain only its selected immutable object until request cancellation; it cannot hold a packager lock, provider event or mutable playlist builder.

### Bootstrap status contract

| Status | Meaning |
| ---: | --- |
| 400 | malformed path/query/JSON/ticket syntax or unsupported mode |
| 401 | syntactically valid but unknown ticket |
| 403 | insecure/untrusted transport, missing/wrong Origin or current authorization denied |
| 410 | expired, invalidated or already redeemed ticket |
| 413 | body over 2 KiB |
| 429 | ticket, lease or request limit exceeded |
| 503 | packager capacity/startup unavailable; `Retry-After: 1` |

The route is `404` when the backend is disabled. Failed bootstrap never replaces the current primary delivery.

## Security, origin, proxy and cache policy

Version 1 is same-origin and HTTPS-only. There is no cleartext-loopback exception because the playback credential requires a `Secure` cookie. The server captures the socket peer before generic forwarded-header rewriting. `Forwarded` and `X-Forwarded-*` transport/host/address information is honored only when that peer is inside an explicitly configured trusted-proxy CIDR. The bootstrap/keepalive Origin must exactly match a configured HTTPS origin; wildcard, suffix and reflected origins are forbidden.

Media `GET`/`HEAD` requests may omit `Origin` because native media stacks do so. If `Origin` is present it must be exact; an explicit cross-site `Sec-Fetch-Site` is rejected. No response emits `Access-Control-Allow-Origin`; `OPTIONS` is not a credential workaround. Relative playlist URIs prevent forwarded-host injection.

Every response carries `Referrer-Policy: no-referrer`, `X-Content-Type-Options: nosniff` and the existing restrictive security headers. Playlist content type is `application/vnd.apple.mpegurl`; MP4 objects use `video/mp4` or `audio/mp4`. Errors have fixed small bodies and never echo IDs, cookies, tickets or paths.

All HLS responses use:

```text
Cache-Control: private, no-store, max-age=0
Pragma: no-cache
Vary: Cookie, Accept-Encoding
```

Version 1 therefore forbids CDN/shared-cache delivery. The reverse proxy must disable response buffering and compression for MP4 objects. Playlist responses implement negotiated gzip with `Vary: Accept-Encoding`; LL-HLS acceptance requires gzip when the client advertises it. Public LL-HLS acceptance additionally requires client-facing HTTP/2 or HTTP/3. Playlist responses carry RFC 9218 urgency 1 and media objects urgency 2 through 6, never giving a higher-bitrate rendition higher priority than a lower one. Caddy may proxy to the local service over HTTP, but only its configured HTTPS listener is a supported client entry point.

Access logs MUST normalize every lease resource to a template such as `/api/media/hls/:lease/:variant/:object`; neither public IDs nor raw query strings are logged. Application logs and metric labels contain no ticket, cookie, public ID, playlist URI, view-only token or login credential. The existing compact view-only fragment remains a login bootstrap only.

Version 1 does not encrypt media objects separately and is not DRM. TLS plus per-request authorization protects delivery in transit; any bytes already received by a client cannot be revoked retroactively.

### Fixed limits

- 128 active playback leases process-wide;
- 512 simultaneous HLS HTTP requests process-wide;
- four simultaneous requests and two simultaneous blocking requests (playlist reload or hinted part) per lease;
- 64 simultaneous blocking requests process-wide;
- bootstrap: two requests per minute per session, burst two;
- keepalive: six per minute per lease, burst two;
- playlists: 12 requests per second per lease, burst 24;
- init/segment/part: 32 requests per second per lease, burst 64;
- bootstrap body: 2 KiB; generated playlist: 64 KiB;
- init object: 2 MiB; part: 1 MiB; parent segment: 8 MiB;
- blocking reload wait: seven seconds;
- header/read timeout: five seconds; ordinary response write timeout: 30 seconds.

All counters use monotonic bounded token buckets. Limit failure is local to the lease/request and never blocks capture or another delivery.

## Private mode, revocation and shutdown

`MediaDelivery.SetPaused(true)` is the authoritative private-mode transition for this backend. A paused delivery:

- stops contributing to packager demand;
- wakes its blocking playlist requests;
- returns `503` plus `Retry-After: 1` for playlists and media objects;
- causes the cooperative client to pause, remove `src`, call `load()` and clear decoded/native buffers;
- remains attached so `SetPaused(false)` can resume through a fresh bootstrap or fresh-generation playlist without changing profile authority.

If all leases are paused, packager idle teardown follows the same 15-second grace. Resume after teardown starts a new generation.

Logout, kick, session deletion, `CanWatch` loss, token-rotation service recreation, primary-delivery replacement and shutdown invalidate the HTTP lease immediately. New requests then return the uniform `404`; blocked requests wake before close. The client clears media on the matching event-plane transition and displays a terminal/retry state. Authorization/policy failures never trigger automatic fallback.

A cooperative foreground client must stop rendering within two seconds of revocation. The server must stop new media bytes within one second. A hostile client may retain bytes it fetched while authorized: the visible playlist contains at most three completed six-second parents plus the current partial parent. This bounded post-revocation possession is an inherent HTTP-download limitation and must be part of the security acceptance record, not hidden as instantaneous revocation.

During process shutdown, viewer leases close first, blocked requests wake, packager workers end, retained objects are released, provider subscriptions close, and only then may capture stop. The total backend shutdown deadline remains within the central five-second delivery timeout.

## Bounded queues, storage and isolation

Each video rendition and the shared audio worker requests a provider queue capacity of 64 with `drop_newest`. Internal worker hand-off is at most eight media events per rendition; published-object notification is at most two pending notifications because playlist snapshots can be regenerated from the store. No queue waits in provider fan-out.

Any provider or internal overflow invalidates the affected packaging generation. The worker drops until the next aligned video IDR/common audio boundary, publishes a discontinuity and resumes; it never grows latency by replaying a backlog. If one rendition cannot recover within 12 seconds, it is removed from new master playlists while the other renditions continue. If audio fails, the whole set enters failed/warming state because version 1 does not silently change requested media composition.

Version 1 stores objects only in memory:

- six completed parents per rendition plus the current parent;
- at most 42 parts per rendition (six parents plus current at six one-second parts each);
- one current and one immediately previous init section per rendition;
- 64 MiB aggregate retained-object hard limit;
- no segment, playlist, ticket, cookie or lease file on disk.

Publication that would exceed an object-size or aggregate limit fails the affected generation and evicts only already-unadvertised oldest objects. It never evicts an advertised object merely to make room. If the bound cannot be restored, the packager fails closed and existing WebRTC/WebCodecs deliveries continue.

A response writes from an immutable object outside all store/packager locks. Kernel/proxy backpressure, a client that reads one byte at a time, aborted range requests and repeated playlist polling can consume only the fixed request slots and rate budget. They cannot slow provider delivery, change an adaptive WebRTC source or create another encoder.

## Observability contract

The implementation MUST expose fixed-label metrics equivalent to:

```text
neko_media_hls_leases{mode,state}
neko_media_hls_bootstrap_total{mode,result}
neko_media_hls_requests{resource,state}
neko_media_hls_requests_total{mode,resource,result}
neko_media_hls_request_duration_seconds{mode,resource}
neko_media_hls_blocked_reloads{state}
neko_media_hls_packagers{variant,state}
neko_media_hls_packager_starts_total{variant,result}
neko_media_hls_generations_total{variant,reason}
neko_media_hls_objects{variant,kind}
neko_media_hls_retained_bytes{variant,kind}
neko_media_hls_published_bytes_total{variant,kind}
neko_media_hls_publish_delay_seconds{variant,kind}
neko_media_hls_drops_total{variant,stage,kind,reason}
```

Allowed `mode`, `resource`, `state`, `variant`, `kind`, `result`, `stage` and `reason` values are compile-time allowlists. No session, public ID, source-generated string, URL, query or credential is a label. `neko_media_deliveries`, provider-subscription metrics, process CPU/RSS and capture-pipeline metrics remain the cross-backend evidence.

Structured logs MAY include the existing session ID at controlled lifecycle points but not on every object request. Required events are packager start/ready/idle-stop/failure, rendition removal/rejoin, generation transition, lease open/pause/resume/close and limit rejection. Public IDs and credentials are redacted by construction.

## Configuration and rollback

The planned server namespace is `media.hls` with these deployment controls:

- `enabled: false` by default;
- exact `allowed_origins`;
- explicit `trusted_proxies` CIDRs;
- `max_leases: 128` and `max_requests: 512`;
- modes `hls` and `ll-hls`, with neither selected automatically;
- the fixed three-variant profile and memory/queue limits above.

Repository deployment must use a separate sanitized Compose overlay. Omitting it leaves routes, negotiation, packagers and client choices absent. Rollback is: select WebRTC in the client, omit the HLS overlay, recreate only the Neko service, verify the route returns `404`, and confirm unchanged WebRTC A/V/control. No base Compose or `master` change is part of this prototype.

## Repository implementation phases

### Phase 1 — access and deterministic media model

- add default-off validated config and capability descriptors;
- add the HLS event messages, ten-second bootstrap ticket, playback-lease/cookie model and exact pre-route security helpers;
- add pure playlist/object/generation models, fixed limits and golden conventional/LL-HLS fixtures;
- add authorization, replay, expiry, path/query, cookie-scope, range, rate and redaction tests;
- do not register a usable media route, start a packager or add a client player yet.

Repository status on 2026-09-23: Phase 1 is implemented on `testing`. The server has validated default-off `media.hls` configuration and authenticated current/legacy event negotiation, digest-only ten-second one-use tickets, independent digest-only 30-second path-scoped cookie leases, exact pre-route security/path/query/range/rate/redaction helpers, immutable generation/object models with the fixed memory/count ceilings, and deterministic master/conventional/LL-HLS playlist renderers backed by golden fixtures. Negotiation re-resolves the live connected `CanWatch` session, passive/view-only sessions may negotiate receive media, session loss or permission change invalidates pending tickets, and HLS negotiation payloads are excluded from WebSocket payload logs. The planned bootstrap/resource paths are models only and are deliberately not registered with the HTTP manager. No provider subscription, media conversion, packaging, HTTP delivery, browser player, deployment overlay or backend selection was added. Focused tests are present but were **NOT EXECUTED IN CODEX**; target-server verification remains required.

### Phase 2 — shared packager and HTTP delivery

- add the one-audio/three-video shared packager set consuming bounded provider subscriptions;
- add H.264/AAC transcoding, aligned fMP4 muxing, generation/discontinuity handling and the bounded memory store;
- register the bootstrap/resource routes only when enabled and connect per-session deliveries to the shared set;
- add fixed metrics, request cancellation, slow-reader isolation and shutdown cleanup;
- retain WebRTC and WebCodecs behavior unchanged.

### Phase 3 — isolated passive client

- add pinned local player support and Apple-native versus MSE feature detection;
- add explicit manual `HLS` and `Low-Latency HLS` selection only when advertised, while keeping absent/invalid state on WebRTC;
- bootstrap without URL credentials, preserve unrelated query/fragment state, and implement pause/revoke/terminal cleanup;
- expose receive-only limitations and never start another backend automatically;
- preserve native iOS fullscreen/PiP where the selected player supports it.

### Phase 4 — deployment, observability and grouped acceptance

- add a separate opt-in Compose overlay and credential-safe evidence collector;
- build and validate exact images on the target server;
- execute the security, role, device, latency, resource and induced-isolation matrix below;
- keep the feature default-off unless a later explicit operator decision promotes it.

## Target-server acceptance matrix

All evidence is tied to one exact `testing` commit, image digest and sanitized result directory. Codex does not execute these checks.

### Automated and protocol gates

1. Client tests/type/lint/build and full relevant Go tests/build pass in validation containers.
2. Golden playlists pass strict parser tests and, where available, Apple `mediastreamvalidator`/`hlsreport`; fMP4 boxes, timestamps, CODECS, target durations and discontinuities are independently inspected.
3. Fuzz malformed playlists/directives, paths, ticket/cookie inputs, range headers and mux input without panic, allocation escape or credential output.
4. Disabled routes return `404`. Enabled public HTTPS and direct-origin probes cover missing/wrong Origin, cleartext, forged forwarded headers, malformed/replayed/expired tickets, missing/wrong cookies, traversal, duplicate/unknown queries, ranges, limits and response headers.
5. Logs, metrics, browser URL/history, `Referer` and captured proxy access logs contain no ticket, cookie, view-only token or raw public lease path.

### Room, role and lifecycle gates

1. Keep one ordinary WebRTC member and one admin connected while a view-only HLS viewer joins the same room.
2. Confirm matching changing desktop/audio, member presence and private-mode pause/resume without a second room or capture.
3. Repeat all view-only input/control/chat/file/plugin/inbound-media denial probes from [`VIEW_ONLY_SHARING.md`](VIEW_ONLY_SHARING.md); changing the receive backend must not weaken one boundary.
4. Exercise HLS delivery replacement by WebRTC, logout, kick, `CanWatch` loss, compact-token service recreation and shutdown. New bytes stop within one second, cooperative display clears within two seconds, and the bounded already-fetched-data caveat is recorded.
5. Confirm ordinary/admin WebRTC control and the explicit WebCodecs path still work after HLS teardown.

### Device and playback gates

For each row in the device matrix, record exact device/runtime, mode, playback path, codec probe, startup, audio, A/V sync, fullscreen/PiP availability, refresh, foreground/background return and terminal failure behavior. Run at least ten cold starts and one ten-minute stream per required mode. Conventional HLS on the actual Smart-TV/constrained target is mandatory before a compatibility claim; LL-HLS failure must remain explicit and may not silently relabel conventional HLS.

### Latency, adaptation and failure gates

1. Collect at least 30 synchronized visible-clock samples per mode/device and report median/p95/max against the fixed targets.
2. Shape a viewer through rates above high, between medium/high, between low/medium and below low; record rendition switches, stalls and recovery.
3. Restart one source generation and change resolution once. Every client must cross an explicit discontinuity without old-generation mixing or permanent black/audio output.
4. Pause reads, read one byte at a time, abort segments and issue bounded future blocking reloads from one client. The offender alone may stall/fail.

### Resource and cross-backend isolation gates

Measure no viewer, one HLS viewer, three viewers on the same rendition, three viewers across variants, and one induced slow viewer for five minutes each:

- the second/third viewer must increase leases/HTTP traffic but not packager count;
- retained HLS bytes stay at or below 64 MiB and segment storage causes zero persistent-disk writes;
- all three renditions together remain below 80% sustained total CPU on the target host and leave the Neko container healthy with zero restarts;
- adding two viewers to an existing rendition adds no more than 10 percentage points process CPU and 32 MiB RSS over the one-viewer steady state;
- healthy WebRTC/WebCodecs viewers show no HLS-caused delivery close, source switch, queue-drop increase, A/V interruption or control regression;
- after the last lease plus 15 seconds, HLS subscriptions, workers, requests and retained bytes return to zero.

Failure of a numeric gate keeps the prototype default-off and is evidence for tuning, not permission to weaken the recorded result.

## Fixed decisions and deferred choices

Fixed for version 1:

- H.264 High 3.1 plus stereo AAC-LC in separate fMP4 renditions;
- six-second parents, one-second LL parts, two-second aligned GOPs;
- one shared bounded packager set, three video workers and one audio worker;
- per-viewer central delivery plus short-lived path-scoped HttpOnly cookie;
- same-origin HTTPS direct-origin delivery with no CDN/shared cache;
- memory-only bounded retention and no automatic fallback;
- explicit conventional versus low-latency modes and exact acceptance gates.

Still evidence-led after implementation:

- actual device pass/fail results and whether LL-HLS materially helps those devices;
- measured final BANDWIDTH values and any target-specific encoder tuning;
- whether a compatible provider-side H.264/AAC source can remove transcode cost without adding capture;
- whether DASH expands the actual target matrix;
- any eventual automatic backend policy.

## Standards and primary references

- [RFC 8216 — HTTP Live Streaming](https://www.rfc-editor.org/rfc/rfc8216)
- [HTTP Live Streaming 2nd Edition, draft-pantos-hls-rfc8216bis-22](https://datatracker.ietf.org/doc/draft-pantos-hls-rfc8216bis/22/) — current work in progress used for LL-HLS tags and server profile
- [Apple HLS Authoring Specification for Apple Devices](https://developer.apple.com/documentation/http-live-streaming/hls-authoring-specification-for-apple-devices/)
- [Apple HLS Authoring Specification appendixes](https://developer.apple.com/documentation/http-live-streaming/hls-authoring-specification-for-apple-devices-appendixes)
- [Apple: Enabling Low-Latency HLS](https://developer.apple.com/documentation/http-live-streaming/enabling-low-latency-http-live-streaming-hls)
- [W3C Media Source Extensions](https://www.w3.org/TR/media-source-2/)
- [hls.js compatibility and feature-detection contract](https://github.com/video-dev/hls.js/blob/master/README.md)
