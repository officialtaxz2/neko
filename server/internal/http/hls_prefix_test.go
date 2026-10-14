package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrefixedHLSRoutesRemainOutsideGenericCORS(t *testing.T) {
	router := newRouter(WithPathPrefix("/room"), WithCORS(func(string) bool { return true }))
	serve := func(w http.ResponseWriter, request *http.Request) error {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	router.Post("/api/media/hls/session", serve)
	router.Get("/health", serve)
	for _, path := range []string{"/room/api/media/hls/session", "/room/health"} {
		method := http.MethodGet
		if path == "/room/api/media/hls/session" {
			method = http.MethodPost
		}
		request := httptest.NewRequest(method, "https://neko.example"+path, nil)
		request.Header.Set("Origin", "https://other.example")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusNoContent {
			t.Fatalf("prefixed route returned %d", response.Code)
		}
		cors := response.Header().Get("Access-Control-Allow-Origin")
		if method == http.MethodPost && cors != "" {
			t.Fatal("HLS inherited generic CORS headers")
		}
		if method == http.MethodGet && cors == "" {
			t.Fatal("ordinary route lost generic CORS handling")
		}
	}
}
