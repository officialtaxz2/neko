package mediahls

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
	"time"
)

func TestInitSegmentsCarryExactCodecsAndFragmentedMovieMetadata(t *testing.T) {
	avcC := []byte{1, 100, 0, 31, 0xff, 0xe1, 0, 1, 0x67, 1, 0, 1, 0x68}
	video, err := buildVideoInit(1280, 720, avcC)
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range [][]byte{[]byte("ftyp"), []byte("moov"), []byte("avc1"), []byte("avcC"), avcC, []byte("mvex"), []byte("trex")} {
		if !bytes.Contains(video, marker) {
			t.Fatalf("video init lacks %q", marker)
		}
	}
	audio, err := buildAudioInit([]byte{0x11, 0x90})
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range [][]byte{[]byte("mp4a"), []byte("esds"), []byte{0x05, 0x02, 0x11, 0x90}} {
		if !bytes.Contains(audio, marker) {
			t.Fatalf("audio init lacks %x", marker)
		}
	}
	if _, err := buildVideoInit(1280, 720, []byte{1, 77, 0, 31}); !errors.Is(err, ErrInvalidCodecConfig) {
		t.Fatalf("non-High profile accepted: %v", err)
	}
	if _, err := buildVideoInit(1280, 720, []byte{1, 100, 0, 31}); !errors.Is(err, ErrInvalidCodecConfig) {
		t.Fatalf("incomplete AVC configuration accepted: %v", err)
	}
}

func TestFragmentMuxerProducesMonotonicMoofMdatWithBoundedSamples(t *testing.T) {
	muxer := newFragmentMuxer(90_000)
	fragment, err := muxer.fragment([]fragmentSample{
		{data: []byte{0, 0, 0, 1, 0x65}, dts: 0, pts: 0, duration: 3_600, keyframe: true},
		{data: []byte{0, 0, 0, 1, 0x41}, dts: 3_600, pts: 3_600, duration: 3_600},
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(fragment[4:8]) != "moof" || !bytes.Contains(fragment, []byte("tfdt")) || !bytes.Contains(fragment, []byte("trun")) || !bytes.Contains(fragment, []byte("mdat")) {
		t.Fatalf("invalid fragment boxes: %x", fragment)
	}
	moofSize := int(binary.BigEndian.Uint32(fragment[:4]))
	if moofSize < 8 || moofSize+8 >= len(fragment) || string(fragment[moofSize+4:moofSize+8]) != "mdat" {
		t.Fatalf("invalid moof/mdat boundary: moof=%d total=%d", moofSize, len(fragment))
	}
	if muxer.sequence != 1 {
		t.Fatalf("sequence = %d", muxer.sequence)
	}
	if _, err := muxer.fragment([]fragmentSample{{data: []byte{1}, dts: 10, pts: 9, duration: 1}}); !errors.Is(err, ErrInvalidObject) {
		t.Fatalf("negative composition offset accepted: %v", err)
	}
}

func TestDurationTicksRoundsCodecFrameDurations(t *testing.T) {
	if got := durationTicks(1024*time.Second/48_000, 48_000); got != 1024 {
		t.Fatalf("AAC duration ticks = %d", got)
	}
	if got := durationTicks(time.Second/30, 90_000); got != 3000 {
		t.Fatalf("30-fps duration ticks = %d", got)
	}
}
