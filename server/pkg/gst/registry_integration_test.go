//go:build hlsintegration

package gst

import (
	"testing"
	"time"
)

// A native parse failure must not need the sample registry. Holding the
// registry deliberately keeps this independent of plugin initialization speed:
// the prior constructor cannot return at all until we release that lock.
func TestNativeParseFailureDoesNotWaitForSampleRegistry(t *testing.T) {
	const missingElement = "neko_registry_probe_missing_element"
	if CheckElement(missingElement) == nil {
		t.Fatal("regression fixture element must be absent")
	}

	pipelinesLock.Lock()
	result := make(chan error, 1)
	go func() {
		_, err := CreatePipeline(missingElement)
		result <- err
	}()

	blocked := false
	var parseErr error
	select {
	case parseErr = <-result:
	case <-time.After(2 * time.Second):
		blocked = true
	}
	pipelinesLock.Unlock()

	if blocked {
		// Drain after unlocking so the expected negative control leaves no
		// constructor goroutine pending before this test reports its failure.
		select {
		case <-result:
		case <-time.After(5 * time.Second):
			t.Fatal("native parse did not return after registry release")
		}
		t.Fatal("native pipeline parse waited for sample registry")
	}
	if parseErr == nil {
		t.Fatal("expected fixed malformed-pipeline parse failure")
	}
}
