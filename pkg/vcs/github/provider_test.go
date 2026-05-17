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
