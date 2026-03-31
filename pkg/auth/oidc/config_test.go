package oidc

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestConfigValidate_UsesManualEndpoints(t *testing.T) {
	cfg := &Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		AuthorizeUrl: "https://issuer.example.com/auth",
		TokenUrl:     "https://issuer.example.com/token",
		UserInfoUrl:  "https://issuer.example.com/userinfo",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestConfigValidate_DiscoveryPopulatesEndpoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/issuer/.well-known/openid-configuration" {
			t.Fatalf("unexpected discovery path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"authorization_endpoint": "https://issuer.example.com/auth",
			"token_endpoint": "https://issuer.example.com/token",
			"userinfo_endpoint": "https://issuer.example.com/userinfo",
			"supported_scopes": ["openid", "email", "profile"]
		}`))
	}))
	defer server.Close()

	originalClient := discoveryHTTPClient
	discoveryHTTPClient = server.Client()
	defer func() {
		discoveryHTTPClient = originalClient
	}()

	cfg := &Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Host:         server.URL + "/issuer",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if cfg.AuthorizeUrl != "https://issuer.example.com/auth" {
		t.Fatalf("AuthorizeUrl = %q, want %q", cfg.AuthorizeUrl, "https://issuer.example.com/auth")
	}

	if cfg.TokenUrl != "https://issuer.example.com/token" {
		t.Fatalf("TokenUrl = %q, want %q", cfg.TokenUrl, "https://issuer.example.com/token")
	}

	if cfg.UserInfoUrl != "https://issuer.example.com/userinfo" {
		t.Fatalf("UserInfoUrl = %q, want %q", cfg.UserInfoUrl, "https://issuer.example.com/userinfo")
	}
}

func TestConfigValidate_DiscoveryRequiresSupportedScopes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"authorization_endpoint": "https://issuer.example.com/auth",
			"token_endpoint": "https://issuer.example.com/token",
			"userinfo_endpoint": "https://issuer.example.com/userinfo",
			"supported_scopes": ["openid"]
		}`))
	}))
	defer server.Close()

	originalClient := discoveryHTTPClient
	discoveryHTTPClient = server.Client()
	defer func() {
		discoveryHTTPClient = originalClient
	}()

	cfg := &Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Host:         server.URL,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate should have failed when required OIDC scopes are not supported")
	}
}

func TestBuildDiscoveryURL_AppendsWellKnownPath(t *testing.T) {
	got, err := buildDiscoveryURL("https://issuer.example.com/realms/dev/")
	if err != nil {
		t.Fatalf("buildDiscoveryURL returned error: %v", err)
	}

	want := "https://issuer.example.com/realms/dev/.well-known/openid-configuration"
	if got != want {
		t.Fatalf("buildDiscoveryURL = %q, want %q", got, want)
	}
}

func TestProviderGetAuthorizeURL_UsesRequiredScopes(t *testing.T) {
	provider := &Provider{
		ClientID:     "client-id",
		AuthorizeUrl: "https://issuer.example.com/auth",
		RedirectUrl:  "https://terralist.example.com/v1/api/auth/redirect",
	}

	authorizeURL := provider.GetAuthorizeUrl("state-value")
	parsed, err := url.Parse(authorizeURL)
	if err != nil {
		t.Fatalf("GetAuthorizeUrl returned invalid URL %q: %v", authorizeURL, err)
	}

	if got := parsed.Query().Get("scope"); got != "openid email" {
		t.Fatalf("scope = %q, want %q", got, "openid email")
	}
}
