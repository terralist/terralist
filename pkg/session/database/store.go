package database

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"terralist/pkg/auth"
	db "terralist/pkg/database"
	"terralist/pkg/session"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func init() {
	gob.Register(&auth.User{})
	gob.Register(map[any]any{})
}

// SessionRecord is the GORM model for a database-backed session.
type SessionRecord struct {
	ID        string `gorm:"primaryKey"`
	Data      []byte `gorm:"type:blob"`
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SessionRecord) TableName() string {
	return "sessions"
}

// Store is a concrete implementation of session.Store backed by a database.
type Store struct {
	cookieName string
	secret     string
	maxAge     int
	database   db.Engine
}

func (s *Store) Get(r *http.Request) (session.Session, error) {
	cookie, err := r.Cookie(s.cookieName)
	if err != nil {
		return s.New(r)
	}

	sessionID := cookie.Value
	if !s.verifyID(sessionID) {
		return s.New(r)
	}

	rawID := s.extractRawID(sessionID)

	var record SessionRecord
	err = s.database.Handler().Where("id = ? AND expires_at > ?", rawID, time.Now()).First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.New(r)
		}
		return nil, fmt.Errorf("failed to fetch session: %w", err)
	}

	sess := newSession(rawID, false)
	if err := s.decode(record.Data, sess); err != nil {
		return s.New(r)
	}

	return sess, nil
}

func (s *Store) New(_ *http.Request) (session.Session, error) {
	id, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	return newSession(id, true), nil
}

func (s *Store) Save(r *http.Request, w http.ResponseWriter, sess session.Session) error {
	impl, ok := sess.(*Session)
	if !ok {
		return fmt.Errorf("unsupported session type")
	}

	data, err := s.encode(impl)
	if err != nil {
		return fmt.Errorf("failed to encode session: %w", err)
	}

	record := SessionRecord{
		ID:        impl.id,
		Data:      data,
		ExpiresAt: time.Now().Add(time.Duration(s.maxAge) * time.Second),
	}

	if impl.isNew {
		if err := s.database.Handler().Create(&record).Error; err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
		impl.isNew = false
	} else {
		if err := s.database.Handler().Save(&record).Error; err != nil {
			return fmt.Errorf("failed to update session: %w", err)
		}
	}

	signedID := s.signID(impl.id)
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName,
		Value:    signedID,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   s.maxAge,
	})

	return nil
}

// Cleanup removes expired sessions from the database. It is intended to be
// called periodically from a background goroutine.
func (s *Store) Cleanup() {
	result := s.database.Handler().Where("expires_at < ?", time.Now()).Delete(&SessionRecord{})
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to clean up expired sessions")
		return
	}
	if result.RowsAffected > 0 {
		log.Debug().Int64("count", result.RowsAffected).Msg("Cleaned up expired sessions")
	}
}

func (s *Store) encode(sess *Session) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(sess.values); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *Store) decode(data []byte, sess *Session) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(&sess.values)
}

func (s *Store) signID(id string) string {
	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write([]byte(id))
	sig := hex.EncodeToString(mac.Sum(nil))
	return id + "." + sig
}

func (s *Store) verifyID(signedID string) bool {
	dotIdx := len(signedID) - 65 // 64 hex chars + 1 dot
	if dotIdx <= 0 || signedID[dotIdx] != '.' {
		return false
	}

	id := signedID[:dotIdx]
	expected := s.signID(id)
	return hmac.Equal([]byte(signedID), []byte(expected))
}

func (s *Store) extractRawID(signedID string) string {
	dotIdx := len(signedID) - 65
	if dotIdx <= 0 {
		return signedID
	}
	return signedID[:dotIdx]
}

func generateID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
