package legacy

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
