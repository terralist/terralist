package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/apikey"
	"terralist/internal/server/services"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/rbac"
	"terralist/pkg/session/cookie"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"
)

// setupApiKeyRouter creates a gin test router with the API key controller registered.
// When user is non-nil, a middleware injects the user into the context before
// the controller's authentication middleware runs, simulating an authenticated request.
func setupApiKeyRouter(
	t *testing.T,
	user *auth.User,
	policyCSV string,
) (*gin.Engine, *services.MockStandaloneApiKeyService, *services.MockAuthorityService) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	mockService := services.NewMockStandaloneApiKeyService(t)
	mockAuthorityService := services.NewMockAuthorityService(t)

	enforcer, err := rbac.NewEnforcer("", "readonly")
	if err != nil && policyCSV == "" {
		t.Fatalf("failed to create enforcer: %v", err)
	}

	if policyCSV != "" {
		enforcer, err = rbac.NewEnforcerFromString(policyCSV, "readonly")
		if err != nil {
			t.Fatalf("failed to create enforcer from policy: %v", err)
		}
	}

	jwtManager, err := jwt.New("test-signing-secret")
	if err != nil {
		t.Fatalf("failed to create JWT manager: %v", err)
	}

	store, err := (&cookie.Creator{}).New(&cookie.Config{
		Name:   "test-session",
		Secret: "test-secret",
	})
	if err != nil {
		t.Fatalf("failed to create session store: %v", err)
	}

	authentication := &handlers.Authentication{
		JWT:   jwtManager,
		Store: store,
	}
	authorization := &handlers.Authorization{
		Enforcer:         enforcer,
		AuthorityService: mockAuthorityService,
	}

	controller := &DefaultApiKeyController{
		Service:        mockService,
		Authentication: authentication,
		Authorization:  authorization,
	}

	router := gin.New()
	group := router.Group("/v1")
	paths := controller.Paths()
	groups := make([]*gin.RouterGroup, len(paths))
	for i, p := range paths {
		groups[i] = group.Group(p)
	}

	// Inject user before controller middleware runs
	for _, g := range groups {
		if user != nil {
			u := user
			g.Use(func(ctx *gin.Context) {
				ctx.Set("user", u)
				ctx.Set("userName", u.Name)
				ctx.Set("userEmail", u.Email)
			})
		}
	}

	controller.Subscribe(groups...)

	return router, mockService, mockAuthorityService
}

func TestApiKeyController_List(t *testing.T) {
	Convey("Subject: Listing API keys", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}

		Convey("Given an authenticated user with get permission on api-keys", func() {
			policy := `p, test-user, api-keys, get, *, allow`
			router, mockService, _ := setupApiKeyRouter(t, user, policy)

			mockService.On("List").Return([]apikey.ApiKeyDTO{
				{ID: "key-1", Name: "ci-key", CreatedBy: "test@example.com"},
			}, nil)

			Convey("When GET /api/api-keys is called", func() {
				req := httptest.NewRequest(http.MethodGet, "/v1/api/api-keys/", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Convey("Then it should return 200 with the keys", func() {
					So(w.Code, ShouldEqual, http.StatusOK)

					var body []apikey.ApiKeyDTO
					err := json.Unmarshal(w.Body.Bytes(), &body)
					So(err, ShouldBeNil)
					So(body, ShouldHaveLength, 1)
					So(body[0].Name, ShouldEqual, "ci-key")
				})
			})
		})

		Convey("Given an authenticated user with only readonly role (no explicit api-keys policy)", func() {
			router, mockService, _ := setupApiKeyRouter(t, user, "")

			mockService.On("List").Return([]apikey.ApiKeyDTO{
				{ID: "key-1", Name: "ci-key", CreatedBy: "someone@example.com"},
			}, nil)

			Convey("When GET /api/api-keys is called", func() {
				req := httptest.NewRequest(http.MethodGet, "/v1/api/api-keys/", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Convey("Then it should return 200 with an empty list (all filtered out)", func() {
					So(w.Code, ShouldEqual, http.StatusOK)

					var body []apikey.ApiKeyDTO
					err := json.Unmarshal(w.Body.Bytes(), &body)
					So(err, ShouldBeNil)
					So(body, ShouldHaveLength, 0)
				})
			})
		})
	})
}

func TestApiKeyController_Create(t *testing.T) {
	Convey("Subject: Creating an API key", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}

		Convey("Given an authenticated user with create permission", func() {
			policy := `p, test-user, api-keys, create, *, allow`
			router, mockService, _ := setupApiKeyRouter(t, user, policy)

			Convey("When POST /api/api-keys is called with valid body", func() {
				body := apikey.CreateApiKeyDTO{
					Name:     "ci-key",
					ExpireIn: 24,
					Policies: []apikey.CreatePolicyDTO{
						{Resource: "modules", Action: "get", Object: "*", Effect: "allow"},
					},
				}
				jsonBody, _ := json.Marshal(body)

				mockService.On("Create",
					"ci-key",
					"test@example.com",
					24,
					[]apikey.Policy{
						{Resource: "modules", Action: "get", Object: "*", Effect: "allow"},
					},
				).Return("generated-uuid", nil)

				req := httptest.NewRequest(http.MethodPost, "/v1/api/api-keys/", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Convey("Then it should return 201 with the key ID", func() {
					So(w.Code, ShouldEqual, http.StatusCreated)

					var result map[string]string
					err := json.Unmarshal(w.Body.Bytes(), &result)
					So(err, ShouldBeNil)
					So(result["id"], ShouldEqual, "generated-uuid")
					So(result["name"], ShouldEqual, "ci-key")
				})
			})
		})

		Convey("Given an authenticated user without create permission", func() {
			// readonly role doesn't grant create
			router, _, _ := setupApiKeyRouter(t, user, "")

			Convey("When POST /api/api-keys is called", func() {
				body := apikey.CreateApiKeyDTO{
					Name:     "ci-key",
					Policies: []apikey.CreatePolicyDTO{
						{Resource: "modules", Action: "get", Object: "*", Effect: "allow"},
					},
				}
				jsonBody, _ := json.Marshal(body)

				req := httptest.NewRequest(http.MethodPost, "/v1/api/api-keys/", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Convey("Then it should return 403", func() {
					So(w.Code, ShouldEqual, http.StatusForbidden)
				})
			})
		})

		Convey("Given an unauthenticated request", func() {
			router, _, _ := setupApiKeyRouter(t, nil, "")

			Convey("When POST /api/api-keys is called", func() {
				body := `{"name":"ci-key","policies":[{"resource":"modules","action":"get","object":"*","effect":"allow"}]}`
				req := httptest.NewRequest(http.MethodPost, "/v1/api/api-keys/", bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Convey("Then it should return 401", func() {
					So(w.Code, ShouldEqual, http.StatusUnauthorized)
				})
			})
		})
	})
}

func TestApiKeyController_Delete(t *testing.T) {
	Convey("Subject: Deleting an API key", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}

		Convey("Given an authenticated user with delete permission", func() {
			policy := `p, test-user, api-keys, delete, *, allow`
			router, mockService, _ := setupApiKeyRouter(t, user, policy)

			mockService.On("Delete", "some-key-id").Return(nil)

			Convey("When DELETE /api/api-keys/:id is called", func() {
				req := httptest.NewRequest(http.MethodDelete, "/v1/api/api-keys/some-key-id", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Convey("Then it should return 200", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})
		})

		Convey("Given an authenticated user without delete permission", func() {
			router, _, _ := setupApiKeyRouter(t, user, "")

			Convey("When DELETE /api/api-keys/:id is called", func() {
				req := httptest.NewRequest(http.MethodDelete, "/v1/api/api-keys/some-key-id", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				Convey("Then it should return 403", func() {
					So(w.Code, ShouldEqual, http.StatusForbidden)
				})
			})
		})
	})
}
