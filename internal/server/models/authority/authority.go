package authority

import (
	"terralist/internal/server/models/module"
	"terralist/internal/server/models/provider"
	"terralist/pkg/database/entity"

	"github.com/samber/lo"
)

type Authority struct {
	entity.Entity

	Name      string `gorm:"not null;uniqueIndex"`
	PolicyURL string `gorm:"not null"`
	Public    bool   `gorm:"not null;default:false"`
	Owner     string `gorm:"not null;index"`

	// UpstreamHostname represents the upstream registry that this authority stands for.
	UpstreamHostname *string `gorm:"uniqueIndex:idx_authorities_upstream"`
	// UpstreamNamespace represents the upstream registry's namespace that this authority stands for.
	// It's providers can be addressed through the network mirror.
	UpstreamNamespace *string `gorm:"uniqueIndex:idx_authorities_upstream"`
	// UpstreamURL is where the upstream registry is reached; empty means https://<UpstreamHostname>.
	UpstreamURL *string
	// UpstreamToken authenticates against it and is sealed at rest.
	UpstreamToken       *string
	UpstreamTokenSealed bool `gorm:"-"`
	// UpstreamEnabled decides if the upstream is enabled. Nothing is fetched unless is set.
	UpstreamEnabled bool `gorm:"not null;default:false"`
	// UpstreamDefaultPolicy decides what the Rules do not decide.
	UpstreamDefaultPolicy string `gorm:"not null;default:allow"`

	Rules []Rule `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Keys      []Key               `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ApiKeys   []ApiKey            `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Modules   []module.Module     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Providers []provider.Provider `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (Authority) TableName() string {
	return "authorities"
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
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	PolicyURL         string      `json:"policy_url"`
	Public            bool        `json:"public"`
	UpstreamHostname  string      `json:"upstream_hostname"`
	UpstreamNamespace string      `json:"upstream_namespace"`
	UpstreamURL       string      `json:"upstream_url"`
	UpstreamToken     string      `json:"upstream_token,omitempty"`
	UpstreamHasToken  bool        `json:"upstream_has_token"`
	UpstreamEnabled   bool        `json:"upstream_enabled"`
	UpstreamPolicy    string      `json:"upstream_default_policy"`
	Rules             []RuleDTO   `json:"rules"`
	Keys              []KeyDTO    `json:"keys"`
	ApiKeys           []ApiKeyDTO `json:"api_keys"`
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

		ApiKeys: lo.Map(a.ApiKeys, func(a ApiKey, _ int) ApiKeyDTO {
			return a.ToDTO()
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

		ApiKeys: lo.Map(d.ApiKeys, func(a ApiKeyDTO, _ int) ApiKey {
			return a.ToApiKey()
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
