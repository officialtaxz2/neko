# Adaptive Quality Target-Server Record

Keep raw metrics and logs outside the repository because session IDs and host details can be sensitive. Commit only a reviewed and sanitized result summary when the run is complete.

## Provenance

- Date/time (UTC):
- Operator:
- Server commit:
- Git worktree clean: yes / no
- Container image name:
- Container image ID:
- Host CPU/RAM:
- Docker and Compose versions:
- Browser versions for H1/H2/C:
- Legacy client/protocol applicable: yes / no

## Constraint setup

- Shaper/tool:
- Scope and client IP/interface for C:
- How scope was verified:
- How actual throughput/latency/loss was verified:
- Confirmation that server/SSH and H1/H2 were not shaped:

## Viewer/session mapping

| Viewer | Session ID | Client/device/network |
| --- | --- | --- |
| H1 |  |  |
| H2 |  |  |
| C |  |  |

## Phase results

Use snapshot filenames as evidence references. Drop values are phase deltas, not lifetime counter totals.

| Phase | H1 tier / drop delta | H2 tier / drop delta | C tier / drop delta | C target bit/s | Stream bit/s | Host CPU/RAM | Evidence files | Result/notes |
| --- | --- | --- | --- | ---: | ---: | --- | --- | --- |
| baseline |  |  |  |  |  |  |  |  |
| C 1.3 Mbit/s |  |  |  |  |  |  |  |  |
| C 0.7 Mbit/s |  |  |  |  |  |  |  |  |
| restored |  |  |  |  |  |  |  |  |

## Recovery and regression checks

- [ ] C recovered through `medium` to `high` after impairment removal.
- [ ] H1 and H2 remained on `high` and usable throughout.
- [ ] H1 and H2 gained no peer-local video-drop delta.
- [ ] Unused `medium` and `low` pipelines returned to zero listeners/inactive.
- [ ] Refresh/rejoin and transient interruption of C did not move or stall H1/H2.
- [ ] Brave profile, downloads and managed policy remained available.
- [ ] Login, audio, control, fullscreen and relevant mobile/touch behavior passed.
- [ ] Legacy path repeated successfully, or was recorded as not applicable.
- [ ] No GStreamer error, reconnect storm or sustained host saturation occurred.
- [ ] Estimator and stream rates were plausible and used the same bit/s scale.

## Deviations and tuning

Record each deviation before changing the opt-in YAML. Include the exact old/new value and the evidence that motivated it. Do not change `docker-compose.yaml` or promote the overlay to the default during measurement tuning.

## Verdict

- [ ] Accepted without tuning.
- [ ] Accepted after documented tuning and rerun.
- [ ] Rejected; stable base Compose retained.

Reason:
