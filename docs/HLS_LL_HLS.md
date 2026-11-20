# HLS and LL-HLS

HLS (about 20 s behind) and LL-HLS (about 3–5 s) are a manual alternative to WebRTC for watching. They
are receive-only: no control, no input. `docker-compose.hls.yaml` enables them (`media.hls.*`); it needs
the adaptive overlay, because HLS transcodes the `high` tier. Server code: `server/internal/mediahls/`;
client: `client/core/src/transport/hls.ts` with the MSE player in `transport/hls/player.ts`.

## Delivery

One shared packager for the room, not one per viewer: one subscription to the `high` VP8 video and one to
the Opus audio feed three H.264 workers and one AAC worker. Both modes serve the same objects.

| Rendition | Output | `AVERAGE-BANDWIDTH` / `BANDWIDTH` incl. audio |
| --- | --- | --- |
| `high` | 1280×720, 25 fps | 3,128,000 / 4,000,000 |
| `medium` | 854×480, 20 fps | 1,228,000 / 1,500,000 |
| `low` | 640×360, 15 fps | 493,000 / 650,000 |

- The input is the `high` VP8 tier, i.e. the full desktop at any size and frame rate, plus stereo 48 kHz
  Opus (`workerFormatMatches`). Every rendition is scaled to its fixed output above; a desktop that is not
  16:9 gets black bars instead of distorted pixels. A size or rate change while the packager runs (for
  example a new desktop resolution) restarts it with `format_change`.
- fMP4, closed two-second GOPs, six-second segments; LL-HLS adds one-second parts, blocking reloads,
  preload hints and rendition reports. `HOLD-BACK` is 18 s, `PART-HOLD-BACK` 3 s.
- A start is ready after three complete segments (HLS) or one complete segment (LL-HLS), within 28 s.
  A cold LL-HLS start therefore takes at least one six-second segment.
- When the last viewer leaves or pauses, the packager stops after a 15-second grace.
- Objects live only in memory: six segments per rendition, at most 64 MiB in total.

## Session and lease

1. Over the event socket the client sends `media/hls/capabilities/request` and `media/hls/create`; the
   server checks the live session and `CanWatch` and answers `media/hls/offer` with a one-time ticket
   (ten seconds, stored as a SHA-256 digest).
2. `POST /api/media/hls/session` with the ticket in the JSON body and the exact `Origin` returns the master
   playlist path and sets `__Secure-neko-hls` (HttpOnly, Secure, SameSite=Strict, scoped to that lease's
   path). No credential ever appears in a playlist URL.
3. The lease expires after 30 s without a playlist request or `keepalive`. Logout, kick, loss of
   `CanWatch`, a new primary delivery and shutdown revoke it; unknown and revoked leases get the same 404.
4. Private mode pauses the lease (`503` with `Retry-After: 1`); the client clears its buffers.

Limits: at most 128 leases and 512 requests in total (`media.hls.max_leases` and
`media.hls.max_requests` can only lower them); fixed per lease: 8 simultaneous requests, 4 of them
blocking, 12 playlists/s and 32 objects/s; blocking reload 7 s. Every response is
`Cache-Control: private, no-store`; there is no CDN or shared cache.

## Configuration

- `NEKO_MEDIA_HLS_ALLOWED_ORIGINS`: the exact public `https://` origin, without path.
- `NEKO_MEDIA_HLS_TRUSTED_PROXIES`: the reverse proxy's address as Neko sees it (for a proxy on the host,
  usually the Docker network gateway). Forwarded headers count only from there.
- `NEKO_MEDIA_HLS_MODES`: `hls ll-hls` (default) or `hls`.

HTTPS is mandatory; there is no cleartext exception, because the lease cookie is `Secure`. To turn HLS
off, remove `docker-compose.hls.yaml` from `COMPOSE_FILE`, run `docker compose down`, then
`bash setup.sh --no-build`.

## Reverse proxy

Caddy streams responses and has no read or write timeout by default, so the seven-second blocking
reloads pass unchanged; HTTP/2, which LL-HLS needs, is on by default for HTTPS sites. In the Neko site's
`reverse_proxy` (see `deploy/caddy-hls.example`):

- `header_up -Forwarded`: Neko rejects a request that carries both `Forwarded` and `X-Forwarded-*`.
- `transport http { compression off }`: MP4 objects must stay uncompressed; the browser's own gzip
  request for playlists still reaches Neko.
- No response buffering, no caching and no CORS on `/api/media/hls/*`, and no `handle_path` that strips a
  prefix, because that breaks the cookie path.
- Logs must not contain lease paths: the example's global log filter rewrites them to
  `/api/media/hls/:redacted` and drops request and response headers, also in Caddy's runtime log.

`deploy/check-ll-hls.sh` checks read-only on the server: advertised modes, HTTP/2, viewer RTT (LL-HLS
needs a p95 RTT of at most 333 ms for one-second parts) and, while an LL-HLS stream plays, held reloads and
parts. `deploy/check-hls-http.py` checks the HTTP boundary through `docker-compose.validation.yaml`.

## Metrics and tests

`neko_media_hls_*`: `leases{mode,state}`, `bootstrap_total{mode,result}`, `requests_total{mode,resource,result}`,
`request_duration_seconds`, `blocked_reloads`, `packagers{variant,state}`, `generations_total{variant,reason}`,
`retained_bytes`, `publish_delay_seconds` and `drops_total{variant,stage,kind,reason}`. No label carries a
session, lease or URL.

The browser test plays the real packager output through the production MSE player in WebKitGTK; the
geometry test feeds 1920×1080 and 1280×1024 desktops through the real codecs:

```bash
# needs node with the client dependencies, xvfb-run and python3 with WebKit2 4.1 GI bindings
cd server && go test -count=1 -tags 'hlsintegration hlsbrowser' -run 'TestRealCodecsLowLatencyReadyAfterSlowSourceStart|TestRealCodecsAcceptAnyHighSourceGeometry|TestBrowserPlayback' -v ./internal/mediahls/
```

`.github/workflows/hls_browser_test.yml` runs it on pull requests that touch the HLS server or client code.

## Open

- Desktops other than 1280×720 are tested only with real codecs in CI, not on the server. Each of the
  three video workers decodes the full `high` picture, which at 1920×1080 is 2.25 times the pixels of
  720p; the packager's CPU use there is not measured.
- Native playback (Safari, iPhone/iPad, Smart TVs) is not confirmed; an iPad's LL-HLS sessions became ready
  but did not play, cause unknown.
- Reliable cold starts with the 28-second window, and latency on the real path, are not measured.
- The decoder in front of H.264 can lose a VP8 frame silently until the next keyframe (up to 4 s).
