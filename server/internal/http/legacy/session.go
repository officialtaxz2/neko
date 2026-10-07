package legacy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	oldTypes "github.com/m1k1o/neko/server/internal/http/legacy/types"

	"github.com/m1k1o/neko/server/internal/api"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

var (
	ErrWebsocketSend  = fmt.Errorf("failed to send message to websocket")
	ErrBackendRespone = fmt.Errorf("error response from backend")
)

type memberStruct struct {
	member    *oldTypes.Member
	connected bool
	sent      bool
}

type session struct {
	r      *http.Request
	h      *LegacyHandler
	ctx    context.Context
	cancel context.CancelFunc

	logger     zerolog.Logger
	serverAddr string
	pathPrefix string

	id, ip     string
	token      string
	name       string
	isAdmin    bool
	isViewOnly bool
	client     *http.Client

	lastHostID         string
	lockedControls     bool
	lockedLogins       bool
	lockedFileTransfer bool
	sessions           map[string]*memberStruct

	connClient    *websocket.Conn
	connBackend   *websocket.Conn
	clientWriter  *utils.WebSocketWriter
	backendWriter *utils.WebSocketWriter
}

func (h *LegacyHandler) newSession(r *http.Request) *session {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil // disable proxy for local requests
	transport.ResponseHeaderTimeout = 15 * time.Second
	ctx, cancel := context.WithCancel(r.Context())

	return &session{
		r:          r,
		h:          h,
		ctx:        ctx,
		cancel:     cancel,
		logger:     h.logger,
		serverAddr: h.serverAddr,
		pathPrefix: h.pathPrefix,
		client: &http.Client{
			Transport: transport,
		},
		sessions: make(map[string]*memberStruct),
	}
}

func (s *session) req(method, reqPath string, headers http.Header, request io.Reader) (io.ReadCloser, http.Header, error) {
	return s.reqWithContext(s.ctx, method, reqPath, headers, request)
}

func (s *session) reqWithContext(ctx context.Context, method, reqPath string, headers http.Header, request io.Reader) (io.ReadCloser, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, method, "http://"+s.serverAddr+path.Join(s.pathPrefix, reqPath), request)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range headers {
		req.Header[k] = v
	}

	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		defer res.Body.Close()

		body, _ := io.ReadAll(res.Body)
		// try to unmarsal as json error message
		var apiErr struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return nil, nil, fmt.Errorf("%w: %s", ErrBackendRespone, apiErr.Message)
		}
		// return raw body if failed to unmarshal
		return nil, nil, fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	return res.Body, res.Header, nil
}

func (s *session) apiReq(method, path string, request, response any) error {
	ctx, cancel := context.WithTimeout(s.ctx, 15*time.Second)
	defer cancel()
	return s.apiReqWithContext(ctx, method, path, request, response)
}

func (s *session) apiReqWithContext(ctx context.Context, method, path string, request, response any) error {
	reqBody, err := json.Marshal(request)
	if err != nil {
		return err
	}

	headers := http.Header{
		"Content-Type": []string{"application/json"},
	}

	resBody, _, err := s.reqWithContext(ctx, method, path, headers, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resBody.Close()

	if resBody == nil {
		return nil
	}

	if response == nil {
		io.Copy(io.Discard, resBody)
		return nil
	}

	return json.NewDecoder(resBody).Decode(response)
}

// send message to client (in old format)
func (s *session) toClient(payload any) error {
	err := s.clientWriter.SendJSON(payload)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrWebsocketSend, err)
	}

	return nil
}

// send message to backend (in new format)
func (s *session) toBackend(event string, payload any) error {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = s.backendWriter.SendJSON(types.WebSocketMessage{
		Event:   event,
		Payload: rawPayload,
	})
	if err != nil {
		return fmt.Errorf("%w: %s", ErrWebsocketSend, err)
	}

	return nil
}

func (s *session) create(username, password string) error {
	data := api.SessionDataPayload{}

	err := s.apiReq(http.MethodPost, "/api/login", api.SessionLoginPayload{
		Username: username,
		Password: password,
	}, &data)
	if err != nil {
		return err
	}

	s.id, s.ip = data.ID, getIp(s.r)
	s.h.mu.Lock()
	s.h.sessionIPs[s.id] = s.ip // save session ip by id
	s.h.mu.Unlock()
	s.token = data.Token
	s.name = data.Profile.Name
	s.isAdmin = data.Profile.IsAdmin
	s.isViewOnly = data.Profile.IsViewOnly

	// if Cookie auth, the token will be empty
	if s.token == "" {
		return fmt.Errorf("token not found - make sure you are not using Cookie auth on the server")
	}

	return nil
}

func (s *session) destroy() {
	defer s.client.CloseIdleConnections()
	s.cancel()

	// Session cleanup must survive the canceled bridge/request context, but
	// cannot wait indefinitely. File bodies retain their request lifetime.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := s.apiReqWithContext(ctx, http.MethodPost, "/api/logout", nil, nil)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to logout")
	}

	// remove session id from ip map
	s.h.mu.Lock()
	delete(s.h.sessionIPs, s.id)
	s.h.mu.Unlock()
}
