package mediahls

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func testLeaseStore(t *testing.T) (*LeaseStore, LeaseOffer, time.Time) {
	t.Helper()
	store, err := NewLeaseStore(2, 8)
	if err != nil { t.Fatal(err) }
	store.random = deterministicRandom()
	now := time.Unix(1_700_000_000,0)
	offer, err := store.Issue(LeaseBinding{SessionID:"viewer", Mode:ModeLLHLS}, now)
	if err != nil { t.Fatal(err) }
	return store, offer, now
}

func TestLeaseAuthorizationCookieScopeReplacementAndExpiry(t *testing.T) {
	store, offer, now := testLeaseStore(t)
	if len(offer.PublicID) != PublicIDEncodedLength || len(offer.Secret) != LeaseSecretEncodedLength { t.Fatalf("offer = %#v", offer) }
	cookie := offer.Cookie
	if cookie.Name != LeaseCookieName || cookie.Path != "/api/media/hls/"+offer.PublicID+"/" || cookie.MaxAge != 30 || !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode { t.Fatalf("cookie = %#v", cookie) }
	if _, err := store.Authenticate(offer.PublicID, offer.Secret, false, now); err != nil { t.Fatal(err) }
	if _, err := store.Authenticate(offer.PublicID, offer.Secret+"x", false, now); !errors.Is(err, ErrLeaseNotFound) { t.Fatalf("wrong secret error = %v", err) }
	replacement, err := store.Issue(LeaseBinding{SessionID:"viewer", Mode:ModeHLS}, now)
	if err != nil { t.Fatal(err) }
	if _, err := store.Authenticate(offer.PublicID, offer.Secret, false, now); !errors.Is(err, ErrLeaseNotFound) { t.Fatalf("replaced lease error = %v", err) }
	if _, err := store.Authenticate(replacement.PublicID, replacement.Secret, false, now.Add(LeaseLifetime)); !errors.Is(err, ErrLeaseNotFound) { t.Fatalf("expired lease error = %v", err) }
}

func TestLeaseSlidingActivityPauseAndRateLimits(t *testing.T) {
	store, offer, now := testLeaseStore(t)
	objectPermit, objectSnapshot, err := store.Acquire(offer.PublicID, offer.Secret, RequestObject, false, now.Add(20*time.Second))
	if err != nil { t.Fatal(err) }
	objectPermit.Release()
	if !objectSnapshot.ExpiresAt.Equal(offer.ExpiresAt) { t.Fatal("object request extended lease") }
	playlistPermit, playlistSnapshot, err := store.Acquire(offer.PublicID, offer.Secret, RequestPlaylist, true, now.Add(20*time.Second))
	if err != nil { t.Fatal(err) }
	playlistPermit.Release()
	if !playlistSnapshot.ExpiresAt.Equal(offer.ExpiresAt) { t.Fatal("playlist admission extended lease before a successful response") }
	playlistSnapshot, err = store.Authenticate(offer.PublicID, offer.Secret, true, now.Add(20*time.Second))
	if err != nil { t.Fatal(err) }
	if !playlistSnapshot.ExpiresAt.Equal(now.Add(20*time.Second+LeaseLifetime)) { t.Fatal("successful playlist did not extend lease") }
	permits := make([]*RequestPermit,0,KeepAliveBurst)
	for attempt:=0; attempt<KeepAliveBurst; attempt++ { permit,_,err := store.Acquire(offer.PublicID,offer.Secret,RequestKeepAlive,false,now.Add(21*time.Second)); if err != nil { t.Fatal(err) }; permits=append(permits,permit) }
	if _,_,err := store.Acquire(offer.PublicID,offer.Secret,RequestKeepAlive,false,now.Add(21*time.Second)); !errors.Is(err,ErrRequestLimit) { t.Fatalf("rate error = %v",err) }
	for _, permit := range permits { permit.Release() }
	if !store.SetPaused("viewer",true) { t.Fatal("pause failed") }
	if _,err := store.Authenticate(offer.PublicID,offer.Secret,false,now.Add(22*time.Second)); !errors.Is(err,ErrLeasePaused) { t.Fatalf("pause error = %v",err) }
	if expiresAt, err := store.Expiration(offer.PublicID, offer.Secret, now.Add(22*time.Second)); err != nil || !expiresAt.Equal(playlistSnapshot.ExpiresAt) {
		t.Fatalf("paused expiration = %v, %v", expiresAt, err)
	}
}

func TestLeaseRequestConcurrencyIsLocal(t *testing.T) {
	store, offer, now := testLeaseStore(t)
	permits := make([]*RequestPermit,0,MaximumLeaseRequests)
	for index:=0; index<MaximumLeaseRequests; index++ { permit,_,err := store.Acquire(offer.PublicID,offer.Secret,RequestObject,index<MaximumLeaseBlocking,now); if err != nil { t.Fatal(err) }; permits=append(permits,permit) }
	if _,_,err := store.Acquire(offer.PublicID,offer.Secret,RequestObject,false,now); !errors.Is(err,ErrRequestLimit) { t.Fatalf("concurrency error = %v",err) }
	for _,permit := range permits { permit.Release() }
}

func TestPausedKeepAliveRenewsWithoutAdmittingMediaOrRevivingExpiredLeases(t *testing.T) {
	store, offer, now := testLeaseStore(t)
	store.SetPaused("viewer", true)
	for _, seconds := range []int{15, 30, 45} {
		at := now.Add(time.Duration(seconds) * time.Second)
		permit, _, err := store.Acquire(offer.PublicID, offer.Secret, RequestKeepAlive, false, at)
		if err != nil { t.Fatal(err) }
		renewed, err := store.RenewKeepAlive(offer.PublicID, offer.Secret, at)
		permit.Release()
		if err != nil || !renewed.Paused || !renewed.ExpiresAt.Equal(at.Add(LeaseLifetime)) {
			t.Fatalf("paused renewal failed: %v", err)
		}
		if _, err := store.Authenticate(offer.PublicID, offer.Secret, true, at); !errors.Is(err, ErrLeasePaused) {
			t.Fatalf("media authentication while paused = %v", err)
		}
		for _, class := range []RequestClass{RequestPlaylist, RequestObject} {
			if _, _, err := store.Acquire(offer.PublicID, offer.Secret, class, false, at); !errors.Is(err, ErrLeasePaused) {
				t.Fatalf("media admission while paused = %v", err)
			}
		}
	}
	wrong := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	if wrong == offer.Secret { wrong = "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB" }
	if _, err := store.RenewKeepAlive(offer.PublicID, wrong, now.Add(46*time.Second)); !errors.Is(err, ErrLeaseNotFound) {
		t.Fatalf("wrong renewal credential = %v", err)
	}
	store.SetPaused("viewer", false)
	if _, err := store.Authenticate(offer.PublicID, offer.Secret, false, now.Add(46*time.Second)); err != nil { t.Fatal(err) }
	if _, err := store.RenewKeepAlive(offer.PublicID, offer.Secret, now.Add(75*time.Second)); !errors.Is(err, ErrLeaseNotFound) {
		t.Fatalf("expired lease revived = %v", err)
	}
	replacement, err := store.Issue(LeaseBinding{SessionID: "viewer", Mode: ModeHLS}, now.Add(80*time.Second))
	if err != nil { t.Fatal(err) }
	store.InvalidateSession("viewer")
	if _, err := store.RenewKeepAlive(replacement.PublicID, replacement.Secret, now.Add(81*time.Second)); !errors.Is(err, ErrLeaseNotFound) {
		t.Fatalf("revoked lease revived = %v", err)
	}
}

func TestPausedKeepAliveStillEnforcesRateLimit(t *testing.T) {
	store, offer, now := testLeaseStore(t)
	store.SetPaused("viewer", true)
	for index := 0; index < KeepAliveBurst; index++ {
		permit, _, err := store.Acquire(offer.PublicID, offer.Secret, RequestKeepAlive, false, now)
		if err != nil { t.Fatal(err) }
		permit.Release()
	}
	if _, _, err := store.Acquire(offer.PublicID, offer.Secret, RequestKeepAlive, false, now); !errors.Is(err, ErrRequestLimit) {
		t.Fatalf("paused keepalive rate limit = %v", err)
	}
}

func TestLeasePauseAndRevocationWakeOutstandingRequests(t *testing.T) {
	store, offer, now := testLeaseStore(t)
	paused, err := store.ChangeChannel(offer.PublicID, offer.Secret, now)
	if err != nil { t.Fatal(err) }
	if !store.SetPaused("viewer", true) { t.Fatal("pause failed") }
	if _, err := store.ChangeChannel(offer.PublicID, offer.Secret, now); !errors.Is(err, ErrLeasePaused) {
		t.Fatalf("late pause watcher error = %v", err)
	}
	select {
	case <-paused:
	default:
		t.Fatal("pause did not wake request")
	}

	replacement, err := store.Issue(LeaseBinding{SessionID:"viewer", Mode:ModeHLS}, now)
	if err != nil { t.Fatal(err) }
	revoked, err := store.ChangeChannel(replacement.PublicID, replacement.Secret, now)
	if err != nil { t.Fatal(err) }
	store.Invalidate(replacement.PublicID)
	select {
	case <-revoked:
	default:
		t.Fatal("revocation did not wake request")
	}
}
