package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0 gstreamer-video-1.0

#include "gst.h"
*/
import "C"
import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/m1k1o/neko/server/pkg/types"
)

var (
	pSerial       int32
	pipelines     = make(map[int]*pipeline)
	pipelinesLock sync.Mutex
	registry      *C.GstRegistry
)

func init() {
	C.gst_init(nil, nil)
	registry = C.gst_registry_get()
}

type Pipeline interface {
	Src() string
	Sample() chan types.Sample
	Dropped() <-chan struct{}
	// attach sink or src to pipeline
	AttachAppsink(sinkName string)
	AttachAppsrc(srcName string)
	// control pipeline lifecycle
	Play()
	Pause()
	Destroy()
	Push(buffer []byte)
	// PushSample preserves the encoded provider timeline and keyframe flag for
	// transcode/remux pipelines. Push remains the compatibility helper for
	// callers without timestamp metadata.
	PushSample(sample types.Sample) bool
	// modify the property of a bin
	SetPropInt(binName string, prop string, value int) bool
	SetCapsFramerate(binName string, numerator, denominator int) bool
	SetCapsResolution(binName string, width, height int) bool
	// emit video keyframe
	EmitVideoKeyframe() bool
}

type pipeline struct {
	id      int
	logger  zerolog.Logger
	src     string
	ctx     *C.GstPipelineCtx
	sample  chan types.Sample
	dropped chan struct{}
	done    chan struct{}
}

func CreatePipeline(pipelineStr string) (Pipeline, error) {
	return CreatePipelineWithSampleCapacity(pipelineStr, 4)
}

func CreatePipelineWithSampleCapacity(pipelineStr string, sampleCapacity int) (Pipeline, error) {
	if sampleCapacity < 1 {
		return nil, fmt.Errorf("sample capacity must be positive")
	}
	id := atomic.AddInt32(&pSerial, 1)

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	var gstError *C.GError
	ctx := C.gstreamer_pipeline_create(pipelineStrUnsafe, C.int(id), &gstError)

	if gstError != nil {
		defer C.g_error_free(gstError)
		return nil, fmt.Errorf("(pipeline error) %s", C.GoString(gstError.message))
	}

	p := &pipeline{
		id:     int(id),
		logger: log.With().
			Str("module", "capture").
			Str("submodule", "gstreamer").
			Int("pipeline_id", int(id)).Logger(),
		src:     pipelineStr,
		ctx:     ctx,
		sample:  make(chan types.Sample, sampleCapacity),
		dropped: make(chan struct{}, 1),
		done:    make(chan struct{}),
	}

	pipelines[p.id] = p
	return p, nil
}

func (p *pipeline) Src() string {
	return p.src
}

func (p *pipeline) Sample() chan types.Sample {
	return p.sample
}

func (p *pipeline) Dropped() <-chan struct{} {
	return p.dropped
}

func (p *pipeline) AttachAppsink(sinkName string) {
	sinkNameUnsafe := C.CString(sinkName)
	defer C.free(unsafe.Pointer(sinkNameUnsafe))

	C.gstreamer_pipeline_attach_appsink(p.ctx, sinkNameUnsafe)
}

func (p *pipeline) AttachAppsrc(srcName string) {
	srcNameUnsafe := C.CString(srcName)
	defer C.free(unsafe.Pointer(srcNameUnsafe))

	C.gstreamer_pipeline_attach_appsrc(p.ctx, srcNameUnsafe)
}

func (p *pipeline) Play() {
	C.gstreamer_pipeline_play(p.ctx)
}

func (p *pipeline) Pause() {
	C.gstreamer_pipeline_pause(p.ctx)
}

func (p *pipeline) Destroy() {
	pipelinesLock.Lock()
	delete(pipelines, p.id)
	close(p.done)
	pipelinesLock.Unlock()

	C.gstreamer_pipeline_destory(p.ctx)
	close(p.sample)
	close(p.dropped)
	C.free(unsafe.Pointer(p.ctx))
}

func (p *pipeline) Push(buffer []byte) {
	bytes := C.CBytes(buffer)
	defer C.free(bytes)

	C.gstreamer_pipeline_push(p.ctx, bytes, C.int(len(buffer)))
}

func (p *pipeline) PushSample(sample types.Sample) bool {
	if len(sample.Data) == 0 {
		return false
	}
	bytes := C.CBytes(sample.Data)
	defer C.free(bytes)

	pts := C.guint64(0)
	ptsValid := sample.PTSValid && sample.PTS >= 0
	if ptsValid {
		pts = C.guint64(sample.PTS)
	}
	dts := C.guint64(0)
	dtsValid := sample.DTSValid && sample.DTS >= 0
	if dtsValid {
		dts = C.guint64(sample.DTS)
	}
	duration := C.guint64(0)
	if sample.Duration > 0 {
		duration = C.guint64(sample.Duration)
	}

	return C.gstreamer_pipeline_push_sample(
		p.ctx,
		bytes,
		C.int(len(sample.Data)),
		pts,
		C.gboolean(boolToGBoolean(ptsValid)),
		dts,
		C.gboolean(boolToGBoolean(dtsValid)),
		duration,
		C.gboolean(boolToGBoolean(sample.DeltaUnit)),
	) == C.TRUE
}

func boolToGBoolean(value bool) C.int {
	if value {
		return C.TRUE
	}
	return C.FALSE
}

func (p *pipeline) SetPropInt(binName string, prop string, value int) bool {
	cBinName := C.CString(binName)
	defer C.free(unsafe.Pointer(cBinName))

	cProp := C.CString(prop)
	defer C.free(unsafe.Pointer(cProp))

	cValue := C.int(value)

	p.logger.Debug().Msgf("setting prop %s of %s to %d", prop, binName, value)

	ok := C.gstreamer_pipeline_set_prop_int(p.ctx, cBinName, cProp, cValue)
	return ok == C.TRUE
}

func (p *pipeline) SetCapsFramerate(binName string, numerator, denominator int) bool {
	cBinName := C.CString(binName)
	cNumerator := C.int(numerator)
	cDenominator := C.int(denominator)

	defer C.free(unsafe.Pointer(cBinName))

	p.logger.Debug().Msgf("setting caps framerate of %s to %d/%d", binName, numerator, denominator)

	ok := C.gstreamer_pipeline_set_caps_framerate(p.ctx, cBinName, cNumerator, cDenominator)
	return ok == C.TRUE
}

func (p *pipeline) SetCapsResolution(binName string, width, height int) bool {
	cBinName := C.CString(binName)
	cWidth := C.int(width)
	cHeight := C.int(height)

	defer C.free(unsafe.Pointer(cBinName))

	p.logger.Debug().Msgf("setting caps resolution of %s to %dx%d", binName, width, height)

	ok := C.gstreamer_pipeline_set_caps_resolution(p.ctx, cBinName, cWidth, cHeight)
	return ok == C.TRUE
}

func (p *pipeline) EmitVideoKeyframe() bool {
	ok := C.gstreamer_pipeline_emit_video_keyframe(p.ctx)
	return ok == C.TRUE
}

// gst-inspect-1.0
func CheckPlugins(plugins []string) error {
	var plugin *C.GstPlugin
	for _, pluginstr := range plugins {
		plugincstr := C.CString(pluginstr)
		plugin = C.gst_registry_find_plugin(registry, plugincstr)
		C.free(unsafe.Pointer(plugincstr))
		if plugin == nil {
			return fmt.Errorf("required gstreamer plugin %s not found", pluginstr)
		}
	}

	return nil
}

func CheckElement(element string) error {
	elementcstr := C.CString(element)
	factory := C.gst_element_factory_find(elementcstr)
	C.free(unsafe.Pointer(elementcstr))
	if factory == nil {
		return fmt.Errorf("required gstreamer element %s not found", element)
	}
	C.gst_object_unref(C.gpointer(factory))
	return nil
}

//export goHandlePipelineBuffer
func goHandlePipelineBuffer(
	pipelineID C.int,
	buf C.gpointer,
	bufLen C.int,
	pts C.guint64,
	ptsValid C.gboolean,
	dts C.guint64,
	dtsValid C.gboolean,
	duration C.guint64,
	deltaUnit C.gboolean,
	width C.gint,
	height C.gint,
	frameRateNumerator C.gint,
	frameRateDenominator C.gint,
	codecConfig C.gpointer,
	codecConfigLen C.int,
) {
	defer C.g_free(buf)
	if codecConfig != nil {
		defer C.g_free(codecConfig)
	}

	pipelinesLock.Lock()
	pipeline, ok := pipelines[int(pipelineID)]
	pipelinesLock.Unlock()

	if ok {
		sample := types.Sample{
			Data:                 C.GoBytes(unsafe.Pointer(buf), bufLen),
			Length:               int(bufLen),
			Timestamp:            time.Now(),
			PTS:                  time.Duration(pts),
			DTS:                  time.Duration(dts),
			PTSValid:             ptsValid == C.TRUE,
			DTSValid:             dtsValid == C.TRUE,
			Duration:             time.Duration(duration),
			DeltaUnit:            deltaUnit == C.TRUE,
			Width:                uint32(width),
			Height:               uint32(height),
			FrameRateNumerator:   uint32(frameRateNumerator),
			FrameRateDenominator: uint32(frameRateDenominator),
			CodecConfig:          C.GoBytes(unsafe.Pointer(codecConfig), codecConfigLen),
		}
		select {
		case <-pipeline.done:
			return
		default:
		}
		select {
		case pipeline.sample <- sample:
		default:
			select {
			case pipeline.dropped <- struct{}{}:
			default:
			}
		}
	} else {
		log.Warn().
			Str("module", "capture").
			Str("submodule", "gstreamer").
			Int("pipeline_id", int(pipelineID)).
			Msgf("discarding sample, pipeline not found")
	}
}

//export goPipelineLog
func goPipelineLog(pipelineID C.int, levelUnsafe *C.char, msgUnsafe *C.char) {
	levelStr := C.GoString(levelUnsafe)
	msg := C.GoString(msgUnsafe)

	level, _ := zerolog.ParseLevel(levelStr)
	log.WithLevel(level).
		Str("module", "capture").
		Str("submodule", "gstreamer").
		Int("pipeline_id", int(pipelineID)).
		Msg(msg)
}
