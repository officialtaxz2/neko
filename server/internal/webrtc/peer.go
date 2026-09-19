package webrtc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/pion/interceptor/pkg/cc"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/internal/webrtc/payload"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/event"
	"github.com/m1k1o/neko/server/pkg/utils"
)

type WebRTCPeerCtx struct {
	mu           sync.Mutex
	id           string
	logger       zerolog.Logger
	session      types.Session
	done         chan struct{}
	doneOnce     sync.Once
	closeOnce    sync.Once
	initialOffer *webrtc.SessionDescription
	metrics      *metrics
	connection   *webrtc.PeerConnection
	// bandwidth estimator
	estimator     cc.BandwidthEstimator
	estimateTrend *utils.TrendDetector
	// backend-neutral encoded-media provider
	media types.EncodedMediaProvider
	// tracks & channels
	audioTrack  *Track
	videoTrack  *Track
	dataChannel *webrtc.DataChannel
	rtcpChannel chan []rtcp.Packet
	// config
	iceTrickle      bool
	estimatorConfig config.WebRTCEstimator
	paused          bool
	videoAuto       bool
	videoDisabled   bool
	audioDisabled   bool
}

//
// connection
//

func (peer *WebRTCPeerCtx) CreateOffer(ICERestart bool) (*webrtc.SessionDescription, error) {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	offer, err := peer.connection.CreateOffer(&webrtc.OfferOptions{
		ICERestart: ICERestart,
	})
	if err != nil {
		return nil, err
	}

	return peer.setLocalDescription(offer)
}

func (peer *WebRTCPeerCtx) CreateAnswer() (*webrtc.SessionDescription, error) {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	answer, err := peer.connection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	return peer.setLocalDescription(answer)
}

func (peer *WebRTCPeerCtx) setLocalDescription(description webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	if !peer.iceTrickle {
		// Create channel that is blocked until ICE Gathering is complete
		gatherComplete := webrtc.GatheringCompletePromise(peer.connection)

		if err := peer.connection.SetLocalDescription(description); err != nil {
			return nil, err
		}

		<-gatherComplete
	} else {
		if err := peer.connection.SetLocalDescription(description); err != nil {
			return nil, err
		}
	}

	return peer.connection.LocalDescription(), nil
}

func (peer *WebRTCPeerCtx) SetRemoteDescription(desc webrtc.SessionDescription) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	return peer.connection.SetRemoteDescription(desc)
}

func (peer *WebRTCPeerCtx) SetCandidate(candidate webrtc.ICECandidateInit) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	return peer.connection.AddICECandidate(candidate)
}

// TODO: Add shutdown function?
func (peer *WebRTCPeerCtx) Destroy() {
	peer.closeOnce.Do(func() {
		peer.mu.Lock()
		defer peer.mu.Unlock()

		var err error
		if peer.connection.ConnectionState() != webrtc.PeerConnectionStateClosed {
			err = peer.connection.Close()
		}
		peer.logger.Err(err).Msg("peer connection destroyed")
	})
}

func (peer *WebRTCPeerCtx) ID() string {
	return peer.id
}

func (peer *WebRTCPeerCtx) SessionID() string {
	return peer.session.ID()
}

func (peer *WebRTCPeerCtx) Backend() string {
	return "webrtc"
}

func (peer *WebRTCPeerCtx) Close() error {
	peer.Destroy()
	return nil
}

func (peer *WebRTCPeerCtx) Done() <-chan struct{} {
	return peer.done
}

func (peer *WebRTCPeerCtx) estimatorReader() {
	conf := peer.estimatorConfig

	// if estimator is not in debug mode, use a nop logger
	var debugLogger zerolog.Logger
	if conf.Debug {
		debugLogger = peer.logger.With().Str("component", "estimator").Logger().Level(zerolog.DebugLevel)
	} else {
		debugLogger = zerolog.Nop()
	}

	// if estimator is disabled, do nothing
	if peer.estimator == nil {
		return
	}

	// use a ticker to get current client target bitrate
	ticker := time.NewTicker(conf.ReadInterval)
	defer ticker.Stop()

	state := newEstimatorObservationState(time.Now())

	for range ticker.C {
		now := time.Now()
		targetBitrate := peer.estimator.GetTargetBitrate()
		peer.metrics.SetReceiverEstimatedTargetBitrate(float64(targetBitrate))

		// if peer connection is closed, stop reading
		if peer.connection.ConnectionState() == webrtc.PeerConnectionStateClosed {
			break
		}

		// if estimation or video is disabled, do nothing
		if !peer.videoAuto || peer.videoDisabled || peer.paused || conf.Passive {
			continue
		}

		// get trend direction to decide if we should upgrade or downgrade
		peer.estimateTrend.AddValue(int64(targetBitrate))
		direction := peer.estimateTrend.GetDirection()

		// Get the current measured and configured tier rates. The estimator target
		// covers the peer's complete RTP delivery, so the downgrade reference also
		// includes this peer's audio plus a small transport reserve.
		stream, ok := peer.videoTrack.Source()
		if !ok {
			debugLogger.Warn().Msg("looks like we don't have a stream yet, skipping bitrate estimation")
			continue
		}

		// if stream bitrate is 0, we need to wait for some time until we get a valid value
		streamId, streamBitrate := stream.ID, stream.Bitrate
		if streamBitrate == 0 {
			debugLogger.Warn().Msg("looks like stream bitrate is 0, we need to wait for some time")
			continue
		}

		audioBitrate := uint64(0)
		if !peer.audioDisabled && peer.audioTrack != nil {
			if audioStream, ok := peer.audioTrack.Source(); ok {
				audioBitrate = audioStream.Bitrate
			}
		}

		downgradeReferenceBitrate := deliveryBitrateReference(
			streamBitrate,
			stream.NominalBitrate,
			audioBitrate,
			conf.TransportReserve,
		)
		decision := state.observe(
			now,
			direction,
			targetBitrate,
			downgradeReferenceBitrate,
			conf,
		)

		debugLogger.Info().
			Float64("current_video_diff", bitrateRatio(targetBitrate, streamBitrate)).
			Float64("current_delivery_diff", bitrateRatio(targetBitrate, downgradeReferenceBitrate)).
			Int("target_bitrate", targetBitrate).
			Uint64("stream_bitrate", streamBitrate).
			Uint64("stream_nominal_bitrate", stream.NominalBitrate).
			Uint64("audio_bitrate", audioBitrate).
			Uint64("downgrade_reference_bitrate", downgradeReferenceBitrate).
			Uint64("downgrade_floor_bitrate", downgradeBitrateFloor(downgradeReferenceBitrate, conf.DowngradeDeficitThreshold)).
			Float64("transport_reserve", conf.TransportReserve).
			Float64("downgrade_deficit_threshold", conf.DowngradeDeficitThreshold).
			Bool("insufficient", decision.insufficient).
			Str("direction", direction.String()).
			Msg("got bitrate from estimator")

		if decision.stalled {
			debugLogger.Warn().
				Time("stalled_since", state.stalledSince).
				Msg("neutral estimate remains materially below the current delivery requirement")
		}

		if decision.downgrade {
			err := peer.SetVideo(types.PeerVideoRequest{
				Selector: &types.StreamSelector{
					ID:   streamId,
					Type: types.StreamSelectorTypeLower,
				},
			})
			if err != nil && err != types.ErrWebRTCStreamNotFound {
				peer.logger.Warn().Err(err).Msg("failed to downgrade video stream")
			}
			state.markDowngrade(now)

			if err == types.ErrWebRTCStreamNotFound {
				debugLogger.Info().Msg("looks like we are already on the lowest stream")
			} else {
				debugLogger.Info().Msg("downgraded video stream")
			}
			continue
		}

		if !decision.upgradeReady {
			continue
		}

		// Resolve the next stream before applying the upgrade threshold. When the
		// next stream declares a nominal rate, compare the estimate with the rate
		// we are about to select instead of the content-dependent current rate.
		upgradeStream, ok := types.SelectMediaSource(peer.media.Sources(types.MediaKindVideo), types.MediaSelector{
			ID:   streamId,
			Type: types.MediaSelectorTypeHigher,
		})
		if !ok {
			debugLogger.Info().Msg("looks like we are already on the highest stream")
			continue
		}

		upgradeNominalBitrate := upgradeStream.NominalBitrate
		upgradeReferenceBitrate := deliveryBitrateReference(
			referenceBitrateForUpgrade(streamBitrate, upgradeNominalBitrate),
			0,
			audioBitrate,
			conf.TransportReserve,
		)
		if !estimatedBitrateSupportsUpgrade(targetBitrate, upgradeReferenceBitrate, conf.UpgradeDiffThreshold) {
			debugLogger.Debug().
				Float64("current_delivery_diff", bitrateRatio(targetBitrate, downgradeReferenceBitrate)).
				Float64("upgrade_diff", float64(targetBitrate)/float64(upgradeReferenceBitrate)).
				Str("upgrade_stream_id", upgradeStream.ID).
				Uint64("upgrade_reference_bitrate", upgradeReferenceBitrate).
				Bool("nominal_reference", upgradeNominalBitrate != 0).
				Float64("threshold", conf.UpgradeDiffThreshold).
				Msgf("looks like we don't have enough bitrate to accomodate higher stream, " +
					"therefore we should wait for some more time")
			continue
		}

		err := peer.SetVideo(types.PeerVideoRequest{
			Selector: &types.StreamSelector{
				ID: upgradeStream.ID,
			},
		})
		if err != nil && err != types.ErrWebRTCStreamNotFound {
			peer.logger.Warn().Err(err).Msg("failed to upgrade video stream")
		}
		state.markUpgrade(now)

		if err != types.ErrWebRTCStreamNotFound {
			debugLogger.Info().Msg("upgraded video stream")
		}
	}
}

type estimatorDecision struct {
	downgrade    bool
	upgradeReady bool
	insufficient bool
	stalled      bool
}

type estimatorObservationState struct {
	stableSince       time.Time
	unstableSince     time.Time
	stalledSince      time.Time
	lastUpgradeTime   time.Time
	lastDowngradeTime time.Time
}

func newEstimatorObservationState(now time.Time) estimatorObservationState {
	stableSince, unstableSince, stalledSince := initialEstimatorObservationTimes(now)
	return estimatorObservationState{
		stableSince:   stableSince,
		unstableSince: unstableSince,
		stalledSince:  stalledSince,
	}
}

func (state *estimatorObservationState) observe(
	now time.Time,
	direction utils.TrendDirection,
	targetBitrate int,
	currentReferenceBitrate uint64,
	conf config.WebRTCEstimator,
) estimatorDecision {
	insufficient := estimatedBitrateRequiresDowngrade(
		targetBitrate,
		currentReferenceBitrate,
		conf.DowngradeDeficitThreshold,
	)

	if direction != utils.TrendDirectionNeutral || !insufficient {
		state.stalledSince = now
	}
	stalled := direction == utils.TrendDirectionNeutral &&
		insufficient &&
		now.Sub(state.stalledSince) > conf.StalledDuration

	// A downward trend is not itself proof that the current tier no longer
	// fits. Conversely, an insufficient estimate must not count toward the
	// stable window used for an opposite-direction upgrade.
	if direction == utils.TrendDirectionDownward || insufficient {
		state.stableSince = now
	}

	congested := insufficient && (direction == utils.TrendDirectionDownward || stalled)
	if congested {
		if now.Sub(state.lastDowngradeTime) < conf.DowngradeBackoff ||
			now.Sub(state.unstableSince) < conf.UnstableDuration {
			return estimatorDecision{insufficient: true, stalled: stalled}
		}
		return estimatorDecision{downgrade: true, insufficient: true, stalled: stalled}
	}

	state.unstableSince = now
	if direction == utils.TrendDirectionDownward ||
		now.Sub(state.lastUpgradeTime) < conf.UpgradeBackoff ||
		now.Sub(state.stableSince) < conf.StableDuration {
		return estimatorDecision{insufficient: insufficient, stalled: stalled}
	}

	return estimatorDecision{upgradeReady: true, insufficient: insufficient, stalled: stalled}
}

func (state *estimatorObservationState) markDowngrade(now time.Time) {
	state.lastDowngradeTime = now
}

func (state *estimatorObservationState) markUpgrade(now time.Time) {
	state.lastUpgradeTime = now
}

func initialEstimatorObservationTimes(now time.Time) (stableSince, unstableSince, stalledSince time.Time) {
	return now, now, now
}

func deliveryBitrateReference(measuredVideoBitrate, nominalVideoBitrate, audioBitrate uint64, transportReserve float64) uint64 {
	videoBitrate := measuredVideoBitrate
	if nominalVideoBitrate > videoBitrate {
		videoBitrate = nominalVideoBitrate
	}

	mediaBitrate := videoBitrate + audioBitrate
	if transportReserve <= 0 {
		return mediaBitrate
	}
	return uint64(math.Ceil(float64(mediaBitrate) * (1 + transportReserve)))
}

func downgradeBitrateFloor(referenceBitrate uint64, toleratedDeficit float64) uint64 {
	if referenceBitrate == 0 {
		return 0
	}
	if toleratedDeficit < 0 {
		return referenceBitrate
	}
	if toleratedDeficit >= 1 {
		return 0
	}
	return uint64(math.Ceil(float64(referenceBitrate) * (1 - toleratedDeficit)))
}

func estimatedBitrateRequiresDowngrade(targetBitrate int, referenceBitrate uint64, toleratedDeficit float64) bool {
	if toleratedDeficit < 0 {
		return true
	}
	if referenceBitrate == 0 {
		return false
	}
	return targetBitrate < int(downgradeBitrateFloor(referenceBitrate, toleratedDeficit))
}

func bitrateRatio(targetBitrate int, referenceBitrate uint64) float64 {
	if referenceBitrate == 0 {
		return 0
	}
	return float64(targetBitrate) / float64(referenceBitrate)
}

func referenceBitrateForUpgrade(measuredBitrate, nominalBitrate uint64) uint64 {
	if nominalBitrate != 0 {
		return nominalBitrate
	}
	return measuredBitrate
}

func estimatedBitrateSupportsUpgrade(targetBitrate int, referenceBitrate uint64, threshold float64) bool {
	if threshold < 0 {
		return true
	}
	if referenceBitrate == 0 {
		return false
	}

	return float64(targetBitrate)/float64(referenceBitrate) >= 1+threshold
}

func (peer *WebRTCPeerCtx) SetPaused(isPaused bool) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	peer.videoTrack.SetPaused(isPaused || peer.videoDisabled)
	peer.audioTrack.SetPaused(isPaused || peer.audioDisabled)

	peer.logger.Info().Bool("is_paused", isPaused).Msg("set paused")
	peer.paused = isPaused

	return nil
}

func (peer *WebRTCPeerCtx) Paused() bool {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	return peer.paused
}

//
// video
//

func (peer *WebRTCPeerCtx) SetVideo(r types.PeerVideoRequest) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	modified := false

	// video disabled
	if r.Disabled != nil {
		disabled := *r.Disabled

		// update only if changed
		if peer.videoDisabled != disabled {
			peer.videoDisabled = disabled
			peer.videoTrack.SetPaused(disabled || peer.paused)

			peer.logger.Info().Bool("disabled", disabled).Msg("set video disabled")
			modified = true
		}
	}

	// video selector
	if r.Selector != nil {
		selector := *r.Selector

		// get requested video stream from selector
		// Resolve and set the encoded source behind the backend-neutral provider.
		changed, err := peer.videoTrack.SetSource(peer.media, selector)
		if err != nil {
			if errors.Is(err, types.ErrMediaSourceNotFound) {
				return types.ErrWebRTCStreamNotFound
			}
			return err
		}

		// update only if stream changed
		if changed {
			stream, _ := peer.videoTrack.Source()
			videoID := stream.ID
			peer.metrics.SetVideoID(videoID)

			peer.logger.Info().Str("video_id", videoID).Msg("set video")
			modified = true
		}
	}

	// video auto
	if r.Auto != nil {
		videoAuto := *r.Auto

		if peer.estimator == nil || peer.estimatorConfig.Passive {
			peer.logger.Warn().Msg("estimator is disabled or in passive mode, cannot change video auto")
			videoAuto = false // ensure video auto is disabled
		}

		// update only if video auto changed
		if peer.videoAuto != videoAuto {
			peer.videoAuto = videoAuto

			peer.logger.Info().Bool("video_auto", videoAuto).Msg("set video auto")
			modified = true
		}
	}

	// send video signal if modified
	if modified {
		go func() {
			// in goroutine because of mutex and we don't want to block
			resolvedVideo := peer.Video()
			peer.session.Send(event.SIGNAL_VIDEO, resolvedVideo)
		}()
	}

	return nil
}

func (peer *WebRTCPeerCtx) Video() types.PeerVideo {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	// get current video stream ID
	ID := ""
	stream, ok := peer.videoTrack.Source()
	if ok {
		ID = stream.ID
	}

	return types.PeerVideo{
		Disabled: peer.videoDisabled,
		ID:       ID,
		Video:    ID, // TODO: Remove, used for backward compatibility
		Auto:     peer.videoAuto,
	}
}

//
// audio
//

func (peer *WebRTCPeerCtx) SetAudio(r types.PeerAudioRequest) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	modified := false

	// audio disabled
	if r.Disabled != nil {
		disabled := *r.Disabled

		// update only if changed
		if peer.audioDisabled != disabled {
			peer.audioDisabled = disabled
			peer.audioTrack.SetPaused(disabled || peer.paused)

			peer.logger.Info().Bool("disabled", disabled).Msg("set audio disabled")
			modified = true
		}
	}

	// send video signal if modified
	if modified {
		go func() {
			// in goroutine because of mutex and we don't want to block
			resolvedAudio := peer.Audio()
			peer.session.Send(event.SIGNAL_AUDIO, resolvedAudio)
		}()
	}

	return nil
}

func (peer *WebRTCPeerCtx) Audio() types.PeerAudio {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	return types.PeerAudio{
		Disabled: peer.audioDisabled,
	}
}

//
// data channel
//

func (peer *WebRTCPeerCtx) SendCursorPosition(x, y int) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	// do not send cursor position to host
	if peer.session.IsHost() {
		return nil
	}

	header := payload.Header{
		Event:  payload.OP_CURSOR_POSITION,
		Length: 7,
	}

	data := payload.CursorPosition{
		X: uint16(x),
		Y: uint16(y),
	}

	buffer := &bytes.Buffer{}

	if err := binary.Write(buffer, binary.BigEndian, header); err != nil {
		return err
	}

	if err := binary.Write(buffer, binary.BigEndian, data); err != nil {
		return err
	}

	return peer.dataChannel.Send(buffer.Bytes())
}

func (peer *WebRTCPeerCtx) SendCursorImage(cur *types.CursorImage, img []byte) error {
	peer.mu.Lock()
	defer peer.mu.Unlock()

	header := payload.Header{
		Event:  payload.OP_CURSOR_IMAGE,
		Length: uint16(11 + len(img)),
	}

	data := payload.CursorImage{
		Width:  cur.Width,
		Height: cur.Height,
		Xhot:   cur.Xhot,
		Yhot:   cur.Yhot,
	}

	buffer := &bytes.Buffer{}

	if err := binary.Write(buffer, binary.BigEndian, header); err != nil {
		return err
	}

	if err := binary.Write(buffer, binary.BigEndian, data); err != nil {
		return err
	}

	if err := binary.Write(buffer, binary.BigEndian, img); err != nil {
		return err
	}

	return peer.dataChannel.Send(buffer.Bytes())
}
