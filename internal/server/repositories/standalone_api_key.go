package repositories

import (
	"fmt"
	"time"

	"terralist/internal/server/models/apikey"
	"terralist/pkg/database"

	"github.com/google/uuid"
)

// StandaloneApiKeyRepository describes a service that can interact with the standalone API keys database.
type StandaloneApiKeyRepository interface {
	// Find searches for a specific ApiKey by its ID.
	Find(id uuid.UUID) (*apikey.ApiKey, error)

	// FindWithPolicies searches for a specific ApiKey and eagerly loads its policies.
	FindWithPolicies(id uuid.UUID) (*apikey.ApiKey, error)

	// Create creates a new ApiKey along with its policies in a single transaction.
	Create(key *apikey.ApiKey) (*apikey.ApiKey, error)

	// Delete removes an ApiKey and its associated policies from the database.
	Delete(id uuid.UUID) error

	// List returns all non-expired ApiKeys.
	List() ([]apikey.ApiKey, error)
}

// DefaultStandaloneApiKeyRepository is a concrete implementation of StandaloneApiKeyRepository.
type DefaultStandaloneApiKeyRepository struct {
	Database database.Engine
}

func (r *DefaultStandaloneApiKeyRepository) Find(id uuid.UUID) (*apikey.ApiKey, error) {
	key := &apikey.ApiKey{}

	if err := r.Database.Handler().
		Where("id = ?", id).
		First(key).
		Error; err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	if key.Expiration != nil && time.Now().Unix() > key.Expiration.Unix() {
		r.Database.Handler().Delete(key)
		return nil, fmt.Errorf("%w", ErrApiKeyExpired)
	}

	return key, nil
}

func (r *DefaultStandaloneApiKeyRepository) FindWithPolicies(id uuid.UUID) (*apikey.ApiKey, error) {
	key := &apikey.ApiKey{}

	if err := r.Database.Handler().
		Preload("Policies").
		Where("id = ?", id).
		First(key).
		Error; err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	if key.Expiration != nil && time.Now().Unix() > key.Expiration.Unix() {
		r.Database.Handler().Delete(key)
		return nil, fmt.Errorf("%w", ErrApiKeyExpired)
	}

	return key, nil
}

func (r *DefaultStandaloneApiKeyRepository) Create(key *apikey.ApiKey) (*apikey.ApiKey, error) {
	if err := r.Database.Handler().Create(key).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	return key, nil
}

func (r *DefaultStandaloneApiKeyRepository) Delete(id uuid.UUID) error {
	if err := r.Database.Handler().
		Where("id = ?", id).
		Delete(&apikey.ApiKey{}).
		Error; err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	return nil
}

func (r *DefaultStandaloneApiKeyRepository) List() ([]apikey.ApiKey, error) {
	var keys []apikey.ApiKey

	if err := r.Database.Handler().
		Preload("Policies").
		Find(&keys).
		Error; err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseFailure, err)
	}

	return keys, nil
}
