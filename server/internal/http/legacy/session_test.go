package legacy

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

func TestLegacySessionCancellationInterruptsLoopbackRequest(t *testing.T) {
	started := make(chan struct{})
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
		close(started)
		select {
		case <-r.Context().Done():
		case <-time.After(3 * time.Second):
		}
	}))
	defer backend.Close()
	handler := New(strings.TrimPrefix(backend.URL, "http://"), "")
	session := handler.newSession(httptest.NewRequest(http.MethodGet, "/ws", nil))
	defer session.cancel()
	defer session.client.CloseIdleConnections()
	result := make(chan error, 1)
	go func() { result <- session.apiReq(http.MethodGet, "/api/room/settings", nil, nil) }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("loopback request did not start")
	}
	session.cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("request not canceled: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("loopback request survived session cancellation")
	}
}

func TestLegacyStreamingDoesNotGetAnAPITotalDeadline(t *testing.T) {
	session := New("127.0.0.1:8080", "").newSession(httptest.NewRequest(http.MethodGet, "/file", nil))
	defer session.cancel()
	defer session.client.CloseIdleConnections()
	if session.client.Timeout != 0 {
		t.Fatal("large transfers inherited API timeout")
	}
	transport := session.client.Transport.(*http.Transport)
	if transport.ResponseHeaderTimeout <= 0 {
		t.Fatal("response headers remain unbounded")
	}
}

func TestCreateUsesExistingSessionCookie(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/whoami":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			if got := r.Header.Get("Cookie"); got != "NEKO_SESSION=session-token" {
				t.Fatalf("cookie = %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "oauth:user-123",
				"profile": types.MemberProfile{
					Name:    "Ada Lovelace",
					IsAdmin: true,
				},
			})
		case "/api/protected":
			if got := r.Header.Get("Cookie"); got != "NEKO_SESSION=session-token" {
				t.Fatalf("authenticated request cookie = %q", got)
			}
			_ = json.NewEncoder(w).Encode(true)
		default:
			http.NotFound(w, r)
		}
	}))
	defer backend.Close()

	request := httptest.NewRequest(http.MethodGet, "http://neko.test/ws", nil)
	request.Header.Set("Cookie", "NEKO_SESSION=session-token")
	handler := &LegacyHandler{sessionIPs: make(map[string]string)}
	session := &session{
		r:          request,
		h:          handler,
		serverAddr: strings.TrimPrefix(backend.URL, "http://"),
		ctx:        context.Background(),
		client:     backend.Client(),
	}

	if err := session.create("", ""); err != nil {
		t.Fatal(err)
	}
	if session.id != "oauth:user-123" || session.name != "Ada Lovelace" || !session.isAdmin {
		t.Fatalf("session = %#v", session)
	}
	if session.cookie != "NEKO_SESSION=session-token" || session.token != "" {
		t.Fatalf("authentication state = cookie %q token %q", session.cookie, session.token)
	}
	if err := session.apiReq(http.MethodGet, "/api/protected", nil, nil); err != nil {
		t.Fatal(err)
	}
}
