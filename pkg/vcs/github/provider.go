package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"terralist/pkg/vcs"
	"terralist/pkg/version"

	"github.com/gin-gonic/gin"
)

type Provider struct {
	WebhookSecret string

	AccessToken       string
	AppID             int
	AppInstallationID int
	AppPrivateKeyPath string
	BaseURL           string
}

var _ vcs.Provider = &Provider{}

func (p *Provider) GetHeaders() map[string]string {
	headers := map[string]string{
		"Accept":               "application/vnd.github.v3+json",
		"X-GitHub-Api-Version": "2022-11-28",
	}
	if p.AppID != 0 && p.AppInstallationID != 0 && p.AppPrivateKeyPath != "" {
		headers["Authorization"] = "Bearer " + p.AccessToken
	}
	if p.AccessToken != "" {
		headers["Authorization"] = "Bearer " + p.AccessToken
	}
	return headers
}

func (p *Provider) Authenticate(ctx *gin.Context, body []byte) error {
	if p.WebhookSecret == "" {
		return nil
	}
	signatureHeader := ctx.GetHeader("X-Hub-Signature-256")
	if signatureHeader == "" {
		return fmt.Errorf("missing signature")
	}
	const prefix = "sha256="
	if !strings.HasPrefix(signatureHeader, prefix) {
		return fmt.Errorf("invalid signature format")
	}
	gotHex := strings.TrimPrefix(signatureHeader, prefix)
	got, err := hex.DecodeString(gotHex)
	if err != nil {
		return fmt.Errorf("invalid signature encoding")
	}
	mac := hmac.New(sha256.New, []byte(p.WebhookSecret))
	mac.Write(body)
	want := mac.Sum(nil)
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}

func (p *Provider) BuildReleaseEventFromWebhook(body []byte) (*vcs.ReleaseEvent, error) {
	var payload ReleasePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode github webhook: %w", err)
	}

	if payload.Action != "published" {
		return nil, nil
	}

	if payload.Release.Draft {
		return nil, nil
	}

	semVer := version.Version(strings.TrimPrefix(payload.Release.TagName, "v"))
	if !semVer.Valid() {
		return nil, fmt.Errorf("invalid release tag %q", payload.Release.TagName)
	}

	archive := payload.Release.ZipballURL
	if archive == "" {
		archive = payload.Release.TarballURL
	}
	if archive == "" {
		return nil, fmt.Errorf("release has no zipball_url or tarball_url")
	}

	ev := &vcs.ReleaseEvent{
		Source:           vcs.ReleaseSourceGitHub,
		Tag:              payload.Release.TagName,
		SemVer:           string(semVer),
		Draft:            payload.Release.Draft,
		Prerelease:       payload.Release.Prerelease,
		ModuleArchiveURL: archive,
	}

	fullName := strings.TrimSpace(payload.Repository.FullName)
	for _, a := range payload.Release.Assets {
		if a.Name == "" {
			continue
		}
		var u string
		if a.ID != 0 && fullName != "" {
			u = fmt.Sprintf("https://api.github.com/repos/%s/releases/assets/%d", fullName, a.ID)
		} else if a.BrowserDownloadURL != "" {
			u = a.BrowserDownloadURL
		}
		if u == "" {
			continue
		}
		ev.Assets = append(ev.Assets, vcs.ReleaseAsset{Name: a.Name, URL: u})
	}

	if payload.Repository.HTMLURL != "" {
		if c, err := vcs.CanonicalVCSRepoURL(payload.Repository.HTMLURL); err == nil {
			ev.RepoURL = c
		}
	} else if payload.Repository.FullName != "" {
		if c, err := vcs.CanonicalVCSRepoURL("https://github.com/" + payload.Repository.FullName); err == nil {
			ev.RepoURL = c
		}
	}

	return ev, nil
}
