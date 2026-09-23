package mediahls

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrRequestRejected = errors.New("HLS request rejected")
	ErrInvalidPath     = errors.New("invalid HLS resource path")
	ErrInvalidQuery    = errors.New("invalid HLS playlist query")
	ErrInvalidRange    = errors.New("invalid HLS byte range")
)

type RequestError struct { Status int; Reason string }

func (err *RequestError) Error() string { return fmt.Sprintf("%v: %s", ErrRequestRejected, err.Reason) }
func (err *RequestError) Unwrap() error { return ErrRequestRejected }

func rejectRequest(status int, reason string) error { return &RequestError{Status: status, Reason: reason} }

func RequestStatus(err error) int {
	var requestError *RequestError
	if errors.As(err, &requestError) { return requestError.Status }
	return http.StatusInternalServerError
}

type SecurityPolicy struct {
	origins map[string]struct{}
	proxies []netip.Prefix
}

func NewSecurityPolicy(allowedOrigins, trustedProxies []string) (*SecurityPolicy, error) {
	if len(allowedOrigins) == 0 {
		return nil, fmt.Errorf("%w: at least one exact HTTPS origin is required", ErrInvalidConfig)
	}
	origins := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.Path != "" || parsed.ForceQuery || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.String() != origin {
			return nil, fmt.Errorf("%w: allowed origins must be exact HTTPS origins", ErrInvalidConfig)
		}
		if _, exists := origins[origin]; exists {
			return nil, fmt.Errorf("%w: duplicate allowed origin", ErrInvalidConfig)
		}
		origins[origin] = struct{}{}
	}
	proxies := make([]netip.Prefix, 0, len(trustedProxies))
	for _, value := range trustedProxies {
		value = strings.TrimSpace(value)
		if value == "" { continue }
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			address, addressErr := netip.ParseAddr(value)
			if addressErr != nil {
				return nil, fmt.Errorf("%w: trusted proxy must be an IP or CIDR", ErrInvalidConfig)
			}
			prefix = netip.PrefixFrom(address, address.BitLen())
		}
		proxies = append(proxies, prefix.Masked())
	}
	return &SecurityPolicy{origins: origins, proxies: proxies}, nil
}

func (policy *SecurityPolicy) CheckBootstrap(request *http.Request) error {
	if request == nil || request.Method != http.MethodPost || request.URL == nil || !canonicalRequestPath(request.URL) || request.URL.Path != BootstrapPath || request.URL.RawQuery != "" || request.URL.ForceQuery || request.Header.Get("Content-Type") != "application/json" {
		return rejectRequest(http.StatusBadRequest, "invalid_bootstrap_request")
	}
	if !policy.secureTransport(request) || !policy.exactOrigin(request.Header.Get("Origin")) {
		return rejectRequest(http.StatusForbidden, "bootstrap_origin_rejected")
	}
	return nil
}

func (policy *SecurityPolicy) CheckKeepAlive(request *http.Request) error {
	if request == nil || request.Method != http.MethodPost || request.URL == nil || !canonicalRequestPath(request.URL) || request.URL.RawQuery != "" || request.URL.ForceQuery {
		return rejectRequest(http.StatusBadRequest, "invalid_keepalive_request")
	}
	resource, err := ParseResourcePath(request.URL.Path)
	if err != nil || resource.Kind != ResourceKeepAlive {
		return rejectRequest(http.StatusBadRequest, "invalid_keepalive_path")
	}
	if !policy.secureTransport(request) || !policy.exactOrigin(request.Header.Get("Origin")) {
		return rejectRequest(http.StatusForbidden, "keepalive_rejected")
	}
	return nil
}

func (policy *SecurityPolicy) CheckMedia(request *http.Request) error {
	if request == nil || (request.Method != http.MethodGet && request.Method != http.MethodHead) || request.URL == nil || !canonicalRequestPath(request.URL) {
		return rejectRequest(http.StatusBadRequest, "invalid_media_request")
	}
	if !policy.secureTransport(request) {
		return rejectRequest(http.StatusForbidden, "media_request_rejected")
	}
	if site := strings.ToLower(strings.TrimSpace(request.Header.Get("Sec-Fetch-Site"))); site == "cross-site" {
		return rejectRequest(http.StatusForbidden, "media_cross_site_rejected")
	}
	if origin := request.Header.Get("Origin"); origin != "" && !policy.exactOrigin(origin) {
		return rejectRequest(http.StatusForbidden, "media_origin_rejected")
	}
	if _, err := ParseResourcePath(request.URL.Path); err != nil {
		return rejectRequest(http.StatusBadRequest, "media_path_rejected")
	}
	return nil
}

func (policy *SecurityPolicy) exactOrigin(origin string) bool {
	_, ok := policy.origins[origin]
	return ok
}

func (policy *SecurityPolicy) secureTransport(request *http.Request) bool {
	if request.TLS != nil { return true }
	peer, ok := remoteAddress(request.RemoteAddr)
	if !ok || !policy.trusted(peer) { return false }
	forwarded := strings.TrimSpace(request.Header.Get("Forwarded"))
	xfp := strings.TrimSpace(request.Header.Get("X-Forwarded-Proto"))
	if forwarded != "" && xfp != "" { return false }
	if forwarded != "" {
		if strings.Contains(forwarded, ",") { return false }
		for _, parameter := range strings.Split(forwarded, ";") {
			name, value, found := strings.Cut(strings.TrimSpace(parameter), "=")
			if found && strings.EqualFold(name, "proto") {
				return strings.EqualFold(strings.Trim(value, `"`), "https")
			}
		}
		return false
	}
	return !strings.Contains(xfp, ",") && strings.EqualFold(xfp, "https")
}

func canonicalRequestPath(value *url.URL) bool {
	return value != nil && value.RawPath == "" && value.EscapedPath() == value.Path
}

func (policy *SecurityPolicy) trusted(address netip.Addr) bool {
	for _, prefix := range policy.proxies {
		if prefix.Contains(address) { return true }
	}
	return false
}

func remoteAddress(value string) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(value)
	if err != nil { host = value }
	address, err := netip.ParseAddr(strings.Trim(host, "[]"))
	return address, err == nil
}

func ApplySecurityHeaders(header http.Header) {
	header.Set("Referrer-Policy", "no-referrer")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Cache-Control", "private, no-store, max-age=0")
	header.Set("Pragma", "no-cache")
	header.Set("Vary", "Cookie, Accept-Encoding")
}

type ResourceKind string

const (
	ResourceMaster    ResourceKind = "master"
	ResourcePlaylist  ResourceKind = "playlist"
	ResourceInit      ResourceKind = "init"
	ResourceSegment   ResourceKind = "segment"
	ResourcePart      ResourceKind = "part"
	ResourceKeepAlive ResourceKind = "keepalive"
)

type ResourcePath struct {
	PublicID  string
	Variant   string
	Kind      ResourceKind
	Object    string
	Generation uint64
	Sequence  uint64
	Part      uint64
}

var objectNamePattern = regexp.MustCompile(`^(init-([1-9][0-9]*)\.mp4|seg-([1-9][0-9]*)\.m4s|part-([1-9][0-9]*)-(0|[1-9][0-9]*)\.m4s)$`)

func ParseResourcePath(path string) (ResourcePath, error) {
	if !strings.HasPrefix(path, "/api/media/hls/") || strings.Contains(path, "%") || strings.Contains(path, "//") || strings.Contains(path, "..") || strings.HasSuffix(path, "/") {
		return ResourcePath{}, ErrInvalidPath
	}
	parts := strings.Split(strings.TrimPrefix(path, "/api/media/hls/"), "/")
	if len(parts) < 2 || ValidatePublicID(parts[0]) != nil {
		return ResourcePath{}, ErrInvalidPath
	}
	result := ResourcePath{PublicID: parts[0]}
	if len(parts) == 2 {
		switch parts[1] {
		case "master.m3u8": result.Kind, result.Object = ResourceMaster, parts[1]
		case "keepalive": result.Kind, result.Object = ResourceKeepAlive, parts[1]
		default: return ResourcePath{}, ErrInvalidPath
		}
		return result, nil
	}
	if len(parts) != 3 || (parts[1] != "audio" && parts[1] != "high" && parts[1] != "medium" && parts[1] != "low") {
		return ResourcePath{}, ErrInvalidPath
	}
	result.Variant, result.Object = parts[1], parts[2]
	if parts[2] == "index.m3u8" { result.Kind = ResourcePlaylist; return result, nil }
	matches := objectNamePattern.FindStringSubmatch(parts[2])
	if matches == nil { return ResourcePath{}, ErrInvalidPath }
	parse := func(value string) (uint64, error) { return strconv.ParseUint(value, 10, 64) }
	var err error
	switch {
	case matches[2] != "":
		result.Kind = ResourceInit
		result.Generation, err = parse(matches[2])
	case matches[3] != "":
		result.Kind = ResourceSegment
		result.Sequence, err = parse(matches[3])
	default:
		result.Kind = ResourcePart
		result.Sequence, err = parse(matches[4])
		if err == nil { result.Part, err = parse(matches[5]) }
	}
	if err != nil { return ResourcePath{}, ErrInvalidPath }
	return result, nil
}

func ValidatePublicID(publicID string) error {
	if len(publicID) != PublicIDEncodedLength { return ErrInvalidPath }
	raw, err := base64.RawURLEncoding.DecodeString(publicID)
	if err != nil || len(raw) != PublicIDEntropyBytes { return ErrInvalidPath }
	return nil
}

type PlaylistDirectives struct { MSN uint64; Part uint64; HasMSN bool; HasPart bool }

func ParsePlaylistQuery(mode string, values url.Values, currentMSN, currentPart uint64) (PlaylistDirectives, error) {
	if mode == ModeHLS {
		if len(values) != 0 { return PlaylistDirectives{}, ErrInvalidQuery }
		return PlaylistDirectives{}, nil
	}
	if mode != ModeLLHLS { return PlaylistDirectives{}, ErrInvalidMode }
	for key, all := range values {
		if (key != "_HLS_msn" && key != "_HLS_part") || len(all) != 1 { return PlaylistDirectives{}, ErrInvalidQuery }
	}
	parse := func(key string) (uint64, bool) {
		all, ok := values[key]
		if !ok { return 0, false }
		value := all[0]
		if value == "" || (len(value) > 1 && value[0] == '0') || strings.HasPrefix(value, "+") || strings.HasPrefix(value, "-") { return 0, false }
		parsed, err := strconv.ParseUint(value, 10, 64)
		return parsed, err == nil
	}
	msn, hasMSN := parse("_HLS_msn")
	part, hasPart := parse("_HLS_part")
	if _, present := values["_HLS_msn"]; present && !hasMSN { return PlaylistDirectives{}, ErrInvalidQuery }
	if _, present := values["_HLS_part"]; present && !hasPart { return PlaylistDirectives{}, ErrInvalidQuery }
	if hasPart && !hasMSN { return PlaylistDirectives{}, ErrInvalidQuery }
	if hasMSN && msn > currentMSN+2 { return PlaylistDirectives{}, ErrInvalidQuery }
	if hasPart && part > currentPart+3 { return PlaylistDirectives{}, ErrInvalidQuery }
	return PlaylistDirectives{MSN: msn, Part: part, HasMSN: hasMSN, HasPart: hasPart}, nil
}

type ByteRange struct { Start int64; End int64 }

func ParseByteRange(value string, size int64) (ByteRange, error) {
	if size <= 0 || value == "" || !strings.HasPrefix(value, "bytes=") || strings.Contains(value, ",") {
		return ByteRange{}, ErrInvalidRange
	}
	startText, endText, ok := strings.Cut(strings.TrimPrefix(value, "bytes="), "-")
	if !ok || (startText == "" && endText == "") { return ByteRange{}, ErrInvalidRange }
	if startText == "" {
		suffix, err := strconv.ParseInt(endText, 10, 64)
		if err != nil || suffix <= 0 { return ByteRange{}, ErrInvalidRange }
		if suffix > size { suffix = size }
		return ByteRange{Start: size - suffix, End: size - 1}, nil
	}
	start, err := strconv.ParseInt(startText, 10, 64)
	if err != nil || start < 0 || start >= size { return ByteRange{}, ErrInvalidRange }
	end := size - 1
	if endText != "" {
		end, err = strconv.ParseInt(endText, 10, 64)
		if err != nil || end < start { return ByteRange{}, ErrInvalidRange }
		if end >= size { end = size - 1 }
	}
	return ByteRange{Start: start, End: end}, nil
}

func NormalizeAccessPath(path string) string {
	if path == BootstrapPath { return BootstrapPath }
	if !strings.HasPrefix(path, "/api/media/hls/") { return path }
	return "/api/media/hls/:lease/:variant/:object"
}
