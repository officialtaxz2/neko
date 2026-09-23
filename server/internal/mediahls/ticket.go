package mediahls

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"
)

const (
	TicketEntropyBytes  = 24
	TicketEncodedLength = 32
	TicketLifetime      = 10 * time.Second
)

var (
	ErrTicketMalformed = errors.New("malformed HLS bootstrap ticket")
	ErrTicketUnknown   = errors.New("unknown HLS bootstrap ticket")
	ErrTicketBinding   = errors.New("HLS bootstrap ticket binding invalid")
	ErrTicketGone      = errors.New("HLS bootstrap ticket expired, invalidated, or already redeemed")
)

type BootstrapRequest struct {
	Mode       string
	VariantIDs []string
	Audio      bool
}

type TicketBinding struct {
	SessionID string
	Backend   string
	Version   uint8
	Mode      string
	Request   BootstrapRequest
}

type TicketOffer struct {
	Ticket    string
	ExpiresAt time.Time
}

type ticketEntry struct {
	binding   TicketBinding
	expiresAt time.Time
	gone      bool
}

type TicketStore struct {
	mu        sync.Mutex
	entries   map[[sha256.Size]byte]ticketEntry
	bySession map[string][sha256.Size]byte
	random    func([]byte) (int, error)
}

func NewTicketStore() *TicketStore {
	return &TicketStore{entries: make(map[[sha256.Size]byte]ticketEntry), bySession: make(map[string][sha256.Size]byte), random: rand.Read}
}

func (store *TicketStore) Issue(binding TicketBinding, now time.Time) (TicketOffer, error) {
	if err := validateTicketBinding(binding); err != nil {
		return TicketOffer{}, err
	}
	if now.IsZero() {
		now = time.Now()
	}
	raw := make([]byte, TicketEntropyBytes)
	read, err := store.random(raw)
	if err != nil {
		return TicketOffer{}, fmt.Errorf("generate HLS bootstrap ticket: %w", err)
	}
	if read != len(raw) {
		return TicketOffer{}, fmt.Errorf("generate HLS bootstrap ticket: short random read %d", read)
	}
	ticket := base64.RawURLEncoding.EncodeToString(raw)
	if len(ticket) != TicketEncodedLength {
		return TicketOffer{}, errors.New("unexpected HLS bootstrap ticket length")
	}
	digest := sha256.Sum256([]byte(ticket))
	expiresAt := now.Add(TicketLifetime)

	store.mu.Lock()
	defer store.mu.Unlock()
	store.cleanupLocked(now)
	if _, exists := store.entries[digest]; exists {
		return TicketOffer{}, errors.New("HLS bootstrap ticket collision")
	}
	if previous, ok := store.bySession[binding.SessionID]; ok {
		entry := store.entries[previous]
		entry.gone = true
		store.entries[previous] = entry
	}
	binding.Request.VariantIDs = slices.Clone(binding.Request.VariantIDs)
	store.entries[digest] = ticketEntry{binding: binding, expiresAt: expiresAt}
	store.bySession[binding.SessionID] = digest
	return TicketOffer{Ticket: ticket, ExpiresAt: expiresAt}, nil
}

func (store *TicketStore) Redeem(ticket string, now time.Time) (TicketBinding, error) {
	if err := ValidateTicketSyntax(ticket); err != nil {
		return TicketBinding{}, err
	}
	if now.IsZero() {
		now = time.Now()
	}
	digest := sha256.Sum256([]byte(ticket))
	store.mu.Lock()
	defer store.mu.Unlock()
	store.cleanupLocked(now)
	entry, ok := store.entries[digest]
	if !ok {
		return TicketBinding{}, ErrTicketUnknown
	}
	if entry.gone || !now.Before(entry.expiresAt) {
		entry.gone = true
		store.entries[digest] = entry
		return TicketBinding{}, ErrTicketGone
	}
	if err := validateTicketBinding(entry.binding); err != nil {
		entry.gone = true
		store.entries[digest] = entry
		return TicketBinding{}, fmt.Errorf("%w: %v", ErrTicketBinding, err)
	}
	entry.gone = true
	store.entries[digest] = entry
	if current, ok := store.bySession[entry.binding.SessionID]; ok && current == digest {
		delete(store.bySession, entry.binding.SessionID)
	}
	entry.binding.Request.VariantIDs = slices.Clone(entry.binding.Request.VariantIDs)
	return entry.binding, nil
}

func (store *TicketStore) InvalidateSession(sessionID string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if digest, ok := store.bySession[sessionID]; ok {
		entry := store.entries[digest]
		entry.gone = true
		store.entries[digest] = entry
		delete(store.bySession, sessionID)
	}
}

func (store *TicketStore) Pending(sessionID string, now time.Time) bool {
	if now.IsZero() { now = time.Now() }
	store.mu.Lock()
	defer store.mu.Unlock()
	store.cleanupLocked(now)
	digest, ok := store.bySession[sessionID]
	if !ok { return false }
	entry, ok := store.entries[digest]
	return ok && !entry.gone && now.Before(entry.expiresAt)
}

func (store *TicketStore) cleanupLocked(now time.Time) {
	for digest, entry := range store.entries {
		if !now.Before(entry.expiresAt) {
			entry.gone = true
			store.entries[digest] = entry
			if current, ok := store.bySession[entry.binding.SessionID]; ok && current == digest {
				delete(store.bySession, entry.binding.SessionID)
			}
		}
		if !now.Before(entry.expiresAt.Add(TicketLifetime)) {
			delete(store.entries, digest)
		}
	}
}

func ValidateTicketSyntax(ticket string) error {
	if len(ticket) != TicketEncodedLength { return ErrTicketMalformed }
	raw, err := base64.RawURLEncoding.DecodeString(ticket)
	if err != nil || len(raw) != TicketEntropyBytes { return ErrTicketMalformed }
	return nil
}

func validateTicketBinding(binding TicketBinding) error {
	if binding.SessionID == "" || binding.Backend != BackendName || binding.Version != ProtocolVersion || binding.Mode != binding.Request.Mode || !ValidMode(binding.Mode) || !binding.Request.Audio {
		return ErrTicketBinding
	}
	expected := []string{"high", "medium", "low"}
	if !slices.Equal(binding.Request.VariantIDs, expected) {
		return ErrTicketBinding
	}
	return nil
}
