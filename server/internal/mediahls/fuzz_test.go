package mediahls

import (
	"net/url"
	"testing"
)

// Exercise unauthenticated request parsers together without issuing credentials,
// opening a delivery or invoking GStreamer. Panics fail the fuzz target.
func FuzzHLSRequestBoundary(f *testing.F) {
	f.Add("/api/media/hls/AAAAAAAAAAAAAAAAAAAAAA/high/index.m3u8", "_HLS_msn=1&_HLS_part=6", "bytes=0-31", []byte(`{"ticket":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}`))
	f.Add("/api/media/hls/short/../master.m3u8", "_HLS_skip=YES&_HLS_part=-1", "bytes=0-1,4-5", []byte(`{"ticket":"short","extra":true}`))
	f.Fuzz(func(t *testing.T, resourcePath, query, byteRange string, body []byte) {
		if len(resourcePath)+len(query)+len(byteRange)+len(body) > 64*1024 {
			t.Skip()
		}
		_, _ = ParseResourcePath(resourcePath)
		if values, err := url.ParseQuery(query); err == nil {
			_, _ = ParsePlaylistQuery(ModeHLS, values, 1, 0)
			_, _ = ParsePlaylistQuery(ModeLLHLS, values, 1, 0)
		}
		_, _ = ParseByteRange(byteRange, 1024)
		_, _ = DecodeBootstrapPayload(body)
	})
}
