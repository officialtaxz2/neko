# HLS / LL-HLS target-server evidence

Status: **PENDING**. A row is passed only with exact target evidence. Use
`NOT RUN`, `PASS`, `FAIL` or `DEFERRED` and give the reason for every omission.
No credential, cookie, ticket, compact view-only token or public lease ID belongs
in this summary. Raw evidence remains outside Git in the private directory.

- Date / operator:
- Exact implementation commit / clean branch:
- Base and Brave image IDs (from `images.txt`):
- Running image ID, initial health and restart-count baseline:
- HLS modes enabled / native versus MSE paths:
- Exact source profile / host CPU count / tested proxy version:
- Static review record and outstanding dependency advisories:

| Gate | Result | Evidence / limitation |
| --- | --- | --- |
| Exact client tests, type check and build | NOT RUN | |
| Go suite, WebSocket + HLS fuzz and server build | NOT RUN | |
| Base/Brave image builds | NOT RUN | |
| Dependency advisories classified, applicable repairs reviewed | NOT RUN | |
| Default-off public route, WebRTC/WebCodecs and selector behavior | NOT RUN | |
| Enabled public HTTPS invalid-input/header checks | NOT RUN | |
| Direct cleartext denial / actual untrusted-peer spoofing denial | NOT RUN | |
| Valid bootstrap, replay, expiry and scoped-cookie renewal | NOT RUN | |
| Generated init/fragments/playlist independent inspection | NOT RUN | |
| Range / HEAD / gzip / prefix / no-CORS / safe access logs | NOT RUN | |
| View-only full server denial matrix | NOT RUN | |
| Room events, private pause/resume, revoke, replace, shutdown | NOT RUN | |
| WebRTC + WebCodecs recovery/fullscreen regression matrix | NOT RUN | |
| HLS available-device conventional playback | NOT RUN | |
| LL-HLS HTTP/2 or HTTP/3 + path RTT prerequisite | NOT RUN | |
| LL-HLS available-device playback | NOT RUN | |
| Affected Hisense VIDAA/Odin TV, exact model/firmware/browser | NOT RUN | Record availability/version; no compatibility assumption |
| Startup / latency / A/V drift / ten-minute stream | NOT RUN | |
| Mixed-backend slow-viewer and adaptive down/up isolation | NOT RUN | |
| No viewer / 1 viewer / 3 same / 3 variants resources | NOT RUN | |
| Bounded retention, no persistent media writes, idle cleanup | NOT RUN | |
| Rollback and fresh default-off confirmation | NOT RUN | |

## Device measurements

### Common-plane / transport comparison follow-up

Use [the current Workplan handoff](WORKPLAN.md#implemented-block-and-target-handoff--2026-10-07)
for the authorized common-event/WebRTC changes. Historical exact-8741 playback
does not pass these new gates. Preserve technical HLS targets separately from
the user's still-unnumbered preference for small passive delay.

| Additional gate | Result | Evidence / limitation |
|---|---|---|
| Actual Vue/BaseClient/config/partial-peer/writer/legacy regression checks | NOT RUN | |
| Relevant Go race, incoming PLI lifecycle and native URI-drop checks | NOT RUN | |
| Healthy viewers isolated from slow event reader; terminal/revoke ordering | NOT RUN | |
| WebRTC short/expired/failed recovery, Pause/native PiP, no stale reopen | NOT RUN | |
| iPhone Play/muted fallback, fullscreen/clipboard and file transfers | NOT RUN | |
| VIDAA/Odin quiet vs chat/control sequence, sound on/off | NOT RUN | |
| Ten cold/ten warm starts and 30 glass-to-glass samples per eligible path | NOT RUN | |
| Like-quality CPU/RAM/egress/cleanup, per-path recurring use | NOT RUN | |
| Accepted passive delay and keep/restrict/retire decision | NOT RUN | |


For each device and mode record exact model/OS/browser, playback path, at least
ten cold starts, a ten-minute unshaped run and at least 30 synchronized visible
clock samples. Report median/p95/max startup and latency separately. Native
fullscreen/PiP, background return and Safari's required Play gesture need their
own results. Request timings alone do not measure displayed-picture latency.

## Resource and isolation phases

Record snapshots before/after each five-minute phase, process CPU/RSS deltas,
container restart delta, packager/lease/request/subscription counts, retained
bytes, drops and healthy-peer behavior. Apply the fixed limits in `HLS_LL_HLS.md`;
do not silently weaken a failed numeric threshold.

## Findings and disposition

- Confirmed defects / exact repair commit / verification:
- Remaining hypotheses and missing evidence:
- Recorded bounded checkpoint scope (if any):
- Full prototype acceptance: **NOT CLAIMED** until every applicable gate passes.
- `master` promotion: **NOT AUTHORIZED** by this testing checkpoint.
