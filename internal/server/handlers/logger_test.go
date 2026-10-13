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

func logRequest(t *testing.T, method, target string, authenticate gin.HandlerFunc) map[string]any {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	previous := log.Logger
	log.Logger = zerolog.New(&buf)
	t.Cleanup(func() { log.Logger = previous })

	router := gin.New()
	router.Use(Logger())
	router.Any("/*path", authenticate, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(method, target, nil))

	var entry map[string]any
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	return entry
}

func TestLogger_IncludesAuthenticatedUser(t *testing.T) {
	entry := logRequest(t, http.MethodDelete, "/v1/api/modules/acme/vpc/aws/1.0.0/remove", func(c *gin.Context) {
		c.Set("userName", "apikey:1234")
		c.Set("userEmail", "ci@example.com")
	})

	assert.Equal(t, "apikey:1234", entry["user"])
	assert.Equal(t, "ci@example.com", entry["user_email"])
	assert.Equal(t, "DELETE", entry["method"])
}

func TestLogger_OmitsUserForAnonymousRequests(t *testing.T) {
	entry := logRequest(t, http.MethodGet, "/v1/modules/acme/vpc/aws/versions", func(c *gin.Context) {})

	assert.NotContains(t, entry, "user")
	assert.NotContains(t, entry, "user_email")
	assert.Equal(t, float64(http.StatusOK), entry["status"])
}

func TestLogger_RedactsCredentialQueryParams(t *testing.T) {
	tests := map[string]struct {
		target string
		want   string
	}{
		"download token": {
			target: "/v1/modules/acme/vpc/aws/1.0.0/archive?token=eyJhbGciOi.secret&fetch=1",
			want:   "/v1/modules/acme/vpc/aws/1.0.0/archive?token=REDACTED&fetch=1",
		},
		"oauth code": {
			target: "/v1/api/auth/redirect?code=4%2F0Abc&state=xyz",
			want:   "/v1/api/auth/redirect?code=REDACTED&state=xyz",
		},
		"no credentials": {
			target: "/v1/api/modules?page=2",
			want:   "/v1/api/modules?page=2",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			entry := logRequest(t, http.MethodGet, tt.target, func(c *gin.Context) {})

			assert.Equal(t, tt.want, entry["path"])
		})
	}
}
