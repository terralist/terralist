package authority

import (
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/provider"
	"terralist/pkg/database/entity"

	"github.com/samber/lo"
)

type Authority struct {
	entity.Entity

	Name      string              `gorm:"not null;uniqueIndex"`
	PolicyURL string              `gorm:"not null"`
	Public    bool                `gorm:"not null;default:false"`
	Owner     string              `gorm:"not null;index"`
	Keys      []Key               `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ApiKeys   []ApiKey            `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Modules   []module.Module     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Providers []provider.Provider `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Authority) TableName() string {
	return "authorities"
}

type AuthorityDTO struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	PolicyURL string      `json:"policy_url"`
	Public    bool        `json:"public"`
	Keys      []KeyDTO    `json:"keys"`
	ApiKeys   []ApiKeyDTO `json:"api_keys"`
}

func (a Authority) ToDTO() AuthorityDTO {
	return AuthorityDTO{
		ID:        a.ID.String(),
		Name:      a.Name,
		PolicyURL: a.PolicyURL,
		Public:    a.Public,

		Keys: lo.Map(a.Keys, func(k Key, _ int) KeyDTO {
			return k.ToKeyDTO()
		}),

		ApiKeys: lo.Map(a.ApiKeys, func(a ApiKey, _ int) ApiKeyDTO {
			return a.ToDTO()
		}),
	}
}

func (d AuthorityDTO) ToAuthority() Authority {
	return Authority{
		Name:      d.Name,
		PolicyURL: d.PolicyURL,
		Public:    d.Public,

		Keys: lo.Map(d.Keys, func(k KeyDTO, _ int) Key {
			return k.ToKey()
		}),

		ApiKeys: lo.Map(d.ApiKeys, func(a ApiKeyDTO, _ int) ApiKey {
			return a.ToApiKey()
		}),
	}
}

type AuthorityCreateDTO struct {
	Name      string `json:"name"`
	PolicyURL string `json:"policy_url"`
	Public    bool   `json:"public"`
	Owner     string `json:"owner"`
}

func (d AuthorityCreateDTO) ToAuthority() Authority {
	return Authority{
		Name:      d.Name,
		PolicyURL: d.PolicyURL,
		Public:    d.Public,
		Owner:     d.Owner,
	}
}
