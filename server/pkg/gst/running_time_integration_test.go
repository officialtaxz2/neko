//go:build hlsintegration

package gst

import (
	"testing"
	"time"

	"github.com/m1k1o/neko/server/pkg/types"
)

func TestEncoderSegmentMapsToRunningTimeWithoutChangingCapture(t *testing.T) {
	if err := CheckElement("x264enc"); err != nil {
		t.Fatal(err) // the explicit integration gate must not skip missing codecs
	}
	source := "videotestsrc num-buffers=4 ! video/x-raw,format=I420,width=64,height=64,framerate=25/1 ! x264enc tune=zerolatency bframes=0 byte-stream=false ! video/x-h264,stream-format=avc,alignment=au ! appsink name=appsink sync=false"
	first := func(running bool) types.Sample {
		t.Helper()
		var pipeline Pipeline
		var err error
		if running {
			pipeline, err = CreatePipelineWithRunningTimeSamples(source, 8)
		} else {
			pipeline, err = CreatePipelineWithSampleCapacity(source, 8)
		}
		if err != nil {
			t.Fatal(err)
		}
		defer pipeline.Destroy()
		pipeline.AttachAppsink("appsink")
		pipeline.Play()
		select {
		case sample := <-pipeline.Sample():
			if !sample.PTSValid || !sample.DTSValid || len(sample.CodecConfig) == 0 {
				t.Fatal("encoder sample lacks timestamps or AVC configuration")
			}
			return sample
		case <-time.After(5 * time.Second):
			t.Fatal("encoder produced no sample")
			return types.Sample{}
		}
	}
	raw := first(false)
	normalized := first(true)
	if raw.PTS < time.Hour || raw.DTS < time.Hour || normalized.PTS != 0 || normalized.DTS != 0 {
		t.Fatalf("unexpected raw/normalized encoder clock: raw=%s normalized PTS=%s DTS=%s", raw.PTS, normalized.PTS, normalized.DTS)
	}
}
