# HLS / LL-HLS Phase 4 target-server validation

Repository assets prepared on 2026-10-04. **The automated/image preparation and
activation/invalid-input gates passed at exact application `93f1fa63`;
the first playback attempt FAILED and normal login then timed out.
Read-only diagnosis, rollback recovery and exact 80020d99 repair-image
preparation passed. Default-off repair-image deployment and operator-confirmed
normal login/picture/audio/control passed. Same-image HLS activation/19 denial
probes passed, but HLS failed again with an operator-reported all-stream outage.
Read-only diagnosis/default-off restoration passed, showing two timeline-gap
rejections and earlier valid server delivery without a sampled process crash.
The isolated target GOP A/B gate passed at repair commit 97ba4ad9 on 2026-10-05:
old code reproduced two timeline gaps; all three repaired codec tests passed,
including 30.19 seconds of scene cuts in generation 1. Exact 97ba4ad9
automated/image preparation then passed with Repair-Check-Exitcode 0: 47 client
tests, TypeScript/build, 13 Go packages, both fuzz jobs, all three codec tests
and server/base/Brave builds. Default-off repair-image deployment then passed
with Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes. The
operator reported the requested normal browser check works. The subsequent HLS
attempt failed with bootstrap failure and an operator-confirmed WebRTC outage.
Latest enablement CLI/HTTP outcomes are not supplied. Read-only diagnosis and
same-image default-off restoration passed with exit 0, healthy service and 2/2
disabled-route probes; the operator confirmed normal login/picture/audio again.
Saved startup analysis also passed: audio/high/medium subscriptions remained,
low reached only its capture-create marker. The isolated 48f4acf2 gate failed
on smooth cold readiness. Paired diagnosis from helper 88f2b25d then completed
with exit 0: each constructor passed two starts and failed one; medium began
before the high anchor and had only two parents at the failed deadline.
The isolated anchor A/B at repair 414639d2 then passed with
Anchor-Check-Exitcode 0: the old initial-IDR failure reproduced and seven
corrected checks each passed three repetitions; all six real-codec fixtures
stayed in generation 1. The subsequent full exact-414639d2 preparation FAILED
with Repair-Prepare-Exitcode 1: smooth readiness restarted with worker_failure
and failed its generation-1 assertion at 20.03 s; scene cuts passed at 30.19 s
in generation 1. Earlier client/type/build, Go/fuzz and registry/mapping/anchor
checks passed, but new base/Brave image steps were not reached. The 53034495
worker diagnosis completed with exit 0: all six checks passed in three fresh
processes, each smooth fixture ready at 18.11 s in generation 1; the earlier
worker failure was not reproduced. The controlled audio-anchor A/B at repair
71a14d21 passed with Audio-Anchor-Exitcode 0: old AAC blockage and an
audio/anchor/queue_full restart reproduced under injected high delay; all eight
corrected checks passed in three fresh processes and scene cuts passed once
(25 top-level passes). All seven real-codec fixtures stayed in generation 1.
Full exact-71a14d21 preparation then passed with Repair-Prepare-Exitcode 0:
all nine selected startup checks passed in the rebuilt GStreamer 1.26.2 image,
including normal/delayed-high readiness and scene cuts in generation 1;
base/Brave images were built and the running service stayed unchanged.
Default-off 71 deployment then passed with Baseline-Exitcode 0, healthy service
and 2/2 disabled-route probes. The operator confirmed the requested normal
browser check works without HLS. Same-image 71a14d21 conventional-HLS activation
then passed with Enable-Exitcode 0, healthy service, an enable snapshot and
17/17 public plus 2/2 cleartext-denial probes. The subsequent HLS attempt failed
with "HLS bootstrap failed; retry manually" after connecting; the operator
reported only WebRTC streaming works. Read-only diagnosis and same-image
default-off restoration passed; the operator confirmed normal login/picture/audio.
The diagnosis found 759/564 medium/low keyframe-admission drops and cumulative
part/segment publication only for audio/high. The isolated source-phase
diagnostic then passed: its aligned control became ready in 18.11 seconds,
and three cold skew runs reproduced not-ready in generation 1 at 24.02 seconds.
The common high-source fan-out repair A/B subsequently passed at a7ffb8b1 with
Shared-Clock-Exitcode 0: the old defect reproduced once; all fifteen corrected
checks passed in three fresh processes, plus one sustained scene-cut check
(46 positive top-level passes). All ten real-codec fixtures stayed in generation 1.
Full exact-a7ffb8b1 preparation then passed with Repair-Prepare-Exitcode 0:
all thirteen selected native/startup checks passed in the rebuilt GStreamer
1.26.2 image, including four real-codec fixtures in generation 1, and base/Brave
images were built. The supplied excerpt starts inside that codec-image build;
preceding client/Go/fuzz steps are covered by final script success without
separately shown fresh counts. Default-off exact-a7ffb8b1 deployment then
passed with Baseline-Exitcode 0, healthy service, a private baseline snapshot
and 2/2 disabled-route probes; the operator reported the requested normal
browser check works. NEXT section 2's same-image conventional-HLS activation
and HTTP denial probes, then section 3's bounded picture/audio attempt.
Working HLS playback
and grouped acceptance remain
PENDING.**
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
readiness/lease. [Startup repairs](HLS_STARTUP_REPAIR_2026-10-04.md) passed the new exact 80020d99 automated/image gate, including both real-codec
tests. The original normal-login blocker remains unconfirmed.
The subsequent rollback returned Restore-Exitcode 0 and a healthy prior image;
the operator confirmed normal login, picture and audio work again. Production HLS capture, picture/audio and valid authorization
remain pending. The
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

### Exact-a7ffb8b1 preparation after the shared-video-clock A/B passed

The complete supplied A/B output identifies repair
`a7ffb8b13448a8329c6df24fdcb182ac32ca398c` and the unchanged exact-71 codec image.
The old independent-source defect reproduced once at 24.02 seconds. Each of
the three positive processes passed all fifteen selected checks; sustained
scene cuts passed once. Normal readiness fixtures passed at 18.08 seconds,
delayed-high at 18.07 seconds and skewed-source at 18.83 seconds in each run.
Scene cuts passed at 30.07 seconds with continued parent advancement. All ten
real-codec fixtures stayed in generation 1 with all four tracks ready and no
rejected native pushes. The skewed video outputs shared IDRs at
30.8/32.8/34.8/36.8 seconds. The negative signature is an expected-defect
reproduction, not playback acceptance. Checkout, markers and live service
were unchanged. **NOT EXECUTED IN CODEX; supplied isolated repair A/B passed.**

The following completed preparation used a new private evidence directory.
It moved the clean `testing` checkout to the repair and ran the established
automated/image gate; it does not replace the running default-off 71 service,
edit Caddy or enable HLS. The permitted starting commits include the repair
itself so a stopped preparation can be resumed without an obsolete-head check.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_REPAIR_PREPARE'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git branch --show-current)" = testing
test -z "$(git status --porcelain=v1)"

repair_commit="a7ffb8b13448a8329c6df24fdcb182ac32ca398c"
case "$(git rev-parse HEAD)" in
  71a14d2174dafbc12b1880adde6dc68176bfe9af|"$repair_commit") ;;
  *) printf 'Unerwarteter Checkout; keine Änderung ausgeführt.\n' >&2; exit 1 ;;
esac
git fetch origin testing
git merge --ff-only "$repair_commit"
test "$(git rev-parse HEAD)" = "$repair_commit"

umask 077
bash deploy/validate-hls-phase4.sh \
  "../neko-hls-results-${repair_commit:0:12}"
NEKO_HLS_REPAIR_PREPARE
printf 'Repair-Prepare-Exitcode: %s\n' "$?"
```

The new output is `/opt/docker/nekoNew/neko-hls-results-a7ffb8b13448`.
Only final success validates its exact-commit marker. A failure leaves it
`PENDING`; preserve the output and the existing exact-71 evidence/images.
The old successful marker does not validate the new repair. Dependency-audit
exit 1 remains a report to classify, not a passing security audit.

This preparation subsequently passed with Repair-Prepare-Exitcode 0 and
AUTOMATED/IMAGE GATE PASSED. All thirteen selected native/startup checks passed;
the four real-codec fixtures stayed in generation 1 with normal/delayed/skewed
test durations 18.08/18.07/18.83 seconds and sustained scene cuts 30.08 seconds.
The application image is `my-neko/brave:hls-a7ffb8b13448`. Earlier client/Go/fuzz
output is outside the supplied excerpt, so only final script success covers
those preceding stages; do not assign fresh individual counts from this tail.
The running default-off 71 service was unchanged. Dependency-audit exit 1
remains an open report item. **NOT EXECUTED IN CODEX; supplied full preparation passed.**

### Exact-a7ffb8b1 default-off deployment and browser confirmation passed

The prepared exact repair image was deployed default-off, followed by the
operator's reported normal-browser confirmation. Same-image enablement,
bounded HLS playback and the grouped authorization/lifecycle/device/resource
matrix remain pending.

The following completed block restarted Neko, retained the adaptive/WebCodecs
overlays and saved the working 71 image as the new repair's rollback tag before
replacement. The deployer checks the exact preparation marker and image ID
before stopping the service; if the new image fails to become healthy, it
restores the saved image without the HLS overlay. No new build or Caddy change
is needed. Keep HEAD at the prepared repair, rather than newer documentation
commits on `origin/testing`, and retain both evidence directories.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_BASELINE'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "a7ffb8b13448a8329c6df24fdcb182ac32ca398c"

umask 077
output=/opt/docker/nekoNew/neko-hls-results-a7ffb8b13448
test "$(stat -c %a "$output")" = 700

bash deploy/deploy-hls-media.sh baseline "$output"
docker compose -f docker-compose.validation.yaml run --rm -T \
  hls-http-checks disabled </dev/null
NEKO_HLS_BASELINE
printf 'Baseline-Exitcode: %s\n' "$?"
```

The supplied output identifies exact application
`a7ffb8b13448a8329c6df24fdcb182ac32ca398c`, deployer blob
`c6f52dc80fdf605ec908f3fe3856ce23e015e494` and image
`my-neko/brave:hls-a7ffb8b13448`. The container became healthy, its private
baseline snapshot was recorded, both disabled bootstrap/media probes returned
their expected 404 (2/2), and Baseline-Exitcode was 0. The operator then replied
"borwser klappt" to the requested private-window normal login/picture/audio/
control check on `https://neko.taxzvps.de/` without a media override. This is a
bounded reported browser checkpoint, not a device/resource or room-event matrix.

Do not repeat baseline or pull documentation commits. NEXT section 2's
separate exact-a7ffb8b1 same-image enablement/denial-probe block, then section
3's one bounded admin HLS attempt alongside a working WebRTC viewer. Existing
71 evidence and rollback tags remain available. **NOT EXECUTED IN CODEX;
supplied default-off deployment and reported browser checkpoint passed,
same-image enablement/valid playback pending.**

### Exact-71a14d21 preparation/default-off deployment/browser checkpoint passed

The supplied output ended with AUTOMATED/IMAGE GATE PASSED and
Repair-Prepare-Exitcode 0. It begins inside the codec-image build, so earlier
client/type/build, Go/fuzz and server-check output is not included in this
excerpt; those preceding steps are covered by the unchanged fail-fast script's
reported final success. Do not infer fresh per-check counts from this tail.

The rebuilt GStreamer 1.26.2 codec image explicitly passed all nine selected
startup checks. Normal readiness passed at 18.11 s, delayed-high readiness at
18.09 s and scene cuts at 30.19 s, all in generation 1 with zero rejected pushes.
The video-IDR, AAC-drainage/timestamp-offset, cancellation/overflow and native
registry/mapping checks passed as well; the overflow/closure warnings belong
to deliberate unit cases. mediahls package duration was 66.704 s.
my-neko/base:hls-71a14d2174da and my-neko/brave:hls-71a14d2174da were built.
The final snapshot and success marker are under
/opt/docker/nekoNew/neko-hls-results-71a14d2174da. Dependency-audit report exit 1
is an open classification/remediation item, not a passing security audit.
During preparation, the running default-off 97 service remained unchanged. Full 414 preparation
remains a separate failed checkpoint; retain its PENDING marker and evidence.

The following block deployed the prepared exact-71 image with HLS disabled and
passed with Baseline-Exitcode 0. It restarted Neko, retained the adaptive/WebCodecs
overlays and saved the current image as the rollback tag before replacement.
The deployer checks the exact
preparation marker and image ID before touching the service, and restores the
saved image if the new container fails to become healthy. No Caddy change or
new build is needed. Keep application HEAD at 71; later documentation commits
on origin/testing are not the prepared image checkpoint.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_BASELINE'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "71a14d2174dafbc12b1880adde6dc68176bfe9af"

umask 077
output=/opt/docker/nekoNew/neko-hls-results-71a14d2174da
test "$(stat -c %a "$output")" = 700

bash deploy/deploy-hls-media.sh baseline "$output"
docker compose -f docker-compose.validation.yaml run --rm -T \
  hls-http-checks disabled </dev/null
NEKO_HLS_BASELINE
printf 'Baseline-Exitcode: %s\n' "$?"
```

The supplied baseline output identifies application
71a14d2174dafbc12b1880adde6dc68176bfe9af and deployer blob
c6f52dc80fdf605ec908f3fe3856ce23e015e494. my-neko/brave:hls-71a14d2174da deployed
healthy, the private baseline snapshot was recorded and both disabled HLS
bootstrap/media probes returned their expected 404 (2/2). Do not repeat the
baseline as the next step. The existing Caddy configuration was unchanged.

The operator replied "Ja ohne klappt alles" to the requested normal
login/picture/audio/control check without a media override. This closes the
default-off 71 browser checkpoint. It does not validate enabled HLS or identify
the earlier all-stream outage's cause. Section 2's exact-71 same-image
enablement/19 denial probes subsequently passed; NEXT section 3's bounded playback.
Enabled HLS and grouped acceptance remain pending. **NOT EXECUTED IN CODEX;
supplied preparation/baseline and operator-confirmed browser checkpoint passed.**

### Controlled audio-anchor A/B and completed full preparation block (historical)

The supplied worker diagnostic from helper 53034495 completed with
Worker-Diagnostic-Exitcode 0. All six checks passed in three fresh processes;
each smooth fixture reached readiness in 18.11 s in generation 1, with
audio/high/medium/low retaining parents MSN 1/2/3 and zero rejected pushes.
The audio first-output delay was 19/19/20 ms and high's 96/74/74 ms. The
overflow/closure warnings are from deliberate lifecycle unit cases, not the
real-codec intervals. The prior failure did not recur, so this observation
does not identify its cause or clear the failed full-preparation marker.

Static review identified a bounded startup hazard: the initial-output wait
also stops AAC drainage while high's initial video IDR is pending. At 48 kHz
with 1024-sample AAC access units, a stopped eight-frame output handoff can
fill within a small fraction of a second. The narrow repair leaves only video
waiting for high. AAC continues through ordinary admission, which discards
unanchored samples and retains the common-timeline offsets of later samples.
No queue/deadline/encoder/clock change or new buffer is added.

The supplied [validate-hls-audio-anchor.sh](../deploy/validate-hls-audio-anchor.sh)
run at repair 71a14d2174dafbc12b1880adde6dc68176bfe9af passed with
Audio-Anchor-Exitcode 0. Application HEAD remained 414639d2 and the running
default-off 97ba4ad9 service was unchanged. The helper pinned the image ID
recorded by the worker diagnostic, verified eight baseline source blobs and
held native/capture/encoder sources constant. Both sides used identical new
fixtures and the same fixed logging; only packager.go differed.

Both negative controls failed as required: the old audio-drainage unit reached
its fixed wait marker at 1.00 s, and the native fixture restarted into exactly
two generations with audio/anchor/queue_full. That fixture delayed only
generation 1's first high input by 256 ms, derived from twelve AAC frame periods.
The first audio worker closed at 207 ms after reading only one output; high
had not produced a frame. The negative codec test failed its generation-1
assertion at 18.39 s. This is an injected test condition, not a production
setting or proof of the earlier unobserved worker failure.

The corrected side passed eight startup checks in each of three fresh
processes, including normal/delayed-high codec readiness, initial video-IDR
retention, AAC drainage/timestamp offset, cancellation/overflow and registry/
encoder mapping. Scene cuts passed once in the first process: 25 top-level
passes in total. Normal readiness was 18.10/18.11/18.11 s; delayed-high readiness
was 18.09 s each; scene cuts passed at 30.09 s. All seven real-codec fixtures
stayed in generation 1 with no rejected pushes. Every cold fixture retained
all four tracks' parents MSN 1/2/3; the scene-cut fixture advanced to MSN 3/4/5.
The intentional overflow/closure warnings belong to their lifecycle unit cases.
Each process used -count=1 and a 90 s per-package timeout, with unchanged 24 s
HLS readiness. See [provenance and exact results](HLS_STARTUP_REPAIR_2026-10-04.md#controlled-audio-anchor-ab-passed--2026-10-05).

This passes the controlled AAC-overflow comparison, not full preparation or
live playback. The earlier failed 414 preparation marker remains PENDING.
**NOT EXECUTED IN CODEX; supplied target A/B passed.** Full exact-71a14d21
tests/images, default-off deployment and normal-browser checkpoint subsequently
passed as recorded above; enabled live acceptance remains pending.

The following block was run to fast-forward to the exact application repair
and prepare its tests/images in a separate directory. It passed; do not repeat
it as the current NEXT step. The newer documentation-only commit on
origin/testing records this evidence; it is not the application checkpoint.
This block never replaces the running default-off 97ba4ad9 service.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_REPAIR_PREPARE'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git branch --show-current)" = testing
test -z "$(git status --porcelain=v1)"

repair_commit="71a14d2174dafbc12b1880adde6dc68176bfe9af"
current_commit="$(git rev-parse HEAD)"
case "$current_commit" in
  414639d2493ad1d2106399337c647d1b007f1e9d|"$repair_commit") ;;
  *) printf 'Unexpected application commit: %s\n' "$current_commit" >&2; exit 1 ;;
esac
git fetch origin testing
git merge --ff-only "$repair_commit"
test "$(git rev-parse HEAD)" = "$repair_commit"

umask 077
bash deploy/validate-hls-phase4.sh "../neko-hls-results-${repair_commit:0:12}"
NEKO_HLS_REPAIR_PREPARE
printf 'Repair-Prepare-Exitcode: %s\n' "$?"
```

The preparation runs client tests/type/build, Go tests and both bounded fuzz
jobs, all nine selected real-codec/startup checks, audit-report collection, the
server build and uniquely tagged base/Brave image builds. It records success
only after all required steps pass. Review the output before any default-off deployment or
fresh browser check. Keep HLS disabled and preserve the private 414 failure
and A/B evidence, the working 97 image and prior rollback evidence. The codec
validation tag moves to 71 source; historical helpers that require 414 source
will reject that mutable tag. The audio-anchor A/B pins its historical image
ID independently. No image pruning is part of this block.

### Failed full preparation and completed worker diagnosis (historical)

The full preparation output ended with Repair-Prepare-Exitcode 1. All 47 client
tests, TypeScript/build, 13 Go packages, both 30 s fuzz jobs, the server build,
registry independence, encoder timestamp mapping and anchor lifecycle checks
passed. Smooth real-codec readiness reached generation 2 after a worker_failure
restart and failed its generation-1 assertion at 20.03 s; all four renditions
were ready in the replacement generation. Scene cuts passed at 30.19 s in
generation 1. The mediahls codec package failed at 50.553 s. No HLS sample
rejection/timeline-gap marker appears in that supplied codec interval.

The script stopped before dependency auditing and the new base/Brave image
build steps. Its preparation marker remains PENDING; do not deploy or enable
414 or change that marker manually. The passing controlled 71 A/B justifies
new exact-repair preparation without clearing the failed 414 marker. At this
historical checkpoint the service remained the confirmed default-off 97
deployment; the checkout and mutable codec-validation tag contained 414.
Old helpers that expect the 97 image intentionally reject this new tag.

The completed [diagnose-hls-worker-startup.sh](../deploy/diagnose-hls-worker-startup.sh) used
the exact 414 checkout and private output
/opt/docker/nekoNew/neko-hls-results-414639d2493a. Helper 53034495 was fetched
without merging it. The helper pins the rebuilt codec image ID, checks
eight original source blobs and mounts only fixed restart-stage logging and
bounded fixture observations. The first two replaced workers plus the latest
worker are retained per track; restarting cannot erase the first attempt.
Input push failure, output overflow/closure, anchor-wait overflow/closure and
audio output stall have distinct fixed stage/reason fields.

Three fresh containers/processes each run six startup checks, including smooth
readiness, registry/mapping and anchor lifecycle checks. They retain ordinary
queue sizes, codecs, clocks, restart decisions and the 24 s readiness deadline.
Each process uses -count=1 and the fixed three attempts include failures; no
retry-until-pass is allowed. The already-passing 30 s scene-cut fixture is not
repeated in this diagnosis. Exit 0 means all observations completed, even if a
test failed; it is not a preparation/playback gate. Raw logs contain synthetic
fixtures only and remain in a new private subdirectory. The helper never
accesses the live service, changes HEAD or replaces preparation markers.
The intentional overflow/closed-channel unit cases also emit worker restart
markers; classify real-codec failures only within their own test interval.
Observation code/helper: **NOT EXECUTED IN CODEX; supplied target results above**.

### Earlier passing anchor A/B and failed preparation block (historical)

The diagnosis/restoration block below completed with Diagnostic-Exitcode 0 and
Recovery-Exitcode 0 at 97ba4ad9. The image returned healthy without HLS and both
disabled-route probes passed. The operator confirmed normal login/picture/audio
again. Do not repeat recovery or enablement now. The saved diagnostic has 84 log
lines, one negotiation rejection, one not-ready bootstrap and no sampled
process exit/OOM or generation/lease-open marker. It does not establish the
rejection's role or prove that startup was never attempted.

The supplied saved analysis from helper 7a51ecbb passed with
Saved-Check-Exitcode 0: audio/high/medium subscriptions persisted, low source
generation remained zero and the fixed sequence stopped at its capture-create
marker before idle-stop scheduling. The earlier matching environment record
had HLS enabled. A global Go sample-registry lock held across native pipeline
construction is corrected in the repository; it can spread a constructor stall
to every media callback. The exact native blocker remains unconfirmed.

The supplied [registry A/B gate](../deploy/validate-hls-startup-isolation.sh)
at repair 48f4acf2 reproduced the expected old-constructor mutex wait (2.00 s).
The corrected registry test and encoder timestamp mapping passed; the smooth
fixture failed readiness at 24.02 s, while scene cuts passed at 30.19 s in
generation 1. Overall Isolation-Check-Exitcode was 1. This blocked rebuild and
deployment; that output did not identify the incomplete rendition.

The supplied [paired diagnostic](../deploy/diagnose-hls-codec-startup.sh) at
helper 88f2b25d completed with Startup-Diagnostic-Exitcode 0. Both original and
isolated constructors passed two starts and failed one. In each failure, medium
observed its initial 30 s IDR before the high anchor and had parents 2/3 only;
the other three renditions had parents 1/2/3. No input rejection or generation
restart was recorded. This supports the pre-anchor admission defect and does
not attribute it to the registry correction. Diagnostic success is not HLS
acceptance; no checkout/service/preparation-marker changes occurred.

The supplied [anchor A/B gate](../deploy/validate-hls-anchor-startup.sh) at repair
414639d2 verified the seven image-source blobs and held the registry correction
constant on both sides. Old packager.go failed the ordered initial-IDR test with
the expected marker; corrected packager.go passed all seven required checks in
three repetitions. Smooth real-codec readiness passed at 18.11/18.10/18.09 s;
scene cuts passed at 30.19/30.20/30.09 s, always in generation 1. Positive-control
exit and Anchor-Check-Exitcode were 0. Cancellation, overflow and closed drop
channels passed as well. The checkout, live service and prepared-image markers
were unchanged. See [the exact record](HLS_STARTUP_REPAIR_2026-10-04.md).

The following block was run to fast-forward the clean testing checkout from
97ba4ad9 to application 414639d2493ad1d2106399337c647d1b007f1e9d and attempt
full preparation. It failed as recorded above. Preserve both private output
directories and images; do not repeat it as the current NEXT step.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_REPAIR_PREPARE'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git branch --show-current)" = testing
test -z "$(git status --porcelain=v1)"

repair_commit="414639d2493ad1d2106399337c647d1b007f1e9d"
current_commit="$(git rev-parse HEAD)"
case "$current_commit" in
  97ba4ad9ab3e635da936a58c8a7ec795da05ba46|"$repair_commit") ;;
  *) printf 'Unexpected application commit: %s\n' "$current_commit" >&2; exit 1 ;;
esac
git fetch origin testing
git merge --ff-only "$repair_commit"
test "$(git rev-parse HEAD)" = "$repair_commit"

umask 077
bash deploy/validate-hls-phase4.sh "../neko-hls-results-${repair_commit:0:12}"
NEKO_HLS_REPAIR_PREPARE
printf 'Repair-Prepare-Exitcode: %s\n' "$?"
```

This runs the full selected client/Go tests, type/build, both bounded fuzz jobs,
same-commit codec checks and uniquely tagged base/Brave image builds. It never
replaces the running default-off 97ba4ad9 service. The codec-validation tag moves
to the newly built source; earlier A/B helpers deliberately reject that tag if
its source is no longer 97ba4ad9. Private 97 evidence and the old image ID remain
available; no image pruning is part of this block.

Any future full preparation must pass with final exit 0 before deploying an image
without HLS. Then obtain a fresh normal-browser confirmation before another
enablement. The earlier A/B pass is bounded fixture evidence, **NOT EXECUTED IN
CODEX**; full preparation, new-image deployment, native capture, browser/media
and grouped acceptance remain pending. Dependency-audit findings remain open;
their nonzero report status is not a passing security audit. Keep HLS disabled.

**Completed file-only analysis block:** extract [summarize-hls-startup.py](../deploy/summarize-hls-startup.py) from
the reviewed helper commit into the existing private directory, recording that
commit as in the operator block. Leave application HEAD at 97ba4ad9. Run:

```bash
output=/opt/docker/nekoNew/neko-hls-results-97ba4ad9ab3e
python3 "$output/summarize-hls-startup.py" "$output" \
  97ba4ad9ab3e635da936a58c8a7ec795da05ba46
```

It reads only the latest saved failure archive, prints fixed stage/rejection
enums and allowlisted capture/delivery metrics, and changes no service state.
Raw files remain private; share only its printed summary. An earlier matching
environment record and a room-wide stage sequence do not identify a correlated
cause. The supplied execution passed; **NOT EXECUTED IN CODEX**.
See [the repair record](HLS_STARTUP_REPAIR_2026-10-04.md) for supplied evidence.

**Completed incident recovery block:** retained for the record, not the next step.

Do not rerun preparation or enablement while this incident remains open. The
operator reported HLS bootstrap failure and WebRTC outage after the confirmed
97ba4ad9 default-off browser checkpoint. The latest activation CLI/HTTP results
are not supplied. Capture the failed state before restarting. The existing
diagnostic stores raw logs privately and prints only its fixed safe summary.
If it fails, retain its exit code and continue restoration. Baseline mode uses
the previously confirmed prepared 97ba4ad9 image without HLS, preserving the
adaptive/WebCodecs overlays; the saved rollback tag holds the prior image.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_DIAG_RESTORE'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "97ba4ad9ab3e635da936a58c8a7ec795da05ba46"
umask 077
output=/opt/docker/nekoNew/neko-hls-results-97ba4ad9ab3e
test "$(stat -c %a "$output")" = 700
diagnostic_status=0
bash deploy/diagnose-hls-playback.sh "$PWD" "$output" || diagnostic_status=$?
printf 'Diagnostic-Exitcode: %s\n' "$diagnostic_status"
bash deploy/deploy-hls-media.sh baseline "$output"
docker compose -f docker-compose.validation.yaml run --rm -T \
  hls-http-checks disabled </dev/null
NEKO_HLS_DIAG_RESTORE
printf 'Recovery-Exitcode: %s\n' "$?"
```

Supply only this block's printed safe summary and deployment/probe output.
Do not paste its private raw application/supervisor logs. Confirm normal
login/picture/audio/control in a fresh browser window after recovery. Diagnosis,
restoration and fresh browser recovery subsequently passed. Review the incident evidence
before any further HLS attempt. Keep checkout, evidence,
Caddy and master unchanged. This block is **NOT EXECUTED IN CODEX**.

The earlier preparation block below passed at exact 97ba4ad9 with
Repair-Check-Exitcode 0; keep its prepared images and private directory.

For this incident, fetch `origin/testing` and fast-forward to the exact reviewed
GOP-repair application commit `97ba4ad9ab3e635da936a58c8a7ec795da05ba46` below.
The isolated comparison passed; the restored live service remains on the prior
80020d99 image with HLS disabled during this new preparation. Preserve its
private results and rollback image. Later documentation-only commits need not
move the new application checkpoint.
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
repair_commit="97ba4ad9ab3e635da936a58c8a7ec795da05ba46"
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
integration tests, including sustained scene cuts, before building exact
uniquely tagged base/Brave images.
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

### GOP-repair default-off deployment/browser checkpoint passed

The supplied 97ba4ad9 preparation passed 47 client tests, TypeScript/build,
all 13 configured Go packages, both fuzz jobs, the trailing server/plugin build,
all three real-codec integration tests with GStreamer 1.26.2, and uniquely
tagged base/Brave builds. Readiness passed in 18.11 seconds; sustained scene
cuts passed in 30.19 seconds in generation 1. The running service was not
replaced. Audit exit 1 remains open findings. See
[the full repair record](HLS_STARTUP_REPAIR_2026-10-04.md) for exact limits.
These are supplied target results, **NOT EXECUTED IN CODEX**.

Keep application HEAD at 97ba4ad9ab3e635da936a58c8a7ec795da05ba46. Its tracked
deployer already contains the reviewed baseline action (blob
c6f52dc80fdf605ec908f3fe3856ce23e015e494); no helper extraction or checkout
update is needed. Run from /opt/docker/nekoNew/neko:

```bash
output=/opt/docker/nekoNew/neko-hls-results-97ba4ad9ab3e
bash deploy/deploy-hls-media.sh baseline "$output"
docker compose -f docker-compose.validation.yaml run --rm -T \
  hls-http-checks disabled </dev/null
```

The helper checks the exact preparation marker and image ID before replacement,
saves the current image for rollback and recreates Neko without the HLS overlay,
preserving adaptive/WebCodecs. Active sessions disconnect. A failed health
startup attempts the saved prior image. The supplied baseline run returned
Baseline-Exitcode 0 with helper blob c6f52dc80fdf605ec908f3fe3856ce23e015e494,
healthy my-neko/brave:hls-97ba4ad9ab3e and 2/2 public disabled-route 404 probes.
The operator then reported the requested normal login/picture/audio/control
check works. This is bounded target evidence, **NOT EXECUTED IN CODEX**; wider
role/recovery/device checks are not implied. The subsequent HLS attempt failed
and WebRTC also stopped working per operator report; use the diagnosis/recovery
record and saved-evidence summary above before any further enablement. Caddy and old private evidence are retained.
Do not infer enabled HLS playback from this baseline.

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

### Prior 80020d99 image: default-off deployment/browser smoke passed

The supplied preparation tail ended with Check-Exitcode 0 at exact 80020d99.
All 13 Go packages, both fuzz jobs, the trailing server/plugin build, the visible
client build and both uniquely tagged image builds passed. Both real-codec tests
passed with GStreamer 1.26.2; conventional readiness was logged in generation 1
and the all-four-rendition test passed in 18.11 seconds. The client test count was
not included in the supplied tail. Audit exit 1 remains open findings. This is
supplied target evidence, **NOT EXECUTED IN CODEX**; valid browser HLS playback
has not passed.

The following records the completed prior checkpoint; use the new 97ba4ad9
result directory only after its preparation passes and is reviewed.
For that prior checkpoint, application HEAD was kept at
80020d99477a58318f210b7e14d19cdd92991a6d. Fetch the
reviewed operator-tooling commit, record its full hash, and extract
[deploy-hls-media.sh](../deploy/deploy-hls-media.sh) into the existing private
../neko-hls-results-80020d99477a directory. Check that script with bash -n.
Invoke it from /opt/docker/nekoNew/neko with the explicit repository argument:

```bash
output=/opt/docker/nekoNew/neko-hls-results-80020d99477a
bash "$output/deploy-hls-media.sh" baseline "$output" "$PWD"
docker compose -f docker-compose.validation.yaml run --rm -T hls-http-checks disabled
```

The operator block pins the reviewed helper commit; its blob and the application
commit are recorded by the script. Baseline mode verifies the preparation marker
and image ID from private images.txt before changing the service. It captures the
currently working image under a rollback tag and recreates Neko using the
prepared repair image, with adaptive/WebCodecs and without the HLS overlay.
Active sessions disconnect. Failed health startup restores the saved prior image;
a public-probe or manual-browser failure requires explicit rollback/review.
The existing application-checkout rollback helper can use this new directory.
The supplied baseline run passed with Baseline-Exitcode 0 using helper commit
a7669dd184138ba53ea9398d31d0ca706ee0b400, blob
c6f52dc80fdf605ec908f3fe3856ce23e015e494. Image
my-neko/brave:hls-80020d99477a started healthy and both disabled-route probes
passed. This is target evidence, **NOT EXECUTED IN CODEX**. No additional
application change/build is required.

The operator confirmed normal login, picture, audio and control all work in the
requested private browser window at https://neko.taxzvps.de/. No wider
room-event/device matrix or enabled HLS playback was reported.
The subsequent section 2 activation passed, but the operator again reported HLS
bootstrap failure followed by all streams stopping. The subsequent read-only
diagnostic, default-off restoration and isolated GOP comparison passed, as
recorded below; exact 97ba4ad9 preparation and default-off deployment then passed.
The operator then reported normal browser checks work. Its subsequent HLS
attempt failed, and the operator confirmed WebRTC outage. Diagnosis/recovery
subsequently passed; saved-evidence analysis is now the next step. The original
activation used section 2's plain
image/probe helper, keeping this same application commit and evidence directory.
Do not repeat the completed Caddy source merge. Public valid-lease playback,
production capture skew and the full lifecycle/device/resource matrix remain
pending. No claim of the original timeout's exact cause follows from this gate.

## 2. Explicit conventional HLS deployment and invalid-input checks

The separate `docker-compose.hls.yaml` requires adaptive source geometry and
exact HTTPS/proxy values; empty HLS values reuse the reviewed WebCodecs values.
Its initial mode is only `hls`. Base Compose remains default-off.

### Exact-a7ffb8b1 same-image activation and denial probes — NEXT

Exact-repair preparation, healthy default-off deployment with 2/2 disabled
probes and the reported normal-browser checkpoint passed. Enable conventional
HLS in that same prepared image without changing the checkout, Caddy, source
profile, readiness deadline or queue limits. The reviewed deployer preserves
the saved rollback tag `my-neko/brave:rollback-hls-a7ffb8b13448` and compares
the current prepared image ID/marker before stopping the service. A health
failure attempts rollback automatically. An HTTP-check failure requires
review/default-off restoration before valid playback.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_ENABLE'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "a7ffb8b13448a8329c6df24fdcb182ac32ca398c"

umask 077
output=/opt/docker/nekoNew/neko-hls-results-a7ffb8b13448
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

This activation/17-public-plus-2-cleartext-denial gate is pending at a7ffb8b1.
Even a passing result does not open a valid lease or prove HLS picture/audio.
After healthy deployment and all nineteen probes pass, perform section 3's
single bounded admin picture/audio attempt alongside WebRTC. Preserve the
failed state through read-only diagnosis before any default-off recovery if
that attempt fails. Raw logs/credentials remain private. **NOT EXECUTED IN
CODEX; same-image activation/HTTP and enabled playback pending.**

### Exact-71a14d21 same-image activation and 19 denial probes passed

Exact-71 preparation, healthy default-off deployment/2 disabled probes and
the requested normal-browser checkpoint have passed. The operator confirmed
normal behavior without HLS. Use the unchanged reviewed deployer in the same
application/image/evidence checkpoint; do not pull later documentation commits
or repeat Caddy merge/reload. Existing adaptive/WebCodecs overlays and the
reviewed host-Caddy trust/logging configuration remain in use.

The following block passed with Enable-Exitcode 0. It restarted Neko with the
conventional-HLS overlay, then passed 17 public invalid-input cases and two
direct cleartext denials. The deployer
checks the prepared image ID and keeps the prior saved rollback tag. A startup
health failure attempts restoration automatically; an HTTP-probe failure
requires review/restoration rather than continuing to valid playback.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_ENABLE'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "71a14d2174dafbc12b1880adde6dc68176bfe9af"

umask 077
output=/opt/docker/nekoNew/neko-hls-results-71a14d2174da
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

The supplied output identifies application
71a14d2174dafbc12b1880adde6dc68176bfe9af and deployer blob
c6f52dc80fdf605ec908f3fe3856ce23e015e494. my-neko/brave:hls-71a14d2174da started
healthy; its private enable snapshot was recorded, and all 17 public checks
passed with expected statuses/headers plus both direct cleartext checks passed
with 403. Do not repeat activation as the next step. These synthetic invalid
credentials did not open a valid HLS lease or demonstrate packager readiness,
picture/audio or mixed-viewer isolation. NEXT one conventional-HLS admin
attempt with normal WebRTC behavior observed alongside it. If the reported
failure recurs, capture the read-only playback diagnosis before a same-image
default-off baseline restore; keep raw credentials/logs private. This block is
**NOT EXECUTED IN CODEX; supplied exact-71 enabled CLI/HTTP passed, playback pending**.

### Earlier Caddy/93f1fa63, 80020d99 and 97ba4ad9 activations (historical)

After the initial single-site guard stop and read-only
inspection, activation helper `2484a022` passed (exit 0). All 11 hosts and other
configuration were preserved; Caddy reloaded, synthetic runtime-error redaction
passed, `my-neko/brave:hls-93f1fa637ae3` started healthy, and the 17 public plus
two cleartext-denial probes passed. See [the Caddy record](HLS_LL_HLS_CADDY.md)
for the logging qualifications. The application stayed at `93f1fa63` for the
successful rollback recovery in section 3. Exact 80020d99 repair-image
preparation and default-off deployment/browser smoke passed. Its subsequent
same-image activation also passed with Enable-Exitcode 0, healthy service and
19/19 HTTP denial probes; the operator again reported failed HLS bootstrap and
all streams stopping afterward. Diagnosis/default-off recovery and the isolated
GOP A/B gate subsequently passed. Exact 97ba4ad9 preparation and default-off
deployment also passed; the operator reported the requested normal browser
check works. Its subsequent live HLS attempt failed with bootstrap failure, and
the operator confirmed WebRTC outage. The latest activation CLI/HTTP results
are not supplied. Diagnosis/default-off restoration and normal browser recovery
subsequently passed. The saved-evidence and codec comparisons later supported
the 71 repair; its passing preparation/default-off/browser checkpoint and
current enablement block are recorded above.
Do not reuse the old preparation
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
a new exact repair-image gate, the plain enable/probe block below was prepared if Caddy
and its logging configuration remain as reviewed, after normal login/picture/
audio/control have been confirmed on the new default-off image. The latest live
attempt failed; do not repeat this block until diagnosis and recovery have been
reviewed and a supported next change/checkpoint is prepared:

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_ENABLE'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "97ba4ad9ab3e635da936a58c8a7ec795da05ba46"
umask 077
output=/opt/docker/nekoNew/neko-hls-results-97ba4ad9ab3e
test "$(stat -c %a "$output")" = 700
bash deploy/deploy-hls-media.sh enable "$output"
docker compose -f docker-compose.validation.yaml run --rm -T hls-http-checks enabled </dev/null
docker compose -f docker-compose.validation.yaml run --rm -T \
  -e NEKO_PUBLIC_BASE_URL=http://127.0.0.1:8082 hls-http-checks insecure-denied </dev/null
NEKO_HLS_ENABLE
printf 'Enable-Exitcode: %s\n' "$?"
```

Adjust the direct port/prefix to the actual bind when needed. The deploy helper
keeps the rollback tag saved before the first baseline replacement,
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

### Failed playback/login incident: recovery and default-off repair baseline passed

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
mandatory real-codec validation job. Those tests/builds/images passed at exact
80020d99 with Check-Exitcode 0. Its default-off deployment then passed with
Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator
confirmed normal login/picture/audio/control work. Enabled HLS acceptance is
still pending. Same-image activation passed with Enable-Exitcode 0, healthy
service and 19/19 HTTP denial probes, but HLS failed again. The target Caddy
source already contains the reviewed changes. Preserve the failed-attempt
diagnostic before restoring default-off; do not repeat the source merger or
activation before review. Later repeat the bounded picture/audio checkpoint
and full valid-delivery matrix. The failure cause is not established by the
static startup findings.

### Repeated failed playback: preserve evidence, then restore default-off

The exact 80020d99 image passed activation and 19 HTTP denial probes, but the
operator reported `HLS bootstrap failed; retry manually`, tentatively recalled
a brief initial picture, then reported all streams stopped. This does not
establish an actual process crash or its cause. Normal playback passed with
HLS disabled immediately before activation.

Keep application HEAD at 80020d99477a58318f210b7e14d19cdd92991a6d and use the
existing private ../neko-hls-results-80020d99477a directory. Before stopping the
container, run [diagnose-hls-playback.sh](../deploy/diagnose-hls-playback.sh)
(blob 5b4b064926f83eccc28fe8bd596db93ba9c19ff1). Raw logs remain private in a
new timestamped report directory; share only the fixed-marker safe summary.
Capture/record its exit code even if it fails, then use the extracted reviewed
helper (blob c6f52dc80fdf605ec908f3fe3856ce23e015e494) in baseline mode with
the explicit repository argument. This restores the same prepared image with
HLS disabled; startup failure attempts the saved prior image. Do not delete
evidence, rebuild, change Caddy or retry enabled playback before review.
Both this incident's diagnostic and restoration passed with exit 0. The
container was healthy with no restart/OOM or sampled Neko exit; the summary
showed one successful bootstrap/lease, 23 successful segment requests and two
timeline-gap sample rejections. Same-image default-off restoration started
healthy and passed 2/2 disabled-route probes. Fresh browser confirmation remains
pending. These are supplied results, **NOT EXECUTED IN CODEX**.

### Isolated scene-cut GOP comparison passed (live HLS remains disabled)

The [repair record](HLS_STARTUP_REPAIR_2026-10-04.md) documents the exact evidence
and bounded scenecut=0 correction. Keep application HEAD at 80020d99 and its
existing private result directory. Fetch the reviewed repair commit and extract
[validate-hls-gop-repair.sh](../deploy/validate-hls-gop-repair.sh) privately;
record that full hash in gop-repair-commit.txt and check the script with bash -n.
This procedure completed at 97ba4ad9 on 2026-10-05; it is retained as evidence,
not the current NEXT block.

```bash
output=/opt/docker/nekoNew/neko-hls-results-80020d99477a
repair_commit="$(cat "$output/gop-repair-commit.txt")"
bash "$output/validate-hls-gop-repair.sh" "$PWD" "$output" "$repair_commit"
```

The helper verifies baseline source blobs in the existing codec image and pins
its ID. It requires the new hard-scene-cut test to fail with a timeline-gap
marker on old code, then all three integration tests to pass with only the HLS
transcoder correction mounted. Runs have no network, room credentials or
running-service attachment. Unexpected controls fail closed; reports remain
private. The supplied result has GOP-Check-Exitcode 0. The negative control
returned the required exit 1 after two low-rendition timeline gaps, generations
2/3 and failure to become ready in 24.01 seconds. The positive control returned
0: segment timestamp mapping passed, smooth startup passed in 18.11 seconds and
the scene-cut test passed in 30.19 seconds, remaining in generation 1 while all
four renditions advanced two further complete parents. Its pinned codec image
was sha256:ed572e4ef4cb4dbbf94b56027f4ac19ac47837a74aebe7a4b14860282a5b4a7f.
The [repair record](HLS_STARTUP_REPAIR_2026-10-04.md) contains full provenance and
limits. These are supplied target results, **NOT EXECUTED IN CODEX**; they do
not establish the all-stream outage's unique cause or production playback.

Full checks/build and fresh-image preparation subsequently passed at exact
97ba4ad9, followed by healthy default-off deployment and 2/2 disabled-route
probes and reported normal browser checks, as recorded in section 1. NEXT is
the separate same-image HLS activation/probe block.
Preserve prior evidence and the existing Caddy configuration.

### First bounded picture/audio checkpoint (repeat after diagnosis/repair)

**Current checkpoint: exact application/image a7ffb8b1, HLS disabled.**
The prior exact-71 HLS attempt failed at bootstrap; read-only diagnosis and
same-image default-off recovery passed. The common-source repair A/B, full
exact-repair preparation, healthy default-off deployment/2 disabled probes and
reported normal-browser checkpoint then passed at a7ffb8b1. NEXT section 2's
same-image activation/19 denial probes before this playback attempt. No valid
HLS picture/audio or five-minute room-event interval is demonstrated yet.
Use the new `/opt/docker/nekoNew/neko-hls-results-a7ffb8b13448` checkpoint and
retain the 71 evidence and both rollback tags separately; do not reuse old paths.

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
earlier playback attempts failed; the full five-minute checkpoint has not
passed at the current a7ff repair.

While HLS is active, the existing collector can preserve private metrics:

```bash
cd /opt/docker/nekoNew/neko
bash deploy/collect-hls-media.sh snapshot ../neko-hls-results-a7ffb8b13448 hls-first-playback
```

If bootstrap/playback fails or the ordinary WebRTC stream stops, preserve the
failed state through the read-only diagnosis below before restoring the
confirmed same-image default-off baseline. Do not keep retrying or restart
before capture. The diagnostic prints only a fixed safe summary; raw logs stay
private. Its failure must not prevent the baseline restoration.

```bash
set +e
bash -e -o pipefail <<'NEKO_HLS_DIAG_RESTORE'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "a7ffb8b13448a8329c6df24fdcb182ac32ca398c"

umask 077
output=/opt/docker/nekoNew/neko-hls-results-a7ffb8b13448
test "$(stat -c %a "$output")" = 700
diagnostic_status=0
bash deploy/diagnose-hls-playback.sh "$PWD" "$output" || diagnostic_status=$?
printf 'Diagnostic-Exitcode: %s\n' "$diagnostic_status"
bash deploy/deploy-hls-media.sh baseline "$output"
docker compose -f docker-compose.validation.yaml run --rm -T \
  hls-http-checks disabled </dev/null
NEKO_HLS_DIAG_RESTORE
printf 'Recovery-Exitcode: %s\n' "$?"
```

Share only the fixed diagnostic/deployment/probe output and confirm normal
browser behavior after restoration. This current exact-a7ff block has not
been needed or executed yet. The prior exact-71 diagnostic/recovery passed
with both exit codes 0 and reported normal browser behavior; that prior
result does not validate an a7ff enabled attempt. **NOT EXECUTED IN CODEX;
current enabled playback/conditional diagnosis-recovery pending.**

### Isolated source-clock-phase diagnostic after exact-71 recovery

Keep the running service HLS-disabled at application
71a14d2174dafbc12b1880adde6dc68176bfe9af. The saved live diagnosis found one
not-ready bootstrap, one started/idle-stopped packager generation and
medium/low keyframe-admission drops of 759/564. Cumulative init publication
exists for all tracks, but part/segment publication only for audio/high. The
zero object/running gauges follow the idle teardown and do not show that no
objects were ever produced. No sampled Neko exit/OOM was found.

The existing native fixture starts all video sources at the same PTS. Static
review identifies unequal source/GOP phases as a candidate omitted by that
fixture: a first medium/low IDR before high's anchor is discarded, and later
IDRs in odd one-second buckets cannot join the common six-second parent.
The live summary does not measure those timestamp phases.

[diagnose-hls-clock-skew.sh](../deploy/diagnose-hls-clock-skew.sh) mounts only
diagnostic test files into the existing exact-71 codec-validation image,
pins its immutable ID and verifies eight recorded source blobs plus unchanged
production sources. It runs one aligned positive control and three predetermined
cold processes with artificial high/medium/low PTS offsets of 800/50/100 ms.
These are test conditions, not recommended settings or measured capture skew.
The extra `hlsdiagnostic` tag excludes the expected-defect check from ordinary
codec-image/full-preparation acceptance. Exit 0 requires generation 1, no
rejected native pushes, flowing IDRs, audio/high ready and medium/low blocked
with no parts/parents. An unexpected signature stops the diagnostic.

Fetch the selected helper commit from `origin/testing`, export just its helper
outside Git into the existing private exact-71 output, check it with `bash -n`,
then invoke:

```bash
bash "$output/diagnose-hls-clock-skew.sh" "$PWD" "$output" "$helper_commit"
```

The operator block supplies the reviewed full helper commit. Do not merge/pull,
rebuild, change Caddy, enable HLS or replace preparation markers for this step.
Synthetic fixture output can be shared; live raw logs remain private.
This diagnostic subsequently passed at helper
409482b47e4ed962a2f9a3fd129114a53563f15a with Clock-Skew-Exitcode 0. The aligned
control reached all-four-track readiness in generation 1 at 18.11 seconds.
All three skew runs reproduced not-ready at 24.02 seconds, with audio/high
ready, medium/low at part index -1 and no parts/parents, flowing IDRs and no
rejected native pushes. This is a controlled defect reproduction, not the
actual unmeasured live timestamp phases. **NOT EXECUTED IN CODEX; supplied
control/reproduction passed.**

### Shared-video-clock repair A/B while the confirmed baseline stays running

The repair feeds one exact high-video subscription's same immutable encoded
units into all three existing scale/encode workers. One audio subscription
remains separate. The first video frame, provider PTS/DTS and fixed GOP phase
are shared; timestamps are not shifted to disguise skew. Capability output
IDs/geometry/rates stay high/1280x720/25, medium/854x480/20 and low/640x360/15,
with all input `source_id` values now high. Input FORMAT still requires the
complete exact high profile; emitted output caps are checked per rendition.
Provider capacity 64, native handoff 8, 24-second readiness and HTTP/security
limits are unchanged. A failed encoder input restarts the HLS generation;
partial construction and cancellation close the two owned subscriptions and
all created encoders. Three high-resolution decoders change CPU/RSS costs;
the later resource comparison must be repeated.

The following [validate-hls-shared-clock.sh](../deploy/validate-hls-shared-clock.sh)
gate subsequently passed at exact repair a7ffb8b1 with Shared-Clock-Exitcode 0.
For that comparison, the checkout/service stayed at default-off exact application
71a14d2174dafbc12b1880adde6dc68176bfe9af with the existing 0700 evidence directory.
The helper pins the codec image ID recorded by the supplied diagnosis,
sha256:2c885aa463ee9514b120d541e9852f4e9cd5cc334a72e9373cacf6896c963c48,
verifies eleven old source blobs and unchanged native/capture/encoder sources,
and mounts the selected repair sources read-only into isolated containers.
No image build, checkout move, live service access/restart, preparation-marker
replacement or Caddy edit is part of this gate.

The old 409482b4 fixture must reproduce its expected defect once. Three fixed
cold positive processes then run fifteen selected checks, including same-unit
fan-out and peer-failure cleanup, partial/successful construction cleanup,
cold FORMAT/generation handling, initial-IDR/AAC lifecycle, timestamp mapping,
normal readiness, delayed high input and the skewed-source case. Sustained
scene cuts run once in the first process. Positive codec fixtures require all
four tracks ready in generation 1, exactly one audio/high subscription, no
rejected native pushes and IDRs on the common clock. Suite timeout is 120
seconds to include the added fixture; each production readiness deadline
remains 24 seconds. Any unexpected signature/failure stops the gate.

Export the helper from the reviewed full repair commit outside Git and check
it with `bash -n`, then invoke:

```bash
bash "$output/validate-hls-shared-clock.sh" "$PWD" "$output" "$repair_commit"
```

The supplied operator block pinned repair
`a7ffb8b13448a8329c6df24fdcb182ac32ca398c`. Its old-code signature reproduced
once at 24.02 seconds; all fifteen positive checks passed in three processes,
plus one scene-cut check (46 positive top-level passes). All ten codec fixtures
remained in generation 1. Full exact-repair tests/image preparation in section 1
then passed with Repair-Prepare-Exitcode 0. Exact-repair default-off deployment
and reported browser confirmation then passed with Baseline-Exitcode 0, healthy
service and 2/2 disabled probes. Same-image enablement/denial probes are NEXT;
HLS-enabled live/device/lifecycle acceptance follows separately. **NOT EXECUTED
IN CODEX; supplied repair A/B, full preparation and default-off checkpoint
passed, enabled/live acceptance pending.**

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
