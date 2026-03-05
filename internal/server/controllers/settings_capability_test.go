package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"terralist/internal/server/handlers"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/rbac"
	"terralist/pkg/session"
	"terralist/pkg/session/cookie"

	"github.com/gin-gonic/gin"
)

func TestSettingsCapabilityController_AllowsByPolicy(t *testing.T) {
	jwtManager, authz, store := testSettingsDeps(t, `
p, alice@example.com, settings, get, page, allow
`)

	controller := &DefaultSettingsCapabilityController{
		Authentication: &handlers.Authentication{
			JWT:   jwtManager,
			Store: store,
		},
		Authorization:   authz,
		AuthorizedUsers: "admin@example.com",
	}

	router := ginRouterWithController(controller)

	token, err := jwtManager.Build(auth.User{
		Name:  "alice",
		Email: "alice@example.com",
	}, 3600)
	if err != nil {
		t.Fatalf("failed to build jwt token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auth/capabilities/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body struct {
		Allowed bool `json:"allowed"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}

	if !body.Allowed {
		t.Fatalf("expected allowed=true")
	}
}

func TestSettingsCapabilityController_DeniesReadonlyDefault(t *testing.T) {
	jwtManager, authz, store := testSettingsDeps(t, "")

	controller := &DefaultSettingsCapabilityController{
		Authentication: &handlers.Authentication{
			JWT:   jwtManager,
			Store: store,
		},
		Authorization:   authz,
		AuthorizedUsers: "alice@example.com",
	}

	router := ginRouterWithController(controller)

	token, err := jwtManager.Build(auth.User{
		Name:  "bob",
		Email: "bob@example.com",
	}, 3600)
	if err != nil {
		t.Fatalf("failed to build jwt token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auth/capabilities/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body struct {
		Allowed bool `json:"allowed"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}

	if body.Allowed {
		t.Fatalf("expected allowed=false")
	}
}

func TestSettingsCapabilityController_RequiresAuthentication(t *testing.T) {
	jwtManager, authz, store := testSettingsDeps(t, "")

	controller := &DefaultSettingsCapabilityController{
		Authentication: &handlers.Authentication{
			JWT:   jwtManager,
			Store: store,
		},
		Authorization: authz,
	}

	router := ginRouterWithController(controller)

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auth/capabilities/settings", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestSettingsCapabilityController_AllowsByAuthorizedUsers(t *testing.T) {
	jwtManager, authz, store := testSettingsDeps(t, "")

	controller := &DefaultSettingsCapabilityController{
		Authentication: &handlers.Authentication{
			JWT:   jwtManager,
			Store: store,
		},
		Authorization:   authz,
		AuthorizedUsers: "alice,alice@example.com",
	}

	router := ginRouterWithController(controller)

	token, err := jwtManager.Build(auth.User{
		Name:  "alice",
		Email: "alice@example.com",
	}, 3600)
	if err != nil {
		t.Fatalf("failed to build jwt token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/api/auth/capabilities/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body struct {
		Allowed bool `json:"allowed"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}

	if !body.Allowed {
		t.Fatalf("expected allowed=true")
	}
}

func testSettingsDeps(t *testing.T, policy string) (jwt.JWT, *handlers.Authorization, session.Store) {
	t.Helper()

	jwtManager, err := jwt.New("settings-capability-secret")
	if err != nil {
		t.Fatalf("failed to create jwt manager: %v", err)
	}

	policyPath := filepath.Join(t.TempDir(), "policy.csv")
	if err := os.WriteFile(policyPath, []byte(policy), 0600); err != nil {
		t.Fatalf("failed to write policy: %v", err)
	}

	enforcer, err := rbac.NewEnforcer(policyPath, "readonly")
	if err != nil {
		t.Fatalf("failed to create enforcer: %v", err)
	}

	store, err := (&cookie.Creator{}).New(&cookie.Config{
		Secret: "settings-cookie-secret",
	})
	if err != nil {
		t.Fatalf("failed to create cookie store: %v", err)
	}

	return jwtManager, &handlers.Authorization{
		Enforcer: enforcer,
	}, store
}

func ginRouterWithController(controller *DefaultSettingsCapabilityController) *gin.Engine {
	router := gin.New()
	apiGroup := api.NewRouterGroup(router, &api.RouterGroupOptions{
		Prefix: "/v1",
	})
	apiGroup.Register(controller)
	return router
}
