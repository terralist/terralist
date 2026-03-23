package e2e

import (
	"net/http"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderListVersions(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/null/versions"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/nonexistent/versions"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("success", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/null/versions"), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, body, "versions")

		versions := body["versions"].([]any)
		require.Len(t, versions, 1)

		version := versions[0].(map[string]any)
		assert.Equal(t, "3.2.4", version["version"])

		platforms := version["platforms"].([]any)
		require.Len(t, platforms, 1)

		platform := platforms[0].(map[string]any)
		assert.Equal(t, runtime.GOOS, platform["os"])
		assert.Equal(t, runtime.GOARCH, platform["arch"])
	})
}

func TestProviderDownload(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/null/3.2.4/download/%s/%s", runtime.GOOS, runtime.GOARCH), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("not found provider", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/nonexistent/3.2.4/download/%s/%s", runtime.GOOS, runtime.GOARCH), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("not found version", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/null/9.9.9/download/%s/%s", runtime.GOOS, runtime.GOARCH), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("success", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/null/3.2.4/download/%s/%s", runtime.GOOS, runtime.GOARCH), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, runtime.GOOS, body["os"])
		assert.Equal(t, runtime.GOARCH, body["arch"])
		assert.Contains(t, body, "download_url")
		assert.Contains(t, body, "shasum")
		assert.Contains(t, body, "shasums_url")
		assert.Contains(t, body, "shasums_signature_url")
		assert.Contains(t, body, "signing_keys")
	})
}

func TestProviderUpload(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodPost, apiURL("/v1/api/providers/hashicorp/null/4.0.0/upload"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("no body", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodPost, apiURL("/v1/api/providers/hashicorp/null/4.0.0/upload"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("duplicate version", func(t *testing.T) {
		body := map[string]any{
			"protocols": []string{"6.0"},
			"shasums": map[string]string{
				"url":           "https://example.com/shasums",
				"signature_url": "https://example.com/shasums.sig",
			},
			"platforms": []map[string]string{
				{
					"os":           runtime.GOOS,
					"arch":         runtime.GOARCH,
					"download_url": "https://example.com/provider.zip",
					"shasum":       "abc123",
				},
			},
		}
		resp := doAuthRequest(t, http.MethodPost, apiURL("/v1/api/providers/hashicorp/null/3.2.4/upload"), body)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})
}

func TestProviderDeleteUnauthenticated(t *testing.T) {
	t.Run("delete version", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodDelete, apiURL("/v1/api/providers/hashicorp/null/3.2.4/remove"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("delete provider", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodDelete, apiURL("/v1/api/providers/hashicorp/null/remove"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
