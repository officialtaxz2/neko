package mediaws

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	ProtocolName    = "neko.media.v1"
	ProtocolVersion = uint8(1)
	HeaderLength    = 64

	MaxSafeInteger        = uint64(1<<53 - 1)
	MaxMetadata           = 4 * 1024
	MaxCodecConfig        = 64 * 1024
	MaxAudioPayload       = 64 * 1024
	MaxVideoPayload       = 8 * 1024 * 1024
	MaxUnitDurationMicros = uint64(10_000_000)
	MaxSourceIDLength     = 256
	MaxVideoDimension     = 16_383
)

var (
	ErrInvalidRecord = errors.New("invalid media websocket record")
	ErrInvalidJSON   = errors.New("invalid media websocket json")
)

type RecordType uint8

const (
	RecordFormat        RecordType = 1
	RecordUnit          RecordType = 2
	RecordDiscontinuity RecordType = 3
	RecordEnd           RecordType = 4
)

type Kind uint8

const (
	KindNone  Kind = 0
	KindAudio Kind = 1
	KindVideo Kind = 2
)

const (
	FlagKeyframe      uint8 = 1 << 0
	FlagPTSValid      uint8 = 1 << 1
	FlagDTSValid      uint8 = 1 << 2
	FlagConfigPresent uint8 = 1 << 3
	knownFlags              = FlagKeyframe | FlagPTSValid | FlagDTSValid | FlagConfigPresent
)

type Record struct {
	Type       RecordType
	Kind       Kind
	Flags      uint8
	TrackID    uint32
	Generation uint64
	Sequence   uint64
	PTS        int64
	DTS        int64
	Duration   uint64
	Metadata   []byte
	Payload    []byte
}

type FormatMetadata struct {
	Schema               string `json:"schema"`
	SourceID             string `json:"source_id"`
	SourceGeneration     uint64 `json:"source_generation"`
	Codec                string `json:"codec"`
	MIMEType             string `json:"mime_type"`
	ClockRate            uint32 `json:"clock_rate"`
	Channels             uint16 `json:"channels"`
	CodedWidth           uint32 `json:"coded_width"`
	CodedHeight          uint32 `json:"coded_height"`
	DisplayWidth         uint32 `json:"display_width"`
	DisplayHeight        uint32 `json:"display_height"`
	FrameRateNumerator   uint32 `json:"frame_rate_numerator"`
	FrameRateDenominator uint32 `json:"frame_rate_denominator"`
	NominalBitrate       uint64 `json:"nominal_bitrate"`
}

type DiscontinuityMetadata struct {
	Schema string `json:"schema"`
	Reason string `json:"reason"`
}

type EndMetadata struct {
	Schema string `json:"schema"`
	Reason string `json:"reason"`
}

func MarshalRecord(record Record) ([]byte, error) {
	if err := validateRecord(record); err != nil {
		return nil, err
	}

	total := HeaderLength + len(record.Metadata) + len(record.Payload)
	out := make([]byte, total)
	copy(out[0:4], []byte("NEKO"))
	out[4] = ProtocolVersion
	out[5] = byte(record.Type)
	out[6] = byte(record.Kind)
	out[7] = record.Flags
	binary.BigEndian.PutUint16(out[8:10], HeaderLength)
	binary.BigEndian.PutUint16(out[10:12], 0)
	binary.BigEndian.PutUint32(out[12:16], uint32(len(record.Metadata)))
	binary.BigEndian.PutUint32(out[16:20], uint32(len(record.Payload)))
	binary.BigEndian.PutUint32(out[20:24], record.TrackID)
	binary.BigEndian.PutUint64(out[24:32], record.Generation)
	binary.BigEndian.PutUint64(out[32:40], record.Sequence)
	binary.BigEndian.PutUint64(out[40:48], uint64(record.PTS))
	binary.BigEndian.PutUint64(out[48:56], uint64(record.DTS))
	binary.BigEndian.PutUint64(out[56:64], record.Duration)
	copy(out[HeaderLength:], record.Metadata)
	copy(out[HeaderLength+len(record.Metadata):], record.Payload)
	return out, nil
}

func ParseRecord(data []byte) (Record, error) {
	if len(data) < HeaderLength {
		return Record{}, fmt.Errorf("%w: truncated header", ErrInvalidRecord)
	}
	if !bytes.Equal(data[:4], []byte("NEKO")) {
		return Record{}, fmt.Errorf("%w: magic", ErrInvalidRecord)
	}
	if data[4] != ProtocolVersion {
		return Record{}, fmt.Errorf("%w: version", ErrInvalidRecord)
	}
	if binary.BigEndian.Uint16(data[8:10]) != HeaderLength {
		return Record{}, fmt.Errorf("%w: header length", ErrInvalidRecord)
	}
	if binary.BigEndian.Uint16(data[10:12]) != 0 {
		return Record{}, fmt.Errorf("%w: reserved", ErrInvalidRecord)
	}

	metadataLength := uint64(binary.BigEndian.Uint32(data[12:16]))
	payloadLength := uint64(binary.BigEndian.Uint32(data[16:20]))
	total := uint64(HeaderLength) + metadataLength + payloadLength
	if total != uint64(len(data)) {
		return Record{}, fmt.Errorf("%w: length", ErrInvalidRecord)
	}

	record := Record{
		Type:       RecordType(data[5]),
		Kind:       Kind(data[6]),
		Flags:      data[7],
		TrackID:    binary.BigEndian.Uint32(data[20:24]),
		Generation: binary.BigEndian.Uint64(data[24:32]),
		Sequence:   binary.BigEndian.Uint64(data[32:40]),
		PTS:        int64(binary.BigEndian.Uint64(data[40:48])),
		DTS:        int64(binary.BigEndian.Uint64(data[48:56])),
		Duration:   binary.BigEndian.Uint64(data[56:64]),
	}
	if err := validateEncodedLengths(record.Type, record.Kind, metadataLength, payloadLength); err != nil {
		return Record{}, err
	}

	metadataEnd := HeaderLength + int(metadataLength)
	record.Metadata = append([]byte(nil), data[HeaderLength:metadataEnd]...)
	record.Payload = append([]byte(nil), data[metadataEnd:]...)
	if err := validateRecord(record); err != nil {
		return Record{}, err
	}
	return record, nil
}

// validateEncodedLengths rejects impossible or oversized records before ParseRecord
// allocates copies of attacker-controlled metadata or payload bytes.
func validateEncodedLengths(recordType RecordType, kind Kind, metadataLength, payloadLength uint64) error {
	if metadataLength > MaxMetadata {
		return fmt.Errorf("%w: metadata too large", ErrInvalidRecord)
	}
	switch recordType {
	case RecordFormat:
		// VP8 and raw Opus have no decoder-description bytes in version 1.
		if kind != KindAudio && kind != KindVideo {
			return fmt.Errorf("%w: format kind", ErrInvalidRecord)
		}
		if metadataLength == 0 || payloadLength != 0 {
			return fmt.Errorf("%w: format lengths", ErrInvalidRecord)
		}
	case RecordUnit:
		if metadataLength != 0 || payloadLength == 0 {
			return fmt.Errorf("%w: unit lengths", ErrInvalidRecord)
		}
		switch kind {
		case KindAudio:
			if payloadLength > MaxAudioPayload {
				return fmt.Errorf("%w: audio payload too large", ErrInvalidRecord)
			}
		case KindVideo:
			if payloadLength > MaxVideoPayload {
				return fmt.Errorf("%w: video payload too large", ErrInvalidRecord)
			}
		default:
			return fmt.Errorf("%w: unit kind", ErrInvalidRecord)
		}
	case RecordDiscontinuity, RecordEnd:
		if metadataLength == 0 || payloadLength != 0 {
			return fmt.Errorf("%w: lifecycle lengths", ErrInvalidRecord)
		}
	default:
		return fmt.Errorf("%w: record type", ErrInvalidRecord)
	}
	return nil
}

func validateRecord(record Record) error {
	if record.Flags&^knownFlags != 0 {
		return fmt.Errorf("%w: unknown flags", ErrInvalidRecord)
	}
	if err := validateEncodedLengths(record.Type, record.Kind, uint64(len(record.Metadata)), uint64(len(record.Payload))); err != nil {
		return err
	}
	if record.Generation > MaxSafeInteger || record.Sequence > MaxSafeInteger {
		return fmt.Errorf("%w: integer range", ErrInvalidRecord)
	}
	if record.PTS < 0 || record.DTS < 0 {
		return fmt.Errorf("%w: negative timestamp", ErrInvalidRecord)
	}
	if uint64(record.PTS) > MaxSafeInteger || uint64(record.DTS) > MaxSafeInteger || record.Duration > MaxSafeInteger {
		return fmt.Errorf("%w: timestamp range", ErrInvalidRecord)
	}

	switch record.Type {
	case RecordFormat:
		if err := validateTrack(record.Kind, record.TrackID); err != nil {
			return err
		}
		if record.Generation == 0 || record.Sequence != 0 || record.PTS != 0 || record.DTS != 0 || record.Duration != 0 {
			return fmt.Errorf("%w: format lifecycle fields", ErrInvalidRecord)
		}
		if record.Flags != 0 || len(record.Payload) != 0 {
			return fmt.Errorf("%w: version 1 format config", ErrInvalidRecord)
		}
		var metadata FormatMetadata
		if err := DecodeStrictJSON(record.Metadata, &metadata); err != nil {
			return err
		}
		if err := RequireJSONFields(record.Metadata,
			"schema", "source_id", "source_generation", "codec", "mime_type", "clock_rate", "channels",
			"coded_width", "coded_height", "display_width", "display_height",
			"frame_rate_numerator", "frame_rate_denominator", "nominal_bitrate",
		); err != nil {
			return err
		}
		if err := validateFormatMetadata(record.Kind, metadata); err != nil {
			return err
		}

	case RecordUnit:
		if err := validateTrack(record.Kind, record.TrackID); err != nil {
			return err
		}
		if record.Generation == 0 {
			return fmt.Errorf("%w: unit fields", ErrInvalidRecord)
		}
		if record.Flags&FlagConfigPresent != 0 {
			return fmt.Errorf("%w: unit config flag", ErrInvalidRecord)
		}
		if record.Kind == KindAudio && record.Flags&FlagKeyframe != 0 {
			return fmt.Errorf("%w: audio keyframe flag", ErrInvalidRecord)
		}
		if record.Flags&FlagDTSValid == 0 && record.DTS != 0 {
			return fmt.Errorf("%w: invalid dts must be zero", ErrInvalidRecord)
		}
		if record.Duration == 0 || record.Duration > MaxUnitDurationMicros {
			return fmt.Errorf("%w: duration", ErrInvalidRecord)
		}

	case RecordDiscontinuity:
		if err := validateTrack(record.Kind, record.TrackID); err != nil {
			return err
		}
		if record.Flags != 0 || record.Generation == 0 || record.Sequence != 0 || record.PTS != 0 || record.DTS != 0 || record.Duration != 0 {
			return fmt.Errorf("%w: discontinuity fields", ErrInvalidRecord)
		}
		var metadata DiscontinuityMetadata
		if err := DecodeStrictJSON(record.Metadata, &metadata); err != nil {
			return err
		}
		if err := RequireJSONFields(record.Metadata, "schema", "reason"); err != nil {
			return err
		}
		if metadata.Schema != "neko.media.discontinuity/1" || !allowedDiscontinuityReason(metadata.Reason) {
			return fmt.Errorf("%w: discontinuity metadata", ErrInvalidRecord)
		}

	case RecordEnd:
		if record.Kind != KindNone || record.TrackID != 0 || record.Flags != 0 || record.Generation != 0 || record.Sequence != 0 || record.PTS != 0 || record.DTS != 0 || record.Duration != 0 {
			return fmt.Errorf("%w: end fields", ErrInvalidRecord)
		}
		var metadata EndMetadata
		if err := DecodeStrictJSON(record.Metadata, &metadata); err != nil {
			return err
		}
		if err := RequireJSONFields(record.Metadata, "schema", "reason"); err != nil {
			return err
		}
		if metadata.Schema != "neko.media.end/1" || !allowedEndReason(metadata.Reason) {
			return fmt.Errorf("%w: end metadata", ErrInvalidRecord)
		}
	}
	return nil
}

func validateTrack(kind Kind, track uint32) error {
	if (kind == KindAudio && track == 1) || (kind == KindVideo && track == 2) {
		return nil
	}
	return fmt.Errorf("%w: kind/track", ErrInvalidRecord)
}

func validateFormatMetadata(kind Kind, metadata FormatMetadata) error {
	if metadata.Schema != "neko.media.format/1" || metadata.SourceID == "" || len(metadata.SourceID) > MaxSourceIDLength || metadata.SourceGeneration == 0 || metadata.SourceGeneration > MaxSafeInteger || metadata.NominalBitrate > MaxSafeInteger {
		return fmt.Errorf("%w: format metadata", ErrInvalidRecord)
	}
	if kind == KindVideo {
		if metadata.Codec != "vp8" || metadata.MIMEType != "video/VP8" || metadata.ClockRate != 90000 || metadata.Channels != 0 ||
			metadata.CodedWidth == 0 || metadata.CodedWidth > MaxVideoDimension || metadata.CodedHeight == 0 || metadata.CodedHeight > MaxVideoDimension ||
			metadata.DisplayWidth == 0 || metadata.DisplayWidth > MaxVideoDimension || metadata.DisplayHeight == 0 || metadata.DisplayHeight > MaxVideoDimension ||
			metadata.FrameRateNumerator == 0 || metadata.FrameRateDenominator == 0 {
			return fmt.Errorf("%w: video format metadata", ErrInvalidRecord)
		}
		return nil
	}
	if metadata.Codec != "opus" || metadata.MIMEType != "audio/opus" || metadata.ClockRate != 48000 || metadata.Channels != 2 ||
		metadata.CodedWidth != 0 || metadata.CodedHeight != 0 || metadata.DisplayWidth != 0 || metadata.DisplayHeight != 0 ||
		metadata.FrameRateNumerator != 0 || metadata.FrameRateDenominator != 0 {
		return fmt.Errorf("%w: audio format metadata", ErrInvalidRecord)
	}
	return nil
}

func allowedDiscontinuityReason(reason string) bool {
	switch reason {
	case "source_switch", "source_restart", "format_change", "timestamp_reset", "server_overflow", "browser_resync", "resumed":
		return true
	default:
		return false
	}
}

func allowedEndReason(reason string) bool {
	switch reason {
	case "normal", "revoked", "replaced", "backend_error", "shutdown":
		return true
	default:
		return false
	}
}

func DecodeStrictJSON(data []byte, out any) error {
	if len(data) == 0 || len(data) > MaxMetadata {
		return fmt.Errorf("%w: json length", ErrInvalidJSON)
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return err
	}
	return nil
}

// RequireJSONFields distinguishes required zero-value or nullable fields from
// omitted fields after DecodeStrictJSON has validated the fixed schema.
func RequireJSONFields(data []byte, required ...string) error {
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	for _, field := range required {
		if _, ok := fields[field]; !ok {
			return fmt.Errorf("%w: missing field %s", ErrInvalidJSON, field)
		}
	}
	return nil
}

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var walk func() error
	walk = func() error {
		token, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
		}
		delim, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]struct{}{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
				}
				key, ok := keyToken.(string)
				if !ok {
					return fmt.Errorf("%w: object key", ErrInvalidJSON)
				}
				if _, exists := seen[key]; exists {
					return fmt.Errorf("%w: duplicate key", ErrInvalidJSON)
				}
				seen[key] = struct{}{}
				if err := walk(); err != nil {
					return err
				}
			}
			if _, err := decoder.Token(); err != nil {
				return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
			}
			return nil
		case '[':
			for decoder.More() {
				if err := walk(); err != nil {
					return err
				}
			}
			if _, err := decoder.Token(); err != nil {
				return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
			}
			return nil
		default:
			return fmt.Errorf("%w: delimiter", ErrInvalidJSON)
		}
	}
	if err := walk(); err != nil {
		return err
	}
	return ensureJSONEOF(decoder)
}

func ensureJSONEOF(decoder *json.Decoder) error {
	if _, err := decoder.Token(); err == io.EOF {
		return nil
	} else if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	return fmt.Errorf("%w: trailing data", ErrInvalidJSON)
}
