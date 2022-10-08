package services

import (
	"errors"
	"terralist/internal/server/models/authority"
	"terralist/internal/server/repositories"

	"github.com/google/uuid"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

// AuthorityService describes a service that can interact with the authorities database
type AuthorityService interface {
	// Get returns an authority with a specific ID
	Get(id uuid.UUID) (*authority.Authority, error)

	// GetAll returns all authorities for a given owner
	GetAll(owner string) ([]*authority.Authority, error)

	// Create creates a new authority
	Create(authority.AuthorityCreateDTO) error

	// AddKey adds a new key to an existing authority
	AddKey(uuid.UUID, authority.KeyDTO) error

	// RemoveKey removes an existing key from an existing authority
	// If no keys are left, the entire authority is removed
	RemoveKey(uuid.UUID, uuid.UUID) error

	// Delete removes an existing authority
	Delete(id uuid.UUID) error
}

// DefaultAuthorityService is a concrete implementation of AuthorityService
type DefaultAuthorityService struct {
	AuthorityRepository repositories.AuthorityRepository
}

func (s *DefaultAuthorityService) Get(id uuid.UUID) (*authority.Authority, error) {
	return s.AuthorityRepository.Find(id)
}

func (s *DefaultAuthorityService) GetAll(owner string) ([]*authority.Authority, error) {
	return s.AuthorityRepository.FindAll(owner)
}

func (s *DefaultAuthorityService) Create(in authority.AuthorityCreateDTO) error {
	a := in.ToAuthority()

	_, err := s.AuthorityRepository.Upsert(a)
	if err != nil {
		return err
	}

	return nil
}

func (s *DefaultAuthorityService) AddKey(authorityID uuid.UUID, in authority.KeyDTO) error {
	a, err := s.AuthorityRepository.Find(authorityID)
	if err != nil {
		return err
	}

	a.Keys = append(a.Keys, in.ToKey())

	_, err = s.AuthorityRepository.Upsert(*a)
	return err
}

func (s *DefaultAuthorityService) RemoveKey(authorityID uuid.UUID, keyID uuid.UUID) error {
	a, err := s.AuthorityRepository.Find(authorityID)
	if err != nil {
		return err
	}

	l := len(a.Keys)
	for i, key := range a.Keys {
		if key.ID == keyID {
			a.Keys = append(a.Keys[:i], a.Keys[i+1:]...)
			break
		}
	}

	// If no key was deleted
	if l == len(a.Keys) {
		return ErrKeyNotFound
	}

	// If there was only 1 key
	if l == 1 {
		return s.AuthorityRepository.Delete(authorityID)
	}

	_, err = s.AuthorityRepository.Upsert(*a)
	return err
}

func (s *DefaultAuthorityService) Delete(id uuid.UUID) error {
	return s.AuthorityRepository.Delete(id)
}
