package services

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"terralist/internal/server/models/apikey"
	"terralist/internal/server/repositories"
	"terralist/pkg/auth"
	"terralist/pkg/rbac"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

var ErrInvalidPolicy = errors.New("invalid policy")

// StandaloneApiKeyService describes a service that manages standalone API keys with RBAC policies.
type StandaloneApiKeyService interface {
	// Authenticate validates an API key and returns the associated user and policies.
	Authenticate(key string) (*auth.User, []apikey.Policy, error)

	// Create creates a new API key with the given policies.
	Create(name, createdBy string, expireIn int, policies []apikey.Policy) (string, error)

	// Delete removes an API key.
	Delete(key string) error

	// List returns all API keys with their policies.
	List() ([]apikey.ApiKeyDTO, error)
}

// DefaultStandaloneApiKeyService is a concrete implementation of StandaloneApiKeyService.
type DefaultStandaloneApiKeyService struct {
	Repository repositories.StandaloneApiKeyRepository
}

func (s *DefaultStandaloneApiKeyService) Authenticate(key string) (*auth.User, []apikey.Policy, error) {
	id, err := uuid.Parse(key)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrCannotParseID, err)
	}

	k, err := s.Repository.FindWithPolicies(id)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidKey, err)
	}

	user := &auth.User{
		Name:  fmt.Sprintf("apikey:%s", k.ID.String()),
		Email: k.CreatedBy,
	}

	return user, k.Policies, nil
}

func (s *DefaultStandaloneApiKeyService) Create(name, createdBy string, expireIn int, policies []apikey.Policy) (string, error) {
	if err := validatePolicies(policies); err != nil {
		return "", err
	}

	key := &apikey.ApiKey{
		Name:      name,
		CreatedBy: createdBy,
		Policies:  policies,
	}

	if expireIn > 0 {
		exp := time.Now().Add(time.Duration(expireIn) * time.Hour)
		key.Expiration = &exp
	}

	key, err := s.Repository.Create(key)
	if err != nil {
		return "", err
	}

	return key.ID.String(), nil
}

func (s *DefaultStandaloneApiKeyService) Delete(key string) error {
	id, err := uuid.Parse(key)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCannotParseID, err)
	}

	return s.Repository.Delete(id)
}

func (s *DefaultStandaloneApiKeyService) List() ([]apikey.ApiKeyDTO, error) {
	keys, err := s.Repository.List()
	if err != nil {
		return nil, err
	}

	return lo.Map(keys, func(k apikey.ApiKey, _ int) apikey.ApiKeyDTO {
		return k.ToDTO()
	}), nil
}

func validatePolicies(policies []apikey.Policy) error {
	for i, p := range policies {
		if !slices.Contains(rbac.Resources, p.Resource) && p.Resource != "*" {
			return fmt.Errorf("%w: policy %d has invalid resource %q", ErrInvalidPolicy, i, p.Resource)
		}

		if !slices.Contains(rbac.Actions, p.Action) && p.Action != "*" {
			return fmt.Errorf("%w: policy %d has invalid action %q", ErrInvalidPolicy, i, p.Action)
		}

		if !slices.Contains(rbac.Effects, p.Effect) {
			return fmt.Errorf("%w: policy %d has invalid effect %q", ErrInvalidPolicy, i, p.Effect)
		}

		if p.Object == "" {
			return fmt.Errorf("%w: policy %d has empty object", ErrInvalidPolicy, i)
		}
	}

	return nil
}
