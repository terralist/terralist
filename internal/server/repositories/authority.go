package repositories

import (
	"errors"
	"fmt"

	"terralist/internal/server/models/authority"
	"terralist/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthorityRepository describes a service that can interact with the authority database
type AuthorityRepository interface {
	// Find searches for a specific authority
	Find(uuid.UUID) (*authority.Authority, error)

	// FindAll searches for all authorities created by a specific owner
	FindAll(owner string) ([]*authority.Authority, error)

	// Upsert either updates or creates a new (if it does not already exist) authority
	Upsert(authority.Authority) (*authority.Authority, error)

	// Delete removes an authority with all its data (api keys, providers)
	Delete(uuid.UUID) error
}

// DefaultAuthorityRepository is a concrete implementation of AuthorityRepository
type DefaultAuthorityRepository struct {
	Database database.Engine
}

func (r *DefaultAuthorityRepository) Find(id uuid.UUID) (*authority.Authority, error) {
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

func (r *DefaultAuthorityRepository) FindAll(owner string) ([]*authority.Authority, error) {
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
		a := a
		asp = append(asp, &a)
	}

	return asp, nil
}

func (r *DefaultAuthorityRepository) Upsert(a authority.Authority) (*authority.Authority, error) {
	if !a.Empty() {
		current, err := r.Find(a.ID)
		if err == nil {
			if a.PolicyURL == "" {
				a.PolicyURL = current.PolicyURL
			}

			a.Name = current.Name
			a.Owner = current.Owner

			for _, currentKey := range current.Keys {
				shouldUpdate := false
				for i, newKey := range a.Keys {
					if currentKey.KeyId == newKey.KeyId {
						shouldUpdate = true
						a.Keys[i].ID = currentKey.ID
						break
					}
				}

				if !shouldUpdate {
					a.Keys = append(a.Keys, currentKey)
				}
			}
		}
	}

	if err := r.Database.Handler().Save(&a).Error; err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *DefaultAuthorityRepository) Delete(id uuid.UUID) error {
	a, err := r.Find(id)
	if err != nil {
		return err
	}

	if err := r.Database.Handler().Delete(a).Error; err != nil {
		return err
	}

	return nil
}
