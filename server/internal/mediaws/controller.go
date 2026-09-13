package mediaws

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

const (
	MaximumConnections       = 128
	InvalidAttachesPerMinute = 20
	VerifierConcurrency      = 64
	MaximumInvalidAddresses  = 4096
)

type ControllerConfig struct {
	AllowedOrigins        []string
	TrustedProxies        []string
	AllowInsecureLoopback bool
	MaxConnections        int
}

type invalidAttachWindow struct {
	started time.Time
	count   int
}

type Controller struct {
	sessions   types.SessionManager
	deliveries types.MediaDeliveryManager
	tickets    *TicketStore
	origins    map[string]struct{}
	proxies    []netip.Prefix
	allowInsecureLoopback bool
	maximum    int
	verify     chan struct{}

	mu          sync.Mutex
	connections int
	attaching   map[string]struct{}
	invalid     map[string]invalidAttachWindow
	now         func() time.Time
	upgrader    websocket.Upgrader
}

func NewController(sessions types.SessionManager, deliveries types.MediaDeliveryManager, tickets *TicketStore, config ControllerConfig) (*Controller, error) {
	if sessions == nil || deliveries == nil || tickets == nil {
		return nil, errors.New("media websocket controller dependencies are required")
	}
	maximum := config.MaxConnections
	if maximum == 0 {
		maximum = MaximumConnections
	}
	if maximum < 1 || maximum > MaximumConnections {
		return nil, errors.New("media websocket max connections must be between 1 and 128")
	}
	origins := make(map[string]struct{}, len(config.AllowedOrigins))
	for _, origin := range config.AllowedOrigins {
		origin = strings.TrimSpace(origin)
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.Path != "" || parsed.ForceQuery || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.String() != origin {
			return nil, errors.New("media websocket allowed origins must be exact http(s) origins")
		}
		origins[origin] = struct{}{}
	}
	proxies := make([]netip.Prefix, 0, len(config.TrustedProxies))
	for _, value := range config.TrustedProxies {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			address, addressErr := netip.ParseAddr(value)
			if addressErr != nil {
				return nil, errors.New("media websocket trusted proxy must be an IP or CIDR")
			}
			prefix = netip.PrefixFrom(address, address.BitLen())
		}
		proxies = append(proxies, prefix.Masked())
	}

	return &Controller{
		sessions:   sessions,
		deliveries: deliveries,
		tickets:    tickets,
		origins:    origins,
		proxies:    proxies,
		allowInsecureLoopback: config.AllowInsecureLoopback,
		maximum:    maximum,
		verify:     make(chan struct{}, VerifierConcurrency),
		attaching:  make(map[string]struct{}),
		invalid:    make(map[string]invalidAttachWindow),
		now:        time.Now,
		upgrader: websocket.Upgrader{
			Subprotocols:      []string{ProtocolName},
			EnableCompression: false,
			CheckOrigin:       func(*http.Request) bool { return true },
		},
	}, nil
}

func (controller *Controller) Handle(w http.ResponseWriter, r *http.Request) error {
	now := controller.now()
	address, peer, trusted := controller.clientAddress(r)
	if controller.invalidBlocked(address, now) {
		return controller.httpError(http.StatusTooManyRequests, "rate_limited")
	}
	if r.URL.RawQuery != "" || r.URL.ForceQuery {
		controller.recordInvalid(address, now)
		return controller.httpError(http.StatusBadRequest, "query_not_allowed")
	}
	secure, scheme := controller.secureTransport(r, peer, trusted)
	if !secure || !controller.allowOrigin(r.Header.Get("Origin"), scheme, r.Host) {
		controller.recordInvalid(address, now)
		return controller.httpError(http.StatusForbidden, "origin_rejected")
	}
	ticket, status := parseAttachmentProtocols(r.Header.Values("Sec-WebSocket-Protocol"))
	if status != 0 {
		controller.recordInvalid(address, now)
		return controller.httpError(status, "subprotocol_rejected")
	}

	select {
	case controller.verify <- struct{}{}:
	default:
		return controller.httpError(http.StatusTooManyRequests, "verifier_busy")
	}
	binding, err := controller.tickets.Redeem(ticket, now)
	<-controller.verify
	if err != nil {
		controller.recordInvalid(address, now)
		switch {
		case errors.Is(err, ErrTicketMalformed):
			return controller.httpError(http.StatusBadRequest, "ticket_malformed")
		case errors.Is(err, ErrTicketGone):
			return controller.httpError(http.StatusGone, "ticket_gone")
		default:
			return controller.httpError(http.StatusUnauthorized, "ticket_invalid")
		}
	}
	if binding.Backend != BackendName || binding.Version != ProtocolVersion {
		controller.recordInvalid(address, now)
		return controller.httpError(http.StatusUnauthorized, "ticket_binding")
	}
	session, ok := controller.sessions.Get(binding.SessionID)
	if !ok || session == nil {
		controller.recordInvalid(address, now)
		return controller.httpError(http.StatusUnauthorized, "session_missing")
	}
	if !session.Profile().CanWatch || !session.State().IsConnected {
		controller.recordInvalid(address, now)
		return controller.httpError(http.StatusForbidden, "watch_denied")
	}
	if status := controller.reserve(session.ID()); status != 0 {
		result := "attach_conflict"
		if status == http.StatusTooManyRequests {
			result = "connection_limit"
		}
		return controller.httpError(status, result)
	}
	reserved := true
	defer func() {
		if reserved {
			controller.releaseReservation(session.ID())
		}
	}()

	connection, err := controller.upgrader.Upgrade(w, r, nil)
	if err != nil {
		mediaWebSocketHandshakes.WithLabelValues("upgrade_error").Inc()
		return nil
	}
	if connection.Subprotocol() != ProtocolName {
		_ = connection.Close()
		mediaWebSocketHandshakes.WithLabelValues("protocol_error").Inc()
		return nil
	}
	request := types.MediaDeliveryRequest{
		Backend: BackendName,
		Audio:   binding.Request.Audio.Enabled,
		Video:   binding.Request.Video.Enabled,
		AudioSelector: types.MediaSelector{
			Type: types.MediaSelectorTypeExact,
			ID:   binding.Request.Audio.SourceID,
		},
		VideoSelector: types.MediaSelector{
			Type: types.MediaSelectorTypeExact,
			ID:   binding.Request.Video.SourceID,
		},
	}
	ctx := context.WithValue(r.Context(), socketAttachmentContextKey{}, &socketAttachment{connection: connection})
	delivery, err := controller.deliveries.Open(ctx, session, request)
	if err != nil {
		code, reason := attachmentOpenClose(err)
		_ = connection.SetWriteDeadline(time.Now().Add(CloseTimeout))
		_ = connection.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason), time.Now().Add(CloseTimeout))
		_ = connection.Close()
		mediaWebSocketHandshakes.WithLabelValues("backend_error").Inc()
		return nil
	}
	controller.markAttached(session.ID())
	reserved = false
	mediaWebSocketHandshakes.WithLabelValues("success").Inc()
	<-delivery.Done()
	controller.releaseConnection()
	return nil
}

func attachmentOpenClose(err error) (int, string) {
	switch {
	case errors.Is(err, types.ErrMediaDeliveryNotAllowed):
		return 4401, "attachment_invalid"
	case errors.Is(err, ErrCodecUnsupported), errors.Is(err, types.ErrMediaSourceNotFound):
		return 4406, "unsupported_format"
	default:
		return 4500, "backend_open_failed"
	}
}

func (controller *Controller) reserve(sessionID string) int {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if controller.connections >= controller.maximum {
		return http.StatusTooManyRequests
	}
	if _, exists := controller.attaching[sessionID]; exists {
		return http.StatusConflict
	}
	controller.connections++
	controller.attaching[sessionID] = struct{}{}
	return 0
}

func (controller *Controller) markAttached(sessionID string) {
	controller.mu.Lock()
	delete(controller.attaching, sessionID)
	controller.mu.Unlock()
}

func (controller *Controller) releaseReservation(sessionID string) {
	controller.mu.Lock()
	delete(controller.attaching, sessionID)
	if controller.connections > 0 {
		controller.connections--
	}
	controller.mu.Unlock()
}

func (controller *Controller) releaseConnection() {
	controller.mu.Lock()
	if controller.connections > 0 {
		controller.connections--
	}
	controller.mu.Unlock()
}

func (controller *Controller) invalidBlocked(address string, now time.Time) bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	window, ok := controller.invalid[address]
	if !ok || now.Sub(window.started) >= time.Minute || now.Before(window.started) {
		return false
	}
	return window.count >= InvalidAttachesPerMinute
}

func (controller *Controller) recordInvalid(address string, now time.Time) {
	controller.mu.Lock()
	if _, exists := controller.invalid[address]; !exists && len(controller.invalid) >= MaximumInvalidAddresses {
		var oldestAddress string
		var oldestTime time.Time
		for candidateAddress, candidate := range controller.invalid {
			if oldestAddress == "" || candidate.started.Before(oldestTime) {
				oldestAddress = candidateAddress
				oldestTime = candidate.started
			}
		}
		delete(controller.invalid, oldestAddress)
	}
	window, ok := controller.invalid[address]
	if !ok || now.Sub(window.started) >= time.Minute || now.Before(window.started) {
		window = invalidAttachWindow{started: now}
	}
	window.count++
	controller.invalid[address] = window
	for key, candidate := range controller.invalid {
		if now.Sub(candidate.started) >= 2*time.Minute {
			delete(controller.invalid, key)
		}
	}
	controller.mu.Unlock()
}

func (controller *Controller) clientAddress(r *http.Request) (string, netip.Addr, bool) {
	original := utils.OriginalRemoteAddr(r)
	peer := parseAddress(original)
	trusted := controller.isTrustedProxy(peer)
	address := peer.String()
	if trusted {
		forwarded := strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
		if candidate, err := netip.ParseAddr(strings.TrimSpace(forwarded)); err == nil {
			address = candidate.Unmap().String()
		} else if candidate, err := netip.ParseAddr(strings.TrimSpace(r.Header.Get("X-Real-IP"))); err == nil {
			address = candidate.Unmap().String()
		}
	}
	if address == "invalid IP" || address == "" {
		address = "unknown"
	}
	return address, peer, trusted
}

func parseAddress(value string) netip.Addr {
	host, _, err := net.SplitHostPort(value)
	if err != nil {
		host = value
	}
	address, _ := netip.ParseAddr(strings.Trim(host, "[]"))
	return address.Unmap()
}

func (controller *Controller) isTrustedProxy(address netip.Addr) bool {
	if !address.IsValid() {
		return false
	}
	for _, prefix := range controller.proxies {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}

func (controller *Controller) secureTransport(r *http.Request, peer netip.Addr, trusted bool) (bool, string) {
	if r.TLS != nil {
		return true, "https"
	}
	if trusted && r.Header.Get("X-Forwarded-Proto") == "https" {
		return true, "https"
	}
	host := r.Host
	if name, _, err := net.SplitHostPort(host); err == nil {
		host = name
	}
	hostAddress, _ := netip.ParseAddr(strings.Trim(host, "[]"))
	if controller.allowInsecureLoopback && peer.IsLoopback() && hostAddress.IsLoopback() &&
		r.Header.Get("Forwarded") == "" && r.Header.Get("X-Forwarded-For") == "" &&
		r.Header.Get("X-Forwarded-Proto") == "" && r.Header.Get("X-Real-IP") == "" {
		return true, "http"
	}
	return false, ""
}

func (controller *Controller) allowOrigin(origin, scheme, host string) bool {
	if origin == "" {
		return false
	}
	if len(controller.origins) > 0 {
		_, ok := controller.origins[origin]
		return ok
	}
	return origin == scheme+"://"+host
}

func parseAttachmentProtocols(values []string) (string, int) {
	protocols := make([]string, 0, 2)
	for _, value := range values {
		for _, protocol := range strings.Split(value, ",") {
			protocols = append(protocols, strings.TrimSpace(protocol))
		}
	}
	if len(protocols) == 0 || protocols[0] != ProtocolName {
		return "", http.StatusUpgradeRequired
	}
	if len(protocols) != 2 {
		return "", http.StatusBadRequest
	}
	const ticketPrefix = "neko.media.ticket."
	if !strings.HasPrefix(protocols[1], ticketPrefix) {
		return "", http.StatusBadRequest
	}
	ticket := strings.TrimPrefix(protocols[1], ticketPrefix)
	if err := validateTicketSyntax(ticket); err != nil {
		return "", http.StatusBadRequest
	}
	return ticket, 0
}

func (controller *Controller) httpError(status int, result string) error {
	mediaWebSocketHandshakes.WithLabelValues(result).Inc()
	return utils.HttpError(status, http.StatusText(status)).WithInternalMsg(result)
}
