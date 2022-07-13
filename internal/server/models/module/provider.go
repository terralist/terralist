package module

import (
	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

type Provider struct {
	entity.Entity
	ParentID  uuid.UUID
	Name      string `gorm:"not null"`
	Namespace string `gorm:"not null"`
	Source    string `gorm:"not null"`
	Version   string `gorm:"not null"`
}

func (Provider) TableName() string {
	return "module_providers"
}

type ProviderDTO struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Source    string `json:"source"`
	Version   string `json:"version"`
}

func (d ProviderDTO) ToProvider() Provider {
	return Provider{
		Name:      d.Name,
		Namespace: d.Namespace,
		Source:    d.Source,
		Version:   d.Version,
	}
}
