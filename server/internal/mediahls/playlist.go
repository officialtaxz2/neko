package mediahls

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"time"
)

var ErrInvalidPlaylist = errors.New("invalid HLS playlist model")

type MasterVariant struct {
	ID string
	URI string
	Bandwidth uint64
	AverageBandwidth uint64
	Width uint32
	Height uint32
	FrameRate float64
	VideoCodec string
}

type MasterPlaylist struct { AudioURI string; Variants []MasterVariant }

func (playlist MasterPlaylist) Render() ([]byte, error) {
	if !relativePlaylistURI(playlist.AudioURI) || len(playlist.Variants) == 0 || len(playlist.Variants) > 3 { return nil, ErrInvalidPlaylist }
	variants := append([]MasterVariant(nil), playlist.Variants...)
	rank := map[string]int{"high": 0, "medium": 1, "low": 2}
	sort.SliceStable(variants, func(i,j int) bool { return rank[variants[i].ID] < rank[variants[j].ID] })
	var builder strings.Builder
	builder.WriteString("#EXTM3U\n#EXT-X-VERSION:7\n#EXT-X-INDEPENDENT-SEGMENTS\n")
	fmt.Fprintf(&builder, "#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"audio\",NAME=\"Audio\",DEFAULT=YES,AUTOSELECT=YES,URI=\"%s\"\n", playlist.AudioURI)
	seen := map[string]struct{}{}
	for _, variant := range variants {
		if _, exists := seen[variant.ID]; exists || (variant.ID != "high" && variant.ID != "medium" && variant.ID != "low") || !relativePlaylistURI(variant.URI) || variant.Bandwidth == 0 || variant.AverageBandwidth == 0 || variant.AverageBandwidth > variant.Bandwidth || variant.Width == 0 || variant.Height == 0 || variant.FrameRate <= 0 || variant.VideoCodec != "avc1.64001f" {
			return nil, ErrInvalidPlaylist
		}
		seen[variant.ID] = struct{}{}
		fmt.Fprintf(&builder, "#EXT-X-STREAM-INF:BANDWIDTH=%d,AVERAGE-BANDWIDTH=%d,RESOLUTION=%dx%d,FRAME-RATE=%.3f,CODECS=\"%s,mp4a.40.2\",AUDIO=\"audio\"\n%s\n", variant.Bandwidth, variant.AverageBandwidth, variant.Width, variant.Height, variant.FrameRate, variant.VideoCodec, variant.URI)
	}
	return boundedPlaylist(builder.String())
}

type Segment struct { URI string; Sequence uint64; Duration float64; ProgramDateTime time.Time; Discontinuity bool }
type Part struct { URI string; Duration float64; Independent bool }
type RenditionReport struct { URI string; LastMSN uint64; LastPart uint64 }

type MediaPlaylist struct {
	Mode string
	MediaSequence uint64
	DiscontinuitySequence uint64
	MapURI string
	Segments []Segment
	Parts []Part
	PartsProgramDateTime time.Time
	PreloadHint string
	RenditionReports []RenditionReport
}

func (playlist MediaPlaylist) Render() ([]byte, error) {
	if !ValidMode(playlist.Mode) || playlist.MediaSequence == 0 || !relativeInitURI(playlist.MapURI) || len(playlist.Segments) > 3 { return nil, ErrInvalidPlaylist }
	if playlist.Mode == ModeHLS && (len(playlist.Segments) != 3 || len(playlist.Parts) != 0 || playlist.PreloadHint != "" || len(playlist.RenditionReports) != 0) { return nil, ErrInvalidPlaylist }
	if playlist.Mode == ModeLLHLS && ((len(playlist.Segments) == 0 && len(playlist.Parts) < 3) || len(playlist.Parts) > 6 || playlist.PartsProgramDateTime.IsZero() || !relativeMediaURI(playlist.PreloadHint)) { return nil, ErrInvalidPlaylist }
	version := 7
	if playlist.Mode == ModeLLHLS { version = 9 }
	var builder strings.Builder
	fmt.Fprintf(&builder, "#EXTM3U\n#EXT-X-VERSION:%d\n#EXT-X-TARGETDURATION:6\n#EXT-X-MEDIA-SEQUENCE:%d\n#EXT-X-DISCONTINUITY-SEQUENCE:%d\n", version, playlist.MediaSequence, playlist.DiscontinuitySequence)
	if playlist.Mode == ModeLLHLS {
		builder.WriteString("#EXT-X-PART-INF:PART-TARGET=1\n#EXT-X-SERVER-CONTROL:CAN-BLOCK-RELOAD=YES,HOLD-BACK=18,PART-HOLD-BACK=3\n")
	} else {
		builder.WriteString("#EXT-X-SERVER-CONTROL:HOLD-BACK=18\n")
	}
	fmt.Fprintf(&builder, "#EXT-X-MAP:URI=\"%s\"\n", playlist.MapURI)
	for _, segment := range playlist.Segments {
		if !relativeMediaURI(segment.URI) || math.Abs(segment.Duration-6) > 0.0005 || segment.ProgramDateTime.IsZero() { return nil, ErrInvalidPlaylist }
		if segment.Discontinuity { builder.WriteString("#EXT-X-DISCONTINUITY\n") }
		fmt.Fprintf(&builder, "#EXT-X-PROGRAM-DATE-TIME:%s\n#EXTINF:%.3f,\n%s\n", segment.ProgramDateTime.UTC().Format(time.RFC3339Nano), segment.Duration, segment.URI)
	}
	if playlist.Mode == ModeLLHLS {
		if len(playlist.Segments) != 0 {
			last := playlist.Segments[len(playlist.Segments)-1]
			expected := last.ProgramDateTime.Add(time.Duration(last.Duration * float64(time.Second)))
			if !playlist.PartsProgramDateTime.Equal(expected) { return nil, ErrInvalidPlaylist }
		}
		fmt.Fprintf(&builder, "#EXT-X-PROGRAM-DATE-TIME:%s\n", playlist.PartsProgramDateTime.UTC().Format(time.RFC3339Nano))
		for _, part := range playlist.Parts {
			if !relativeMediaURI(part.URI) || part.Duration < 0.85 || part.Duration > 1 { return nil, ErrInvalidPlaylist }
			fmt.Fprintf(&builder, "#EXT-X-PART:DURATION=%.3f,URI=\"%s\"", part.Duration, part.URI)
			if part.Independent { builder.WriteString(",INDEPENDENT=YES") }
			builder.WriteByte('\n')
		}
		fmt.Fprintf(&builder, "#EXT-X-PRELOAD-HINT:TYPE=PART,URI=\"%s\"\n", playlist.PreloadHint)
		seen := map[string]struct{}{}
		for _, report := range playlist.RenditionReports {
			if !renditionReportURI(report.URI) || report.LastMSN == 0 { return nil, ErrInvalidPlaylist }
			if _, exists := seen[report.URI]; exists { return nil, ErrInvalidPlaylist }
			seen[report.URI] = struct{}{}
			fmt.Fprintf(&builder, "#EXT-X-RENDITION-REPORT:URI=\"%s\",LAST-MSN=%d,LAST-PART=%d\n", report.URI, report.LastMSN, report.LastPart)
		}
	}
	return boundedPlaylist(builder.String())
}

func relativePlaylistURI(value string) bool { return relativeURI(value) && strings.HasSuffix(value, ".m3u8") }
func renditionReportURI(value string) bool {
	if relativePlaylistURI(value) { return true }
	for _, variant := range []string{"audio", "high", "medium", "low"} {
		if value == "../"+variant+"/index.m3u8" { return true }
	}
	return false
}
func relativeInitURI(value string) bool { return relativeURI(value) && strings.HasSuffix(value, ".mp4") }
func relativeMediaURI(value string) bool { return relativeURI(value) && strings.HasSuffix(value, ".m4s") }
func relativeURI(value string) bool {
	if value == "" || strings.HasPrefix(value, "/") || strings.Contains(value, "\\") || strings.Contains(value, "..") { return false }
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "" && parsed.Host == "" && parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == "" && parsed.Path == value
}

func boundedPlaylist(value string) ([]byte, error) {
	if !strings.HasSuffix(value, "\n") || len(value) > MaximumPlaylistBytes { return nil, ErrInvalidPlaylist }
	return []byte(value), nil
}
