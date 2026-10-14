package message

type MediaHLSCapabilitiesRequest struct {
	Version int    `json:"version"`
	Mode    string `json:"mode"`
}

type MediaHLSVariant struct {
	ID               string `json:"id"`
	SourceID         string `json:"source_id"`
	VideoCodec       string `json:"video_codec"`
	Container        string `json:"container"`
	Width            uint32 `json:"width"`
	Height           uint32 `json:"height"`
	FrameRate        uint32 `json:"frame_rate"`
	AverageBandwidth uint64 `json:"average_bandwidth"`
	Bandwidth        uint64 `json:"bandwidth"`
}

type MediaHLSLimits struct {
	MaxLeases             int `json:"max_leases"`
	MaxRequests           int `json:"max_requests"`
	MaxRequestsPerLease   int `json:"max_requests_per_lease"`
	MaxBlockingPerLease   int `json:"max_blocking_per_lease"`
	MaxBlockingRequests   int `json:"max_blocking_requests"`
	IdleExpiresInMS       int `json:"idle_expires_in_ms"`
	BlockingReloadWaitMS  int `json:"blocking_reload_wait_ms"`
	MaximumPlaylistBytes  int `json:"maximum_playlist_bytes"`
}

type MediaHLSCapabilities struct {
	Version    int               `json:"version"`
	Backend    string            `json:"backend"`
	Modes      []string          `json:"modes"`
	Container  string            `json:"container"`
	VideoCodec string            `json:"video_codec"`
	AudioCodec string            `json:"audio_codec"`
	AudioRate  uint32            `json:"audio_rate"`
	Variants   []MediaHLSVariant `json:"variants"`
	Limits     MediaHLSLimits    `json:"limits"`
}

type MediaHLSCreate struct {
	Version int    `json:"version"`
	Backend string `json:"backend"`
	Mode    string `json:"mode"`
}

type MediaHLSOffer struct {
	Version     int    `json:"version"`
	Backend     string `json:"backend"`
	Mode        string `json:"mode"`
	Path        string `json:"path"`
	Ticket      string `json:"ticket"`
	ExpiresInMS int64  `json:"expires_in_ms"`
}
