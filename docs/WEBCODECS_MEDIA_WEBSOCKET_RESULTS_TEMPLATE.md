# WebCodecs/media-WebSocket target-server record

Keep raw metrics and filtered logs outside the repository because host details and timing can be sensitive. Never record `.env`, passwords, cookies, share tokens, one-time media tickets, ticket-bearing subprotocol headers, full media URLs or event/control payloads. Commit only a reviewed and sanitized summary.

## Provenance

- Date/time (UTC):
- Operator:
- Exact `testing` commit:
- Git worktree clean: yes / no
- Base/adaptive/WebCodecs overlays used:
- Base image ID:
- Brave image ID:
- Validation image ID:
- Host CPU/RAM:
- Docker/Compose versions:
- Public scheme: HTTPS / loopback-only HTTP
- Immediate trusted-proxy topology reviewed: yes / no / not applicable
- Browser/OS/device versions:
- Capture codec/profile/dimensions/frame rate:

## Exact-commit automated checks

- [ ] Client `npm ci`, tests, lint and production build passed in the validation container.
- [ ] Accumulated server package tests passed.
- [ ] `FuzzParseRecord` completed for 30 seconds without a failure.
- [ ] Server/plugin build passed.
- [ ] Local base and Brave images built at the exact commit.

Evidence/notes:

## Default invariance

- [ ] Base deployment without the WebCodecs overlay returned 404 for `/api/media/ws`.
- [ ] Ordinary client sent no prototype request and remained on WebRTC.
- [ ] WebRTC A/V, control/data channel, reconnect and applicable view-only/iOS checks passed.
- [ ] No new persistent volume/state or extra capture pipeline appeared.

Evidence/notes:

## Functional and authorization matrix

| Case | Expected | Evidence/result |
| --- | --- | --- |
| Ordinary opt-in viewer | VP8 video + 48 kHz stereo Opus |  |
| Desktop + phone fullscreen enter/exit | Native where supported; explicit app fallback for canvas otherwise |  |
| Admin opt-in viewer | VP8 video + 48 kHz stereo Opus |  |
| View-only `CanWatch` viewer | Receive succeeds; input stays denied |  |
| Session without `CanWatch` | Denied before upgrade |  |
| Private mode pause/resume | Clean new generation |  |
| Foreground iPhone continuous interval | At least 10 min; no canvas clear, server `client_audio_underflow`, `resync_limit` or retry |  |
| Logout/kick/revocation/replacement | Output stops within 2 s; no retry |  |
| Unsupported codec/browser | Explicit rejection; session survives |  |
| **Use WebRTC** | Unchanged WebRTC A/V and control path |  |

## Ten clean desktop joins

Measure from media-socket open using the same supported Chromium/Brave target and method for every run.

| Join | First video ms | Audible synced A/V ms | Glass-to-glass ms | Absolute A/V skew ms | Notes |
| ---: | ---: | ---: | ---: | ---: | --- |
| 1 |  |  |  |  |  |
| 2 |  |  |  |  |  |
| 3 |  |  |  |  |  |
| 4 |  |  |  |  |  |
| 5 |  |  |  |  |  |
| 6 |  |  |  |  |  |
| 7 |  |  |  |  |  |
| 8 |  |  |  |  |  |
| 9 |  |  |  |  |  |
| 10 |  |  |  |  |  |
| p95 |  |  |  |  |  |

- Same-device WebRTC glass-to-glass p95:
- WebCodecs delta from WebRTC p95:
- Any A/V excursion over 200 ms lasting at least 1 second:

Acceptance: video p95 <= 2000 ms; audible A/V p95 <= 3000 ms; glass-to-glass p95 <= 500 ms and <= 250 ms worse than WebRTC; absolute A/V skew p95 <= 80 ms, with no sustained > 200 ms excursion.

## Slow-client isolation, recovery and adaptive down/up

- Constraint/shaper and exact scope:
- Proof that only W (the WebCodecs viewer) was constrained:
- Healthy WebRTC viewers H1/H2:
- Snapshot/log evidence files:

| Phase | W queue/drop/resync/connection delta | H1/H2 quality and usability | Capture pipelines | Result/notes |
| --- | --- | --- | ---: | --- |
| healthy start |  |  |  |  |
| W constrained/stalled |  |  |  |  |
| W restored/reconnected |  |  |  |  |
| adaptive induced down |  |  |  |  |
| adaptive restored up |  |  |  |  |

- [ ] Provider and egress bounds held.
- [ ] Only W resynchronized or closed; H1/H2 did not stall or switch because of W.
- [ ] Current-generation keyframe resumed within 2 seconds of successful reconnect.
- [ ] No stale frame or parallel reconnect appeared.
- [ ] Fresh one-time tickets and at most 1/2/5/10-second attempts were observed.
- [ ] Logout/kick/revocation suppressed retries.
- [ ] Accepted adaptive three-viewer down/up isolation remained intact.

## Five-minute resource comparison and cleanup

Use the same source and workload. Record deltas, not unrelated host totals.

| Measure | Stable WebRTC baseline | +1 WebCodecs viewer | 60 s after disconnect | Acceptance |
| --- | ---: | ---: | ---: | --- |
| Average process CPU (% of one core) |  |  |  | increase <= 10 points |
| Process RSS MiB |  |  |  | increase <= 128; cleanup within 32 |
| Go goroutines |  |  |  | no repeated-cycle growth |
| Heap allocation MiB |  |  |  | returns to bounded level |
| Capture pipelines |  |  |  | no extra encoder/pipeline |
| Manager deliveries/subscriptions |  |  |  | returns to baseline |
| Payload and calculated envelope overhead |  |  | n/a | within 10% |

Repeated connect/disconnect count:

## Malformed and hostile input

| Case group | Expected/result | Offender isolated | Logs credential/payload-safe | No restart/leak |
| --- | --- | --- | --- | --- |
| Query/origin/secure-transport boundary |  |  |  |  |
| Missing/wrong protocol/version |  |  |  |  |
| Malformed/unknown/replayed/expired ticket |  |  |  |  |
| Flags/reserved/type/length/size parser cases |  |  |  |  |
| Invalid/control-rate JSON and binary input |  |  |  |  |

## Rollback

- [ ] Removed `media=webcodecs-ws` from clients.
- [ ] Removed the WebCodecs overlay and recreated only the Neko service.
- [ ] `/api/media/ws` returned 404 again.
- [ ] Normal WebRTC A/V and control passed.
- [ ] No database, volume or capture-format rollback was needed.

## Deviations and verdict

Record each deviation before changing an opt-in value. Never change the stable base Compose merely to pass this matrix.

- [ ] Accepted for the documented target/browser scope.
- [ ] Rejected; stable WebRTC deployment retained.
- [ ] Incomplete; no acceptance claim.

Reason and remaining limitations:
