# HLS / LL-HLS Phase 4 target-server validation

Repository assets prepared on 2026-10-04. **Target validation is PENDING.**
Nothing here was executed in Codex. Use one block at a time, review its output,
then advance. Keep one exact `testing` implementation commit/image throughout
the matrix. No `master` promotion or full acceptance follows from a smoke check.

Read [the fixed contract](HLS_LL_HLS.md), [Caddy review](HLS_LL_HLS_CADDY.md),
[observability](HLS_LL_HLS_OBSERVABILITY.md) and
[static findings](STABILITY_REVIEW_2026-10-04.md). The established server path is
`/opt/docker/nekoNew/neko`; keep the existing adaptive and WebCodecs overlays.

## 1. Prepare exact tests and images without replacing the service

Pull the reviewed `origin/testing` commit using `git pull --ff-only`. Verify the
expected full hash, branch and clean tracked worktree. The result directory
below is private and outside Git; reuse it for all remaining blocks.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_CHECK'
cd /opt/docker/nekoNew/neko
test "$(git branch --show-current)" = testing
test -z "$(git status --porcelain=v1)"
caddy version
systemctl is-active caddy
bash deploy/validate-hls-phase4.sh ../neko-hls-results
NEKO_HLS_CHECK
printf 'Check-Exitcode: %s\n' "$?"
```

The helper checks Compose quietly, shell syntax and the HTTP checker, runs
client tests/type/build and the relevant Go suite with 30-second WebSocket and
HLS request-boundary fuzz jobs, then builds exact uniquely tagged base/Brave
images. It never stops/recreates the running service. A success marker is written
only after the final baseline snapshot. A failed rerun invalidates that marker.

`dependency-audit.json` and its exit code are private. Audit findings can return
nonzero without stopping image preparation; they are **not a passing security
gate**. Review advisory/package/version, production versus build reachability
and applicable repairs before final acceptance. An unavailable registry or
invalid report leaves this review pending. Do not run `npm audit fix --force`.
If a repair changes the commit, repeat the exact automated/image gate.

Check default-off public routes with the credential-free probe:

```bash
cd /opt/docker/nekoNew/neko
docker compose -f docker-compose.validation.yaml run --rm -T hls-http-checks disabled
```

Its base URL defaults to the reviewed HLS/WebCodecs HTTPS origin. For a prefix,
pass the actual public base URL through `-e NEKO_PUBLIC_BASE_URL=...`; it contains
no credentials. Confirm existing WebRTC/WebCodecs playback, control/recovery,
fullscreen and absent HLS choices before enabling. Complete the Caddy routing,
trust, streaming and log review before creating valid HLS credentials.

## 2. Explicit conventional HLS deployment and invalid-input checks

The separate `docker-compose.hls.yaml` requires adaptive source geometry and
exact HTTPS/proxy values; empty HLS values reuse the reviewed WebCodecs values.
Its initial mode is only `hls`. Base Compose remains default-off.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_ENABLE'
cd /opt/docker/nekoNew/neko
bash deploy/deploy-hls-media.sh enable ../neko-hls-results
docker compose -f docker-compose.validation.yaml run --rm -T hls-http-checks enabled
docker compose -f docker-compose.validation.yaml run --rm -T \
  -e NEKO_PUBLIC_BASE_URL=http://127.0.0.1:8082 hls-http-checks insecure-denied
NEKO_HLS_ENABLE
printf 'Enable-Exitcode: %s\n' "$?"
```

Adjust the direct port/prefix to the actual bind when needed. The deploy helper
captures the running image under a rollback tag **before** stopping the service,
preserves adaptive/WebCodecs, enables HLS using the tested image and waits for
health. Startup failure attempts an immediate restore without the HLS overlay.
HTTP probe failure does not itself roll back; review/rollback explicitly.

The HTTP probe uses synthetic unknown tickets/cookies, disables redirects/proxy
inheritance and prints only fixed case/status/header verdicts. It proves no valid
bootstrap, packager, range, rate, replay or revocation behavior. The direct
cleartext check sends no forwarding header; spoofed TLS metadata must separately
be tested from a peer **outside** the configured trusted proxy set, or marked
unexecuted. A Docker bridge peer inside a trusted CIDR is not that test.

## 3. Valid delivery, authorization and lifecycle

With an ordinary WebRTC member, an admin and an explicit view-only HLS viewer
in the same room, check changing picture/audio and shared presence. HLS is
receive-only even for admin diagnostics. Repeat the complete server-enforced
denial matrix from [VIEW_ONLY_SHARING.md](VIEW_ONLY_SHARING.md); the receive
backend must not permit input, control, chat, file, plugin or inbound media.

Using controlled browser/local tools without exporting raw credentials, test
valid POST bootstrap, one-use replay, ten-second ticket expiry, cookie flags
and narrow root/prefixed paths, 15-second renewal and 30-second idle expiry.
Keep private mode enabled for at least **45 seconds**: video buffers clear,
media requests deny with `503`, keepalive still renews with `204`, unused
packagers stop, and resume uses the same live lease after fresh readiness.
Exercise revoked sharing, logout, kick, `CanWatch` loss, delivery replacement,
service recreation and shutdown. New bytes stop within one second and the
cooperative display clears within two seconds; record the already-fetched-data
caveat. Admin private-mode behavior remains independent of passive viewers.

Inspect real generated master/children/init/fragments with an independent fMP4
inspector and, where available, Apple validators. Check codecs/timestamps,
aligned IDRs, parent rollover, discontinuities and no old-generation mixing.
Test HEAD/gzip/ranges and invalid/duplicate/unknown/future queries with an
authenticated lease; confirm bounded limits and cancellation on downstream
abort. The new fuzz job covers request parsers, not a complete mux/player fuzz
campaign. Record missing validator/mux-fuzz cases as unexecuted.

Repeat the reported event sequence (join, message, take/release/grant control)
on available viewers. On the affected TV when available, compare chat sound
on/off first, then each supported backend with identical content. The actual
colleague TV is currently **DEFERRED**, not passed. Preserve explicit WebCodecs.

## 4. Devices, latency and LL-HLS

Record exact model/OS/browser, H.264/AAC capability, native/MSE path, Play
gesture, fullscreen/PiP, refresh and foreground/background return per device.
Run ten cold starts and ten minutes per tested mode. Collect at least 30
synchronized visible-clock samples: conventional p95 startup and displayed
latency at most 24 seconds; LL-HLS at most six seconds. Both require the fixed
A/V drift and stall gates. ICE connection and HTTP timing are separate from
first moving picture/audio and displayed latency.

Only after actual public HTTP/2 or HTTP/3 and path p95 RTT <=333 ms are recorded,
advertise both modes by applying `NEKO_MEDIA_HLS_MODES='hls ll-hls'` to the
**same** enable helper. Viper environment slices use whitespace; `hls,ll-hls`
is invalid. Keep that choice in the server's ignored `.env` for subsequent
recreates, and record it. Select `Low-Latency HLS` explicitly, check blocking
reload cancellation and repeated six-part parent rollover; no automatic fallback.

## 5. Mixed-backend isolation/resources and rollback

Follow the five-minute phases and fixed thresholds in
[HLS_LL_HLS_OBSERVABILITY.md](HLS_LL_HLS_OBSERVABILITY.md). Include healthy
WebRTC/WebCodecs viewers, an induced slow reader, constrained adaptive down/up
and source restart/resolution change. Record packager count, CPU/RSS, retention,
drop/close counters, A/V and control. Never infer healthy-viewer isolation from
an unshaped single-viewer stream. Verify zero persistent media-object writes and
idle cleanup separately from ordinary browser-profile writes.

Return clients explicitly to WebRTC, then:

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_ROLLBACK'
cd /opt/docker/nekoNew/neko
bash deploy/deploy-hls-media.sh rollback ../neko-hls-results
docker compose -f docker-compose.validation.yaml run --rm -T hls-http-checks disabled
NEKO_HLS_ROLLBACK
printf 'Rollback-Exitcode: %s\n' "$?"
```

Rollback reads the saved exact prior image, omits HLS and retains adaptive plus
WebCodecs. Confirm health, zero packagers and existing playback/control.
Do not delete the saved image/result directory during the checkpoint.

Complete the private `RESULTS.md`. Commit only a sanitized evidence summary
with exact hash/image and explicit omissions. Missing device, numeric, hostile
input or dependency gates prevent a full prototype-acceptance claim. `master`
remains pinned until a later explicit operator promotion decision.
