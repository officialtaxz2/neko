package webrtc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	pionmedia "github.com/pion/webrtc/v4/pkg/media"
	"github.com/rs/zerolog"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

const webRTCSubscriptionQueueCapacity = 2

type mediaSampleWriter interface {
	WriteSample(pionmedia.Sample) error
}

type Track struct {
	logger  zerolog.Logger
	track   *webrtc.TrackLocalStaticSample
	writer  mediaSampleWriter
	kind    types.MediaKind
	metrics *metrics

	rtcpCh chan []rtcp.Packet

	paused       bool
	provider     types.EncodedMediaProvider
	source       types.MediaSource
	subscription types.MediaSubscription
	streamMu     sync.Mutex
	readerWg     sync.WaitGroup
}

type trackOption func(*Track)

func WithRtcpChan(rtcp chan []rtcp.Packet) trackOption {
	return func(track *Track) {
		track.rtcpCh = rtcp
	}
}

func WithMetrics(metrics *metrics) trackOption {
	return func(track *Track) {
		track.metrics = metrics
	}
}

func rtpCodecFromMediaSource(source types.MediaSource) (codec.RTPCodec, error) {
	rtpCodec, ok := codec.ParseStr(source.Codec.Name)
	if !ok {
		return codec.RTPCodec{}, fmt.Errorf("unknown encoded-media codec: %s", source.Codec.Name)
	}
	if !strings.EqualFold(rtpCodec.Capability.MimeType, source.Codec.MIMEType) {
		return codec.RTPCodec{}, fmt.Errorf("encoded-media MIME type %s does not match codec %s", source.Codec.MIMEType, source.Codec.Name)
	}
	if (source.Kind == types.MediaKindAudio) != rtpCodec.IsAudio() ||
		(source.Kind == types.MediaKindVideo) != rtpCodec.IsVideo() {
		return codec.RTPCodec{}, fmt.Errorf("encoded-media kind %s does not match codec %s", source.Kind, source.Codec.Name)
	}
	return rtpCodec, nil
}

func NewTrack(logger zerolog.Logger, source types.MediaSource, connection *webrtc.PeerConnection, opts ...trackOption) (*Track, error) {
	rtpCodec, err := rtpCodecFromMediaSource(source)
	if err != nil {
		return nil, err
	}

	id := string(source.Kind)
	track, err := webrtc.NewTrackLocalStaticSample(rtpCodec.Capability, id, "stream")
	if err != nil {
		return nil, err
	}

	result := &Track{
		logger: logger.With().Str("id", id).Logger(),
		track:  track,
		writer: track,
		kind:   source.Kind,
		rtcpCh: nil,
	}

	for _, option := range opts {
		option(result)
	}

	sender, err := connection.AddTrack(result.track)
	if err != nil {
		return nil, err
	}

	go result.rtcpReader(sender)
	return result, nil
}

func (track *Track) Shutdown() {
	track.RemoveSource()
	track.readerWg.Wait()
}

func (track *Track) rtcpReader(sender *webrtc.RTPSender) {
	for {
		packets, _, err := sender.ReadRTCP()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
				track.logger.Debug().Msg("track rtcp reader closed")
				return
			}

			track.logger.Warn().Err(err).Msg("failed to read track rtcp")
			continue
		}

		if track.rtcpCh != nil {
			track.rtcpCh <- packets
		}
	}
}

func (track *Track) subscriptionReader(subscription types.MediaSubscription) {
	defer track.readerWg.Done()
	for event := range subscription.Events() {
		switch event.Type {
		case types.MediaEventTypeFormat:
			track.streamMu.Lock()
			if track.subscription == subscription {
				track.source = types.CloneMediaSource(event.Source)
			}
			track.streamMu.Unlock()
		case types.MediaEventTypeUnit:
			err := track.writer.WriteSample(pionmedia.Sample{
				Data:      event.Unit.Data,
				Duration:  event.Unit.Duration,
				Timestamp: event.Unit.CapturedAt,
			})
			if err != nil && !errors.Is(err, io.ErrClosedPipe) {
				track.logger.Warn().Err(err).Msg("failed to write sample to track")
			}
		}
	}
}

func (track *Track) OnMediaSubscriptionDrop(drop types.MediaDrop) {
	if track.metrics != nil && drop.Reason == "queue_full" {
		track.metrics.IncTrackDroppedSample(string(track.kind))
	}
	track.logger.Trace().Str("reason", drop.Reason).Msg("dropping sample: media subscription queue full")
}

func (track *Track) subscriptionRequest(sourceID string) types.SourceSubscriptionRequest {
	return types.SourceSubscriptionRequest{
		Kind:           track.kind,
		Selector:       types.MediaSelector{ID: sourceID},
		QueueCapacity:  webRTCSubscriptionQueueCapacity,
		OverflowPolicy: types.MediaOverflowDropNewest,
		Backend:        "webrtc",
		Observer:       track,
	}
}

func (track *Track) SetSource(provider types.EncodedMediaProvider, selector types.MediaSelector) (bool, error) {
	source, ok := types.SelectMediaSource(provider.Sources(track.kind), selector)
	if !ok {
		return false, types.ErrMediaSourceNotFound
	}

	track.streamMu.Lock()
	defer track.streamMu.Unlock()
	if track.provider == provider && track.source.ID == source.ID {
		return false, nil
	}

	if track.paused {
		if track.subscription != nil {
			if err := track.subscription.Switch(context.Background(), types.MediaSelector{ID: source.ID}); err != nil {
				return false, err
			}
		}
		track.provider = provider
		track.source = source
		return true, nil
	}

	if track.subscription != nil {
		if err := track.subscription.Switch(context.Background(), types.MediaSelector{ID: source.ID}); err != nil {
			return false, err
		}
		track.provider = provider
		track.source = source
		return true, nil
	}

	subscription, err := provider.Subscribe(context.Background(), track.subscriptionRequest(source.ID))
	if err != nil {
		return false, err
	}

	track.provider = provider
	track.source = source
	track.subscription = subscription
	track.readerWg.Add(1)
	go track.subscriptionReader(subscription)
	return true, nil
}

func (track *Track) RemoveSource() {
	track.streamMu.Lock()
	subscription := track.subscription
	track.subscription = nil
	track.provider = nil
	track.source = types.MediaSource{}
	track.streamMu.Unlock()

	if subscription != nil {
		if err := subscription.Close(); err != nil {
			track.logger.Warn().Err(err).Msg("failed to close media subscription")
		}
	}
}

func (track *Track) Source() (types.MediaSource, bool) {
	track.streamMu.Lock()
	defer track.streamMu.Unlock()
	if track.source.ID == "" {
		return types.MediaSource{}, false
	}
	if track.subscription != nil {
		return track.subscription.Source(), true
	}
	return types.CloneMediaSource(track.source), true
}

func (track *Track) SetPaused(paused bool) {
	track.streamMu.Lock()
	if track.paused == paused {
		track.streamMu.Unlock()
		return
	}

	if track.subscription != nil {
		if err := track.subscription.SetPaused(paused); err != nil {
			track.streamMu.Unlock()
			track.logger.Warn().Err(err).Msg("failed to change media subscription paused state")
			return
		}
		track.paused = paused
		track.streamMu.Unlock()
		return
	}

	if paused || track.provider == nil || track.source.ID == "" {
		track.paused = paused
		track.streamMu.Unlock()
		return
	}

	provider := track.provider
	source := track.source
	track.streamMu.Unlock()

	subscription, err := provider.Subscribe(context.Background(), track.subscriptionRequest(source.ID))
	if err != nil {
		track.logger.Warn().Err(err).Msg("failed to resume media subscription")
		return
	}

	track.streamMu.Lock()
	if track.subscription != nil || track.paused == paused {
		track.streamMu.Unlock()
		_ = subscription.Close()
		return
	}
	track.subscription = subscription
	track.paused = false
	track.readerWg.Add(1)
	track.streamMu.Unlock()
	go track.subscriptionReader(subscription)
}

func (track *Track) Paused() bool {
	track.streamMu.Lock()
	defer track.streamMu.Unlock()
	return track.paused
}
