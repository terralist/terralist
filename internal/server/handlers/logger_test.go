package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func logRequest(t *testing.T, authenticate gin.HandlerFunc) map[string]any {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	previous := log.Logger
	log.Logger = zerolog.New(&buf)
	t.Cleanup(func() { log.Logger = previous })

	router := gin.New()
	router.Use(Logger())
	router.DELETE("/v1/api/modules/acme/vpc/aws/1.0.0/remove", authenticate, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodDelete, "/v1/api/modules/acme/vpc/aws/1.0.0/remove", nil)
	router.ServeHTTP(httptest.NewRecorder(), req)

	var entry map[string]any
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	return entry
}

func TestLogger_IncludesAuthenticatedUser(t *testing.T) {
	entry := logRequest(t, func(c *gin.Context) {
		c.Set("userName", "apikey:1234")
		c.Set("userEmail", "ci@example.com")
	})

	assert.Equal(t, "apikey:1234", entry["user"])
	assert.Equal(t, "ci@example.com", entry["user_email"])
	assert.Equal(t, "DELETE", entry["method"])
}

func TestLogger_OmitsUserForAnonymousRequests(t *testing.T) {
	entry := logRequest(t, func(c *gin.Context) {})

	assert.NotContains(t, entry, "user")
	assert.NotContains(t, entry, "user_email")
	assert.Equal(t, float64(http.StatusOK), entry["status"])
}
