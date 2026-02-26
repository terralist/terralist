package repositories

import (
	"errors"
	"fmt"
	"slices"

	"terralist/internal/server/models/authority"
	"terralist/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthorityRepository describes a service that can interact with the authority database.
type AuthorityRepository interface {
	// Find searches for a specific authority by its ID.
	FindByID(uuid.UUID) (*authority.Authority, error)

	// Find searches for a specific authority by its name.
	FindByName(string) (*authority.Authority, error)

	// FindAll searches for all authorities.
	FindAll() ([]*authority.Authority, error)

	// FindAllByOwner searches for all authorities created by a specific owner.
	FindAllByOwner(owner string) ([]*authority.Authority, error)

	// Upsert either updates or creates a new (if it does not already exist) authority.
	Upsert(authority.Authority) (*authority.Authority, error)

	// Delete removes an authority with all its data (api keys, providers).
	Delete(uuid.UUID) error
}

// DefaultAuthorityRepository is a concrete implementation of AuthorityRepository.
type DefaultAuthorityRepository struct {
	Database database.Engine
}

func (r *DefaultAuthorityRepository) FindByID(id uuid.UUID) (*authority.Authority, error) {
	a := &authority.Authority{}

	err := r.Database.Handler().
		Where("id = ?", id).
		Preload("Keys").
		Preload("ApiKeys").
		First(&a).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no authority found")
		} else {
			return nil, fmt.Errorf("error while querying the database: %v", err)
		}
	}

	return a, nil
}

func (r *DefaultAuthorityRepository) FindByName(name string) (*authority.Authority, error) {
	a := &authority.Authority{}

	err := r.Database.Handler().
		Where("name = ?", name).
		Preload("Keys").
		Preload("ApiKeys").
		First(&a).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no authority found")
		} else {
			return nil, fmt.Errorf("error while querying the database: %v", err)
		}
	}

	return a, nil
}

func (r *DefaultAuthorityRepository) FindAll() ([]*authority.Authority, error) {
	as := []authority.Authority{}

	err := r.Database.Handler().
		Preload("Keys").
		Preload("ApiKeys").
		Preload("Modules").
		Preload("Modules.Versions").
		Preload("Providers").
		Preload("Providers.Versions").
		Find(&as).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no authority found")
		} else {
			return nil, fmt.Errorf("error while querying the database: %v", err)
		}
	}

	asp := make([]*authority.Authority, len(as))
	for i, a := range as {
		asp[i] = &a
	}

	return asp, nil
}

func (r *DefaultAuthorityRepository) FindAllByOwner(owner string) ([]*authority.Authority, error) {
	as := []authority.Authority{}

	err := r.Database.Handler().
		Where(&authority.Authority{Owner: owner}).
		Preload("Keys").
		Preload("ApiKeys").
		Find(&as).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no authority found")
		} else {
			return nil, fmt.Errorf("error while querying the database: %v", err)
		}
	}

	asp := []*authority.Authority{}
	for _, a := range as {
		asp = append(asp, &a)
	}

	return asp, nil
}

func (r *DefaultAuthorityRepository) Upsert(a authority.Authority) (*authority.Authority, error) {
	toDeleteKeys := make([]authority.Key, 0, len(a.Keys))

	if !a.Empty() {
		current, err := r.FindByID(a.ID)
		if err == nil {
			a.Name = current.Name
			a.Owner = current.Owner
		}

		for _, key := range current.Keys {
			if !slices.Contains(a.Keys, key) {
				toDeleteKeys = append(toDeleteKeys, key)
			}
		}
	}

	if err := r.Database.Handler().Transaction(func(tx *gorm.DB) error {
		if len(toDeleteKeys) > 0 {
			if err := tx.Delete(&toDeleteKeys).Error; err != nil {
				return err
			}
		}

		if err := tx.Save(&a).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *DefaultAuthorityRepository) Delete(id uuid.UUID) error {
	a, err := r.FindByID(id)
	if err != nil {
		return err
	}

	if err := r.Database.Handler().Delete(a).Error; err != nil {
		return err
	}

	return nil
}
