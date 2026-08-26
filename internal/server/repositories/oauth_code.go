package repositories

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"terralist/internal/server/models/oauth"
	"terralist/pkg/database"

	"gorm.io/gorm"
)

// OAuthCodeRepository stores OAuth code components behind opaque, short-lived,
// single-use codes.
type OAuthCodeRepository interface {
	// Put stores the given code components and returns an opaque code that can
	// be exchanged for them exactly once.
	Put(components oauth.CodeComponents) (string, error)

	// Take resolves and removes the code components behind an opaque code. The
	// second return value is false when the code is unknown or expired.
	Take(code string) (*oauth.CodeComponents, bool)
}

// DefaultOAuthCodeRepository is a concrete implementation of OAuthCodeRepository.
type DefaultOAuthCodeRepository struct {
	Database database.Engine
	TTL      time.Duration

	now func() time.Time
}

func (r *DefaultOAuthCodeRepository) clock() time.Time {
	if r.now != nil {
		return r.now()
	}

	return time.Now()
}

func (r *DefaultOAuthCodeRepository) Put(components oauth.CodeComponents) (string, error) {
	data, err := json.Marshal(components)
	if err != nil {
		return "", fmt.Errorf("could not encode code components: %w", err)
	}

	code, err := newOAuthCode()
	if err != nil {
		return "", err
	}

	now := r.clock()

	if err := r.Database.Handler().
		Where("expires_at < ?", now).
		Delete(&oauth.Code{}).
		Error; err != nil {
		return "", fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	entry := &oauth.Code{
		Code:       code,
		Components: string(data),
		ExpiresAt:  now.Add(r.TTL),
	}

	if err := r.Database.Handler().Create(entry).Error; err != nil {
		return "", fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	return code, nil
}

func (r *DefaultOAuthCodeRepository) Take(code string) (*oauth.CodeComponents, bool) {
	var entry oauth.Code

	err := r.Database.Handler().Transaction(func(tx *database.DB) error {
		if err := tx.Where("code = ?", code).First(&entry).Error; err != nil {
			return err
		}

		// The delete is the atomic claim: only the transaction that actually
		// removes the row may resolve the code, so a code cannot be used twice
		// even when two exchanges race for it.
		result := tx.Delete(&entry)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
	if err != nil {
		return nil, false
	}

	if r.clock().After(entry.ExpiresAt) {
		return nil, false
	}

	var components oauth.CodeComponents
	if err := json.Unmarshal([]byte(entry.Components), &components); err != nil {
		return nil, false
	}

	return &components, true
}

func newOAuthCode() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("could not generate oauth code: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
