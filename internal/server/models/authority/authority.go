package authority

import (
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/provider"
	"terralist/pkg/database/entity"

	"github.com/ssoroka/slice"
)

type Authority struct {
	entity.Entity

	Name      string              `gorm:"size:900;not null;uniqueIndex"`
	PolicyURL string              `gorm:"not null"`
	Owner     string              `gorm:"size:900;not null;index"`
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
	Keys      []KeyDTO    `json:"keys"`
	ApiKeys   []ApiKeyDTO `json:"api_keys"`
}

func (a Authority) ToDTO() AuthorityDTO {
	return AuthorityDTO{
		ID:        a.ID.String(),
		Name:      a.Name,
		PolicyURL: a.PolicyURL,

		Keys: slice.Map[Key, KeyDTO](a.Keys, func(k Key) KeyDTO {
			return k.ToKeyDTO()
		}),

		ApiKeys: slice.Map[ApiKey, ApiKeyDTO](a.ApiKeys, func(a ApiKey) ApiKeyDTO {
			return a.ToDTO()
		}),
	}
}

func (d AuthorityDTO) ToAuthority() Authority {
	return Authority{
		Name:      d.Name,
		PolicyURL: d.PolicyURL,

		Keys: slice.Map[KeyDTO, Key](d.Keys, func(k KeyDTO) Key {
			return k.ToKey()
		}),

		ApiKeys: slice.Map[ApiKeyDTO, ApiKey](d.ApiKeys, func(a ApiKeyDTO) ApiKey {
			return a.ToApiKey()
		}),
	}
}

type AuthorityCreateDTO struct {
	Name      string `json:"name"`
	PolicyURL string `json:"policy_url"`
	Owner     string `json:"owner"`
}

func (d AuthorityCreateDTO) ToAuthority() Authority {
	return Authority{
		Name:      d.Name,
		PolicyURL: d.PolicyURL,
		Owner:     d.Owner,
	}
}
