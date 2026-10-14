package legacy

import (
	"bytes"
	"encoding/json"
	"testing"

	oldMessage "github.com/m1k1o/neko/server/internal/http/legacy/message"
)

func TestSystemInitSessionIDIsOptInOnly(t *testing.T) {
	ordinary, err := json.Marshal(oldMessage.SystemInit{})
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(ordinary, []byte(`"session_id"`)) {
		t.Fatalf("ordinary system init exposed session_id: %s", ordinary)
	}

	selected, err := json.Marshal(oldMessage.SystemInit{SessionID: "session-1"})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(selected, []byte(`"session_id":"session-1"`)) {
		t.Fatalf("selected system init omitted session_id: %s", selected)
	}
}
