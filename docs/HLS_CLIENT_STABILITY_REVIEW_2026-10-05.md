# HLS client startup and pause/recovery review — 2026-10-05

Status: **EXACT-7DCC3C5E TRACE CAPTURED FREEZE / EXACT-8741 TARGET PREPARATION PASSED / ACTIVATION AND LIVE ACCEPTANCE PENDING**.
The [2026-10-07 repair record](HLS_PLAYLIST_WINDOW_REPAIR_2026-10-07.md)
records the supplied successful old/new, 66-client-test/type/build, HLS
wire/package and image gate with Prepare-Exitcode 0. Target checkout is exact-8741;
live exact-7dcc was retained. NEXT activate the prepared image and check sustained
HLS playback alongside WebRTC. No additional old-image test or new build is needed.
All target results are supplied, **NOT EXECUTED IN CODEX**; live acceptance is pending.
The supplied exact-68 target result passed all 60 tests, TypeScript and build;
scoped image preparation subsequently passed with Image-Prepare-Exitcode 0.
All execution results below are supplied from the server, **NOT EXECUTED IN CODEX**.
At that checkpoint checkout and running deployment were exact application
`68dbdd4a8dd798886302b235c1f8f208452e0c6e`, conventional HLS enabled, with
Start-Exitcode 0 and a healthy container. The operator reports HLS needed
Retry, then worked without problems; reliable first-start acceptance remains open.
Read-only diagnosis subsequently passed: one not-ready and one successful
bootstrap, generation-1 server delivery and no sampled process exit/OOM.
The saved timing summary subsequently passed with Timing-Exitcode 0: one
request approximately 24 seconds, one at most 1 ms. The server-only candidate
now raises conventional readiness allowance from 24 to 28 seconds, preserving
the outer 30-second limits. Exact-8f54970f focused preparation subsequently
passed with Prepare-Exitcode 0, all 13 selected native checks and both candidate
images. Target checkout is exact-8f; live was exact-68 at preparation. After
the requested activation, an apparently frozen first picture was reported;
reload produced moving HLS video/audio. Read-only diagnosis subsequently passed
with Diagnostic-Exitcode 0 and matching prepared/live exact-8f image IDs, two
successful bootstraps, all four workers in one generation and 135 successful
segments; no sampled process exit/OOM or fixed error. Initial browser state
remains unknown. The client-only start-order candidate's preparation at exact
`7dcc3c5e4ba0279f35727eb4bed2e86bf721ad3b` subsequently passed with
Prepare-Exitcode 0: both images and private evidence recorded, live exact-8f
retained. Earlier isolated old/new client tests/type/build are covered by final
helper success, without separately visible counts in the copied build tail.
Target checkout is now exact-7dcc. After the requested activation, the operator
reports moving first picture after about 20 seconds, then a frozen picture
requiring reload. The subsequent progress diagnostic passed, confirming the
prepared/live exact-7dcc image and healthy service. Its valid 12.012-second sample
was idle; the operator confirmed HLS was closed/logged out. This does not capture
the freeze or establish a WebRTC-join/video-start cause. The corrected `4593a6f9`
progress helper subsequently passed: one active lease, all four workers and
continuing audio/video publication and successful HTTP media requests over
12.011 seconds, with no interval restart; VP8 rows are retained. The operator
confirms frozen picture/audio loss during this sample without reload, then later
"HLS HTTP connection failed; retry manually". The freeze is captured despite
server flow; the later HTTP failure was outside the sampled interval. NEXT the
read-only browser timing summary on the existing failed page, preserving the
service; no speculative timeout/buffer changes or repeated build/server gates.
Browser-helper execution is PENDING / NOT EXECUTED IN CODEX. Prior server
results are supplied evidence; the corrected saved-startup
helper has no fresh target execution. Reliable first start and
sustained/wider acceptance remain open.

## What the existing evidence establishes

PC/Helium playback worked after Retry and one frozen-picture reload. The
operator later reported possible random reconnects/room actions without enough
detail to correlate them. The exact-73 read-only diagnostic showed a healthy
container, six successful bootstraps, 394 successful segment requests and one
not-ready bootstrap, without a sampled process exit/OOM or fixed error marker.
These cumulative observations do not establish one uninterrupted interval or
the cause of a specific browser failure. See the
[readiness and diagnosis record](HLS_CLIENT_READINESS_REPAIR_2026-10-05.md).

## Confirmed client defects corrected by inspection

1. **Pause time could be counted as a stall.** `lastProgress` was retained
   through a deliberate Pause. On resume, `video.play()` can make `paused`
   false before the asynchronous `playing` event updates the clock. The next
   successful master poll could therefore terminate immediately after a pause
   longer than 20 seconds. Resume now resets the observed media time and clock
   before invoking Play. Repeated Play on an already unpaused frozen player
   does not refresh that clock.
2. **The stall watchdog could preempt initial readiness.** A first pending
   Play with `paused=false` allowed the 20-second progress check to run while
   the separate 30-second first-readiness deadline was still active. A
   per-player readiness flag now admits progress checks only after current
   `canplay`/`playing`, or the existing deadline's ready-state confirmation.
   First readiness resets the progress observation; player cleanup clears the
   flag. Startup remains bounded at 30 seconds, and a ready unpaused player
   still terminates after more than 20 seconds without media-time progress.
3. **HTTP and readiness failures shared one counter.** After three valid 503
   not-ready polls, one network failure incremented the same counter and met
   the three-HTTP-error terminal threshold. Those counters are now separate:
   a valid 503 clears consecutive HTTP failures, readiness remains bounded by
   30 not-ready polls, and successful validated media clears both counters.
   An authoritative private-mode transition starts a fresh readiness interval
   while preserving the same lease and periodic renewal. Repeated unchanged
   room pause state does not reset the counters.

The same lifecycle block also handles fast attachment readiness after the MSE
factory returns, preserving explicit autoplay/manual Play. A queued `playing`
callback is ignored if the element has already been paused; a genuine current
`playing` event establishes both playable and playing state. An asynchronous
old lease-watch continuation rechecks its controller generation after player
attachment before observing progress. These are defensive lifecycle repairs;
the supplied browser evidence does not show which event ordering occurred.

Bootstrap HTTP 503 now produces the fixed detail
`HLS media was unavailable at startup; retry manually`. This separates a
server-returned availability failure from the generic bootstrap failure
without exposing response bodies, URLs, tickets or cookies. Server 503 also
covers paused media, unsupported codec and missing source, so the new text
does not claim a proven readiness timeout. The one-time ticket is not retried
automatically; terminal Retry and manual WebRTC choice remain explicit.

## Room-event and surrounding-code findings

The reviewed client join/leave, chat and control handlers update member/chat/
control state; they do not directly stop or replace the HLS controller. The
legacy server bridge derives `media/hls/state.paused` from private mode and
admin status, independently of control locks. Unchanged pause state returns
before clearing a player. Video component track/stream and WebRTC recovery
handlers already exclude HLS, and its playing watcher leaves element playback
to the controller. No direct event-to-HLS-reconnect defect was established in
those paths. This static finding does not rule out browser load, actual event
socket loss or an unobserved server/delivery failure.

The scoped MSE loader's URL/body/concurrency bounds and generation cleanup,
server bootstrap error mapping, packager readiness/idle windows and private
resume path were also inspected. The existing cumulative not-ready count
did not by itself justify changing cold-start timeouts, codec/GOPs, buffering,
rendition selection or automatic recovery. The later supplied duration evidence
and bounded server-only candidate are recorded at the end of this document.

## Regression coverage prepared for the target

Eight additional deterministic controller tests cover:

- a first pending Play still waiting at 25 seconds, then genuine readiness and
  the unchanged later stall bound, through native and MSE paths;
- a 46-second intentional Pause followed by Play before its playing event,
  through both paths;
- repeated Play on a frozen unpaused player retaining the stall deadline;
- an already queued playing callback after deliberate Pause;
- immediate MSE readiness with autoplay enabled and disabled;
- valid 503 warm-up polls followed by one network failure, recovery, and the
  unchanged three-consecutive-HTTP-failure terminal bound;
- a private pause/resume with a fresh but still finite 30-poll warm-up budget,
  periodic keepalive and one bootstrap lease;
- fixed 503/generic bootstrap details and no automatic ticket retry.

Existing readiness-deadline, stale-player, autoplay, private renewal and native
revocation tests remain. The never-ready case also checks a pending Play with
`paused=false`, retaining the original 30-second startup cleanup bound.
Static diff and surrounding-source review completed;
no dependency, package script, backend or deployment configuration changed.
The supplied target gate below passed the full client tests, TypeScript check
and client build. Scoped candidate image preparation and healthy activation
also passed; reliable first-start and wider browser acceptance remain **OPEN**.

## One grouped target checkpoint later

The isolated client-checks and scoped image preparation stages below have now
passed, followed by healthy exact-68 activation and read-only diagnosis.
Continue with the saved-file duration summary at the end of this record;
do not rerun passed preparation without changed source or a new failure.

Keep the now-running exact-68 service and private evidence. Do not repeat
completed activation/denial probes or ask the operator to answer another
ad-hoc questionnaire. The image is deployed; its first-start reliability and
wider live acceptance remain open.

The completed first stage ran the existing client validation service against
an isolated checkout of the exact reviewed candidate, with the server fixture
tree retained:

```bash
# Real target server only; cwd is the isolated candidate Git/source root.
docker compose -f docker-compose.validation.yaml run --rm -T client-checks </dev/null
```

That service runs `npm ci`, the full client suite, `npm run lint` and
`npm run build` in its disposable container. It does not replace Neko. Record
the full candidate commit and final result. The older
`validate-hls-client-readiness.sh` and `prepare-hls-client-repair.sh` helpers are
pinned to the earlier a7ff-to-73 checkpoint; do not reuse them for this repair
or treat the prior 52-test result as validation of these changes.

The completed image stage uses the validated unchanged backend and prior-image
provenance. Group the next live browser check into one clearly described
checkpoint with a simultaneous WebRTC viewer. The bounded first activation
check below precedes wider cold/warm, join and deliberate Pause/resume checks.
Record one compact result; collect failed-state diagnostics before recovery
only if needed. Candidate activation has now passed; the supplied post-Retry
playback report does not establish reliable first start or the wider matrix.

The reported frozen picture and event-associated reconnects remain
uncorrelated; these source repairs are not a confirmed explanation of all
observed failures. Device, authorization/lifecycle, latency/resource/isolation,
dependency and final grouped promotion gates remain open. `master` stays
pinned; WebRTC remains the default and fallback stays manual.

## Supplied exact-68 isolated client gate passed

The complete supplied output pins candidate
`68dbdd4a8dd798886302b235c1f8f208452e0c6e` and private report
`/opt/docker/nekoNew/neko-hls-client-check-68dbdd4a-lJbw35xa`. The command fetched
Git objects, exported that candidate with `git archive` and used the existing
disposable `client-checks` service. It did not move the application checkout or
replace the running service.

All 60 client tests passed, including the eight new regression cases and the
expanded pending-Play startup bound: zero failures, cancellations or skips.
`tsc --noEmit` then passed. Vite 6.4.3 built 673 modules and the fresh main bundle
`index-urce48rY.js`; the build completed in 5.50 seconds. The output ended with
`CLIENT-CHECK PASSED: 68dbdd4a8dd798886302b235c1f8f208452e0c6e` and
`Client-Check-Exitcode: 0`. The command writes `candidate-commit.txt`,
`client-check-commit.txt` and `client-check.log` into that private report.

`npm ci` reported 20 affected package entries (11 low, 3 moderate, 5 high,
1 critical), matching the earlier aggregate classification. This abbreviated
output supplies no fresh advisory-by-advisory audit or security clearance.
Keep [dependency maintenance](DEPENDENCY_AUDIT_2026-10-04.md) open; no package
or lockfile changed and no automatic audit fix was applied. The chunk-size
warning is distinct from a build failure.

These are supplied target results, **NOT EXECUTED IN CODEX**. They validate the
exact client candidate, not a fresh server/codec/fuzz run, image deployment or
browser playback interval. The reported frozen picture and event-associated
reconnects remain uncorrelated pending live candidate validation.

## Completed image preparation: exact-68 while exact-73 stays running

`deploy/prepare-hls-client-stability.sh` is a separate target-only helper,
statically reviewed in the later tooling commit. Keep the application candidate
at exact `68dbdd4a`; extract the helper from its supplied exact tooling commit
into the existing private client report rather than checking out tooling as an
untested application image.

The helper consumes these five arguments:

```text
REPOSITORY  /opt/docker/nekoNew/neko
BASE_OUTPUT /opt/docker/nekoNew/neko-hls-results-73d5ff6d2911
CLIENT_REPORT /opt/docker/nekoNew/neko-hls-client-check-68dbdd4a-lJbw35xa
OUTPUT_DIR  /opt/docker/nekoNew/neko-hls-results-68dbdd4a8dd7
HELPER_COMMIT the full tooling commit provided in the operator block
```

Before moving the checkout, it checks the clean `testing` history, helper blob,
private exact-73 preparation record, exact-68 client markers and every exported
client/validation/fixture file against Git. It also requires unchanged
server/runtime/apps/build/Compose sources and dependency manifests, and checks
the live/tagged exact-73 image ID against the prior preparation record.

It then fast-forwards the application checkout to exact-68, records private
evidence, builds fresh `my-neko/base:hls-68dbdd4a8dd7` and
`my-neko/brave:hls-68dbdd4a8dd7`, saves image IDs and checks that the original
live container/image remain unchanged. The marker is written only after both
builds and the snapshot succeed. No stop/up/recreate/reload action is included.
Server/codec/fuzz evidence is explicitly inherited through the identical
exact-73/a7ff backend; the passed 60-test result is not represented as an A/B
reproduction of old code.

The operator block first runs `bash -n` on the extracted helper, then invokes
it once with these arguments. Supplied syntax checking and image preparation
passed as recorded below; activation/live acceptance were still pending at
that preparation stage. All execution is supplied, **NOT EXECUTED IN CODEX**.
Keep old images/evidence and `master`;
no completed HTTP-denial gate is repeated during this preparation block.

## Supplied exact-68 image preparation and subsequent activation

The supplied complete target output used tooling commit
`28d081a4430b43ba03d074157e009670f93b6b9a` and preparer blob
`a3494e08414ddb5cf6d2f95294f6fc6325d1a93d`. The extracted helper passed
`bash -n`, checked the passed private client report and baseline provenance,
then fast-forwarded the clean application checkout from exact-73 to exact-68.

The Docker client stage rebuilt with Vite 6.4.3: 673 modules, fresh main bundle
`index-urce48rY.js`, completed in 5.23 seconds. The unchanged server/runtime
build layers were cached. Both `my-neko/base:hls-68dbdd4a8dd7` and
`my-neko/brave:hls-68dbdd4a8dd7` were built; the Brave image installed version
1.96.61. No fresh Go, codec or fuzz run is shown or claimed. Backend evidence
is inherited through the identical exact-73/a7ff source and preparation chain.

The private directory `/opt/docker/nekoNew/neko-hls-results-68dbdd4a8dd7`
contains the candidate image-ID records, inherited/client evidence references,
preparation log, snapshot and completed validation marker. The helper's final
checks found the prior exact-73 container and image unchanged. The output ended
with `CLIENT STABILITY IMAGE GATE PASSED; backend evidence inherited; running
service unchanged.` and `Image-Prepare-Exitcode: 0`. Actual candidate image IDs
are retained in private `images.txt`, not printed in this supplied output;
Docker export digests are not substituted for those inspected IDs.

These are supplied target results, **NOT EXECUTED IN CODEX**. Application
checkout exact-68 and live container exact-73 are distinct states.

The supplied operator block below activated the prepared image once, using the
existing deployer and exact evidence directory. It required the application
checkout to stay pinned; later docs/tooling commits were not deployed:

```bash
set +e
bash -Ee -o pipefail <<'NEKO_HLS_START'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "68dbdd4a8dd798886302b235c1f8f208452e0c6e"
umask 077
output=/opt/docker/nekoNew/neko-hls-results-68dbdd4a8dd7
test "$(stat -c %a "$output")" = 700
bash deploy/deploy-hls-media.sh enable "$output"
NEKO_HLS_START
printf 'Start-Exitcode: %s\n' "$?"
```

This restarts Neko briefly. The unchanged deployer verifies the preparation
marker/image ID and saves the running image for rollback before stopping it.
The current conventional-HLS overlay and Caddy configuration are retained;
no second default-off baseline restart or repeat of the unchanged 19 denial
probes is required at this client-only checkpoint. No fresh public denial
result is claimed. A healthy-container result does not establish media success.

After Start-Exitcode 0, open a fresh private browser window at
`https://neko.taxzvps.de/?media=hls` and log in as admin for diagnostics. Observe
whether moving picture/audio start without Retry. Keep it running for five
minutes; in a separate normal WebRTC window at `https://neko.taxzvps.de/`, send
one chat message and take/release control once. Report start without Retry or
Retry needed, five-minute stability or the exact fixed UI error, and whether
WebRTC continues. No multi-part questionnaire is needed. Preserve the failed
state for a bounded diagnostic if it fails; do not infer a warm-up cause from
"HLS failed" alone.

At preparation, activation and this bounded browser interval remained pending.
The subsequent activation/playback result is recorded below. Wider device,
authorization/lifecycle, latency/resource/isolation, dependency and final grouped
acceptance remain pending. Preserve prior images/evidence and pinned `master`.

## Supplied exact-68 activation passed; first start still needs Retry

The supplied target command pinned application
`68dbdd4a8dd798886302b235c1f8f208452e0c6e` and the existing private directory
`/opt/docker/nekoNew/neko-hls-results-68dbdd4a8dd7`. It invoked the unchanged
deployer blob `c6f52dc80fdf605ec908f3fe3856ce23e015e494` with `enable`. Container
replacement completed healthy on `my-neko/brave:hls-68dbdd4a8dd7`, a private
enable snapshot was recorded, and the output ended with `Start-Exitcode: 0`.

The operator reports HLS first needed Retry HLS, then worked without problems.
Asked for the fixed first-failure detail, the operator could not recall it and
described only the default failed text; this does not identify an error branch.
The detailed first error, elapsed time, separately observed audio, exact
five-minute/room-event interval and concurrent WebRTC result are not supplied.
Do not silently mark those individual checks passed or classify this as a
confirmed packager, player or prior 30-second-deadline failure.

Static review confirms conventional bootstrap waits for all four tracks to
have three complete six-second parents, bounded by `ConventionalReadyWindow`
24 seconds. The client separately bounds the bootstrap request at 30 seconds
and initial player readiness at 30 seconds after attachment. A failed readiness
wait releases its reference; the last reference schedules a 15-second idle grace,
so packaging can remain active during that interval;
a manual Retry obtains a fresh ticket and can join already warmed media.
This explains how Retry could help a cold start, not which stage failed here.
No timeout, buffering, codec or automatic-ticket-retry change is justified yet.

NEXT retain the working service and collect the existing bounded read-only
diagnostic. Its fixed bootstrap result counters distinguish server not-ready,
successful lease creation and response-write failures; packager markers and
worker/object metrics provide surrounding evidence. Counts remain cumulative
and bounded logs are not an attempt-correlated trace. The helper is unchanged
at blob `ddfe001618f8733447f4bdd328a8519db9acef06` and already exists at the
application checkout, so no fetch/pull or restart is needed:

```bash
set +e
bash -Ee -o pipefail <<'NEKO_HLS_DIAG'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "68dbdd4a8dd798886302b235c1f8f208452e0c6e"
bash deploy/diagnose-hls-playback.sh "$PWD" \
  /opt/docker/nekoNew/neko-hls-results-68dbdd4a8dd7
NEKO_HLS_DIAG
printf 'Diagnostic-Exitcode: %s\n' "$?"
```

The helper writes raw evidence only inside a new private subdirectory and
prints its allowlisted summary. Share the printed summary, not private logs,
cookies or complete lease/WebSocket URLs. It does not request credentialed
playback, stop/recreate the container or change configuration.

Activation passed and post-Retry playback is reported, **NOT EXECUTED IN CODEX**.
This diagnostic subsequently passed as recorded below. Reliable first start
and wider live acceptance remain open; retain the existing images/evidence
and pinned `master`.

## Supplied exact-68 diagnosis passed; server availability rejection recorded

The supplied complete diagnostic ended with `Diagnostic-Exitcode: 0`, using
unchanged helper blob `ddfe001618f8733447f4bdd328a8519db9acef06`. It captured
137 application-log lines and available metrics. The exact-68 container was
running/healthy, with zero container restarts, no OOM kill and no sampled
supervisor Neko exit or fixed GStreamer/media error marker. Its inspected image
ID was `sha256:0669e21d1d3c12694f447348d76f1449846eda9992fff358672f14d5d07d38ae`.
The unavailable `gst-inspect` CLI version is not evidence of missing codecs.

The startup counters contain exactly one `bootstrap_total{result="not_ready"}`
and one success. All four workers started successfully once; each track has
one `startup` generation. There is one packager-ready marker, one opened/
active/closed lease and one idle-grace stop. All tracks published init, parts
and segments, with 17 successful segment requests, 36 master requests,
13 child-playlist requests, three init requests and two keepalives.
This establishes successful server media delivery after an availability
rejection and fits the reported Retry, without individual attempt correlation.

No active lease/worker or retained object remains at capture. That agrees with
the lease-close and idle-stop sequence; it does not demonstrate a crash or
claim that HLS was still playing at the diagnostic instant. The counters do
not establish five uninterrupted minutes, room-event behavior or simultaneous
WebRTC success.

The `not_ready` label aggregates the server's HTTP-503 availability mapping:
readiness wait failure, codec/source unavailability and paused media can share
that label. Also, `waitReady` maps both its own deadline and parent request
cancellation to `ErrPackagerNotReady`. Thus a server rejection is established,
but a 24-second cold-start timeout is not yet measured or proved. No timeout/
buffering or automatic fallback change has been made.

NEXT read the bootstrap-duration histogram already retained in private
`metrics.prom` using the extended file-only `deploy/summarize-hls-startup.py`.
It extracts only the registered `mode`/`resource="bootstrap"` histogram and
fixed bucket boundaries, finite nonnegative totals and integer observation
counts. Unexpected/duplicate labels, other resources, escaped values and
non-finite values cannot reach the output; duplicate valid series fail with
fixed text. The existing private-file, symlink, exact-commit and size guards
remain in place. No service access, credentialed request, dependency change or
fresh playback is required.

The emitted `bootstrap_request_durations` contains the cumulative observation
count, total seconds and cumulative buckets. Durations aggregate successes
and failures; buckets provide ranges, not an exact failed-attempt time or the
instant media first became ready. This bounded evidence can distinguish a
long wait from an immediate rejection and guide the next focused repair.

The extended helper blob is `9f99fe05046c32956d7850a96277de03ede2e657`.
At this checkpoint its target execution was pending; the supplied successful
result and next candidate are recorded below. It was NOT EXECUTED IN CODEX.

## Supplied saved timing gate passed; bounded server startup candidate

The supplied operator block extracted `deploy/summarize-hls-startup.py` from
tooling commit `833cff60ff7420c670ed681ce6b96f660053714f` into the existing
private exact-68 directory, without changing the application checkout or
running container. It ended with **Timing-Exitcode: 0**. The summary used saved
diagnostic stamp `20261005T202801416992672Z` and the same 137 log lines.

The conventional bootstrap histogram has two observations totaling
24.001096458 seconds. The cumulative 0.001-second bucket has one observation;
all buckets through 15 seconds have one, and the 30-second/+Inf buckets have
two. Consequently one request took at most 1 ms and the other approximately
24 seconds (24.000096458 to 24.001096458 seconds from the aggregate).
These are server request durations, not stream latency or first-picture time.
The duration histogram combines result classes and is not individually joined
to the one `not_ready` and one successful bootstrap counter.

The saved sequence has one startup generation, a scheduled idle stop before
packager readiness and one subsequently opened lease. Together with the
reported successful Retry, this strongly fits a first request exhausting the
24-second readiness window and a quick warm join. It does not establish the
exact instant all four tracks became ready, prove which duration belongs to
which result or identify the source of the extra startup time. Static review
found readiness notifications on publication and no missing wake-up in that
path; this is not a claim of exhaustive concurrency verification.

The bounded implementation candidate changes only
`ConventionalReadyWindow` in `server/internal/mediahls/packager.go`, from 24 to
28 seconds. This allows four additional seconds for conventional readiness
while remaining below the existing 30-second client bootstrap and server
response-write bounds. It does not add an automatic retry, ticket reuse or
backend switch. Earlier request work, cancellation or response delivery can
consume the remaining outer margin; 28 seconds is not an end-to-end guarantee.
LL-HLS still waits six seconds. The same conventional constant also applies
to private lease resume. All-four-track readiness, three six-second parents,
idle grace, codec/GOP, capture subscriptions, client and HTTP security remain
unchanged. The p95 24-second first-picture/audio and glass-to-glass acceptance
targets remain open and are not relaxed by this allowance.

At implementation the candidate was statically reviewed, with target runtime
verification pending and NOT EXECUTED IN CODEX. Its subsequent preparation
passed as recorded below. It may remove the observed boundary failure; a reliable
first-start fix cannot be claimed until target cold-start playback succeeds.
No new test merely asserting the constant is added. Existing full HLS package
tests and selected native startup/anchor/shared-input/scene-cut checks are the
focused automated gate; actual target cold start remains necessary.

## Next: prepare the server-only candidate without replacing live exact-68

`deploy/prepare-hls-startup-window.sh` is a separate target-only helper invoked
with repository, baseline evidence directory, new evidence directory and the
full selected candidate commit. It verifies a clean `testing` fast-forward
from exact-68, its own selected-commit blob, the baseline preparation marker
and both tagged/live image IDs. Source guards require the entire server delta
to be exactly the 24-to-28-second constant change, and unchanged client,
runtime, app, build, Compose, collector and deployer sources. Client/runtime/
configuration evidence is inherited from the exact-68 checkpoint; it is not
reported as a freshly rerun client or HTTP-denial matrix.

Only after those guards does the helper fast-forward the checkout and create
a new private 0700 directory outside Git and prior evidence. It builds
`server-checks` first, ensuring the native validation image inherits candidate
sources; runs `go test ./internal/mediahls -count=1`; builds and runs the
existing native HLS validation image; and builds separate candidate base/Brave
tags with the bundle from unchanged client sources. The server Dockerfile also compiles the
candidate. The native command uses `-count=1` and includes conventional
readiness, delayed high anchor, skewed sources and sustained scene cuts.
Unchanged client tests, unrelated Go packages, fuzz jobs and the 19 public
denial probes are not repeated at this focused checkpoint.

The helper has no stop/up/recreate/reload action. It verifies unchanged
container ID and live image before recording the private snapshot and final
candidate validation marker. A failed or incomplete preparation leaves the
marker pending and preserves prior evidence. Its syntax, focused tests and
image builds are **NOT EXECUTED IN CODEX**. The latest supplied live image
remains `my-neko/brave:hls-68dbdd4a8dd7`; no candidate deployment has occurred.

After preparation passes, activate the exact prepared image in one separate
operator step and check one fresh private-window HLS start without Retry;
retain a failure for the bounded diagnostic if necessary. Group the sustained
picture/audio/room-event comparison later. Frozen-picture, device, passive
authorization/lifecycle, numeric latency, resource/isolation, dependency and
final grouped acceptance remain open. Preserve all previous images/evidence
and pinned `master`.

## Supplied exact-8f54970f preparation passed; activation pending

The supplied target excerpt ends with **Prepare-Exitcode: 0** and
`STARTUP WINDOW PREPARATION PASSED; running service unchanged`. It records
the private snapshot and completed preparation in
`/opt/docker/nekoNew/neko-hls-results-8f54970f025e`, from the unchanged helper
blob `40f9536ac204342bbf5b8070380cc05ece637a21` at exact application
`8f54970f025e3491a123540cc94870508b66f119`. Native-test timestamps are
2026-10-05; this supplied result was carried forward after the interruption
on 2026-10-07. No verification was executed in Codex.

All 13 selected top-level native tests visibly passed. The four real-codec
fixtures remained in generation 1: normal conventional readiness at 18.09
seconds, delayed high anchor at 18.07 seconds, skewed sources at 18.84 seconds
and sustained scene cuts at 30.07 seconds. The package totals were 0.270
seconds for `pkg/gst` and 85.451 seconds for `internal/mediahls`. Controlled
negative tests emitted the expected missing-pipeline, queue-full, closed-drop
and push-failed warnings and passed; these are not live crash evidence.

The excerpt begins during the native image's dependency installation. Earlier
server-validation build/compilation and full HLS package tests are covered by
the strict helper's final success, without separately visible results or a
fresh test count. The base/Brave candidate images were exported successfully
as `my-neko/base:hls-8f54970f025e` and `my-neko/brave:hls-8f54970f025e`;
the Brave package remained 1.96.61. Unchanged client dependency/source/build
layers were cached, so no fresh client test/type/build run is claimed here.
The image inspection record and validation marker were saved privately;
BuildKit export hashes are not substituted for an inspected runtime image ID.

The helper verified unchanged container ID/live image before final success.
Target checkout is now exact-8f54970f, but the running image remains
`my-neko/brave:hls-68dbdd4a8dd7`. Automated readiness in these controlled
fixtures does not demonstrate the actual browser first-start correction.
Activation and one cold target-browser start without Retry remain pending.
Preserve previous images/evidence; `master` remains at `d9105ef8`.

## Next: activate the prepared image and observe one cold first start

Close existing HLS tabs before activation so an old viewer cannot reconnect
and warm the new packager before the intended first attempt. The existing
deployer blob `c6f52dc80fdf605ec908f3fe3856ce23e015e494` is unchanged. Keep the
target checkout pinned at the prepared application; later documentation-only
commits are not replacement deployment candidates.

```bash
set +e
bash -Ee -o pipefail <<'NEKO_HLS_START'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "8f54970f025e3491a123540cc94870508b66f119"

umask 077
output=/opt/docker/nekoNew/neko-hls-results-8f54970f025e
test "$(stat -c %a "$output")" = 700
bash deploy/deploy-hls-media.sh enable "$output"
NEKO_HLS_START
printf 'Start-Exitcode: %s\n' "$?"
```

This briefly restarts Neko, checks the candidate preparation marker/image ID
and retains the previous image for rollback. No second baseline restart,
rebuild or repeat of unchanged denial probes is requested. If container startup
fails, the existing deployer restores the saved prior image without the HLS
overlay; a healthy-container result does not establish playback success.

After Start-Exitcode 0, open one fresh private window at
`https://neko.taxzvps.de/?media=hls`, log in as admin and let that first attempt
reach moving picture/audio or its fixed failure detail without Retry/reload.
Report whether it started on the first attempt and the fixed detail if it
failed; retain the failed state for bounded diagnosis. Keep this step distinct
from the later sustained picture/audio/room-event and concurrent WebRTC check.
Device, authorization/lifecycle, numeric latency/resource/isolation, dependency
and final grouped acceptance remain open.

## Operator first-picture/reload report and requested read-only diagnosis

After the requested exact-8f activation, the operator reports the connecting
message, then a picture that apparently remained still. Page reload was needed
for moving video and audio, with HLS still selected. Record this as successful
playback after recovery, not a passed reliable first start. No fixed terminal
error, exact duration, initial Play/paused/muted/buffer state, separately
observed five-minute interval or concurrent WebRTC result was supplied.
The activation CLI, Start-Exitcode and inspected live image were not supplied
either; the report alone does not prove which image was active. No further
activation/rebuild is requested before capturing existing evidence.

Static follow-up reviewed controller initial canplay/playing handling, pending
Play and muted fallback, manual Play overlay, progress watchdog, native/MSE
attachment, scoped-loader queuing and the video-component HLS exclusions.
Canplay establishes playable media but does not establish observed playback
progress. Autoplay rejection has one muted attempt and the manual Play action;
the progress watchdog requires desired, ready, unpaused playback. These paths
do not establish whether the reported first picture was paused, buffering,
stalled delivery or another player failure. No new confirmed source defect or
attempt-correlated cause was identified; the implementation is unchanged.
No timeout/buffering/codec or automatic fallback change is made on this report.

The next target-only block prints the prepared image ID and invokes unchanged
diagnostic blob `ddfe001618f8733447f4bdd328a8519db9acef06`. Its safe summary
contains the actual container image ID, bootstrap/HTTP counters, generation/
worker/publication metrics and bounded fixed error markers. Comparing the
prepared and live IDs confirms whether the selected image is active, rather
than substituting checkout state for runtime evidence. The helper saves raw
evidence privately and prints only the allowlisted summary. It does not make
credentialed playback requests, stop/recreate Neko or change configuration.

```bash
set +e
bash -Ee -o pipefail <<'NEKO_HLS_DIAG'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "8f54970f025e3491a123540cc94870508b66f119"
umask 077
docker image inspect my-neko/brave:hls-8f54970f025e \
  --format 'Prepared image ID: {{.Id}}'
bash deploy/diagnose-hls-playback.sh "$PWD" \
  /opt/docker/nekoNew/neko-hls-results-8f54970f025e
NEKO_HLS_DIAG
printf 'Diagnostic-Exitcode: %s\n' "$?"
```

The page has already been reloaded, so current browser state cannot reproduce
the earlier frozen state and cumulative server counters may include both
attempts. This limitation must accompany the diagnosis; successful segment
requests cannot prove rendered/moving video or audible playback. The new
diagnostic result was pending at this checkpoint and is supplied below.
Further player-specific evidence may be needed if these server observations
do not isolate the defect.
Runtime checks were NOT EXECUTED IN CODEX. Keep the working service, previous
images/evidence, default/manual backend choices and pinned `master`.

## Supplied exact-8f diagnosis and client start-order candidate — 2026-10-07

The supplied read-only block passed with **Diagnostic-Exitcode 0**. Both the
prepared image and inspected running container report
`sha256:8221f57d0ed227af9356a4190335d8cb1a9d3192521ef369ca5708f8c7ee2962`,
confirming the exact-8f image is active. The container is healthy/running with
zero Docker restarts and no OOM kill or sampled Neko process exit. The bounded
390-line application sample contains no fixed error marker. The GStreamer
version probe is unavailable; that does not establish missing codecs.

The captured cumulative metrics show:

- two successful conventional-HLS bootstraps and no recorded not-ready result
  series in this snapshot;
- one startup generation and one successful worker start per audio/high/medium/
  low track, with all four workers still running;
- one active lease, zero opening leases, one init, 42 parts and six segments
  currently stored per track, with cumulative publication for all tracks;
- six successful init, 28 keepalive, 422 master, 129 playlist and **135 segment**
  requests. The fixed log sequence includes two lease opens and one close.

These observations fit two browser connections around the reported reload in
one shared generation. They do not correlate either bootstrap/lease to the
initial frozen picture, prove rendering or audio at that time, or establish
the cause of the browser issue. No server crash/restart mechanism is supported
by this sample. The activation CLI/Start-Exitcode remains unsupplied, although
the live image and healthy server delivery are now confirmed.

Static client inspection identifies a narrower improvement: the existing
automatic Play request waits for `canplay`, or for synchronous attachment that
already set the readiness flag. A first frame can be available before that
event. The [HTML play algorithm](https://html.spec.whatwg.org/multipage/media.html#dom-media-play)
permits requesting Play before future data is available, with a pending promise
until playback can start. The pinned
[hls.js 1.7.3 buffer attachment source](https://github.com/video-dev/hls.js/blob/v1.7.3/src/controller/buffer-controller.ts)
sets the media source during attachment; its
[control API](https://github.com/video-dev/hls.js/blob/v1.7.3/docs/API.md#fourth-step-control-through-video-element)
uses the video element's Play method.

The client-only candidate therefore invokes the existing Play path once the
current source/player has attached, when autoplay is desired and the element
is paused. It no longer requires `playbackReady` for that request. A pending
Play is still serialized; `canplay`/`playing` determine readiness and playing
state independently. The 30-second initial readiness and 20-second progress
budgets, manual Play, one muted attempt, disabled autoplay, user/private Pause,
generation guards, server codecs/configuration and backend choice are unchanged.
This removes an unnecessary readiness-event dependency; it does **not** prove
that the reported first picture missed canplay or that the live issue is fixed.

Three new target-only controller cases cover early Play with no data or only
a first frame on native/MSE players, disabled autoplay and pre-readiness Pause,
and a stale rejected Play after private resume. Existing blocked-autoplay
coverage now exercises both player kinds with rejection at attachment and no
unrequested repeat on readiness. The existing explicit pending-Play/Pause case
disables automatic Play so it continues to isolate the manual operation.
At candidate creation, execution was **PENDING / NOT EXECUTED IN CODEX**.
The supplied preparation result below subsequently closes that target gate.

### Prepared client/image gate, retaining the working service

Use `deploy/prepare-hls-client-start.sh REPOSITORY BASE_OUTPUT OUTPUT_DIR CANDIDATE`
from the exact reviewed candidate commit supplied with the operator command.
The target helper requires a clean exact-8f testing checkout, private validated
8f image evidence and a live image matching that evidence. It checks the
candidate/helper provenance and limits the delta to controller/tests, this
helper and documentation; backend, dependencies and deployment sources must
match the validated baseline.

It archives the candidate client/fixtures into a new private output directory,
checks that the old controller fails the specific missing-early-Play assertion,
then runs the complete candidate client tests, TypeScript and build in an
isolated validation container. After that gate passes, it fast-forwards only
the checkout, assembles candidate base/Brave images and records preparation
markers. It verifies that the same live Neko container/image remains in place;
it has no service activation or Caddy reload operation. Existing backend/native/
fuzz evidence is inherited where previously supplied, not freshly rerun.
Old image tags and evidence directories remain retained. If preparation fails,
preserve its private output/log and the working service.

No additional ad-hoc browser questionnaire or repeat denial probe is requested
for preparation. Candidate activation and one fresh-window first-start attempt
follow only after preparation success. Sustained moving video/audio, concurrent
WebRTC, room events, authorization/lifecycle, device/resource/latency, dependency
maintenance and final grouped acceptance remain open. `master` stays pinned.

## Supplied exact-7dcc3c5e preparation passed — 2026-10-07

The supplied output ends with **CLIENT START PREPARATION PASSED** and
**Prepare-Exitcode 0** from helper
`83b5924d39b3840875b8cf1a1b687a09fc54de31`, selected at application
`7dcc3c5e4ba0279f35727eb4bed2e86bf721ad3b`. The copied tail begins inside the
base-image dependency installation; the earlier isolated old/new assertion,
full client tests, TypeScript and client build are covered by the strict
helper's final success, but no fresh test count or client bundle hash is
separately visible in this excerpt. Do not invent those values or treat this
as live first-start acceptance. These are supplied target results,
**NOT EXECUTED IN CODEX**.

The excerpt explicitly shows completed base and Brave image exports as
`my-neko/base:hls-7dcc3c5e4ba0` and `my-neko/brave:hls-7dcc3c5e4ba0`.
The Brave build uses the candidate base and installs Brave 1.96.61. Private
image records, preparation markers and the final live-service snapshot were
recorded in `/opt/docker/nekoNew/neko-hls-results-7dcc3c5e4ba0`. The helper
guards show the exact-8f container/image remained unchanged through preparation.
Target checkout is now exact-7dcc3c5e; the live service remains exact-8f with
conventional HLS enabled. Identical backend/configuration/dependency sources
inherit previously supplied exact-8f evidence; no new server/native/fuzz or
HTTP-denial execution is claimed.

### Next: activate the prepared image once, then observe its first start

Keep the server checkout pinned at the prepared application. Later
documentation-only commits are not a replacement deployment candidate. Use
the unchanged deployer blob `c6f52dc80fdf605ec908f3fe3856ce23e015e494`:

```bash
set +e
bash -Ee -o pipefail <<'NEKO_HLS_START'
trap 'printf "Abbruch in Zeile %s, Exitcode %s\n" "$LINENO" "$?" >&2' ERR
cd /opt/docker/nekoNew/neko
test "$(git rev-parse HEAD)" = "7dcc3c5e4ba0279f35727eb4bed2e86bf721ad3b"

umask 077
output=/opt/docker/nekoNew/neko-hls-results-7dcc3c5e4ba0
test "$(stat -c %a "$output")" = 700
bash deploy/deploy-hls-media.sh enable "$output"
NEKO_HLS_START
printf 'Start-Exitcode: %s\n' "$?"
```

This briefly restarts Neko with the prepared candidate, checks the preparation
marker/image record and saves the prior image for rollback. No baseline
restart, rebuild or repeat of passed denial probes is requested. If candidate
container startup fails, the existing deployer restores the saved prior image
without the HLS overlay; healthy startup alone does not prove browser playback.

After Start-Exitcode 0, open a new private browser window at
`https://neko.taxzvps.de/?media=hls`, log in as admin, and let the first attempt
reach moving video or its fixed failure message without Retry/page reload.
If playback is deliberately blocked or muted by browser policy, the normal
Play/unmute action remains valid and must not be mistaken for a transport
failure. Report whether the first attempt reached moving video/audio, or the
fixed detail if it failed. Leave a failed state in place for bounded diagnosis.
Do not request another unrelated questionnaire at this checkpoint.

At the time of that activation request, candidate activation and first-start
evidence were **PENDING**. The subsequent operator report is recorded below. Sustained
video/audio, concurrent WebRTC, room events, authorization/lifecycle, device,
resource/latency/isolation, dependency maintenance and grouped acceptance stay
open; no promotion to `master` is authorized.

## Moving first picture then freeze — 2026-10-07

After the requested exact-7dcc activation, the operator reports moving HLS
picture after approximately 20 seconds, followed later by another frozen
picture requiring page reload. This is partial first-picture evidence, not
sustained playback acceptance or a measured latency percentile. The activation
CLI/Start-Exitcode and active image ID are not separately supplied. Exact-8f
remains the last independently confirmed live image until the new diagnostic
checks the prepared/live exact-7dcc image ID. The operator suggests a WebRTC
participant joining and starting a video as a possible trigger; its timing
and the frozen player's state/error/audio are uncorrelated.

### Static interaction review

- `server/internal/capture/streamsink.go`: a new video listener requests a
  keyframe when the keyframe-wait lobby was empty. Its listener start/stop paths
  create/destroy capture only when there are zero listeners, so an existing HLS
  subscription retains that shared pipeline during an ordinary new listener
  attachment. The keyframe event in
  `server/pkg/gst/gst.c` does not itself rebuild the capture pipeline.
- `server/internal/capture/media.go`: capture-generation, backwards DTS and
  actual format changes can advance a source generation; a keyframe alone does
  not. HLS consumes its independent audio/high subscriptions and fans the high
  source out to its three encoders in `server/internal/mediahls/packager.go`.
  Provider/worker failures and source discontinuities remain possible and have
  fixed metrics; no runtime absence of those failures is claimed.
- An actual desktop resolution change rebuilds video capture in
  `server/internal/capture/manager.go`. This is distinct from a new viewer or
  normal browser video playback; no resolution change was reported here.
- `client/src/neko/index.ts` handles member connection by adding member/chat
  state, without stopping the HLS controller. The receive-only guards in
  `client/src/components/video.vue` prevent the normal WebRTC stream watcher
  from overwriting the HLS element's media source. No direct join-triggered
  teardown/source overwrite was found in these inspected paths.

These findings constrain the hypothesis; they do not prove that a keyframe,
changing video content, resource pressure, browser buffering or audio focus
could not contribute. No speculative playback/capture/buffer change is made.

### Next: one bounded read-only progress diagnostic

The optional `--progress` mode in `deploy/diagnose-hls-playback.sh` retains the
original two-argument diagnostic. In progress mode it requires the application
checkout/preparation marker to agree and the running container image ID to
match the prepared exact-commit image record before sampling. Fetch the helper
from the reviewed helper commit into the existing private exact-7dcc results
directory, keeping the application checkout at
`7dcc3c5e4ba0279f35727eb4bed2e86bf721ad3b`; do not pull/merge the helper history
into the application or redeploy. The operator command supplies its full
helper commit. Run `bash -n` on the extracted helper, then:

```bash
bash "$output/diagnose-hls-progress.sh" "$PWD" "$output" --progress
```

Two credential-free loopback `/metrics` samples are taken 12 seconds apart.
The helper collects bounded private logs and selected container metadata, then
prints only fixed allowlisted labels/markers and numeric values. Its summary
separates capture bytes, HLS source units, published media bytes, HTTP results,
leases/workers, source generations, capture creation and packager-generation/
discontinuity counters. Successful WebRTC opens are aggregate counts; there
are no participant IDs or media URLs. New helper execution and Bash syntax
verification are **PENDING / NOT EXECUTED IN CODEX**; only static review and
Git whitespace checks are claimed here.

Keep the viewer open during sampling. If a freeze happens again, sample before
Retry/reload when possible; no induced join/stress test is requested. Reload
has erased the earlier browser state, so a normal sample after reload cannot
reconstruct that failure. Counter deltas are valid only if both scrapes succeed,
process-start metadata agrees and no observed counter regresses. Missing series
are not proof of zero production. Logs/cumulative counters may span several
viewers and attempts; production/request progress does not prove successful
browser decoding or establish which viewer requested a segment.

The diagnostic makes no service, image, proxy or media-session change and sends
no credentialed playback request. No repeated build/client/native/fuzz/HTTP-denial
gate is needed for this helper-only change. Preserve previous images/evidence;
reliable startup, sustained video/audio, room-event isolation and grouped
authorization/device/resource acceptance remain open. `master` stays pinned.

## Supplied progress diagnostic captured idle HLS — 2026-10-07

The supplied `02526dca` helper run passed Bash syntax checking and completed with
**Diagnostic-Exitcode 0**, using helper blob
`64bd2ea7a6e3b63d7eafe7db0a5890cb53d27a1e`. Its precondition confirms the running
image matches the prepared exact application
`7dcc3c5e4ba0279f35727eb4bed2e86bf721ad3b`:
`sha256:c445e6541db2eb4ddf821874480a37be86542f749d5bb83e0cc67ce9640bac25`.
The container was healthy/running, restart count zero, OOM false, with no sampled
Neko exit or fixed error marker. These are supplied target results,
**NOT EXECUTED IN CODEX**, independently confirming the new live image without
claiming the original activation CLI was supplied.

Both metric scrapes succeeded, process metadata agreed, no observed counter
reset occurred, and the interval was 12.012 seconds (`delta_valid: true`).
HLS active/opening leases, audio/high subscriptions and all four worker gauges
were zero before and after. HLS source-unit/publication deltas were zero, with
no new HTTP results, generations, discontinuities, drops, capture creations or
WebRTC opens. Cumulative logs show two lease opens/closures, one packager start/
ready/idle stop and seven negotiation rejections. The cumulative high-video
source-restart discontinuity is one, with no interval increment; each HLS track
has only one startup-labelled generation. None of these cumulative observations
correlates the earlier freeze to a WebRTC participant or establishes a crash.

Capability discovery in `client/src/neko/index.ts` also probes the other HLS
mode; conventional-only configuration in `docker-compose.hls.yaml` rejects
LL-HLS requests in `server/internal/mediahls/negotiator.go`. A negotiation
rejection count alone therefore does not establish failed conventional playback.
The seven sampled rejections lack mode/reason correlation in this summary;
do not loosen authorization or enable LL-HLS based on that count.

The operator confirmed **HLS was closed or logged out during sampling**.
Idle leases and the packager stop are therefore consistent with expected cleanup;
the zero deltas are not evidence of stalled active playback. A terminal client
failure also stops its periodic requests, and eventual lease expiry can lead
to idle cleanup while a failed page remains open. Distinguish that state from
an active moving/frozen viewer before interpreting a later sample.

Static inspection found why VP8 capture rows cannot appear in that summary:
the diagnostic progress parser and saved-startup parser allowed `vp8` in their
fixed codec allowlist but rejected digits in their preceding label-value regex.
Both parsers now admit digits **before the unchanged allowlist checks**. This
restores VP8 rows without admitting arbitrary codecs, identifiers or URLs. It
changes only diagnostic helpers; application/media code and previous accepted
HLS/lease/source metrics are unaffected. The omitted VP8 rows cannot be inferred
from the supplied output. New helper execution is **PENDING / NOT EXECUTED IN
CODEX**; static review and Git whitespace checks are the local verification.

NEXT fetch the corrected helper outside Git, keep the application checkout at
exact-7dcc and run the same bounded progress diagnostic with an authenticated
HLS viewer left open. Start sampling while it is playing, or promptly during
a recurrence before Retry/reload when possible. No build, activation, restart
or induced room action is requested. A moving sample demonstrates only that
interval; a sample after idle expiry cannot reconstruct the earlier player
failure. Keep prior images/evidence and the remaining acceptance gates open.

## Supplied active progress sample passed — 2026-10-07

The supplied `4593a6f9` progress helper run passed Bash syntax checking and ended
with **Diagnostic-Exitcode 0**, using corrected helper blob
`60cf47ddf8489c2397a5770a4c9523ae09a50b56`. The application remains exact
`7dcc3c5e4ba0279f35727eb4bed2e86bf721ad3b`; prepared/live image ID remains
`sha256:c445e6541db2eb4ddf821874480a37be86542f749d5bb83e0cc67ce9640bac25`.
The container was healthy/running, restart count zero and OOM false, with no
sampled Neko exit or fixed error marker. Both scrapes succeeded, process
metadata agreed and no observed counter reset occurred. The interval was
12.011 seconds with `delta_valid: true`. These are supplied target results,
**NOT EXECUTED IN CODEX**.

| Signal | Observation over the interval |
| --- | --- |
| HLS sessions | One active lease persisted; no opening lease |
| Source subscriptions | One audio and one high-video subscription persisted |
| Source units delivered to HLS | 601 audio and 300 high-video units |
| Capture bytes | +192,921 Opus and +3,039,667 high VP8 bytes |
| Packagers | All four workers running before and after |
| Publication | Every audio/high/medium/low track published new parts and segments |
| HTTP results | Success: 12 masters, 2 playlists, 2 segments, 1 keepalive; no observed failure increment |
| Pipeline/generation changes | No capture-creation, HLS-generation or source-discontinuity increment |
| Drops and WebRTC opens | No observed HLS drop or successful WebRTC-open increment |

The restored VP8 capture rows confirm the diagnostic parser correction is
exercised. Low/medium capture bytes remained zero, while their HLS renditions
published positive bytes: this fits the intended one-high-source fan-out and
does not show stalled HLS lower renditions. The separately corrected
`summarize-hls-startup.py` is statically reviewed but has no fresh target result.

Cumulative startup-labelled HLS generations are now two per track; source
audio/high generation gauges stayed at two throughout the interval. The
previously recorded idle stop and a second cold start fit those observations;
they do not establish a process crash or restart loop. The high-video
source-restart discontinuity remains one without interval increment. Six
cumulative WebRTC opens and nine negotiation rejections cannot correlate a
join with the earlier browser freeze; no new successful WebRTC open was
observed during this sample.

This gate confirms active server production and successful media HTTP requests
for one bounded interval. It does not establish the browser's media-time,
buffer/paused/decoder state, audible audio, uninterrupted five-minute playback,
reliable cold start or room-event isolation. The operator then clarified that
initial moving picture/audio stopped, and this diagnostic was run during the
frozen picture/audio loss without reload. The page later reported the fixed
HTTP-connection failure below. Do not label the freeze fixed or infer the later
HTTP response status from this earlier sample. No new build, deployment or
routine repetition of the passed diagnostic is requested.

## Confirmed browser freeze and later HTTP failure — 2026-10-07

The operator's browser report correlates frozen picture and lost audio with
the supplied active 12.011-second server interval: capture inputs, every HLS
rendition's publication, one lease/all workers and successful HTTP requests
continued. This narrows the failure to playback/delivery beyond the observed
server production; it does not prove complete/valid decoding of the responses
or exclude an earlier media/timeline problem. The page was not reloaded, and
later showed **HLS HTTP connection failed; retry manually**. The later HTTP
failure was not captured by the supplied server interval.

In `client/src/neko/hls/controller.ts`, that detail is emitted after three
consecutive failures in the serial lease-watch loop. Master and keepalive
requests use the default **1,000 ms for headers and the complete body**;
bootstrap uses 30 seconds, native child validation five seconds, and the MSE
loader has its separate maximum 12-second request budget. The generic HTTP
detail also covers nonterminal HTTP status, network/abort/body-read errors and
invalid master validation. It does **not** identify which of these occurred.
Valid 503 not-ready responses have a separate readiness budget. The 20-second
player progress watchdog checks only a ready, desired, unpaused player after a
successful master poll. Freeze and later generic HTTP failure must therefore
remain separate observations; no timeout/buffer/lease policy change is justified
from the generic message alone.

### Completed step: read the failed page's retained request timings

That checkpoint requested `deploy/inspect-hls-browser-timing.js` in DevTools
**on the existing Neko HLS page**, without Retry/reload. The helper performs a single synchronous read
of retained same-origin HLS Resource Timing entries and selected current video
numeric/boolean fields. It sends no requests, reads no cookies/storage/headers
and changes no player, timer, fetch implementation or performance buffer. It
prints only fixed resource/rendition classes and numeric timing/status/body-size
fields; URL, lease ID and query strings are never printed. The operator message
supplies the complete snippet; no script download or server action is required.
The supplied result is now recorded below; it was **NOT EXECUTED IN CODEX**.
Static inspection and Git whitespace checks are the local verification.

The [W3C Resource Timing specification](https://www.w3.org/TR/resource-timing/)
includes completed/aborted fetch timings, while blocked preconditions can lack
entries; its buffer may be full and `responseStatus` is not available in every
engine. The helper retains zero versus missing status but neither proves a
timeout. Compare the numeric request order/duration/status and freshness; no
recorded recent failure is not proof that all later probes succeeded. It cannot
read response bodies or identify master-validation errors, full client decoder
state or why an aborted request was canceled. Terminal cleanup already clears
the video's original buffer/source, so current video fields after the failure
cannot reconstruct the original freeze.

Preserve the exact-7dcc service and old images/evidence. If retained timing
entries are insufficient, a later bounded browser trace must be designed before
another reproduction. Do not request raw HAR, full WebSocket URLs or all console
logs. The WebRTC-join hypothesis, reliable start, sustained playback and grouped
authorization/device/resource/isolation acceptance remain open; no promotion.

## Supplied retained browser timings — 2026-10-07

The PC/Helium console attachment supplies the `hls-browser-timing-v1` JSON.
The operator pasted the previously supplied equivalent formatted snippet;
no separate file/blob execution claim is made. This is supplied browser evidence,
**NOT EXECUTED IN CODEX**. The application remains exact-7dcc; no new server
change or preparation result is supplied.

| Field | Supplied observation |
| --- | --- |
| Query time since navigation | 732,848 ms |
| Retained Resource Timing entries | 250; 236 matched HLS in the preceding summary |
| Last retained master | Started 181,362 ms; duration 1,001 ms; no first byte/body/transfer metadata; status zero |
| Most preceding master requests | 16–30 ms with 637 decoded-body bytes; some later requests 58–152 ms |
| Retained medium segment requests | 68–93 ms with roughly 800–852 kB decoded-body data; statuses include both 200 and zero |
| Current video after failure | Paused, ready/network states zero, time zero, no current media error |

The last retained request ended at approximately 182,363 ms, about 550,485 ms
(9.17 minutes) before the query. The full buffer and old entries cannot provide
the later consecutive-failure sequence. A repeated snapshot of this buffer
cannot recover omitted timings. There is no proof that all requests after the
retained boundary succeeded or failed, or when the freeze began relative to it.

The 1,001-ms empty master is consistent with the controller's 1,000-ms deadline
and supports a timeout-like attempt. Resource Timing does not identify who
canceled it; this alone does not prove the three errors required for the later
terminal HTTP detail or explain the preceding frozen picture/audio loss.
In particular, **status zero is not failure proof**: several preceding fast
master/segment entries have substantial body/transfer metadata despite that
status. These fields do not prove playlist validity or browser decoding.
Current video fields fit the inspected terminal cleanup (`clearHLSVideo`)
and cannot reveal the original paused/buffer/decoder state. Copied heartbeat
messages do not establish uninterrupted WebSocket/event delivery or room-action
causality. No timeout, buffering, lease or authorization policy is changed.

## Completed investigation: one bounded browser trace

Use [`deploy/trace-hls-browser.js`](../deploy/trace-hls-browser.js) in DevTools on
the **existing Neko HLS page**, before one manual Retry if it is already failed.
Leave that tab open/foreground; do not reload or induce additional room actions.
The helper does not Retry itself. An initial old terminal state does not finish
the trace until new HLS activity is observed. It prints one JSON report two
seconds after a new terminal error, or after a **five-minute maximum**, then
removes its observers/listeners/timers and temporary stop function. For an
earlier finish use `window.__nekoHLSBrowserTraceStop()`; pasting it again ends
the previous trace before arming a new one. Closing/reloading the page ends the
context and can lose the report. Foreground timers may still be delayed by a
busy browser; duration and visibility samples expose that limitation.

It uses an unbuffered Resource `PerformanceObserver` for new completions,
samples selected numeric/boolean video buffer/time/frame/UI fields every two
seconds and records fixed media events. Request start times are relative to
installation and can be negative for a request already in flight. Output has
fixed resource/rendition/error classifications and bounded rows: 180 recent
HTTP entries, 40 requests lasting at least 900 ms (not classified as timeouts),
155 media samples, 80 events and fixed request counts. Discard counts expose
truncation. Only the first four buffered ranges are sampled; the full range
count is recorded. Neither bodies nor decoder internals are inspected. The
helper sends no traffic, touches no player/fetch/performance-buffer policy and
reads/prints no cookies, storage, headers, URLs, lease identifiers or raw errors.

The [W3C Performance Timeline queue algorithm](https://www.w3.org/TR/performance-timeline/#queue-a-performanceentry)
queues matching observer entries before checking the retained global buffer.
This avoids relying on the already full 250-entry snapshot, subject to the
browser's implementation and omitted/precondition-blocked requests. It cannot
recover the old failure; installing it before Retry is necessary to capture a
new failure alongside media state before terminal cleanup. Zero/missing status
still does not prove an error; frame counts/media time do not prove audible
sound or valid timestamps. Successful requests with frozen buffers/time would
direct a later focused player/timeline review, while repeated empty near-deadline
masters would support a separate HTTP-watch investigation. Neither result may
be invented before supplied execution.

The complete five-minute report was subsequently supplied on 2026-10-07:
moving video begins at 27.149 seconds, stalls at media time 23.948 seconds with
513 frames, then seeks forward every six seconds while the buffered range and
frame count remain fixed. Audio stops loading after four segments; medium
video continues to load. No new terminal HTTP error appears before the time
limit. This is a captured failure, not five-minute acceptance.
See [the supplied trace and bounded repairs](HLS_PLAYLIST_WINDOW_REPAIR_2026-10-07.md)
for the rolling discontinuity defect, seek-only watchdog correction and exact
target preparation scope. The earlier trace instructions are historical;
another trace on the old image is not requested. Exact-8741 repair preparation
subsequently passed with Prepare-Exitcode 0 and all 66 client tests/type/build,
the old/new wire check, full HLS package and server/base/Brave builds. Target
checkout is exact-8741; live exact-7dcc was retained. Supplied target evidence,
**NOT EXECUTED IN CODEX**. Activation and live sustained acceptance remain pending.
