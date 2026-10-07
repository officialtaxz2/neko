# Stability review and outstanding device evidence

## Authorized improvement block — 2026-10-07

The operator subsequently authorized the comparative plan. Independent B1–B3
and B6 source corrections and B4/B9 handoff assets are implemented and statically
reviewed. **Tests, builds and device/runtime checks: NOT EXECUTED IN CODEX.**
The target is now the healthy exact-8d conventional-HLS-enabled candidate; preserve the earlier
accepted exact-8741 HLS image/evidence. Historical passes below do not validate these
new common-event/WebRTC changes.

**Supplied exact-f03 preparation:** Prepare-Exitcode 1, private evidence
`/opt/docker/nekoNew/neko-stability-f03bc4bcf68b-20261007T164845Z`. All 80 client
tests/type/build and the selected normal Go suite passed; race checks passed
for utils/config/legacy/event/handler/WebRTC. The capture race invocation failed
an event-order assertion, not a supplied data-race diagnostic. Base/Brave builds
and the final retained-live check were not reached; no activation was issued.
The candidate is not deployable on that evidence.

Static inspection confirms an older publication-window defect in the unchanged
capture provider: a received initial format was not yet marked handed off when
a consumer triggered the next generation, allowing its discontinuity to be
coalesced away. The narrow repair marks the selected immutable format under the
dequeue mutex before sending. It preserves unselected-format coalescing and
changes no payload, media clocks, encoder topology or queue limits. HLS already
handles the cold same-generation first-caps transition. Deterministic fixtures
cover both publication boundaries; the scoped target gate requires the old
failure and 100 repaired subscription repetitions under race, then repeats
backend/race/server/images. Supplied exact-8d preparation subsequently passed
with Prepare-Exitcode 0: the helper reports all those gates passed and live
service retained; Base/Brave exports and the private snapshot are visible.
The excerpt starts inside the base build, so earlier individual assertions/counts
are not separately visible. Unchanged client evidence alone is inherited from
f03. Evidence: `/opt/docker/nekoNew/neko-stability-8d8126c92903-20261007T171646Z`.
Subsequent exact-8d HLS-off activation passed with Baseline-Exitcode 0: healthy
`my-neko/brave:hls-8d8126c92903`, private baseline snapshot and 2/2 disabled
bootstrap/media 404 probes. The operator reports general PC operation works;
duration and individual action/outage/Pause results are not supplied. Hisense
VIDAA/Odin is unavailable because the TV belongs to a friend. Its device and
event-cause gate remains deferred/unverified; general PC success does not close
it. Subsequent conventional-HLS activation passed with Enable-Exitcode 0,
healthy exact-8d service/private enable snapshot and all 17 public plus 2 direct-
cleartext denial probes, including required headers. Valid-credential leases,
packaging and rendered A/V are separate from these synthetic boundary checks.
The operator subsequently explicitly confirmed the requested at-least-ten-minute
PC HLS moving-picture/audio interval with a concurrent working WebRTC tab,
without observed freeze, Retry or reload. This closes that bounded sustained-
playback/coexistence gate at exact-8d. It is supplied browser observation, not a
captured trace, scripted event/outage/role matrix, numerical resource/latency
comparison or iPhone/TV acceptance. Do not repeat the passed interval or server
gates. NEXT is the read-only B4 resource baseline/comparison and remaining
available-device/grouped coverage; unavailable VIDAA stays explicitly open.
This is no demonstrated cause of VIDAA failure or
of the earlier HLS freeze, and no performance advantage is claimed.

- **F1 / B2:** a bounded FIFO writer now owns each normal event socket, including
  both legacy bridge legs; 128 records/16 MiB including in-flight writes, 5 s
  operation/terminal-flush ceilings. Overflow closes its connection; no room or
  security event is silently discarded while retaining a supposedly healthy
  peer. Revocation precedes flush. JSON encoding remains synchronous and queue
  ceilings are per socket, not total service memory. Reader handoffs can cancel;
  local JSON API/handshake operations are 15 s and cleanup is separately 5 s.
  Streaming file bodies retain cancellation without an API total-duration cap.
- **F2/F3 / B3:** initial ICE checking is not success; startup stays bounded at
  15 s. Client/server transient grace is 8 s, without extension by rechecking;
  failed/revoked peers close immediately. Non-trickle gathering is bounded and
  cancellable before peer locking. Vue methods own actual playback state, with
  one foreground progress/reattachment owner and the existing three-attempt
  ceiling. Pause/native PiP intent, autoplay/Play fallback, seek/metadata,
  dropped-frame evidence, legacy undefined play returns and stale promises are
  handled separately. Track removal uses that same owner instead of another
  timer; late WebCodecs audio resumes cannot mutate a paused/replaced player.
  Browser counters remain proxies, not proof of visible
  moving pictures/audible sound or a universal blackscreen remedy.
- **F4 / B1:** inbound PLI loops exit on track/peer closure; missing pipeline
  IDs derive deterministically, nominal-rate order only when every rate is
  known, otherwise explicit ladder order is required. Empty default selection
  is rejected before peer creation and partially initialized peers close.
  Pipeline-string shorthand and native GFile/URI/CString ownership are repaired.
  One mutex now covers bitrate buckets and their reset; no subscription API or
  native sample/clock topology was replaced.
- **F8/F9 / B6:** remove unused Vue CLI configuration, use the Vite development
  launcher while retaining its legacy port input, remove the special
  four-second development-host timeout,
  honor configured ICE server lists and correct unproven room-event cause
  comments. Re-delivery of the same track object no longer stops that object.
  Safari uses the existing manual clipboard fallback. Dependency remediation
  remains separately scoped and open.

New target-only regressions cover real Vue class binding/lifecycle, actual
BaseClient timers and stale peers, seek/Pause/autoplay/progress, writer FIFO,
buffer ownership/overflow/blocked-peer isolation, legacy cancellation and
configuration/partial-peer cleanup. See [the exact target handoff](WORKPLAN.md#implemented-block-and-target-handoff--2026-10-07).
New live/device outcomes and performance improvement are **unmeasured**; supplied
f03 automated results and their capture failure are recorded above.

Required devices are desktop, Smart-TV and iPhone. The user specifically reports
Hisense **VIDAA / Odin** WebRTC abort/blackscreen during control request/release,
chat or similar actions, resolved by reload. Treat this as operator evidence of
an event-correlated symptom, not proof of renegotiation, ICE failure, decoder
exhaustion or chat-audio causality. Model, firmware, browser build and exact
failed layer remain open; do not infer them from older platform descriptions.
Target checks will be performed by the operator. Small passive delay is wanted;
no numerical accepted limit exists, and the HLS 24-s contract alone does not
establish suitability. B5/B7/B8 require the recorded usage/resource decisions;
all current paths remain available, without new automatic fallback.

## Comparative fork and transport review — 2026-10-07

**Original review scope: analysis and documentation at a91d9388, before the
subsequent authorization and source changes recorded above.** Baseline findings
in this section retain that original source state; the following
chronological records retain the evidence available at their respective commits.
The actionable sequence is [the current Workplan NEXT](WORKPLAN.md#next).

### Snapshot and strength of evidence

The initial working tree was clean on `testing`, at documentation HEAD
`a91d9388d75f3d042f212398a1dcff1af198992d`. The latest application change is
`8741f7880d9a709e6dc17924d2af0564a949ccd9`; `master` and `origin/master` remain
`d9105ef8`. The local `upstream/master` still points to the September integration
base `b0f01cedea68893e85a3fd852c0521238c285695`. A read-only GitHub API comparison
independently found current upstream master
`a2cb38dd10e20774b5f3ecf8b8c5d0b07a2051f5` (2026-10-04): 18 subsequent commits,
71 changed files. No fetch, merge, branch update or upstream adoption was done.
See [the upstream assessment](UPSTREAM_SYNC_AUDIT.md#read-only-upstream-comparison--2026-10-07).

Use these evidence classes throughout this assessment:

- **S — static:** actual source/configuration/history establishes a mechanism or
  boundary; it does not establish its runtime cost or the cause of a live incident.
- **T — supplied target evidence:** a recorded operator/server result at the named
  commit. Nothing was freshly executed in Codex. A later source change requires
  relevant revalidation; earlier device results are not fresh exact-8741 results.
- **O — open:** an unmeasured comparison, suspected cause, unsupported device or
  unknown usage. No numerical performance benefit is inferred from architecture.

The decisive existing results are:

| Evidence | What it supports | What it does not support |
|---|---|---|
| WebRTC adaptive checkpoints `bfaca84e`, `2efcc6b1`, `ddf15cee` (T) | The documented three-viewer scenario; a 20-minute healthy high-tier hold; real constrained downgrade; peer isolation; a bounded lower-tier recovery probe | A universally optimal profile, all devices, or cost versus another transport |
| WebCodecs bounded checkpoints `86893473`, `6b6cd328`, selection `12cfe43b` (T) | Roughly 15 minutes of technical playback; a corrected foreground iPhone interval over ten minutes with acceptable reported picture/audio and no retry; role/private-mode/manual-selection checks within their recorded limits | TV compatibility, numerical latency/pacing, background recovery, resource advantage or induced slow-client acceptance |
| Exact-8741 preparation and activation, PC/Helium (T) | Both old rolling-playlist/watchdog defects reproduced in isolation; 66 client tests/type/build plus stated wire/package/image gates passed; operator confirmed at least five minutes of moving conventional-HLS picture/audio without Retry/reload while WebRTC continued working | Reliable cold start, a scripted room-event matrix, native Safari/TV or LL-HLS acceptance, ten-minute/full Phase 4 acceptance, direct CPU/RAM/latency comparison or proof of the earlier live parser cause |

The exact HLS evidence and omissions remain in
[the rolling-window repair record](HLS_PLAYLIST_WINDOW_REPAIR_2026-10-07.md).
Preserve its images and private evidence at
`/opt/docker/nekoNew/neko-hls-results-8741f7880d9a`. This review needs no new
server command, rebuild, old-image trace or repeat of the passed five-minute gate.

### Actual transport responsibilities and shared dependencies

There are **three media choices, but two distinct kinds of WebSocket**:

1. **Normal event/session WebSocket:** the browser currently logs in through
   legacy `/ws`; the server adapter performs a local HTTP login and opens an
   internal `/api/ws` connection. It carries session initialization, SDP/ICE,
   chat, members, control grants/releases, capability negotiation, media tickets
   and permission/private-mode changes. It is mandatory for every media choice.
   High-rate WebRTC keyboard/mouse/touch input uses its RTCDataChannel; the event
   socket's room/control messages are a different function.
2. **WebRTC media:** a peer receives shared capture output over RTP, with ICE,
   RTCP and optionally the estimator/quality ladder. It also supplies interactive
   input through the data channel. It remains the normal default.
3. **Dedicated media WebSocket `/api/media/ws`:** an additional connection
   carries encoded VP8/Opus records, readiness/feedback/resync and heartbeat
   messages. A worker, WebCodecs decoders, AudioWorklet and canvas perform playback.
   It does not replace login, chat, room actions or the normal socket, and adds
   no interactive input transport. Selection is explicit and receive-only.
4. **HLS/LL-HLS HTTPS:** event-plane negotiation authorizes bootstrap and a
   scoped playback lease; playlists/init/parts/segments carry passive media.
   Native video or pinned `hls.js`/MSE plays H.264/AAC. Conventional and low-latency
   modes share one packager, workers and retained objects; they are not two
   independent capture systems. Eligibility remains passive/view-only plus the
   explicit administrative diagnostic path.

The central delivery manager owns one primary delivery per session and checks
the live `CanWatch` permission before/after attachment. The provider shares
immutable encoded data with bounded subscriber queues and explicit source
format, PTS/DTS validity, generations and keyframe/discontinuity state. These
boundaries have value even if an optional transport is later removed. Separate
queues and per-viewer leases contain many slow-consumer failures; they cannot
isolate a common event connection, native capture failure or host CPU exhaustion.

### Comparison of benefits, costs and evidence

The numbers below are configuration/acceptance targets unless explicitly marked
T. Lower transport overhead, quicker startup or a lower CPU/RAM footprint has
**not** been measured across these paths.
The retained media topology is unchanged; common-event bounds and the WebRTC
grace/progress repair above are now static source properties, with new target
evidence pending.

| Dimension | WebRTC | WebCodecs + media WebSocket | Conventional HLS | LL-HLS |
|---|---|---|---|---|
| Product role (S) | Interactive default; passive viewing also supported | Explicit receive-only experiment; same room | Explicit passive compatibility candidate | Passive candidate where lower HTTP-streaming delay is needed |
| Actual compatibility (T/O) | Existing desktop/mobile deployment; TV incident unresolved; omitted iPhone recovery phases stay open | Desktop and corrected foreground iPhone evidence; no affected-TV test | PC/Helium evidence at 8741; native Apple/TV acceptance open | No supplied live acceptance on the required devices |
| Capability limits (S/O) | RTCPeerConnection, negotiated codec, ICE/data channel and autoplay constraints | Worker VideoDecoder **and** AudioDecoder for exact VP8/Opus, AudioWorklet/AudioContext/canvas; no native PiP; viewport fullscreen on tested iPhone | Native H.264/AAC HLS or suitable MSE+hls.js, plus the modern Neko application/event socket; a TV's ability to play an m3u8 alone is insufficient | HLS client support plus blocking/part semantics and suitable HTTP/proxy/network behavior |
| Startup (S/T/O) | Login → signaling/ICE → first frame → autoplay; comparative distribution missing | Event login → ticket/socket → format/keyframe/audio buffering; fast-start targets unmeasured | Cold packager produces all four tracks; fixtures near 18.1 s and historical first picture near 20 s are different observations; 28 s server readiness allowance inside 30 s HTTP limit, not a measured p95 | Six-second readiness allowance; live startup unknown |
| Latency and bandwidth (S/O) | RTP/RTCP with congestion feedback; ICE may use relay; actual candidate and overhead matter | Reliable ordered TCP; loss can delay later media. Fixed/manual source, no WebRTC estimator; framing and feedback add traffic | Six-second parents and buffer/hold-back trade latency for robustness; ABR chooses among outputs; transcoding and HTTP overhead can increase bytes | One-second parts and blocking reload add request/state cost; lower-latency benefit unmeasured. Contract requires path p95 RTT ≤333 ms |
| Recovery (S/T/O) | One fresh-login owner and one bounded element repair; matched client/server 8-s transient grace and real progress checks are implemented but await target validation | Local audio reanchor and bounded same-backend resync/retry; no automatic backend fallback; corrected foreground interval passed | Player recovery plus independent HTTP/lease/progress checks; explicit Retry; latest rolling-window fixes passed bounded playback | Same lease/player safety plus more reload/part/discontinuity cases; runtime evidence open |
| Capture and encoding (S) | Shares demanded configured source; optional high/medium/low capture pipelines can run concurrently | Reuses the selected VP8/Opus source; per-viewer delivery and browser decoding, no additional server encoder | Pins shared `high` VP8 input plus Opus; **two provider subscriptions, three separate VP8 decoders/H.264 encoders and one Opus→AAC worker**, all shared by viewers | Same four workers and objects as conventional HLS; disabling LL mode alone does not save those encoders |
| CPU/RAM (S/O) | Pion/packetization, feedback and demanded capture cost; no clean cross-transport baseline | Queue/parser/writer and worker/browser decode cost; receiving on TCP can concentrate backlog; comparative footprint unknown | Largest identifiable additional native conversion machinery; 64 MiB object limit is not a total RSS limit; first viewer starts every variant | Similar native conversion cost, with additional HTTP waiter/playlist state; exact delta unmeasured |
| Failure isolation (S/T/O) | Per-peer media queues and accepted estimator isolation; common event writes now bounded but new slow-reader acceptance pending; shared capture/host remain | Per-delivery queue/write/decoder limits; induced hostile/slow-client and shared-event tests remain open | Shared packager means one rendition/audio-generation failure can affect many HLS viewers; bounded HTTP readers; five-minute coexistence supports only that interval | Same shared-packager boundary; blocking-request isolation needs its own gate |
| Authorization (S/T) | Live session/view-only controls, WebRTC/input permissions | TLS/exact Origin, ten-second one-use ticket via subprotocol, live permission recheck, credential-free source backend | TLS/exact Origin, one-use bootstrap and sliding 30-second HttpOnly path-scoped cookie lease, 15-second renewal, current session permission | Same security plus strict bounded reload/query handling |
| Maintenance (S) | Required core lifecycle; Pion, legacy signaling, input and estimator | Separate wire protocol, worker decoders, audio clock/resync, canvas/frame ownership and browser feature changes | Native and MSE players, pinned hls.js, transcoder, custom fMP4/playlist/retention/lease/security machinery | Further playlist, blocking reload, preload/rendition-report and proxy test cases |

The WebCodecs specification guarantees no particular codec implementation;
exact runtime probes remain necessary. Browser WebSocket also supplies no receive
backpressure. These limitations explain the need for the current bounded decode
and recovery machinery, not a measured disadvantage on the target devices.
Sources: [WebCodecs specification](https://www.w3.org/TR/webcodecs/),
[WebSocket API](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket),
[hls.js compatibility](https://github.com/video-dev/hls.js#compatibility).

### Findings and disposition

The following identifiers map to the planned blocks in `WORKPLAN.md`. They are
baseline findings; their implemented dispositions are recorded above, with
remaining evidence conditions in `WORKPLAN.md`.

**F1 — Common event delivery has unbounded waits (S); prioritize it across all
backends.** `SessionManagerCtx.Broadcast` sends sequentially. Normal
`WebSocketPeerCtx.Send/Ping` and the legacy bridge's two writers serialize writes
without a write deadline; `Destroy` first sends a disconnect and can wait before
closing. Legacy local HTTP requests have neither a client timeout nor a request
context tied to session cancellation. This is a structural path by which a slow
event reader/backend can delay broadcasts or cleanup; no supplied trace proves
it caused the TV failures. Add bounded operations and then, where needed, one
bounded writer owner with ordered lifecycle handling. Do not substitute
unbounded per-message goroutines or silently drop authorization/control events.
Relevant source: `server/internal/session/manager.go`,
`server/internal/websocket/peer.go`, `server/internal/http/legacy/session.go` and
`server/internal/http/legacy/handler.go`. Alternative media sockets do not bypass
this common dependency.

**F2 — WebRTC recovery policies conflict (S).** Client `base.ts` waits eight
seconds for a disconnected ICE peer to recover; server `webrtc/manager.go`
destroys a peer immediately on Pion peer `Disconnected` or `Failed`. ICE state
and peer state are different callbacks, so actual ordering needs a target trace;
the client window is not an end-to-end guarantee. The initial client timer is
also cleared at ICE `checking`, and `peerConnected` includes that state without
a separate first-frame deadline. Align bounded transient-disconnect ownership,
first-picture progress and final failure instead of adding another reconnect
loop. Existing 1/2/5/10-second fresh-login retries, stale-identity guards, explicit
logout/revocation stops and Safari's Play fallback remain useful. There is no
complete client/legacy ICE-restart flow merely because the current server API has
a `signalRestart` handler.

**F3 — WebRTC player state still has the callback problem already corrected in
WebCodecs (S, runtime effect O).** `video.vue` registers class-field arrow
callbacks for canplay/playing/pause/waiting/stalled/timeupdate while those
callbacks mutate private primitive recovery fields. Vue-class-component 7
constructs a synthetic data instance; class-field arrows can retain that instance
instead of the live Vue instance. The earlier WebCodecs repair and Vue's
[documented caveat](https://class-component.vuejs.org/guide/caveats.html), backed
by the [exact v7.2.6 implementation](https://raw.githubusercontent.com/vuejs/vue-class-component/v7.2.6/src/data.ts),
support reviewing these as bound prototype methods. No current WebRTC browser
reproduction or TV attribution is claimed. Independently, loadedmetadata uses
the canplay handler, any timeupdate marks progress, and startup before first play
has no equivalent element watchdog. Deliberate Pause/autoplay denial must be
distinguished from a stall. Consolidate one playback-progress owner, separate
from network recovery; do not copy HLS buffer/time constants into WebRTC.

**F4 — Some upstream correctness fixes are absent; capture convergence can
reduce special machinery (S).** The inbound published-track RTCP PLI ticker
goroutine ranges a channel that `ticker.Stop()` does not close; upstream #697
supplies an exit signal. This applies to admitted incoming microphone/camera
tracks, not every passive viewer. Upstream #711 derives IDs for configured
pipelines and guards an empty selection. Here missing IDs can make valid pipeline
definitions invisible to the media provider; the signal fallback still indexes
`videos[0]`, but the preceding current `openPeer` rejects empty provider sources,
so an end-to-end panic is **not** established. Preserve explicit quality ordering:
alphabetical `high, low, medium` is not the fork's `high, medium, low` ladder.
The upstream subscription API also addresses listener identity, per-sample
snapshot allocation, pipeline-start races, shutdown and bitrate locking. Local
bitrate bucket updates/reset use different locks while only the published gauge
is atomic; native teardown ordering/race checks are required. Adapt the low-level
capture seam once behind the richer existing provider, preserving timestamp
validity/generations and the fork's corrected **bits/s** units. Do not retain two
parallel subscription systems or replace the authorization/delivery manager.

**F5 — Adaptive quality has demonstrated value; keep the validated behavior
and default-off profile (S/T).** Receiver-loss/NACK evidence prevents a loss-free
GCC collapse alone from lowering quality; a one-tier peer-local probe breaks
lower-tier application-limited recovery deadlock with bounded backoff. Their
negative controls and supplied target downgrade/recovery/isolation results
justify this complexity. They do not justify universal bitrate/quantizer values
or enabling the multi-pipeline profile everywhere. Multiple active source tiers
cost captures/encoders; new HLS demand keeps high active. Historical whole-Brave
resource samples are not isolated estimator or transport measurements. Retain
the single-pipeline default, fixed diagnostics and the accepted profile; tune
only from a controlled target baseline.

**F6 — HLS pays a real conversion/topology cost; measured benefit is still
bounded (S/T/O).** `mediahls/transcoder.go` independently decodes the common VP8
high source for each of three H.264 outputs. Opus is decoded and encoded as AAC.
No second desktop capture is created by HLS itself, but it can start/retain high
capture and adds four native conversion workers. This is another lossy video
generation; a larger H.264 target rate cannot restore details already lost in
VP8. H.264 targets 3,000/1,100/365 kbit/s plus AAC 128 kbit/s and advertised
playlist bandwidth estimates are not measured egress or quality improvements.
The benefit is another passive playback ecosystem, with latest PC/Helium
coexistence evidence. A reduced ladder, shared decode/raw tee or a source codec
change is a separate design only if resources/quality make it necessary; clocks,
anchors and source isolation make it a risky premature rewrite.

**F7 — HLS source-format and lifecycle claims need a precise boundary (S/O).**
`workerFormatMatches` requires high VP8 at 1280×720, 25/1 fps and stereo 48-kHz
Opus. A changed generation with the supported format can be recovered; arbitrary
resolution/codec changes are not transparent supported input. Existing broad
resolution-change acceptance text must be read with this constraint and resolved
before promising seamless changes. The independent master probe (one-second
full-body deadline), lease renewal, readiness budget, player errors and progress
watchdog serve different purposes but overlap in failure reporting. Consolidate
their ownership/status where possible; preserve native-player revocation and
lease renewal. The old timeout/parser hypotheses are not confirmed causes, so
do not raise timeouts or buffers speculatively after the passed 8741 checkpoint.

**F8 — Ordinary room events do not statically explain the reported TV failure
(S/O).** Join/chat/control handlers do not directly clear or reconnect the media
delivery. A new capture listener requests a shared-source keyframe; other
listeners remain attached. This can create a bitrate burst, but causality is
unmeasured. A video-store comment claims these events renegotiate fresh tracks;
that causal assertion is not supported by the traced server path. Replacement
of old same-kind tracks and listener cleanup still serve a valid purpose and
must not be removed on that comment alone. Default chat sounds allocate a new
Audio element and chat history grows without a fixed bound. These are separate
device/long-session candidates, not TV diagnoses. Compare sound on/off as one
variable, record socket/ICE/player failure separately, and change history/render
behavior only with actual resource evidence and retained product requirements.

**F9 — Small maintenance cleanup and dependency work have higher confidence
than removing a useful transport (S/O).** `client/vue.config.js` is a Vue CLI
artifact while package scripts use Vite; legacy `VUE_APP_*` environment fallbacks
and special development-host timeout logic need an actual-caller inventory.
No automatic migration/removal is justified for existing deployments. Client
Google STUN injection checks only `stun:` and can augment an intentional TURN-only
or `stuns:` configuration. Establish intended operator policy, actual ICE
candidate/relay use and startup phases before simplifying it. Non-trickle
gathering and external address discovery also need bounded cancellation review;
more public STUN servers are not a demonstrated performance fix. Keep mobile
touch/input, view-only enforcement and the chat/tooltip/Axios containment repairs.
The supplied dependency audit contains 20 package entries, not 20 proven reachable
production flaws; unpatched packages and Vue 2 maintenance remain open. Review
runtime/build reachability and perform scoped updates separately from recovery
or transport changes. See [the dependency classification](DEPENDENCY_AUDIT_2026-10-04.md).

### Recommendation and alternatives

**Retain WebRTC as the core/default and improve its common event, startup and
recovery paths first. Retain conventional HLS as a bounded opt-in passive
compatibility candidate. Keep WebCodecs implemented but freeze expansion until
comparative evidence identifies a real niche. Keep LL-HLS unadvertised in the
conventional-only deployment and defer its acceptance/expansion until there is a
latency requirement.** The HLS server feature remains default-off in the
repository even though the supplied current target has conventional HLS enabled.
No deletion, new transport, automatic fallback, broad upstream merge or master
promotion is authorized by this assessment.

| Option | Benefit | Cost or lost behavior | Decision condition |
|---|---|---|---|
| Keep all three media paths; qualify each explicitly | Preserves tested WebCodecs iPhone behavior and HLS compatibility candidate | Three delivery/client lifecycle surfaces plus LL modes remain to maintain | Worthwhile if actual device/network needs and accepted resource budgets justify both optional paths |
| WebRTC + conventional HLS; retire WebCodecs later | Likely simpler long-term split: interactive RTP and passive HTTP ecosystem | Loses manual VP8/Opus-over-HTTPS receive path and its tested foreground behavior; saved preferences/overlays/contracts need migration | Favor only if inventory and same-device comparison show no retained WebCodecs use or unique accepted benefit |
| WebRTC + WebCodecs; retire HLS later | Avoids HLS native transcodes, custom packaging and HLS HTTP lifecycle | Loses native passive HLS ecosystem and ABR; TV compatibility remains unresolved | Favor if HLS has no required compatible devices or fails resource/quality gates and WebCodecs covers actual receive needs |
| WebRTC only | Smallest delivery architecture and one primary media stack | Loses both alternative receive paths; offers no compatibility remedy for devices that cannot use WebRTC | Favor only if all required devices/networks pass WebRTC and users no longer need either optional path |
| Remove only LL-HLS, retain conventional HLS | Reduces reload/part/client/proxy acceptance combinations | Requires contract/parser/advertisement cleanup; most conversion cost and some shared fMP4 machinery remain | Reasonable after confirming no low-delay passive requirement; current mode restriction already provides a reversible first step |

Actual active use of each path, required device list and tolerated latency remain
unknown. The operator's experiments establish usage during checkpoints, not an
ongoing production population. Optional-path retirement should therefore begin
with an explicit reversible disablement decision and end with removal only after
its compatibility and saved-preference consequences are accepted. Do not remove
the **normal event WebSocket**, central authorization, provider or interactive
data channel while retiring the **dedicated media WebSocket**.

The implementation and measurement blocks, their dependencies, target checks and
rollback are recorded in [WORKPLAN.md](WORKPLAN.md#next). This is a completed
static assessment with open empirical/device questions, not full transport
acceptance or a claim to have reproduced the Smart-TV incident.

## Historical HLS source follow-up — 2026-10-05 to 2026-10-07

Historical requested source-only follow-up (2026-10-05):
[the HLS client stability review](HLS_CLIENT_STABILITY_REVIEW_2026-10-05.md)
corrects paused-time stalls, monitoring before first readiness and mixed
HTTP/readiness error counts, with defensive player-event handling. Eight new
regression cases passed in the supplied exact-68 isolated 60-test/type/build
gate (Client-Check-Exitcode 0). Scoped exact-68 image preparation subsequently
passed with Image-Prepare-Exitcode 0 while the exact-73 service stayed running;
exact-68 activation then passed healthy with Start-Exitcode 0. The operator
reports HLS works after Retry; reliable first start and wider live acceptance
remain open. All results are supplied target evidence, NOT EXECUTED IN CODEX.
The reported frozen-picture/room-event symptoms are not yet correlated to these
defects. Exact-68 read-only diagnosis subsequently passed: one not-ready and
one successful bootstrap, generation-1 delivery and 17 successful segments,
without a sampled process exit/OOM; normal idle-stop cleanup was recorded.
The saved timing summary then passed: one request approximately 24 seconds,
one at most 1 ms, two requests totaling 24.001096458 seconds. The sequence fits
a readiness deadline followed by a warmed Retry, without attempt correlation.
The new bounded server-only candidate changes conventional startup allowance
from 24 to 28 seconds inside the unchanged 30-second HTTP/client limits. It is
not yet a verified fix. Exact-8f54970f focused preparation subsequently passed
with Prepare-Exitcode 0: all 13 selected native checks passed, including the four
real-codec fixtures in generation 1; candidate base/Brave images and private
evidence were recorded. Unchanged client build layers were cached. Target
checkout is exact-8f; live was exact-68 at preparation. After the requested
activation, the operator reports an apparently frozen first picture, then
moving HLS video/audio after page reload. Activation CLI and initial
player/error/timing evidence are missing; reliable first start remains open.
Read-only diagnosis subsequently passed with Diagnostic-Exitcode 0 and matching
prepared/live exact-8f image IDs: healthy container, two successful bootstraps,
one startup generation per track, all four workers running, one active lease
and 135 successful segment requests, without a sampled exit/OOM or fixed error.
This does not identify the browser's frozen-picture cause. A client-only
start-order candidate requests existing autoplay after attachment rather than
waiting for canplay. Exact-7dcc3c5e preparation subsequently passed with
Prepare-Exitcode 0: both images and private snapshot/marker recorded, live
exact-8f retained. Earlier isolated old/new client tests/type/build precede
the supplied base-image-build tail and are covered by final helper success,
without visible counts. Backend evidence is inherited from unchanged sources.
After the requested exact-7dcc activation, the operator reports moving first
picture after about 20 seconds, then another frozen picture requiring reload.
The subsequent progress diagnostic passed, confirming the prepared exact-7dcc
live image and healthy service. HLS was idle in both valid 12.012-second samples,
with zero leases/subscriptions/workers and no new publication/HTTP/generation
events; the operator confirmed HLS was closed/logged out. This does not capture
the earlier browser freeze. The original activation CLI remains unsupplied.
A WebRTC participant joining/starting video is an uncorrelated hypothesis.
Static review found a keyframe request for a new shared-capture listener, with
HLS listeners retained; no direct join-triggered HLS teardown/source overwrite
was found. The corrected `4593a6f9` progress helper subsequently passed with
Diagnostic-Exitcode 0 and valid 12.011-second deltas: one active lease, all four
workers, continuing audio/video publication and successful HTTP media requests,
with no interval generation/discontinuity/drop/capture-creation/WebRTC-open
increments. VP8 capture rows are visible. The operator confirms the sample was
taken during frozen HLS picture/audio loss without reload. Later the page failed
with "HLS HTTP connection failed; retry manually"; that later failure is outside
the sampled server interval. The supplied retained browser timings show one
1,001-ms master with no response metadata; the full 250-entry buffer omits later
failures. Status-zero entries can have body data; terminal cleanup cleared the
video. The full five-minute trace now captures a freeze at 23.948 seconds of
media, fixed buffer/513 frames, audio loading stopped after four segments and
repeated seek-only jumps while medium-video HTTP delivery continues. No new
terminal HTTP error is captured. Static inspection confirms a rolling-playlist
discontinuity defect and seek-only watchdog bypass; both are repaired on testing.
Exact-8741 target preparation subsequently passed: controlled old defects,
66 client tests/type/build, repaired wire/HLS-package checks and candidate images.
Checkout/live deployment is now exact-8741: activation passed with Activate-Exitcode 0
and a healthy candidate image. The operator confirmed at least five minutes of
moving HLS picture/audio without Retry/reload while WebRTC continued working.
This bounded PC/Helium sustained-playback gate is passed; [remaining grouped
acceptance](HLS_PLAYLIST_WINDOW_REPAIR_2026-10-07.md#remaining-acceptance)
stays open. No new server command, rebuild, old-image trace or timeout tuning.
The uncaptured parser error as the old audio-stop cause remains an inference.
Supplied target/browser results, NOT EXECUTED IN CODEX; the corrected saved-startup
helper has no fresh execution. Repeated cold-start and wider acceptance remain open.
Preserve prior evidence.
An explicit room-action sequence was not supplied for the passed interval. The wider device,
authorization/lifecycle and resource gates stay open.

Operator direction recorded on 2026-10-04. This is a required review and validation plan, not a claim that the reported device failures have been reproduced or fixed.

## Reported behavior and evidence limits

- Some older Smart-TV viewers reportedly lose their WebRTC connection when another participant joins, posts a chat message, takes control or releases control. Quiet continuous viewing reportedly works. The affected television belongs to a colleague and is currently unavailable for direct testing.
- Hisense VIDAA is a suspected affected platform. Exact model, firmware, operating-system and browser versions are unknown; do not generalize this report to every VIDAA device.
- The existing WebCodecs/media-WebSocket path has **not been tested on those televisions**. It remains implemented and must not be removed or declared unsuitable on that evidence.
- Some modern mobile clients reportedly take longer to show the first WebRTC picture. Treat startup delay as a separate observation from event-triggered TV disconnects until evidence connects them.
- Source inspection found that join, chat and several control transitions call `chat.newMessage`; with `chat_sound` enabled it creates a new `Audio('chat.mp3')` element. The default is enabled. This is a concrete investigation candidate, **not a confirmed cause**. Do not change unrelated media settings simultaneously when testing it.
- HLS is a compatibility candidate for passive viewers, not a proven fix or a demonstrated stability improvement over either existing backend. The event/session connection remains authoritative for every backend.

## Historical implementation and validation order

This records the earlier HLS sequence. New application work now waits for the
operator's decision on [the comparative plan](WORKPLAN.md#next); the sequence
below does not authorize further repairs or target execution in this task.

HLS/LL-HLS Phases 1–3 and the separate Phase 4 repository assets are implemented on `testing`. The unavailable television does not block repository implementation or checks on available devices; its device acceptance remains pending. Keep WebRTC as the default, preserve explicit WebCodecs selection and leave HLS default-off. No automatic fallback or `master` promotion is authorized.

After the implementation blocks, prepare the separate HLS Phase 4 deployment/observability assets. Before final grouped target-server validation and before any promotion decision, perform the mandatory static review below. Fix clearly established defects in their own implementation block as they are found; the final review is not a reason to postpone them.

The [2026-10-04 review record](STABILITY_REVIEW_2026-10-04.md) covers source/configuration items 1–7 and repairs legacy notification error handling plus paused HLS keepalive. Those repairs passed automated target checks at e55; their live matrix remains pending. Item 8 is now [classified against the supplied audit](DEPENDENCY_AUDIT_2026-10-04.md), with additional chat/compiler, member-tooltip and Axios containment changes passing the exact 47-test/type/build/image gate at `93f1fa63`. Deployment/browser checks, package remediation and final security acceptance remain open. Follow [the Phase 4 runbook](HLS_LL_HLS_VALIDATION.md). The subsequent failed HLS playback and normal-login incident has supplied read-only diagnostic evidence; [startup repairs and a mandatory real-codec gate](HLS_STARTUP_REPAIR_2026-10-04.md) passed the target automated/image gate at 80020d99; live acceptance remains pending. The recorded pre-HLS runtime was restored healthy with exit 0 and operator-confirmed normal login/picture/audio. Default-off deployment passed with Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login/picture/audio/control. Same-image HLS activation passed with Enable-Exitcode 0, healthy service and 19/19 HTTP denial probes, but HLS failed and the operator reported all streams stopped afterward. The read-only diagnostic found one ready packager/lease, 23 successful segment requests and two timeline-gap rejections, with no sampled Neko exit/OOM. Default-off restoration passed with Recovery-Exitcode 0 and 2/2 disabled-route probes; fresh browser confirmation is pending. The HLS-only fixed-GOP correction passed the isolated target GOP A/B gate at 97ba4ad9: old code reproduced two timeline gaps and all three repaired codec tests passed, including 30.19 seconds of scene cuts in generation 1. Exact 97ba4ad9 automated/image preparation passed with Repair-Check-Exitcode 0 (47 client tests, 13 Go packages, both fuzz jobs, all three codec tests and base/Brave builds). Default-off deployment of my-neko/brave:hls-97ba4ad9ab3e passed with Baseline-Exitcode 0, healthy service and 2/2 disabled-route probes. The operator reported the requested normal browser check works. The subsequent HLS attempt at the prepared 97ba4ad9 checkpoint failed with "HLS bootstrap failed; retry manually", and the operator confirmed WebRTC also stopped working. The latest enablement CLI/HTTP results have not been supplied. Read-only diagnosis passed with 84 log lines, one not-ready bootstrap, one negotiation rejection and no sampled exit/OOM or generation/lease-open markers. Same-image default-off restoration passed with Recovery-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login/picture/audio work again. The saved-startup summary passed with Saved-Check-Exitcode 0: audio/high/medium subscriptions persisted, low capture reached only its create marker, and the saved matching environment had HLS enabled. Paired startup diagnosis at helper 88f2b25d completed with exit 0: both constructors passed two starts and failed one, with medium losing its initial pre-anchor IDR and reaching only two parents. This narrows the smooth readiness defect independently of registry isolation. The 414639d2 isolated anchor A/B gate passed with Anchor-Check-Exitcode 0: old code reproduced the fixed initial-IDR failure; seven corrected checks each passed three times, with every real-codec fixture staying in generation 1. Full exact-414639d2 preparation FAILED with Repair-Prepare-Exitcode 1: 47 client tests, type/build, 13 Go packages, both fuzz jobs, registry/mapping and anchor lifecycle checks passed, but smooth readiness restarted with worker_failure and failed its generation-1 assertion at 20.03 seconds; scene cuts passed at 30.19 seconds in generation 1. Base/Brave image steps were not reached. Worker diagnosis at helper 53034495 then completed with exit 0: all six checks passed in three fresh processes, each smooth fixture ready in 18.11 seconds in generation 1 with no rejected pushes; the earlier worker failure was not reproduced. The controlled audio-anchor A/B at repair 71a14d21 passed with Audio-Anchor-Exitcode 0: the old AAC hold reproduced both the blocked-drainage unit failure and an audio/anchor/queue_full restart under 256 ms high-input delay; all eight corrected startup checks passed in three fresh processes and scene cuts passed once (25 top-level passes). All seven real-codec fixtures stayed in generation 1; normal readiness was 18.10/18.11/18.11 seconds and delayed-high readiness 18.09 seconds each. This verifies the controlled AAC-overflow mechanism, not the cause of the earlier unobserved worker failure or live all-stream outage. Full exact-71a14d21 preparation then passed with Repair-Prepare-Exitcode 0: the rebuilt GStreamer 1.26.2 image passed all nine selected startup checks, including normal readiness at 18.11 seconds, delayed-high readiness at 18.09 seconds and scene cuts at 30.19 seconds, all in generation 1; base/Brave images were built and the service remained unchanged. The supplied tail starts inside the codec-image build; earlier client/Go/fuzz steps are covered by the script's reported final success, not separately shown in this excerpt. Default-off deployment of my-neko/brave:hls-71a14d2174da then passed with Baseline-Exitcode 0, a healthy service and 2/2 disabled-route probes. The operator confirmed the requested normal browser check works without HLS at 71a14d21. Same-image conventional-HLS activation at 71a14d21 then passed with Enable-Exitcode 0, healthy service, a private enable snapshot and 17/17 public plus 2/2 cleartext-denial probes. The subsequent operator-reported HLS attempt at 71a14d21 went from connecting to failed with "HLS bootstrap failed; retry manually"; the operator reported only WebRTC streaming works. No successful HLS picture/audio or room-event interval is demonstrated. Read-only diagnosis passed with 950 captured lines, one not-ready bootstrap, one started/idle-stopped packager generation, medium/low keyframe-admission drops of 759/564 and cumulative part/segment publication only for audio/high. No sampled Neko exit/OOM was found. Same-image default-off restoration passed with Recovery-Exitcode 0, healthy service and 2/2 disabled-route probes; the operator confirmed normal login/picture/audio work again. The isolated source-clock-phase diagnosis at helper 409482b4 passed with Clock-Skew-Exitcode 0: aligned control ready in 18.11 seconds; all three skew runs reproduced not-ready at 24.02 seconds in generation 1 with audio/high ready, medium/low blocked, flowing IDRs and no rejected native pushes. This proves the controlled phase-admission defect, not the exact unmeasured live phases. A common high-source fan-out repair is implemented and statically reviewed: one shared video provider subscription feeds the existing three scaled encoders, plus one audio subscription; provider PTS/DTS are preserved. The isolated common-source repair A/B passed at a7ffb8b1 with Shared-Clock-Exitcode 0: the old defect reproduced once; 46 positive top-level checks passed across three cold processes including one scene-cut check, with all ten real-codec fixtures in generation 1. Full exact-a7ffb8b1 preparation then passed with Repair-Prepare-Exitcode 0: all thirteen selected native/startup checks passed in the rebuilt GStreamer 1.26.2 image, including the four real-codec fixtures in generation 1, and base/Brave images were built without changing the running service. The supplied excerpt starts inside the codec-image build; earlier client/Go/fuzz stages are covered by final script success without separately shown fresh counts. Default-off deployment of my-neko/brave:hls-a7ffb8b13448 then passed with Baseline-Exitcode 0, healthy service, a private baseline snapshot and 2/2 disabled-route probes; the operator reported the requested normal browser check works. Same-image a7ffb8b1 HLS activation passed with Enable-Exitcode 0, healthy service and 19/19 HTTP denial probes. The operator then reported first HLS picture and the compact streaming label, followed after roughly 30 seconds by "HLS playback did not become ready; retry manually"; Retry HLS restored playback and WebRTC continued working. Static inspection found the initial client readiness deadline was never disarmed on canplay/playing. Client-only repair 73d5ff6d cancels it on those current-player events, arms it before attachment and retains the independent startup/stall bounds; Read-only target diagnosis then passed with a healthy a7ff image, two successful HLS bootstraps and 486 successful segment requests, with no sampled process exit/OOM; the packager stopped after idle grace. The isolated target client gate passed with Client-Check-Exitcode 0: the old timer defect reproduced, all 52 repaired client tests plus type/build passed, and checkout/live service remained at a7ff. These are supplied target results, NOT EXECUTED IN CODEX. Scoped exact-73d5ff6d image preparation then passed with Client-Image-Exitcode 0: fresh client bundle index-CrHQRMnq.js, cached unchanged server/runtime layers, both base/Brave images and private snapshot/marker recorded while enabled a7ff stayed running. Exact-73d5ff6d default-off deployment then passed with Baseline-Exitcode 0, healthy my-neko/brave:hls-73d5ff6d2911, a private baseline snapshot and 2/2 disabled-route probes; the operator confirmed the requested normal browser check works without a media override. Same-image conventional-HLS activation then passed with Enable-Exitcode 0, healthy service, a private enable snapshot and 19/19 denial probes. The operator reports PC/Helium HLS playback after an initial Retry and one frozen-picture/page-reload incident; WebRTC kept working and later HLS worked normally. A HLS failed message was confirmed without its detailed error or exact timing, so startup reliability and an uninterrupted room-event interval remain unverified. Checkout/live image is now 73d5ff6d with conventional HLS enabled. Read-only exact-73 diagnosis then passed with Diagnostic-Exitcode 0: healthy service, no sampled exit/OOM or fixed error markers, one active HLS lease/all four workers running, six successful bootstraps and 394 successful segment requests. One not-ready bootstrap supports readiness as a possible initial-Retry explanation without attempt correlation; two startup-labelled generations and one idle stop do not establish a crash loop. The operator cannot confirm the exact uninterrupted interval and mentions possible random reconnects/room actions without correlation. NEXT consolidate startup/frozen-picture/room-event investigation into one bounded later validation step; no further ad-hoc operator check requested at this checkpoint; startup/recovery and grouped acceptance pending; the all-stream symptom's cause is unconfirmed. The earlier normal-login blocker remains unconfirmed.

## Mandatory final static review

Review the integrated code and relevant configuration at an identified `testing` commit:

1. Trace join, chat, control take/release/grant and member-state updates through event handlers, stores and UI. Check that ordinary room events cannot accidentally replace, clear or reconnect a viewer's media delivery.
2. Inspect chat sounds, notifications, emotes, DOM/media-element changes and old-browser API/error handling. Identify repeated audio creation, unhandled failures and work that could interfere with playback. Separate confirmed code defects from device hypotheses.
3. Trace startup from login/event initialization through ICE readiness or alternative-backend negotiation, first decodable video and playback. Inspect deadlines, cold-source/keyframe admission and autoplay/user-gesture handling; a connected session is not evidence of a displayed picture.
4. Review recovery state transitions, stale callbacks, concurrent requests, bounded retries, generation changes and ownership of timers/listeners/sockets/decoders. Check logout, kick, permission loss, private-mode pause/resume and backend replacement cleanup.
5. Recheck server-enforced view-only boundaries and the common event/session authorization above WebRTC, WebCodecs and HLS. Receive transport never grants control rights.
6. Inspect queue/buffer limits, source demand, shared HLS workers, slow-viewer isolation and resource release. A bounded queue alone does not prove device performance or acceptable total CPU load.
7. Review configuration consistency and quality-related choices without speculative tuning. Change bitrate, quantizer, latency or buffer settings only with an explicit goal and a comparable baseline.
8. Review the target-server dependency audit report and distinguish runtime dependencies from build-only dependencies, affected versions and reachable behavior. At `e85d8568`, `npm ci` reported 20 findings including one critical, but no detailed advisory report was supplied. Classify and address confirmed applicable findings without a blind breaking dependency update.

Record findings, affected files, the reason for each repair and required target checks. Review the resulting final diff again. No broad rewrite, additional transport or dependency migration is implied by this review.

## Target-server and device follow-up

All project tests, type checks, builds, containers and device/media checks are **NOT EXECUTED IN CODEX**. Run the relevant accumulated suites and grouped acceptance at the final exact commit and image digest on the real target server.

- On an available affected TV, repeat a fixed sequence of joins, messages and control take/release/grant while another viewer watches. First compare WebRTC with chat sound enabled versus disabled, changing only that setting; then compare explicit WebCodecs and HLS on the same device and content when each is supported.
- Record whether the failure is event-WebSocket closure/session loss, ICE/data-channel failure, media decoder/player failure, a stall or only a UI/playback change. Keep credential-bearing URLs, tickets and cookies out of collected evidence.
- Record the TV model/firmware/browser and actual codec/player capability when the device becomes available. Until then, keep the TV result untested rather than passed or failed.
- Measure mobile startup stages separately: event login, ICE connection (for WebRTC), first frame and actual playback. Compare Wi-Fi/mobile-network conditions on the same client before changing defaults.
- Run the existing role/private-mode/reconnect matrix plus cross-backend resource/isolation gates from `WORKPLAN.md` and `HLS_LL_HLS.md`. Measure visual quality, pacing, startup and latency; static inspection cannot establish those outcomes.

Completion means the static findings are reviewed and repairs are verified on the target server. Missing device phases remain explicitly open and do not become passes through a bounded checkpoint. Any WebCodecs removal or promotion to `master` needs a later explicit operator decision supported by the comparison results.
