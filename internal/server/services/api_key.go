package services

import (
	"errors"
	"fmt"
	"terralist/internal/server/models/authority"
	"terralist/internal/server/repositories"
	"terralist/pkg/auth"
	"terralist/pkg/metrics"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCannotParseID = errors.New("cannot parse")
	ErrInvalidKey    = errors.New("invalid key")
)

// ApiKeyService describes a service that can interact with the API keys database.
type ApiKeyService interface {
	// GetUserDetails checks if a given key is granted and returns the owner of
	// the key; if the key is invalid, it will return an error.
	GetUserDetails(key string) (*auth.User, error)

	// Grant allocates a new key; It takes an input argument which can control the
	// duration of the key. If you don't want your key to expire, set the argument
	// to 0.
	Grant(authorityID uuid.UUID, name string, expireIn int) (string, error)

	// Revoke removes a key from the database.
	Revoke(key string) error
}

// DefaultApiKeyService is a concrete implementation of ApiKeyService.
type DefaultApiKeyService struct {
	AuthorityService AuthorityService
	ApiKeyRepository repositories.ApiKeyRepository
}

func (s *DefaultApiKeyService) GetUserDetails(key string) (*auth.User, error) {
	id, err := uuid.Parse(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCannotParseID, err)
	}

	apiKey, err := s.ApiKeyRepository.Find(id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidKey, err)
	}

	authority, err := s.AuthorityService.GetByID(apiKey.AuthorityID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidKey, err)
	}

	return &auth.User{
		Email:       authority.Owner,
		Authority:   authority.Name,
		AuthorityID: apiKey.AuthorityID.String(),
	}, nil
}

func (s *DefaultApiKeyService) Grant(authorityID uuid.UUID, name string, expireIn int) (string, error) {
	apiKey := &authority.ApiKey{
		AuthorityID: authorityID,
		Name:        name,
	}

	if expireIn > 0 {
		exp := time.Now().Add(time.Duration(expireIn) * time.Hour)
		apiKey.Expiration = &exp
	}

	apiKey, err := s.ApiKeyRepository.Create(apiKey)
	if err != nil {
		return "", err
	}

	// Update metrics after creating API key
	s.updateApiKeysMetrics(authorityID)

	return apiKey.ID.String(), nil
}

func (s *DefaultApiKeyService) Revoke(key string) error {
	id, err := uuid.Parse(key)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCannotParseID, err)
	}

	// Get the API key before deleting to update metrics
	apiKey, err := s.ApiKeyRepository.Find(id)
	if err != nil {
		return err
	}

	err = s.ApiKeyRepository.Delete(id)
	if err != nil {
		return err
	}

	// Update metrics after revoking API key
	s.updateApiKeysMetrics(apiKey.AuthorityID)

	return nil
}

// updateApiKeysMetrics updates the API keys metrics for a specific authority.
func (s *DefaultApiKeyService) updateApiKeysMetrics(authorityID uuid.UUID) {
	authority, err := s.AuthorityService.GetByID(authorityID)
	if err != nil {
		return
	}

	now := time.Now()
	activeCount := 0
	expiredCount := 0

	for _, apiKey := range authority.ApiKeys {
		if apiKey.Expiration == nil || apiKey.Expiration.After(now) {
			activeCount++
		} else {
			expiredCount++
		}
	}

	metrics.SetApiKeysCount(authority.Name, "active", float64(activeCount))
	metrics.SetApiKeysCount(authority.Name, "expired", float64(expiredCount))
}
