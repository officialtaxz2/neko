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
