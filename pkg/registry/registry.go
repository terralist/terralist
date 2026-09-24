// Package registry is a client for the Terraform provider and module registry
// protocols, used to read upstream registries such as registry.terraform.io.
package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	// discoveryPath is the service discovery document of a registry host.
	discoveryPath = "/.well-known/terraform.json"

	// providersService and modulesService are the service discovery keys of
	// the provider and module registry protocols.
	providersService = "providers.v1"
	modulesService   = "modules.v1"

	// locationHeader carries the source location of a module version in the
	// module registry protocol download response.
	locationHeader = "X-Terraform-Get"

	defaultTimeout = 30 * time.Second
)

var (
	// ErrNotFound is returned when the registry does not know the requested
	// provider, version or platform.
	ErrNotFound = errors.New("not found in the upstream registry")
)

// Client reads providers and modules from one upstream registry.
type Client struct {
	baseURL string
	token   string
	http    *http.Client

	mu       sync.Mutex
	services map[string]*url.URL
}

// Option configures a Client.
type Option func(*Client)

// WithToken authenticates every request with a bearer token.
func WithToken(token string) Option {
	return func(c *Client) {
		c.token = token
	}
}

// WithHTTPClient replaces the HTTP client used for every request.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.http = client
	}
}

// New creates a client for the registry served at the given base URL, for
// example https://registry.terraform.io.
func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: defaultTimeout},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Version is an entry of the provider registry protocol version list.
type Version struct {
	Version   string     `json:"version"`
	Protocols []string   `json:"protocols"`
	Platforms []Platform `json:"platforms"`
}

// Platform identifies an operating system and architecture a version was
// built for.
type Platform struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// Download is the provider registry protocol package metadata.
type Download struct {
	Protocols           []string    `json:"protocols"`
	OS                  string      `json:"os"`
	Arch                string      `json:"arch"`
	Filename            string      `json:"filename"`
	DownloadURL         string      `json:"download_url"`
	ShaSumsURL          string      `json:"shasums_url"`
	ShaSumsSignatureURL string      `json:"shasums_signature_url"`
	ShaSum              string      `json:"shasum"`
	SigningKeys         SigningKeys `json:"signing_keys"`
}

// SigningKeys holds the keys a registry advertises for a package.
type SigningKeys struct {
	GPGPublicKeys []GPGPublicKey `json:"gpg_public_keys"`
}

// GPGPublicKey is an armored OpenPGP public key advertised by a registry.
type GPGPublicKey struct {
	KeyID          string `json:"key_id"`
	ASCIIArmor     string `json:"ascii_armor"`
	TrustSignature string `json:"trust_signature"`
	Source         string `json:"source"`
	SourceURL      string `json:"source_url"`
}

// ModuleVersion is an entry of the module registry protocol version list.
type ModuleVersion struct {
	Version string `json:"version"`
}

// ProviderVersions lists the versions of a provider.
func (c *Client) ProviderVersions(ctx context.Context, namespace, name string) ([]Version, error) {
	var body struct {
		Versions []Version `json:"versions"`
	}

	if err := c.getJSON(ctx, providersService, path.Join(namespace, name, "versions"), &body); err != nil {
		return nil, err
	}

	return body.Versions, nil
}

// ProviderDownload fetches the package metadata of a provider version for a
// platform.
func (c *Client) ProviderDownload(ctx context.Context, namespace, name, version, os, arch string) (*Download, error) {
	var body Download

	if err := c.getJSON(ctx, providersService, path.Join(namespace, name, version, "download", os, arch), &body); err != nil {
		return nil, err
	}

	return &body, nil
}

// ModuleVersions lists the versions of a module.
func (c *Client) ModuleVersions(ctx context.Context, namespace, name, system string) ([]ModuleVersion, error) {
	var body struct {
		Modules []struct {
			Versions []ModuleVersion `json:"versions"`
		} `json:"modules"`
	}

	if err := c.getJSON(ctx, modulesService, path.Join(namespace, name, system, "versions"), &body); err != nil {
		return nil, err
	}

	var versions []ModuleVersion
	for _, m := range body.Modules {
		versions = append(versions, m.Versions...)
	}

	return versions, nil
}

// ModuleLocation resolves the source location of a module version, a go-getter
// address the module can be fetched from. A relative location is resolved
// against the download endpoint, as Terraform does.
func (c *Client) ModuleLocation(ctx context.Context, namespace, name, system, version string) (string, error) {
	endpoint, err := c.endpoint(ctx, modulesService, path.Join(namespace, name, system, version, "download"))
	if err != nil {
		return "", err
	}

	resp, err := c.get(ctx, endpoint.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return "", fmt.Errorf("request to %s returned status %d", endpoint, resp.StatusCode)
	}

	location := resp.Header.Get(locationHeader)
	if location == "" {
		return "", fmt.Errorf("%s did not return a %s header", endpoint, locationHeader)
	}

	if strings.HasPrefix(location, "/") || strings.HasPrefix(location, "./") || strings.HasPrefix(location, "../") {
		relative, err := url.Parse(location)
		if err != nil {
			return "", fmt.Errorf("invalid module location %q: %w", location, err)
		}

		location = endpoint.ResolveReference(relative).String()
	}

	return location, nil
}

// endpoint builds the URL of a path under one of the discovered services.
func (c *Client) endpoint(ctx context.Context, service, relative string) (*url.URL, error) {
	base, err := c.discover(ctx, service)
	if err != nil {
		return nil, err
	}

	return base.JoinPath(relative), nil
}

// discover resolves the URL of a registry service, fetching the service
// discovery document once.
func (c *Client) discover(ctx context.Context, service string) (*url.URL, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.services == nil {
		var announced map[string]string
		if err := c.getJSONFrom(ctx, c.baseURL+discoveryPath, &announced); err != nil {
			return nil, fmt.Errorf("service discovery failed: %w", err)
		}

		base, err := url.Parse(c.baseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid registry URL %q: %w", c.baseURL, err)
		}

		c.services = make(map[string]*url.URL, len(announced))
		for name, location := range announced {
			serviceURL, err := url.Parse(location)
			if err != nil {
				return nil, fmt.Errorf("invalid %s service URL %q: %w", name, location, err)
			}

			c.services[name] = base.ResolveReference(serviceURL)
		}
	}

	serviceURL, ok := c.services[service]
	if !ok {
		return nil, fmt.Errorf("registry %s does not offer the %s service", c.baseURL, service)
	}

	return serviceURL, nil
}

// getJSON performs a GET under one of the discovered services and decodes the
// JSON response into out.
func (c *Client) getJSON(ctx context.Context, service, relative string, out any) error {
	endpoint, err := c.endpoint(ctx, service, relative)
	if err != nil {
		return err
	}

	return c.getJSONFrom(ctx, endpoint.String(), out)
}

// getJSONFrom performs a GET and decodes the JSON response into out.
func (c *Client) getJSONFrom(ctx context.Context, endpoint string, out any) error {
	resp, err := c.get(ctx, endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request to %s returned status %d", endpoint, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("could not decode the response of %s: %w", endpoint, err)
	}

	return nil
}

// get performs an authenticated GET. A 404 answer is reported as ErrNotFound;
// any other status is left to the caller.
func (c *Client) get(ctx context.Context, endpoint string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", endpoint, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, fmt.Errorf("%s: %w", endpoint, ErrNotFound)
	}

	return resp, nil
}
