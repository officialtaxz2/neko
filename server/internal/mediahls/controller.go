package mediahls

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/pkg/types"
)

type Controller struct {
	logger     zerolog.Logger
	sessions   types.SessionManager
	deliveries types.MediaDeliveryManager
	tickets    *TicketStore
	leases     *LeaseStore
	backend    *Backend
	security   *SecurityPolicy
	now        func() time.Time
}

func NewController(sessions types.SessionManager, deliveries types.MediaDeliveryManager, tickets *TicketStore, leases *LeaseStore, backend *Backend, config Config) (*Controller, error) {
	if sessions == nil || deliveries == nil || tickets == nil || leases == nil || backend == nil {
		return nil, ErrInvalidConfig
	}
	normalized, err := NormalizeConfig(config)
	if err != nil {
		return nil, err
	}
	security, err := NewSecurityPolicy(normalized.AllowedOrigins, normalized.TrustedProxies)
	if err != nil {
		return nil, err
	}
	return &Controller{
		logger: log.With().Str("module", "mediahls").Str("submodule", "http").Logger(),
		sessions: sessions, deliveries: deliveries, tickets: tickets,
		leases: leases, backend: backend, security: security, now: time.Now,
	}, nil
}

func (controller *Controller) Route(router types.Router) {
	router.Post(BootstrapPath, controller.Bootstrap)
	router.Post("/api/media/hls/{publicID}/keepalive", controller.KeepAlive)
	router.Get("/api/media/hls/{publicID}/*", controller.Media)
	router.Head("/api/media/hls/{publicID}/*", controller.Media)
}

type bootstrapResponse struct {
	Mode            string `json:"mode"`
	Master          string `json:"master"`
	IdleExpiresInMS int64  `json:"idle_expires_in_ms"`
}

func (controller *Controller) Bootstrap(w http.ResponseWriter, request *http.Request) error {
	started := controller.now()
	clearDeadlines := boundHLSRequest(w)
	defer clearDeadlines()
	ApplySecurityHeaders(w.Header())
	mode := ModeHLS
	result := "bad_request"
	defer func() {
		hlsBootstrap.WithLabelValues(metricMode(mode), result).Inc()
		hlsRequestDuration.WithLabelValues(metricMode(mode), "bootstrap").Observe(time.Since(started).Seconds())
	}()
	if err := controller.security.CheckBootstrap(request); err != nil {
		controller.writeStatus(w, RequestStatus(err))
		return nil
	}
	body, err := io.ReadAll(io.LimitReader(request.Body, MaximumBootstrapBody+1))
	if err != nil {
		result = "read_error"
		controller.writeStatus(w, http.StatusBadRequest)
		return nil
	}
	payload, err := DecodeBootstrapPayload(body)
	if err != nil {
		if errors.Is(err, ErrBootstrapBodyTooLarge) {
			result = "too_large"
			controller.writeStatus(w, http.StatusRequestEntityTooLarge)
		} else {
			controller.writeStatus(w, http.StatusBadRequest)
		}
		return nil
	}
	binding, err := controller.tickets.Redeem(payload.Ticket, controller.now())
	if err != nil {
		switch {
		case errors.Is(err, ErrTicketGone):
			result = "gone"
			controller.writeStatus(w, http.StatusGone)
		case errors.Is(err, ErrTicketMalformed):
			controller.writeStatus(w, http.StatusBadRequest)
		default:
			result = "unauthorized"
			controller.writeStatus(w, http.StatusUnauthorized)
		}
		return nil
	}
	mode = binding.Mode
	session, ok := controller.sessions.Get(binding.SessionID)
	if !ok || session == nil || !session.Profile().CanWatch || !session.State().IsConnected {
		result = "denied"
		controller.writeStatus(w, http.StatusForbidden)
		return nil
	}
	ctx := context.WithValue(request.Context(), bootstrapContextKey{}, &bootstrapAttachment{mode: binding.Mode})
	opened, err := controller.deliveries.Open(ctx, session, types.MediaDeliveryRequest{Backend: BackendName, Audio: true, Video: true})
	if err != nil {
		result = bootstrapErrorResult(err)
		if result == "rate_limited" {
			controller.logger.Warn().Str("session_id", binding.SessionID).Str("mode", mode).Msg("HLS lease limit rejected")
		}
		status := hlsOpenErrorStatus(err)
		if status == http.StatusServiceUnavailable {
			w.Header().Set("Retry-After", "1")
		}
		controller.writeStatus(w, status)
		return nil
	}
	delivery, ok := opened.(*Delivery)
	if !ok {
		_ = opened.Close()
		result = "backend_error"
		controller.writeStatus(w, http.StatusInternalServerError)
		return nil
	}
	offer := delivery.Offer()
	http.SetCookie(w, offer.Cookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(bootstrapResponse{Mode: binding.Mode, Master: "/api/media/hls/" + offer.PublicID + "/master.m3u8", IdleExpiresInMS: LeaseLifetime.Milliseconds()}); err != nil {
		controller.logger.Warn().Err(err).Msg("HLS bootstrap response write failed")
		_ = delivery.Close()
		result = "write_error"
		return nil
	}
	result = "success"
	return nil
}

func (controller *Controller) KeepAlive(w http.ResponseWriter, request *http.Request) error {
	started := controller.now()
	clearDeadlines := boundHLSRequest(w)
	defer clearDeadlines()
	ApplySecurityHeaders(w.Header())
	resource, parseErr := ParseResourcePath(request.URL.Path)
	mode := ModeHLS
	result := "bad_request"
	defer func() { controller.observeRequest(mode, ResourceKeepAlive, result, started) }()
	if parseErr != nil {
		controller.writeStatus(w, http.StatusBadRequest)
		return nil
	}
	if err := controller.security.CheckKeepAlive(request); err != nil {
		controller.writeStatus(w, RequestStatus(err))
		return nil
	}
	secret, ok := leaseCookie(request)
	if !ok {
		result = "not_found"
		controller.writeStatus(w, http.StatusNotFound)
		return nil
	}
	permit, snapshot, err := controller.leases.Acquire(resource.PublicID, secret, RequestKeepAlive, false, controller.now())
	if err != nil {
		result = controller.writeLeaseError(w, err)
		return nil
	}
	defer permit.Release()
	mode = snapshot.Mode
	if _, err := controller.extendLease(w, resource.PublicID, secret); err != nil {
		result = controller.writeLeaseError(w, err)
		return nil
	}
	w.WriteHeader(http.StatusNoContent)
	result = "success"
	return nil
}

func (controller *Controller) Media(w http.ResponseWriter, request *http.Request) error {
	started := controller.now()
	clearDeadlines := boundHLSRequest(w)
	defer clearDeadlines()
	ApplySecurityHeaders(w.Header())
	resource, err := ParseResourcePath(request.URL.Path)
	if err != nil {
		controller.writeStatus(w, http.StatusBadRequest)
		return nil
	}
	mode := ModeHLS
	result := "bad_request"
	defer func() { controller.observeRequest(mode, resource.Kind, result, started) }()
	if err := controller.security.CheckMedia(request); err != nil {
		controller.writeStatus(w, RequestStatus(err))
		return nil
	}
	secret, ok := leaseCookie(request)
	if !ok {
		result = "not_found"
		controller.writeStatus(w, http.StatusNotFound)
		return nil
	}
	snapshot, err := controller.leases.Authenticate(resource.PublicID, secret, false, controller.now())
	if err != nil {
		result = controller.writeLeaseError(w, err)
		return nil
	}
	mode = snapshot.Mode
	sessionID := snapshot.SessionID
	blocking := false
	directives := PlaylistDirectives{}
	if resource.Kind == ResourcePlaylist {
		currentMSN, currentPart, positionErr := controller.backend.packager.Position(resource.Variant)
		if positionErr != nil {
			result = "not_ready"
			w.Header().Set("Retry-After", "1")
			controller.writeStatus(w, http.StatusServiceUnavailable)
			return nil
		}
		directives, err = ParsePlaylistQuery(mode, request.URL.Query(), currentMSN, currentPart)
		if err != nil {
			controller.writeStatus(w, http.StatusBadRequest)
			return nil
		}
		blocking = directives.HasMSN && (directives.MSN > currentMSN || (directives.MSN == currentMSN && (!directives.HasPart || directives.Part >= currentPart)))
	} else if request.URL.RawQuery != "" || request.URL.ForceQuery {
		controller.writeStatus(w, http.StatusBadRequest)
		return nil
	}
	if resource.Kind == ResourcePart {
		if _, found := controller.backend.packager.Object(resource.Variant, resource.Object); !found {
			currentMSN, currentPart, positionErr := controller.backend.packager.Position(resource.Variant)
			blocking = positionErr == nil && resource.Sequence == currentMSN && resource.Part == currentPart
		}
	}
	class := RequestObject
	if resource.Kind == ResourceMaster || resource.Kind == ResourcePlaylist {
		class = RequestPlaylist
	}
	permit, snapshot, err := controller.leases.Acquire(resource.PublicID, secret, class, blocking, controller.now())
	if err != nil {
		if errors.Is(err, ErrRequestLimit) {
			controller.logger.Warn().Str("session_id", sessionID).Str("mode", mode).Str("resource", metricResource(resource.Kind)).Msg("HLS request limit rejected")
		}
		result = controller.writeLeaseError(w, err)
		return nil
	}
	defer permit.Release()
	mode = snapshot.Mode
	leaseContext, stopWatch, watchErr := controller.watchLease(w, request, resource.PublicID, secret)
	if watchErr != nil {
		result = controller.writeLeaseError(w, watchErr)
		return nil
	}
	defer stopWatch()
	hlsRequests.WithLabelValues(metricResource(resource.Kind), "active").Inc()
	defer hlsRequests.WithLabelValues(metricResource(resource.Kind), "active").Dec()

	switch resource.Kind {
	case ResourceMaster:
		data, renderErr := controller.backend.packager.Master()
		if renderErr != nil {
			result = "not_ready"
			w.Header().Set("Retry-After", "1")
			controller.writeStatus(w, http.StatusServiceUnavailable)
			return nil
		}
		if _, leaseErr := controller.extendLease(w, resource.PublicID, secret); leaseErr != nil {
			result = controller.writeLeaseError(w, leaseErr)
			return nil
		}
		result = controller.writePlaylist(w, request, data)
	case ResourcePlaylist:
		playlistContext := leaseContext
		cancel := func() {}
		if blocking {
			playlistContext, cancel = context.WithTimeout(playlistContext, BlockingReloadWait)
			hlsBlockedReloads.WithLabelValues("waiting").Inc()
			defer hlsBlockedReloads.WithLabelValues("waiting").Dec()
		}
		defer cancel()
		data, renderErr := controller.backend.packager.Playlist(playlistContext, mode, resource.Variant, directives)
		if renderErr != nil {
			if errors.Is(renderErr, context.DeadlineExceeded) {
				result = "timeout"
				w.Header().Set("Retry-After", "1")
				controller.writeStatus(w, http.StatusServiceUnavailable)
			} else if errors.Is(renderErr, context.Canceled) {
				if request.Context().Err() != nil {
					result = "canceled"
				} else {
					_, leaseErr := controller.leases.Authenticate(resource.PublicID, secret, false, controller.now())
					result = controller.writeLeaseError(w, leaseErr)
				}
			} else {
				result = "not_ready"
				w.Header().Set("Retry-After", "1")
				controller.writeStatus(w, http.StatusServiceUnavailable)
			}
			return nil
		}
		if _, leaseErr := controller.extendLease(w, resource.PublicID, secret); leaseErr != nil {
			result = controller.writeLeaseError(w, leaseErr)
			return nil
		}
		result = controller.writePlaylist(w, request, data)
	case ResourceInit, ResourcePart, ResourceSegment:
		object, found := controller.backend.packager.Object(resource.Variant, resource.Object)
		if !found && resource.Kind == ResourcePart && blocking {
			objectContext, cancel := context.WithTimeout(leaseContext, BlockingReloadWait)
			object, found = controller.backend.packager.WaitObject(objectContext, resource.Variant, resource.Object)
			waitErr := objectContext.Err()
			cancel()
			if !found && waitErr != nil {
				if errors.Is(waitErr, context.DeadlineExceeded) {
					result = "timeout"
					w.Header().Set("Retry-After", "1")
					controller.writeStatus(w, http.StatusServiceUnavailable)
				} else if request.Context().Err() != nil {
					result = "canceled"
				} else {
					_, leaseErr := controller.leases.Authenticate(resource.PublicID, secret, false, controller.now())
					result = controller.writeLeaseError(w, leaseErr)
				}
				return nil
			}
		}
		if !found {
			result = "not_found"
			controller.writeStatus(w, http.StatusNotFound)
			return nil
		}
		result = controller.writeObject(w, request, object)
	default:
		controller.writeStatus(w, http.StatusBadRequest)
		return nil
	}
	if result == "success" && resource.Kind == ResourcePlaylist {
		if delivery, found := controller.backend.Delivery(resource.PublicID, snapshot.SessionID); found {
			delivery.MarkActive()
		}
	}
	return nil
}

func (controller *Controller) writePlaylist(w http.ResponseWriter, request *http.Request, data []byte) string {
	if len(data) == 0 || len(data) > MaximumPlaylistBytes {
		controller.writeStatus(w, http.StatusInternalServerError)
		return "backend_error"
	}
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Priority", "u=1")
	if acceptsGzip(request.Header.Get("Accept-Encoding")) {
		var compressed bytes.Buffer
		writer := gzip.NewWriter(&compressed)
		if _, err := writer.Write(data); err != nil {
			controller.writeStatus(w, http.StatusInternalServerError)
			return "backend_error"
		}
		if err := writer.Close(); err != nil {
			controller.writeStatus(w, http.StatusInternalServerError)
			return "backend_error"
		}
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Length", strconv.Itoa(compressed.Len()))
		w.WriteHeader(http.StatusOK)
		if request.Method == http.MethodHead {
			return "success"
		}
		if _, err := w.Write(compressed.Bytes()); err != nil {
			return "write_error"
		}
		return "success"
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	if request.Method == http.MethodHead {
		return "success"
	}
	if _, err := w.Write(data); err != nil {
		return "write_error"
	}
	return "success"
}

func (controller *Controller) writeObject(w http.ResponseWriter, request *http.Request, object MediaObject) string {
	data := object.bytesView()
	w.Header().Set("Priority", objectPriority(object.Variant()))
	start, end := int64(0), int64(len(data)-1)
	status := http.StatusOK
	if value := request.Header.Get("Range"); value != "" {
		parsed, err := ParseByteRange(value, int64(len(data)))
		if err != nil {
			w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", len(data)))
			controller.writeStatus(w, http.StatusRequestedRangeNotSatisfiable)
			return "range_invalid"
		}
		start, end, status = parsed.Start, parsed.End, http.StatusPartialContent
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(data)))
	}
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", object.ContentType())
	w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	w.WriteHeader(status)
	if request.Method == http.MethodHead {
		return "success"
	}
	if _, err := w.Write(data[start : end+1]); err != nil {
		return "write_error"
	}
	return "success"
}

func objectPriority(variant string) string {
	switch variant {
	case "audio":
		return "u=2"
	case "low":
		return "u=3"
	case "medium":
		return "u=4"
	default:
		return "u=5"
	}
}

func (controller *Controller) watchLease(w http.ResponseWriter, request *http.Request, publicID, secret string) (context.Context, func(), error) {
	changed, err := controller.leases.ChangeChannel(publicID, secret, controller.now())
	if err != nil {
		return request.Context(), func() {}, err
	}
	ctx, cancel := context.WithCancel(request.Context())
	done := make(chan struct{})
	go func() {
		select {
		case <-changed:
			cancel()
			_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(time.Second))
		case <-ctx.Done():
		case <-done:
		}
	}()
	return ctx, func() { cancel(); close(done) }, nil
}

func (controller *Controller) observeRequest(mode string, resource ResourceKind, result string, started time.Time) {
	mode = metricMode(mode)
	resourceLabel := metricResource(resource)
	hlsRequestsTotal.WithLabelValues(mode, resourceLabel, boundedRequestResult(result)).Inc()
	hlsRequestDuration.WithLabelValues(mode, resourceLabel).Observe(time.Since(started).Seconds())
}

func (controller *Controller) writeLeaseError(w http.ResponseWriter, err error) string {
	switch {
	case errors.Is(err, ErrLeasePaused):
		w.Header().Set("Retry-After", "1")
		controller.writeStatus(w, http.StatusServiceUnavailable)
		return "paused"
	case errors.Is(err, ErrRequestLimit):
		controller.writeStatus(w, http.StatusTooManyRequests)
		return "rate_limited"
	default:
		controller.writeStatus(w, http.StatusNotFound)
		return "not_found"
	}
}

func (*Controller) writeStatus(w http.ResponseWriter, status int) { w.WriteHeader(status) }

func leaseCookie(request *http.Request) (string, bool) {
	cookie, err := request.Cookie(LeaseCookieName)
	returnValue := ""
	if err == nil && cookie != nil {
		returnValue = cookie.Value
	}
	return returnValue, err == nil && len(returnValue) == LeaseSecretEncodedLength
}

func acceptsGzip(value string) bool {
	for _, item := range strings.Split(value, ",") {
		parts := strings.Split(item, ";")
		if !strings.EqualFold(strings.TrimSpace(parts[0]), "gzip") {
			continue
		}
		quality := 1.0
		for _, parameter := range parts[1:] {
			name, raw, found := strings.Cut(strings.TrimSpace(parameter), "=")
			if !found || !strings.EqualFold(strings.TrimSpace(name), "q") {
				continue
			}
			parsed, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
			if err != nil || parsed < 0 || parsed > 1 {
				return false
			}
			quality = parsed
		}
		return quality > 0
	}
	return false
}

func refreshLeaseCookie(w http.ResponseWriter, publicID, secret string) {
	http.SetCookie(w, LeaseCookie(publicID, secret))
}

func (controller *Controller) extendLease(w http.ResponseWriter, publicID, secret string) (LeaseSnapshot, error) {
	snapshot, err := controller.leases.Authenticate(publicID, secret, true, controller.now())
	if err != nil {
		return LeaseSnapshot{}, err
	}
	refreshLeaseCookie(w, publicID, secret)
	return snapshot, nil
}

func boundHLSRequest(w http.ResponseWriter) func() {
	controller := http.NewResponseController(w)
	_ = controller.SetReadDeadline(time.Now().Add(HeaderReadTimeout))
	_ = controller.SetWriteDeadline(time.Now().Add(ResponseWriteTimeout))
	return func() {
		_ = controller.SetReadDeadline(time.Time{})
		_ = controller.SetWriteDeadline(time.Time{})
	}
}

func bootstrapErrorResult(err error) string {
	switch hlsOpenErrorStatus(err) {
	case http.StatusServiceUnavailable:
		return "not_ready"
	case http.StatusTooManyRequests:
		return "rate_limited"
	case http.StatusForbidden:
		return "denied"
	default:
		return "backend_error"
	}
}

func boundedRequestResult(value string) string {
	switch value {
	case "success", "bad_request", "not_found", "not_ready", "paused", "rate_limited", "range_invalid", "timeout", "canceled", "write_error":
		return value
	default:
		return "backend_error"
	}
}
