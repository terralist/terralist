package authority

import "terralist/pkg/database/entity"

type Authority struct {
	entity.Entity

	Name      string `gorm:"not null"`
	PolicyURL string `gorm:"not null"`
	Keys      []Key  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Authority) TableName() string {
	return "authority"
}

func (a Authority) ToAuthorityKeysDTO() AuthorityKeysDTO {
	var keys []AuthorityKeyDTO
	for _, k := range a.Keys {
		keys = append(keys, AuthorityKeyDTO{
			KeyDTO:    k.ToKeyDTO(),
			Source:    a.Name,
			SourceURL: a.PolicyURL,
		})
	}

	return AuthorityKeysDTO{
		Keys: keys,
	}
}

type AuthorityKeysDTO struct {
	Keys []AuthorityKeyDTO `json:"gpg_public_keys"`
}

type AuthorityKeyDTO struct {
	KeyDTO
	Source    string `json:"string"`
	SourceURL string `json:"source_url"`
}
