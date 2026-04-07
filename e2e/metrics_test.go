package e2e

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsEndpoint(t *testing.T) {
	resp, err := httpClient().Get(config.MetricsURL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	content := string(body)
	assert.Contains(t, content, "terralist_uptime_seconds")
	assert.Contains(t, content, "terralist_build_info")
	assert.Contains(t, content, "http_requests_total")
}

func TestMetricsNotOnMainPort(t *testing.T) {
	if config.MetricsURL == config.URL {
		t.Skip("Metrics served on main port, skipping isolation test")
	}

	resp, err := httpClient().Get(config.URL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
