package mediaws

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

type backendTestLease struct {
	mu      sync.Mutex
	state   types.MediaDeliveryState
	valid   bool
}

func (*backendTestLease) ID() string        { return "webcodecs-ws-1" }
func (*backendTestLease) SessionID() string { return "viewer" }
func (*backendTestLease) Backend() string   { return BackendName }
func (lease *backendTestLease) SetState(state types.MediaDeliveryState) bool {
	lease.mu.Lock()
	lease.state = state
	valid := lease.valid
	lease.mu.Unlock()
	return valid
}
func (lease *backendTestLease) Valid() bool {
	lease.mu.Lock()
	defer lease.mu.Unlock()
	return lease.valid
}

type backendTestSubscription struct {
	mu     sync.Mutex
	source types.MediaSource
	events chan types.MediaEvent
	paused bool
	closed bool
}

func (subscription *backendTestSubscription) ID() string { return "subscription" }
func (subscription *backendTestSubscription) Source() types.MediaSource {
	subscription.mu.Lock()
	defer subscription.mu.Unlock()
	return types.CloneMediaSource(subscription.source)
}
func (subscription *backendTestSubscription) Events() <-chan types.MediaEvent { return subscription.events }
func (*backendTestSubscription) Switch(_ context.Context, _ types.MediaSelector) error {
	return nil
}
func (subscription *backendTestSubscription) SetPaused(paused bool) error {
	subscription.mu.Lock()
	subscription.paused = paused
	subscription.mu.Unlock()
	return nil
}
func (subscription *backendTestSubscription) Close() error {
	subscription.mu.Lock()
	subscription.closed = true
	subscription.mu.Unlock()
	return nil
}

func videoTestSource() types.MediaSource {
	return types.MediaSource{
		ID:         "high",
		Kind:       types.MediaKindVideo,
		Generation: 1,
		Codec: types.MediaCodec{
			Name:      "vp8",
			MIMEType:  "video/VP8",
			ClockRate: 90000,
		},
		Width:                1920,
		Height:               1080,
		FrameRateNumerator:   25,
		FrameRateDenominator: 1,
		NominalBitrate:       1_996_800,
	}
}

func audioTestSource() types.MediaSource {
	return types.MediaSource{
		ID:         "audio",
		Kind:       types.MediaKindAudio,
		Generation: 1,
		Codec: types.MediaCodec{
			Name:      "opus",
			MIMEType:  "audio/opus",
			ClockRate: 48000,
			Channels:  2,
		},
	}
}

func newBackendTestDelivery(kind Kind, source types.MediaSource) (*Delivery, *deliveryTrack) {
	lease := &backendTestLease{valid: true}
	delivery := newDelivery(lease, nil)
	subscription := &backendTestSubscription{source: source, events: make(chan types.MediaEvent)}
	delivery.addTrack(kind, subscription)
	return delivery, delivery.tracks[kind]
}

func addBackendTestTrack(delivery *Delivery, kind Kind, source types.MediaSource) *deliveryTrack {
	subscription := &backendTestSubscription{source: source, events: make(chan types.MediaEvent)}
	delivery.addTrack(kind, subscription)
	return delivery.tracks[kind]
}

func TestDeliverySerializesCompleteFormatAndKeyframeGatedUnits(t *testing.T) {
	delivery, track := newBackendTestDelivery(KindVideo, videoTestSource())
	source := videoTestSource()
	if err := delivery.handleMediaEvent(track, types.MediaEvent{Type: types.MediaEventTypeFormat, Source: source}); err != nil {
		t.Fatal(err)
	}
	format, ok := delivery.queue.pop()
	if !ok {
		t.Fatal("complete FORMAT was not queued")
	}
	parsed, err := ParseRecord(format.encoded)
	if err != nil || parsed.Type != RecordFormat || parsed.Generation != 1 {
		t.Fatalf("FORMAT = %#v, err = %v", parsed, err)
	}

	delivery.mu.Lock()
	delivery.ready = true
	delivery.mu.Unlock()
	delta := types.EncodedMediaUnit{Generation: 1, PTS: time.Millisecond, Duration: 40 * time.Millisecond, Data: []byte{0x01}}
	if err := delivery.handleMediaEvent(track, types.MediaEvent{Type: types.MediaEventTypeUnit, Source: source, Unit: delta}); err != nil {
		t.Fatal(err)
	}
	if _, ok := delivery.queue.pop(); ok {
		t.Fatal("delta frame crossed initial keyframe gate")
	}

	keyframe := types.EncodedMediaUnit{Generation: 1, PTS: 2 * time.Millisecond, PTSValid: true, Keyframe: true, Data: []byte{0x00}}
	if err := delivery.handleMediaEvent(track, types.MediaEvent{Type: types.MediaEventTypeUnit, Source: source, Unit: keyframe}); err != nil {
		t.Fatal(err)
	}
	unit, ok := delivery.queue.pop()
	if !ok {
		t.Fatal("keyframe UNIT was not queued")
	}
	parsed, err = ParseRecord(unit.encoded)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Type != RecordUnit || parsed.Sequence != 0 || parsed.Duration != 40000 || parsed.Flags&(FlagKeyframe|FlagPTSValid) != FlagKeyframe|FlagPTSValid {
		t.Fatalf("video UNIT = %#v", parsed)
	}
}

func TestDeliveryRejectsVP8MarkerMismatchAndDerivesOpusDuration(t *testing.T) {
	videoDelivery, videoTrack := newBackendTestDelivery(KindVideo, videoTestSource())
	videoTrack.formatSent = true
	videoTrack.lastProviderGen = 1
	videoDelivery.ready = true
	mismatch := types.EncodedMediaUnit{Generation: 1, Keyframe: true, Duration: time.Millisecond, Data: []byte{0x01}}
	if err := videoDelivery.handleMediaEvent(videoTrack, types.MediaEvent{Type: types.MediaEventTypeUnit, Source: videoTestSource(), Unit: mismatch}); err == nil {
		t.Fatal("VP8 keyframe marker mismatch was accepted")
	}

	audioDelivery, audioTrack := newBackendTestDelivery(KindAudio, audioTestSource())
	if err := audioDelivery.handleMediaEvent(audioTrack, types.MediaEvent{Type: types.MediaEventTypeFormat, Source: audioTestSource()}); err != nil {
		t.Fatal(err)
	}
	_, _ = audioDelivery.queue.pop()
	audioDelivery.ready = true
	unit := types.EncodedMediaUnit{Generation: 1, PTS: time.Millisecond, Data: []byte{0x00}}
	if err := audioDelivery.handleMediaEvent(audioTrack, types.MediaEvent{Type: types.MediaEventTypeUnit, Source: audioTestSource(), Unit: unit}); err != nil {
		t.Fatal(err)
	}
	queued, ok := audioDelivery.queue.pop()
	if !ok {
		t.Fatal("audio UNIT was not queued")
	}
	parsed, err := ParseRecord(queued.encoded)
	if err != nil || parsed.Duration != 10000 || parsed.Flags&FlagPTSValid != 0 {
		t.Fatalf("audio UNIT = %#v, err = %v", parsed, err)
	}
}

func TestDeliveryResyncClearsMediaAndAdvancesGeneration(t *testing.T) {
	delivery, track := newBackendTestDelivery(KindVideo, videoTestSource())
	track.formatSent = true
	track.lastProviderGen = 1
	delivery.ready = true
	delivery.queue.pushMedia(testQueuedUnit(KindVideo, 10))
	if err := delivery.performResync(KindVideo, "server_overflow"); err != nil {
		t.Fatal(err)
	}
	first, ok := delivery.queue.pop()
	if !ok || first.record.Type != RecordDiscontinuity || first.record.Generation != 2 {
		t.Fatalf("resync discontinuity = %#v/%v", first.record, ok)
	}
	second, ok := delivery.queue.pop()
	if !ok || second.record.Type != RecordFormat || second.record.Generation != 2 {
		t.Fatalf("resync format = %#v/%v", second.record, ok)
	}
	if _, ok := delivery.queue.pop(); ok {
		t.Fatal("stale media remained after resync")
	}
	if !track.awaitKeyframe || track.nextSequence != 0 || track.transitioning {
		t.Fatalf("resynchronized track = %#v", track)
	}
}

func TestDeliveryDefersCommonResyncUntilEveryFormatExists(t *testing.T) {
	delivery, video := newBackendTestDelivery(KindVideo, videoTestSource())
	audio := addBackendTestTrack(delivery, KindAudio, audioTestSource())
	if err := delivery.handleMediaEvent(video, types.MediaEvent{Type: types.MediaEventTypeFormat, Source: videoTestSource()}); err != nil {
		t.Fatal(err)
	}
	_, _ = delivery.queue.pop()

	delivery.requestResync(KindNone, "server_overflow")
	select {
	case request := <-delivery.resync:
		t.Fatalf("common resync escaped before all formats: %#v", request)
	default:
	}
	if !video.transitioning {
		t.Fatal("published video was not gated while common resync waited for audio")
	}

	if err := delivery.handleMediaEvent(audio, types.MediaEvent{Type: types.MediaEventTypeFormat, Source: audioTestSource()}); err != nil {
		t.Fatal(err)
	}
	select {
	case request := <-delivery.resync:
		if request.kind != KindNone || request.reason != "server_overflow" {
			t.Fatalf("released common resync = %#v", request)
		}
	default:
		t.Fatal("common resync was not released after the audio FORMAT")
	}
	if !video.transitioning || !audio.transitioning {
		t.Fatal("tracks were not held behind the common lifecycle transition")
	}
}

func TestDeliveryDiscontinuityGatesUnitsUntilResyncLifecycle(t *testing.T) {
	delivery, track := newBackendTestDelivery(KindVideo, videoTestSource())
	if err := delivery.handleMediaEvent(track, types.MediaEvent{Type: types.MediaEventTypeFormat, Source: videoTestSource()}); err != nil {
		t.Fatal(err)
	}
	_, _ = delivery.queue.pop()
	delivery.ready = true

	source := videoTestSource()
	source.Generation = 2
	if err := delivery.handleMediaEvent(track, types.MediaEvent{
		Type:   types.MediaEventTypeDiscontinuity,
		Source: source,
		Discontinuity: types.MediaDiscontinuity{
			Generation: 2,
			Reason:     "source_restart",
		},
	}); err != nil {
		t.Fatal(err)
	}
	if !track.transitioning {
		t.Fatal("provider discontinuity did not close the media gate")
	}
	if err := delivery.handleMediaEvent(track, types.MediaEvent{
		Type:   types.MediaEventTypeUnit,
		Source: source,
		Unit:   types.EncodedMediaUnit{Generation: 2, PTS: time.Millisecond, Duration: time.Millisecond, Keyframe: true, Data: []byte{0x00}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, ok := delivery.queue.pop(); ok {
		t.Fatal("UNIT crossed a pending provider lifecycle transition")
	}
	if err := delivery.handleMediaEvent(track, types.MediaEvent{Type: types.MediaEventTypeFormat, Source: source}); err != nil {
		t.Fatal(err)
	}
	select {
	case request := <-delivery.resync:
		if request.kind != KindVideo || request.reason != "source_restart" {
			t.Fatalf("provider transition resync = %#v", request)
		}
	default:
		t.Fatal("provider FORMAT did not request the pending resync")
	}
}

func TestDeliveryCloseReasonProducesExactTerminalRecord(t *testing.T) {
	delivery, _ := newBackendTestDelivery(KindVideo, videoTestSource())
	if err := delivery.CloseWithReason(types.MediaDeliveryCloseReplaced); err != nil {
		t.Fatal(err)
	}
	record, ok := delivery.queue.pop()
	if !ok {
		t.Fatal("terminal END record was not queued")
	}
	parsed, err := ParseRecord(record.encoded)
	if err != nil {
		t.Fatal(err)
	}
	var metadata EndMetadata
	err = DecodeStrictJSON(parsed.Metadata, &metadata)
	if err != nil || metadata.Reason != "replaced" {
		t.Fatalf("END metadata = %#v, err = %v", metadata, err)
	}
	select {
	case request := <-delivery.closing:
		if request.code != 4408 || request.failed {
			t.Fatalf("replacement close request = %#v", request)
		}
	default:
		t.Fatal("replacement WebSocket close was not queued")
	}
}

func TestDeliveryFeedbackDetectsBoundedRenderedPTSLag(t *testing.T) {
	delivery, track := newBackendTestDelivery(KindVideo, videoTestSource())
	track.formatSent = true
	delivery.ready = true
	delivery.readyAt = time.Unix(1_700_000_000, 0)
	delivery.recordUnitSent(Record{Type: RecordUnit, Kind: KindVideo, Generation: 1, Sequence: 0, PTS: 0})
	delivery.recordUnitSent(Record{Type: RecordUnit, Kind: KindVideo, Generation: 1, Sequence: 1, PTS: 600_001})

	feedback := feedbackControl{Video: &feedbackKind{
		Generation: 1,
		Received:   2,
		Decoded:    2,
		Rendered:   1,
	}}
	if !delivery.handleFeedback(feedback, delivery.readyAt.Add(time.Second)) {
		t.Fatal("valid progress feedback was rejected")
	}
	select {
	case request := <-delivery.resync:
		if request.kind != KindNone || request.reason != "progress_timeout" {
			t.Fatalf("rendered-PTS resync = %#v", request)
		}
	default:
		t.Fatal("rendered PTS more than 500 ms behind did not request a common resync")
	}
}

func TestOpusPacketDurationBounds(t *testing.T) {
	if duration, ok := opusPacketDurationMicros([]byte{0x00}); !ok || duration != 10000 {
		t.Fatalf("single Opus frame duration = %d/%v", duration, ok)
	}
	if _, ok := opusPacketDurationMicros([]byte{0x03}); ok {
		t.Fatal("truncated code-3 Opus packet accepted")
	}
	if _, ok := opusPacketDurationMicros([]byte{0x03, 49}); ok {
		t.Fatal("over-limit Opus frame count accepted")
	}
}
