package config

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestCaptureDerivesIDsWithoutChangingExplicitOrder(t *testing.T) {
	for _, test := range []struct {
		name, pipelines string
		ids, want []string
	}{
		{"nominal ladder", `{"small":{"nominal_bitrate":500000},"large":{"nominal_bitrate":2000000}}`, nil, []string{"large", "small"}},
		{"partial rates require explicit ladder", `{"z":{"nominal_bitrate":2000000},"a":{}}`, nil, []string{"a", "z"}},
		{"bare shorthand", `{"z":"videotestsrc ! vp8enc ! appsink","a":"videotestsrc ! vp8enc ! appsink"}`, nil, []string{"a", "z"}},
		{"explicit order", `{"small":{},"large":{}}`, []string{"small", "large"}, []string{"small", "large"}},
		{"empty uses normal default", `{}`, nil, []string{"main"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)
			var capture Capture
			if err := capture.Init(&cobra.Command{}); err != nil { t.Fatal(err) }
			viper.Set("capture.video.pipelines", test.pipelines)
			if test.ids != nil { viper.Set("capture.video.ids", test.ids) }
			capture.Set()
			if !reflect.DeepEqual(capture.VideoIDs, test.want) { t.Fatalf("IDs = %v, want %v", capture.VideoIDs, test.want) }
			if test.name == "bare shorthand" && capture.VideoPipelines["a"].GstPipeline == "" { t.Fatal("shorthand lost its pipeline") }
		})
	}
}
