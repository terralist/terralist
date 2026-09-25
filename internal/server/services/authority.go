package services

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/repositories"
	"terralist/pkg/secret"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

var (
	ErrKeyNotFound  = errors.New("key not found")
	ErrRuleNotFound = errors.New("rule not found")

	// upstreamHostnameRegexp accepts DNS hostnames with an optional port.
	upstreamHostnameRegexp = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*(:[0-9]{1,5})?$`)

	// upstreamNamespaceRegexp accepts provider registry namespaces.
	upstreamNamespaceRegexp = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
)

// AuthorityService describes a service that can interact with the authorities database.
type AuthorityService interface {
	// Get returns an authority with a specific ID.
	GetByID(id uuid.UUID) (*authority.Authority, error)

	// Get returns an authority with a specific name.
	GetByName(name string) (*authority.Authority, error)

	// GetByUpstream returns the authority standing for an upstream registry
	// hostname and namespace.
	GetByUpstream(hostname, namespace string) (*authority.Authority, error)

	// GetAll returns all authorities.
	GetAll() ([]*authority.Authority, error)

	// GetAllByOwner returns all authorities for a given owner.
	GetAllByOwner(owner string) ([]*authority.Authority, error)

	// Create creates a new authority.
	Create(authority.AuthorityCreateDTO) (*authority.AuthorityDTO, error)

	// Update updates an existing authority.
	Update(uuid.UUID, authority.AuthorityDTO) (*authority.AuthorityDTO, error)

	// AddKey adds a new key to an existing authority.
	AddKey(uuid.UUID, authority.KeyDTO) (*authority.KeyDTO, error)

	// RemoveKey removes an existing key from an existing authority.
	// If no keys are left, the entire authority is removed.
	RemoveKey(uuid.UUID, uuid.UUID) error

	AddRule(uuid.UUID, authority.RuleDTO) (*authority.RuleDTO, error)

	RemoveRule(uuid.UUID, uuid.UUID) error

	// Delete removes an existing authority.
	Delete(id uuid.UUID) error
}

// DefaultAuthorityService is a concrete implementation of AuthorityService.
type DefaultAuthorityService struct {
	AuthorityRepository repositories.AuthorityRepository

	// Sealer protects upstream tokens at rest. Without it, no upstream token
	// can be stored.
	Sealer *secret.Sealer
}

func (s *DefaultAuthorityService) GetByID(id uuid.UUID) (*authority.Authority, error) {
	return s.AuthorityRepository.FindByID(id)
}

func (s *DefaultAuthorityService) GetByName(name string) (*authority.Authority, error) {
	return s.AuthorityRepository.FindByName(name)
}

func (s *DefaultAuthorityService) GetByUpstream(hostname, namespace string) (*authority.Authority, error) {
	return s.AuthorityRepository.FindByUpstream(hostname, namespace)
}

func (s *DefaultAuthorityService) GetAll() ([]*authority.Authority, error) {
	return s.AuthorityRepository.FindAll()
}

func (s *DefaultAuthorityService) GetAllByOwner(owner string) ([]*authority.Authority, error) {
	return s.AuthorityRepository.FindAllByOwner(owner)
}

func (s *DefaultAuthorityService) Create(in authority.AuthorityCreateDTO) (*authority.AuthorityDTO, error) {
	a := in.ToAuthority()

	if err := s.normalizeUpstream(&a); err != nil {
		return nil, err
	}

	created, err := s.AuthorityRepository.Upsert(a)
	if err != nil {
		return nil, err
	}

	dto := created.ToDTO()
	return &dto, nil
}

func (s *DefaultAuthorityService) Update(id uuid.UUID, in authority.AuthorityDTO) (*authority.AuthorityDTO, error) {
	a := in.ToAuthority()
	a.ID = id

	// A token is never sent back to clients, so an update without one keeps
	// the token already stored.
	if a.UpstreamToken == nil {
		if current, err := s.AuthorityRepository.FindByID(id); err == nil && current.UpstreamToken != nil {
			a.UpstreamToken = current.UpstreamToken
			a.UpstreamTokenSealed = true
		}
	}

	if err := s.normalizeUpstream(&a); err != nil {
		return nil, err
	}

	// ApiKeys are not managed from within authority API
	// With this, we make sure we don't touch them while updating the authority
	apiKeys := a.ApiKeys
	a.ApiKeys = nil

	updated, err := s.AuthorityRepository.Upsert(a)
	if err != nil {
		return nil, err
	}

	// Put back missing ApiKeys
	updated.ApiKeys = apiKeys

	dto := updated.ToDTO()
	return &dto, nil
}

func (s *DefaultAuthorityService) AddKey(authorityID uuid.UUID, in authority.KeyDTO) (*authority.KeyDTO, error) {
	a, err := s.AuthorityRepository.FindByID(authorityID)
	if err != nil {
		return nil, err
	}

	a.Keys = append(a.Keys, in.ToKey())

	updated, err := s.AuthorityRepository.Upsert(*a)
	if err != nil {
		return nil, err
	}

	// The find operation cannot fail if the upsert method passes
	updatedKey, _ := lo.Find(updated.Keys, func(key authority.Key) bool {
		return key.KeyId == in.ToKey().KeyId
	})

	dto := updatedKey.ToKeyDTO()
	return &dto, nil
}

func (s *DefaultAuthorityService) RemoveKey(authorityID uuid.UUID, keyID uuid.UUID) error {
	a, err := s.AuthorityRepository.FindByID(authorityID)
	if err != nil {
		return err
	}

	l := len(a.Keys)
	for i, key := range a.Keys {
		if key.ID == keyID {
			a.Keys = append(a.Keys[:i], a.Keys[i+1:]...)
			break
		}
	}

	// If no key was deleted
	if l == len(a.Keys) {
		return ErrKeyNotFound
	}

	// If there was only 1 key
	if l == 1 {
		return s.AuthorityRepository.Delete(authorityID)
	}

	_, err = s.AuthorityRepository.Upsert(*a)
	return err
}

func (s *DefaultAuthorityService) AddRule(authorityID uuid.UUID, in authority.RuleDTO) (*authority.RuleDTO, error) {
	rule := in.ToRule()
	if err := rule.Validate(); err != nil {
		return nil, err
	}

	a, err := s.AuthorityRepository.FindByID(authorityID)
	if err != nil {
		return nil, err
	}

	a.Rules = append(a.Rules, rule)

	updated, err := s.AuthorityRepository.Upsert(*a)
	if err != nil {
		return nil, err
	}

	dto := updated.Rules[len(updated.Rules)-1].ToDTO()
	return &dto, nil
}

func (s *DefaultAuthorityService) RemoveRule(authorityID uuid.UUID, ruleID uuid.UUID) error {
	a, err := s.AuthorityRepository.FindByID(authorityID)
	if err != nil {
		return err
	}

	if !lo.ContainsBy(a.Rules, func(r authority.Rule) bool { return r.ID == ruleID }) {
		return ErrRuleNotFound
	}

	return s.AuthorityRepository.DeleteRule(ruleID)
}

func (s *DefaultAuthorityService) Delete(id uuid.UUID) error {
	return s.AuthorityRepository.Delete(id)
}

// normalizeUpstream validates the upstream settings of an authority: it
// lowercases the hostname, defaults the namespace to the authority name and
// the policy to allow, trims the URL and seals the token.
func (s *DefaultAuthorityService) normalizeUpstream(a *authority.Authority) error {
	if a.UpstreamDefaultPolicy == "" {
		a.UpstreamDefaultPolicy = authority.PolicyAllow
	}

	if a.UpstreamDefaultPolicy != authority.PolicyAllow && a.UpstreamDefaultPolicy != authority.PolicyDeny {
		return fmt.Errorf("invalid upstream default policy %q, expected %s or %s", a.UpstreamDefaultPolicy, authority.PolicyAllow, authority.PolicyDeny)
	}

	if a.UpstreamHostname == nil {
		if a.UpstreamNamespace != nil {
			return fmt.Errorf("an upstream namespace requires an upstream hostname")
		}

		if a.UpstreamEnabled {
			return fmt.Errorf("enabling the upstream requires an upstream hostname")
		}

		a.UpstreamURL = nil
		a.UpstreamToken = nil

		return nil
	}

	hostname, err := NormalizeUpstreamHostname(*a.UpstreamHostname)
	if err != nil {
		return err
	}

	a.UpstreamHostname = &hostname

	if a.UpstreamNamespace == nil {
		a.UpstreamNamespace = &a.Name
	}

	if !upstreamNamespaceRegexp.MatchString(*a.UpstreamNamespace) {
		return fmt.Errorf("invalid upstream namespace %q", *a.UpstreamNamespace)
	}

	if a.UpstreamURL != nil {
		parsed, err := url.Parse(*a.UpstreamURL)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
			return fmt.Errorf("invalid upstream URL %q, expected an http or https URL", *a.UpstreamURL)
		}

		trimmed := strings.TrimRight(*a.UpstreamURL, "/")
		a.UpstreamURL = &trimmed
	}

	if a.UpstreamToken != nil && !a.UpstreamTokenSealed {
		if s.Sealer == nil {
			return fmt.Errorf("storing an upstream token requires the upstream-secret option")
		}

		sealed, err := s.Sealer.Seal(*a.UpstreamToken)
		if err != nil {
			return fmt.Errorf("could not seal the upstream token: %w", err)
		}

		a.UpstreamToken = &sealed
		a.UpstreamTokenSealed = true
	}

	return nil
}

// NormalizeUpstreamHostname lowercases an upstream registry hostname and
// rejects anything that is not a DNS hostname with an optional port.
func NormalizeUpstreamHostname(hostname string) (string, error) {
	normalized := strings.ToLower(hostname)
	if !upstreamHostnameRegexp.MatchString(normalized) {
		return "", fmt.Errorf("invalid upstream hostname %q", hostname)
	}

	return normalized, nil
}
