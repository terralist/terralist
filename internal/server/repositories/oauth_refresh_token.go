package repositories

import (
	"errors"
	"fmt"
	"time"

	"terralist/internal/server/models/oauth"
	"terralist/pkg/database"

	"gorm.io/gorm"
)

// OAuthRefreshTokenRepository stores the refresh sessions issued to OAuth
// clients.
type OAuthRefreshTokenRepository interface {
	// Create stores a new refresh session, removing the expired ones.
	Create(token *oauth.RefreshToken) error

	// FindByHash returns the session whose current token has the given hash.
	FindByHash(hash string) (*oauth.RefreshToken, error)

	// FindByPreviousHash returns the session whose last rotated-out token has
	// the given hash.
	FindByPreviousHash(hash string) (*oauth.RefreshToken, error)

	// Rotate replaces the token hash of a session, but only while its current
	// hash is still currentHash. It returns false when another request rotated
	// or removed the session first.
	Rotate(id, currentHash, newHash, previousHash string, expiresAt time.Time) (bool, error)

	// UpdateGrant replaces the sealed upstream grant of a session.
	UpdateGrant(id, grant string) error

	// Delete removes a session.
	Delete(id string) error

	// DeleteAll removes every session.
	DeleteAll() error
}

// DefaultOAuthRefreshTokenRepository is a concrete implementation of
// OAuthRefreshTokenRepository.
type DefaultOAuthRefreshTokenRepository struct {
	Database database.Engine

	now func() time.Time
}

func (r *DefaultOAuthRefreshTokenRepository) clock() time.Time {
	if r.now != nil {
		return r.now()
	}

	return time.Now()
}

func (r *DefaultOAuthRefreshTokenRepository) Create(token *oauth.RefreshToken) error {
	now := r.clock()

	if err := r.Database.Handler().
		Where("expires_at < ? OR session_expires_at < ?", now, now).
		Delete(&oauth.RefreshToken{}).
		Error; err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	if err := r.Database.Handler().Create(token).Error; err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	return nil
}

func (r *DefaultOAuthRefreshTokenRepository) FindByHash(hash string) (*oauth.RefreshToken, error) {
	return r.findBy("token_hash", hash)
}

func (r *DefaultOAuthRefreshTokenRepository) FindByPreviousHash(hash string) (*oauth.RefreshToken, error) {
	return r.findBy("previous_token_hash", hash)
}

func (r *DefaultOAuthRefreshTokenRepository) findBy(column, hash string) (*oauth.RefreshToken, error) {
	if hash == "" {
		return nil, ErrNotFound
	}

	var token oauth.RefreshToken
	if err := r.Database.Handler().Where(column+" = ?", hash).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	return &token, nil
}

func (r *DefaultOAuthRefreshTokenRepository) Rotate(id, currentHash, newHash, previousHash string, expiresAt time.Time) (bool, error) {
	result := r.Database.Handler().
		Model(&oauth.RefreshToken{}).
		Where("id = ? AND token_hash = ?", id, currentHash).
		Updates(map[string]any{
			"token_hash":          newHash,
			"previous_token_hash": previousHash,
			"expires_at":          expiresAt,
		})
	if result.Error != nil {
		return false, fmt.Errorf("%w: %v", ErrDatabaseFailure, result.Error)
	}

	return result.RowsAffected == 1, nil
}

func (r *DefaultOAuthRefreshTokenRepository) UpdateGrant(id, grant string) error {
	if err := r.Database.Handler().
		Model(&oauth.RefreshToken{}).
		Where("id = ?", id).
		Update("upstream_grant", grant).
		Error; err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	return nil
}

func (r *DefaultOAuthRefreshTokenRepository) Delete(id string) error {
	if err := r.Database.Handler().
		Where("id = ?", id).
		Delete(&oauth.RefreshToken{}).
		Error; err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	return nil
}

func (r *DefaultOAuthRefreshTokenRepository) DeleteAll() error {
	if err := r.Database.Handler().
		Where("1 = 1").
		Delete(&oauth.RefreshToken{}).
		Error; err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	return nil
}
