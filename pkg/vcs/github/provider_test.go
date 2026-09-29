package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"terralist/pkg/vcs"
	"testing"

	"github.com/gin-gonic/gin"
)

var _ vcs.Provider = &Provider{}

func TestBuildReleaseEventFromWebhookGitHubRelease(t *testing.T) {
	body := []byte(`{
		"action": "published",
		"release": {
			"tag_name": "v2.0.0",
			"draft": false,
			"prerelease": false,
			"zipball_url": "https://api.github.com/repos/o/r/zipball/refs/tags/v2.0.0",
			"tarball_url": "https://api.github.com/repos/o/r/tarball/refs/tags/v2.0.0",
			"assets": [
				{"id": 101, "name": "terraform-provider-x_2.0.0_SHA256SUMS", "browser_download_url": "https://a/sums"},
				{"id": 102, "name": "terraform-provider-x_2.0.0_linux_amd64.zip", "browser_download_url": "https://a/z"}
			]
		},
		"repository": {"full_name": "o/r", "html_url": "https://github.com/o/r"},
		"installation": {"id": 42}
	}`)
	var provider Provider
	ev, err := provider.BuildReleaseEventFromWebhook(body)
	if err != nil {
		t.Fatal(err)
	}
	if ev.SemVer != "2.0.0" || ev.ModuleArchiveURL == "" || len(ev.Assets) != 2 {
		t.Fatalf("%+v", ev)
	}
	if ev.Assets[0].URL != "https://api.github.com/repos/o/r/releases/assets/101" {
		t.Fatalf("asset url %q", ev.Assets[0].URL)
	}
	if ev.Assets[1].URL != "https://api.github.com/repos/o/r/releases/assets/102" {
		t.Fatalf("asset url %q", ev.Assets[1].URL)
	}
	if ev.RepoURL != "https://github.com/o/r" {
		t.Fatalf("repo url %q", ev.RepoURL)
	}
}

func TestBuildReleaseEventFromWebhookGitHubReleaseIgnored(t *testing.T) {
	body := []byte(`{"action":"created","release":{"tag_name":"v1"}}`)
	var provider Provider
	_, err := provider.BuildReleaseEventFromWebhook(body)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVerifyGitHubSignature(t *testing.T) {
	secret := "s"
	body := []byte(`{}`)
	provider := &Provider{
		WebhookSecret: secret,
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if err := provider.Authenticate(&gin.Context{
		Request: &http.Request{
			Header: map[string][]string{
				"X-Hub-Signature-256": {sig},
			},
		},
	}, body); err != nil {
		t.Fatal(err)
	}
	if err := provider.Authenticate(&gin.Context{
		Request: &http.Request{
			Header: map[string][]string{
				"X-Hub-Signature-256": {"sha256=deadbeef"},
			},
		},
	}, body); err == nil {
		t.Fatal("expected err")
	}
	if err := provider.Authenticate(&gin.Context{
		Request: &http.Request{
			Header: map[string][]string{
				"X-Hub-Signature-256": {sig},
			},
		},
	}, body); err != nil {
		t.Fatal(err)
	}
}

func TestAuthenticateWithoutSecret(t *testing.T) {
	provider := &Provider{}
	if err := provider.Authenticate(&gin.Context{
		Request: &http.Request{Header: map[string][]string{}},
	}, []byte(`{}`)); err == nil {
		t.Fatal("expected unsigned webhook to be rejected when no secret is configured")
	}
}

func TestAuthenticateMissingSignature(t *testing.T) {
	provider := &Provider{WebhookSecret: "s"}
	if err := provider.Authenticate(&gin.Context{
		Request: &http.Request{Header: map[string][]string{}},
	}, []byte(`{}`)); err == nil {
		t.Fatal("expected unsigned webhook to be rejected")
	}
}

func TestConfigRequiresWebhookSecret(t *testing.T) {
	cfg := &Config{BaseURL: "github.com"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected missing webhook secret to fail validation")
	}
	cfg.WebhookSecret = "s"
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestGetHeadersSendsCredentialsOnlyToGitHub(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		urls     []string
		wantAuth bool
	}{
		{"api", "https://github.com", []string{"https://api.github.com/repos/o/r/zipball/v1"}, true},
		{"codeload", "https://github.com", []string{"https://codeload.github.com/o/r/zip/v1"}, true},
		{"github", "https://github.com", []string{"https://github.com/o/r/releases/download/v1/a.zip"}, true},
		{"host case", "https://github.com", []string{"https://API.GitHub.com/repos/o/r/zipball/v1"}, true},
		{"all trusted", "https://github.com", []string{"https://api.github.com/a", "https://github.com/b"}, true},
		{"no urls", "https://github.com", nil, false},
		{"other host", "https://github.com", []string{"https://attacker.example/archive.zip"}, false},
		{"lookalike host", "https://github.com", []string{"https://github.com.attacker.example/a.zip"}, false},
		{"other subdomain", "https://github.com", []string{"https://pages.github.com/a.zip"}, false},
		{"plain http", "https://github.com", []string{"http://api.github.com/repos/o/r/zipball/v1"}, false},
		{"userinfo", "https://github.com", []string{"https://user@api.github.com/a"}, false},
		{"other port", "https://github.com", []string{"https://api.github.com:8443/a"}, false},
		{"one untrusted", "https://github.com", []string{"https://api.github.com/a", "https://attacker.example/b"}, false},
		{"empty url", "https://github.com", []string{""}, false},
		{"enterprise", "https://ghe.example.com", []string{"https://ghe.example.com/api/v3/repos/o/r/zipball/v1"}, true},
		{"enterprise to github.com", "https://ghe.example.com", []string{"https://api.github.com/repos/o/r/zipball/v1"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &Provider{AccessToken: "token", BaseURL: tt.baseURL}
			_, gotAuth := provider.GetHeaders(tt.urls)["Authorization"]
			if gotAuth != tt.wantAuth {
				t.Fatalf("Authorization sent = %v, want %v", gotAuth, tt.wantAuth)
			}
		})
	}
}
