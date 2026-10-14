package message

import "github.com/m1k1o/neko/server/pkg/types"

type MediaCapabilitiesRequest struct {
	Version int `json:"version"`
}

type MediaCapabilitySource struct {
	ID                   string              `json:"id"`
	Codec                string              `json:"codec"`
	MIMEType             string              `json:"mime_type"`
	ClockRate            uint32              `json:"clock_rate"`
	Channels             uint16              `json:"channels"`
	CodedWidth           uint32              `json:"coded_width"`
	CodedHeight          uint32              `json:"coded_height"`
	FrameRateNumerator   uint32              `json:"frame_rate_numerator"`
	FrameRateDenominator uint32              `json:"frame_rate_denominator"`
	NominalBitrate       uint64              `json:"nominal_bitrate"`
	Selector             types.MediaSelector `json:"selector"`
}

type MediaCapabilities struct {
	Version  int                     `json:"version"`
	Backend  string                  `json:"backend"`
	Protocol string                  `json:"protocol"`
	Audio    []MediaCapabilitySource `json:"audio"`
	Video    []MediaCapabilitySource `json:"video"`
}

type MediaCreateChoice struct {
	SourceID string `json:"source_id"`
	Codec    string `json:"codec"`
}

type MediaCreate struct {
	Version int                `json:"version"`
	Backend string             `json:"backend"`
	Audio   *MediaCreateChoice `json:"audio"`
	Video   *MediaCreateChoice `json:"video"`
}

type MediaOffer struct {
	Version     int    `json:"version"`
	Backend     string `json:"backend"`
	Protocol    string `json:"protocol"`
	Path        string `json:"path"`
	Ticket      string `json:"ticket"`
	ExpiresInMS int64  `json:"expires_in_ms"`
}
