package authority

import (
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/provider"
	"terralist/pkg/database/entity"
)

type Authority struct {
	entity.Entity

	Name      string              `gorm:"not null;uniqueIndex"`
	PolicyURL string              `gorm:"not null"`
	Owner     string              `gorm:"not null;uniqueIndex"`
	Keys      []Key               `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ApiKeys   []ApiKey            `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Modules   []module.Module     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Providers []provider.Provider `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Authority) TableName() string {
	return "authorities"
}

type AuthorityCreateDTO struct {
	Name      string `json:"name"`
	PolicyURL string `json:"policy_url"`
}

func (d AuthorityCreateDTO) ToAuthority() Authority {
	return Authority{
		Name:      d.Name,
		PolicyURL: d.PolicyURL,
	}
}
