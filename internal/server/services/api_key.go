package services

import (
	"errors"
	"fmt"
	"time"

	authModel "terralist/internal/server/models/auth"
	"terralist/internal/server/repositories"
	"terralist/pkg/auth"

	"github.com/google/uuid"
)

var (
	ErrCannotParseID = errors.New("cannot parse")
	ErrInvalidKey    = errors.New("invalid key")
)

// ApiKeyService describes a service that can interact with the API keys database
type ApiKeyService interface {
	// GetUserDetails checks if a given key is granted and returns the owner of
	// the key; if the key is invalid, it will return an error
	GetUserDetails(key string) (*auth.User, error)

	// Grant allocates a new key; It takes an input argument which can control the
	// duration of the key. If you don't want your key to expire, set the argument
	// to 0.
	Grant(authorityID uuid.UUID, expireIn int) (string, error)

	// Revoke removes a key from the database
	Revoke(key string) error
}

// DefaultApiKeyService is a concrete implementation of ApiKeyService
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
		AuthorityID: apiKey.AuthorityID.String(),
	}, nil
}

func (s *DefaultApiKeyService) Grant(authorityID uuid.UUID, expireIn int) (string, error) {
	apiKey := &authModel.ApiKey{
		AuthorityID: authorityID,
	}

	if expireIn > 0 {
		exp := time.Now().Add(time.Duration(expireIn) * time.Hour)
		apiKey.Expiration = &exp
	}

	apiKey, err := s.ApiKeyRepository.Create(apiKey)
	if err != nil {
		return "", err
	}

	return apiKey.ID.String(), nil
}

func (s *DefaultApiKeyService) Revoke(key string) error {
	id, err := uuid.Parse(key)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCannotParseID, err)
	}

	return s.ApiKeyRepository.Delete(id)
}
