package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorityUpdate(t *testing.T) {
	name := fmt.Sprintf("updated-%d", time.Now().UnixNano())

	created := readJSON(t, doAuthRequest(t, http.MethodPost, apiURL("/v1/api/authorities"), map[string]any{
		"name":       name,
		"policy_url": "",
	}))
	id, _ := created["id"].(string)
	require.NotEmpty(t, id)

	keyResp := doAuthRequest(t, http.MethodPost, apiURL("/v1/api/authorities/%s/keys", id), map[string]string{
		"key_id":          "KEEPME",
		"ascii_armor":     "armor",
		"trust_signature": "",
	})
	keyResp.Body.Close()
	require.Equal(t, http.StatusOK, keyResp.StatusCode)

	// An update replaces the whole authority, so it is sent back as read.
	authority := readJSON(t, doAuthRequest(t, http.MethodGet, apiURL("/v1/api/authorities/%s", id), nil))
	authority["public"] = true
	authority["upstream_hostname"] = "registry.terraform.io"

	resp := doAuthRequest(t, http.MethodPatch, apiURL("/v1/api/authorities/%s", id), authority)
	body := readJSON(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode, body)

	updated := readJSON(t, doAuthRequest(t, http.MethodGet, apiURL("/v1/api/authorities/%s", id), nil))
	assert.Equal(t, true, updated["public"])
	assert.Equal(t, "registry.terraform.io", updated["upstream_hostname"])

	keys, _ := updated["keys"].([]any)
	require.Len(t, keys, 1)
	key, _ := keys[0].(map[string]any)
	assert.Equal(t, "KEEPME", key["key_id"])
}
