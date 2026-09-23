package mediahls

import (
	"fmt"
	"sync"
	"time"

	"github.com/m1k1o/neko/server/pkg/gst"
	"github.com/m1k1o/neko/server/pkg/types"
)

const (
	ProviderQueueCapacity = 64
	WorkerHandoffCapacity = 8
	AudioBitrate          = 128_000
)

type transcoder interface {
	Samples() <-chan types.Sample
	Drops() <-chan struct{}
	Push(types.EncodedMediaUnit) bool
	CapturedAt(types.Sample) time.Time
	Close()
}

type transcoderFactory interface {
	NewAudio(types.MediaSource) (transcoder, error)
	NewVideo(Variant, types.MediaSource) (transcoder, error)
}

type gstTranscoderFactory struct{}

func validateGSTTranscoderElements() error {
	for _, element := range []string{
		"appsrc", "appsink", "queue",
		"opusparse", "opusdec", "audioconvert", "audioresample", "voaacenc", "aacparse",
		"vp8dec", "videoconvert", "videoscale", "videorate", "x264enc", "h264parse",
	} {
		if err := gst.CheckElement(element); err != nil {
			return fmt.Errorf("HLS transcoder element %s: %w", element, err)
		}
	}
	return nil
}

func (gstTranscoderFactory) NewAudio(source types.MediaSource) (transcoder, error) {
	if source.Kind != types.MediaKindAudio || source.Codec.Name != "opus" || source.Codec.ClockRate != 48_000 || source.Codec.Channels != 2 {
		return nil, ErrCodecUnsupported
	}
	pipeline := fmt.Sprintf(
		"appsrc name=appsrc is-live=true format=time block=false max-buffers=%d max-bytes=0 max-time=0 caps=audio/x-opus,rate=48000,channels=2 ! queue max-size-buffers=%d max-size-bytes=0 max-size-time=0 leaky=downstream ! opusparse ! opusdec ! audioconvert ! audioresample ! audio/x-raw,rate=48000,channels=2 ! voaacenc bitrate=%d ! aacparse ! audio/mpeg,mpegversion=4,stream-format=raw ! appsink name=appsink max-buffers=%d drop=true sync=false",
		WorkerHandoffCapacity,
		WorkerHandoffCapacity,
		AudioBitrate,
		WorkerHandoffCapacity,
	)
	return newGSTTranscoder(pipeline)
}

func (gstTranscoderFactory) NewVideo(variant Variant, source types.MediaSource) (transcoder, error) {
	if source.Kind != types.MediaKindVideo || source.ID != variant.SourceID || source.Codec.Name != "vp8" {
		return nil, ErrCodecUnsupported
	}
	bitrate := (variant.AverageBandwidth - AudioBitrate) / 1_000
	peakBitrate := (variant.Bandwidth - AudioBitrate) / 1_000
	keyframeInterval := variant.FrameRate * 2
	pipeline := fmt.Sprintf(
		"appsrc name=appsrc is-live=true format=time block=false max-buffers=%d max-bytes=0 max-time=0 caps=video/x-vp8 ! queue max-size-buffers=%d max-size-bytes=0 max-size-time=0 leaky=downstream ! vp8dec ! videoconvert ! videoscale ! videorate ! video/x-raw,format=I420,width=%d,height=%d,framerate=%d/1 ! x264enc bitrate=%d key-int-max=%d bframes=0 byte-stream=false aud=true speed-preset=veryfast tune=zerolatency option-string=vbv-maxrate=%d:vbv-bufsize=%d ! h264parse config-interval=-1 ! video/x-h264,stream-format=avc,alignment=au,profile=high,level=(string)3.1 ! appsink name=appsink max-buffers=%d drop=true sync=false",
		WorkerHandoffCapacity,
		WorkerHandoffCapacity,
		variant.Width,
		variant.Height,
		variant.FrameRate,
		bitrate,
		keyframeInterval,
		peakBitrate,
		peakBitrate*2,
		WorkerHandoffCapacity,
	)
	return newGSTTranscoder(pipeline)
}

type gstTranscoder struct {
	pipeline     gst.Pipeline
	timelineMu   sync.Mutex
	timelineSet  bool
	capturedBase time.Time
}

func newGSTTranscoder(pipelineSource string) (*gstTranscoder, error) {
	pipeline, err := gst.CreatePipelineWithSampleCapacity(pipelineSource, WorkerHandoffCapacity)
	if err != nil {
		return nil, err
	}
	pipeline.AttachAppsrc("appsrc")
	pipeline.AttachAppsink("appsink")
	pipeline.Play()
	return &gstTranscoder{pipeline: pipeline}, nil
}

func (transcoder *gstTranscoder) Samples() <-chan types.Sample {
	return transcoder.pipeline.Sample()
}

func (transcoder *gstTranscoder) Drops() <-chan struct{} {
	return transcoder.pipeline.Dropped()
}

func (transcoder *gstTranscoder) Push(unit types.EncodedMediaUnit) bool {
	if unit.PTS < 0 {
		return false
	}
	if !unit.CapturedAt.IsZero() {
		transcoder.timelineMu.Lock()
		if !transcoder.timelineSet {
			transcoder.timelineSet = true
			transcoder.capturedBase = unit.CapturedAt.Add(-unit.PTS)
		}
		transcoder.timelineMu.Unlock()
	}
	dts := unit.DTS
	if !unit.DTSValid || dts < 0 {
		dts = unit.PTS
	}
	return transcoder.pipeline.PushSample(types.Sample{
		Data:      unit.Data,
		Length:    len(unit.Data),
		Timestamp: unit.CapturedAt,
		PTS:       unit.PTS,
		DTS:       dts,
		// The provider contract makes normalized PTS usable even when
		// PTSValid is false (that bit only records that it was synthesized).
		PTSValid:  true,
		DTSValid:  true,
		Duration:  unit.Duration,
		DeltaUnit: !unit.Keyframe,
	})
}

func (transcoder *gstTranscoder) CapturedAt(sample types.Sample) time.Time {
	transcoder.timelineMu.Lock()
	defer transcoder.timelineMu.Unlock()
	if !transcoder.timelineSet || !sample.PTSValid {
		return sample.Timestamp
	}
	return transcoder.capturedBase.Add(sample.PTS)
}

func (transcoder *gstTranscoder) Close() {
	transcoder.pipeline.Destroy()
}
