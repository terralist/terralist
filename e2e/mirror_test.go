package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func platformKey() string {
	return fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
}

// mirrorHostname returns the hostname under which Terralist serves its own
// providers through the network mirror protocol.
func mirrorHostname(t *testing.T) string {
	t.Helper()

	parsed, err := url.Parse(config.URL)
	require.NoError(t, err)

	return parsed.Host
}

func TestMirrorListVersions(t *testing.T) {
	hostname := mirrorHostname(t)

	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodGet, apiURL("/providers/%s/hashicorp/null/index.json", hostname))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("upstream nobody stands for", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/registry.terraform.io/unknown/null/index.json"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/hashicorp/nonexistent/index.json", hostname), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("success", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/hashicorp/null/index.json", hostname), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		versions, ok := body["versions"].(map[string]any)
		require.True(t, ok)
		require.Contains(t, versions, "3.2.4")
		require.Contains(t, versions, nullMirrorOnlyVersion)
		require.Contains(t, versions, nullSignedVersion)
		assert.Equal(t, map[string]any{}, versions["3.2.4"])
	})

	t.Run("upstream address", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/registry.terraform.io/hashicorp/null/index.json"), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		versions, ok := body["versions"].(map[string]any)
		require.True(t, ok)
		assert.Contains(t, versions, "3.2.4")
		assert.Contains(t, versions, nullMirrorOnlyVersion)
	})
}

func TestMirrorListArchives(t *testing.T) {
	hostname := mirrorHostname(t)

	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodGet, apiURL("/providers/%s/hashicorp/null/3.2.4.json", hostname))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("upstream nobody stands for", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/registry.terraform.io/unknown/null/3.2.4.json"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("not found provider", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/hashicorp/nonexistent/3.2.4.json", hostname), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("not found version", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/hashicorp/null/9.9.9.json", hostname), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("without json extension", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/hashicorp/null/3.2.4", hostname), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("success", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/hashicorp/null/3.2.4.json", hostname), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		archives, ok := body["archives"].(map[string]any)
		require.True(t, ok)
		require.Contains(t, archives, platformKey())

		archive, ok := archives[platformKey()].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, []any{"zh:" + bootstrap.NullProviderShaSum}, archive["hashes"])

		downloadURL, ok := archive["url"].(string)
		require.True(t, ok)
		require.NotEmpty(t, downloadURL)

		download, err := httpClient().Get(downloadURL)
		require.NoError(t, err)
		defer download.Body.Close()

		require.Equal(t, http.StatusOK, download.StatusCode)

		content, err := io.ReadAll(download.Body)
		require.NoError(t, err)

		sum := sha256.Sum256(content)
		assert.Equal(t, bootstrap.NullProviderShaSum, hex.EncodeToString(sum[:]))
	})

	t.Run("mirror-only version under the upstream address", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/registry.terraform.io/hashicorp/null/%s.json", nullMirrorOnlyVersion), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		archives, ok := body["archives"].(map[string]any)
		require.True(t, ok)
		require.Contains(t, archives, platformKey())

		archive, ok := archives[platformKey()].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, []any{bootstrap.NullMirrorOnlyH1, "zh:" + bootstrap.NullMirrorOnlyShaSum}, archive["hashes"])

		downloadURL, ok := archive["url"].(string)
		require.True(t, ok)

		download, err := httpClient().Get(downloadURL)
		require.NoError(t, err)
		defer download.Body.Close()

		require.Equal(t, http.StatusOK, download.StatusCode)

		content, err := io.ReadAll(download.Body)
		require.NoError(t, err)
		assert.Equal(t, bootstrap.NullMirrorOnlyArchive, content)
	})
}
