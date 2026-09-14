# Interactive Phase 4 target-server validation

This is the operator/Codex hand-off runbook for the default-off WebCodecs plus dedicated media-WebSocket receive prototype. Execute it only on the real target server, one block at a time. Return each block's complete terminal output before continuing so failures and deployment-specific values can be assessed without guessing.

Never paste `.env`, passwords, cookies, view-only tokens, browser storage, event payloads, packet captures, media WebSocket URLs or `Sec-WebSocket-Protocol` values from a successful attachment. The supplied HTTP check uses a fixed invalid ticket and suppresses response bodies. Store raw evidence outside the Git worktree with mode `0700`.

The stable rollback path is always the base deployment without [`../docker-compose.webcodecs-ws.yaml`](../docker-compose.webcodecs-ws.yaml). Do not change or push `master` during this checkpoint.

## 1. Exact commit and safe topology inventory

Start only with a clean `testing` worktree:

```bash
set -Eeuo pipefail
test "$(git branch --show-current)" = testing
test -z "$(git status --porcelain=v1)"
git pull --ff-only origin testing
test "$(git rev-parse HEAD)" = "$(git rev-parse origin/testing)"
printf 'commit=%s\nbranch=%s\n' "$(git rev-parse HEAD)" "$(git branch --show-current)"
docker version
docker compose version
docker compose -f docker-compose.yaml ps
NEKO_CONTAINER_ID="$(docker compose -f docker-compose.yaml ps -q neko)"
if test -n "$NEKO_CONTAINER_ID"; then
  docker inspect "$NEKO_CONTAINER_ID" --format '{{range $name, $network := .NetworkSettings.Networks}}{{printf "network=%s gateway=%s container_ip=%s\n" $name $network.Gateway $network.IPAddress}}{{end}}'
fi
```

The public browser origin and the immediate reverse-proxy peer as seen by Neko are deployment facts, not secrets. Review the output and proxy topology before choosing the allowlist. For a same-host proxy that reaches the published loopback port, the Docker bridge gateway shown to Neko is commonly the immediate peer; for a proxy container on a shared network, use its actual peer address or the narrow required CIDR. Never substitute an entire private network without topology evidence.

## 2. Exact-commit tests and builds

The validation service includes the accumulated Phase 1–3 server suite, 30-second parser fuzz target and server/plugin build:

```bash
set -Eeuo pipefail
export NEKO_VALIDATION_COMMIT="$(git rev-parse HEAD)"
docker compose -f docker-compose.validation.yaml pull client-checks
docker compose -f docker-compose.validation.yaml run --rm client-checks
docker compose -f docker-compose.validation.yaml build --pull server-checks
docker compose -f docker-compose.validation.yaml run --rm server-checks
./build my-neko/base:latest -y
./build my-neko/brave:latest -y
docker image inspect my-neko/base:latest my-neko/brave:latest --format 'image={{index .RepoTags 0}} id={{.Id}} created={{.Created}}'
```

Do not deploy if any command fails. Preserve the complete output in the private evidence directory once it has been checked for credentials.

## 3. Default-off invariance

First validate and deploy the accepted stack without the WebCodecs overlay. The adaptive overlay is included because the final grouped run must also repeat its three-viewer isolation cycle:

```bash
set -Eeuo pipefail
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml config --quiet
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml up -d --force-recreate neko
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml ps
```

Set the public base URL/origin only in the current shell; neither value contains credentials:

```bash
export NEKO_PUBLIC_BASE_URL='https://neko.example'
export NEKO_PUBLIC_ORIGIN='https://neko.example'
./deploy/check-media-websocket-http.sh disabled
```

Replace the example with the exact browser-visible origin, with no path or trailing route. Then open one ordinary URL without `media=webcodecs-ws` and confirm normal WebRTC login, A/V, control/data channel, refresh/reconnect and applicable view-only/iOS behavior. Browser developer tools must show no `media/capabilities/request`, `media/create`, prototype worker or `/api/media/ws` attempt.

## 4. Opt-in configuration and activation

Add these values to the ignored target-server `.env` using an editor; do not print the file. HTTPS behind the reviewed proxy is the normal target configuration:

```dotenv
NEKO_MEDIA_WEBCODECS_WS_ALLOWED_ORIGINS=https://neko.example
NEKO_MEDIA_WEBCODECS_WS_TRUSTED_PROXIES=172.20.0.1/32
NEKO_MEDIA_WEBCODECS_WS_ALLOW_INSECURE_LOOPBACK=false
NEKO_MEDIA_WEBCODECS_WS_MAX_CONNECTIONS=128
```

Replace both examples with the exact reviewed values. The deployment overlay intentionally takes one exact origin and one immediate proxy IP/CIDR; a multi-origin or multi-proxy deployment needs a separately reviewed config-file form instead of an improvised environment list. Cleartext is permitted only for a direct loopback-only diagnostic and requires the explicit `true` setting; never use it for a remote browser.

Validate without printing the interpolated configuration, then activate all three files:

```bash
set -Eeuo pipefail
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml -f docker-compose.webcodecs-ws.yaml config --quiet
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml -f docker-compose.webcodecs-ws.yaml up -d --force-recreate neko
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml -f docker-compose.webcodecs-ws.yaml ps
./deploy/check-media-websocket-http.sh enabled
NEKO_PUBLIC_BASE_URL='http://127.0.0.1:8082' ./deploy/check-media-websocket-http.sh insecure-denied
```

Adjust `8082` only if the loopback HTTP binding was changed. The expected fixed-invalid probe statuses are 400/403/426/401 as printed by the helper, and direct cleartext must remain 403. A difference can indicate proxy rewriting, a wrong trusted-peer CIDR, an incorrect exact Origin or an unexpected route; diagnose it before using a real ticket.

## 5. Evidence initialization and functional matrix

```bash
set -Eeuo pipefail
RESULT_DIR="../neko-webcodecs-results-$(date -u +%Y%m%dT%H%M%SZ)"
./deploy/collect-webcodecs-media.sh init "$RESULT_DIR"
./deploy/collect-webcodecs-media.sh snapshot "$RESULT_DIR" optin-idle
printf 'result_dir=%s\n' "$RESULT_DIR"
```

Retain that path for later blocks. Use one normal member, one admin and one view-only `CanWatch` session in turn with the exact query `?media=webcodecs-ws`. Confirm VP8 video plus 48-kHz stereo Opus, central Play/mute/volume/fullscreen, the receive-only limitation and denied view-only input. Test fullscreen enter and exit on desktop and the available target phone: native element fullscreen is preferred, while a phone that cannot natively fullscreen the WebCodecs canvas must enter the viewport-filling app mode and expose its compress/exit button; browser chrome may remain on iPhone. Exercise private-mode pause/resume, explicit **Use WebRTC**, unsupported-browser/codec behavior, logout, kick, authorization revocation and session replacement exactly as specified in [`WEBCODECS_MEDIA_WEBSOCKET.md`](WEBCODECS_MEDIA_WEBSOCKET.md). Output must stop within two seconds where required and terminal policy/auth failures must not retry.

After any client recovery correction, keep one iPhone WebCodecs client visible in the foreground for at least ten continuous minutes with no role/window switching. Short local audio rebuffering may be audible, but it must not clear the canvas, emit a server `client_audio_underflow`, reach `resync_limit` or start a media-only retry. Record before/after metrics and the operator's visible-picture/audio observation; a desktop-only interval cannot close this regression.

Record results in the private copy of [`WEBCODECS_MEDIA_WEBSOCKET_RESULTS_TEMPLATE.md`](WEBCODECS_MEDIA_WEBSOCKET_RESULTS_TEMPLATE.md), then take a labeled snapshot and filtered logs.

## 6. Startup, latency and synchronization

Perform ten clean joins on the same supported desktop Chromium/Brave target. Measure from media-socket open to first rendered frame and audible synchronized A/V. Measure glass-to-glass latency and A/V skew with the same reproducible external method used for a same-device WebRTC baseline. Browser UI impressions alone are not sufficient for the numeric gates.

Acceptance is: video p95 <= 2 seconds; audible A/V p95 <= 3 seconds; glass-to-glass p95 <= 500 ms and <= 250 ms worse than WebRTC; absolute A/V skew p95 <= 80 ms with no > 200 ms excursion lasting one second. Preserve the ten raw observations, method and browser/device versions.

## 7. Slow-client isolation and adaptive down/up

Run healthy WebRTC viewers H1/H2 and opt-in viewer W on the same source. Snapshot before impairment, constrain or stall only W with a deployment-appropriate shaper, snapshot during the constraint, remove it and snapshot after recovery. The exact shaping commands are selected only after the interface/client topology is known; a host-wide rule is not acceptable evidence.

```bash
./deploy/collect-webcodecs-media.sh snapshot "$RESULT_DIR" isolation-start
# Apply the reviewed W-only constraint and observe the bounded failure/recovery.
./deploy/collect-webcodecs-media.sh snapshot "$RESULT_DIR" isolation-constrained
# Remove the constraint and wait for successful recovery.
./deploy/collect-webcodecs-media.sh snapshot "$RESULT_DIR" isolation-restored
./deploy/collect-webcodecs-media.sh logs "$RESULT_DIR" 60m
```

Only W may resynchronize/close. H1/H2 must remain usable and must not change quality because of W. Provider/egress caps must hold, reconnect must resume current-generation video from a keyframe within two seconds, and attempts must stay serialized at 1/2/5/10 seconds with fresh tickets. Repeat the already accepted adaptive three-viewer induced down/up cycle and prove H1/H2 remain `high`.

## 8. Five-minute resource and cleanup comparison

Collect a five-minute stable WebRTC baseline and the same workload with one opt-in viewer. Use metrics over the full windows plus repeated connect/disconnect cycles; a single `docker stats` sample alone cannot establish average CPU.

```bash
./deploy/collect-webcodecs-media.sh snapshot "$RESULT_DIR" resource-webrtc-start
# Hold the documented WebRTC baseline for five minutes.
./deploy/collect-webcodecs-media.sh snapshot "$RESULT_DIR" resource-webrtc-end
./deploy/collect-webcodecs-media.sh snapshot "$RESULT_DIR" resource-webcodecs-start
# Hold the identical workload plus one opt-in viewer for five minutes.
./deploy/collect-webcodecs-media.sh snapshot "$RESULT_DIR" resource-webcodecs-end
# Disconnect the opt-in viewer, wait 60 seconds, then:
./deploy/collect-webcodecs-media.sh snapshot "$RESULT_DIR" resource-cleanup-60s
```

Apply the exact CPU/RSS/payload/pipeline/goroutine/delivery limits in the contract and the PromQL in [`WEBCODECS_MEDIA_WEBSOCKET_OBSERVABILITY.md`](WEBCODECS_MEDIA_WEBSOCKET_OBSERVABILITY.md).

## 9. Hostile-input evidence and rollback

The target-server test block covers strict record/control parser cases, ticket lifecycle, queue/rate bounds and private close mappings. The live HTTP helper covers wrong/missing Origin, query rejection, wrong/missing subprotocol and malformed/unknown ticket behavior through the actual reverse proxy. Also verify that successful/replayed/expired ticket cases, binary/control-rate abuse and disconnect cleanup affect only the offending browser and create no credential/payload log entry or server restart.

After evidence capture, roll back without state migration:

```bash
set -Eeuo pipefail
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml up -d --force-recreate neko
docker compose -f docker-compose.yaml -f docker-compose.adaptive.yaml ps
./deploy/check-media-websocket-http.sh disabled
```

Remove the prototype query from clients and confirm ordinary WebRTC A/V and control once more. The unused WebCodecs values may remain in ignored `.env`, but omitting the overlay is the authoritative disablement. Acceptance may be recorded only after every applicable matrix item has evidence; otherwise record the checkpoint as incomplete or rejected.
