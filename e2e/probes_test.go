package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthProbe(t *testing.T) {
	resp := doUnauthRequest(t, http.MethodGet, apiURL("/check/healthz"))
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadinessProbe(t *testing.T) {
	resp := doUnauthRequest(t, http.MethodGet, apiURL("/check/readyz"))
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
