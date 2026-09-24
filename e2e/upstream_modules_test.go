package e2e

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The dir/template versions used below are all published upstream. Each test
// uses its own version so that the fetches of one test do not change what
// another one observes.
const (
	upstreamModuleDownloadVersion = "1.0.2"
	upstreamModuleFetchVersion    = "1.0.1"
	upstreamModuleInitVersion     = "1.0.0"
)

func moduleVersions(t *testing.T, name, system string) (int, []string) {
	t.Helper()

	resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/modules/hashicorp/%s/%s/versions", name, system), nil)
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return resp.StatusCode, nil
	}

	body := readJSON(t, resp)
	modules, _ := body["modules"].([]any)
	require.Len(t, modules, 1)
	mod, _ := modules[0].(map[string]any)
	entries, _ := mod["versions"].([]any)

	var versions []string
	for _, e := range entries {
		entry, _ := e.(map[string]any)
		version, _ := entry["version"].(string)
		versions = append(versions, version)
	}

	return resp.StatusCode, versions
}

func moduleLocation(t *testing.T, version string) string {
	t.Helper()

	resp := doAuthRequestNoRedirect(t, apiURL("/v1/modules/hashicorp/dir/template/%s/download", version))
	defer resp.Body.Close()

	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	return resp.Header.Get("X-Terraform-Get")
}

func TestUpstreamModuleVersions(t *testing.T) {
	t.Run("upstream versions are listed", func(t *testing.T) {
		status, versions := moduleVersions(t, "dir", "template")

		require.Equal(t, http.StatusOK, status)
		assert.Contains(t, versions, upstreamModuleDownloadVersion)
		assert.Contains(t, versions, upstreamModuleFetchVersion)
	})

	t.Run("modules the default policy denies are not listed", func(t *testing.T) {
		status, _ := moduleVersions(t, "consul", "aws")

		assert.Equal(t, http.StatusNotFound, status)
	})
}

func TestUpstreamModuleDownload(t *testing.T) {
	expected := fmt.Sprintf("/v1/modules/hashicorp/dir/template/%s/archive?token=", upstreamModuleDownloadVersion)

	t.Run("download points at the archive route with a token", func(t *testing.T) {
		location := moduleLocation(t, upstreamModuleDownloadVersion)

		assert.Contains(t, location, expected)
	})

	t.Run("the archive route fetches the module and hands storage to go-getter", func(t *testing.T) {
		location := moduleLocation(t, upstreamModuleDownloadVersion)
		require.Contains(t, location, expected)

		resp, err := noRedirectClient().Get(location)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNoContent, resp.StatusCode, "the archive route accepts the token without credentials")
		storage := resp.Header.Get("X-Terraform-Get")
		assert.True(t, strings.HasPrefix(storage, "http"), "expected a storage location, got %q", storage)
		assert.NotContains(t, storage, "/archive")
	})

	t.Run("fetched version is then served from storage", func(t *testing.T) {
		location := moduleLocation(t, upstreamModuleDownloadVersion)

		assert.NotContains(t, location, "/archive")
		assert.True(t, strings.HasPrefix(location, "http"))
	})
}

func TestUpstreamModuleFetch(t *testing.T) {
	url := apiURL("/v1/api/modules/hashicorp/dir/template/%s/fetch", upstreamModuleFetchVersion)

	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodPost, url)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("fetches the version", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodPost, url, map[string]any{})
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode, body)

		location := moduleLocation(t, upstreamModuleFetchVersion)
		assert.NotContains(t, location, "/archive")
	})

	t.Run("denied module", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodPost, apiURL("/v1/api/modules/hashicorp/consul/aws/0.1.0/fetch"), map[string]any{})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
	})
}

func TestTerraformUpstreamModuleInit(t *testing.T) {
	requireTerraformCapable(t)
	host := registryHost(t)

	dir := setupTerraformProject(t, host, fmt.Sprintf(`
module "templates" {
  source  = "%s/hashicorp/dir/template"
  version = "%s"

  base_dir      = "${path.module}/templates"
  template_vars = {}
}
`, host, upstreamModuleInitVersion))

	out := runTerraform(t, dir, "init")
	assert.Contains(t, out, "Terraform has been successfully initialized")
}
