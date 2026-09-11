package e2e

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mirrorTestHostname is a hostname seeded by the deletion tests, kept apart
// from the bootstrapped registry.terraform.io data.
const mirrorTestHostname = "mirror.example.com"

func platformKey() string {
	return fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
}

// mirrorUpload uploads the bootstrapped null provider package under the given
// coordinates and returns the response.
func mirrorUpload(t *testing.T, hostname, namespace, name, version string) *http.Response {
	t.Helper()

	files := mirrorUploadFiles(nullProviderArchiveName(), bootstrap.NullProviderH1, bootstrap.NullProviderArchive)

	return doAuthMultipartRequest(t, apiURL("/v1/api/mirror/%s/%s/%s/%s/upload", hostname, namespace, name, version), files)
}

// requireMirrorUpload uploads the bootstrapped null provider package under
// the deletion tests hostname and fails the test if the upload is rejected.
func requireMirrorUpload(t *testing.T, namespace, name, version string) {
	t.Helper()

	resp := mirrorUpload(t, mirrorTestHostname, namespace, name, version)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
}

func TestMirrorListVersions(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodGet, apiURL("/providers/registry.terraform.io/hashicorp/null/index.json"))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/registry.terraform.io/hashicorp/nonexistent/index.json"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("success", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/registry.terraform.io/hashicorp/null/index.json"), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		versions, ok := body["versions"].(map[string]any)
		require.True(t, ok)
		require.Contains(t, versions, "3.2.4")
		assert.Equal(t, map[string]any{}, versions["3.2.4"])
	})
}

func TestMirrorGetVersion(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodGet, apiURL("/providers/registry.terraform.io/hashicorp/null/3.2.4.json"))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("not found provider", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/registry.terraform.io/hashicorp/nonexistent/3.2.4.json"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("not found version", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/registry.terraform.io/hashicorp/null/9.9.9.json"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("without json extension", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/registry.terraform.io/hashicorp/null/3.2.4"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("success", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/registry.terraform.io/hashicorp/null/3.2.4.json"), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		archives, ok := body["archives"].(map[string]any)
		require.True(t, ok)
		require.Contains(t, archives, platformKey())

		archive, ok := archives[platformKey()].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, []any{bootstrap.NullProviderH1}, archive["hashes"])

		downloadURL, ok := archive["url"].(string)
		require.True(t, ok)
		require.NotEmpty(t, downloadURL)

		download, err := httpClient().Get(downloadURL)
		require.NoError(t, err)
		defer download.Body.Close()

		require.Equal(t, http.StatusOK, download.StatusCode)

		content, err := io.ReadAll(download.Body)
		require.NoError(t, err)
		assert.Equal(t, bootstrap.NullProviderArchive, content)
	})
}

func TestMirrorUpload(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodPost, apiURL("/v1/api/mirror/registry.terraform.io/hashicorp/null/4.0.0/upload"))
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("not multipart", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodPost, apiURL("/v1/api/mirror/registry.terraform.io/hashicorp/null/4.0.0/upload"), map[string]any{})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing metadata", func(t *testing.T) {
		resp := doAuthMultipartUpload(t, apiURL("/v1/api/mirror/registry.terraform.io/hashicorp/null/4.0.0/upload"), "archives", nullProviderArchiveName(), bootstrap.NullProviderArchive)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("duplicate platform", func(t *testing.T) {
		resp := mirrorUpload(t, "registry.terraform.io", "hashicorp", "null", "3.2.4")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("invalid version", func(t *testing.T) {
		resp := mirrorUpload(t, "registry.terraform.io", "hashicorp", "null", "latest")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("hash mismatch", func(t *testing.T) {
		files := mirrorUploadFiles(nullProviderArchiveName(), "h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", bootstrap.NullProviderArchive)

		resp := doAuthMultipartRequest(t, apiURL("/v1/api/mirror/registry.terraform.io/hashicorp/null/4.0.0/upload"), files)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		index := doAuthRequest(t, http.MethodGet, apiURL("/providers/registry.terraform.io/hashicorp/null/index.json"), nil)
		body := readJSON(t, index)
		versions, _ := body["versions"].(map[string]any)
		assert.NotContains(t, versions, "4.0.0")
	})

	t.Run("archive not listed in metadata", func(t *testing.T) {
		files := mirrorUploadFiles("other.zip", bootstrap.NullProviderH1, bootstrap.NullProviderArchive)
		files[1].FileName = nullProviderArchiveName()

		resp := doAuthMultipartRequest(t, apiURL("/v1/api/mirror/registry.terraform.io/hashicorp/null/4.0.0/upload"), files)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})
}

func TestMirrorDelete(t *testing.T) {
	hostname := mirrorTestHostname

	t.Run("unauthenticated", func(t *testing.T) {
		for _, path := range []string{
			"/v1/api/mirror/%s/hashicorp/null/3.2.4",
			"/v1/api/mirror/%s/hashicorp/null",
			"/v1/api/mirror/%s/hashicorp",
			"/v1/api/mirror/%s",
		} {
			resp := doUnauthRequest(t, http.MethodDelete, apiURL(path, hostname))
			resp.Body.Close()

			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, path)
		}
	})

	t.Run("not found", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodDelete, apiURL("/v1/api/mirror/%s/hashicorp/null/3.2.4", hostname), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// Seed a hostname with two namespaces so that every deletion scope can be
	// exercised against real data.
	requireMirrorUpload(t, "team-a", "null", "1.0.0")
	requireMirrorUpload(t, "team-a", "null", "1.0.1")
	requireMirrorUpload(t, "team-a", "random", "1.0.0")
	requireMirrorUpload(t, "team-b", "null", "1.0.0")

	t.Run("delete version", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodDelete, apiURL("/v1/api/mirror/%s/team-a/null/1.0.0", hostname), nil)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		index := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/team-a/null/index.json", hostname), nil)
		body := readJSON(t, index)
		versions, _ := body["versions"].(map[string]any)
		assert.NotContains(t, versions, "1.0.0")
		assert.Contains(t, versions, "1.0.1")
	})

	t.Run("delete provider", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodDelete, apiURL("/v1/api/mirror/%s/team-a/random", hostname), nil)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		index := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/team-a/random/index.json", hostname), nil)
		defer index.Body.Close()

		assert.Equal(t, http.StatusNotFound, index.StatusCode)
	})

	t.Run("delete namespace", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodDelete, apiURL("/v1/api/mirror/%s/team-a", hostname), nil)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		index := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/team-a/null/index.json", hostname), nil)
		defer index.Body.Close()

		assert.Equal(t, http.StatusNotFound, index.StatusCode)

		other := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/team-b/null/index.json", hostname), nil)
		defer other.Body.Close()

		assert.Equal(t, http.StatusOK, other.StatusCode)
	})

	t.Run("delete hostname", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodDelete, apiURL("/v1/api/mirror/%s", hostname), nil)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		index := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/team-b/null/index.json", hostname), nil)
		defer index.Body.Close()

		assert.Equal(t, http.StatusNotFound, index.StatusCode)
	})
}
