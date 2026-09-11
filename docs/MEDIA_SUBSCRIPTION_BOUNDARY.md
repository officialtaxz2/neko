# Backend-neutral encoded-media subscription boundary

Status: **design complete on `testing` as of 2026-09-11; no alternative media backend is implemented**.

This document fixes the architecture contract for the next implementation block. It is intentionally more concrete than a product direction, but it does not claim that WebCodecs/WebSocket, HLS/LL-HLS, DASH or WebTransport exists in the repository.

## Scope

The boundary must let more than one receive-media backend consume the existing shared Neko desktop without duplicating the room, weakening authorization or allowing one slow consumer to block another.

It covers:

- discovery and selection of encoded audio/video sources;
- demand-driven, bounded subscriptions to those sources;
- media timing, format changes, keyframes and discontinuities;
- participant-scoped delivery lifecycle and revocation;
- backend registration, capability reporting and observability;
- a compatibility migration of the existing WebRTC sender.

It deliberately does not define a final WebCodecs wire format, HLS segment duration, automatic fallback algorithm, Smart-TV support matrix or new control transport. Those belong to later prototypes and measurements.

## Current code seam

The implemented receive path is:

```text
GStreamer appsink
    -> types.Sample
    -> capture.StreamSinkManagerCtx
    -> types.SampleListener
    -> webrtc.Track bounded channel
    -> Pion TrackLocalStaticSample
    -> browser RTCPeerConnection
```

The useful existing properties are:

- `server/internal/capture/streamsink.go` starts an encoder on its first listener and stops it after the last listener;
- video listeners wait in a keyframe lobby before receiving samples;
- `StreamSelectorManagerCtx` resolves the ordered video variants;
- capture fan-out releases its listener lock before dispatch;
- each WebRTC `Track` owns a bounded two-sample queue and drops locally when full;
- current stream bitrate, nominal bitrate, listener count and peer-local drops are observable.

The coupling that must be removed is:

- `types.StreamSinkManager` and `types.StreamSelector` are exposed directly to WebRTC;
- codec metadata is represented by `codec.RTPCodec`, which contains Pion/RTP-specific fields;
- the WebRTC track owns subscription, pause, switching, queueing and drop semantics at once;
- listener identity relies on reflected pointers and cross-stream movement relies on the concrete capture implementation;
- `types.Sample.Timestamp` is Go arrival time. The GStreamer bridge exports duration and delta/keyframe state, but not buffer PTS, DTS, caps or a source generation;
- `Session.SetWebRTCPeer` and `SetWebRTCConnected` make the generic `IsWatching` state backend-specific.

The existing broadcast and screencast managers are not the new abstraction. They build separate purpose-specific capture pipelines and do not provide participant-scoped authorization, source selection or non-blocking subscriber isolation.

## Non-negotiable invariants

1. There remains one shared server-side desktop and one logical room.
2. Authentication, host/control authority and plugin permissions remain server-side and independent of media transport.
3. `CanWatch` authorizes receiving media. `IsViewOnly` removes interaction but does not create a separate room or capture.
4. No media backend may grant control, clipboard, file, chat-send, admin or media-publish capability.
5. Capture fan-out and every subscriber queue are bounded and non-blocking.
6. A backend or participant failure closes only its own delivery/subscription unless the shared encoder itself fails.
7. WebRTC remains the default and must preserve its current protocol, queue/drop behavior, adaptive selection and metrics during the first refactor.
8. Alternative backends remain opt-in until their own target-device evidence exists.

## Chosen shape

Two separate lifecycle layers are required:

```text
                                  authenticated Session
                                           |
                                           | CanWatch policy + scoped lease
                                           v
GStreamer -> EncodedMediaProvider -> MediaDeliveryManager -> registered backend
                  |                         |                    |
                  | source subscription     |                    +-- WebRTC per viewer
                  v                         |                    +-- WebCodecs/WS per viewer
        encoded audio/video events          |                    `-- HLS packager shared,
                                            |                        viewer leases separate
                                            v
                                  generic watching state
```

### Source subscription

A source subscription connects a delivery backend to encoded capture output. It is not itself a user session and it carries no login credential.

This distinction permits:

- one WebRTC or WebSocket source subscription per participant;
- one HLS packaging subscription per active variant, shared by multiple authorized playlist viewers;
- capture demand to be counted accurately without pretending every HTTP segment request is a new encoder listener.

### Participant delivery

A participant delivery binds an authenticated Neko session to one registered backend. Only the delivery manager can create it. The manager checks `session.Profile().CanWatch`, creates a scoped lease and owns revocation and generic watching state.

The backend receives only the minimum media grant and opaque identifiers it needs. It must not receive the member/admin password, the view-only share credential or authority to mutate the participant profile.

## Backend-neutral data contract

The first implementation should place transport-independent contracts in `server/pkg/types/media.go`. Exact Go naming may change only if required to avoid an import cycle; the semantics below are fixed.

```go
type MediaKind string // audio or video

type MediaCodec struct {
    Name       string
    MIMEType   string
    ClockRate  uint32
    Channels   uint16
    Parameters map[string]string
    Config     []byte
}

type MediaSource struct {
    ID                   string
    Kind                 MediaKind
    Codec                MediaCodec
    Width                uint32
    Height               uint32
    FrameRateNumerator   uint32
    FrameRateDenominator uint32
    NominalBitrate       uint64
}

type EncodedMediaUnit struct {
    Generation uint64
    Sequence   uint64
    PTS        time.Duration
    DTS        time.Duration
    DTSValid   bool
    Duration   time.Duration
    Keyframe   bool
    Data       []byte
}

type MediaEvent struct {
    Type          MediaEventType // format, unit, discontinuity, end
    Source        MediaSource
    Unit          EncodedMediaUnit
    Discontinuity MediaDiscontinuity
}
```

`MediaCodec` must not embed Pion or WebRTC types. A WebRTC adapter may map it to `codec.RTPCodec`; an HLS packager may map the same description to its muxer.

`Config` represents codec initialization data when the format needs it. It must be immutable and must never contain credentials. Width, height and frame rate may be zero for audio or when not yet known, but a `format` event must update them before a backend that requires those fields starts delivery.

`Data` is immutable after publication. A consumer that mutates or retains a buffer outside its subscription event lifetime must copy it. The implementation must document and test the selected ownership rule; silent mutation of a shared sample is forbidden.

## Timing and discontinuity contract

Arrival time from `time.Now()` is not a valid cross-backend presentation timestamp. The GStreamer bridge must expose buffer PTS, DTS and duration, and the provider must normalize them to a manager-owned monotonic timeline.

The contract is:

- `Generation` changes whenever a pipeline is recreated, caps/codec change, screen resize invalidates the stream, or timestamps reset;
- `Sequence` increases within one source generation;
- PTS/DTS use one declared time unit; valid DTS is monotonic within a generation, while PTS may reorder only when the codec requires presentation reordering;
- `DTSValid == false` explicitly represents codecs or stages without a meaningful decode timestamp;
- audio and video expose enough common timeline information for a muxer/player to synchronize them;
- a subscriber first receives `format`, then video begins on a keyframe;
- overflow, source restart and format change produce an explicit discontinuity rather than an unexplained timestamp jump;
- end-of-stream and subscription close are distinct events.

If the initial refactor cannot prove common audio/video timestamp normalization, it may preserve current WebRTC delivery behind a compatibility adapter, but no HLS or WebCodecs prototype may be declared ready until the timing requirement is implemented and tested.

## Provider and subscription contract

The provider owns source discovery, capture demand, keyframe admission and switching:

```go
type EncodedMediaProvider interface {
    Sources(kind MediaKind) []MediaSource
    Subscribe(context.Context, SourceSubscriptionRequest) (MediaSubscription, error)
}

type MediaSubscription interface {
    ID() string
    Source() MediaSource
    Events() <-chan MediaEvent
    Switch(context.Context, MediaSelector) error
    SetPaused(bool) error
    Close() error
}
```

Required behavior:

- `Subscribe` resolves one source and starts its pipeline before reporting success;
- video admission waits for a keyframe;
- `Switch` resolves and starts the target before detaching the source, emits a discontinuity/format transition and resumes video on a target keyframe;
- `SetPaused(true)` removes capture demand; resume restores demand and keyframe admission;
- `Close` is idempotent, releases demand exactly once and closes the event channel after its final event;
- direct access to `MoveListenerTo` is hidden behind this contract;
- callers cannot request an unbounded queue.

Selectors retain the useful current meanings (`exact`, `nearest`, `lower`, `higher`) but move out of the WebRTC contract. Ordered video IDs and nominal bitrates remain provider data.

## Backpressure and isolation

Every source subscription has a bounded manager-owned queue. Dispatch into that queue never waits for a consumer.

Queue capacity and overflow policy are trusted backend configuration, not client input. The initial WebRTC adapter must keep its effective capacity of two samples and current drop-new behavior so the refactor does not silently change accepted behavior.

That means the migrated WebRTC path has one two-unit subscription queue rather than stacking a new queue in front of the existing two-sample `Track.sample` channel. The adapter consumes the subscription queue directly when writing to Pion.

Later backends may choose only defined bounded policies:

- drop the newest unit;
- discard video until the next keyframe and emit a discontinuity;
- close/restart a failed packaging subscription.

An HLS packager must never build an unbounded backlog. If it misses data needed for a valid segment, it ends that segment generation, emits the appropriate playlist discontinuity and resumes from a keyframe. Reliable ordered WebSocket delivery must likewise discard/resynchronize instead of converting receiver slowness into ever-growing latency.

Format, discontinuity and end events are not ordinary droppable media units. On overflow the manager must clear or resynchronize sample backlog so the lifecycle event is delivered, or close the subscription with an explicit error; it must never silently lose a format generation change.

## Delivery manager and backend contract

The delivery manager is the only participant-facing entry point:

```go
type MediaDeliveryManager interface {
    Backends() []MediaBackendDescriptor
    Open(context.Context, Session, MediaDeliveryRequest) (MediaDelivery, error)
    CloseSession(sessionID string)
}

type MediaBackend interface {
    Name() string
    Capabilities() MediaBackendCapabilities
    Open(context.Context, MediaLease, MediaDeliveryRequest) (MediaDelivery, error)
}
```

The manager must:

1. authenticate the existing session and reject `CanWatch == false`;
2. resolve an enabled backend and intersect requested media with server capabilities;
3. create a short-lived, backend-scoped lease without exposing the login/share credential;
4. open at most one primary receive delivery per session in the first implementation;
5. update generic watching state independently from WebRTC connection state;
6. revoke the lease and close delivery on logout, session deletion, explicit stop, backend failure or server shutdown.

During server shutdown, participant deliveries and backend subscriptions close before the shared capture provider. This preserves an explicit final event/error path and prevents consumers from outliving their source.

Private-mode pause must be represented above individual backends so all transports stop receiving consistently.

Backend capabilities describe transport mechanics only, such as receive audio/video, server-side variant switching, native adaptive playback and media publishing support. They never override member authorization.

## Session and security boundary

The generic session contract needs backend-neutral delivery attachment/state methods. Existing WebRTC methods remain compatibility shims until all current handlers have migrated.

`IsConnected` continues to describe the event/session connection. `IsWatching` becomes true only while an authorized primary media delivery is active. For connectionless HTTP playback, the HLS backend must use an expiring viewer lease/heartbeat rather than treating every segment GET as a permanent connected peer.

For HTTP streaming, a native player may not support custom authorization headers. A later HLS prototype may therefore need an opaque token in a playlist path. If so, it must be a new short-lived delivery lease that is:

- scoped to one session, backend and receive-only operation;
- unrelated to and unable to replace the member/admin/view-only login credential;
- redacted from access logs and metrics;
- revoked with the session and invalid after service recreation;
- never accepted by room, control, plugin or media-publish APIs.

The compact view-only fragment remains only a login bootstrap. It must not become a permanent segment credential.

## Current-to-target mapping

| Current artifact | Target responsibility |
| --- | --- |
| `types.Sample` | compatibility input mapped to `EncodedMediaUnit` with real timing/generation |
| `types.SampleListener` | replaced internally by a bounded `MediaSubscription` event queue |
| `StreamSinkManagerCtx` | encoded source implementation behind `EncodedMediaProvider` |
| `StreamSelectorManagerCtx` | source catalog/selector implementation |
| `webrtc.Track` | WebRTC delivery adapter consuming a media subscription |
| `WebRTCManagerCtx.CreatePeer` | WebRTC signaling/control plus an opened media delivery |
| `Session.SetWebRTCConnected` | compatibility shim over generic primary-delivery watching state |
| broadcast/screencast managers | remain separate until a later explicit convergence decision |

## Observability contract

No metric or log may contain a credential or delivery URL token. The implementation must expose at least:

- active source subscriptions by backend, source ID and media kind;
- active participant deliveries by backend and state;
- queue depth/capacity where practical;
- delivered units/bytes;
- dropped units by backend, kind, source and reason;
- discontinuity/resynchronization counts;
- backend open/close/error counts;
- source generation and pipeline start/stop counts.

Session IDs may remain in debug logs and the existing bounded per-session metrics where already accepted, but they must not be added to low-value high-cardinality metrics without a concrete diagnostic need.

## First implementation block: compatibility refactor

The next code block must change structure without adding another client-visible transport:

1. Add the pure media descriptors, events, provider/subscription and delivery interfaces.
2. Extend the GStreamer bridge to expose the timing/format data needed by the contract while preserving the current sample data path.
3. Implement a capture-backed provider with demand-driven start/stop, keyframe admission, bounded queues, switch/pause/close and discontinuity handling.
4. Add a delivery manager/registry and generic session watching lifecycle.
5. Convert the existing WebRTC sender into the first backend adapter. Preserve current SDP/signaling, data channels, queue capacity/drop behavior, estimator selection, private-mode pause and metrics.
6. Keep every existing config key, HTTP/WebSocket event and default deployment behavior unchanged.
7. Do not add WebCodecs, a media WebSocket endpoint, HLS routes, an HLS packager or automatic backend selection in this block.

Required focused tests include:

- source discovery and selector ordering;
- first/last subscription pipeline demand;
- keyframe-gated video start;
- switch, pause, resume and idempotent close;
- generation/format/discontinuity ordering;
- bounded overflow that never blocks another subscription;
- unchanged WebRTC effective queue/drop and adaptive-selection behavior;
- `CanWatch == false` denial and view-only receive allowance;
- delivery/session revocation and shutdown cleanup;
- absence of tokens in logs/metric labels by construction.

Runtime/build verification remains target-server work under `AGENTS.md`. The compatibility refactor is not complete until focused Go tests, server/plugin build, local image build and the existing ordinary/admin/view-only/adaptive regression smoke tests pass there.

## Prototype gates after the refactor

### WebCodecs plus dedicated media WebSocket

The first interactive fallback candidate may reuse VP8 where the client reports a supported `VideoDecoder` configuration. It still requires:

- an explicitly versioned binary frame protocol carrying format, generation, PTS/DTS, duration, keyframe and payload length;
- audio as well as video, with capability probing rather than codec assumptions;
- a bounded server and browser queue with drop-to-keyframe/resync behavior;
- independent control transport work, because the current mouse/keyboard path is a WebRTC data channel;
- measured latency and loss behavior. WebSocket is not presumed better on a constrained path merely because it avoids ICE.

### HLS / Low-Latency HLS

The first passive prototype must:

- use the same delivery authorization and room session;
- share one packager output per active variant rather than one encoder per viewer;
- begin with an opt-in codec set supported by actual target devices; current VP8/Opus must not be assumed to provide native HLS compatibility;
- define segment/part retention, discontinuity and cleanup limits;
- prove that slow HTTP clients cannot block capture or healthy WebRTC viewers;
- measure startup time, end-to-end latency, CPU, memory and storage/I/O before choosing HLS versus LL-HLS defaults.

## Fixed decisions and remaining open choices

Fixed by this design:

- the boundary is encoded media, not a second desktop capture;
- source subscriptions and participant deliveries are different lifecycles;
- authorization is centralized above media backends;
- timing, format, generation and discontinuity are first-class;
- queues are bounded and non-blocking;
- WebRTC remains the default during migration;
- backend selection is explicit/opt-in before any automatic fallback;
- WebCodecs/WebSocket is evaluated for interactive compatibility, HLS/LL-HLS separately for passive compatibility, and WebTransport only afterward.

Still open for later evidence-led prototype blocks:

- exact WebCodecs/WebSocket framing and client queue sizes;
- codec combinations on the real desktop, iPhone, iPad and target Smart-TV browsers;
- HLS versus LL-HLS segment/part durations and latency budget;
- whether DASH materially expands the actual device matrix;
- explicit capability negotiation and eventual automatic fallback rules;
- whether broadcast/screencast should later consume the same provider.

## Sources

- Current repository code in `server/pkg/types/capture.go`, `server/pkg/gst/`, `server/internal/capture/`, `server/internal/webrtc/`, `server/internal/session/` and `server/cmd/serve.go`.
- Upstream [v3 rewrite issue #371](https://github.com/m1k1o/neko/issues/371), which separates connection, media streaming and control and requires one interface between media backends and the rest of the system.
- Upstream [alternative-media issue #690](https://github.com/m1k1o/neko/issues/690), whose proposed order starts with encoded-media subscriptions before WebCodecs/WebSocket and WebTransport.
