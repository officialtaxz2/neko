# HLS / Low-Latency HLS passive delivery contract

Status: **Phases 1–3 and Phase 4 assets implemented on testing. Exact-68 client
tests/type/build, scoped image preparation and healthy activation passed.
Checkout/live were then exact-68 with conventional HLS enabled. HLS playback
worked after Retry per operator report; reliable first start and wider live
acceptance remain open. Read-only diagnosis passed with one not-ready and one
successful bootstrap, generation-1 delivery and no sampled crash. The saved
timing summary passed: one request approximately 24 seconds and one at most
1 ms. A server-only candidate raises conventional startup allowance to 28
seconds within the existing 30-second HTTP/client limits. Exact-8f54970f focused
target preparation subsequently passed with Prepare-Exitcode 0, all 13 selected
native checks and both candidate images. Target checkout is exact-8f; live was
exact-68 at preparation. After the requested activation, the operator reports
an apparently frozen first picture, then moving HLS video/audio after reload.
Read-only diagnosis then passed with Diagnostic-Exitcode 0 and matching
prepared/live exact-8f image IDs: healthy container, two successful bootstraps,
all four workers in one startup generation and 135 successful segments,
without a sampled exit/OOM or fixed error. Browser first-start state remains
unknown. A client-only start-order candidate requests autoplay after attachment
without waiting for canplay. Exact-7dcc3c5e preparation subsequently passed
with Prepare-Exitcode 0, both candidate images and private evidence recorded,
live exact-8f retained. Earlier client tests/type/build are covered by final
helper success; their counts are outside the supplied build tail. Checkout
is now exact-7dcc3c5e. After the requested activation, moving first picture after
about 20 seconds then a frozen picture/reload are operator-reported. The subsequent
progress diagnostic passed, confirming the exact-7dcc prepared/live image and
healthy service. Both valid 12.012-second samples were idle; the operator confirmed
HLS was closed/logged out. This does not capture the freeze or establish a
WebRTC-join cause. The corrected `4593a6f9` progress helper subsequently passed:
one active lease, all four workers, continuing audio/video publication and
successful HTTP media requests over 12.011 seconds, with no interval restart;
VP8 rows are retained. The operator confirms frozen picture/audio loss during
this sample without reload, then a later "HLS HTTP connection failed; retry
manually". The freeze is captured despite server flow; the later HTTP failure
was outside the measured interval. The supplied retained browser timings show
one 1,001-ms master without response metadata, but a full 250-entry buffer omits
later failures. Status zero with body data is not failure proof. The complete
five-minute trace now captures a freeze near media time 24 seconds, audio loading
stopped after four segments and repeated seeks while video delivery continues.
Exact-8741 preparation subsequently passed: old defects reproduced, 66 client
tests/type/build, repaired wire/HLS-package checks and candidate images passed.
Checkout/live deployment is now exact-8741: activation passed with Activate-Exitcode 0
and a healthy candidate image. The operator confirmed at least five minutes of
moving HLS picture/audio without Retry/reload while WebRTC continued working.
The bounded PC/Helium sustained-playback gate is passed; [remaining grouped
acceptance](HLS_PLAYLIST_WINDOW_REPAIR_2026-10-07.md#remaining-acceptance)
stays open. No new server command or rebuild is needed. Supplied target/browser
results, NOT EXECUTED IN CODEX; the corrected saved-startup helper has no fresh
execution. Repeated cold-start reliability remains open.
Authorization/lifecycle and grouped
device/resource acceptance
remain pending; no automatic selection or full HLS acceptance claim exists**.

This document is the normative contract for the first default-off passive/view-only HTTP-streaming prototype. It specializes the encoded-source/subscription and participant-delivery boundary in [`MEDIA_SUBSCRIPTION_BOUNDARY.md`](MEDIA_SUBSCRIPTION_BOUNDARY.md). It does not authorize a second desktop capture, a stable-deployment change or a claim that any untested device supports the proposed path.

The key words **MUST**, **MUST NOT**, **SHOULD** and **MAY** describe the intended implementation contract.

## Scope and non-goals

Version 1 provides receive-only live audio/video for an already authenticated participant in the existing logical Neko room. It has two explicit modes:

- `hls`: conventional live HLS for the broadest passive-device compatibility;
- `ll-hls`: the same media objects plus Low-Latency HLS playlists, partial segments and blocking reloads.

Both modes remain unavailable unless the dedicated default-off server flag enables them through the separate `docker-compose.hls.yaml` overlay. It initially advertises only conventional `hls`; `ll-hls` requires its recorded public-protocol/RTT prerequisite and explicit enablement. WebRTC remains the default interactive backend. `webcodecs-ws` remains an explicit independent receive prototype. There is no automatic fallback between any backend.

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
        | one bounded high-video subscription, common encoded-unit fan-out
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

There is one process-wide packager set for the single active Neko room, not one encoder or muxer per viewer. One audio worker owns one audio provider subscription. One high-video provider subscription feeds the same immutable encoded units to all three video decode/encode/mux workers. All video variants therefore begin from the same first source keyframe and provider clock. Both subscriptions and all four workers are shared across `hls` and `ll-hls` viewers. Regular and low-latency modes render different playlist views over the same retained objects.

The 2026-10-05 source-phase diagnosis reproduced the prior independent-source
admission defect in three cold runs. The common-video-input correction's A/B and
full preparation subsequently passed at `a7ffb8b1`. Later client/playlist repairs
culminated in supplied exact-8741 preparation/activation and at least five
minutes of moving PC/Helium conventional-HLS picture/audio without Retry/reload
alongside continuing WebRTC. This supersedes the original
one-video-subscription-per-rendition topology. It does not claim the failed live
source phases were measured or that grouped/native/TV/LL-HLS/resource acceptance
passed. See [the current comparative assessment](STABILITY_REVIEW.md#comparative-fork-and-transport-review--2026-10-07)
and [remaining acceptance](HLS_PLAYLIST_WINDOW_REPAIR_2026-10-07.md#remaining-acceptance).

Current implementation constraint: `workerFormatMatches` requires the common
`high` input to be VP8 at 1280×720 and 25/1 fps, with stereo 48-kHz Opus audio.
Output scaling does not make arbitrary input resolution/codec changes supported.
Distinguish a supported same-format generation restart from an unsupported input
format in source-change acceptance. Transparent arbitrary-resolution recovery
and a shared-decode/raw-source redesign are undecided future work; neither is
implied by the existing broad source-change test wording.

The first authorized lease starts all three configured variants so the multivariant playlist is internally consistent and can adapt immediately. The hard version-1 maximum is three video variants plus one audio rendition. When the last unpaused lease closes, the set enters a 15-second idle grace and then closes subscriptions, workers and retained objects. A new viewer after teardown starts a new packager generation.

The packager consumes the existing encoded provider. On the current VP8/Opus deployment each video worker decodes the same high VP8 input and scales/re-encodes it to its advertised output. All variant capability `source_id` values are `high`; output IDs remain `high`, `medium`, `low`. This preserves provider PTS/DTS and avoids independently phased tier captures. It reduces HLS provider subscriptions from four to two, but three high-resolution decoders replace the prior differently sized inputs; CPU/RSS costs must be remeasured. A future shared raw-frame tee remains a separate optimization. No second desktop capture or per-viewer encoder is created.

Per-viewer ownership remains in `MediaDeliveryManager`:

1. the authenticated current session is resolved;
2. current `CanWatch` is required;
3. one credential-free `MediaLease` and HLS delivery are prepared, and backend open waits for the mode-specific ready generation within the fixed startup deadline;
4. that delivery replaces any prior primary delivery for the session;
5. the delivery holds a reference to the shared packager set but never owns its capture;
6. logout, deletion, disconnect after the existing grace, `CanWatch == false`, replacement or shutdown revokes it centrally.

`IsConnected` continues to describe the event connection. `IsWatching` becomes true only after an authenticated media playlist has been served from a ready generation, and false when that delivery is paused, failed or closed. Individual segment GETs are not participants.

## Renditions, timestamps and keyframes

The fixed HLS output ladder on the initial target-server profile is:

| ID | Encoded HLS output geometry/rate | H.264 video target | Initial `AVERAGE-BANDWIDTH` including audio | Initial peak `BANDWIDTH` |
| --- | --- | ---: | ---: | ---: |
| `high` | 1280x720 at 25 fps | 3,000 kbit/s | 3,128,000 | 4,000,000 |
| `medium` | 854x480 at 20 fps | 1,100 kbit/s | 1,228,000 | 1,500,000 |
| `low` | 640x360 at 15 fps | 365 kbit/s | 493,000 | 650,000 |

The packager MUST require a complete high-source provider `FORMAT` matching the configured 1280x720/25-fps VP8 input and a consistent nonzero generation. All renditions use that single input. Their output width, height and frame rate MUST match their emitted H.264 caps and the advertised table; a mismatch is rejected. The packager MUST NOT publish zero or guessed output dimensions. `AVERAGE-BANDWIDTH` and `BANDWIDTH` MUST be replaced with measured values before acceptance and must satisfy the Apple live-stream guidance recorded in the sources below.

All renditions use a closed two-second GOP. With the current rates this means 50, 40 and 30 frames respectively. Their shared source-unit sequence and first PTS establish the same GOP phase; timestamp rebasing of individual renditions or arrival-time substitution is forbidden. IDRs align to the same two-second presentation boundaries across variants. A six-second parent segment begins on an aligned IDR; the parts at 0, 2 and 4 seconds are marked `INDEPENDENT=YES`. Other parts are not advertised as independent.

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

An LL-HLS playlist accepts only decimal `_HLS_msn` and `_HLS_part` directives; `_HLS_part` without `_HLS_msn` is `400`. The server blocks for at most seven seconds for a valid near-future request, wakes on publication/revocation/pause/shutdown, and returns `503` with `Retry-After: 1` when the requested update is still unavailable. Requests more than two parent sequences or three parts ahead in the current/future parent are `400`. Part indices are 0–5; an exact requested index 6 is normalized to part 0 of the next parent before checking the parent bound. A stale previous-parent part request is served from the current playlist and is not compared with the new parent's reset part index. Larger indices, `_HLS_skip` and all unknown query fields are rejected in version 1.

### Required playlist and object shape

The multivariant playlist contains `EXTM3U`, `EXT-X-VERSION:7`, `EXT-X-INDEPENDENT-SEGMENTS`, one `EXT-X-MEDIA:TYPE=AUDIO` entry, and one `EXT-X-STREAM-INF` per ready video rendition. Every stream entry includes measured `BANDWIDTH`, `AVERAGE-BANDWIDTH`, exact `RESOLUTION`, `FRAME-RATE`, the emitted H.264 plus AAC `CODECS` values, and the shared audio group. It contains no absolute URI, credential, session ID or unavailable rendition.

A conventional media playlist contains at least `EXTM3U`, `EXT-X-VERSION:7`, `EXT-X-TARGETDURATION:6`, monotonically increasing `EXT-X-MEDIA-SEQUENCE`, the current `EXT-X-DISCONTINUITY-SEQUENCE`, `EXT-X-SERVER-CONTROL:HOLD-BACK=18`, `EXT-X-MAP`, `EXT-X-PROGRAM-DATE-TIME` and exactly three completed `EXTINF` entries once warm. Conventional backend open becomes ready only after the audio and every advertised video rendition have an init section plus three complete aligned parents. The current server-only candidate waits for at most 28 seconds (24 seconds in the deployed exact-68 image); expiry fails bootstrap with `503` and leaves the prior primary delivery unchanged. The same constant also bounds conventional readiness during private lease resume. The extra four seconds are a bounded startup allowance pending target validation, not a change to segment duration, playback hold-back or the measured acceptance targets below. The outer client bootstrap and server response-write limits remain 30 seconds; earlier work or request cancellation can still end a request sooner.

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

The shared high-video input and the audio worker each request a provider queue
capacity of 64 with `drop_newest`: two provider subscriptions feed four workers.
Internal worker hand-off is at most eight media events per rendition;
published-object notification is at most two pending notifications because
playlist snapshots can be regenerated from the store. No queue waits in
provider fan-out.

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

Repository status on 2026-09-23: Phase 2 is implemented and statically reviewed on `testing`. When and only when `media.hls.enabled=true`, startup now registers a central `hls` delivery backend and the authenticated bootstrap/keepalive/resource routes. One process-wide packager consumes four bounded exact-source subscriptions and shares H.264 High 3.1 video plus stereo AAC-LC fMP4 objects across conventional and low-latency leases. Every worker requires a complete first provider `FORMAT`, an exact advertised dimension/rate match and one consistent nonzero provider generation before accepting media. It publishes aligned one-second parts, two-second IDRs and six-second parents into the fixed memory/count store, maintains explicit generations/discontinuities, evicts only objects beyond the advertised retention floors under aggregate pressure, removes stalled renditions, rejoins on a fresh keyframe and stops after the final unpaused lease plus the 15-second grace. Per-session deliveries, sliding scoped cookies, initially paused attachment, Private Mode with readiness-before-unpause resume, replacement/revocation, bounded blocking reloads, gzip/HEAD/range responses, request deadlines, slow-reader isolation, fixed metrics and credential-safe logs follow the contract above. No client player, client selection, deployment overlay or automatic fallback was added. Focused tests are present but were **NOT EXECUTED IN CODEX**; target-server build/tests, independent fMP4/playlist validation and all runtime/device/resource gates remain required.

### Phase 3 — isolated passive client

- add pinned local player support and Apple-native versus MSE feature detection;
- add explicit manual `HLS` and `Low-Latency HLS` selection only when advertised, while keeping absent/invalid state on WebRTC;
- bootstrap without URL credentials, preserve unrelated query/fragment state, and implement pause/revoke/terminal cleanup;
- expose receive-only limitations and never start another backend automatically;
- preserve native iOS fullscreen/PiP where the selected player supports it.

Repository status on 2026-10-04: Phase 3 is implemented on `testing` and statically reviewed; all executable checks remain **NOT EXECUTED IN CODEX**. `client/src/neko/hls/` owns the controller, strict capability/offer/lease/playlist/URL validation and lazy MSE adapter. The full local `hls.js` build is pinned to **1.7.3** in both manifests, with its locally bundled worker; no CDN script is used. Apple HLS stays on the existing native video element. Other browsers require exact H.264/AAC MSE support, or positive native container/codec probes plus successful real playback; these probes are not device acceptance evidence.

The manual settings and passive playback selector expose only authenticated advertised HLS modes to view-only viewers or admin diagnostics. Exact `?media=hls` and `?media=ll-hls` diagnostic overrides take precedence without writing storage. Absent/invalid selection remains WebRTC; a stored but unavailable HLS mode terminates visibly, with manual Retry HLS and Use WebRTC actions. Selector reload removes only `media` and preserves unrelated queries and the view-only fragment. Selected HLS event sockets receive their session ID without creating an unused WebRTC peer. No replacement input or automatic transport fallback is added.

Bootstrap sends the one-time ticket only in the same-origin HTTPS POST body. The controller retains the opaque public lease URL only in memory and uses the path-scoped HttpOnly cookie. A serialized one-second nonblocking master watchdog surfaces native-loader authorization failures; HTTP deadlines, readiness/stall limits and same-request retries are bounded. Keepalive renews every 15 seconds. The MSE loader permits three active media requests and twelve waiting descriptors, leaving one lease slot for the watchdog/keepalive; playlist/object bodies and front/back buffers are bounded. Errors reaching UI/logs use fixed text and omit tickets, cookie values and lease URLs. Native initial master/child playlists are checked before assigning `src`; subsequent native requests remain owned by the browser and the strict server route.

The initial player readiness deadline is 30 seconds and is disarmed on the
first current-player `canplay` or `playing` event; later buffering cannot
reactivate it. It is armed before attachment so immediate readiness is covered.
The separate 20-second progress watchdog remains active after first readiness
when playback is requested and the element is not paused. The follow-up
[client stability review](HLS_CLIENT_STABILITY_REVIEW_2026-10-05.md) gates that
watchdog until readiness, resets its clock on deliberate resume, separates
HTTP/readiness failure counts and prepares eight new regression cases. This
exact-68 candidate passed the supplied isolated 60-test/type/build gate and
scoped image preparation (Image-Prepare-Exitcode 0), with the working exact-73
service unchanged during preparation. Subsequent exact-68 activation passed
healthy with Start-Exitcode 0. The operator reports HLS works after Retry;
reliable first start and wider browser validation remain open. The next step
is the saved-file bootstrap duration summary: read-only diagnosis passed with
one not-ready and one successful bootstrap plus generation-1 segment delivery,
but does not identify the exact error cause or elapsed time. Retain exact-68.
The earlier client-only correction follows
the first-picture/premature-timeout report at a7ffb8b1. Its isolated target gate
reproduced the old fault and passed all 52 repaired client tests/type/build;
scoped image preparation passed with Client-Image-Exitcode 0, a fresh client
build and cached unchanged server/runtime layers, leaving enabled a7ff running.
Default-off exact-73 deployment then passed healthy with Baseline-Exitcode 0,
2/2 disabled-route probes and the reported normal-browser checkpoint. Same-image
activation then passed with Enable-Exitcode 0, healthy service and 19/19 denial
probes. PC/Helium playback was reported after Retry and one frozen-picture
reload, with WebRTC unaffected; detailed "HLS failed" text/timing is missing.
Conventional warm-up/18-second hold-back and a player failure remain distinct.
Read-only diagnosis then passed with a healthy active exact-73 service, six
bootstrap successes, 394 successful segments and one not-ready bootstrap; no
sampled exit/OOM/fixed error markers. These cumulative counters do not correlate
the frozen picture or confirm an uninterrupted room-event interval. The operator
cannot confirm that interval and mentions possible random reconnects/room actions.
Consolidate the remaining checks into one bounded later step; no extra ad-hoc
operator test requested now. Startup/recovery and sustained acceptance remain open in
[the readiness repair record](HLS_CLIENT_READINESS_REPAIR_2026-10-05.md).

The selected legacy event bridge now emits `media/hls/state` with `{version:1, backend:"hls", paused:boolean}` before `system/init` and on authoritative room settings updates. It derives private pause from `PrivateMode && !IsAdmin`, independently of control locks. Private pause, stop, detach, logout, replacement and terminal failure invalidate callbacks, destroy MSE, remove listeners, pause the element, remove `src`/`srcObject` and call `load()` to discard URL/MSE buffers. WebRTC recovery never uses this cleanup helper. Private resume reuses the still-valid lease and waits for fresh packaging; Safari autoplay still has one muted retry and the explicit Play gesture. Native fullscreen and supported standard/WebKit PiP remain available.

Static integration also corrected cookie scope for `server.path_prefix`, ordered prefix stripping before log/CORS classification, and made fixed six-part LL-HLS parent rollover compatible with the pinned player's reload directives. Focused client tests cover shared server golden playlists, hostile URL/query/credential input, selection/fragment preservation, disabled/ineligible negotiation, private pause/resume, native revocation, stale bootstrap completion and autoplay fallback. Go checks cover exact selection, prefix/CORS/cookies and rollover bounds. The validation container now supplies HLS fixtures to client tests. No target build, playback, latency, TV-compatibility or resource claim follows from these source checks.

The mandatory final review and the operator's untested TV/event-disconnect report are tracked in [`STABILITY_REVIEW.md`](STABILITY_REVIEW.md). Phase 4 must run the accumulated automated/security/role/device/resource gates before any acceptance or promotion decision.

### Phase 4 — deployment, observability and grouped acceptance

- add a separate opt-in Compose overlay and credential-safe evidence collector;
- build and validate exact images on the target server;
- execute the security, role, device, latency, resource and induced-isolation matrix below;
- keep the feature default-off unless a later explicit operator decision promotes it.

Repository status on 2026-10-04: the separate `docker-compose.hls.yaml`, private collector, credential-free HTTP probes, exact-test/image preparation and captured-image enable/rollback helpers are implemented. [Validation](HLS_LL_HLS_VALIDATION.md), [host Caddy review](HLS_LL_HLS_CADDY.md), [observability](HLS_LL_HLS_OBSERVABILITY.md) and the pending [result template](HLS_LL_HLS_RESULTS_TEMPLATE.md) define the next target checkpoint. The overlay retains the fixed adaptive source profile, accepts only reviewed HTTPS/proxy values and leaves base Compose unchanged. Both modes use the whitespace-separated environment value `hls ll-hls`, not CSV.

The [integrated static review](STABILITY_REVIEW_2026-10-04.md) records source/configuration items 1–7 and fixes optional notification error handling plus long private-pause keepalive: a valid paused lease can renew while every media request stays denied. Rate limits, expiry and revocation still apply; renewal does not hold packager demand. Detailed target dependency-advisory classification remains open. New client/Go regression tests and request-parser fuzzing are present, but were **NOT EXECUTED IN CODEX**. Exact target tests/images, Caddy log/cancellation checks and every enabled runtime/device/resource gate remain pending. Phase 4 full acceptance is not claimed.

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
- [Pinned hls.js 1.7.3 player/loader/worker API](https://github.com/video-dev/hls.js/blob/v1.7.3/docs/API.md)
