//go:build hlsintegration && hlsbrowser

package mediahls

// Plays the real packager output through the production MSE player
// (client/src/neko/hls/player.ts with hls.js) in WebKitGTK. The browser starts
// before the packager is ready, as a viewer does after the bootstrap, so a
// playlist hls.js cannot play fails here with the same "HLS player stopped"
// terminal error the UI shows.
//
// Requires node (client dependencies installed), xvfb-run and python3 with
// WebKit2 4.1 GObject bindings (Ubuntu: gir1.2-webkit2-4.1, python3-gi):
//
//	go test -tags 'hlsintegration hlsbrowser' -run TestBrowserPlayback ./internal/mediahls/

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
)

const browserPlaybackSeconds = 45

func TestBrowserPlayback(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	client := filepath.Join(filepath.Dir(file), "..", "..", "..", "client")
	python := os.Getenv("HLS_BROWSER_PYTHON")
	if python == "" {
		python = "python3"
	}
	for _, tool := range []string{"node", "xvfb-run", python} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not available", tool)
		}
	}
	if err := exec.Command(python, "-c", "import gi; gi.require_version('WebKit2', '4.1')").Run(); err != nil {
		t.Skip("WebKit2 4.1 GObject bindings not available")
	}
	static := t.TempDir()
	build := exec.Command("node", filepath.Join("tests", "browser", "build.mjs"), static)
	build.Dir = client
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("player bundle: %v\n%s", err, out)
	}
	for _, mode := range []string{ModeLLHLS, ModeHLS} {
		t.Run(mode, func(t *testing.T) {
			browserPlayback(t, filepath.Join(client, "tests", "browser", "webkit.py"), python, static, mode)
		})
	}
}

func browserPlayback(t *testing.T, script, python, static, mode string) {
	provider := &codecFixtureProvider{origin: time.Now(), errors: make(chan error, 4)}
	packager := newPackager(provider, gstTranscoderFactory{})
	defer packager.Shutdown()
	leases, err := NewLeaseStore(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	backend := &Backend{packager: packager, leases: leases, deliveries: map[string]*Delivery{}}
	var (
		mu         sync.Mutex
		logs       []string
		controller *Controller
		offer      LeaseOffer
	)
	ready := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		logs = append(logs, string(body))
		mu.Unlock()
		t.Logf("%5.1fs %s", time.Since(provider.origin).Seconds(), body)
	})
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		<-ready
		controller.setLeaseCookie(w, offer.Cookie)
		http.Redirect(w, r, "/index.html?mode="+mode+"#/api/media/hls/"+offer.PublicID+"/master.m3u8", http.StatusFound)
	})
	mux.HandleFunc("/api/media/hls/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/keepalive") {
			_ = controller.KeepAlive(w, r)
			return
		}
		_ = controller.Media(w, r)
	})
	mux.Handle("/", http.FileServer(http.Dir(static)))
	server := httptest.NewUnstartedServer(mux)
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()
	policy, err := NewSecurityPolicy([]string{server.URL}, nil)
	if err != nil {
		t.Fatal(err)
	}
	controller = &Controller{logger: log.Logger, leases: leases, backend: backend, security: policy, now: time.Now}

	browser := exec.Command("xvfb-run", "-a", python, script, server.URL+"/start", strconv.Itoa(browserPlaybackSeconds))
	done := make(chan []byte)
	go func() {
		out, _ := browser.CombinedOutput()
		done <- out
	}()
	defer func() {
		select {
		case <-ready:
		default:
			close(ready)
		}
	}()

	// Bootstrap equivalent: the browser waits on /start while the packager
	// cold-starts, then loads the master as soon as readiness is reported.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := packager.Acquire(ctx, mode); err != nil {
		t.Fatal(err)
	}
	defer packager.Release()
	if offer, err = leases.Issue(LeaseBinding{SessionID: "viewer", Mode: mode}, time.Now()); err != nil {
		t.Fatal(err)
	}
	t.Logf("%5.1fs packager ready", time.Since(provider.origin).Seconds())
	close(ready)
	output := <-done

	mu.Lock()
	defer mu.Unlock()
	var first, last float64 = -1, -1
	paused := ""
	for _, line := range logs {
		if strings.HasPrefix(line, "TERMINAL") || strings.Contains(line, `"fatal":true`) {
			t.Errorf("player failed: %s", line)
		}
		fields := strings.Fields(line)
		if len(fields) == 4 && fields[0] == "PROGRESS" {
			value, _ := strconv.ParseFloat(fields[1], 64)
			if first < 0 && value > 0 {
				first = value
			}
			last, paused = value, fields[3]
		}
	}
	if first < 0 || last-first < 5 || paused != "false" {
		t.Errorf("playback did not keep progressing: first %.2fs, last %.2fs, paused %q\nbrowser output:\n%s", first, last, paused, output)
	}
}
