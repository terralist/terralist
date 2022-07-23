package repositories

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"terralist/internal/server/models/authority"
	"terralist/pkg/database"
)

type AuthorityRepository interface {
	Find(uuid.UUID) (*authority.Authority, error)
	Upsert(authority.Authority) (*authority.Authority, error)
	Delete(uuid.UUID) error
}

type DefaultAuthorityRepository struct {
	Database database.Engine
}

func (r *DefaultAuthorityRepository) Find(id uuid.UUID) (*authority.Authority, error) {
	a := &authority.Authority{}

	err := r.Database.Handler().
		Where("id = ?", id).
		Preload("Keys").
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

func (r *DefaultAuthorityRepository) Upsert(a authority.Authority) (*authority.Authority, error) {
	if !a.Empty() {
		current, err := r.Find(a.ID)
		if err == nil {
			if a.PolicyURL == "" {
				a.PolicyURL = current.PolicyURL
			}
			a.Name = current.Name

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
