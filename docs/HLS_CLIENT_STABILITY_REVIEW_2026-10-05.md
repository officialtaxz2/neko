# HLS client startup and pause/recovery review — 2026-10-05

Status: **IMPLEMENTED / STATICALLY REVIEWED / NOT EXECUTED IN CODEX**.
The new client changes and regression cases have no target-server result yet.
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
All new tests, TypeScript checks, builds and browser acceptance are **PENDING**.

## One grouped target checkpoint later

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
