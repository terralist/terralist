package repositories

import (
	"errors"
	"fmt"
	"sort"

	"terralist/internal/server/models/mirror"
	"terralist/pkg/database"
	"terralist/pkg/version"

	"gorm.io/gorm"
)

// MirrorRepository describes a service that can interact with the mirrored
// providers database.
type MirrorRepository interface {
	// Find searches for a specific mirrored provider.
	Find(hostname, namespace, name string) (*mirror.Provider, error)

	// FindByNamespace returns all mirrored providers under a hostname and
	// namespace.
	FindByNamespace(hostname, namespace string) ([]mirror.Provider, error)

	// FindByHostname returns all mirrored providers under a hostname.
	FindByHostname(hostname string) ([]mirror.Provider, error)

	// Upsert either updates or creates a new (if it does not already exist)
	// mirrored provider.
	Upsert(mirror.Provider) (*mirror.Provider, error)

	// Delete removes a mirrored provider with all its data (versions).
	Delete(*mirror.Provider) error

	// DeleteVersion removes a version from a mirrored provider.
	DeleteVersion(p *mirror.Provider, version string) error
}

// DefaultMirrorRepository is a concrete implementation of MirrorRepository.
type DefaultMirrorRepository struct {
	Database database.Engine
}

func (r *DefaultMirrorRepository) Find(hostname, namespace, name string) (*mirror.Provider, error) {
	p := mirror.Provider{}

	err := r.Database.Handler().
		Where("LOWER(hostname) = LOWER(?) AND LOWER(namespace) = LOWER(?) AND LOWER(name) = LOWER(?)", hostname, namespace, name).
		Preload("Versions").
		Preload("Versions.Platforms").
		First(&p).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no mirrored provider found with given arguments (provider %s/%s/%s)", hostname, namespace, name)
		}

		return nil, fmt.Errorf("error while querying the database: %v", err)
	}

	sortVersions(&p)

	return &p, nil
}

func (r *DefaultMirrorRepository) FindByNamespace(hostname, namespace string) ([]mirror.Provider, error) {
	return r.findAll(
		"LOWER(hostname) = LOWER(?) AND LOWER(namespace) = LOWER(?)",
		hostname,
		namespace,
	)
}

func (r *DefaultMirrorRepository) FindByHostname(hostname string) ([]mirror.Provider, error) {
	return r.findAll("LOWER(hostname) = LOWER(?)", hostname)
}

func (r *DefaultMirrorRepository) findAll(query string, args ...any) ([]mirror.Provider, error) {
	var providers []mirror.Provider

	err := r.Database.Handler().
		Where(query, args...).
		Preload("Versions").
		Preload("Versions.Platforms").
		Find(&providers).
		Error

	if err != nil {
		return nil, fmt.Errorf("error while querying the database: %v", err)
	}

	for i := range providers {
		sortVersions(&providers[i])
	}

	return providers, nil
}

func (r *DefaultMirrorRepository) Upsert(p mirror.Provider) (*mirror.Provider, error) {
	if err := r.Database.Handler().Save(&p).Error; err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *DefaultMirrorRepository) Delete(p *mirror.Provider) error {
	return r.Database.Handler().Transaction(func(tx *gorm.DB) error {
		for _, v := range p.Versions {
			if err := deleteVersion(tx, &v); err != nil {
				return err
			}
		}

		return tx.Delete(p).Error
	})
}

func (r *DefaultMirrorRepository) DeleteVersion(p *mirror.Provider, v string) error {
	toDelete := p.GetVersion(v)
	if toDelete == nil {
		return fmt.Errorf("provider %s/%s/%s does not contain version %s", p.Hostname, p.Namespace, p.Name, v)
	}

	if len(p.Versions) == 1 {
		return r.Delete(p)
	}

	return r.Database.Handler().Transaction(func(tx *gorm.DB) error {
		return deleteVersion(tx, toDelete)
	})
}

// deleteVersion removes a version together with its platforms.
func deleteVersion(tx *gorm.DB, v *mirror.Version) error {
	if err := tx.Where("version_id = ?", v.ID).Delete(&mirror.Platform{}).Error; err != nil {
		return err
	}

	return tx.Delete(v).Error
}

// sortVersions orders the provider versions ascending.
func sortVersions(p *mirror.Provider) {
	sort.Slice(p.Versions, func(i, j int) bool {
		lhs := version.Version(p.Versions[i].Version)
		rhs := version.Version(p.Versions[j].Version)

		return version.Compare(lhs, rhs) < 0
	})
}
