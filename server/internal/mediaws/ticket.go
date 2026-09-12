package mediaws

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	TicketEntropyBytes  = 24
	TicketEncodedLength = 32
	TicketLifetime      = 10 * time.Second
)

var (
	ErrTicketMalformed = errors.New("malformed media attach ticket")
	ErrTicketUnknown   = errors.New("unknown media attach ticket")
	ErrTicketBinding   = errors.New("media attach ticket binding invalid")
	ErrTicketGone      = errors.New("media attach ticket expired, invalidated, or already redeemed")
)

type NegotiatedKind struct {
	Enabled  bool
	SourceID string
	Codec    string
}

type NegotiatedRequest struct {
	Audio NegotiatedKind
	Video NegotiatedKind
}

type TicketBinding struct {
	SessionID string
	Backend   string
	Version   uint8
	Request   NegotiatedRequest
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
	return &TicketStore{
		entries:   make(map[[sha256.Size]byte]ticketEntry),
		bySession: make(map[string][sha256.Size]byte),
		random:    rand.Read,
	}
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
		return TicketOffer{}, fmt.Errorf("generate media attach ticket: %w", err)
	}
	if read != len(raw) {
		return TicketOffer{}, fmt.Errorf("generate media attach ticket: short random read %d", read)
	}
	ticket := base64.RawURLEncoding.EncodeToString(raw)
	if len(ticket) != TicketEncodedLength {
		return TicketOffer{}, errors.New("unexpected media attach ticket length")
	}
	digest := sha256.Sum256([]byte(ticket))
	expiresAt := now.Add(TicketLifetime)

	store.mu.Lock()
	defer store.mu.Unlock()
	store.cleanupLocked(now)
	if _, exists := store.entries[digest]; exists {
		return TicketOffer{}, errors.New("media attach ticket collision")
	}
	if previous, ok := store.bySession[binding.SessionID]; ok {
		entry := store.entries[previous]
		entry.gone = true
		store.entries[previous] = entry
	}
	store.entries[digest] = ticketEntry{binding: binding, expiresAt: expiresAt}
	store.bySession[binding.SessionID] = digest
	return TicketOffer{Ticket: ticket, ExpiresAt: expiresAt}, nil
}

// Redeem atomically consumes the one-time ticket and returns the server-owned
// binding needed by the future credential-free attachment route. The route
// must re-resolve SessionID and CanWatch before upgrading the connection.
func (store *TicketStore) Redeem(ticket string, now time.Time) (TicketBinding, error) {
	if err := validateTicketSyntax(ticket); err != nil {
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
	if now.IsZero() {
		now = time.Now()
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.cleanupLocked(now)
	digest, ok := store.bySession[sessionID]
	if !ok {
		return false
	}
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
		// Keep a bounded replay tombstone for one additional ticket lifetime so
		// expired/redeemed tickets can map to HTTP 410 instead of looking unknown.
		if !now.Before(entry.expiresAt.Add(TicketLifetime)) {
			delete(store.entries, digest)
		}
	}
}

func validateTicketSyntax(ticket string) error {
	if len(ticket) != TicketEncodedLength {
		return ErrTicketMalformed
	}
	raw, err := base64.RawURLEncoding.DecodeString(ticket)
	if err != nil || len(raw) != TicketEntropyBytes {
		return ErrTicketMalformed
	}
	return nil
}

func validateTicketBinding(binding TicketBinding) error {
	if binding.SessionID == "" || binding.Backend != BackendName || binding.Version != ProtocolVersion {
		return errors.New("invalid media attach ticket binding")
	}
	if !binding.Request.Video.Enabled || binding.Request.Video.SourceID == "" || len(binding.Request.Video.SourceID) > MaxSourceIDLength || binding.Request.Video.Codec != "vp8" {
		return errors.New("invalid media attach video request")
	}
	if binding.Request.Audio.Enabled {
		if binding.Request.Audio.SourceID == "" || len(binding.Request.Audio.SourceID) > MaxSourceIDLength || binding.Request.Audio.Codec != "opus" {
			return errors.New("invalid media attach audio request")
		}
	} else if binding.Request.Audio.SourceID != "" || binding.Request.Audio.Codec != "" {
		return errors.New("disabled media attach audio request must be empty")
	}
	return nil
}
