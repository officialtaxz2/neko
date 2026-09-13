package mediaws

import "testing"

func testQueuedUnit(kind Kind, payload int) queuedRecord {
	return queuedRecord{
		record:       Record{Type: RecordUnit, Kind: kind},
		encoded:      make([]byte, HeaderLength+payload),
		payloadBytes: payload,
	}
}

func TestDeliveryQueueEnforcesRecordAndByteCaps(t *testing.T) {
	queue := newDeliveryQueue()
	for index := 0; index < EgressRecordCapacity; index++ {
		if !queue.pushMedia(testQueuedUnit(KindVideo, 1)) {
			t.Fatalf("record %d rejected below capacity", index)
		}
	}
	if queue.pushMedia(testQueuedUnit(KindVideo, 1)) {
		t.Fatal("record-cap overflow accepted")
	}

	queue = newDeliveryQueue()
	if !queue.pushMedia(testQueuedUnit(KindVideo, EgressPayloadCapacity)) {
		t.Fatal("payload at byte capacity rejected")
	}
	if queue.pushMedia(testQueuedUnit(KindAudio, 1)) {
		t.Fatal("byte-cap overflow accepted")
	}
}

func TestDeliveryQueuePrioritizesLifecycleAndClearsByKind(t *testing.T) {
	queue := newDeliveryQueue()
	queue.pushMedia(testQueuedUnit(KindAudio, 3))
	queue.pushMedia(testQueuedUnit(KindVideo, 5))
	lifecycle := queuedRecord{record: Record{Type: RecordFormat, Kind: KindVideo}}
	if !queue.pushLifecycle(lifecycle) {
		t.Fatal("lifecycle record rejected")
	}
	first, ok := queue.pop()
	if !ok || first.record.Type != RecordFormat {
		t.Fatalf("first queue record = %#v/%v", first.record, ok)
	}
	if dropped := queue.clearMedia(KindVideo); dropped != 1 {
		t.Fatalf("cleared video records = %d, want 1", dropped)
	}
	remaining, ok := queue.pop()
	if !ok || remaining.record.Kind != KindAudio {
		t.Fatalf("remaining queue record = %#v/%v", remaining.record, ok)
	}
}

func TestDeliveryQueueNeverSilentlyDropsLifecycle(t *testing.T) {
	queue := newDeliveryQueue()
	for index := 0; index < LifecycleRecordCapacity; index++ {
		if !queue.pushLifecycle(queuedRecord{record: Record{Type: RecordFormat, Kind: KindVideo}}) {
			t.Fatalf("lifecycle record %d rejected below capacity", index)
		}
	}
	if queue.pushLifecycle(queuedRecord{record: Record{Type: RecordFormat, Kind: KindVideo}}) {
		t.Fatal("lifecycle overflow was silently accepted")
	}

	end := queuedRecord{record: Record{Type: RecordEnd, Kind: KindNone}}
	queue.finish(end)
	record, ok := queue.pop()
	if !ok || record.record.Type != RecordEnd {
		t.Fatalf("finish record = %#v/%v", record.record, ok)
	}
	if queue.pushMedia(testQueuedUnit(KindVideo, 1)) || queue.pushLifecycle(end) {
		t.Fatal("closed queue accepted another record")
	}
}
