package mediahls

import (
	"errors"
	"testing"
	"time"
)

func testTicketBinding(sessionID, mode string) TicketBinding {
	return TicketBinding{SessionID: sessionID, Backend: BackendName, Version: ProtocolVersion, Mode: mode, Request: BootstrapRequest{Mode: mode, VariantIDs: []string{"high", "medium", "low"}, Audio: true}}
}

func deterministicRandom() func([]byte) (int,error) {
	next := byte(1)
	return func(target []byte) (int,error) { for index := range target { target[index] = next; next++ }; return len(target), nil }
}

func TestTicketSingleUseReplacementAndExpiry(t *testing.T) {
	store := NewTicketStore(); store.random = deterministicRandom()
	now := time.Unix(1_700_000_000,0)
	first, err := store.Issue(testTicketBinding("viewer", ModeHLS), now)
	if err != nil { t.Fatal(err) }
	second, err := store.Issue(testTicketBinding("viewer", ModeLLHLS), now)
	if err != nil { t.Fatal(err) }
	if _, err := store.Redeem(first.Ticket, now); !errors.Is(err, ErrTicketGone) { t.Fatalf("replaced ticket error = %v", err) }
	binding, err := store.Redeem(second.Ticket, now)
	if err != nil { t.Fatal(err) }
	if binding.SessionID != "viewer" || binding.Mode != ModeLLHLS || len(binding.Request.VariantIDs) != 3 { t.Fatalf("binding = %#v", binding) }
	if _, err := store.Redeem(second.Ticket, now); !errors.Is(err, ErrTicketGone) { t.Fatalf("replayed ticket error = %v", err) }
	expiring, err := store.Issue(testTicketBinding("other", ModeHLS), now)
	if err != nil { t.Fatal(err) }
	if _, err := store.Redeem(expiring.Ticket, now.Add(TicketLifetime)); !errors.Is(err, ErrTicketGone) { t.Fatalf("expired ticket error = %v", err) }
}

func TestTicketRejectsSyntaxAndBinding(t *testing.T) {
	store := NewTicketStore(); store.random = deterministicRandom()
	if _, err := store.Issue(TicketBinding{SessionID:"viewer", Backend:BackendName, Version:ProtocolVersion, Mode:ModeHLS}, time.Now()); !errors.Is(err, ErrTicketBinding) { t.Fatalf("binding error = %v", err) }
	for _, value := range []string{"", "not-base64________________________________"} {
		if !errors.Is(ValidateTicketSyntax(value), ErrTicketMalformed) { t.Fatalf("accepted ticket %q", value) }
	}
	valid := "AQIDBAUGBwgJCgsMDQ4PEBESExQVFhcY"
	if payload, err := DecodeBootstrapPayload([]byte(`{"ticket":"` + valid + `"}`)); err != nil || payload.Ticket != valid {
		t.Fatalf("bootstrap payload = %#v, %v", payload, err)
	}
	if _, err := DecodeBootstrapPayload([]byte(`{"ticket":"` + valid + `","extra":true}`)); err == nil {
		t.Fatal("bootstrap payload accepted unknown field")
	}
	if _, err := DecodeBootstrapPayload(make([]byte, MaximumBootstrapBody+1)); !errors.Is(err, ErrBootstrapBodyTooLarge) {
		t.Fatalf("oversized bootstrap error = %v", err)
	}
}
