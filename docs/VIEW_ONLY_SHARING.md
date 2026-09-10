# Server-enforced view-only sharing

Status: **implemented and statically reviewed on `testing`; target-server build, server tests and three-role runtime validation pending**.

This first implementation deliberately keeps WebRTC as the receive-media path. Authorization belongs to the member/session profile and ingress checks, not to WebRTC, so the same passive identity can later be used by an HLS/LL-HLS backend without creating another desktop or room.

## Share credential and URL

The optional multi-user setting `member.multiuser.view_only_token` must be either empty or exactly 64 hexadecimal characters (256 bits). Generate it outside Git:

```bash
openssl rand -hex 32
```

Store the result as `NEKO_VIEW_ONLY_TOKEN` in the ignored deployment `.env`. Never commit or print the resolved Compose configuration. The viewer URL is:

```text
https://neko.example/#/watch/<64-hex-token>
```

The token is a bearer credential. The URL fragment does not enter the HTTP request path, query string, access log or `Referer` header. The client validates the shape locally and carries the credential as the `neko-view.<token>` WebSocket subprotocol; the legacy adapter echoes that protocol and authenticates the session internally. The token is not copied into Vuex or another application storage field. It can still remain in browser history, bookmarks, screenshots and copied links, so every recipient and device holding the URL must be trusted. TLS is required for an Internet-facing deployment, and reverse proxies must not log `Sec-WebSocket-Protocol` values.

An empty value disables share-link login. Startup rejects a malformed token or a token equal to either configured member password. A syntactically invalid WebSocket subprotocol receives HTTP 400; a correctly shaped but disabled or incorrect token is upgraded only far enough to receive the normal authentication disconnect.

## Lifetime and revocation

There is no clock-based expiry. The credential is valid for the lifetime of the configured deployment until the operator rotates or removes it and recreates the Neko service. View-only sessions are intentionally excluded from `session.file` persistence, and stale serialized view-only sessions are ignored on load. Therefore service recreation is the hard revocation boundary for already connected viewers.

To revoke:

1. replace `NEKO_VIEW_ONLY_TOKEN` with a newly generated token, or set it to an empty value;
2. run the same Compose profile's configuration check;
3. recreate the Neko service;
4. distribute a new fragment URL only when sharing remains enabled.

Do not reuse the member password, admin password or a previously disclosed share token.

## Enforced capability boundary

Authentication creates an ephemeral member profile marked `is_view_only`. Session creation and update normalize that marker and forcibly remove admin, host, microphone/media-share, clipboard and inactive-cursor-send capabilities even if another provider or API payload contains conflicting fields.

The boundary is enforced at every current client-to-server entry point:

- HTTP middleware denies profile, session, member and plugin APIs; room watch/configuration reads needed for playback remain available, while control, keyboard, clipboard, uploads and admin routes fail their capability middleware;
- the current and legacy WebSocket adapters use a receive-media signalling allowlist and drop every control, keyboard, touch, clipboard, chat-send, file, admin and unknown plugin event from a view-only session;
- the modern WebRTC data channel accepts only its ping opcode, and the legacy data channel accepts no view-only input;
- server-side host assignment rejects a passive target, including an admin attempting to grant it control;
- inbound WebRTC tracks are stopped, so a forged microphone/camera publication cannot reach the room;
- file-transfer updates expose neither capability nor file names, and upload/download/delete endpoints return 403;
- plugin HTTP routes are interactive-only; chat may be received at the server protocol layer but cannot be sent, and the video-only UI does not expose chat or other room controls.

Denied WebSocket/data-channel attempts are ignored and recorded as warnings rather than disconnecting an otherwise valid passive viewer. Denied HTTP actions return 403. The UI additionally becomes video-only and retains playback, volume, mute, fullscreen and picture-in-picture controls, but UI hiding is not part of the authorization decision.

## Boundary audit record

The implementation audit covered:

- multi-user authentication and login-lock evaluation;
- session creation, profile updates, persistence, host selection and inactive cursors;
- authenticated HTTP APIs, room middleware and batch re-routing;
- current WebSocket dispatch before core and plugin handlers;
- legacy WebSocket translation and legacy HTTP file routes;
- modern and legacy WebRTC data-channel input plus inbound media tracks;
- chat, file-transfer and Open-in-App plugin boundaries;
- member serialization and the Vue client/store/control surface.

The fixed server profile is the authority. `is_view_only` is backend-neutral and must remain authoritative when a later passive media transport is added.

## Target-server build and focused tests

Run only on the real target server at the exact candidate commit:

```bash
git switch testing
git pull --ff-only origin testing
git status --short --branch
git rev-parse HEAD

cd client
npm ci
npm test
npm run lint
npm run build

cd ../server
go test ./pkg/types ./pkg/auth ./internal/member/multiuser ./internal/session ./internal/http/legacy ./internal/websocket ./internal/webrtc
./build
cd ..
```

The focused Go packages cover profile normalization, middleware denial, token validation/authentication, non-persistence, current/legacy WebSocket allowlists and WebRTC data-channel denial. They do not replace the adversarial runtime matrix below.

If the target host has no Go installation, use the repository server image for the same build and tests:

```bash
VALIDATION_IMAGE="my-neko/server-validation:$(git rev-parse --short=8 HEAD)"
docker build -t "$VALIDATION_IMAGE" ./server
docker run --rm "$VALIDATION_IMAGE" \
  go test ./pkg/types ./pkg/auth ./internal/member/multiuser ./internal/session \
  ./internal/http/legacy ./internal/websocket ./internal/webrtc
```

Prepare the ignored `.env` without exposing the generated value, preserve the current image, and recreate the same base or adaptive Compose profile intended for the checkpoint:

```bash
docker image tag my-neko/brave:latest my-neko/brave:pre-view-only
./build my-neko/base:latest -y
./build my-neko/brave:latest -y
docker compose config --quiet
docker compose up -d --force-recreate
docker compose ps
```

If the adaptive overlay is active, use both Compose files consistently for `config`, `up` and rollback. Do not use plain `docker compose config` output because it expands secrets.

## Three-role target-server matrix

Use three simultaneous participants in the same shared desktop: ordinary member `M`, admin `A`, and share-link viewer `V`. Record the commit, image ID, active Compose profile, browsers/devices and all three session IDs.

| Check | Member `M` | Admin `A` | View-only `V` |
|---|---|---|---|
| Join same room and receive audio/video | allowed | allowed | allowed |
| Playback/mute/volume/fullscreen/PiP | allowed | allowed | allowed |
| Request or receive control | according to room policy | allowed | denied |
| Mouse, scroll, buttons, keyboard, touch | host only | host/admin rules | denied |
| Clipboard read/write | permitted host only | permitted host | denied |
| Upload/drop/dialog or file list/download/delete | configured rights | configured admin rights | denied; no file names |
| Publish microphone/camera track | profile/control policy | profile/control policy | denied/stopped |
| Send chat or arbitrary plugin event | configured rights | configured rights | denied |
| Settings, member/session or plugin HTTP mutation | capability policy | allowed where admin | 403 |
| Admin lock/take/give/revoke | denied | allowed | denied |

Execute the following phases without changing roles:

1. Prove the ordinary/admin baseline: `M` requests/releases control; `A` takes, grants and revokes it; locks still work.
2. Join `V` through the fragment link and verify a video-only surface, audio/video playback and the same visible desktop content as `M` and `A`.
3. Attempt to grant `V` control from `A`; verify `V` never becomes host and no input reaches X11.
4. Attempt mouse, scroll, keyboard and touch input from `V`, including crafted current and legacy WebSocket messages where the inspection setup permits it. Verify warning logs and no desktop effect.
5. Using a separate API test session authenticated with the view-only credential, attempt profile/session/member/plugin, control, keyboard, clipboard, upload and file-transfer calls. Verify 403 and confirm no file names are returned. Keep bearer values out of shell history and evidence captures.
6. Attempt to publish an audio track from the view-only session with an instrumented client; verify the server stops it and other participants receive no view-only microphone audio.
7. Refresh and transiently disconnect/reconnect `V`; verify it remains view-only and neither healthy participant reconnects or loses control.
8. Rotate the token and recreate the service. Verify the old link cannot log in, connected old view sessions are gone, the new link can watch, and member/admin credentials still behave normally.
9. Remove the token and recreate once more if sharing should finish disabled; verify ordinary/admin access remains unchanged.

At this same coherent `testing` checkpoint, execute every phase in [`IOS_RECOVERY.md`](IOS_RECOVERY.md). View-only acceptance does not prove no-reload iOS recovery, and a Safari Play gesture remains a separately recorded autoplay-policy fallback.

## Acceptance criteria

The view-only block becomes target-server verified only when:

- all three participants see the same logical room and desktop;
- `V` receives WebRTC audio/video but every matrix action marked denied fails server-side;
- crafted current/legacy WebSocket, data-channel, HTTP/plugin and inbound-media attempts cannot cause an interactive effect or expose file names;
- `A` cannot grant control to `V`, while existing `M`/`A` control and lock semantics remain intact;
- token rotation/removal plus service recreation revokes old links and active view-only sessions;
- ordinary member, admin, mobile/touch, file-transfer, adaptive-quality (when enabled) and iOS recovery regressions required by the grouped checkpoint pass.

Until that evidence exists, report this feature as **implemented and statically reviewed, target-server validation pending**.

## Rollback

First disable sharing by clearing `NEKO_VIEW_ONLY_TOKEN` and recreating the current image. If code rollback is required, restore the preserved image using the same Compose profile:

```bash
NEKO_IMAGE=my-neko/brave:pre-view-only docker compose up -d --force-recreate
```

There is no database or persistent-data migration. An old image does not understand the new setting; leaving the environment variable present is inert there, but clearing it avoids ambiguity.
