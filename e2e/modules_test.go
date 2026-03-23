package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleListVersions(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodGet, apiURL("/v1/modules/hashicorp/subnets/cidr/versions"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/modules/hashicorp/nonexistent/cidr/versions"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("success", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/modules/hashicorp/subnets/cidr/versions"), nil)
		body := readJSON(t, resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, body, "modules")

		modules := body["modules"].([]any)
		require.Len(t, modules, 1)

		mod := modules[0].(map[string]any)
		versions := mod["versions"].([]any)
		assert.Len(t, versions, 1)
	})
}

func TestModuleDownload(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodGet, apiURL("/v1/modules/hashicorp/subnets/cidr/1.0.0/download"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("not found module", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/modules/hashicorp/nonexistent/cidr/1.0.0/download"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("not found version", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/modules/hashicorp/subnets/cidr/9.9.9/download"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("success", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodGet, apiURL("/v1/modules/hashicorp/subnets/cidr/1.0.0/download"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		assert.NotEmpty(t, resp.Header.Get("X-Terraform-Get"))
	})
}

func TestModuleUpload(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		resp := doUnauthRequest(t, http.MethodPost, apiURL("/v1/api/modules/hashicorp/subnets/cidr/1.1.0/upload"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("no body", func(t *testing.T) {
		resp := doAuthRequest(t, http.MethodPost, apiURL("/v1/api/modules/hashicorp/subnets/cidr/1.1.0/upload"), nil)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("duplicate version", func(t *testing.T) {
		body := map[string]string{
			"download_url": "https://example.com/module.zip",
		}
		resp := doAuthRequest(t, http.MethodPost, apiURL("/v1/api/modules/hashicorp/subnets/cidr/1.0.0/upload"), body)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})
}

func TestModuleCreateAndDelete(t *testing.T) {
	// Upload a new version (reuse the same archive, different version number).
	body := map[string]string{
		"download_url": "https://github.com/hashicorp/terraform-cidr-subnets/archive/refs/tags/v1.0.0.zip",
	}
	resp := doAuthRequest(t, http.MethodPost, apiURL("/v1/api/modules/hashicorp/subnets/cidr/1.1.0/upload"), body)
	result := readJSON(t, resp)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, result["errors"])

	// Verify it appears in the versions list.
	resp = doAuthRequest(t, http.MethodGet, apiURL("/v1/modules/hashicorp/subnets/cidr/versions"), nil)
	versionsBody := readJSON(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	modules := versionsBody["modules"].([]any)
	mod := modules[0].(map[string]any)
	versions := mod["versions"].([]any)
	assert.Len(t, versions, 2)

	// Delete the specific version.
	resp = doAuthRequest(t, http.MethodDelete, apiURL("/v1/api/modules/hashicorp/subnets/cidr/1.1.0/remove"), nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify it's gone.
	resp = doAuthRequest(t, http.MethodGet, apiURL("/v1/modules/hashicorp/subnets/cidr/versions"), nil)
	versionsBody = readJSON(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	modules = versionsBody["modules"].([]any)
	mod = modules[0].(map[string]any)
	versions = mod["versions"].([]any)
	assert.Len(t, versions, 1)
}

func TestModuleDeleteUnauthenticated(t *testing.T) {
	resp := doUnauthRequest(t, http.MethodDelete, apiURL("/v1/api/modules/hashicorp/subnets/cidr/1.0.0/remove"), nil)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
