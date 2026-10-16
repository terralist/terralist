package server

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"terralist/pkg/auth"
	"terralist/pkg/cache/memory"
	"terralist/pkg/database/sqlite"
	"terralist/pkg/session/cookie"
	"terralist/pkg/storage"
	"terralist/pkg/storage/local"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

const openAPISpecPath = "../../docs/api/openapi.yaml"

// samlProvider stands for a SAML identity provider, so that the server
// registers the routes it serves only with SAML.
type samlProvider struct{}

func (samlProvider) Name() string                            { return "saml" }
func (samlProvider) GetAuthorizeUrl(string) string           { return "" }
func (samlProvider) GetUserDetails(string, *auth.User) error { return nil }
func (samlProvider) GetSPMetadata() ([]byte, error)          { return nil, nil }

var ginParam = regexp.MustCompile(`[:*]([A-Za-z]+)`)

// TestOpenAPISpecDocumentsEveryRoute keeps the API reference in step with the
// router: every route the server registers, with every optional feature
// enabled, must be documented, and every documented route must exist.
func TestOpenAPISpecDocumentsEveryRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()

	db, err := (&sqlite.Creator{}).New(&sqlite.Config{Path: filepath.Join(dir, "terralist.db")})
	if err != nil {
		t.Fatalf("could not open the database: %v", err)
	}

	store, err := (&cookie.Creator{}).New(&cookie.Config{Name: "_session", Secret: "secret"})
	if err != nil {
		t.Fatalf("could not create the session store: %v", err)
	}

	c, err := (&memory.Creator{}).New(&memory.Config{SweepInterval: 1})
	if err != nil {
		t.Fatalf("could not create the cache: %v", err)
	}

	localResolver := func(name string) storage.Resolver {
		r, err := (&local.Creator{}).New(&local.Config{
			RegistryDirectory:  filepath.Join(dir, name),
			BaseURL:            "http://localhost:5758",
			FilesEndpoint:      "/v1/files",
			TokenSigningSecret: "secret",
			LinkExpire:         1,
		})
		if err != nil {
			t.Fatalf("could not create the %s resolver: %v", name, err)
		}
		return r
	}

	srv, err := NewServer(UserConfig{
		URL:                     "http://localhost:5758",
		TokenSigningSecret:      "secret",
		OAuthStateSecret:        "0123456789abcdef0123456789abcdef",
		LocalTokenSigningSecret: "secret",
		UpstreamCacheTTL:        "1h",
		UpstreamCacheRetention:  "1h",
	}, Config{
		Database:          db,
		Provider:          samlProvider{},
		ModulesResolver:   localResolver("modules"),
		ProvidersResolver: localResolver("providers"),
		Store:             store,
		Cache:             c,
	})
	if err != nil {
		t.Fatalf("could not create the server: %v", err)
	}

	var registered []string
	for _, route := range srv.Router.Routes() {
		registered = append(registered, route.Method+" "+ginParam.ReplaceAllString(route.Path, "{$1}"))
	}

	documented := documentedRoutes(t)

	for _, route := range registered {
		if !slices.Contains(documented, route) {
			t.Errorf("route %s is not documented in %s", route, openAPISpecPath)
		}
	}

	for _, route := range documented {
		if !slices.Contains(registered, route) {
			t.Errorf("%s documents route %s, which the server does not register", openAPISpecPath, route)
		}
	}
}

// documentedRoutes lists the operations of the OpenAPI specification as
// "METHOD path".
func documentedRoutes(t *testing.T) []string {
	t.Helper()

	data, err := os.ReadFile(openAPISpecPath)
	if err != nil {
		t.Fatalf("could not read the OpenAPI specification: %v", err)
	}

	var spec struct {
		Paths map[string]map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		t.Fatalf("could not parse the OpenAPI specification: %v", err)
	}

	methods := []string{"get", "put", "post", "delete", "options", "head", "patch", "trace"}

	var routes []string
	for path, item := range spec.Paths {
		for key := range item {
			if slices.Contains(methods, key) {
				routes = append(routes, strings.ToUpper(key)+" "+path)
			}
		}
	}

	return routes
}
