// Package registry is a client for the Terraform provider registry protocol,
// used to read upstream registries such as registry.terraform.io.
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

	// providersService is the service discovery key of the provider registry
	// protocol.
	providersService = "providers.v1"

	defaultTimeout = 30 * time.Second
)

var (
	// ErrNotFound is returned when the registry does not know the requested
	// provider, version or platform.
	ErrNotFound = errors.New("not found in the upstream registry")
)

// Client reads providers from one upstream registry.
type Client struct {
	baseURL string
	token   string
	http    *http.Client

	mu           sync.Mutex
	providersURL *url.URL
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

// Versions lists the versions of a provider.
func (c *Client) Versions(ctx context.Context, namespace, name string) ([]Version, error) {
	var body struct {
		Versions []Version `json:"versions"`
	}

	if err := c.getProviders(ctx, path.Join(namespace, name, "versions"), &body); err != nil {
		return nil, err
	}

	return body.Versions, nil
}

// Download fetches the package metadata of a provider version for a platform.
func (c *Client) Download(ctx context.Context, namespace, name, version, os, arch string) (*Download, error) {
	var body Download

	if err := c.getProviders(ctx, path.Join(namespace, name, version, "download", os, arch), &body); err != nil {
		return nil, err
	}

	return &body, nil
}

// getProviders performs a GET under the providers service and decodes the JSON
// response into out.
func (c *Client) getProviders(ctx context.Context, relative string, out any) error {
	base, err := c.discover(ctx)
	if err != nil {
		return err
	}

	endpoint := base.JoinPath(relative)

	return c.getJSON(ctx, endpoint.String(), out)
}

// discover resolves the providers service URL of the registry once.
func (c *Client) discover(ctx context.Context) (*url.URL, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.providersURL != nil {
		return c.providersURL, nil
	}

	var services map[string]string
	if err := c.getJSON(ctx, c.baseURL+discoveryPath, &services); err != nil {
		return nil, fmt.Errorf("service discovery failed: %w", err)
	}

	service, ok := services[providersService]
	if !ok || service == "" {
		return nil, fmt.Errorf("registry %s does not offer the %s service", c.baseURL, providersService)
	}

	base, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid registry URL %q: %w", c.baseURL, err)
	}

	serviceURL, err := url.Parse(service)
	if err != nil {
		return nil, fmt.Errorf("invalid %s service URL %q: %w", providersService, service, err)
	}

	c.providersURL = base.ResolveReference(serviceURL)

	return c.providersURL, nil
}

// getJSON performs an authenticated GET and decodes the JSON response into out.
func (c *Client) getJSON(ctx context.Context, endpoint string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request to %s failed: %w", endpoint, err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return fmt.Errorf("%s: %w", endpoint, ErrNotFound)
	case resp.StatusCode != http.StatusOK:
		return fmt.Errorf("request to %s returned status %d", endpoint, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("could not decode the response of %s: %w", endpoint, err)
	}

	return nil
}
