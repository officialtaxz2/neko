package mediaws

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type goldenFixture struct {
	Schema  string `json:"schema"`
	Records []struct {
		Name string `json:"name"`
		Hex  string `json:"hex"`
	} `json:"records"`
}

func TestGoldenFixturesParseAndRoundTrip(t *testing.T) {
	raw, err := os.ReadFile("testdata/neko_media_v1_golden.json")
	if err != nil {
		t.Fatal(err)
	}
	fixture := goldenFixture{}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Schema != "neko.media.fixture/1" {
		t.Fatalf("fixture schema = %q", fixture.Schema)
	}
	expected := map[string]struct {
		recordType RecordType
		kind       Kind
	}{
		"format_video":        {RecordFormat, KindVideo},
		"format_audio":        {RecordFormat, KindAudio},
		"unit_video_key":      {RecordUnit, KindVideo},
		"unit_audio":          {RecordUnit, KindAudio},
		"discontinuity_video": {RecordDiscontinuity, KindVideo},
		"end":                 {RecordEnd, KindNone},
	}
	if len(fixture.Records) != len(expected) {
		t.Fatalf("fixture records = %d, want %d", len(fixture.Records), len(expected))
	}
	for _, item := range fixture.Records {
		t.Run(item.Name, func(t *testing.T) {
			want, ok := expected[item.Name]
			if !ok {
				t.Fatalf("unexpected fixture %q", item.Name)
			}
			encoded, err := hex.DecodeString(item.Hex)
			if err != nil {
				t.Fatal(err)
			}
			record, err := ParseRecord(encoded)
			if err != nil {
				t.Fatal(err)
			}
			if record.Type != want.recordType || record.Kind != want.kind {
				t.Fatalf("record type/kind = %d/%d, want %d/%d", record.Type, record.Kind, want.recordType, want.kind)
			}
			roundTrip, err := MarshalRecord(record)
			if err != nil {
				t.Fatal(err)
			}
			if hex.EncodeToString(roundTrip) != item.Hex {
				t.Fatal("round trip differs")
			}
		})
	}
}

func validVideoUnit() Record {
	return Record{
		Type:       RecordUnit,
		Kind:       KindVideo,
		Flags:      FlagKeyframe | FlagPTSValid,
		TrackID:    2,
		Generation: 1,
		PTS:        1,
		Duration:   40_000,
		Payload:    []byte{1},
	}
}

func TestParseRecordRejectsHeaderAndLengthErrors(t *testing.T) {
	encoded, err := MarshalRecord(validVideoUnit())
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]func([]byte) []byte{
		"truncated": func(data []byte) []byte { return data[:HeaderLength-1] },
		"magic": func(data []byte) []byte { data[0] = 'X'; return data },
		"version": func(data []byte) []byte { data[4] = 2; return data },
		"type": func(data []byte) []byte { data[5] = 99; return data },
		"kind": func(data []byte) []byte { data[6] = 99; return data },
		"flags": func(data []byte) []byte { data[7] = 0x80; return data },
		"header_length": func(data []byte) []byte { data[9] = 63; return data },
		"reserved": func(data []byte) []byte { data[11] = 1; return data },
		"declared_length": func(data []byte) []byte { data[15] = 1; return data },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			candidate := append([]byte(nil), encoded...)
			if _, err := ParseRecord(mutate(candidate)); err == nil {
				t.Fatal("invalid record accepted")
			}
		})
	}
}

func TestDecodeStrictJSONRejectsDuplicateUnknownAndTrailingData(t *testing.T) {
	var metadata EndMetadata
	for _, raw := range []string{
		`{"schema":"neko.media.end/1","schema":"neko.media.end/1","reason":"normal"}`,
		`{"schema":"neko.media.end/1","reason":"normal","extra":true}`,
		`{"schema":"neko.media.end/1","reason":"normal"} {}`,
	} {
		if err := DecodeStrictJSON([]byte(raw), &metadata); err == nil {
			t.Fatalf("invalid JSON accepted: %s", raw)
		}
	}
}

func TestRecordValidationRejectsInvalidTimingConfigAndBounds(t *testing.T) {
	tests := map[string]func(*Record){
		"negative_pts": func(record *Record) { record.PTS = -1 },
		"duration": func(record *Record) { record.Duration = MaxUnitDurationMicros + 1 },
		"generation": func(record *Record) { record.Generation = MaxSafeInteger + 1 },
		"track": func(record *Record) { record.TrackID = 1 },
		"audio_keyframe": func(record *Record) {
			record.Kind = KindAudio
			record.TrackID = 1
		},
		"invalid_dts": func(record *Record) { record.DTS = 1 },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			record := validVideoUnit()
			mutate(&record)
			if _, err := MarshalRecord(record); err == nil {
				t.Fatal("invalid record accepted")
			}
		})
	}

	format := Record{
		Type:       RecordFormat,
		Kind:       KindVideo,
		Flags:      FlagConfigPresent,
		TrackID:    2,
		Generation: 1,
		Metadata:   []byte(`{"schema":"neko.media.format/1","source_id":"high","source_generation":1,"codec":"vp8","mime_type":"video/VP8","clock_rate":90000,"channels":0,"coded_width":1280,"coded_height":720,"display_width":1280,"display_height":720,"frame_rate_numerator":25,"frame_rate_denominator":1,"nominal_bitrate":1996800}`),
		Payload:    []byte{1},
	}
	if _, err := MarshalRecord(format); err == nil {
		t.Fatal("version 1 codec config accepted")
	}

	format.Flags = 0
	format.Payload = nil
	format.Metadata = []byte(strings.Replace(string(format.Metadata), `"channels":0,`, "", 1))
	if _, err := MarshalRecord(format); err == nil {
		t.Fatal("format metadata with omitted zero-value field accepted")
	}

	format.Metadata = []byte(`{"schema":"neko.media.format/1","source_id":"high","source_generation":1,"codec":"vp8","mime_type":"video/VP8","clock_rate":90000,"channels":0,"coded_width":1280,"coded_height":720,"display_width":1280,"display_height":720,"frame_rate_numerator":25,"frame_rate_denominator":1,"nominal_bitrate":1996800}`)
	format.Metadata = []byte(strings.Replace(string(format.Metadata), `"source_id":"high"`, `"source_id":"`+strings.Repeat("x", MaxSourceIDLength+1)+`"`, 1))
	if _, err := MarshalRecord(format); err == nil {
		t.Fatal("over-limit source id accepted")
	}
}

func TestEncodedLengthValidationRejectsBeforePayloadAllocation(t *testing.T) {
	if err := validateEncodedLengths(RecordUnit, KindVideo, 0, MaxVideoPayload+1); err == nil {
		t.Fatal("over-limit video payload accepted")
	}
	if err := validateEncodedLengths(RecordUnit, KindAudio, 0, MaxAudioPayload+1); err == nil {
		t.Fatal("over-limit audio payload accepted")
	}
	if err := validateEncodedLengths(RecordEnd, KindNone, MaxMetadata+1, 0); err == nil {
		t.Fatal("over-limit metadata accepted")
	}
}
