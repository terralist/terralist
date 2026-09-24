package e2e

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The random provider versions used below are all published upstream. Each
// test uses its own version so that the fetches of one test do not change
// what another one observes.
const (
	upstreamListedVersion   = "3.6.3"
	upstreamMirrorVersion   = "3.6.2"
	upstreamRegistryVersion = "3.6.1"
	upstreamFetchVersion    = "3.6.0"
	upstreamInitVersion     = "3.5.1"
)

// noRedirectClient returns an HTTP client that reports redirects instead of
// following them.
func noRedirectClient() *http.Client {
	client := httpClient()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	return client
}

// doAuthRequestNoRedirect performs an authenticated GET without following
// redirects.
func doAuthRequestNoRedirect(t *testing.T, url string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer x-api-key:"+config.MasterAPIKey)

	resp, err := noRedirectClient().Do(req)
	require.NoError(t, err)

	return resp
}

func mirrorIndex(t *testing.T, hostname, namespace, name string) (int, map[string]any) {
	t.Helper()

	resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/%s/%s/index.json", hostname, namespace, name), nil)
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return resp.StatusCode, nil
	}

	body := readJSON(t, resp)
	versions, _ := body["versions"].(map[string]any)

	return resp.StatusCode, versions
}

func TestUpstreamVersions(t *testing.T) {
	t.Run("upstream versions are listed under the upstream address", func(t *testing.T) {
		status, versions := mirrorIndex(t, "registry.terraform.io", "hashicorp", upstreamProvider)

		require.Equal(t, http.StatusOK, status)
		assert.Contains(t, versions, upstreamListedVersion)
		assert.NotContains(t, versions, upstreamDeniedVersion, "a version matched by a deny rule is not listed")
	})

	t.Run("upstream versions are listed under Terralist's own address", func(t *testing.T) {
		status, versions := mirrorIndex(t, mirrorHostname(t), "hashicorp", upstreamProvider)

		require.Equal(t, http.StatusOK, status)
		assert.Contains(t, versions, upstreamListedVersion)
	})

	t.Run("providers the default policy denies are not listed", func(t *testing.T) {
		status, _ := mirrorIndex(t, "registry.terraform.io", "hashicorp", "aws")

		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("registry protocol lists upstream versions", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/%s/versions", upstreamProvider), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		versions, ok := body["versions"].([]any)
		require.True(t, ok)

		var listed []string
		for _, v := range versions {
			version, _ := v.(map[string]any)
			number, _ := version["version"].(string)
			listed = append(listed, number)
		}

		assert.Contains(t, listed, upstreamListedVersion)
		assert.NotContains(t, listed, upstreamDeniedVersion)
	})
}

func TestUpstreamMirrorDownload(t *testing.T) {
	hostname := "registry.terraform.io"
	fileName := fmt.Sprintf("terraform-provider-%s_%s_%s_%s.zip", upstreamProvider, upstreamMirrorVersion, runtime.GOOS, runtime.GOARCH)

	t.Run("version document lists upstream packages relative to itself", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/hashicorp/%s/%s.json", hostname, upstreamProvider, upstreamMirrorVersion), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		archives, ok := body["archives"].(map[string]any)
		require.True(t, ok)
		require.Contains(t, archives, platformKey())

		archive, _ := archives[platformKey()].(map[string]any)
		url, _ := archive["url"].(string)
		assert.True(t, strings.HasPrefix(url, fileName+"?token="), "expected a relative link with a token, got %q", url)

		hashes, _ := archive["hashes"].([]any)
		require.Len(t, hashes, 1)
		hash, _ := hashes[0].(string)
		assert.True(t, strings.HasPrefix(hash, "zh:"))
	})

	t.Run("the listed link downloads the package without credentials", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/hashicorp/%s/%s.json", hostname, upstreamProvider, upstreamMirrorVersion), nil)
		body := readJSON(t, resp)
		archives, _ := body["archives"].(map[string]any)
		archive, _ := archives[platformKey()].(map[string]any)
		link, _ := archive["url"].(string)
		require.True(t, strings.HasPrefix(link, fileName+"?token="))

		download, err := noRedirectClient().Get(apiURL("/providers/%s/hashicorp/%s/%s", hostname, upstreamProvider, link))
		require.NoError(t, err)
		defer download.Body.Close()

		require.Equal(t, http.StatusFound, download.StatusCode)
		assert.NotEmpty(t, download.Header.Get("Location"))
	})

	t.Run("denied version has no document", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/hashicorp/%s/%s.json", hostname, upstreamProvider, upstreamDeniedVersion), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("first download fetches the package and redirects to storage", func(t *testing.T) {
		resp := doAuthRequestNoRedirect(t, apiURL("/providers/%s/hashicorp/%s/%s", hostname, upstreamProvider, fileName))
		defer resp.Body.Close()

		require.Equal(t, http.StatusFound, resp.StatusCode)
		location := resp.Header.Get("Location")
		require.NotEmpty(t, location)

		download, err := httpClient().Get(location)
		require.NoError(t, err)
		defer download.Body.Close()
		assert.Equal(t, http.StatusOK, download.StatusCode)
	})

	t.Run("fetched package is then served from storage", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/hashicorp/%s/%s.json", hostname, upstreamProvider, upstreamMirrorVersion), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		archives, _ := body["archives"].(map[string]any)
		archive, _ := archives[platformKey()].(map[string]any)
		url, _ := archive["url"].(string)
		assert.True(t, strings.HasPrefix(url, "http"), "expected an absolute storage URL, got %q", url)
		assert.NotContains(t, url, "/providers/"+hostname)
	})

	t.Run("denied package cannot be downloaded", func(t *testing.T) {
		denied := fmt.Sprintf("terraform-provider-%s_%s_%s_%s.zip", upstreamProvider, upstreamDeniedVersion, runtime.GOOS, runtime.GOARCH)
		resp := doAuthRequest(t, http.MethodGet, apiURL("/providers/%s/hashicorp/%s/%s", hostname, upstreamProvider, denied), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestUpstreamRegistryDownload(t *testing.T) {
	resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/providers/hashicorp/%s/%s/download/%s/%s", upstreamProvider, upstreamRegistryVersion, runtime.GOOS, runtime.GOARCH), nil)
	body := readJSON(t, resp)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, body, "shasum")
	assert.Contains(t, body, "shasums_url")
	assert.Contains(t, body, "shasums_signature_url")

	keys, _ := body["signing_keys"].(map[string]any)
	gpgKeys, _ := keys["gpg_public_keys"].([]any)
	require.NotEmpty(t, gpgKeys, "the upstream signing keys are served with the version")

	downloadURL, _ := body["download_url"].(string)
	expected := fmt.Sprintf("/providers/%s/hashicorp/%s/terraform-provider-%s_%s_%s_%s.zip?token=", mirrorHostname(t), upstreamProvider, upstreamProvider, upstreamRegistryVersion, runtime.GOOS, runtime.GOARCH)
	require.True(t, strings.Contains(downloadURL, expected), "expected the download to point at the mirror route with a token, got %q", downloadURL)

	download, err := noRedirectClient().Get(downloadURL)
	require.NoError(t, err)
	defer download.Body.Close()
	assert.Equal(t, http.StatusFound, download.StatusCode, "the link downloads without credentials")
}

func TestUpstreamFetch(t *testing.T) {
	url := apiURL("/v1/api/providers/hashicorp/%s/%s/fetch", upstreamProvider, upstreamFetchVersion)

	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodPost, url)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("without platforms", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodPost, url, map[string]any{"platforms": []string{}})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("fetches the requested platforms", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodPost, url, map[string]any{"platforms": []string{platformKey(), "plan9_mips"}})
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		results, ok := body["results"].([]any)
		require.True(t, ok)
		require.Len(t, results, 2)

		fetched, _ := results[0].(map[string]any)
		assert.Equal(t, platformKey(), fetched["platform"])
		assert.Empty(t, fetched["error"])

		unknown, _ := results[1].(map[string]any)
		assert.NotEmpty(t, unknown["error"])

		document := doAuthRequest(t, http.MethodGet, apiURL("/providers/registry.terraform.io/hashicorp/%s/%s.json", upstreamProvider, upstreamFetchVersion), nil)
		doc := readJSON(t, document)
		archives, _ := doc["archives"].(map[string]any)
		archive, _ := archives[platformKey()].(map[string]any)
		archiveURL, _ := archive["url"].(string)
		assert.True(t, strings.HasPrefix(archiveURL, "http"), "expected the fetched package to be served from storage, got %q", archiveURL)
	})
}

func TestUpstreamRules(t *testing.T) {
	rulesURL := apiURL("/v1/api/authorities/%s/rules", bootstrap.HashicorpAuthorityID)

	t.Run("invalid rule", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodPost, rulesURL, map[string]string{"kind": "bucket", "name": "*", "version": "*", "effect": "deny"})
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("a deny rule hides a version until it is removed", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodPost, rulesURL, map[string]string{"kind": "provider", "name": upstreamProvider, "version": upstreamListedVersion, "effect": "deny"})
		rule := readJSON(t, resp)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		ruleID, _ := rule["id"].(string)
		require.NotEmpty(t, ruleID)

		_, versions := mirrorIndex(t, "registry.terraform.io", "hashicorp", upstreamProvider)
		assert.NotContains(t, versions, upstreamListedVersion)

		remove := doAuthRequest(t, http.MethodDelete, rulesURL+"/"+ruleID, nil)
		remove.Body.Close()
		require.Equal(t, http.StatusOK, remove.StatusCode)

		_, versions = mirrorIndex(t, "registry.terraform.io", "hashicorp", upstreamProvider)
		assert.Contains(t, versions, upstreamListedVersion)
	})

	t.Run("unknown rule", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodDelete, rulesURL+"/00000000-0000-0000-0000-000000000000", nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestUpstreamAutoCreate(t *testing.T) {
	t.Run("an allowlisted upstream namespace is created on first request", func(t *testing.T) {
		status, versions := mirrorIndex(t, "registry.terraform.io", "integrations", "github")

		require.Equal(t, http.StatusOK, status)
		assert.NotEmpty(t, versions)

		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/api/authorities"), nil)
		body, err := readJSONArray(resp)
		require.NoError(t, err)

		var found map[string]any
		for _, a := range body {
			if a["name"] == "integrations" {
				found = a
			}
		}
		require.NotNil(t, found, "expected the integrations authority to be created")
		assert.Equal(t, "registry.terraform.io", found["upstream_hostname"])
		assert.Equal(t, true, found["upstream_enabled"])
	})
}

func TestTerraformUpstreamInit(t *testing.T) {
	requireTerraformCapable(t)
	host := registryHost(t)

	config := fmt.Sprintf(`
terraform {
  required_providers {
    random = {
      source  = "hashicorp/%s"
      version = "%s"
    }
  }
}

resource "random_pet" "test" {}
`, upstreamProvider, upstreamInitVersion)

	for _, run := range []string{"first init fetches from the upstream", "second init is served from storage"} {
		t.Run(run, func(t *testing.T) {
			dir := setupTerraformMirrorProject(t, host, config)

			out := runTerraform(t, dir, "init")
			assert.Contains(t, out, fmt.Sprintf("Installed hashicorp/%s v%s", upstreamProvider, upstreamInitVersion))
			assert.Contains(t, out, "Terraform has been successfully initialized")

			lock, err := os.ReadFile(filepath.Join(dir, ".terraform.lock.hcl"))
			require.NoError(t, err)
			assert.Contains(t, string(lock), fmt.Sprintf(`provider "registry.terraform.io/hashicorp/%s"`, upstreamProvider))
		})
	}
}
