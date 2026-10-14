package mediahls

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

// Exercise the packager and rendered wire playlists across multiple window
// advances. The same segment URI must keep its discontinuity number when an
// earlier tag disappears, including the independent alternate audio track.
func TestPackagerRollingPlaylistPreservesDiscontinuityNumbers(t *testing.T) {
	provider := newFakeHLSProvider()
	packager := newPackager(provider, fakeTranscoderFactory{})
	var workers []*packagerWorker
	for _, id := range []string{"high", "audio", "medium", "low"} {
		kind := types.MediaKindVideo
		if id == "audio" {
			kind = types.MediaKindAudio
		}
		source := provider.Sources(kind)[0]
		var variant Variant
		for _, candidate := range FixedVariants() {
			if candidate.ID == id {
				variant = candidate
			}
		}
		worker, err := packager.newRenditionWorker(id, variant, source, 1, 7, 1, time.Unix(0, 0))
		if err != nil {
			t.Fatal(err)
		}
		workers = append(workers, worker)
		packager.tracks[id] = worker.track
	}
	t.Cleanup(func() {
		closePackagerWorkers(workers, false)
		packager.Shutdown()
	})
	previous := make(map[string]map[string]uint64)
	for second := 0; second <= 36; second++ {
		for _, worker := range workers {
			if !worker.transcoder.Push(startupUnit(second)) {
				t.Fatal("fixture transcoder rejected sample")
			}
			if err := packager.acceptSample(worker.track, <-worker.transcoder.Samples()); err != nil {
				t.Fatalf("%s at %ds: %v", worker.track.id, second, err)
			}
		}
		if second < 18 || second%6 != 0 {
			continue
		}
		for _, mode := range []string{ModeHLS, ModeLLHLS} {
			for _, worker := range workers {
				data, err := packager.Playlist(context.Background(), mode, worker.track.id, PlaylistDirectives{})
				if err != nil {
					t.Fatalf("%s/%s at %ds: %v", mode, worker.track.id, second, err)
				}
				numbers := playlistSegmentDiscontinuities(t, string(data))
				if len(numbers) != 3 {
					t.Fatalf("%s/%s: expected three parent segments, got %d", mode, worker.track.id, len(numbers))
				}
				key := mode + "/" + worker.track.id
				overlaps := 0
				for uri, number := range numbers {
					if oldNumber, found := previous[key][uri]; found {
						overlaps++
						if oldNumber != number {
							t.Fatalf("playlist discontinuity changed for retained segment: %s %s at %ds: %d -> %d", key, uri, second, oldNumber, number)
						}
					}
				}
				if second > 18 && overlaps != 2 {
					t.Fatalf("%s: expected two retained parent segments, got %d", key, overlaps)
				}
				for uri, number := range numbers {
					if number != 8 {
						t.Fatalf("%s %s at %ds: same-generation discontinuity = %d, want 8", key, uri, second, number)
					}
				}
				previous[key] = numbers
			}
		}
	}
}

func playlistSegmentDiscontinuities(t *testing.T, playlist string) map[string]uint64 {
	t.Helper()
	numbers := make(map[string]uint64)
	var number uint64
	for _, line := range strings.Split(playlist, "\n") {
		switch {
		case strings.HasPrefix(line, "#EXT-X-DISCONTINUITY-SEQUENCE:"):
			parsed, err := strconv.ParseUint(strings.TrimPrefix(line, "#EXT-X-DISCONTINUITY-SEQUENCE:"), 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			number = parsed
		case line == "#EXT-X-DISCONTINUITY":
			number++
		case strings.HasPrefix(line, "seg-") && strings.HasSuffix(line, ".m4s"):
			numbers[line] = number
		}
	}
	return numbers
}

// LL-HLS keeps listing the parts of the most recent completed parents, so a
// player at PART-HOLD-BACK can cross a parent rollover part by part. Every
// listed part must still be retained; conventional playlists list none.
func TestPackagerLowLatencyListsRecentParentParts(t *testing.T) {
	provider := newFakeHLSProvider()
	packager := newPackager(provider, fakeTranscoderFactory{})
	var workers []*packagerWorker
	for _, id := range []string{"high", "audio", "medium", "low"} {
		kind := types.MediaKindVideo
		if id == "audio" {
			kind = types.MediaKindAudio
		}
		source := provider.Sources(kind)[0]
		var variant Variant
		for _, candidate := range FixedVariants() {
			if candidate.ID == id {
				variant = candidate
			}
		}
		worker, err := packager.newRenditionWorker(id, variant, source, 1, 7, 1, time.Unix(0, 0))
		if err != nil {
			t.Fatal(err)
		}
		workers = append(workers, worker)
		packager.tracks[id] = worker.track
	}
	t.Cleanup(func() {
		closePackagerWorkers(workers, false)
		packager.Shutdown()
	})
	for second := 0; second <= 31; second++ {
		for _, worker := range workers {
			if !worker.transcoder.Push(startupUnit(second)) {
				t.Fatal("fixture transcoder rejected sample")
			}
			if err := packager.acceptSample(worker.track, <-worker.transcoder.Samples()); err != nil {
				t.Fatalf("%s at %ds: %v", worker.track.id, second, err)
			}
		}
		if second == 4 {
			// Before the first parent completes, the media sequence names the
			// parent of the listed parts.
			data, err := packager.Playlist(context.Background(), ModeLLHLS, "high", PlaylistDirectives{})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), "#EXT-X-MEDIA-SEQUENCE:1\n") || !strings.Contains(string(data), `URI="part-1-0.m4s"`) {
				t.Fatalf("startup LL-HLS playlist:\n%s", data)
			}
		}
	}
	// Parents 1–5 are complete (3–5 visible) and part 6-0 is current.
	for _, worker := range workers {
		id := worker.track.id
		conventional, err := packager.Playlist(context.Background(), ModeHLS, id, PlaylistDirectives{})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(conventional), "#EXT-X-PART") {
			t.Fatalf("%s: conventional playlist lists parts:\n%s", id, conventional)
		}
		data, err := packager.Playlist(context.Background(), ModeLLHLS, id, PlaylistDirectives{})
		if err != nil {
			t.Fatal(err)
		}
		var listed []string
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "#EXT-X-PART:") {
				uri := line[strings.Index(line, `URI="`)+5:]
				listed = append(listed, uri[:strings.Index(uri, `"`)])
			}
		}
		if len(listed) != 13 || listed[0] != "part-4-0.m4s" || listed[12] != "part-6-0.m4s" {
			t.Fatalf("%s: listed parts %v\n%s", id, listed, data)
		}
		if strings.Contains(string(data), "part-3-") {
			t.Fatalf("%s: oldest visible parent still lists parts:\n%s", id, data)
		}
		for _, uri := range listed {
			if _, found := packager.Object(id, uri); !found {
				t.Fatalf("%s: listed part %s is not retained", id, uri)
			}
		}
	}
}
