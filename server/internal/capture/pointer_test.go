package capture

import (
	"testing"

	"github.com/m1k1o/neko/server/pkg/types"
)

func TestPointerInVideo(t *testing.T) {
	tests := []struct {
		name      string
		pipelines map[string]types.VideoConfig
		ids       []string
		want      bool
	}{
		{"no ids", map[string]types.VideoConfig{"main": {ShowPointer: true}}, nil, false},
		{"default pipeline without pointer", map[string]types.VideoConfig{"main": {ShowPointer: false}}, []string{"main"}, false},
		{"all tiers with pointer", map[string]types.VideoConfig{
			"high": {ShowPointer: true}, "medium": {ShowPointer: true}, "low": {ShowPointer: true},
		}, []string{"high", "medium", "low"}, true},
		{"one tier without pointer", map[string]types.VideoConfig{
			"high": {ShowPointer: true}, "low": {ShowPointer: false},
		}, []string{"high", "low"}, false},
		{"unselectable pipeline is ignored", map[string]types.VideoConfig{
			"main": {ShowPointer: true}, "spare": {ShowPointer: false},
		}, []string{"main"}, true},
		{"missing pipeline", map[string]types.VideoConfig{}, []string{"main"}, false},
		{"gst pipeline with pointer", map[string]types.VideoConfig{
			"main": {GstPipeline: "ximagesrc display-name={display} show-pointer=true ! appsink name=appsink"},
		}, []string{"main"}, true},
		{"gst pipeline without pointer", map[string]types.VideoConfig{
			"main": {GstPipeline: "ximagesrc display-name={display} show-pointer=false ! appsink name=appsink", ShowPointer: true},
		}, []string{"main"}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := pointerInVideo(test.pipelines, test.ids); got != test.want {
				t.Fatalf("pointerInVideo() = %v, want %v", got, test.want)
			}
		})
	}
}
