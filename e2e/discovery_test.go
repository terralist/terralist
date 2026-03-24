package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformServiceDiscovery(t *testing.T) {
	resp := doUnauthRequest(t, http.MethodGet, apiURL("/.well-known/terraform.json"))
	body := readJSON(t, resp)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, body, "login.v1")
	assert.Contains(t, body, "modules.v1")
	assert.Contains(t, body, "providers.v1")
}
