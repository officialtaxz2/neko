package utils

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Safety ceilings for the common event plane, not latency tuning targets.
	WebSocketWriteTimeout  = 5 * time.Second
	webSocketQueuedRecords = 128
	webSocketQueuedBytes   = 16 * 1024 * 1024
)

var (
	ErrWebSocketWriterClosed = errors.New("websocket writer closed")
	ErrWebSocketQueueFull    = errors.New("websocket event queue full")
)

type webSocketWriteConnection interface {
	SetWriteDeadline(time.Time) error
	WriteMessage(int, []byte) error
	WriteControl(int, []byte, time.Time) error
	Close() error
}

type webSocketWrite struct {
	messageType int
	data        []byte
	last        bool
}

// WebSocketWriter gives each event connection one data writer. Send never waits
// for network I/O. Overflow terminates the connection instead of dropping a room
// or authorization event. The byte/record ceilings include the in-flight write.
type WebSocketWriter struct {
	connection webSocketWriteConnection
	queue      chan webSocketWrite
	stop       chan struct{}
	done       chan struct{}
	mu         sync.Mutex
	bytes      int
	records    int
	closing    bool
	closed     bool
	err        error
	closeTimer *time.Timer
}

func NewWebSocketWriter(connection webSocketWriteConnection) *WebSocketWriter {
	writer := &WebSocketWriter{
		connection: connection,
		queue:      make(chan webSocketWrite, webSocketQueuedRecords),
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
	}
	go writer.run()
	return writer
}

func (writer *WebSocketWriter) Done() <-chan struct{} { return writer.done }

func (writer *WebSocketWriter) SendJSON(payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return writer.Send(websocket.TextMessage, data)
}

func (writer *WebSocketWriter) Send(messageType int, data []byte) error {
	return writer.enqueue(webSocketWrite{messageType: messageType, data: data})
}

// CloseAfterJSON preserves terminal-message order on a healthy socket. The
// writer and a force-close timer bound the flush; revocation must happen first.
func (writer *WebSocketWriter) CloseAfterJSON(payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		writer.abort(err)
		return
	}
	_ = writer.enqueue(webSocketWrite{messageType: websocket.TextMessage, data: data, last: true})
}

func (writer *WebSocketWriter) CloseGracefully() {
	_ = writer.enqueue(webSocketWrite{messageType: -1, last: true})
}

// Close interrupts a blocked write independently of the queue and its mutex.
func (writer *WebSocketWriter) Close() { writer.abort(ErrWebSocketWriterClosed) }

func (writer *WebSocketWriter) enqueue(request webSocketWrite) error {
	writer.mu.Lock()
	if writer.closing {
		err := writer.err
		if err == nil {
			err = ErrWebSocketWriterClosed
		}
		writer.mu.Unlock()
		return err
	}
	if writer.records >= webSocketQueuedRecords || len(request.data) > webSocketQueuedBytes-writer.bytes {
		writer.mu.Unlock()
		writer.abort(ErrWebSocketQueueFull)
		return ErrWebSocketQueueFull
	}
	// Callers can release or mutate their buffer as soon as Send returns.
	request.data = append([]byte(nil), request.data...)
	writer.records++
	writer.bytes += len(request.data)
	if request.last {
		writer.closing = true
		writer.closeTimer = time.AfterFunc(WebSocketWriteTimeout, func() {
			writer.abort(ErrWebSocketWriterClosed)
		})
	}
	writer.queue <- request // Capacity was checked, including in-flight records.
	writer.mu.Unlock()
	return nil
}

func (writer *WebSocketWriter) abort(err error) {
	writer.mu.Lock()
	if writer.closed {
		writer.mu.Unlock()
		return
	}
	writer.closed, writer.closing, writer.err = true, true, err
	if writer.closeTimer != nil {
		writer.closeTimer.Stop()
	}
	close(writer.stop)
	// Release queued payloads now, rather than retaining them with a dead peer.
	for {
		select {
		case request := <-writer.queue:
			writer.records--
			writer.bytes -= len(request.data)
		default:
			writer.mu.Unlock()
			_ = writer.connection.Close()
			return
		}
	}
}

func (writer *WebSocketWriter) run() {
	defer close(writer.done)
	for {
		select {
		case <-writer.stop:
			return
		case request := <-writer.queue:
			select {
			case <-writer.stop:
				writer.release(request)
				return
			default:
			}
			var err error
			if request.messageType != -1 {
				err = writer.connection.SetWriteDeadline(time.Now().Add(WebSocketWriteTimeout))
				if err == nil {
					err = writer.connection.WriteMessage(request.messageType, request.data)
				}
			}
			writer.release(request)
			if err != nil {
				writer.abort(err)
				return
			}
			if request.last {
				_ = writer.connection.WriteControl(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
					time.Now().Add(WebSocketWriteTimeout))
				writer.abort(nil)
				return
			}
		}
	}
}

func (writer *WebSocketWriter) release(request webSocketWrite) {
	writer.mu.Lock()
	writer.records--
	writer.bytes -= len(request.data)
	writer.mu.Unlock()
}
