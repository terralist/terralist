package repositories

import (
	"errors"
	"fmt"
	"time"

	"terralist/internal/server/models/authority"
	"terralist/pkg/database"

	"github.com/google/uuid"
)

var (
	ErrApiKeyExpired = errors.New("api key expired")
)

// ApiKeyRepository describes a service that can interact with the API keys database
type ApiKeyRepository interface {
	// Find searches for a specific ApiKey
	Find(id uuid.UUID) (*authority.ApiKey, error)

	// Create creates a new ApiKey
	Create(*authority.ApiKey) (*authority.ApiKey, error)

	// Delete removes an ApiKey from the database
	Delete(id uuid.UUID) error
}

// DefaultApiKeyRepository is a concrete implementation of ApiKeyRepository
type DefaultApiKeyRepository struct {
	Database database.Engine
}

func (r *DefaultApiKeyRepository) Find(id uuid.UUID) (*authority.ApiKey, error) {
	apiKey := &authority.ApiKey{}

	if err := r.Database.Handler().
		Where("id = ?", id).
		First(apiKey).
		Error; err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	if apiKey.Expiration != nil && time.Now().Unix() > apiKey.Expiration.Unix() {
		r.Database.Handler().Delete(apiKey)
		return nil, ErrApiKeyExpired
	}

	return apiKey, nil
}

func (r *DefaultApiKeyRepository) Create(apiKey *authority.ApiKey) (*authority.ApiKey, error) {
	if err := r.Database.Handler().Create(apiKey).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	return apiKey, nil
}

func (r *DefaultApiKeyRepository) Delete(id uuid.UUID) error {
	if err := r.Database.Handler().
		Where("id = ?", id).
		Delete(&authority.ApiKey{}).
		Error; err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	return nil
}
