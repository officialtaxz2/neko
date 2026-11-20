# Upstream sync

The fork merges upstream `m1k1o/neko` `master`; it never overwrites. In a conflict, upstream's
implementation is the standard. Fork code stays only where switching would remove a fork capability
(media delivery and HLS, adaptive quality, view-only, the queued WebSocket writer), layered on top of
upstream's code. Take upstream's tests as the base and add the fork's checks with the same fakes. Each
merge is reviewed in its pull request; this file lists what stays different afterwards. Last merged
upstream commit: `8fd303c` (PR #37, 2026-11-20).

## Deliberate divergences

| Area | Upstream | Fork | Why |
| --- | --- | --- | --- |
| WebSocket writes and teardown (#728) | `peer.mu` around synchronous writes; `Destroy()` sends, then closes | Queued `utils.WebSocketWriter`; `Destroy()` uses `CloseAfterJSON`; `connect()` closes after the writer flushed | A slow client cannot block broadcasts, and `SYSTEM_DISCONNECT` arrives before the close. Upstream's #728 tests pass. |
| Peer teardown (#740, #743) | Destroys the WebRTC peer | Also revokes the session's media delivery in `profileChanged`, `Delete` and `Disconnect` | Deliveries belong to the fork's media layer; covered by `session/delivery_test.go`. |
| `signalRequest` | Creates the peer, then validates the selection | Validates first, destroys a partly built peer | No peer for an empty selection, no leaked peer. |
| Capture subscriptions (#699) | `EncodedStream.Subscribe`, `EncodedMediaSource` | Listener API (`StreamSinkManager`) behind the fork's `EncodedMediaProvider` | WebRTC and HLS consume the provider. See below. |
| Capture config | Pipeline IDs as configured | `nominal_bitrate`, ID derivation that respects it, nil-safe shorthand hook | Adaptive quality. |
| `apps/brave` | Locked-down `policies.json` | Only the restrictions a shared session needs (extensions, downloads, upload dialogs allowed) | Extensions are installed by hand; history and tabs are kept. |
| `apps/brave/supervisord*.conf` | `command=` only | `stopasgroup=true` and `--no-default-browser-check` | `/usr/bin/brave-browser` does not forward the stop signal; without it Brave is killed and loses its tabs. |
| `apps/helium` | does not exist | Fork-only app, also started through `runtime/chromium-unlock` | Keeps Helium's profile (`helium/` subdirectory) free of stale locks. |
| Client | React client | Design in `client/src/fork/`; trackpad, `setInputMode` and `HLSTransport` in `client/core`; other fork code marked "the fork" | See `client/src/fork/README.md`. |

## Capture subscription API (#699)

The fork kept its `capture/manager.go`, `streamselector.go`, `streamsink.go`, `pkg/types/capture.go` and
`webrtc/{manager,peer,track}.go`, and ported upstream's concrete fixes instead: leaky one-frame raw video
queues (#712, `TestVideoConfigGetPipelineQueues`); sample duration from the PTS gap with the
`neko_capture_streamsink_frames` counter (only valid PTS values count, gaps above 10 s keep the buffer
duration); no-copy appsink delivery (`gst_buffer_map`) combined with the fork's PTS/DTS, validity flags
and caps. Bitrate locking, the RTCP PLI goroutine exit and the empty video ID guard were already fixed in
the fork. Every sync will conflict in these files again; moving `capture/media.go` onto `Subscribe` is a
separate task.

## Rollback note

`setup.sh --rollback` to an image built before PR #37 has no `chromium-unlock` and Compose no longer has
the `cleanup` service: after an unclean stop, delete `SingletonLock`, `SingletonSocket` and
`SingletonCookie` from `NEKO_PROFILE_DIR` by hand.
