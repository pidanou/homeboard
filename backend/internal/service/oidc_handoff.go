package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

const handoffTTL = 60 * time.Second

var ErrHandoffCodeInvalid = errors.New("oidc: handoff code invalid, expired, or already used")

type handoffEntry struct {
	jwt     string
	expires time.Time
}

// OIDCHandoffStore hands a freshly-minted JWT from the backend's OIDC callback
// (a server-to-server redirect target) to the SPA via a short-lived, single-use
// opaque code — the JWT itself never appears in a URL. Safe for a single backend
// replica only (in-memory); a horizontally-scaled deployment would need a shared
// store instead.
type OIDCHandoffStore struct {
	mu      sync.Mutex
	entries map[string]handoffEntry
}

func NewOIDCHandoffStore() *OIDCHandoffStore {
	return &OIDCHandoffStore{entries: make(map[string]handoffEntry)}
}

// Put stores jwt behind a new random code, valid for 60 seconds, and returns the code.
func (s *OIDCHandoffStore) Put(jwt string) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	code := base64.RawURLEncoding.EncodeToString(buf)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.evictExpiredLocked()
	s.entries[code] = handoffEntry{jwt: jwt, expires: time.Now().Add(handoffTTL)}
	return code, nil
}

// Consume returns the JWT for code and invalidates it. A code can only be
// redeemed once, and only within its TTL.
func (s *OIDCHandoffStore) Consume(code string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[code]
	delete(s.entries, code)
	if !ok || time.Now().After(entry.expires) {
		return "", ErrHandoffCodeInvalid
	}
	return entry.jwt, nil
}

// evictExpiredLocked sweeps stale entries. Called with mu held, opportunistically
// from Put so the map doesn't grow unbounded between logins.
func (s *OIDCHandoffStore) evictExpiredLocked() {
	now := time.Now()
	for code, entry := range s.entries {
		if now.After(entry.expires) {
			delete(s.entries, code)
		}
	}
}
