package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type originalRemoteAddrContextKey struct{}

func SetOriginalRemoteAddr(ctx context.Context, address string) context.Context {
	return context.WithValue(ctx, originalRemoteAddrContextKey{}, address)
}

func OriginalRemoteAddr(r *http.Request) string {
	if address, ok := r.Context().Value(originalRemoteAddrContextKey{}).(string); ok && address != "" {
		return address
	}
	return r.RemoteAddr
}

func HttpRequestGET(url string) (string, error) {
	// Used only for configured external-IP discovery, never file transfers.
	client := &http.Client{Timeout: 15 * time.Second}
	rsp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		return "", fmt.Errorf("IP discovery returned HTTP %d", rsp.StatusCode)
	}
	const maxResponse = 4096
	buf, err := io.ReadAll(io.LimitReader(rsp.Body, maxResponse+1))
	if err != nil {
		return "", err
	}
	if len(buf) > maxResponse {
		return "", fmt.Errorf("IP discovery response too large")
	}

	return string(bytes.TrimSpace(buf)), nil
}
