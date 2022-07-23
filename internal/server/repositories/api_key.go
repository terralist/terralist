package repositories

import (
	"fmt"
	"github.com/google/uuid"
	"terralist/internal/server/models/authority"
	"terralist/pkg/database"
	"time"
)

// ApiKeyRepository describes a service that can interact with the API keys database
type ApiKeyRepository interface {
	// Exists checks if a given key is granted
	Exists(key string) bool

	// Grant allocates a new key; It takes an input argument which can control the
	// duration of the key. If you don't want your key to expire, set the argument
	// to 0.
	Grant(expireIn int) (string, error)

	// Revoke removes a key from the database
	Revoke(key string) error
}

// DefaultApiKeyRepository is a concrete implementation of ApiKeyRepository
type DefaultApiKeyRepository struct {
	Database database.Engine
}

func (r *DefaultApiKeyRepository) Exists(key string) bool {
	id, _ := uuid.Parse(key)

	apiKey := &authority.ApiKey{}

	if err := r.Database.Handler().
		Where("id = ?", id).
		First(apiKey).
		Error; err != nil {
		return false
	}

	if time.Now().Unix() > apiKey.Expiration.Unix() {
		r.Database.Handler().Delete(apiKey)
		return false
	}

	return true
}

func (r *DefaultApiKeyRepository) Grant(expireIn int) (*authority.ApiKey, error) {
	apiKey := &authority.ApiKey{}

	if expireIn != 0 {
		apiKey.Expiration = time.Now().Add(time.Duration(expireIn) * time.Hour)
	}

	if err := r.Database.Handler().Create(apiKey).Error; err != nil {
		return nil, fmt.Errorf("could not create api key: %v", err)
	}

	return apiKey, nil
}

func (r *DefaultApiKeyRepository) Revoke(key string) error {
	id, _ := uuid.Parse(key)

	err := r.Database.Handler().
		Where("id = ?", id).
		Delete(&authority.ApiKey{}).
		Error

	if err != nil {
		return fmt.Errorf("could not revoke key: %v", err)
	}

	return nil
}
