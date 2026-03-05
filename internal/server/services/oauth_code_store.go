package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"terralist/internal/server/models/oauth"
)

// OAuthCodeStore stores OAuth code components behind opaque short-lived codes.
type OAuthCodeStore interface {
	Put(components oauth.CodeComponents) (string, error)
	Take(code string) (*oauth.CodeComponents, bool)
}

type oauthCodeEntry struct {
	components oauth.CodeComponents
	expiresAt  time.Time
}

type inMemoryOAuthCodeStore struct {
	ttl time.Duration
	now func() time.Time

	mu    sync.Mutex
	codes map[string]oauthCodeEntry
}

func NewInMemoryOAuthCodeStore(ttl time.Duration) OAuthCodeStore {
	return &inMemoryOAuthCodeStore{
		ttl:   ttl,
		now:   time.Now,
		codes: map[string]oauthCodeEntry{},
	}
}

func (s *inMemoryOAuthCodeStore) Put(components oauth.CodeComponents) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	code, err := newOAuthCodeID()
	if err != nil {
		return "", err
	}

	now := s.now()
	s.pruneExpired(now)
	s.codes[code] = oauthCodeEntry{
		components: components,
		expiresAt:  now.Add(s.ttl),
	}

	return code, nil
}

func (s *inMemoryOAuthCodeStore) Take(code string) (*oauth.CodeComponents, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	s.pruneExpired(now)

	entry, ok := s.codes[code]
	if !ok || now.After(entry.expiresAt) {
		delete(s.codes, code)
		return nil, false
	}

	delete(s.codes, code)
	components := entry.components
	return &components, true
}

func (s *inMemoryOAuthCodeStore) pruneExpired(now time.Time) {
	for code, entry := range s.codes {
		if now.After(entry.expiresAt) {
			delete(s.codes, code)
		}
	}
}

func newOAuthCodeID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("could not generate oauth code id: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
