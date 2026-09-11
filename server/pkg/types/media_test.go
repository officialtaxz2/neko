package types

import (
	"reflect"
	"strings"
	"testing"
)

func TestSelectMediaSourcePreservesConfiguredOrdering(t *testing.T) {
	sources := []MediaSource{
		{ID: "high", Bitrate: 2_000_000},
		{ID: "medium", Bitrate: 750_000},
		{ID: "low", Bitrate: 330_000},
	}

	tests := []struct {
		name     string
		selector MediaSelector
		want     string
	}{
		{name: "exact", selector: MediaSelector{ID: "medium"}, want: "medium"},
		{name: "next lower by id", selector: MediaSelector{ID: "medium", Type: MediaSelectorTypeLower}, want: "low"},
		{name: "next higher by id", selector: MediaSelector{ID: "medium", Type: MediaSelectorTypeHigher}, want: "high"},
		{name: "nearest prefers fitting source", selector: MediaSelector{Bitrate: 700_000, Type: MediaSelectorTypeNearest}, want: "low"},
		{name: "legacy lower bitrate chooses lowest", selector: MediaSelector{Bitrate: 1_000_000, Type: MediaSelectorTypeLower}, want: "low"},
		{name: "legacy higher bitrate chooses highest", selector: MediaSelector{Bitrate: 500_000, Type: MediaSelectorTypeHigher}, want: "high"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := SelectMediaSource(sources, test.selector)
			if !ok {
				t.Fatal("SelectMediaSource() did not resolve a source")
			}
			if got.ID != test.want {
				t.Fatalf("SelectMediaSource() = %q, want %q", got.ID, test.want)
			}
		})
	}
}

func TestCloneMediaSourceCopiesMutableCodecData(t *testing.T) {
	source := MediaSource{
		Codec: MediaCodec{
			Parameters: map[string]string{"profile": "original"},
			Config:     []byte{1, 2, 3},
		},
	}
	clone := CloneMediaSource(source)
	clone.Codec.Parameters["profile"] = "changed"
	clone.Codec.Config[0] = 9

	if source.Codec.Parameters["profile"] != "original" {
		t.Fatal("CloneMediaSource() shared the codec parameter map")
	}
	if source.Codec.Config[0] != 1 {
		t.Fatal("CloneMediaSource() shared the codec config buffer")
	}
}

func TestSelectMediaSourceIgnoresUnavailableBitrates(t *testing.T) {
	sources := []MediaSource{
		{ID: "high", Bitrate: 2_000_000},
		{ID: "medium", Bitrate: 750_000},
		{ID: "low", Bitrate: 0},
	}
	source, ok := SelectMediaSource(sources, MediaSelector{
		Type: MediaSelectorTypeLower, Bitrate: 1_000_000,
	})
	if !ok || source.ID != "medium" {
		t.Fatalf("lower selection with unavailable low source = %#v/%v, want medium/true", source, ok)
	}
}

func TestMediaLeaseAndMetricDimensionsExposeNoCredentials(t *testing.T) {
	lease := reflect.TypeOf((*MediaLease)(nil)).Elem()
	for index := 0; index < lease.NumMethod(); index++ {
		name := strings.ToLower(lease.Method(index).Name)
		if strings.Contains(name, "token") || strings.Contains(name, "password") || strings.Contains(name, "credential") {
			t.Fatalf("MediaLease exposes credential-like method %q", name)
		}
	}

	request := reflect.TypeOf(MediaDeliveryRequest{})
	for index := 0; index < request.NumField(); index++ {
		name := strings.ToLower(request.Field(index).Name)
		if strings.Contains(name, "token") || strings.Contains(name, "password") || strings.Contains(name, "credential") {
			t.Fatalf("MediaDeliveryRequest exposes credential-like field %q", name)
		}
	}
}
