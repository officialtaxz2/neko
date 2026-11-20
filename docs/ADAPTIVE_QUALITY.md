# Adaptive quality

Each WebRTC viewer gets its own VP8 tier, chosen by a per-viewer estimator in
`server/internal/webrtc/peer.go`. `docker-compose.adaptive.yaml` mounts `deploy/adaptive-quality.yaml`
as `NEKO_CONFIG`; `.env.example` enables it through `COMPOSE_FILE`, and the HLS overlay needs it. Without
the overlay Neko runs upstream's single `main` pipeline.

## Tiers

| ID | Size at 1920×1080 | fps | `nominal_bitrate` (bit/s) |
| --- | ---: | ---: | ---: |
| `high` | 1920×1080 | 25 | 3,993,600 |
| `medium` | 1280×720 | 25 | 1,601,600 |
| `low` | 1280×720 | 20 | 748,800 |
| `lowest` | 960×540 | 15 | 332,800 |

- The first ID is the start tier. A tier's encoder runs only while someone watches it, so CPU grows with
  the number of tiers in use (up to four encoders with four threads each).
- `nominal_bitrate` equals the encoder target and is the reference for stepping up, so an easy-to-encode
  picture on the current tier cannot make the next tier look affordable.
- Lower tiers scale bilinearly (`videoscale method=1`); nearest-neighbour made thin text vanish.
- Encoder: CBR, full quantizer range (`max-quantizer: 63`), fixed speed (`cpu-used: -4` on `high`, `-6`
  below; a positive value lets libvpx speed up under CPU contention and lose 3–7 dB),
  `static-threshold: 100` for screen content.
- Keyframes come on demand: when a tier gains a viewer, on PLI/FIR and after a locally dropped frame, at
  most one per 500 ms per tier; otherwise every 4 seconds (`keyframe-max-dist`). `max-intra-bitrate` caps
  a keyframe at about 0.36 s of the tier's rate.

## Decision rules

The estimator reads Pion's GCC target every `read_interval` (2 s) and compares it with the viewer's
complete-delivery reference: `max(measured, nominal video) + measured audio`, times `1 + transport_reserve`
(0.05).

- **Down one tier** only when all hold: the target is more than `diff_threshold` (0.15) below the
  reference, the receiver confirms congestion, the trend is downward or this has lasted
  `stalled_duration` (8 s) without one, the shortage has lasted `unstable_duration` (6 s), and
  `downgrade_backoff` (30 s) has passed since the last downgrade.
- **Congestion** is confirmed only by loss the receiver still reports after retransmission: per RTCP
  report interval, lost packets of the video SSRC divided by expected packets, at least
  `congestion_loss_threshold` (0.02) and at least four packets. NACKed packets that arrived on
  retransmission do not count; reports older than two read intervals do not count. An interval whose
  highest sequence number did not advance keeps the previous verdict.
- **Up one tier** after a clean `stable_duration` (12 s) when the target exceeds the next tier's reference
  by `upgrade_diff_threshold` (0.15) and `upgrade_backoff` (5 s) has passed.
- **Recovery probe:** a viewer that was downgraded automatically may try the tier it lost once the lower
  tier was clean for the stable window and `recovery_probe_interval` (30 s) has passed, whatever the target
  says (Pion's loss controller can pin it at 100,000 bit/s). Every step back toward a lost tier is
  validated like a probe: a failed step doubles the wait, up to `recovery_probe_max_backoff` (2 min).

All state is per viewer. `congestion_loss_threshold: 0` restores the old any-loss-or-NACK rule.

## Diagnosis

The estimator logs at debug level (`debug: true` in the profile):

```bash
docker compose logs -f neko | grep -E "got bitrate from estimator|downgraded video stream|upgraded video stream|recovery probe|set video"
```

On each `got bitrate from estimator` line, compare `receiver_loss_ratio` (counts) with
`receiver_nack_ratio` (logged only): a high NACK share with zero loss is a path that loses first
transmissions but repairs them, and must not cost a tier. Further fields: `target_bitrate`, the reference
and floor, `insufficient`, the trend, `receiver_interval_unjudged` and `receiver_congestion_confirmed`.
A probe log with `capacity_qualified=true` is a step the target alone would have allowed.

Metrics (`/metrics`, `server.metrics: true`): `neko_capture_streamsink_bitrate{video_id}`,
`neko_webrtc_video_listeners` (current tier per session), `neko_webrtc_receiver_estimated_target_bitrate`,
`neko_webrtc_receiver_congestion_evidence`, `neko_webrtc_recovery_probe_*`,
`neko_webrtc_track_dropped_samples_total`, `neko_capture_streamsink_keyframes` and `_keyframe_requests`.
Above 30 forced keyframes per minute on one tier, raise `minKeyframeRequestInterval` to one second.

## Open checks on the server

- Since #32: the home desktop stays on `high` for at least ten minutes of full-screen motion while
  `receiver_loss_ratio` stays below 0.02, whatever the NACK share and target show. Every downgrade follows
  reads with `receiver_congestion_confirmed=true`.
- A constrained second viewer (about 0.7 Mbit/s) still leaves `high` and returns within three minutes
  after the limit is lifted, while a healthy viewer never changes tier.
- CPU and frame rate (`neko_capture_streamsink_frames`) with several tiers active; HLS output for
  artifacts, since its decoder can lose a VP8 frame silently until the next keyframe.
- A path whose losses NACK repairs too late for playout freezes without counting as congestion. Compare
  `receiver_loss_ratio` with the browser's freeze count in `chrome://webrtc-internals`.

## Rollback

- Encoder settings: `cpu-used: "4"` without `static-threshold`, or `keyframe-max-dist` at one second per
  tier without `max-intra-bitrate`.
- Image: `bash setup.sh --rollback`.
- Whole feature: set `COMPOSE_FILE=docker-compose.yaml` in `.env` (this also turns HLS off), run
  `docker compose down`, then `bash setup.sh --no-build`.
