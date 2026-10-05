# HLS client startup deadline repair — 2026-10-05

Status: client-only repair `73d5ff6d29110e3dd06999726a7e88718d09ea23` is
implemented and statically reviewed on `testing`. Target A/B, client tests,
type/build, new image preparation/deployment and sustained playback are
**PENDING / NOT EXECUTED IN CODEX**. The running application remains exact
`a7ffb8b13448a8329c6df24fdcb182ac32ca398c`, with conventional HLS enabled.

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
No new live diagnostic output has yet been supplied.

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

## Next target block: read-only diagnosis and isolated client A/B

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

WebRTC continues working per the operator, so no immediate recovery restart
is required for this client gate. If normal playback also fails, capture the
read-only diagnosis and restore the confirmed same-image default-off baseline
using [the recovery block](HLS_LL_HLS_VALIDATION.md#first-bounded-pictureaudio-checkpoint-repeat-after-diagnosisrepair).
Keep a7ff/71 evidence and rollback tags. Earlier all-stream outage causes,
TV compatibility, sustained A/V, authorization/lifecycle, resource/isolation,
dependency remediation and grouped acceptance remain open. `master` stays
pinned; there is no automatic fallback or promotion.
