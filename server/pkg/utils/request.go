package utils

import (
	"bytes"
	"context"
	"io"
	"net/http"
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
	rsp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	buf, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(buf)), nil
}
