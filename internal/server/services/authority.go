package services

import (
	"errors"
	"terralist/internal/server/models/authority"
	"terralist/internal/server/repositories"

	"github.com/google/uuid"
	"github.com/ssoroka/slice"
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

// AuthorityService describes a service that can interact with the authorities database.
type AuthorityService interface {
	// Get returns an authority with a specific ID.
	GetByID(id uuid.UUID) (*authority.Authority, error)

	// Get returns an authority with a specific name.
	GetByName(name string) (*authority.Authority, error)

	// GetAll returns all authorities.
	GetAll() ([]*authority.Authority, error)

	// GetAllByOwner returns all authorities for a given owner.
	GetAllByOwner(owner string) ([]*authority.Authority, error)

	// Create creates a new authority.
	Create(authority.AuthorityCreateDTO) (*authority.AuthorityDTO, error)

	// Update updates an existing authority.
	Update(uuid.UUID, authority.AuthorityDTO) (*authority.AuthorityDTO, error)

	// AddKey adds a new key to an existing authority.
	AddKey(uuid.UUID, authority.KeyDTO) (*authority.KeyDTO, error)

	// RemoveKey removes an existing key from an existing authority.
	// If no keys are left, the entire authority is removed.
	RemoveKey(uuid.UUID, uuid.UUID) error

	// Delete removes an existing authority.
	Delete(id uuid.UUID) error
}

// DefaultAuthorityService is a concrete implementation of AuthorityService.
type DefaultAuthorityService struct {
	AuthorityRepository repositories.AuthorityRepository
}

func (s *DefaultAuthorityService) GetByID(id uuid.UUID) (*authority.Authority, error) {
	return s.AuthorityRepository.FindByID(id)
}

func (s *DefaultAuthorityService) GetByName(name string) (*authority.Authority, error) {
	return s.AuthorityRepository.FindByName(name)
}

func (s *DefaultAuthorityService) GetAll() ([]*authority.Authority, error) {
	return s.AuthorityRepository.FindAll()
}

func (s *DefaultAuthorityService) GetAllByOwner(owner string) ([]*authority.Authority, error) {
	return s.AuthorityRepository.FindAllByOwner(owner)
}

func (s *DefaultAuthorityService) Create(in authority.AuthorityCreateDTO) (*authority.AuthorityDTO, error) {
	a := in.ToAuthority()

	created, err := s.AuthorityRepository.Upsert(a)
	if err != nil {
		return nil, err
	}

	dto := created.ToDTO()
	return &dto, nil
}

func (s *DefaultAuthorityService) Update(id uuid.UUID, in authority.AuthorityDTO) (*authority.AuthorityDTO, error) {
	a := in.ToAuthority()
	a.ID = id

	// ApiKeys are not managed from within authority API
	// With this, we make sure we don't touch them while updating the authority
	apiKeys := a.ApiKeys
	a.ApiKeys = nil

	updated, err := s.AuthorityRepository.Upsert(a)
	if err != nil {
		return nil, err
	}

	// Put back missing ApiKeys
	updated.ApiKeys = apiKeys

	dto := updated.ToDTO()
	return &dto, nil
}

func (s *DefaultAuthorityService) AddKey(authorityID uuid.UUID, in authority.KeyDTO) (*authority.KeyDTO, error) {
	a, err := s.AuthorityRepository.FindByID(authorityID)
	if err != nil {
		return nil, err
	}

	a.Keys = append(a.Keys, in.ToKey())

	updated, err := s.AuthorityRepository.Upsert(*a)
	if err != nil {
		return nil, err
	}

	// The find operation cannot fail if the upsert method passes
	updatedKey, _ := slice.Find(updated.Keys, func(key authority.Key) bool {
		return key.KeyId == in.ToKey().KeyId
	})

	dto := updatedKey.ToKeyDTO()
	return &dto, nil
}

func (s *DefaultAuthorityService) RemoveKey(authorityID uuid.UUID, keyID uuid.UUID) error {
	a, err := s.AuthorityRepository.FindByID(authorityID)
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
