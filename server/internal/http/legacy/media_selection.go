package legacy

import (
	"net/http"

	"github.com/m1k1o/neko/server/pkg/types"
)

// Forward the authoritative per-session pause state to cooperative HLS players.
// Control locks are unrelated to private mode. Admin diagnostics retain access.
func (s *session) sendHLSState(settings types.Settings) error {
	values := s.r.URL.Query()["media"]
	if len(values) != 1 || (values[0] != "hls" && values[0] != "ll-hls") {
		return nil
	}
	return s.toClient(struct {
		Event   string `json:"event"`
		Version int    `json:"version"`
		Backend string `json:"backend"`
		Paused  bool   `json:"paused"`
	}{"media/hls/state", 1, "hls", settings.PrivateMode && !s.isAdmin})
}

const webCodecsMediaBackend = "webcodecs-ws"

func webCodecsMediaSelected(request *http.Request) bool {
	values, exists := request.URL.Query()["media"]
	return exists && len(values) == 1 && values[0] == webCodecsMediaBackend
}

// Exact receive-only opt-ins must not create an unused WebRTC peer. They do
// not enable a backend or authorize playback; negotiation still checks both.
func receiveOnlyMediaSelected(request *http.Request) bool {
	values, exists := request.URL.Query()["media"]
	if !exists || len(values) != 1 {
		return false
	}
	return values[0] == webCodecsMediaBackend || values[0] == "hls" || values[0] == "ll-hls"
}
