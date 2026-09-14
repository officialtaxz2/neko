package mediaws

import (
	"errors"
	"math"
	"time"

	"github.com/gorilla/websocket"
	"github.com/m1k1o/neko/server/pkg/types"
)

func (delivery *Delivery) writeLoop() closeRequest {
	ping := time.NewTicker(PingPeriod)
	defer ping.Stop()

	for {
		if record, ok := delivery.queue.pop(); ok {
			if err := delivery.writeRecord(record); err != nil {
				return delivery.closeResult(backendClose("write_failed"))
			}
			if !delivery.recordOutbound(len(record.encoded), time.Now()) {
				delivery.close(backpressureClose("outbound_rate"))
			}
			continue
		}

		select {
		case request := <-delivery.closing:
			for {
				record, ok := delivery.queue.pop()
				if !ok {
					break
				}
				if err := delivery.writeRecord(record); err != nil {
					return request
				}
			}
			_ = delivery.connection.SetWriteDeadline(time.Now().Add(CloseTimeout))
			_ = delivery.connection.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(request.code, request.reason), time.Now().Add(CloseTimeout))
			return request
		case <-delivery.queue.notify:
		case <-ping.C:
			if err := delivery.writePing(); err != nil {
				return backendClose("ping_failed")
			}
		}
	}
}

func (delivery *Delivery) closeResult(fallback closeRequest) closeRequest {
	delivery.mu.Lock()
	defer delivery.mu.Unlock()
	if delivery.requestedClose != nil {
		return *delivery.requestedClose
	}
	return fallback
}

func (delivery *Delivery) writeRecord(record queuedRecord) error {
	started := time.Now()
	if err := delivery.connection.SetWriteDeadline(started.Add(WriteTimeout)); err != nil {
		return err
	}
	if err := delivery.connection.WriteMessage(websocket.BinaryMessage, record.encoded); err != nil {
		return err
	}
	mediaWebSocketWriteSeconds.Observe(time.Since(started).Seconds())
	mediaWebSocketRecords.WithLabelValues(metricKind(record.record.Kind), metricRecordType(record.record.Type)).Inc()
	if record.record.Type == RecordUnit {
		mediaWebSocketBytes.WithLabelValues(metricKind(record.record.Kind)).Add(float64(record.payloadBytes))
		delivery.recordUnitSent(record.record)
	}
	return nil
}

func (delivery *Delivery) recordUnitSent(record Record) {
	delivery.mu.Lock()
	defer delivery.mu.Unlock()
	track := delivery.tracks[record.Kind]
	if track == nil || track.generation != record.Generation {
		return
	}
	track.sentUnits = record.Sequence + 1
	track.sentTimeline = append(track.sentTimeline, sentTimestamp{sequence: record.Sequence, pts: uint64(record.PTS)})
	if len(track.sentTimeline) > SentTimestampHistory {
		copy(track.sentTimeline, track.sentTimeline[len(track.sentTimeline)-SentTimestampHistory:])
		track.sentTimeline = track.sentTimeline[:SentTimestampHistory]
	}
	delivery.sentSinceProgress = true
}

func (delivery *Delivery) writePing() error {
	deadline := time.Now().Add(WriteTimeout)
	if err := delivery.connection.SetWriteDeadline(deadline); err != nil {
		return err
	}
	return delivery.connection.WriteControl(websocket.PingMessage, nil, deadline)
}

func (delivery *Delivery) recordOutbound(bytes int, now time.Time) bool {
	delivery.mu.Lock()
	cutoff := now.Add(-OutboundRateWindow)
	kept := delivery.outbound[:0]
	total := bytes
	for _, sample := range delivery.outbound {
		if sample.at.Before(cutoff) {
			continue
		}
		kept = append(kept, sample)
		total += sample.bytes
	}
	delivery.outbound = append(kept, outboundSample{at: now, bytes: bytes})
	delivery.mu.Unlock()
	return total <= MaximumOutboundBytesPerSec*int(OutboundRateWindow/time.Second)
}

func (delivery *Delivery) readControls() {
	for {
		messageType, raw, err := delivery.connection.ReadMessage()
		if err != nil {
			if delivery.ctx.Err() != nil {
				return
			}
			var closeError *websocket.CloseError
			if errors.As(err, &closeError) {
				request := normalClose()
				request.skipEnd = true
				delivery.close(request)
			} else {
				request := backendClose("read_failed")
				request.skipEnd = true
				delivery.close(request)
			}
			return
		}
		if messageType != websocket.TextMessage {
			delivery.close(protocolClose("binary_client_message"))
			return
		}
		now := time.Now()
		if !delivery.allowControl(now) {
			delivery.close(rateClose("control_rate"))
			return
		}
		control, err := parseControl(raw)
		if err != nil {
			delivery.close(protocolClose("invalid_control"))
			return
		}
		switch control.Type {
		case "ready":
			if !delivery.handleReady(control.Ready, now) {
				delivery.close(protocolClose("invalid_ready"))
				return
			}
		case "feedback":
			if delivery.feedbackRateExceeded(now) {
				delivery.close(rateClose("feedback_rate"))
				return
			}
			if !delivery.handleFeedback(control.Feedback, now) {
				delivery.close(protocolClose("invalid_feedback"))
				return
			}
		case "resync":
			if !delivery.handleClientResync(control.Resync, now) {
				return
			}
		case "stop":
			delivery.close(normalClose())
			return
		}
	}
}

func (delivery *Delivery) allowControl(now time.Time) bool {
	delivery.mu.Lock()
	defer delivery.mu.Unlock()
	if now.After(delivery.controlRefillAt) {
		delivery.controlTokens += now.Sub(delivery.controlRefillAt).Seconds() * ControlRecordsPerSecond
		if delivery.controlTokens > ControlRecordBurst {
			delivery.controlTokens = ControlRecordBurst
		}
		delivery.controlRefillAt = now
	}
	if delivery.controlTokens < 1 {
		return false
	}
	delivery.controlTokens--
	return true
}

func (delivery *Delivery) handleReady(ready readyControl, now time.Time) bool {
	delivery.mu.Lock()
	if delivery.ready || delivery.ctx.Err() != nil {
		delivery.mu.Unlock()
		return false
	}
	audio, hasAudio := delivery.tracks[KindAudio]
	video, hasVideo := delivery.tracks[KindVideo]
	if !hasVideo || !video.formatSent || !readyMatchesTrack(ready.Video, video, KindVideo) || hasAudio != (ready.Audio != nil) {
		delivery.mu.Unlock()
		return false
	}
	if hasAudio && (!audio.formatSent || !readyMatchesTrack(ready.Audio, audio, KindAudio)) {
		delivery.mu.Unlock()
		return false
	}
	delivery.ready = true
	delivery.readyAt = now
	delivery.lastProgressAt = now
	delivery.mu.Unlock()

	if !delivery.lease.SetState(types.MediaDeliveryStateActive) {
		delivery.close(delivery.invalidLeaseClose())
		return true
	}
	delivery.setConnectionState("active")
	return true
}

func readyMatchesTrack(ready *readyKind, track *deliveryTrack, kind Kind) bool {
	if ready == nil || ready.Generation != track.generation {
		return false
	}
	if kind == KindAudio {
		return ready.Codec == "opus" && ready.SampleRate == track.source.Codec.ClockRate && ready.Channels == track.source.Codec.Channels
	}
	return ready.Codec == "vp8" && ready.CodedWidth == track.source.Width && ready.CodedHeight == track.source.Height
}

func (delivery *Delivery) handleFeedback(feedback feedbackControl, now time.Time) bool {
	requestProgressResync := false
	skewResync := false
	delivery.mu.Lock()
	if !delivery.ready {
		delivery.mu.Unlock()
		return false
	}
	audio, hasAudio := delivery.tracks[KindAudio]
	video, hasVideo := delivery.tracks[KindVideo]
	if !hasVideo || hasAudio != (feedback.Audio != nil) || feedback.Video == nil || !feedbackMatchesTrack(feedback.Video, video) || (hasAudio && !feedbackMatchesTrack(feedback.Audio, audio)) {
		delivery.mu.Unlock()
		return false
	}
	progressed := feedback.Video.Rendered > video.lastRendered
	lagMicros, lagKnown := renderedLagMicros(video, feedback.Video.Rendered)
	requestProgressResync = lagKnown && lagMicros > 500_000
	outstanding := feedback.Video.Rendered < video.sentUnits
	video.lastRendered = feedback.Video.Rendered
	maxLag := feedback.Video.BufferedMS
	if hasAudio {
		progressed = progressed || feedback.Audio.Rendered > audio.lastRendered
		if audioLag, known := renderedLagMicros(audio, feedback.Audio.Rendered); known && audioLag > lagMicros {
			lagMicros = audioLag
			lagKnown = true
			requestProgressResync = audioLag > 500_000
		}
		outstanding = outstanding || feedback.Audio.Rendered < audio.sentUnits
		audio.lastRendered = feedback.Audio.Rendered
		if feedback.Audio.BufferedMS > maxLag {
			maxLag = feedback.Audio.BufferedMS
		}
	}
	delivery.lastFeedbackAt = now
	if progressed {
		delivery.lastProgressAt = now
		delivery.progressRecoveryDeadline = time.Time{}
		delivery.sentSinceProgress = outstanding
	}
	absSkew := int64(feedback.AVSkewMS)
	if absSkew < 0 {
		absSkew = -absSkew
	}
	if absSkew > 200 {
		if delivery.skewExceededAt.IsZero() {
			delivery.skewExceededAt = now
		} else if now.Sub(delivery.skewExceededAt) >= time.Second {
			requestProgressResync = true
			skewResync = true
			delivery.skewExceededAt = time.Time{}
		}
	} else {
		delivery.skewExceededAt = time.Time{}
	}
	delivery.mu.Unlock()
	mediaWebSocketClientLagMS.Observe(float64(maxLag))
	mediaWebSocketClientLagMS.Observe(math.Abs(float64(feedback.AVSkewMS)))
	if lagKnown {
		mediaWebSocketClientLagMS.Observe(float64(lagMicros) / 1000)
	}
	if requestProgressResync {
		reason := "progress_timeout"
		if skewResync {
			reason = "av_skew"
		}
		delivery.requestResync(KindNone, reason)
	}
	return true
}

func (delivery *Delivery) feedbackRateExceeded(now time.Time) bool {
	delivery.mu.Lock()
	defer delivery.mu.Unlock()
	if now.After(delivery.feedbackRefillAt) {
		delivery.feedbackTokens += now.Sub(delivery.feedbackRefillAt).Seconds() * FeedbackRecordsPerSecond
		if delivery.feedbackTokens > FeedbackRecordBurst {
			delivery.feedbackTokens = FeedbackRecordBurst
		}
		delivery.feedbackRefillAt = now
	}
	if delivery.feedbackTokens < 1 {
		return true
	}
	delivery.feedbackTokens--
	return false
}

func feedbackMatchesTrack(feedback *feedbackKind, track *deliveryTrack) bool {
	return feedback != nil && feedback.Generation == track.generation && feedback.Rendered >= track.lastRendered && feedback.Received <= track.sentUnits
}

func renderedLagMicros(track *deliveryTrack, rendered uint64) (uint64, bool) {
	if rendered == 0 || rendered > track.sentUnits || len(track.sentTimeline) == 0 {
		return 0, false
	}
	sequence := rendered - 1
	latest := track.sentTimeline[len(track.sentTimeline)-1]
	earliest := track.sentTimeline[0]
	if sequence < earliest.sequence && latest.pts >= earliest.pts {
		// The exact rendered stamp aged out of the fixed history. The retained
		// span is still a safe lower bound for how far that report trails.
		return latest.pts - earliest.pts, true
	}
	for index := len(track.sentTimeline) - 1; index >= 0; index-- {
		stamp := track.sentTimeline[index]
		if stamp.sequence != sequence {
			continue
		}
		if latest.pts < stamp.pts {
			return 0, false
		}
		return latest.pts - stamp.pts, true
	}
	return 0, false
}

func (delivery *Delivery) handleClientResync(resync resyncControl, now time.Time) bool {
	delivery.mu.Lock()
	if !delivery.ready {
		delivery.mu.Unlock()
		delivery.close(protocolClose("resync_before_ready"))
		return false
	}
	cutoff := now.Add(-time.Minute)
	kept := delivery.clientResyncs[:0]
	for _, observed := range delivery.clientResyncs {
		if !observed.Before(cutoff) {
			kept = append(kept, observed)
		}
	}
	if len(kept) >= ClientResyncsPerMinute {
		delivery.clientResyncs = kept
		delivery.mu.Unlock()
		delivery.close(rateClose("resync_rate"))
		return false
	}
	kind := KindNone
	switch resync.Kind {
	case "audio":
		kind = KindAudio
	case "video":
		kind = KindVideo
	}
	track := delivery.tracks[kind]
	if kind == KindNone {
		track = delivery.tracks[KindVideo]
	}
	if track == nil || resync.Generation != track.generation {
		delivery.mu.Unlock()
		delivery.close(protocolClose("resync_generation"))
		return false
	}
	delivery.clientResyncs = append(kept, now)
	delivery.mu.Unlock()
	delivery.logger.Info().
		Str("resync_kind", resync.Kind).
		Str("resync_reason", resync.Reason).
		Msg("client requested media resync")
	if kind == KindAudio {
		kind = KindNone
	}
	delivery.requestResync(kind, "client_"+resync.Reason)
	return true
}

func (delivery *Delivery) handleResyncs() {
	for {
		select {
		case <-delivery.ctx.Done():
			return
		case request := <-delivery.resync:
			for {
				select {
				case pending := <-delivery.resync:
					if pending.kind == KindNone {
						request.kind = KindNone
					}
					if pending.reason == "server_overflow" {
						request.reason = pending.reason
					}
				default:
					goto drained
				}
			}
		drained:
			resyncKind := "all"
			if request.kind != KindNone {
				resyncKind = metricKind(request.kind)
			}
			delivery.logger.Info().
				Str("resync_kind", resyncKind).
				Str("resync_reason", request.reason).
				Msg("processing media resync")
			if !delivery.allowResync(request.reason, time.Now()) {
				delivery.close(backpressureClose("resync_limit"))
				return
			}
			if err := delivery.performResync(request.kind, request.reason); err != nil {
				delivery.close(backendClose("resync_failed"))
				return
			}
		}
	}
}

func (delivery *Delivery) allowResync(reason string, now time.Time) bool {
	delivery.mu.Lock()
	defer delivery.mu.Unlock()
	cutoff := now.Add(-ResyncWindow)
	kept := delivery.resyncs[:0]
	for _, observed := range delivery.resyncs {
		if !observed.Before(cutoff) {
			kept = append(kept, observed)
		}
	}
	if len(kept) >= MaximumResyncsPerWindow-1 {
		delivery.resyncs = kept
		return false
	}
	delivery.resyncs = append(kept, now)
	mediaWebSocketResyncs.WithLabelValues(metricResyncReason(reason)).Inc()
	return true
}

func (delivery *Delivery) performResync(kind Kind, reason string) error {
	delivery.mu.Lock()
	if delivery.paused || delivery.ctx.Err() != nil {
		delivery.mu.Unlock()
		return nil
	}
	kinds := []Kind{kind}
	if kind == KindNone {
		kinds = nil
		if _, ok := delivery.tracks[KindAudio]; ok {
			kinds = append(kinds, KindAudio)
		}
		kinds = append(kinds, KindVideo)
	}
	protocolReason := normalizeResyncProtocolReason(reason)
	records := make([]queuedRecord, 0, len(kinds)*2)
	for _, currentKind := range kinds {
		track := delivery.tracks[currentKind]
		if track == nil || !track.formatSent {
			delivery.mu.Unlock()
			return errors.New("resync track unavailable")
		}
		track.source = track.subscription.Source()
		if !completeSupportedSource(track.source) || track.generation >= MaxSafeInteger {
			delivery.mu.Unlock()
			return errors.New("resync format unavailable")
		}
		track.generation++
		track.nextSequence = 0
		track.awaitKeyframe = currentKind == KindVideo
		track.transitioning = true
		track.lastRendered = 0
		track.sentUnits = 0
		track.sentTimeline = nil
		track.lastProviderGen = track.source.Generation
		track.pendingReason = ""
		discontinuity, err := makeDiscontinuityRecord(currentKind, track.generation, protocolReason)
		if err != nil {
			delivery.mu.Unlock()
			return err
		}
		records = append(records, discontinuity)
	}
	for _, currentKind := range kinds {
		track := delivery.tracks[currentKind]
		format, err := makeFormatRecord(currentKind, track.generation, track.source)
		if err != nil {
			delivery.mu.Unlock()
			return err
		}
		records = append(records, format)
	}
	now := time.Now()
	delivery.lastFeedbackAt = now
	delivery.lastProgressAt = now
	delivery.sentSinceProgress = false
	delivery.skewExceededAt = time.Time{}
	if reason == "progress_timeout" {
		delivery.progressRecoveryDeadline = now.Add(ProgressRecoveryTimeout)
	} else {
		delivery.progressRecoveryDeadline = time.Time{}
	}
	delivery.mu.Unlock()

	for _, currentKind := range kinds {
		if dropped := delivery.queue.clearMedia(currentKind); dropped > 0 {
			mediaWebSocketDrops.WithLabelValues(metricKind(currentKind), "egress", "resync_clear").Add(float64(dropped))
		}
	}
	for _, record := range records {
		if !delivery.queue.pushLifecycle(record) {
			return errors.New("lifecycle queue full during resync")
		}
	}
	delivery.mu.Lock()
	for _, currentKind := range kinds {
		delivery.tracks[currentKind].transitioning = false
	}
	delivery.mu.Unlock()
	return nil
}

func (delivery *Delivery) monitor() {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-delivery.ctx.Done():
			return
		case now := <-ticker.C:
			requestProgressResync := false
			delivery.mu.Lock()
			if delivery.paused {
				delivery.mu.Unlock()
				continue
			}
			if !delivery.ready {
				if now.Sub(delivery.createdAt) >= ReadyTimeout {
					delivery.mu.Unlock()
					delivery.close(codecClose("ready_timeout"))
					return
				}
				delivery.mu.Unlock()
				continue
			}
			feedbackReference := delivery.lastFeedbackAt
			if feedbackReference.IsZero() {
				feedbackReference = delivery.readyAt
			}
			if delivery.sentSinceProgress && now.Sub(feedbackReference) >= FeedbackTimeout {
				delivery.mu.Unlock()
				delivery.close(backpressureClose("feedback_timeout"))
				return
			}
			if !delivery.progressRecoveryDeadline.IsZero() && !now.Before(delivery.progressRecoveryDeadline) {
				delivery.mu.Unlock()
				delivery.close(backpressureClose("progress_timeout"))
				return
			}
			if delivery.sentSinceProgress && delivery.progressRecoveryDeadline.IsZero() && now.Sub(delivery.lastProgressAt) >= RenderedProgressTimeout {
				delivery.progressRecoveryDeadline = now.Add(ProgressRecoveryTimeout)
				requestProgressResync = true
			}
			delivery.mu.Unlock()
			if requestProgressResync {
				delivery.requestResync(KindNone, "progress_timeout")
			}
			if !delivery.lease.Valid() {
				delivery.close(delivery.invalidLeaseClose())
				return
			}
		}
	}
}

func normalizeResyncProtocolReason(reason string) string {
	switch reason {
	case "server_overflow":
		return "server_overflow"
	case "source_switch", "source_restart", "format_change", "timestamp_reset", "resumed":
		return reason
	default:
		return "browser_resync"
	}
}

func metricResyncReason(reason string) string {
	switch reason {
	case "server_overflow", "source_switch", "source_restart", "format_change", "timestamp_reset", "resumed", "progress_timeout", "av_skew":
		return reason
	case "client_queue_overflow", "client_video_compressed_overflow", "client_audio_compressed_overflow", "client_audio_output_overflow", "client_audio_worklet_overflow", "client_decoder_error", "client_timestamp", "client_audio_underflow", "client_av_skew":
		return reason
	default:
		return "other"
	}
}

func protocolClose(reason string) closeRequest {
	return closeRequest{endReason: "backend_error", code: 4400, reason: reason, failed: true}
}

func attachmentClose(reason string) closeRequest {
	return closeRequest{endReason: "revoked", code: 4401, reason: reason}
}

func codecClose(reason string) closeRequest {
	return closeRequest{endReason: "backend_error", code: 4406, reason: reason, failed: true}
}

func rateClose(reason string) closeRequest {
	return closeRequest{endReason: "backend_error", code: 4429, reason: reason, failed: true}
}
