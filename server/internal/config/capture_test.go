package config

import (
	"reflect"
	"strings"
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

// loadCaptureFromEnv mirrors the production config flow: env prefix setup,
// flag registration via Init, then value population via Set.
func loadCaptureFromEnv(t *testing.T, pipelines string) *Capture {
	t.Helper()

	viper.Reset()
	viper.SetEnvPrefix("NEKO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	t.Setenv("NEKO_CAPTURE_VIDEO_PIPELINES", pipelines)

	c := &Capture{}
	cmd := &cobra.Command{}
	if err := c.Init(cmd); err != nil {
		t.Fatal(err)
	}
	c.Set()
	return c
}

func TestVideoPipelinesEnv(t *testing.T) {
	const pipeline = "ximagesrc display-name={display} show-pointer=false ! appsink name=appsink"

	t.Run("object entries with gst_pipeline keep working", func(t *testing.T) {
		c := loadCaptureFromEnv(t, `{"high":{"gst_pipeline":"`+pipeline+`"}}`)

		cfg, ok := c.VideoPipelines["high"]
		if !ok {
			t.Fatalf("pipeline entry missing, got %v", c.VideoPipelines)
		}
		if cfg.GstPipeline != pipeline {
			t.Errorf("expected pipeline %q, got %q", pipeline, cfg.GstPipeline)
		}
	})

	t.Run("bare pipeline string entries are accepted", func(t *testing.T) {
		c := loadCaptureFromEnv(t, `{"high":"`+pipeline+`"}`)

		cfg, ok := c.VideoPipelines["high"]
		if !ok {
			t.Fatalf("bare pipeline string entry was dropped and neko fell back to the default vp8 pipeline, got %v", c.VideoPipelines)
		}
		if cfg.GstPipeline != pipeline {
			t.Errorf("expected pipeline %q, got %q", pipeline, cfg.GstPipeline)
		}
	})
}
