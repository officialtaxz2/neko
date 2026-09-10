# Bounded iOS Transient-Network Recovery

This document describes the client recovery path implemented on `testing` and the exact target-server procedure needed before making a no-reload iOS recovery claim. Runtime/build/device checks are not executed in Codex.

## Implemented recovery boundary

The client treats connection recovery and autoplay permission as separate state machines:

1. **Existing-peer recovery:** an ICE `disconnected` state keeps the current WebSocket, peer and server session alive for eight seconds. If ICE returns to `connected`/`completed`, the timer is cancelled and no new login is created.
2. **Application-level reconnect:** if the WebSocket closes, ICE becomes `failed`/`closed`, or the eight-second ICE window expires, the old socket, peer, data channel, heartbeat, timers, candidates and media references are cleared. A session that had previously connected may then retry with its in-memory login values after 1, 2, 5 and 10 seconds. The four-attempt sequence is owned by one timer and cannot run in parallel.
3. **Safari Play fallback:** once a recovered or replacement peer supplies media, the existing autoplay path first tries normal playback and then muted playback. If Safari still requires user activation, the central Play control remains visible. A required Play tap is a media-policy outcome, not a failed network reconnect.

Initial login failures do not start the retry sequence. Explicit logout, demo mode and every server-directed `system/disconnect` suppress it, preserving authentication and session intent. After exhaustion the populated login form remains available for a manual retry without reloading the page.

Stale WebSocket, peer, data-channel and stream callbacks are identity-checked or detached so a delayed event from the replaced connection cannot tear down or populate the current one. Media-element `srcObject` recovery remains bounded and never calls `video.load()`.

## Static verification boundary

The extracted retry decision and delay schedule have dependency-free Node tests in `client/tests/recovery.test.mjs`. Run all commands below only on the real target server:

```bash
git switch testing
git pull --ff-only origin testing
git status --short --branch
git rev-parse HEAD

cd client
npm ci
npm test
npm run lint
npm run build
cd ..
```

Rebuild and recreate the deployment through the operator's normal `testing` procedure. For the tracked Brave deployment:

```bash
docker image tag my-neko/brave:latest my-neko/brave:pre-ios-recovery
./build my-neko/base:latest -y
./build my-neko/brave:latest -y
docker compose config --quiet
docker compose up -d --force-recreate
docker compose ps
```

If the adaptive overlay is part of the grouped checkpoint, use both Compose files consistently instead. Do not print the resolved Compose model because it contains password values.

## Exact iPhone interruption procedure

### Record before starting

- exact `testing` commit and running image ID;
- iPhone model, iOS version and Safari version;
- whether Safari Web Inspector is available;
- whether the base or adaptive Compose profile is active;
- healthy desktop (`H1`), healthy iPad/second viewer (`H2`) and iPhone (`C`) session IDs;
- initial muted/playing state and whether an initial Play tap was required.

Keep changing desktop content visible on all three viewers. If the adaptive overlay is active, snapshot `video_listeners`, connection state and peer-local drop counters before and after every phase. Keep server logs open and, when Web Inspector is available, retain client messages containing:

```text
peer ice connection state changed
ICE recovery timeout
scheduling application reconnect attempt
starting application reconnect attempt
connected
Autoplay blocked
```

### Phase A — existing peer recovers

1. Join `H1`, `H2` and then `C`; wait 60 seconds with video and audio working.
2. Enable flight mode on `C` for five seconds, then disable it without reloading or touching the page.
3. Observe for 60 seconds.
4. Confirm that `C` returned on the same server session ID and that the client showed ICE `disconnected` followed by `connected`/`completed` without an application-reconnect attempt.
5. Confirm `H1` and `H2` never reconnected, stalled or changed tier. If Safari shows Play after media returns, record and tap it once; do not classify that policy tap as an application reconnect.

### Phase B — replacement session recovers without reload

1. With all viewers healthy again, enable flight mode on `C` for 15 seconds. If this device does not cross the ICE/WebSocket failure boundary in 15 seconds, extend only this interruption until the old peer is demonstrably closed or the ICE recovery timeout fires, and record the actual duration.
2. Disable flight mode. Do not reload Safari, navigate, resubmit the login form or touch Play yet.
3. Observe for up to 90 seconds. Confirm that exactly one reconnect attempt is in flight at a time and no more than four attempts start.
4. Confirm a replacement session becomes connected and supplies video/audio. A new session ID is expected for this application-level login.
5. If the central Play control appears only after media is available, record the autoplay-policy result and tap it once. Acceptance permits this standards-required gesture, but not a page reload or manual login.
6. Confirm `H1` and `H2` remained usable and did not reconnect. Under the adaptive profile, both must remain on `high` and gain no peer-local video-drop delta.

### Phase C — bounded exhaustion and manual recovery

1. Start from a healthy `C`, then keep it offline long enough for all four application retries to finish. Because each WebSocket/peer attempt may itself wait for the 15-second connection timeout, allow up to 90 seconds before declaring the sequence unbounded.
2. Confirm the client logged no more than four `starting application reconnect attempt` messages and then stopped retrying.
3. Restore connectivity. Confirm the populated login form can reconnect manually without a reload.
4. From a healthy session, issue an admin kick or another intentional `system/disconnect`. Confirm that no automatic reconnect attempt follows. This is required evidence that recovery does not override server authorization/session decisions.

## Acceptance criteria

The block can be described as target-server verified only when all of the following are recorded:

- Phase A recovers the same peer/session without reload.
- Phase B creates at most one connection attempt at a time and restores media through a replacement session without reload or manual login.
- No more than four application attempts occur in Phase C, and explicit server disconnect does not retry.
- A central Play tap, if required, occurs only after media is present and is reported separately from network recovery.
- Old session IDs disappear or become inactive; no stale callback tears down the replacement session.
- `H1` and `H2` remain connected and usable throughout; adaptive runs also keep them on `high` with zero peer-local video-drop deltas.
- Touch, trackpad, mobile keyboard/helper, orientation/fullscreen, mute/unmute, audio and control semantics still work after recovery.
- No page reload is used in Phases A or B.

Until this procedure passes on the target iPhone, the repository status is **implemented and statically reviewed, target-server no-reload behavior pending**.

## Rollback

Return to the pre-change `testing` commit or the preserved image tag, recreate the same Compose profile, and repeat login/playback/control smoke checks. The recovery change has no server configuration or persistent-data migration.

```bash
NEKO_IMAGE=my-neko/brave:pre-ios-recovery docker compose up -d --force-recreate
```
