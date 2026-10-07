package utils

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type writerConnection struct {
	mu       sync.Mutex
	messages []webSocketWrite
	deadlines []time.Time
	started  chan struct{}
	gate     chan struct{}
	closed   chan struct{}
	once     sync.Once
	first    sync.Once
	failure  error
}

func newWriterConnection(blocked bool) *writerConnection {
	connection := &writerConnection{started: make(chan struct{}), closed: make(chan struct{}), gate: make(chan struct{})}
	if !blocked { close(connection.gate) }
	return connection
}

func (connection *writerConnection) SetWriteDeadline(deadline time.Time) error {
	connection.mu.Lock()
	defer connection.mu.Unlock()
	connection.deadlines = append(connection.deadlines, deadline)
	return nil
}

func (connection *writerConnection) WriteMessage(kind int, data []byte) error {
	connection.first.Do(func() { close(connection.started) })
	select {
	case <-connection.gate:
	case <-connection.closed:
		return ErrWebSocketWriterClosed
	}
	connection.mu.Lock()
	defer connection.mu.Unlock()
	if connection.failure != nil { return connection.failure }
	connection.messages = append(connection.messages, webSocketWrite{messageType: kind, data: append([]byte(nil), data...)})
	return nil
}

func (connection *writerConnection) WriteControl(kind int, data []byte, _ time.Time) error {
	connection.mu.Lock()
	defer connection.mu.Unlock()
	connection.messages = append(connection.messages, webSocketWrite{messageType: kind, data: append([]byte(nil), data...)})
	return nil
}

func (connection *writerConnection) Close() error {
	connection.once.Do(func() { close(connection.closed) })
	return nil
}

func waitWriter(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(2 * time.Second): t.Fatal("writer did not terminate")
	}
}

func TestWebSocketWriterPreservesTerminalFIFOAndOwnsPayload(t *testing.T) {
	connection := newWriterConnection(true)
	writer := NewWebSocketWriter(connection)
	defer writer.Close()
	data := []byte("before")
	if err := writer.Send(websocket.TextMessage, data); err != nil { t.Fatal(err) }
	waitWriter(t, connection.started)
	copy(data, "mutate")
	writer.CloseAfterJSON(map[string]string{"event": "system/disconnect"})
	if err := writer.SendJSON("after"); !errors.Is(err, ErrWebSocketWriterClosed) { t.Fatalf("send after terminal: %v", err) }
	close(connection.gate)
	waitWriter(t, writer.Done())
	if len(connection.messages) != 3 || string(connection.messages[0].data) != "before" ||
		string(connection.messages[1].data) != `{"event":"system/disconnect"}` || connection.messages[2].messageType != websocket.CloseMessage {
		t.Fatalf("unexpected terminal sequence: %+v", connection.messages)
	}
	for _, deadline := range connection.deadlines {
		if deadline.IsZero() { t.Fatal("unbounded data write") }
	}
}

func TestWebSocketWriterIsolatesBlockedPeerAndClosesOverflow(t *testing.T) {
	slow, healthy := newWriterConnection(true), newWriterConnection(false)
	writer, other := NewWebSocketWriter(slow), NewWebSocketWriter(healthy)
	defer writer.Close()
	defer other.Close()
	if err := writer.SendJSON("blocked"); err != nil { t.Fatal(err) }
	waitWriter(t, slow.started)
	for i := 1; i < webSocketQueuedRecords; i++ {
		if err := writer.SendJSON(i); err != nil { t.Fatal(err) }
	}
	if err := other.SendJSON("healthy event"); err != nil { t.Fatal(err) }
	other.CloseGracefully()
	waitWriter(t, other.Done())
	if len(healthy.messages) != 2 { t.Fatalf("healthy peer stalled: %d messages", len(healthy.messages)) }
	if err := writer.SendJSON("overflow"); !errors.Is(err, ErrWebSocketQueueFull) { t.Fatalf("overflow: %v", err) }
	waitWriter(t, writer.Done())
	if writer.bytes != 0 || writer.records != 0 { t.Fatalf("retained payloads: %d/%d", writer.bytes, writer.records) }
}

func TestWebSocketWriterByteLimitAndWriteFailure(t *testing.T) {
	connection := newWriterConnection(false)
	writer := NewWebSocketWriter(connection)
	if err := writer.Send(websocket.TextMessage, make([]byte, webSocketQueuedBytes+1)); !errors.Is(err, ErrWebSocketQueueFull) { t.Fatal(err) }
	waitWriter(t, writer.Done())
	connection = newWriterConnection(false)
	connection.failure = errors.New("write failed")
	writer = NewWebSocketWriter(connection)
	if err := writer.SendJSON("event"); err != nil { t.Fatal(err) }
	waitWriter(t, writer.Done())
	if err := writer.SendJSON("later"); !errors.Is(err, connection.failure) { t.Fatalf("failure not retained: %v", err) }
}

func TestWebSocketWriterCloseInterruptsInFlightWrite(t *testing.T) {
	connection := newWriterConnection(true)
	writer := NewWebSocketWriter(connection)
	if err := writer.SendJSON("blocked"); err != nil { t.Fatal(err) }
	waitWriter(t, connection.started)
	writer.Close()
	waitWriter(t, writer.Done())
	if writer.bytes != 0 || writer.records != 0 { t.Fatal("in-flight accounting survived Close") }
}
