package mediaws

import "sync"

const (
	ProviderVideoQueueCapacity = 4
	ProviderAudioQueueCapacity = 16
	EgressRecordCapacity       = 24
	EgressPayloadCapacity      = 16 * 1024 * 1024
	LifecycleRecordCapacity    = 4
)

type queuedRecord struct {
	record       Record
	encoded      []byte
	payloadBytes int
}

type deliveryQueue struct {
	mu         sync.Mutex
	lifecycle  []queuedRecord
	media      []queuedRecord
	mediaBytes int
	closed     bool
	notify     chan struct{}
}

func newDeliveryQueue() *deliveryQueue {
	return &deliveryQueue{notify: make(chan struct{}, 1)}
}

func (queue *deliveryQueue) signal() {
	select {
	case queue.notify <- struct{}{}:
	default:
	}
}

func (queue *deliveryQueue) pushLifecycle(record queuedRecord) bool {
	queue.mu.Lock()
	if queue.closed || len(queue.lifecycle) >= LifecycleRecordCapacity {
		queue.mu.Unlock()
		return false
	}
	queue.lifecycle = append(queue.lifecycle, record)
	depth := len(queue.lifecycle) + len(queue.media)
	bytes := queue.mediaBytes
	queue.mu.Unlock()
	mediaWebSocketQueueDepth.Observe(float64(depth))
	mediaWebSocketQueueBytes.Observe(float64(bytes))
	queue.signal()
	return true
}

func (queue *deliveryQueue) pushMedia(record queuedRecord) bool {
	queue.mu.Lock()
	if queue.closed || len(queue.media) >= EgressRecordCapacity || queue.mediaBytes+record.payloadBytes > EgressPayloadCapacity {
		queue.mu.Unlock()
		return false
	}
	queue.media = append(queue.media, record)
	queue.mediaBytes += record.payloadBytes
	depth := len(queue.lifecycle) + len(queue.media)
	bytes := queue.mediaBytes
	queue.mu.Unlock()
	mediaWebSocketQueueDepth.Observe(float64(depth))
	mediaWebSocketQueueBytes.Observe(float64(bytes))
	queue.signal()
	return true
}

func (queue *deliveryQueue) pop() (queuedRecord, bool) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if len(queue.lifecycle) > 0 {
		record := queue.lifecycle[0]
		queue.lifecycle = queue.lifecycle[1:]
		return record, true
	}
	if len(queue.media) > 0 {
		record := queue.media[0]
		queue.media = queue.media[1:]
		queue.mediaBytes -= record.payloadBytes
		return record, true
	}
	return queuedRecord{}, false
}

func (queue *deliveryQueue) clearMedia(kind Kind) int {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	kept := queue.media[:0]
	dropped := 0
	queue.mediaBytes = 0
	for _, record := range queue.media {
		if kind == KindNone || record.record.Kind == kind {
			dropped++
			continue
		}
		kept = append(kept, record)
		queue.mediaBytes += record.payloadBytes
	}
	queue.media = kept
	return dropped
}

func (queue *deliveryQueue) finish(record queuedRecord) {
	queue.mu.Lock()
	queue.closed = true
	queue.media = nil
	queue.mediaBytes = 0
	queue.lifecycle = []queuedRecord{record}
	queue.mu.Unlock()
	queue.signal()
}

func (queue *deliveryQueue) stop() {
	queue.mu.Lock()
	queue.closed = true
	queue.media = nil
	queue.mediaBytes = 0
	queue.lifecycle = nil
	queue.mu.Unlock()
	queue.signal()
}
