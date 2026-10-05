# HLS client startup deadline repair — 2026-10-05

Status: client-only repair `73d5ff6d29110e3dd06999726a7e88718d09ea23` is
implemented and statically reviewed on `testing`. Supplied target results
passed the old-defect reproduction, all 52 client tests, type/build and the
read-only live diagnosis, with Client-Check-Exitcode 0. Scoped exact-73 image
preparation then passed with Client-Image-Exitcode 0. Deployment/browser and
sustained playback remain **PENDING**. All execution was on the target server,
**NOT EXECUTED IN CODEX**. The application checkout is now exact 73d5ff6d;
the live a7ffb8b1 image remains running with conventional HLS enabled until
the next default-off deployment.

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
preparation passed, new image deployment/browser/acceptance pending.**

## Next target block: exact-73 default-off deployment and browser check

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
