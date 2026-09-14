package mediaws

import "testing"

func TestParseReadyControlStrictSchemas(t *testing.T) {
	control, err := parseControl([]byte(`{"type":"ready","version":1,"audio":{"generation":"1","codec":"opus","sample_rate":48000,"channels":2},"video":{"generation":"2","codec":"vp8","coded_width":1920,"coded_height":1080}}`))
	if err != nil {
		t.Fatal(err)
	}
	if control.Type != "ready" || control.Ready.Audio.Generation != 1 || control.Ready.Video.Generation != 2 {
		t.Fatalf("ready control = %#v", control)
	}

	invalid := []string{
		`{"type":"ready","version":1,"video":{"generation":"1","codec":"vp8","coded_width":1920,"coded_height":1080}}`,
		`{"type":"ready","version":1,"audio":null,"video":{"generation":"01","codec":"vp8","coded_width":1920,"coded_height":1080}}`,
		`{"type":"ready","version":1,"audio":{"generation":"1","codec":"opus","sample_rate":48000,"channels":2,"coded_width":0},"video":{"generation":"1","codec":"vp8","coded_width":1920,"coded_height":1080}}`,
		`{"type":"ready","version":1,"audio":null,"video":{"generation":"1","codec":"vp8","coded_width":0,"coded_height":1080}}`,
		`{"type":"ready","version":1,"audio":null,"video":null,"extra":true}`,
		`{"type":"ready","type":"stop","version":1,"audio":null,"video":null}`,
	}
	for _, raw := range invalid {
		if _, err := parseControl([]byte(raw)); err == nil {
			t.Fatalf("invalid ready accepted: %s", raw)
		}
	}
}

func TestParseFeedbackAndResyncBounds(t *testing.T) {
	feedback := `{"type":"feedback","audio":null,"video":{"generation":"1","received":"10","decoded":"9","rendered":"8","compressed_queue":4,"decode_queue":4,"buffered_ms":200,"drops":"1"},"av_skew_ms":-200}`
	control, err := parseControl([]byte(feedback))
	if err != nil {
		t.Fatal(err)
	}
	if control.Feedback.Video.Rendered != 8 || control.Feedback.AVSkewMS != -200 {
		t.Fatalf("feedback control = %#v", control)
	}

	invalid := []string{
		`{"type":"feedback","audio":null,"video":{"generation":"1","received":"8","decoded":"9","rendered":"8","compressed_queue":0,"decode_queue":0,"buffered_ms":0,"drops":"0"},"av_skew_ms":0}`,
		`{"type":"feedback","audio":null,"video":{"generation":"1","received":"10","decoded":"9","rendered":"8","compressed_queue":5,"decode_queue":0,"buffered_ms":0,"drops":"0"},"av_skew_ms":0}`,
		`{"type":"feedback","audio":null,"video":{"generation":"1","received":"10","decoded":"9","rendered":"8","compressed_queue":0,"decode_queue":0,"buffered_ms":201,"drops":"0"},"av_skew_ms":0}`,
		`{"type":"feedback","audio":null,"video":{"generation":"1","received":"10","decoded":"9","rendered":"8","compressed_queue":0,"decode_queue":0,"buffered_ms":0,"drops":"0"},"av_skew_ms":10001}`,
		`{"type":"resync","kind":"all","generation":"0","reason":"decoder_error"}`,
		`{"type":"resync","kind":"all","generation":"1","reason":"other"}`,
		`{"type":"stop","reason":"extra"}`,
		`{"type":"unknown"}`,
	}
	for _, raw := range invalid {
		if _, err := parseControl([]byte(raw)); err == nil {
			t.Fatalf("invalid control accepted: %s", raw)
		}
	}

	resync, err := parseControl([]byte(`{"type":"resync","kind":"video","generation":"3","reason":"decoder_error"}`))
	if err != nil || resync.Resync.Generation != 3 || resync.Resync.Kind != "video" {
		t.Fatalf("resync control = %#v, err = %v", resync, err)
	}
	for _, reason := range []string{"queue_overflow", "video_compressed_overflow", "audio_compressed_overflow", "audio_output_overflow", "audio_worklet_overflow"} {
		if _, err := parseControl([]byte(`{"type":"resync","kind":"all","generation":"3","reason":"` + reason + `"}`)); err != nil {
			t.Fatalf("bounded overflow reason %q rejected: %v", reason, err)
		}
	}
}

func TestClientOverflowResyncReasonsStayBounded(t *testing.T) {
	for _, reason := range []string{"queue_overflow", "video_compressed_overflow", "audio_compressed_overflow", "audio_output_overflow", "audio_worklet_overflow"} {
		internal := "client_" + reason
		if got := metricResyncReason(internal); got != internal {
			t.Fatalf("metric reason %q = %q", internal, got)
		}
		if got := normalizeResyncProtocolReason(internal); got != "browser_resync" {
			t.Fatalf("protocol reason %q = %q", internal, got)
		}
	}
}

func TestParseControlLengthAndSafeIntegers(t *testing.T) {
	oversized := make([]byte, MaxControlRecord+1)
	if _, err := parseControl(oversized); err == nil {
		t.Fatal("oversized control record accepted")
	}
	for _, value := range []string{"", "00", "+1", "-1", "9007199254740992"} {
		if _, err := parseCanonicalUint(value, true); err == nil {
			t.Fatalf("invalid canonical integer accepted: %q", value)
		}
	}
	if value, err := parseCanonicalUint("9007199254740991", false); err != nil || value != MaxSafeInteger {
		t.Fatalf("maximum safe integer = %d, err = %v", value, err)
	}
}
