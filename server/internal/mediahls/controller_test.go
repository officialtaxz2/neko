package mediahls

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestPlaylistHeadMatchesCompressedGetWithoutBody(t *testing.T) {
	controller := &Controller{}
	data := []byte("#EXTM3U\n#EXT-X-VERSION:9\n")

	getRequest := httptest.NewRequest(http.MethodGet, "https://neko.example/index.m3u8", nil)
	getRequest.Header.Set("Accept-Encoding", "gzip")
	getResponse := httptest.NewRecorder()
	if result := controller.writePlaylist(getResponse, getRequest, data); result != "success" {
		t.Fatalf("GET result = %q", result)
	}

	headRequest := httptest.NewRequest(http.MethodHead, "https://neko.example/index.m3u8", nil)
	headRequest.Header.Set("Accept-Encoding", "gzip")
	headResponse := httptest.NewRecorder()
	if result := controller.writePlaylist(headResponse, headRequest, data); result != "success" {
		t.Fatalf("HEAD result = %q", result)
	}

	if headResponse.Body.Len() != 0 {
		t.Fatalf("HEAD body length = %d", headResponse.Body.Len())
	}
	if got, want := headResponse.Header().Get("Content-Length"), getResponse.Header().Get("Content-Length"); got != want {
		t.Fatalf("HEAD length = %q, GET length = %q", got, want)
	}
	if headResponse.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal("HEAD did not preserve gzip representation headers")
	}

	reader, err := gzip.NewReader(bytes.NewReader(getResponse.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded, data) {
		t.Fatalf("decoded playlist = %q", decoded)
	}
}

func TestObjectRangeAndHeadAreBounded(t *testing.T) {
	controller := &Controller{}
	object, err := NewMediaObject("seg-42.m4s", "high", ObjectSegment, 7, 42, 0, "video/mp4", []byte{0, 1, 2, 3, 4, 5})
	if err != nil {
		t.Fatal(err)
	}

	rangeRequest := httptest.NewRequest(http.MethodGet, "https://neko.example/seg-42.m4s", nil)
	rangeRequest.Header.Set("Range", "bytes=2-4")
	rangeResponse := httptest.NewRecorder()
	if result := controller.writeObject(rangeResponse, rangeRequest, object); result != "success" {
		t.Fatalf("range result = %q", result)
	}
	if rangeResponse.Code != http.StatusPartialContent || rangeResponse.Header().Get("Content-Range") != "bytes 2-4/6" || !bytes.Equal(rangeResponse.Body.Bytes(), []byte{2, 3, 4}) {
		t.Fatalf("range response = code %d, headers %#v, body %v", rangeResponse.Code, rangeResponse.Header(), rangeResponse.Body.Bytes())
	}

	headRequest := httptest.NewRequest(http.MethodHead, "https://neko.example/seg-42.m4s", nil)
	headResponse := httptest.NewRecorder()
	if result := controller.writeObject(headResponse, headRequest, object); result != "success" {
		t.Fatalf("HEAD result = %q", result)
	}
	if headResponse.Body.Len() != 0 || headResponse.Header().Get("Content-Length") != strconv.Itoa(object.Size()) {
		t.Fatalf("HEAD response = headers %#v, body length %d", headResponse.Header(), headResponse.Body.Len())
	}

	invalidRequest := httptest.NewRequest(http.MethodGet, "https://neko.example/seg-42.m4s", nil)
	invalidRequest.Header.Set("Range", "bytes=0-1,3-4")
	invalidResponse := httptest.NewRecorder()
	if result := controller.writeObject(invalidResponse, invalidRequest, object); result != "range_invalid" || invalidResponse.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("multi-range response = result %q, code %d", result, invalidResponse.Code)
	}
}

func TestLeaseRefreshCookieRemainsNarrowAndCredentialSafe(t *testing.T) {
	_, offer, _ := testLeaseStore(t)
	response := httptest.NewRecorder()
	refreshLeaseCookie(response, offer.PublicID, offer.Secret)
	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != LeaseCookieName || cookie.Value != offer.Secret || cookie.Path != "/api/media/hls/"+offer.PublicID+"/" || cookie.MaxAge != int(LeaseLifetime.Seconds()) || !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("refresh cookie = %#v", cookie)
	}
}

func TestGzipNegotiationRejectsZeroOrInvalidQuality(t *testing.T) {
	if !acceptsGzip("br, gzip; q=0.5") {
		t.Fatal("positive gzip quality was rejected")
	}
	for _, value := range []string{"gzip;q=0", "gzip;q=bogus", "br"} {
		if acceptsGzip(value) {
			t.Fatalf("gzip accepted for %q", value)
		}
	}
}
