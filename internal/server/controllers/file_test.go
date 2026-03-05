package controllers

import (
	"bytes"
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

	jwtManager, err := jwt.New(localSecret)
	if err != nil {
		t.Fatalf("could not create jwt manager: %v", err)
	}

	return &DefaultFileServer{
		ModulesResolver: resolver,
		JWT:             jwtManager,
	}
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
