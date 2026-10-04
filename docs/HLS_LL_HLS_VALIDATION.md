# HLS / LL-HLS Phase 4 target-server validation

Repository assets prepared on 2026-10-04. **The automated/image preparation and
activation/invalid-input gates passed at exact application `93f1fa63`;
the first playback attempt FAILED and normal login then timed out.
Read-only diagnosis and rollback recovery passed; exact repair-image preparation
and grouped acceptance remain PENDING target verification.**
The supplied output records 47 client tests (including the three new chat
security/formatting regressions), type/build, 13 Go packages, both fuzz jobs
and server/base/Brave builds, with final exit code 0; that preparation left the
running service unchanged.
The two default-off public HTTP 404 checks passed earlier at e55. Caddy 2.6.2
validated the complete merged file, retained all 11 hosts and reloaded through
helper `2484a022`. The synthetic runtime-error redaction, healthy opt-in HLS
deployment, 17 public probes and two cleartext-denial probes all passed with
Activate-Exitcode 0. The operator subsequently reported inability to connect or
no picture with an HLS failure, then normal-login timeouts (tentative /ws 101).
Helper d191b8ea supplied a healthy-image diagnostic but no demonstrated HLS
readiness/lease. [Startup repairs](HLS_STARTUP_REPAIR_2026-10-04.md) are implemented
and await a new exact target gate; the normal-login blocker remains unconfirmed.
The subsequent rollback returned Restore-Exitcode 0 and a healthy prior image;
the operator confirmed normal login, picture and audio work again. Actual HLS
packager/picture/audio and valid authorization have not passed. The
[audit classification](DEPENDENCY_AUDIT_2026-10-04.md) and remaining dependency
work are not a passing security audit. Nothing here was executed
in Codex. Use one block at a time, review its output,
then advance. Keep one exact `testing` implementation commit/image throughout
the matrix. No `master` promotion or full acceptance follows from a smoke check.

Read [the fixed contract](HLS_LL_HLS.md), [Caddy review](HLS_LL_HLS_CADDY.md),
[observability](HLS_LL_HLS_OBSERVABILITY.md) and
[static findings](STABILITY_REVIEW_2026-10-04.md). The established server path is
`/opt/docker/nekoNew/neko`; keep the existing adaptive and WebCodecs overlays.

## 1. Prepare exact tests and images without replacing the service

For this incident, fetch `origin/testing` and fast-forward to the exact reviewed
repair application commit `80020d99477a58318f210b7e14d19cdd92991a6d` below.
Later documentation-only commits need not move this application checkpoint.
Verify the full hash, branch and clean worktree. The result directory
below is private and outside Git; reuse it for all blocks of that exact commit.
When changing the implementation commit, preserve the old directory and use
`../neko-hls-results-<first-12-commit-characters>` for the new checkpoint.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_CHECK'
cd /opt/docker/nekoNew/neko
test "$(git branch --show-current)" = testing
test -z "$(git status --porcelain=v1)"
git fetch origin testing
repair_commit="80020d99477a58318f210b7e14d19cdd92991a6d"
git merge --ff-only "$repair_commit"
test "$(git rev-parse HEAD)" = "$repair_commit"
bash deploy/validate-hls-phase4.sh "../neko-hls-results-${repair_commit:0:12}"
NEKO_HLS_CHECK
printf 'Check-Exitcode: %s\n' "$?"
```

The helper checks Compose quietly, shell syntax and the HTTP checker, runs
client tests/type/build and the relevant Go suite with 30-second WebSocket and
HLS request-boundary fuzz jobs. It builds the same-commit codec-validation image,
reports its GStreamer version and runs required real-codec segment/packager
integration tests before building exact uniquely tagged base/Brave images.
It never stops/recreates the running service. A success marker is written
only after the final baseline snapshot. A failed rerun invalidates that marker.
That snapshot describes the running rollback image. Passing preparation does
not establish runtime behavior of the newly built repair image; verify that
image with HLS disabled before re-enabling the backend.

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

**Current checkpoint:** after the initial single-site guard stop and read-only
inspection, activation helper `2484a022` passed (exit 0). All 11 hosts and other
configuration were preserved; Caddy reloaded, synthetic runtime-error redaction
passed, `my-neko/brave:hls-93f1fa637ae3` started healthy, and the 17 public plus
two cleartext-denial probes passed. See [the Caddy record](HLS_LL_HLS_CADDY.md)
for the logging qualifications. The application stayed at `93f1fa63` for the
successful rollback recovery in section 3. NEXT validate the exact repair
commit/image from section 1; do not reuse the old preparation
marker or repeat the completed source-merging Caddy activation.

The statically reviewed [activation helper](../deploy/activate-hls-phase4.sh)
now uses [merge-hls-caddy.py](../deploy/merge-hls-caddy.py) and additionally:

- edits only the explicit bare Neko proxy and a runtime default logger with no
  existing encoder, rejecting unfamiliar source/proxy/logging structures;
- requires complete adapted JSON equality outside the approved proxy/default
  encoder changes, preserves the other hosts/options/loggers, validates with
  Caddy and rechecks the original before a private backup/reload;
- retains relative import resolution by using a private temporary candidate
  beside the active file, with cleanup on preparation success/failure;
- captures the old Neko image before stopping it;
- uses the deployment stop for one synthetic public 502 and checks the actual
  Caddy journal for normalized URI, absent headers and absent synthetic marker;
- starts the already-tested image, performs the enabled/cleartext denial checks,
  and attempts image/config restoration if any activation gate fails.

It starts only conventional HLS. This intentionally interrupts existing
sessions during deployment; the log check uses no valid lease or credential.
The raw journal remains private. No access logger is added to the supplied
site, and this one error case does not cover arbitrary custom/debug logging.
Valid playback and the grouped matrix still follow separately.

**Historical successful activation block:** application/image remained at
`93f1fa63`. Both `deploy/activate-hls-phase4.sh` and `deploy/merge-hls-caddy.py`
were extracted from helper commit `2484a022` into the private evidence directory.
The helper recorded both tooling blobs and the application commit, passed the
merge/activation/synthetic-error/invalid-input gates, and left valid playback
pending. This was target execution, **NOT EXECUTED IN CODEX**. Do not repeat its
bare-source merge on the already-modified Caddyfile. After incident recovery and
a new exact repair-image gate, use the plain enable/probe block below if Caddy
and its logging configuration remain as reviewed:

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_ENABLE'
cd /opt/docker/nekoNew/neko
commit="$(git rev-parse HEAD)"
bash deploy/deploy-hls-media.sh enable "../neko-hls-results-${commit:0:12}"
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

### Failed playback/login incident: recovery passed, repair preparation NEXT

On 2026-10-04 the operator reported `hls failed`, connection failure/no picture,
then normal-login timeouts. Device/browser and exact player status remain
unconfirmed; `/ws` HTTP 101 was tentatively reported. If confirmed, upgrade
worked but does not prove authentication or session initialization.

The read-only helper from d191b8ea passed with exit 0 at application 93f1fa63:
prepared image healthy/running, zero Docker restarts/OOM, no Neko process exit
in the bounded Supervisor sample, and responding metrics. HLS startup/source-
restart and bootstrap `not_ready`/`backend_error` were recorded, without any
successful readiness/lease-open demonstration. Raw files remain private in
`/opt/docker/nekoNew/neko-hls-results-93f1fa637ae3`. Counts are cumulative/bounded
samples, not correlated browser evidence. No credentialed playback request was
made by the helper. Its safe output was supplied from target; **NOT EXECUTED IN
CODEX**. See [the detailed repair/evidence record](HLS_STARTUP_REPAIR_2026-10-04.md).

The following recovery block was executed successfully from the unchanged
tested checkout, restoring the recorded prior image without the HLS overlay:

```bash
set +e
bash -e -o pipefail <<'NEKO_RESTORE'
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "93f1fa637ae3f14ba41d1bfef39e860d43993ff1"
output=/opt/docker/nekoNew/neko-hls-results-93f1fa637ae3
test "$(stat -c %a "$output")" = 700
bash deploy/deploy-hls-media.sh rollback "$output"
NEKO_RESTORE
printf 'Restore-Exitcode: %s\n' "$?"
```

This recreates Neko, disconnecting active sessions, while retaining the reviewed
Caddy configuration and adaptive/WebCodecs overlays. It restores the recorded
prior runtime, not the latest containment repairs; this is an incident baseline.
The supplied output returned **Restore-Exitcode 0** and healthy configured image
`my-neko/brave:rollback-hls-93f1fa637ae3`. The operator confirmed normal login,
picture and audio work again. Preserve all old evidence/images. The rollback
changed image, overlay and process state together; it does not establish the
original blocker or validate the new HLS repair. No wider matrix was reported.

The repository now includes cold-generation/initial-caps, HLS-only encoder-
segment timestamp, initial videorate gap and bounded C logging repairs plus a
mandatory real-codec validation job. Those tests/builds/images are **PENDING on
target**. Restoration is reviewed; NEXT advance to exact repair commit 80020d99
and prepare its fresh images/result directory using section 1 without replacing
the currently working service. Verify normal
WebRTC with HLS disabled first. The target Caddy source already contains the
reviewed changes; use the plain enable/probe block in section 2 after these gates,
not the completed bare-source merger. Then repeat the bounded picture/audio
checkpoint below and the full valid-delivery matrix. If login still hangs,
collect targeted private event/login diagnostics before further HLS work; its
cause is not established by the static startup findings.

### First bounded picture/audio checkpoint (repeat after diagnosis/repair)

Keep one ordinary WebRTC viewer connected, with changing video/audio in the
shared browser. In a separate private browser window, open the deployment root
with exactly `?media=hls` and log in as an admin using a distinct test name. For
the confirmed root deployment this is `https://neko.taxzvps.de/?media=hls`.
The query is a stateless diagnostic override and does not change the saved
browser preference. Ordinary non-admin members are not eligible for HLS; this
first admin diagnostic is not the view-only authorization acceptance test.

Press Play and enable audio if the browser requires a gesture. Record device,
OS/browser version, approximate login-to-first-moving-picture duration, audible
audio/A-V impression, and the exact fixed status/error if playback fails. Run
five foreground minutes while another WebRTC participant sends a message,
releases/takes control and a fresh participant joins. Record HLS interruption
or terminal state separately from normal buffered display delay, and confirm
the WebRTC viewer remains working. Do not infer the fixed ten-start/ten-minute
numeric acceptance gates from this initial bounded checkpoint. The reported
first playback attempt failed; the full five-minute checkpoint has not passed.

While HLS is active, the existing collector can preserve private metrics:

```bash
cd /opt/docker/nekoNew/neko
bash deploy/collect-hls-media.sh snapshot ../neko-hls-results-93f1fa637ae3 hls-first-playback
```

### Remaining valid-delivery matrix

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
commit="$(git rev-parse HEAD)"
bash deploy/deploy-hls-media.sh rollback "../neko-hls-results-${commit:0:12}"
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
