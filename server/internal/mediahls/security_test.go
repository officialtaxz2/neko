package mediahls

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"testing"
)

const testPublicID = "AAAAAAAAAAAAAAAAAAAAAA"

func TestSecurityPolicyRequiresHTTPSExactOriginAndTrustedProxy(t *testing.T) {
	policy, err := NewSecurityPolicy([]string{"https://neko.example"}, []string{"127.0.0.1/32"})
	if err != nil { t.Fatal(err) }
	request := &http.Request{Method:http.MethodPost, URL:&url.URL{Path:BootstrapPath}, Header:http.Header{"Origin":[]string{"https://neko.example"},"Content-Type":[]string{"application/json"}}, RemoteAddr:"203.0.113.5:1234"}
	request.Header.Set("X-Forwarded-Proto","https")
	if check:=policy.CheckBootstrap(request);!errors.Is(check,ErrRequestRejected)||RequestStatus(check)!=http.StatusForbidden { t.Fatalf("untrusted forwarded HTTPS result = %v",check) }
	request.RemoteAddr="127.0.0.1:1234"
	if err:=policy.CheckBootstrap(request); err!=nil { t.Fatalf("trusted proxy rejected: %v",err) }
	request.Header.Set("Forwarded","proto=https")
	if !errors.Is(policy.CheckBootstrap(request),ErrRequestRejected) { t.Fatal("ambiguous forwarded transport accepted") }
	request.Header.Del("Forwarded")
	request.RemoteAddr="203.0.113.5:1234"; request.TLS=&tls.ConnectionState{}; request.Header.Del("X-Forwarded-Proto")
	if err:=policy.CheckBootstrap(request); err!=nil { t.Fatalf("direct TLS rejected: %v",err) }
	request.Header.Set("Origin","https://evil.example")
	if !errors.Is(policy.CheckBootstrap(request),ErrRequestRejected) { t.Fatal("wrong origin accepted") }
}

func TestMediaSecurityAllowsMissingOriginButRejectsCrossSite(t *testing.T) {
	policy, err := NewSecurityPolicy([]string{"https://neko.example"},nil)
	if err!=nil { t.Fatal(err) }
	request:=&http.Request{Method:http.MethodGet,URL:&url.URL{Path:"/api/media/hls/"+testPublicID+"/high/index.m3u8"},Header:http.Header{},TLS:&tls.ConnectionState{}}
	if err:=policy.CheckMedia(request); err!=nil { t.Fatal(err) }
	request.Header.Set("Sec-Fetch-Site","cross-site")
	if !errors.Is(policy.CheckMedia(request),ErrRequestRejected) { t.Fatal("cross-site request accepted") }
	request.Header.Del("Sec-Fetch-Site");request.URL.RawPath="/api/media/hls/"+testPublicID+"/%68igh/index.m3u8"
	if !errors.Is(policy.CheckMedia(request),ErrRequestRejected) { t.Fatal("non-canonical escaped path accepted") }
}

func TestResourcePathQueryRangeAndRedaction(t *testing.T) {
	paths:=map[string]ResourceKind{
		"/api/media/hls/"+testPublicID+"/master.m3u8":ResourceMaster,
		"/api/media/hls/"+testPublicID+"/audio/init-7.mp4":ResourceInit,
		"/api/media/hls/"+testPublicID+"/high/seg-42.m4s":ResourceSegment,
		"/api/media/hls/"+testPublicID+"/low/part-42-0.m4s":ResourcePart,
	}
	for path,kind:=range paths { parsed,err:=ParseResourcePath(path); if err!=nil||parsed.Kind!=kind { t.Fatalf("path %q = %#v, %v",path,parsed,err) } }
	for _,path:=range []string{"/api/media/hls/"+testPublicID+"/../master.m3u8","/api/media/hls/"+testPublicID+"/high/index.m3u8/","/api/media/hls/not-an-id/master.m3u8"} { if _,err:=ParseResourcePath(path); !errors.Is(err,ErrInvalidPath) { t.Fatalf("accepted path %q",path) } }
	valid:=url.Values{"_HLS_msn":[]string{"42"},"_HLS_part":[]string{"3"}}
	if _,err:=ParsePlaylistQuery(ModeLLHLS,valid,42,2); err!=nil { t.Fatal(err) }
	invalidQueries:=[]url.Values{{"_HLS_part":[]string{"1"}},{"_HLS_skip":[]string{"YES"}},{"_HLS_msn":[]string{"042"}},{"_HLS_msn":[]string{"45"}}}
	for _,values:=range invalidQueries { if _,err:=ParsePlaylistQuery(ModeLLHLS,values,42,2); !errors.Is(err,ErrInvalidQuery) { t.Fatalf("accepted query %#v",values) } }
	rangeValue,err:=ParseByteRange("bytes=10-19",100); if err!=nil||rangeValue.Start!=10||rangeValue.End!=19 { t.Fatalf("range = %#v, %v",rangeValue,err) }
	for _,value:=range []string{"bytes=10-20,30-40","bytes=100-","items=1-2"} { if _,err:=ParseByteRange(value,100); !errors.Is(err,ErrInvalidRange) { t.Fatalf("accepted range %q",value) } }
	if normalized:=NormalizeAccessPath("/api/media/hls/"+testPublicID+"/high/seg-42.m4s"); normalized!="/api/media/hls/:lease/:variant/:object"||containsCredential(normalized,testPublicID) { t.Fatalf("normalized path = %q",normalized) }
	header:=http.Header{};ApplySecurityHeaders(header)
	if header.Get("Referrer-Policy")!="no-referrer"||header.Get("X-Content-Type-Options")!="nosniff"||header.Get("Cache-Control")!="private, no-store, max-age=0"||header.Get("Vary")!="Cookie, Accept-Encoding"{t.Fatalf("security headers = %#v",header)}
}

func containsCredential(value, credential string) bool { return len(credential)>0 && len(value)>=len(credential) && func()bool{ for i:=0;i+len(credential)<=len(value);i++ { if value[i:i+len(credential)]==credential{return true} }; return false }() }
