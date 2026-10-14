package mediaws

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testBinding(session string) TicketBinding {
	return TicketBinding{
		SessionID: session,
		Backend:   BackendName,
		Version:   ProtocolVersion,
		Request: NegotiatedRequest{
			Video: NegotiatedKind{Enabled: true, SourceID: "high", Codec: "vp8"},
			Audio: NegotiatedKind{Enabled: true, SourceID: "audio", Codec: "opus"},
		},
	}
}

func TestTicketSingleUseBindingAndReplacement(t *testing.T) {
	store := NewTicketStore()
	now := time.Unix(1_700_000_000, 0)

	first, err := store.Issue(testBinding("session-a"), now)
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Issue(testBinding("session-a"), now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if first.Ticket == second.Ticket {
		t.Fatal("tickets unexpectedly equal")
	}
	if _, err := store.Redeem(first.Ticket, now.Add(2*time.Second)); !errors.Is(err, ErrTicketGone) {
		t.Fatalf("replaced ticket error = %v", err)
	}
	binding, err := store.Redeem(second.Ticket, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if binding != testBinding("session-a") {
		t.Fatalf("redeemed binding = %#v", binding)
	}
	if _, err := store.Redeem(second.Ticket, now.Add(3*time.Second)); !errors.Is(err, ErrTicketGone) {
		t.Fatalf("replay error = %v", err)
	}
}

func TestTicketExpiresAndInvalidates(t *testing.T) {
	store := NewTicketStore()
	now := time.Unix(1_700_000_000, 0)
	offer, err := store.Issue(testBinding("session-a"), now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Redeem(offer.Ticket, now.Add(TicketLifetime)); !errors.Is(err, ErrTicketGone) {
		t.Fatalf("expired ticket error = %v", err)
	}

	offer, err = store.Issue(testBinding("session-a"), now.Add(2*TicketLifetime))
	if err != nil {
		t.Fatal(err)
	}
	if !store.Pending("session-a", now.Add(2*TicketLifetime)) {
		t.Fatal("fresh ticket is not pending")
	}
	store.InvalidateSession("session-a")
	if store.Pending("session-a", now.Add(2*TicketLifetime)) {
		t.Fatal("invalidated ticket remained pending")
	}
	if _, err := store.Redeem(offer.Ticket, now.Add(2*TicketLifetime)); !errors.Is(err, ErrTicketGone) {
		t.Fatalf("invalidated ticket error = %v", err)
	}
}

func TestTicketRedemptionIsAtomic(t *testing.T) {
	store := NewTicketStore()
	now := time.Unix(1_700_000_000, 0)
	offer, err := store.Issue(testBinding("session-a"), now)
	if err != nil {
		t.Fatal(err)
	}

	var successes atomic.Int32
	var unexpected atomic.Int32
	var wait sync.WaitGroup
	for attempt := 0; attempt < 16; attempt++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if _, err := store.Redeem(offer.Ticket, now); err == nil {
				successes.Add(1)
			} else if !errors.Is(err, ErrTicketGone) {
				unexpected.Add(1)
			}
		}()
	}
	wait.Wait()
	if successes.Load() != 1 || unexpected.Load() != 0 {
		t.Fatalf("redemption results: successes=%d unexpected=%d", successes.Load(), unexpected.Load())
	}
}

func TestTicketRejectsMalformedBindingRandomFailureAndCollision(t *testing.T) {
	store := NewTicketStore()
	now := time.Unix(1_700_000_000, 0)
	if _, err := store.Redeem("short", now); !errors.Is(err, ErrTicketMalformed) {
		t.Fatalf("malformed ticket error = %v", err)
	}
	if _, err := store.Redeem(strings.Repeat("!", TicketEncodedLength), now); !errors.Is(err, ErrTicketMalformed) {
		t.Fatalf("non-Base64URL ticket error = %v", err)
	}
	if _, err := store.Redeem(strings.Repeat("A", TicketEncodedLength), now); !errors.Is(err, ErrTicketUnknown) {
		t.Fatalf("unknown ticket error = %v", err)
	}
	invalid := testBinding("session-a")
	invalid.Request.Video.Codec = "h264"
	if _, err := store.Issue(invalid, now); err == nil {
		t.Fatal("invalid binding accepted")
	}

	store.random = func(buffer []byte) (int, error) { return len(buffer) - 1, nil }
	if _, err := store.Issue(testBinding("session-a"), now); err == nil {
		t.Fatal("short random read accepted")
	}

	store.random = func(buffer []byte) (int, error) {
		for index := range buffer {
			buffer[index] = 0x42
		}
		return len(buffer), nil
	}
	if _, err := store.Issue(testBinding("session-a"), now); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Issue(testBinding("session-b"), now); err == nil {
		t.Fatal("ticket collision accepted")
	}
}
