package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/repositories"
	"terralist/pkg/cache"
	"terralist/pkg/metrics"
	"terralist/pkg/registry"
	"terralist/pkg/secret"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"golang.org/x/sync/singleflight"
)

var (
	// ErrUpstreamDenied is returned when the rules of an authority do not allow
	// serving a version from its upstream.
	ErrUpstreamDenied = errors.New("version not allowed by the upstream rules")

	// ErrUpstreamUnavailable is returned when the upstream registry cannot be
	// read and nothing Terralist holds answers the request.
	ErrUpstreamUnavailable = errors.New("the upstream registry is unavailable")
)

// upstreamFailure marks err as an unavailable upstream, unless it says the
// upstream does not have or may not serve what was asked.
func upstreamFailure(err error) error {
	if err == nil || errors.Is(err, registry.ErrNotFound) || errors.Is(err, ErrUpstreamDenied) {
		return err
	}

	return fmt.Errorf("%w: %v", ErrUpstreamUnavailable, err)
}

// versionExists refuses an upload of a version Terralist already holds,
// telling how to replace one pulled from the upstream.
func versionExists(version string, pulled bool) error {
	if pulled {
		return fmt.Errorf("version %s %w: it was pulled from the upstream, delete it before uploading your own", version, repositories.ErrAlreadyExists)
	}

	return fmt.Errorf("version %s %w", version, repositories.ErrAlreadyExists)
}

// UpstreamVersion is a provider version an authority may serve from its
// upstream registry.
type UpstreamVersion struct {
	Version   string              `json:"version"`
	Protocols []string            `json:"protocols"`
	Platforms []registry.Platform `json:"platforms"`
}

// UpstreamVersionMetadata is what an upstream registry vouches for about one
// provider version: the SHA256SUMS document with its verified signature and
// the digests it lists.
type UpstreamVersionMetadata struct {
	Protocols       []string                `json:"protocols"`
	ShaSums         map[string]string       `json:"shasums"`
	ShaSumsDocument []byte                  `json:"shasums_document"`
	Signature       []byte                  `json:"signature"`
	SigningKeys     []registry.GPGPublicKey `json:"signing_keys"`
}

// UpstreamPackage locates one provider package on the upstream registry.
type UpstreamPackage struct {
	FileName string
	URL      string
	ShaSum   string
}

// UpstreamService reads providers from the upstream registries authorities
// stand for, verifies what it reads and caches it.
type UpstreamService interface {
	// ProviderVersions returns the upstream versions of a provider that the
	// authority allows, or nothing when the authority has no enabled upstream.
	ProviderVersions(a *authority.Authority, name string) ([]UpstreamVersion, error)

	// ProviderVersion returns the verified metadata of an allowed version.
	ProviderVersion(a *authority.Authority, name, version string) (*UpstreamVersionMetadata, error)

	// ProviderPackage locates the package of one platform of an allowed
	// version, with its digest checked against the verified SHA256SUMS.
	ProviderPackage(a *authority.Authority, name, version, os, arch string) (*UpstreamPackage, error)

	// ModuleVersions returns the upstream versions of a module that the
	// authority allows, or nothing when the authority has no enabled upstream.
	ModuleVersions(a *authority.Authority, name, system string) ([]string, error)

	// ModuleLocation returns the go-getter source of an allowed module
	// version.
	ModuleLocation(a *authority.Authority, name, system, version string) (string, error)
}

type DefaultUpstreamService struct {
	Cache      cache.Cache
	TTL        time.Duration
	Retention  time.Duration
	HTTPClient *http.Client
	Verifier   registry.SignatureVerifier
	Sealer     *secret.Sealer

	now       func() time.Time
	clients   sync.Map
	refreshes singleflight.Group
}

// envelope wraps a cached payload with the time it was read from upstream, so
// that freshness is decided by the service and retention by the cache.
type envelope struct {
	FetchedAt time.Time       `json:"fetched_at"`
	Payload   json.RawMessage `json:"payload"`
}

func (s *DefaultUpstreamService) ProviderVersions(a *authority.Authority, name string) ([]UpstreamVersion, error) {
	if !a.UpstreamEnabled {
		return nil, nil
	}

	var versions []UpstreamVersion
	err := s.cached(a, "versions", s.key(a, name, "versions"), &versions, func(ctx context.Context, client *registry.Client) (any, error) {
		upstream, err := client.ProviderVersions(ctx, *a.UpstreamNamespace, name)
		if errors.Is(err, registry.ErrNotFound) {
			return []UpstreamVersion{}, nil
		}
		if err != nil {
			return nil, err
		}

		return lo.Map(upstream, func(v registry.Version, _ int) UpstreamVersion {
			return UpstreamVersion{Version: v.Version, Protocols: v.Protocols, Platforms: v.Platforms}
		}), nil
	})
	if err != nil {
		return nil, err
	}

	return lo.Filter(versions, func(v UpstreamVersion, _ int) bool {
		return a.AllowsUpstream(authority.RuleKindProvider, name, v.Version)
	}), nil
}

func (s *DefaultUpstreamService) ProviderVersion(a *authority.Authority, name, version string) (*UpstreamVersionMetadata, error) {
	if !a.AllowsUpstream(authority.RuleKindProvider, name, version) {
		return nil, ErrUpstreamDenied
	}

	var metadata UpstreamVersionMetadata
	err := s.cached(a, "version", s.key(a, name, version), &metadata, func(ctx context.Context, client *registry.Client) (any, error) {
		return s.readVersion(ctx, client, a, name, version)
	})
	if err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (s *DefaultUpstreamService) ProviderPackage(a *authority.Authority, name, version, os, arch string) (*UpstreamPackage, error) {
	metadata, err := s.ProviderVersion(a, name, version)
	if err != nil {
		return nil, err
	}

	client, err := s.client(a)
	if err != nil {
		return nil, err
	}

	download, err := client.ProviderDownload(context.Background(), *a.UpstreamNamespace, name, version, os, arch)
	if err != nil {
		metrics.RecordUpstreamRequest(*a.UpstreamHostname, "package", "error")
		return nil, err
	}

	if want := provider.PackageFileName(name, version, os, arch); download.Filename != want {
		metrics.RecordUpstreamRequest(*a.UpstreamHostname, "package", "error")
		return nil, fmt.Errorf("upstream advertises package %s of %s/%s %s as %q", want, a.Name, name, version, download.Filename)
	}

	expected, ok := metadata.ShaSums[download.Filename]
	if !ok {
		metrics.RecordUpstreamRequest(*a.UpstreamHostname, "package", "error")
		return nil, fmt.Errorf("package %s is not listed in the verified SHA256SUMS of %s/%s %s", download.Filename, a.Name, name, version)
	}

	if expected != download.ShaSum {
		metrics.RecordUpstreamRequest(*a.UpstreamHostname, "package", "error")
		return nil, fmt.Errorf("package %s advertises digest %s but the verified SHA256SUMS lists %s", download.Filename, download.ShaSum, expected)
	}

	metrics.RecordUpstreamRequest(*a.UpstreamHostname, "package", "success")

	return &UpstreamPackage{
		FileName: download.Filename,
		URL:      download.DownloadURL,
		ShaSum:   download.ShaSum,
	}, nil
}

func (s *DefaultUpstreamService) ModuleVersions(a *authority.Authority, name, system string) ([]string, error) {
	if !a.UpstreamEnabled {
		return nil, nil
	}

	var versions []string
	err := s.cached(a, "module_versions", s.moduleKey(a, name, system, "versions"), &versions, func(ctx context.Context, client *registry.Client) (any, error) {
		upstream, err := client.ModuleVersions(ctx, *a.UpstreamNamespace, name, system)
		if errors.Is(err, registry.ErrNotFound) {
			return []string{}, nil
		}
		if err != nil {
			return nil, err
		}

		return lo.Map(upstream, func(v registry.ModuleVersion, _ int) string { return v.Version }), nil
	})
	if err != nil {
		return nil, err
	}

	ruleName := name + "/" + system

	return lo.Filter(versions, func(v string, _ int) bool {
		return a.AllowsUpstream(authority.RuleKindModule, ruleName, v)
	}), nil
}

func (s *DefaultUpstreamService) ModuleLocation(a *authority.Authority, name, system, version string) (string, error) {
	if !a.AllowsUpstream(authority.RuleKindModule, name+"/"+system, version) {
		return "", ErrUpstreamDenied
	}

	client, err := s.client(a)
	if err != nil {
		return "", err
	}

	location, err := client.ModuleLocation(context.Background(), *a.UpstreamNamespace, name, system, version)
	if err != nil {
		metrics.RecordUpstreamRequest(*a.UpstreamHostname, "module_location", "error")
		return "", err
	}

	metrics.RecordUpstreamRequest(*a.UpstreamHostname, "module_location", "success")

	return location, nil
}

// readVersion fetches the download metadata of a version, then its SHA256SUMS
// document and signature, and verifies the signature with the advertised keys.
func (s *DefaultUpstreamService) readVersion(ctx context.Context, client *registry.Client, a *authority.Authority, name, version string) (*UpstreamVersionMetadata, error) {
	versions, err := client.ProviderVersions(ctx, *a.UpstreamNamespace, name)
	if err != nil {
		return nil, err
	}

	upstream, found := lo.Find(versions, func(v registry.Version) bool { return v.Version == version })
	if !found || len(upstream.Platforms) == 0 {
		return nil, fmt.Errorf("version %s of %s/%s: %w", version, *a.UpstreamNamespace, name, registry.ErrNotFound)
	}

	// Every platform shares the SHA256SUMS document, so any platform's download
	// metadata leads to it.
	platform := upstream.Platforms[0]
	download, err := client.ProviderDownload(ctx, *a.UpstreamNamespace, name, version, platform.OS, platform.Arch)
	if err != nil {
		return nil, err
	}

	document, err := s.fetch(ctx, download.ShaSumsURL)
	if err != nil {
		return nil, fmt.Errorf("could not fetch SHA256SUMS: %w", err)
	}

	signature, err := s.fetch(ctx, download.ShaSumsSignatureURL)
	if err != nil {
		return nil, fmt.Errorf("could not fetch the SHA256SUMS signature: %w", err)
	}

	if _, err := s.Verifier.VerifyShaSums(document, signature, download.SigningKeys.GPGPublicKeys); err != nil {
		return nil, fmt.Errorf("SHA256SUMS of %s/%s %s: %w", *a.UpstreamNamespace, name, version, err)
	}

	sums, err := registry.ParseShaSums(bytes.NewReader(document))
	if err != nil {
		return nil, err
	}

	return &UpstreamVersionMetadata{
		Protocols:       upstream.Protocols,
		ShaSums:         sums,
		ShaSumsDocument: document,
		Signature:       signature,
		SigningKeys:     download.SigningKeys.GPGPublicKeys,
	}, nil
}

// cached answers from the cache while the entry is fresh, refreshes it through
// read otherwise, and falls back to a stale entry when the refresh fails.
func (s *DefaultUpstreamService) cached(a *authority.Authority, operation, key string, out any, read func(context.Context, *registry.Client) (any, error)) error {
	ctx := context.Background()

	stored, ok, err := s.Cache.Get(ctx, key)
	if err != nil {
		log.Warn().Err(err).Str("key", key).Msg("Could not read the upstream cache.")
	}

	var entry envelope
	if ok {
		if err := json.Unmarshal(stored, &entry); err != nil {
			log.Warn().Err(err).Str("key", key).Msg("Discarding an unreadable upstream cache entry.")
			ok = false
		}
	}

	if ok && s.clock().Sub(entry.FetchedAt) < s.TTL {
		return json.Unmarshal(entry.Payload, out)
	}

	client, err := s.client(a)
	if err != nil {
		return err
	}

	// Concurrent requests for the same entry wait for one upstream read.
	raw, err, _ := s.refreshes.Do(key, func() (any, error) {
		return s.refresh(ctx, client, key, read)
	})
	if err != nil {
		if ok {
			metrics.RecordUpstreamRequest(*a.UpstreamHostname, operation, "stale")
			log.Warn().Err(err).Str("upstream", *a.UpstreamHostname).Str("key", key).Msg("Upstream unavailable, serving the last known answer.")

			return json.Unmarshal(entry.Payload, out)
		}

		metrics.RecordUpstreamRequest(*a.UpstreamHostname, operation, "error")

		return err
	}

	metrics.RecordUpstreamRequest(*a.UpstreamHostname, operation, "success")

	return json.Unmarshal(raw.([]byte), out) //nolint:forcetypeassert
}

// refresh reads an entry from the upstream and stores it in the cache.
func (s *DefaultUpstreamService) refresh(ctx context.Context, client *registry.Client, key string, read func(context.Context, *registry.Client) (any, error)) ([]byte, error) {
	payload, err := read(ctx, client)
	if err != nil {
		return nil, err
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	sealed, err := json.Marshal(envelope{FetchedAt: s.clock(), Payload: raw})
	if err != nil {
		return nil, err
	}

	if err := s.Cache.Set(ctx, key, sealed, s.Retention); err != nil {
		log.Warn().Err(err).Str("key", key).Msg("Could not write the upstream cache.")
	}

	return raw, nil
}

// client returns the registry client of an authority, created on first use.
func (s *DefaultUpstreamService) client(a *authority.Authority) (*registry.Client, error) {
	token := ""
	if a.UpstreamToken != nil {
		if s.Sealer == nil {
			return nil, fmt.Errorf("authority %s has an upstream token but the upstream-secret option is not set", a.Name)
		}

		opened, err := s.Sealer.Open(*a.UpstreamToken, a.ID.String())
		if err != nil {
			return nil, fmt.Errorf("could not open the upstream token of authority %s: %w", a.Name, err)
		}

		token = opened
	}

	key := fmt.Sprintf("%s|%s|%s", a.ID, a.UpstreamBaseURL(), token)
	if client, ok := s.clients.Load(key); ok {
		return client.(*registry.Client), nil //nolint:forcetypeassert
	}

	client := registry.New(a.UpstreamBaseURL(), registry.WithToken(token), registry.WithHTTPClient(s.HTTPClient))
	s.clients.Store(key, client)

	return client, nil
}

// fetch downloads a small document such as a SHA256SUMS file.
func (s *DefaultUpstreamService) fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to %s returned status %d", url, resp.StatusCode)
	}

	return registry.ReadLimited(resp.Body, 1<<20)
}

func (s *DefaultUpstreamService) key(a *authority.Authority, name, suffix string) string {
	return fmt.Sprintf("upstream/%s/providers/%s/%s", a.ID, name, suffix)
}

func (s *DefaultUpstreamService) moduleKey(a *authority.Authority, name, system, suffix string) string {
	return fmt.Sprintf("upstream/%s/modules/%s/%s/%s", a.ID, name, system, suffix)
}

// clock returns the current time, overridable in tests.
func (s *DefaultUpstreamService) clock() time.Time {
	if s.now == nil {
		return time.Now()
	}

	return s.now()
}
