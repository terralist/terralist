package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	moduleWebhookVersion   = "1.3.0"
	providerWebhookVersion = "3.2.3"
	moduleArchiveURL       = "https://github.com/hashicorp/terraform-cidr-subnets/archive/refs/tags/v1.0.0.zip"
	moduleWebhookURL       = "/v1/api/modules/hashicorp/subnets/cidr/webhook/github"
	providerWebhookURL     = "/v1/api/providers/hashicorp/null/webhook/github"
)

func moduleReleasePayload(action, tag string) map[string]any {
	return map[string]any{
		"action": action,
		"release": map[string]any{
			"tag_name":    tag,
			"draft":       false,
			"zipball_url": moduleArchiveURL,
		},
		"repository": map[string]any{
			"full_name": "hashicorp/terraform-cidr-subnets",
			"html_url":  "https://github.com/hashicorp/terraform-cidr-subnets",
		},
	}
}

func marshalWebhookPayload(t *testing.T, payload map[string]any) []byte {
	t.Helper()
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	return body
}

func moduleVersionCount(t *testing.T) int {
	t.Helper()

	resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/modules/hashicorp/subnets/cidr/versions"), nil)
	body := readJSON(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	modules, ok := body["modules"].([]any)
	require.True(t, ok)
	require.Len(t, modules, 1)

	mod, ok := modules[0].(map[string]any)
	require.True(t, ok)
	versions, ok := mod["versions"].([]any)
	require.True(t, ok)
	return len(versions)
}

func moduleVersionExists(t *testing.T, version string) bool {
	t.Helper()

	resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/modules/hashicorp/subnets/cidr/versions"), nil)
	body := readJSON(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	modules, ok := body["modules"].([]any)
	require.True(t, ok)
	require.Len(t, modules, 1)

	mod, ok := modules[0].(map[string]any)
	require.True(t, ok)
	versions, ok := mod["versions"].([]any)
	require.True(t, ok)

	for _, v := range versions {
		ver, ok := v.(map[string]any)
		require.True(t, ok)
		if ver["version"] == version {
			return true
		}
	}
	return false
}

func providerVersionExists(t *testing.T, version string) bool {
	t.Helper()

	resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/null/versions"), nil)
	body := readJSON(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	versions, ok := body["versions"].([]any)
	require.True(t, ok)

	for _, v := range versions {
		ver, ok := v.(map[string]any)
		require.True(t, ok)
		if ver["version"] == version {
			return true
		}
	}
	return false
}

func fetchNullProviderReleaseMetadata(t *testing.T, version string) (shasumsURL, shasumsSigURL, downloadURL string) {
	t.Helper()

	registryURL := fmt.Sprintf(
		"https://registry.terraform.io/v1/providers/hashicorp/null/%s/download/%s/%s",
		version, runtime.GOOS, runtime.GOARCH,
	)
	resp, err := http.Get(registryURL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var metadata map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&metadata))

	shasumsURL, _ = metadata["shasums_url"].(string)
	shasumsSigURL, _ = metadata["shasums_signature_url"].(string)
	downloadURL, _ = metadata["download_url"].(string)
	require.NotEmpty(t, shasumsURL)
	require.NotEmpty(t, shasumsSigURL)
	require.NotEmpty(t, downloadURL)
	return shasumsURL, shasumsSigURL, downloadURL
}

func providerReleasePayload(t *testing.T, version string, includeChecksums bool) map[string]any {
	t.Helper()

	shasumsURL, shasumsSigURL, downloadURL := fetchNullProviderReleaseMetadata(t, version)
	prefix := fmt.Sprintf("terraform-provider-null_%s", version)

	assets := []map[string]any{
		{
			"name":                 fmt.Sprintf("%s_%s_%s.zip", prefix, runtime.GOOS, runtime.GOARCH),
			"browser_download_url": downloadURL,
		},
	}
	if includeChecksums {
		assets = append([]map[string]any{
			{
				"name":                 prefix + "_SHA256SUMS",
				"browser_download_url": shasumsURL,
			},
			{
				"name":                 prefix + "_SHA256SUMS.sig",
				"browser_download_url": shasumsSigURL,
			},
		}, assets...)
	}

	return map[string]any{
		"action": "published",
		"release": map[string]any{
			"tag_name":    fmt.Sprintf("v%s", version),
			"draft":       false,
			"zipball_url": fmt.Sprintf("https://github.com/hashicorp/terraform-provider-null/archive/refs/tags/v%s.zip", version),
			"assets":      assets,
		},
		"repository": map[string]any{
			"full_name": "hashicorp/terraform-provider-null",
			"html_url":  "https://github.com/hashicorp/terraform-provider-null",
		},
	}
}

func TestModuleWebhook(t *testing.T) {
	t.Run("invalid signature", func(t *testing.T) {
		body := marshalWebhookPayload(t, moduleReleasePayload("published", "v"+moduleWebhookVersion))
		resp := doWebhookRequestWithSignature(t, apiURL(moduleWebhookURL), body, "sha256=deadbeef")
		result := readJSON(t, resp)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		errors, ok := result["errors"].([]any)
		require.True(t, ok)
		require.NotEmpty(t, errors)
	})

	t.Run("ignored non-published action", func(t *testing.T) {
		before := moduleVersionCount(t)
		body := marshalWebhookPayload(t, moduleReleasePayload("created", "v9.9.9"))
		resp := doWebhookRequest(t, apiURL(moduleWebhookURL), body)
		result := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Empty(t, result["errors"])
		assert.Equal(t, before, moduleVersionCount(t))
	})

	t.Run("unknown authority", func(t *testing.T) {
		body := marshalWebhookPayload(t, moduleReleasePayload("published", "v9.9.9"))
		resp := doWebhookRequest(t, apiURL("/v1/api/modules/nonexistent/subnets/cidr/webhook/github"), body)
		result := readJSON(t, resp)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		errors, ok := result["errors"].([]any)
		require.True(t, ok)
		require.NotEmpty(t, errors)
	})

	t.Run("success", func(t *testing.T) {
		require.False(t, moduleVersionExists(t, moduleWebhookVersion))

		body := marshalWebhookPayload(t, moduleReleasePayload("published", "v"+moduleWebhookVersion))
		resp := doWebhookRequest(t, apiURL(moduleWebhookURL), body)
		result := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Empty(t, result["errors"])
		assert.True(t, moduleVersionExists(t, moduleWebhookVersion))

		t.Cleanup(func() {
			resp := doAuthRequest(t, http.MethodDelete, apiURL("/v1/api/modules/hashicorp/subnets/cidr/%s/remove", moduleWebhookVersion), nil)
			resp.Body.Close()
		})
	})

	t.Run("duplicate version", func(t *testing.T) {
		body := marshalWebhookPayload(t, moduleReleasePayload("published", "v"+moduleWebhookVersion))

		first := doWebhookRequest(t, apiURL(moduleWebhookURL), body)
		firstResult := readJSON(t, first)
		require.Equal(t, http.StatusOK, first.StatusCode, "first upload should succeed")
		assert.Empty(t, firstResult["errors"])

		resp := doWebhookRequest(t, apiURL(moduleWebhookURL), body)
		result := readJSON(t, resp)

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
		errors, ok := result["errors"].([]any)
		require.True(t, ok)
		require.NotEmpty(t, errors)

		t.Cleanup(func() {
			resp := doAuthRequest(t, http.MethodDelete, apiURL("/v1/api/modules/hashicorp/subnets/cidr/%s/remove", moduleWebhookVersion), nil)
			resp.Body.Close()
		})
	})
}

func TestProviderWebhook(t *testing.T) {
	t.Run("invalid signature", func(t *testing.T) {
		body := marshalWebhookPayload(t, providerReleasePayload(t, providerWebhookVersion, true))
		resp := doWebhookRequestWithSignature(t, apiURL(providerWebhookURL), body, "sha256=deadbeef")
		result := readJSON(t, resp)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		errors, ok := result["errors"].([]any)
		require.True(t, ok)
		require.NotEmpty(t, errors)
	})

	t.Run("missing checksum asset", func(t *testing.T) {
		body := marshalWebhookPayload(t, providerReleasePayload(t, providerWebhookVersion, false))
		resp := doWebhookRequest(t, apiURL(providerWebhookURL), body)
		result := readJSON(t, resp)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		errors, ok := result["errors"].([]any)
		require.True(t, ok)
		require.NotEmpty(t, errors)
	})

	t.Run("success", func(t *testing.T) {
		require.False(t, providerVersionExists(t, providerWebhookVersion))

		body := marshalWebhookPayload(t, providerReleasePayload(t, providerWebhookVersion, true))
		resp := doWebhookRequest(t, apiURL(providerWebhookURL), body)
		result := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Empty(t, result["errors"])
		assert.True(t, providerVersionExists(t, providerWebhookVersion))

		downloadResp := doAuthRequest(
			t,
			http.MethodGet,
			apiURL("/v1/providers/hashicorp/null/%s/download/%s/%s", providerWebhookVersion, runtime.GOOS, runtime.GOARCH),
			nil,
		)
		downloadBody := readJSON(t, downloadResp)
		require.Equal(t, http.StatusOK, downloadResp.StatusCode)
		assert.Equal(t, runtime.GOOS, downloadBody["os"])
		assert.Equal(t, runtime.GOARCH, downloadBody["arch"])
		assert.Contains(t, downloadBody, "download_url")

		t.Cleanup(func() {
			resp := doAuthRequest(t, http.MethodDelete, apiURL("/v1/api/providers/hashicorp/null/%s/remove", providerWebhookVersion), nil)
			resp.Body.Close()
		})
	})
}
