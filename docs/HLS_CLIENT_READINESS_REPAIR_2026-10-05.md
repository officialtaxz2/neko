# HLS client startup deadline repair — 2026-10-05

Status: client-only repair `73d5ff6d29110e3dd06999726a7e88718d09ea23` is
implemented and statically reviewed on `testing`. Supplied target results
passed the old-defect reproduction, all 52 client tests, type/build and the
read-only live diagnosis, with Client-Check-Exitcode 0. Scoped exact-73 image
preparation then passed with Client-Image-Exitcode 0. Same-image default-off
deployment passed with Baseline-Exitcode 0, healthy service and 2/2 disabled-route
probes; the operator confirmed the requested normal browser check works.
Same-image HLS activation then passed with Enable-Exitcode 0, healthy service
and 19/19 denial probes. The operator reports HLS playback on PC/Helium after
an initial Retry and one frozen-picture/page-reload incident; WebRTC kept working.
"HLS failed" was shown, but its detailed error, exact timing and player path
are not supplied. The subsequent read-only diagnosis passed with a healthy
exact image, no sampled process exit/OOM or fixed error markers, one active
HLS lease, all four workers running, six bootstrap successes and 394 successful
segment requests. One not-ready bootstrap is recorded, plausibly explaining
Retry without correlation to a specific browser attempt. Startup reliability
and sustained/grouped acceptance remain
**PENDING**. All execution was on the target server, **NOT EXECUTED IN CODEX**.
The checkout/live image remains exact 73d5ff6d, now with conventional HLS enabled.

## Supplied live evidence

Exact-a7ffb8b1 same-image activation passed with Enable-Exitcode 0, a healthy
`my-neko/brave:hls-a7ffb8b13448` container and the private enable snapshot.
All 17 public invalid-input probes and both cleartext-denial probes passed.
The deployer blob remains `c6f52dc80fdf605ec908f3fe3856ce23e015e494`.

The operator then reported first successful HLS picture, the compact `HLS`
indicator while playback worked, and after approximately 30 seconds:

> HLS playback did not become ready; retry manually

Retry HLS restored playback. The operator explicitly confirmed WebRTC keeps
working during this HLS failure. The compact indicator is rendered only for
the client's `streaming` state in `client/src/components/video.vue`; it supports
an earlier observed `playing` event, rather than a player that never started.
Audio, browser/device version, exact elapsed time and sustained playback were
not separately measured. This is first-picture progress, not acceptance of
the five-minute checkpoint, role/device matrix or all-stream incident repair.
The subsequent safe live diagnosis and isolated client output were supplied
and passed as recorded below.

## Static finding and bounded correction

`client/src/neko/hls/controller.ts` armed a 30-second readiness deadline after
player attachment. Its `canplay` and `playing` handlers never cancelled that
timer. When the deadline ran, a then-current `readyState < 3` terminated the
player with the exact reported message, even after successful streaming.
A brief buffer underrun at that instant could therefore be misclassified as
startup failure. This defect is present by source inspection and is consistent
with the reported message/timing; the browser's actual buffer state at failure
has not been captured, and separate packaging/network issues are not ruled out.

The correction cancels the initial deadline on either current-player readiness
event and arms it before native/MSE attachment, so an immediate event cannot
leave a newly assigned timer alive. Player cleanup uses the same idempotent
timer removal. Existing generation guards prevent removed-player callbacks
from cancelling a new player's deadline. A player that never becomes ready
still has its 30-second startup bound; sustained lack of playback progress
still has the separate 20-second watchdog, evaluated by serialized lease
polling. Manual pause and blocked-autoplay Play fallback remain intact.

The repair changes no server, codec, package/lockfile, HLS lease/security limit,
buffer size or transport choice. Native and MSE controller doubles cover
earlier readiness followed by buffering around the old deadline, never-ready
cleanup, retained stall detection, immediate native readiness and stale
callbacks after private resume. The existing blocked-autoplay case now waits
beyond the old deadline before exercising manual Play. These are synthetic
target tests, not native/MSE browser or device compatibility evidence.

## Completed target block: read-only diagnosis and isolated client A/B

Use this block while the checkout and live image remain at a7ffb8b1. It first
saves a credential-safe live summary, then exports the exact repair helper
outside Git. The helper verifies the two client files are the only application
changes; server/runtime/configuration and dependencies must match the baseline.
It copies the client/fixtures into an ephemeral validation container, requires
the old controller to fail the specific new regression assertion, then replaces
only the controller in that copy and runs the complete client tests/type/build.
It installs the existing locked dependencies only in that container.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_CLIENT_CHECK'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "a7ffb8b13448a8329c6df24fdcb182ac32ca398c"

umask 077
output=/opt/docker/nekoNew/neko-hls-results-a7ffb8b13448
test "$(stat -c %a "$output")" = 700

diagnostic_status=0
bash deploy/diagnose-hls-playback.sh "$PWD" "$output" || diagnostic_status=$?
printf 'Diagnostic-Exitcode: %s\n' "$diagnostic_status"

git fetch origin testing
repair_commit="73d5ff6d29110e3dd06999726a7e88718d09ea23"
git show "$repair_commit:deploy/validate-hls-client-readiness.sh" \
  > "$output/validate-hls-client-readiness.sh"
bash -n "$output/validate-hls-client-readiness.sh"
bash "$output/validate-hls-client-readiness.sh" "$PWD" "$output" "$repair_commit"
NEKO_HLS_CLIENT_CHECK
printf 'Client-Check-Exitcode: %s\n' "$?"
```

No checkout move, image build/replacement, Caddy edit or live credentialed
playback attempt is performed. Diagnosis may contain cumulative evidence from
multiple retries; it cannot identify one browser buffer state. Raw live logs
remain private. Share only the diagnostic summary and synthetic client output.
If diagnosis fails, its separate exit code is printed and independent client
checks still run. If old-fault reproduction or any corrected check fails, the
client gate stops without a success marker.

Evidence is saved in a new private `client-readiness-<UTC>` directory below
the existing a7ff output. Its `client-check-commit.txt` becomes the repair hash
only after complete client success. It does not replace `validation-commit.txt`
or qualify a new application image for deployment. Review output before
preparing a new image and repeating normal-browser plus enabled HLS playback.
The running service still serves the old client until that later deployment.

## Supplied diagnosis and exact-client gate passed

The supplied diagnostic returned Diagnostic-Exitcode 0 at application a7ffb8b1.
Its running image ID is
`sha256:e3517c887622e04065a7fec5fae1902c03e2e4e5470946915b9b99457e2992b2`;
the container was healthy with zero restarts, OOM false and no Neko process exit
in the bounded supervisor sample. The diagnostic captured 977 application log
lines with no sampled fixed codec/timeline/crash markers. The GStreamer CLI
version was unavailable; this does not mean the running media library is absent.

Two bootstrap successes, two lease-open/close markers, one packager-ready
generation and one idle-grace stop were recorded. All four tracks have startup
generation 1 and cumulative part/parent publication, including medium/low.
Requests include 486 successful segments, 479 playlists, 1,408 masters, 92
keepalives and seven init objects. At capture all leases/packagers/retained
objects were zero, consistent with the logged stop after leases closed. One
earlier not-ready bootstrap and six negotiation rejections are also present;
these cumulative counts do not identify one attempt or the browser buffer state.
Missing error markers are not proof that every earlier event was error-free.

The isolated helper then reproduced the old readiness-deadline assertion as
required. All 52 repaired client tests passed (zero failures/skips), including
the five new native/MSE/readiness/lifecycle cases, and TypeScript `tsc --noEmit`
plus the production Vite build passed. Final Client-Check-Exitcode was 0 with
CLIENT READINESS A/B/TYPE/BUILD GATE PASSED. Evidence is private at:

`/opt/docker/nekoNew/neko-hls-results-a7ffb8b13448/client-readiness-20261005T171735141870202Z`.

Checkout, images and running service remained at a7ffb8b1. This establishes the
controlled timer correction; it is not a repaired-image browser check. npm
reported 20 existing advisories (11 low, three moderate, five high, one critical)
and Vite reported large chunks; neither prevented this gate. No dependency
update or passing security-audit claim follows. **NOT EXECUTED IN CODEX;
supplied read-only diagnosis and exact client A/B/tests/type/build passed.**

## Completed target block: client-only image preparation

The helper is pinned independently at tooling commit
`4a957f3e1f632896b3da3d4b816f5d4008bdbad4`. Application source remains pinned to
the tested `73d5ff6d29110e3dd06999726a7e88718d09ea23` repair, not newer tooling/
documentation commits. All guards run before the clean checkout advances from
a7ffb8b1 to that repair. The running enabled a7ff image is retained during builds.

The helper checks the successful client marker, exact helper/source blobs and
baseline preparation/image record. It rejects changes outside the two client
files, their readiness helper and documentation. Unchanged server/capture/
runtime/codec/configuration/dependency sources inherit the successful a7ff
backend gate; no repeat Go/fuzz/codec test run or fresh exact-73 backend-test
claim is made. This scoped provenance is saved in `preparation-scope.txt` and
`inherited-backend-commit.txt` beside the client evidence reference. Fresh base/
Brave artifacts are built from the repair with `CLIENT_DIST` explicitly empty.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_CLIENT_IMAGE'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko

umask 077
base_output=/opt/docker/nekoNew/neko-hls-results-a7ffb8b13448
client_report="$base_output/client-readiness-20261005T171735141870202Z"
output=/opt/docker/nekoNew/neko-hls-results-73d5ff6d2911

git fetch origin testing
helper_commit="4a957f3e1f632896b3da3d4b816f5d4008bdbad4"
git show "$helper_commit:deploy/prepare-hls-client-repair.sh" \
  > "$base_output/prepare-hls-client-repair.sh"
bash -n "$base_output/prepare-hls-client-repair.sh"
bash "$base_output/prepare-hls-client-repair.sh" \
  "$PWD" "$base_output" "$client_report" "$output" "$helper_commit"
NEKO_HLS_CLIENT_IMAGE
printf 'Client-Image-Exitcode: %s\n' "$?"
```

The new output directory must be absent; preserve any failed attempt. Its
`validation-commit.txt` stays PENDING until both images, exact image-ID records
and the private snapshot succeed, with the original live container/image still
unchanged. Final success qualifies the scoped repair-image preparation, not
the full enabled acceptance matrix. There is no service recreation, Caddy edit,
test repetition or HLS enablement change in this block. This helper has only
been statically reviewed in Codex; its supplied target syntax/build/preparation
subsequently passed as recorded below.

After reviewing success, deploy the new image default-off using the established
helper/new evidence path, confirm normal login/picture/audio/control, then enable
conventional HLS in that same image and repeat sustained picture/audio alongside
WebRTC. A private/new browser window must load the new client bundle. The later
deployer saves the working a7ff image under the repair's rollback tag. Keep both
old evidence directories and images; no pruning or `master` promotion.

WebRTC continues working per the operator. If normal playback also fails before
new deployment, preserve diagnosis and restore the confirmed a7ff default-off
baseline, keeping the application checkout and matching evidence aligned.
Earlier all-stream outage causes, TV compatibility, sustained A/V, authorization/
lifecycle, resource/isolation, dependency remediation and grouped acceptance
remain open. There is no automatic fallback or full acceptance claim.

## Scoped exact-73 image preparation passed

The complete supplied output identifies helper
`4a957f3e1f632896b3da3d4b816f5d4008bdbad4`, application
`73d5ff6d29110e3dd06999726a7e88718d09ea23`, inherited backend a7ffb8b1 and
final CLIENT REPAIR IMAGE GATE PASSED / Client-Image-Exitcode 0. The clean
checkout advanced from a7ff to 73; the existing enabled a7ff container/image
stayed unchanged through both builds and the private snapshot.

The client Docker stage rebuilt with Vite 6.4.3, transforming 673 modules and
producing `index-CrHQRMnq.js`, matching the isolated passing client gate. The
server build, plugins and common runtime layers were cached; no fresh backend
test/codec/fuzz run or new server compilation is claimed. Both
`my-neko/base:hls-73d5ff6d2911` and `my-neko/brave:hls-73d5ff6d2911` built.
The Brave stage installed version 1.96.61. The existing large-chunk and manual-
page warnings did not stop the build.

Private evidence, exact prepared image-ID records, the scope/inherited-source
records and successful preparation marker are under
`/opt/docker/nekoNew/neko-hls-results-73d5ff6d2911`. The displayed Docker build
manifest/config digests are not a supplied runtime container image-ID check.
The deployer will compare the actual prepared image against `images.txt` before
stopping the current service. **NOT EXECUTED IN CODEX; supplied scoped image
preparation passed; default-off deployment/browser subsequently passed below,
enabled acceptance pending.**

## Completed target block: exact-73 default-off deployment and browser check

Use the existing deployer without fetching/merging newer documentation/tooling
commits. This block restarts Neko on the prepared repair image with HLS disabled,
retains adaptive/WebCodecs overlays, and saves the running a7ff image as
`my-neko/brave:rollback-hls-73d5ff6d2911` before replacement. A failed startup
attempt restores that saved image without HLS. Existing a7ff/71 evidence and
rollback tags remain available; no Caddy edit, image rebuild or pruning.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_BASELINE'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "73d5ff6d29110e3dd06999726a7e88718d09ea23"

umask 077
output=/opt/docker/nekoNew/neko-hls-results-73d5ff6d2911
test "$(stat -c %a "$output")" = 700

bash deploy/deploy-hls-media.sh baseline "$output"
docker compose -f docker-compose.validation.yaml run --rm -T \
  hls-http-checks disabled </dev/null
NEKO_HLS_BASELINE
printf 'Baseline-Exitcode: %s\n' "$?"
```

Review healthy deployment and the two expected disabled-route 404 probes.
Then use a new private browser window at `https://neko.taxzvps.de/` without a
media override and check normal login, changing picture, audio and control.
This loads the new client bundle. Supply the output and browser result before
same-image conventional-HLS enablement and the sustained picture/audio gate.
That enablement/manual playback matrix remains pending; a scoped image build
does not establish the fix on a real browser or TV.

## Exact-73 default-off deployment and reported browser checkpoint passed

The supplied output identifies application
`73d5ff6d29110e3dd06999726a7e88718d09ea23`, unchanged deployer blob
`c6f52dc80fdf605ec908f3fe3856ce23e015e494` and healthy container image
`my-neko/brave:hls-73d5ff6d2911`. The private baseline snapshot was recorded
under `/opt/docker/nekoNew/neko-hls-results-73d5ff6d2911`. Disabled bootstrap
and media returned their expected 404 (2/2); Baseline-Exitcode was 0. The
loopback binding remains 127.0.0.1:8082 and the established UDP ports remain
published. The existing deployer preserves the prior a7ff image under
`my-neko/brave:rollback-hls-73d5ff6d2911`.

The operator replied "Funktioniert ohne media=hls extra." to the requested
normal browser check. This closes the bounded reported default-off checkpoint;
it does not demonstrate the repaired client's enabled HLS playback or a device/
resource/role matrix. **NOT EXECUTED IN CODEX; supplied default-off deployment,
disabled-route probes and reported normal-browser checkpoint passed.**

## Completed target block: exact-73 same-image HLS enablement and playback attempt

Keep the exact application checkout/image/evidence aligned; use the existing
deployer without pulling later documentation commits, rebuilding or changing
Caddy. The existing rollback tag remains the saved a7ff image. Enable only
conventional HLS with the reviewed overlay, then require healthy service and
all seventeen public plus two cleartext-denial probes before valid playback.
These probes verify request boundaries, not the corrected client timer.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_ENABLE'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "73d5ff6d29110e3dd06999726a7e88718d09ea23"

umask 077
output=/opt/docker/nekoNew/neko-hls-results-73d5ff6d2911
test "$(stat -c %a "$output")" = 700

bash deploy/deploy-hls-media.sh enable "$output"
docker compose -f docker-compose.validation.yaml run --rm -T \
  hls-http-checks enabled </dev/null
docker compose -f docker-compose.validation.yaml run --rm -T \
  -e NEKO_PUBLIC_BASE_URL=http://127.0.0.1:8082 \
  hls-http-checks insecure-denied </dev/null
NEKO_HLS_ENABLE
printf 'Enable-Exitcode: %s\n' "$?"
```

After Enable-Exitcode 0, keep one normal WebRTC viewer connected and open a
new private window at `https://neko.taxzvps.de/?media=hls` as an admin diagnostic.
Use Play/unmute if required. Observe changing picture and audible audio for
five foreground minutes, including another participant's message, control
release/take and a fresh join. Record device/browser, any fixed error and whether
WebRTC keeps working. Do not retry away a failure before recording it. The
previous approximately 30-second startup error must not return after readiness.
This initial bounded interval is separate from the full role/device/resource/
numeric acceptance matrix. Enablement and repaired live playback are pending.

If an HTTP gate fails or playback fails, capture the safe diagnosis before
restoring the confirmed same-image default-off baseline as specified in
[the playback runbook](HLS_LL_HLS_VALIDATION.md#first-bounded-pictureaudio-checkpoint-repeat-after-diagnosisrepair).
Keep all evidence/rollback tags. WebRTC stays the default, fallback manual,
and no `master` promotion or full HLS acceptance is implied.

## Exact-73 activation passed; reported playback has startup failures

The supplied enablement identifies application
`73d5ff6d29110e3dd06999726a7e88718d09ea23`, unchanged deployer blob
`c6f52dc80fdf605ec908f3fe3856ce23e015e494`, healthy
`my-neko/brave:hls-73d5ff6d2911` and a private enable snapshot. All seventeen
public invalid-input/auth-boundary probes and both cleartext-denial probes
passed, with Enable-Exitcode 0. No new image build, Caddy change or source change
was part of this block. These probes do not establish valid playback reliability.

The operator tested on a PC with the Chromium-derived Helium browser and
reports: an initial Retry was needed, a frozen picture required one page reload,
and subsequent playback worked normally while WebRTC continued working.
The follow-up characterizes these as startup difficulties and confirms a
"HLS failed" message. The detailed error, browser version, player path,
elapsed times, audio-specific result and uninterrupted five-minute room-event
interval are not supplied. No old 30-second-deadline recurrence is established
without the detailed message. Later working playback does not close first-start
or recovery reliability; no TV or full prototype acceptance claim follows.

## Static distinction: cold startup, buffered delay and a failed player

The current packager requires three complete six-second parents for conventional
HLS readiness (`packager.go` / `playlist.go`); the controlled codec fixtures
previously became ready around 18 seconds. `ConventionalReadyWindow` remains
24 seconds, and the client bootstrap request has a separate 30-second bound.
This preparation precedes attaching the native/MSE player. An idle packager
stops after 15 seconds, so a lone fresh viewer can encounter another cold start.

The playlist advertises HOLD-BACK=18, and the pinned MSE adapter uses
liveSyncDurationCount=3 with six-second parents. This is an intended buffered
distance from the playlist edge, not an additional compulsory 18-second wait
after every login. Completed media can play immediately when joining an
already warm stream. [RFC 8216 section 6.3.3](https://www.rfc-editor.org/rfc/rfc8216.html#section-6.3.3)
explains the conventional live-start recommendation of three target durations;
it does not certify this application's measured end-to-end delay.

The reported 20–30-second lag is a rough operator estimate, not a measured
latency gate. Warm-up/hold-back can explain delayed first picture and content,
but cannot establish the cause of a terminal "HLS failed" or frozen picture
requiring reload. Candidate stages remain bootstrap readiness, HTTP/playlist
delivery, decoder/player state and progress monitoring. The separate initial
readiness and ongoing stall watchdogs must not be conflated. Do not increase
timeouts, change buffering/GOPs or enable LL-HLS without evidence of the failed
stage. No code change is justified by this report alone.

## Completed target block: exact-73 read-only playback diagnosis

Keep the currently working conventional-HLS/WebRTC service running and the
application checkout/image/evidence aligned. This existing helper collects
bounded private logs, container state and cumulative fixed HLS counters; it
makes no credentialed playback attempt or service change. It cannot reconstruct
browser error details erased by a reload. Run once before any restart or rebuild.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_DIAG'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "73d5ff6d29110e3dd06999726a7e88718d09ea23"

umask 077
output=/opt/docker/nekoNew/neko-hls-results-73d5ff6d2911
test "$(stat -c %a "$output")" = 700
bash deploy/diagnose-hls-playback.sh "$PWD" "$output"
NEKO_HLS_DIAG
printf 'Diagnostic-Exitcode: %s\n' "$?"
```

Share only its fixed safe summary; raw logs, lease paths, WebSocket URLs and
credentials remain private. If a future attempt fails, record its exact fixed
UI detail and whether picture/audio ever started before Retry/reload. Do not
repeat enablement or HTTP probes without a new reason. Diagnosis, reliable
first-start/recovery and sustained/device/authorization/resource acceptance
remain open. **NOT EXECUTED IN CODEX; supplied activation/19 denial probes passed,
later playback reported after recovery, startup fault unresolved.**

## Exact-73 read-only diagnosis passed; first-start reliability remains open

The supplied complete safe summary ended with Diagnostic-Exitcode 0 and pins
application `73d5ff6d29110e3dd06999726a7e88718d09ea23`, helper blob
`ddfe001618f8733447f4bdd328a8519db9acef06` and actual running container image ID
`sha256:cf8913d83d2e6fd2b1e5ca3e2af5dbf984f5432ed65722ee738a20775cf1b0cb`.
This supplies a runtime image-ID observation; prior build-export digests alone
did not establish it. The container was healthy/running, with zero restarts,
OOM false and no sampled Neko exits. In 987 application lines, the fixed error
marker counts were empty.

One HLS lease was active; audio/high/medium/low each had a running worker, an
init object, 42 retained parts and six retained parents. All four tracks had
cumulative part/segment publication. The summary records six successful
bootstraps, 16 successful init requests, 75 keepalives, 763 master requests,
377 playlists and 394 segments. These are cumulative counters spanning all
participants/retries, not one uninterrupted browser interval or proof of audio
playback.

Exactly one bootstrap had result `not_ready`. In the existing server path,
packager readiness that is not reached before the wait context ends returns
ErrPackagerNotReady, which maps to HTTP 503; client bootstrap does not retry
that one-time ticket automatically. This supports readiness as a candidate
for the reported initial Retry. The counter does not identify the browser
attempt, elapsed readiness time, deadline versus cancellation, or the cause
of later frozen-picture recovery. Do not assert a proven 24-second cold-start
timeout or recurrence of the old client deadline from this summary.

Two packager starts/readiness markers were recorded, with every track's
generation result labelled `startup`, plus one idle-grace stop. No worker-
failure/source-restart/timeline-gap marker is shown in the bounded summary;
the two starts are consistent with the recorded idle stop and a later fresh
start, not evidence of a crash loop. Seven negotiation rejections are not
correlated to the user-reported failure. The CLI GStreamer version remains
unavailable; this does not contradict the observed working packager.

The current working HLS/WebRTC service was unchanged. **NOT EXECUTED IN CODEX;
supplied read-only diagnosis passed, active delivery demonstrated; reliable
first-start/recovery, exact browser failure cause and grouped acceptance open.**

## Next: consolidate startup, frozen-picture and room-event investigation

The requested follow-up static work is now recorded in
[the client stability review](HLS_CLIENT_STABILITY_REVIEW_2026-10-05.md).
It corrects paused-time stall accounting, premature first-play monitoring and
mixed HTTP/readiness failure counts, plus bounded attachment/event handling
and fixed bootstrap availability detail. The supplied exact-68 isolated target
gate then passed all 60 tests, TypeScript and build with Client-Check-Exitcode 0.
Scoped exact-68 image preparation subsequently passed with
Image-Prepare-Exitcode 0 using helper `28d081a4`, fresh client build and both
images, while the active target remained exact-73. Backend evidence is
inherited through identical sources. Exact-68 activation then passed healthy
with Start-Exitcode 0; checkout/live are now exact-68, conventional HLS enabled.
The operator reports Retry was needed, then HLS worked without problems. The
initial error, timed/event interval and concurrent WebRTC result are not
separately supplied; reliable first start and wider live validation stay open.
NEXT existing read-only playback diagnosis while retaining the working service.
Do not repeat passed client/image/HTTP-denial gates without a new reason.

The operator could not reliably answer the follow-up about the exact five-minute
interval and reports possible random reconnects/room actions, without enough
detail to correlate transport, cause or duration. The operator explicitly found
the sequence of checks confusing. Record this as uncertain symptom evidence,
not a passed interval or a confirmed event-triggered HLS defect. No additional
ad-hoc command, replay or questionnaire is requested at this checkpoint.

Keep the currently working exact-73 service and saved evidence. Do not repeat
completed diagnosis/deployment/19 denial probes without a new failure/change.
Consolidate the remaining first-start, frozen-picture and room-event observations
into one bounded validation step after static triage; give the operator one
clear action and observable result at a time. The uninterrupted picture/audio/
room-event criterion remains open, to be grouped later rather than inferred
from cumulative successful requests. This does not close the larger matrix.

If a first-start failure recurs, record the fixed UI detail below "HLS failed"
and approximate wait before Retry/reload, whether picture/audio ever started,
and whether another HLS viewer was already playing. A warm packager join and
a cold start after the last HLS viewer leaves plus the 15-second idle grace
must be distinguished. Preserve a failed-state diagnostic before recovery if
needed; no broad rebuild, timeout/buffering change or LL-HLS switch is justified
by the single cumulative not-ready counter. The frozen-picture cause remains
open. Passive authorization/lifecycle, device/resource/isolation, dependency
maintenance and final grouped acceptance remain separate pending gates.
