package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"

	"terralist/pkg/api"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/storage"
	storageFactory "terralist/pkg/storage/factory"
	"terralist/pkg/storage/local"

	gjwt "github.com/golang-jwt/jwt"
	"github.com/gin-gonic/gin"
)

const localSecret = "local-secret"

func TestDefaultFileServer_ServesModulesFileWithValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resolver := newLocalResolver(t)
	key := storeTestFile(t, resolver, []byte("artifact content"))

	downloadURL := findFileURL(t, resolver, key)
	token := extractToken(t, downloadURL)

	filesController := newFileServer(t, resolver)
	router := gin.New()
	apiGroup := api.NewRouterGroup(router, &api.RouterGroupOptions{
		Prefix: "/v1",
	})
	apiGroup.Register(filesController)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/files/%s?token=%s", key, token), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if got := w.Body.String(); got != "artifact content" {
		t.Fatalf("unexpected body: %q", got)
	}
}

func TestDefaultFileServer_RejectsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resolver := newLocalResolver(t)
	key := storeTestFile(t, resolver, []byte("artifact content"))

	filesController := newFileServer(t, resolver)
	router := gin.New()
	apiGroup := api.NewRouterGroup(router, &api.RouterGroupOptions{
		Prefix: "/v1",
	})
	apiGroup.Register(filesController)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/files/%s?token=invalid", key), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestDefaultFileServer_RejectsTokenForDifferentFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resolver := newLocalResolver(t)
	keyA := storeTestFile(t, resolver, []byte("artifact-a"))
	keyB, err := resolver.Store(&storage.StoreInput{
		KeyPrefix: "modules/testns/module/testprov",
		FileName:  "artifact-b.zip",
		Reader:    bytes.NewReader([]byte("artifact-b")),
		Size:      int64(len("artifact-b")),
	})
	if err != nil {
		t.Fatalf("could not store second test file: %v", err)
	}

	downloadURL := findFileURL(t, resolver, keyA)
	token := extractToken(t, downloadURL)

	filesController := newFileServer(t, resolver)
	router := gin.New()
	apiGroup := api.NewRouterGroup(router, &api.RouterGroupOptions{
		Prefix: "/v1",
	})
	apiGroup.Register(filesController)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/files/%s?token=%s", keyB, token), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestDefaultFileServer_RejectsTokenWithoutKeyClaim(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resolver := newLocalResolver(t)
	key := storeTestFile(t, resolver, []byte("artifact content"))

	jwtManager, err := jwt.New(localSecret)
	if err != nil {
		t.Fatalf("could not create jwt manager: %v", err)
	}

	// Build a syntactically valid token that does not carry file key identity.
	token, err := jwtManager.Build(nil, 60)
	if err != nil {
		t.Fatalf("could not build token: %v", err)
	}

	filesController := newFileServer(t, resolver)
	router := gin.New()
	apiGroup := api.NewRouterGroup(router, &api.RouterGroupOptions{
		Prefix: "/v1",
	})
	apiGroup.Register(filesController)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/files/%s?token=%s", key, token), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestDefaultFileServer_RejectsModulesTokenForProviderFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resolver := newLocalResolver(t)
	modulesKey := storeTestFile(t, resolver, []byte("module artifact"))
	providerKey, err := resolver.Store(&storage.StoreInput{
		KeyPrefix: "providers/testns/provider/1.0.0",
		FileName:  "provider.zip",
		Reader:    bytes.NewReader([]byte("provider artifact")),
		Size:      int64(len("provider artifact")),
	})
	if err != nil {
		t.Fatalf("could not store provider test file: %v", err)
	}

	modulesURL := findFileURL(t, resolver, modulesKey)
	modulesToken := extractToken(t, modulesURL)

	filesController := &DefaultFileServer{
		ModulesResolver:   resolver,
		ProvidersResolver: resolver,
		JWT:               mustJWT(t, localSecret),
	}
	router := gin.New()
	apiGroup := api.NewRouterGroup(router, &api.RouterGroupOptions{
		Prefix: "/v1",
	})
	apiGroup.Register(filesController)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/files/%s?token=%s", providerKey, modulesToken), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestLocalResolver_TokenContainsObjectKeyClaim(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resolver := newLocalResolver(t)
	key := storeTestFile(t, resolver, []byte("artifact content"))
	downloadURL := findFileURL(t, resolver, key)
	token := extractToken(t, downloadURL)

	claims := &jwt.TokenClaims{}
	parsed, err := gjwt.ParseWithClaims(token, claims, func(token *gjwt.Token) (interface{}, error) {
		return []byte(localSecret), nil
	})
	if err != nil {
		t.Fatalf("expected token to parse: %v", err)
	}
	if !parsed.Valid {
		t.Fatalf("expected token to be valid")
	}

	payload := map[string]string{}
	if err := json.Unmarshal(claims.Data, &payload); err != nil {
		t.Fatalf("expected token data to contain json payload: %v", err)
	}

	if payload["key"] != key {
		t.Fatalf("expected token key claim %q, got %q", key, payload["key"])
	}
}

func newLocalResolver(t *testing.T) storage.Resolver {
	t.Helper()

	homeDir := t.TempDir()
	cfg := &local.Config{
		HomeDirectory:      homeDir,
		RegistryDirectory:  filepath.Join(homeDir, "registry"),
		BaseURL:            "http://localhost:5758",
		FilesEndpoint:      "/v1/files",
		TokenSigningSecret: localSecret,
		LinkExpire:         1,
	}

	resolver, err := storageFactory.NewResolver(storage.LOCAL, cfg)
	if err != nil {
		t.Fatalf("could not create local resolver: %v", err)
	}

	return resolver
}

func newFileServer(t *testing.T, resolver storage.Resolver) *DefaultFileServer {
	t.Helper()

	return &DefaultFileServer{
		ModulesResolver: resolver,
		JWT:             mustJWT(t, localSecret),
	}
}

func mustJWT(t *testing.T, secret string) jwt.JWT {
	t.Helper()
	j, err := jwt.New(secret)
	if err != nil {
		t.Fatalf("could not create jwt manager: %v", err)
	}
	return j
}

func storeTestFile(t *testing.T, resolver storage.Resolver, content []byte) string {
	t.Helper()

	key, err := resolver.Store(&storage.StoreInput{
		KeyPrefix: "modules/testns/module/testprov",
		FileName:  "artifact.zip",
		Reader:    bytes.NewReader(content),
		Size:      int64(len(content)),
	})
	if err != nil {
		t.Fatalf("could not store test file: %v", err)
	}

	return key
}

func findFileURL(t *testing.T, resolver storage.Resolver, key string) string {
	t.Helper()

	url, err := resolver.Find(key)
	if err != nil {
		t.Fatalf("could not find file URL: %v", err)
	}

	return url
}

func extractToken(t *testing.T, rawURL string) string {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("could not parse download url: %v", err)
	}

	token := parsed.Query().Get("token")
	if token == "" {
		t.Fatalf("token missing from url")
	}

	return token
}
