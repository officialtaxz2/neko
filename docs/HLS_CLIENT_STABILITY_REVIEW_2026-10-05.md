# HLS client startup and pause/recovery review — 2026-10-05

Status: **TARGET CLIENT GATE PASSED / IMAGE AND LIVE ACCEPTANCE PENDING**.
The supplied exact-68 target result passed all 60 tests, TypeScript and build.
All execution results below are supplied from the server, **NOT EXECUTED IN CODEX**.
The latest supplied running deployment remains exact application
`73d5ff6d29110e3dd06999726a7e88718d09ea23`, conventional HLS enabled, with
WebRTC working. This review changes repository source only.

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
does not justify changing cold-start timeouts, codec/GOPs, buffering, rendition
selection or automatic recovery. Those values remain as previously specified.

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
and client build. Candidate image preparation and browser acceptance remain
**PENDING**.

## One grouped target checkpoint later

The isolated client-checks stage below has now passed. Continue with exact-68
image preparation described at the end of this record; do not rerun the passed
client stage without a changed source or a new failure.

Keep the working exact-73 service and private evidence while this candidate is
reviewed. Do not repeat completed activation/denial probes or ask the operator
to answer another ad-hoc questionnaire now. The new repository commit is a
candidate, not a deployed or accepted image.

First run the existing client validation service against an isolated checkout
of the exact reviewed candidate, with the server fixture tree retained:

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

Only after that passes, prepare the exact candidate image with the validated
unchanged backend and rollback provenance. Group the live browser checks into
one clearly described checkpoint: cold and warm HLS start, picture/audio during
chat/join/control actions, and deliberate Pause/resume with a simultaneous
WebRTC viewer. Record one compact result; collect failed-state diagnostics
before recovery only if it fails. No image preparation, activation or live
acceptance has been performed for this candidate.

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

## Next operator block: prepare exact-68 images while exact-73 stays running

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
it once with these arguments. Syntax checking, image preparation and live
candidate acceptance are **PENDING / NOT EXECUTED IN CODEX**. Only after the
image block passes should the existing exact-image deployer activate it with
rollback evidence and the grouped browser test. Keep old images/evidence and
`master`; no renewed browser questionnaire or completed HTTP-denial gate is
requested during this preparation block.
