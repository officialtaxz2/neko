package message

import modern "github.com/m1k1o/neko/server/pkg/types/message"

type MediaCapabilitiesRequest struct {
	Event   string `json:"event"`
	Version int    `json:"version"`
}

type MediaCapabilities struct {
	Event    string                         `json:"event"`
	Version  int                            `json:"version"`
	Backend  string                         `json:"backend"`
	Protocol string                         `json:"protocol"`
	Audio    []modern.MediaCapabilitySource `json:"audio"`
	Video    []modern.MediaCapabilitySource `json:"video"`
}

type MediaCreate struct {
	Event   string                    `json:"event"`
	Version int                       `json:"version"`
	Backend string                    `json:"backend"`
	Audio   *modern.MediaCreateChoice `json:"audio"`
	Video   *modern.MediaCreateChoice `json:"video"`
}

type MediaOffer struct {
	Event       string `json:"event"`
	Version     int    `json:"version"`
	Backend     string `json:"backend"`
	Protocol    string `json:"protocol"`
	Path        string `json:"path"`
	Ticket      string `json:"ticket"`
	ExpiresInMS int64  `json:"expires_in_ms"`
}

type MediaHLSCapabilitiesRequest struct {
	Event   string `json:"event"`
	Version int    `json:"version"`
	Mode    string `json:"mode"`
}

type MediaHLSCapabilities struct {
	Event      string                   `json:"event"`
	Version    int                      `json:"version"`
	Backend    string                   `json:"backend"`
	Modes      []string                 `json:"modes"`
	Container  string                   `json:"container"`
	VideoCodec string                   `json:"video_codec"`
	AudioCodec string                   `json:"audio_codec"`
	AudioRate  uint32                   `json:"audio_rate"`
	Variants   []modern.MediaHLSVariant `json:"variants"`
	Limits     modern.MediaHLSLimits    `json:"limits"`
}

type MediaHLSCreate struct {
	Event   string `json:"event"`
	Version int    `json:"version"`
	Backend string `json:"backend"`
	Mode    string `json:"mode"`
}

type MediaHLSOffer struct {
	Event       string `json:"event"`
	Version     int    `json:"version"`
	Backend     string `json:"backend"`
	Mode        string `json:"mode"`
	Path        string `json:"path"`
	Ticket      string `json:"ticket"`
	ExpiresInMS int64  `json:"expires_in_ms"`
}
