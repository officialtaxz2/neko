package mediaws

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

const MaxControlRecord = 4 * 1024

var ErrInvalidControl = errors.New("invalid media websocket control record")

type readyKind struct {
	Generation  uint64
	Codec       string
	SampleRate  uint32
	Channels    uint16
	CodedWidth  uint32
	CodedHeight uint32
}

type readyControl struct {
	Audio *readyKind
	Video *readyKind
}

type feedbackKind struct {
	Generation      uint64
	Received        uint64
	Decoded         uint64
	Rendered        uint64
	CompressedQueue uint32
	DecodeQueue     uint32
	BufferedMS      uint32
	Drops           uint64
}

type feedbackControl struct {
	Audio    *feedbackKind
	Video    *feedbackKind
	AVSkewMS int32
}

type resyncControl struct {
	Kind       string
	Generation uint64
	Reason     string
}

type parsedControl struct {
	Type     string
	Ready    readyControl
	Feedback feedbackControl
	Resync   resyncControl
}

type readyAudioWire struct {
	Generation string `json:"generation"`
	Codec      string `json:"codec"`
	SampleRate uint32 `json:"sample_rate"`
	Channels   uint16 `json:"channels"`
}

type readyVideoWire struct {
	Generation  string `json:"generation"`
	Codec       string `json:"codec"`
	CodedWidth  uint32 `json:"coded_width"`
	CodedHeight uint32 `json:"coded_height"`
}

type readyWire struct {
	Type    string          `json:"type"`
	Version int             `json:"version"`
	Audio   json.RawMessage `json:"audio"`
	Video   json.RawMessage `json:"video"`
}

type feedbackWireKind struct {
	Generation      string `json:"generation"`
	Received        string `json:"received"`
	Decoded         string `json:"decoded"`
	Rendered        string `json:"rendered"`
	CompressedQueue uint32 `json:"compressed_queue"`
	DecodeQueue     uint32 `json:"decode_queue"`
	BufferedMS      uint32 `json:"buffered_ms"`
	Drops           string `json:"drops"`
}

type feedbackWire struct {
	Type     string          `json:"type"`
	Audio    json.RawMessage `json:"audio"`
	Video    json.RawMessage `json:"video"`
	AVSkewMS int32           `json:"av_skew_ms"`
}

type resyncWire struct {
	Type       string `json:"type"`
	Kind       string `json:"kind"`
	Generation string `json:"generation"`
	Reason     string `json:"reason"`
}

func parseControl(data []byte) (parsedControl, error) {
	if len(data) == 0 || len(data) > MaxControlRecord {
		return parsedControl{}, fmt.Errorf("%w: length", ErrInvalidControl)
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return parsedControl{}, fmt.Errorf("%w: json", ErrInvalidControl)
	}
	var discriminator map[string]json.RawMessage
	if err := json.Unmarshal(data, &discriminator); err != nil {
		return parsedControl{}, fmt.Errorf("%w: json", ErrInvalidControl)
	}
	rawType, ok := discriminator["type"]
	if !ok {
		return parsedControl{}, fmt.Errorf("%w: type", ErrInvalidControl)
	}
	var controlType string
	if err := json.Unmarshal(rawType, &controlType); err != nil || controlType == "" {
		return parsedControl{}, fmt.Errorf("%w: type", ErrInvalidControl)
	}

	switch controlType {
	case "ready":
		ready, err := parseReady(data)
		return parsedControl{Type: controlType, Ready: ready}, err
	case "feedback":
		feedback, err := parseFeedback(data)
		return parsedControl{Type: controlType, Feedback: feedback}, err
	case "resync":
		resync, err := parseResync(data)
		return parsedControl{Type: controlType, Resync: resync}, err
	case "stop":
		var stop struct {
			Type string `json:"type"`
		}
		if err := DecodeStrictJSON(data, &stop); err != nil || stop.Type != "stop" {
			return parsedControl{}, fmt.Errorf("%w: stop", ErrInvalidControl)
		}
		if err := RequireJSONFields(data, "type"); err != nil {
			return parsedControl{}, fmt.Errorf("%w: stop", ErrInvalidControl)
		}
		return parsedControl{Type: controlType}, nil
	default:
		return parsedControl{}, fmt.Errorf("%w: unknown type", ErrInvalidControl)
	}
}

func parseReady(data []byte) (readyControl, error) {
	var wire readyWire
	if err := DecodeStrictJSON(data, &wire); err != nil {
		return readyControl{}, fmt.Errorf("%w: ready", ErrInvalidControl)
	}
	if err := RequireJSONFields(data, "type", "version", "audio", "video"); err != nil || wire.Type != "ready" || wire.Version != int(ProtocolVersion) {
		return readyControl{}, fmt.Errorf("%w: ready", ErrInvalidControl)
	}
	audio, err := parseReadyKind(wire.Audio, KindAudio)
	if err != nil {
		return readyControl{}, err
	}
	video, err := parseReadyKind(wire.Video, KindVideo)
	if err != nil {
		return readyControl{}, err
	}
	return readyControl{Audio: audio, Video: video}, nil
}

func parseReadyKind(raw json.RawMessage, kind Kind) (*readyKind, error) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	if kind == KindAudio {
		var wire readyAudioWire
		if err := DecodeStrictJSON(raw, &wire); err != nil {
			return nil, fmt.Errorf("%w: ready audio", ErrInvalidControl)
		}
		generation, err := parseCanonicalUint(wire.Generation, false)
		if err != nil {
			return nil, err
		}
		if err := RequireJSONFields(raw, "generation", "codec", "sample_rate", "channels"); err != nil || wire.Codec != "opus" || wire.SampleRate != 48000 || wire.Channels != 2 {
			return nil, fmt.Errorf("%w: ready audio", ErrInvalidControl)
		}
		return &readyKind{
			Generation: generation,
			Codec:      wire.Codec,
			SampleRate: wire.SampleRate,
			Channels:   wire.Channels,
		}, nil
	}
	var wire readyVideoWire
	if err := DecodeStrictJSON(raw, &wire); err != nil {
		return nil, fmt.Errorf("%w: ready video", ErrInvalidControl)
	}
	generation, err := parseCanonicalUint(wire.Generation, false)
	if err != nil {
		return nil, err
	}
	if err := RequireJSONFields(raw, "generation", "codec", "coded_width", "coded_height"); err != nil || wire.Codec != "vp8" || wire.CodedWidth == 0 || wire.CodedWidth > MaxVideoDimension || wire.CodedHeight == 0 || wire.CodedHeight > MaxVideoDimension {
		return nil, fmt.Errorf("%w: ready video", ErrInvalidControl)
	}
	return &readyKind{
		Generation:  generation,
		Codec:       wire.Codec,
		CodedWidth:  wire.CodedWidth,
		CodedHeight: wire.CodedHeight,
	}, nil
}

func parseFeedback(data []byte) (feedbackControl, error) {
	var wire feedbackWire
	if err := DecodeStrictJSON(data, &wire); err != nil {
		return feedbackControl{}, fmt.Errorf("%w: feedback", ErrInvalidControl)
	}
	if err := RequireJSONFields(data, "type", "audio", "video", "av_skew_ms"); err != nil || wire.Type != "feedback" || wire.AVSkewMS < -10000 || wire.AVSkewMS > 10000 {
		return feedbackControl{}, fmt.Errorf("%w: feedback", ErrInvalidControl)
	}
	audio, err := parseFeedbackKind(wire.Audio, KindAudio)
	if err != nil {
		return feedbackControl{}, err
	}
	video, err := parseFeedbackKind(wire.Video, KindVideo)
	if err != nil {
		return feedbackControl{}, err
	}
	return feedbackControl{Audio: audio, Video: video, AVSkewMS: wire.AVSkewMS}, nil
}

func parseFeedbackKind(raw json.RawMessage, kind Kind) (*feedbackKind, error) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	var wire feedbackWireKind
	if err := DecodeStrictJSON(raw, &wire); err != nil {
		return nil, fmt.Errorf("%w: feedback kind", ErrInvalidControl)
	}
	if err := RequireJSONFields(raw, "generation", "received", "decoded", "rendered", "compressed_queue", "decode_queue", "buffered_ms", "drops"); err != nil {
		return nil, fmt.Errorf("%w: feedback kind", ErrInvalidControl)
	}
	generation, err := parseCanonicalUint(wire.Generation, false)
	if err != nil {
		return nil, err
	}
	received, err := parseCanonicalUint(wire.Received, true)
	if err != nil {
		return nil, err
	}
	decoded, err := parseCanonicalUint(wire.Decoded, true)
	if err != nil {
		return nil, err
	}
	rendered, err := parseCanonicalUint(wire.Rendered, true)
	if err != nil {
		return nil, err
	}
	drops, err := parseCanonicalUint(wire.Drops, true)
	if err != nil {
		return nil, err
	}
	queueLimit := uint32(ProviderVideoQueueCapacity)
	if kind == KindAudio {
		queueLimit = ProviderAudioQueueCapacity
	}
	if wire.CompressedQueue > queueLimit || wire.DecodeQueue > queueLimit || wire.BufferedMS > 200 || decoded > received || rendered > decoded {
		return nil, fmt.Errorf("%w: feedback bounds", ErrInvalidControl)
	}
	return &feedbackKind{
		Generation: generation,
		Received: received,
		Decoded: decoded,
		Rendered: rendered,
		CompressedQueue: wire.CompressedQueue,
		DecodeQueue: wire.DecodeQueue,
		BufferedMS: wire.BufferedMS,
		Drops: drops,
	}, nil
}

func parseResync(data []byte) (resyncControl, error) {
	var wire resyncWire
	if err := DecodeStrictJSON(data, &wire); err != nil {
		return resyncControl{}, fmt.Errorf("%w: resync", ErrInvalidControl)
	}
	if err := RequireJSONFields(data, "type", "kind", "generation", "reason"); err != nil || wire.Type != "resync" {
		return resyncControl{}, fmt.Errorf("%w: resync", ErrInvalidControl)
	}
	if wire.Kind != "all" && wire.Kind != "audio" && wire.Kind != "video" {
		return resyncControl{}, fmt.Errorf("%w: resync kind", ErrInvalidControl)
	}
	switch wire.Reason {
	case "queue_overflow", "decoder_error", "timestamp", "audio_underflow", "av_skew":
	default:
		return resyncControl{}, fmt.Errorf("%w: resync reason", ErrInvalidControl)
	}
	generation, err := parseCanonicalUint(wire.Generation, false)
	if err != nil {
		return resyncControl{}, err
	}
	return resyncControl{Kind: wire.Kind, Generation: generation, Reason: wire.Reason}, nil
}

func parseCanonicalUint(value string, allowZero bool) (uint64, error) {
	if value == "" || (len(value) > 1 && value[0] == '0') || value[0] == '+' || value[0] == '-' {
		return 0, fmt.Errorf("%w: integer", ErrInvalidControl)
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed > MaxSafeInteger || (!allowZero && parsed == 0) {
		return 0, fmt.Errorf("%w: integer", ErrInvalidControl)
	}
	return parsed, nil
}
