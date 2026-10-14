package webrtc

import (
	"sync"
	"testing"
	"time"

	"github.com/pion/interceptor/pkg/cc"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

type raceTestSession struct {
	types.Session
	id string
}

func (s raceTestSession) ID() string { return s.id }

// Metrics register globally, so repeated runs (-count) share one manager.
var raceTestMetrics = newMetricsManager()

type constantEstimator struct {
	cc.BandwidthEstimator
}

func (constantEstimator) GetTargetBitrate() int { return 1_000_000 }

// The estimator goroutine reads the peer's pause/video/audio state while
// signaling handlers change it under peer.mu. Run with -race.
func TestEstimatorReaderStateAccessIsSynchronized(t *testing.T) {
	connection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatal(err)
	}

	peer := &WebRTCPeerCtx{
		logger:     zerolog.Nop(),
		metrics:    raceTestMetrics.getBySession(raceTestSession{id: "estimator-race-test"}),
		connection: connection,
		estimator:  constantEstimator{},
		estimateTrend: utils.NewTrendDetector(utils.TrendDetectorParams{
			RequiredSamples:        8,
			DownwardTrendThreshold: -0.5,
			CollapseValues:         true,
		}),
		audioTrack:      &Track{logger: zerolog.Nop()},
		videoTrack:      &Track{logger: zerolog.Nop()},
		estimatorConfig: config.WebRTCEstimator{Enabled: true, ReadInterval: time.Millisecond},
		videoAuto:       true,
	}

	done := make(chan struct{})
	go func() {
		peer.estimatorReader()
		close(done)
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			_ = peer.SetPaused(i%2 == 0)
			disabled := i%3 == 0
			peer.mu.Lock()
			peer.videoDisabled = disabled
			peer.audioDisabled = !disabled
			peer.videoAuto = i%5 != 0
			peer.mu.Unlock()
			time.Sleep(100 * time.Microsecond)
		}
	}()
	wg.Wait()

	if err := connection.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("estimator reader did not stop after the connection closed")
	}
}
