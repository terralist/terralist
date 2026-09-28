package authority

import (
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/provider"
	"terralist/pkg/database"
	"terralist/pkg/database/entity"

	"github.com/samber/lo"
)

type Authority struct {
	entity.Entity

	Name      string `gorm:"not null"`
	PolicyURL string `gorm:"not null"`
	Public    bool   `gorm:"not null;default:false"`
	Owner     string `gorm:"not null;index"`

	// UpstreamHostname is the upstream registry this authority stands for.
	UpstreamHostname *string
	// UpstreamNamespace is the namespace of that registry whose artifacts the
	// authority serves; its providers are addressed through the network mirror
	// with the upstream address.
	UpstreamNamespace *string
	// UpstreamURL is where the upstream registry is reached; empty means
	// https://<UpstreamHostname>.
	UpstreamURL *string
	// UpstreamToken authenticates against the upstream registry and is sealed
	// at rest.
	UpstreamToken       *string `gorm:"type:text"`
	UpstreamTokenSealed bool    `gorm:"-"`
	// UpstreamEnabled allows fetching from the upstream registry; without it
	// the identity only serves the network mirror address.
	UpstreamEnabled bool `gorm:"not null;default:false"`
	// UpstreamDefaultPolicy decides what the Rules do not decide.
	UpstreamDefaultPolicy string `gorm:"not null;default:allow"`

	Rules []Rule `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Keys      []Key               `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Modules   []module.Module     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Providers []provider.Provider `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Authority) TableName() string {
	return "authorities"
}

// UniqueIndexes hold the name of an authority and the upstream namespace it
// stands for unique, regardless of case, as they are looked up.
var UniqueIndexes = []database.CaseInsensitiveUniqueIndex{
	{Table: "authorities", Name: "idx_authorities_lower_name", Columns: []string{"name"}},
	{Table: "authorities", Name: "idx_authorities_lower_upstream", Columns: []string{"upstream_hostname", "upstream_namespace"}},
}

// AllowsUpstream reports whether a version of an upstream artifact may be
// served through this authority: the upstream must be enabled, no deny rule
// may match, and either the default policy allows or an allow rule matches.
func (a Authority) AllowsUpstream(kind, name, version string) bool {
	if !a.UpstreamEnabled {
		return false
	}

	allowed := a.UpstreamDefaultPolicy != PolicyDeny
	for _, rule := range a.Rules {
		if !rule.Matches(kind, name, version) {
			continue
		}

		if rule.Effect == EffectDeny {
			return false
		}

		allowed = true
	}

	return allowed
}

// UpstreamBaseURL returns the URL the upstream registry is reached at.
func (a Authority) UpstreamBaseURL() string {
	if url := lo.FromPtr(a.UpstreamURL); url != "" {
		return url
	}

	return "https://" + lo.FromPtr(a.UpstreamHostname)
}

type AuthorityDTO struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	PolicyURL         string    `json:"policy_url"`
	Public            bool      `json:"public"`
	UpstreamHostname  string    `json:"upstream_hostname"`
	UpstreamNamespace string    `json:"upstream_namespace"`
	UpstreamURL       string    `json:"upstream_url"`
	UpstreamToken     string    `json:"upstream_token,omitempty"`
	UpstreamHasToken  bool      `json:"upstream_has_token"`
	UpstreamEnabled   bool      `json:"upstream_enabled"`
	UpstreamPolicy    string    `json:"upstream_default_policy"`
	Rules             []RuleDTO `json:"rules"`
	Keys              []KeyDTO  `json:"keys"`
}

func (a Authority) ToDTO() AuthorityDTO {
	return AuthorityDTO{
		ID:                a.ID.String(),
		Name:              a.Name,
		PolicyURL:         a.PolicyURL,
		Public:            a.Public,
		UpstreamHostname:  lo.FromPtr(a.UpstreamHostname),
		UpstreamNamespace: lo.FromPtr(a.UpstreamNamespace),
		UpstreamURL:       lo.FromPtr(a.UpstreamURL),
		UpstreamHasToken:  a.UpstreamToken != nil,
		UpstreamEnabled:   a.UpstreamEnabled,
		UpstreamPolicy:    a.UpstreamDefaultPolicy,

		Rules: lo.Map(a.Rules, func(r Rule, _ int) RuleDTO {
			return r.ToDTO()
		}),

		Keys: lo.Map(a.Keys, func(k Key, _ int) KeyDTO {
			return k.ToKeyDTO()
		}),
	}
}

func (d AuthorityDTO) ToAuthority() Authority {
	return Authority{
		Name:                  d.Name,
		PolicyURL:             d.PolicyURL,
		Public:                d.Public,
		UpstreamHostname:      lo.EmptyableToPtr(d.UpstreamHostname),
		UpstreamNamespace:     lo.EmptyableToPtr(d.UpstreamNamespace),
		UpstreamURL:           lo.EmptyableToPtr(d.UpstreamURL),
		UpstreamToken:         lo.EmptyableToPtr(d.UpstreamToken),
		UpstreamEnabled:       d.UpstreamEnabled,
		UpstreamDefaultPolicy: d.UpstreamPolicy,

		Keys: lo.Map(d.Keys, func(k KeyDTO, _ int) Key {
			return k.ToKey()
		}),
	}
}

type AuthorityCreateDTO struct {
	Name              string `json:"name"`
	PolicyURL         string `json:"policy_url"`
	Public            bool   `json:"public"`
	Owner             string `json:"owner"`
	UpstreamHostname  string `json:"upstream_hostname"`
	UpstreamNamespace string `json:"upstream_namespace"`
	UpstreamURL       string `json:"upstream_url"`
	UpstreamToken     string `json:"upstream_token"`
	UpstreamEnabled   bool   `json:"upstream_enabled"`
	UpstreamPolicy    string `json:"upstream_default_policy"`
}

func (d AuthorityCreateDTO) ToAuthority() Authority {
	return Authority{
		Name:                  d.Name,
		PolicyURL:             d.PolicyURL,
		Public:                d.Public,
		Owner:                 d.Owner,
		UpstreamHostname:      lo.EmptyableToPtr(d.UpstreamHostname),
		UpstreamNamespace:     lo.EmptyableToPtr(d.UpstreamNamespace),
		UpstreamURL:           lo.EmptyableToPtr(d.UpstreamURL),
		UpstreamToken:         lo.EmptyableToPtr(d.UpstreamToken),
		UpstreamEnabled:       d.UpstreamEnabled,
		UpstreamDefaultPolicy: d.UpstreamPolicy,
	}
}
