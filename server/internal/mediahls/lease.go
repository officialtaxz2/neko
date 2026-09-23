package mediahls

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	PublicIDEntropyBytes  = 16
	PublicIDEncodedLength = 22
	LeaseSecretEntropyBytes = 24
	LeaseSecretEncodedLength = 32
	LeaseCookieName = "__Secure-neko-hls"
)

var (
	ErrLeaseNotFound = errors.New("HLS lease not found")
	ErrLeasePaused   = errors.New("HLS lease paused")
	ErrLeaseLimit    = errors.New("HLS lease limit exceeded")
	ErrRequestLimit  = errors.New("HLS request limit exceeded")
)

type LeaseBinding struct { SessionID string; Mode string }

type LeaseOffer struct {
	PublicID string
	Secret string
	ExpiresAt time.Time
	Cookie *http.Cookie
}

type LeaseSnapshot struct { SessionID string; Mode string; PublicID string; ExpiresAt time.Time; Paused bool }

type tokenBucket struct { tokens float64; last time.Time }

type leaseEntry struct {
	binding LeaseBinding
	secretDigest [sha256.Size]byte
	expiresAt time.Time
	paused bool
	active int
	blocking int
	keepalive tokenBucket
	playlists tokenBucket
	objects tokenBucket
}

type LeaseStore struct {
	mu sync.Mutex
	entries map[string]*leaseEntry
	bySession map[string]string
	maximumLeases int
	maximumRequests int
	active int
	blocking int
	random func([]byte) (int,error)
}

func NewLeaseStore(maximumLeases, maximumRequests int) (*LeaseStore, error) {
	if maximumLeases == 0 { maximumLeases = MaximumLeases }
	if maximumRequests == 0 { maximumRequests = MaximumRequests }
	if maximumLeases < 1 || maximumLeases > MaximumLeases || maximumRequests < 1 || maximumRequests > MaximumRequests {
		return nil, ErrInvalidConfig
	}
	return &LeaseStore{entries: map[string]*leaseEntry{}, bySession: map[string]string{}, maximumLeases: maximumLeases, maximumRequests: maximumRequests, random: rand.Read}, nil
}

func (store *LeaseStore) Issue(binding LeaseBinding, now time.Time) (LeaseOffer, error) {
	if binding.SessionID == "" || !ValidMode(binding.Mode) { return LeaseOffer{}, ErrLeaseNotFound }
	if now.IsZero() { now = time.Now() }
	publicID, err := store.randomToken(PublicIDEntropyBytes, PublicIDEncodedLength)
	if err != nil { return LeaseOffer{}, err }
	secret, err := store.randomToken(LeaseSecretEntropyBytes, LeaseSecretEncodedLength)
	if err != nil { return LeaseOffer{}, err }
	digest := sha256.Sum256([]byte(secret))
	expiresAt := now.Add(LeaseLifetime)
	store.mu.Lock()
	defer store.mu.Unlock()
	store.cleanupLocked(now)
	if _, exists := store.entries[publicID]; exists { return LeaseOffer{}, errors.New("HLS public ID collision") }
	previous, replaces := store.bySession[binding.SessionID]
	if !replaces && len(store.entries) >= store.maximumLeases { return LeaseOffer{}, ErrLeaseLimit }
	if replaces { delete(store.entries, previous) }
	store.entries[publicID] = &leaseEntry{binding: binding, secretDigest: digest, expiresAt: expiresAt}
	store.bySession[binding.SessionID] = publicID
	return LeaseOffer{PublicID: publicID, Secret: secret, ExpiresAt: expiresAt, Cookie: LeaseCookie(publicID, secret)}, nil
}

func (store *LeaseStore) randomToken(entropy, encoded int) (string, error) {
	raw := make([]byte, entropy)
	read, err := store.random(raw)
	if err != nil { return "", fmt.Errorf("generate HLS lease credential: %w", err) }
	if read != len(raw) { return "", fmt.Errorf("generate HLS lease credential: short random read %d", read) }
	token := base64.RawURLEncoding.EncodeToString(raw)
	if len(token) != encoded { return "", errors.New("unexpected HLS lease credential length") }
	return token, nil
}

func LeaseCookie(publicID, secret string) *http.Cookie {
	return &http.Cookie{Name: LeaseCookieName, Value: secret, Path: "/api/media/hls/"+publicID+"/", MaxAge: int(LeaseLifetime.Seconds()), Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode}
}

func (store *LeaseStore) Authenticate(publicID, secret string, extend bool, now time.Time) (LeaseSnapshot, error) {
	if ValidatePublicID(publicID) != nil || len(secret) != LeaseSecretEncodedLength { return LeaseSnapshot{}, ErrLeaseNotFound }
	if now.IsZero() { now = time.Now() }
	digest := sha256.Sum256([]byte(secret))
	store.mu.Lock()
	defer store.mu.Unlock()
	store.cleanupLocked(now)
	entry, ok := store.entries[publicID]
	if !ok || !hmac.Equal(digest[:], entry.secretDigest[:]) { return LeaseSnapshot{}, ErrLeaseNotFound }
	if entry.paused { return LeaseSnapshot{}, ErrLeasePaused }
	if extend { entry.expiresAt = now.Add(LeaseLifetime) }
	return snapshot(publicID, entry), nil
}

func (store *LeaseStore) SetPaused(sessionID string, paused bool) bool {
	store.mu.Lock(); defer store.mu.Unlock()
	publicID, ok := store.bySession[sessionID]
	if !ok { return false }
	entry, ok := store.entries[publicID]
	if !ok { return false }
	entry.paused = paused
	return true
}

func (store *LeaseStore) InvalidateSession(sessionID string) {
	store.mu.Lock(); defer store.mu.Unlock()
	if publicID, ok := store.bySession[sessionID]; ok { delete(store.entries, publicID); delete(store.bySession, sessionID) }
}

func (store *LeaseStore) cleanupLocked(now time.Time) {
	for publicID, entry := range store.entries {
		if !now.Before(entry.expiresAt) {
			delete(store.entries, publicID)
			if current, ok := store.bySession[entry.binding.SessionID]; ok && current == publicID { delete(store.bySession, entry.binding.SessionID) }
		}
	}
}

func snapshot(publicID string, entry *leaseEntry) LeaseSnapshot {
	return LeaseSnapshot{SessionID: entry.binding.SessionID, Mode: entry.binding.Mode, PublicID: publicID, ExpiresAt: entry.expiresAt, Paused: entry.paused}
}

type RequestClass uint8
const ( RequestKeepAlive RequestClass = iota + 1; RequestPlaylist; RequestObject )

type RequestPermit struct { store *LeaseStore; publicID string; blocking bool; once sync.Once }
func (permit *RequestPermit) Release() { if permit != nil { permit.once.Do(func(){ permit.store.release(permit.publicID, permit.blocking) }) } }

func (store *LeaseStore) Acquire(publicID, secret string, class RequestClass, blocking bool, now time.Time) (*RequestPermit, LeaseSnapshot, error) {
	if ValidatePublicID(publicID) != nil || len(secret) != LeaseSecretEncodedLength { return nil, LeaseSnapshot{}, ErrLeaseNotFound }
	if now.IsZero() { now = time.Now() }
	digest := sha256.Sum256([]byte(secret))
	store.mu.Lock()
	defer store.mu.Unlock()
	store.cleanupLocked(now)
	entry, ok := store.entries[publicID]
	if !ok || !hmac.Equal(digest[:], entry.secretDigest[:]) { return nil, LeaseSnapshot{}, ErrLeaseNotFound }
	if entry.paused { return nil, LeaseSnapshot{}, ErrLeasePaused }
	if store.active >= store.maximumRequests || entry.active >= MaximumLeaseRequests || (blocking && (store.blocking >= MaximumBlockingRequests || entry.blocking >= MaximumLeaseBlocking)) {
		return nil, LeaseSnapshot{}, ErrRequestLimit
	}
	var bucket *tokenBucket
	var rate, burst float64
	switch class {
	case RequestKeepAlive: bucket, rate, burst = &entry.keepalive, KeepAlivesPerMinute/60.0, KeepAliveBurst
	case RequestPlaylist: bucket, rate, burst = &entry.playlists, PlaylistsPerSecond, PlaylistBurst
	case RequestObject: bucket, rate, burst = &entry.objects, ObjectsPerSecond, ObjectBurst
	default: return nil, LeaseSnapshot{}, ErrRequestLimit
	}
	if !takeToken(bucket, rate, burst, now) { return nil, LeaseSnapshot{}, ErrRequestLimit }
	if class == RequestKeepAlive || class == RequestPlaylist { entry.expiresAt = now.Add(LeaseLifetime) }
	store.active++; entry.active++
	if blocking { store.blocking++; entry.blocking++ }
	return &RequestPermit{store: store, publicID: publicID, blocking: blocking}, snapshot(publicID, entry), nil
}

func takeToken(bucket *tokenBucket, rate, burst float64, now time.Time) bool {
	if bucket.last.IsZero() { bucket.tokens, bucket.last = burst, now
	} else if now.After(bucket.last) { bucket.tokens += now.Sub(bucket.last).Seconds()*rate; if bucket.tokens > burst { bucket.tokens = burst }; bucket.last = now }
	if bucket.tokens < 1 { return false }
	bucket.tokens--
	return true
}

func (store *LeaseStore) release(publicID string, blocking bool) {
	store.mu.Lock(); defer store.mu.Unlock()
	if store.active > 0 { store.active-- }
	if blocking && store.blocking > 0 { store.blocking-- }
	if entry, ok := store.entries[publicID]; ok { if entry.active > 0 { entry.active-- }; if blocking && entry.blocking > 0 { entry.blocking-- } }
}
