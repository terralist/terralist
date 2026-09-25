package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/module"
	"terralist/internal/server/repositories"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/rbac"
	"terralist/pkg/registry"
	"terralist/pkg/session/cookie"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

const moduleArchiveBase = "https://terralist.example.com/v1/modules"

// setupModuleRouter creates a gin test router with the module controller
// registered under /v1. When user is non-nil, a middleware injects the user
// into the context to simulate an authenticated request.
func setupModuleRouter(t *testing.T, user *auth.User, policyCSV string) (*gin.Engine, *services.MockModuleService) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	mockService := services.NewMockModuleService(t)
	mockAuthorityService := services.NewMockAuthorityService(t)
	mockAuthorityService.On("GetByName", "hashicorp").Return(&authority.Authority{Name: "hashicorp"}, nil).Maybe()

	enforcer, err := rbac.NewEnforcer("", "readonly")
	if policyCSV != "" {
		enforcer, err = rbac.NewEnforcerFromString(policyCSV, "readonly")
	}
	if err != nil {
		t.Fatalf("failed to create enforcer: %v", err)
	}

	jwtManager, err := jwt.New("test-signing-secret")
	if err != nil {
		t.Fatalf("failed to create JWT manager: %v", err)
	}

	store, err := (&cookie.Creator{}).New(&cookie.Config{Name: "test-session", Secret: "test-secret"})
	if err != nil {
		t.Fatalf("failed to create session store: %v", err)
	}

	tokens, err := handlers.NewDownloadTokens("test-signing-secret")
	if err != nil {
		t.Fatalf("failed to create download tokens: %v", err)
	}

	controller := &DefaultModuleController{
		ModuleService:    mockService,
		AuthorityService: mockAuthorityService,
		Authentication:   &handlers.Authentication{JWT: jwtManager, Store: store},
		Authorization:    &handlers.Authorization{Enforcer: enforcer, AuthorityService: mockAuthorityService},
		Tokens:           tokens,
		ArchiveBaseURL:   moduleArchiveBase,
	}

	router := gin.New()
	if user != nil {
		router.Use(func(ctx *gin.Context) {
			ctx.Set("user", user)
			ctx.Set("userName", user.Name)
			ctx.Set("userEmail", user.Email)
		})
	}

	api.NewRouterGroup(router, &api.RouterGroupOptions{Prefix: "/v1"}).Register(controller)

	return router, mockService
}

func moduleTokens(t *testing.T) *handlers.DownloadTokens {
	t.Helper()

	tokens, err := handlers.NewDownloadTokens("test-signing-secret")
	if err != nil {
		t.Fatalf("failed to create download tokens: %v", err)
	}

	return tokens
}

func TestModuleController_UpstreamVersions(t *testing.T) {
	Convey("Subject: Listing module versions with the upstream", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}
		url := "/v1/modules/hashicorp/dir/template/versions"
		dto := &module.ListResponseDTO{Modules: []module.ModuleDTO{{Versions: []module.VersionListDTO{{Version: "1.0.2"}}}}}

		Convey("Given a readonly user", func() {
			router, mockService := setupModuleRouter(t, user, "")
			mockService.On("Get", "hashicorp", "dir", "template", false).Return(dto, nil)

			w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

			Convey("Then the upstream should not be included", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})

		Convey("Given a user allowed to create the module", func() {
			router, mockService := setupModuleRouter(t, user, "p, test-user, modules, create, hashicorp/*, allow")
			mockService.On("Get", "hashicorp", "dir", "template", true).Return(dto, nil)

			w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

			Convey("Then the upstream should be included", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}

func TestModuleController_Download(t *testing.T) {
	Convey("Subject: The download location of a module version", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}
		url := "/v1/modules/hashicorp/dir/template/1.0.2/download"
		router, mockService := setupModuleRouter(t, user, "p, test-user, modules, create, hashicorp/*, allow")

		Convey("When the version is stored", func() {
			location := "https://storage.example.com/1.0.2.zip"
			mockService.On("GetVersionURL", "hashicorp", "dir", "template", "1.0.2", true).Return(&location, nil)

			w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

			Convey("Then the storage location is returned untouched", func() {
				So(w.Code, ShouldEqual, http.StatusNoContent)
				So(w.Header().Get("X-Terraform-Get"), ShouldEqual, location)
			})
		})

		Convey("When the version is offered by the upstream only", func() {
			location := moduleArchiveBase + "/hashicorp/dir/template/1.0.2/archive"
			mockService.On("GetVersionURL", "hashicorp", "dir", "template", "1.0.2", true).Return(&location, nil)

			w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

			Convey("Then the archive location carries a token allowing the fetch", func() {
				So(w.Code, ShouldEqual, http.StatusNoContent)
				got := w.Header().Get("X-Terraform-Get")
				So(got, ShouldStartWith, location+"?token=")

				fetch, ok := moduleTokens(t).Verify(strings.TrimPrefix(got, location+"?token="), module.ArchiveSubject("hashicorp", "dir", "template", "1.0.2"))
				So(ok, ShouldBeTrue)
				So(fetch, ShouldBeTrue)
			})
		})
	})
}

func TestModuleController_Archive(t *testing.T) {
	Convey("Subject: Downloading a module archive through the archive route", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}
		url := "/v1/modules/hashicorp/dir/template/1.0.2/archive"
		subject := module.ArchiveSubject("hashicorp", "dir", "template", "1.0.2")

		Convey("Given no credentials but a valid token allowing the fetch", func() {
			router, mockService := setupModuleRouter(t, nil, "")
			token, _ := moduleTokens(t).Sign(subject, true)
			mockService.On("Download", "hashicorp", "dir", "template", "1.0.2", true).Return("https://storage.example.com/1.0.2.zip", nil)

			w := serve(router, httptest.NewRequest(http.MethodGet, url+"?token="+token, nil))

			Convey("Then the storage location is handed to go-getter", func() {
				So(w.Code, ShouldEqual, http.StatusNoContent)
				So(w.Header().Get("X-Terraform-Get"), ShouldEqual, "https://storage.example.com/1.0.2.zip")
			})
		})

		Convey("Given no credentials and a token for another version", func() {
			router, mockService := setupModuleRouter(t, nil, "")
			token, _ := moduleTokens(t).Sign(module.ArchiveSubject("hashicorp", "dir", "template", "1.0.1"), true)

			w := serve(router, httptest.NewRequest(http.MethodGet, url+"?token="+token, nil))

			Convey("Then it should be forbidden and the service not called", func() {
				So(w.Code, ShouldEqual, http.StatusForbidden)
				mockService.AssertNotCalled(t, "Download", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
		})

		Convey("Given a readonly user", func() {
			router, mockService := setupModuleRouter(t, user, "")
			mockService.On("Download", "hashicorp", "dir", "template", "1.0.2", false).Return("", services.ErrFetchRequiresCreate)

			w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

			Convey("Then fetching should be forbidden", func() {
				So(w.Code, ShouldEqual, http.StatusForbidden)
			})
		})

		Convey("Given a user allowed to create the module", func() {
			router, mockService := setupModuleRouter(t, user, "p, test-user, modules, create, hashicorp/*, allow")

			Convey("When the upstream does not know the version", func() {
				mockService.On("Download", "hashicorp", "dir", "template", "1.0.2", true).Return("", repositories.ErrNotFound)

				w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

				Convey("Then it should be not found", func() {
					So(w.Code, ShouldEqual, http.StatusNotFound)
				})
			})

			Convey("When the fetch fails", func() {
				mockService.On("Download", "hashicorp", "dir", "template", "1.0.2", true).Return("", errors.New("upstream down"))

				w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

				Convey("Then it should report a bad gateway", func() {
					So(w.Code, ShouldEqual, http.StatusBadGateway)
					So(w.Body.String(), ShouldContainSubstring, "upstream down")
				})
			})
		})
	})
}

func TestModuleController_Fetch(t *testing.T) {
	Convey("Subject: Fetching a module version from the upstream on demand", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}
		url := "/v1/api/modules/hashicorp/dir/template/1.0.2/fetch"

		post := func(router *gin.Engine) *httptest.ResponseRecorder {
			req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString("{}"))
			req.Header.Set("Content-Type", "application/json")

			return serve(router, req)
		}

		Convey("Given no authenticated user", func() {
			router, _ := setupModuleRouter(t, nil, "")

			Convey("Then it should be unauthorized", func() {
				So(post(router).Code, ShouldEqual, http.StatusUnauthorized)
			})
		})

		Convey("Given a user without create permission", func() {
			router, _ := setupModuleRouter(t, user, "")

			Convey("Then it should be forbidden", func() {
				So(post(router).Code, ShouldEqual, http.StatusForbidden)
			})
		})

		Convey("Given a user with create permission", func() {
			router, mockService := setupModuleRouter(t, user, "p, test-user, modules, create, hashicorp/*, allow")

			Convey("When the fetch succeeds", func() {
				mockService.On("Fetch", "hashicorp", "dir", "template", "1.0.2").Return(nil)
				w := post(router)

				Convey("Then it should succeed", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					var body map[string][]string
					So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)
					So(body["errors"], ShouldBeEmpty)
				})
			})

			Convey("When the fetch fails", func() {
				mockService.On("Fetch", "hashicorp", "dir", "template", "1.0.2").Return(errors.New("upstream down"))
				w := post(router)

				Convey("Then it should report a bad gateway with the reason", func() {
					So(w.Code, ShouldEqual, http.StatusBadGateway)
					So(w.Body.String(), ShouldContainSubstring, "upstream down")
				})
			})

			Convey("When the version is unknown upstream", func() {
				mockService.On("Fetch", "hashicorp", "dir", "template", "1.0.2").Return(fmt.Errorf("hashicorp/dir/template 1.0.2: %w", registry.ErrNotFound))
				w := post(router)

				Convey("Then it should be not found", func() {
					So(w.Code, ShouldEqual, http.StatusNotFound)
				})
			})

			Convey("When a rule denies the version", func() {
				mockService.On("Fetch", "hashicorp", "dir", "template", "1.0.2").Return(fmt.Errorf("hashicorp/dir/template 1.0.2: %w", services.ErrUpstreamDenied))
				w := post(router)

				Convey("Then it should be not found", func() {
					So(w.Code, ShouldEqual, http.StatusNotFound)
				})
			})
		})
	})
}

func TestModuleController_UpstreamUnavailable(t *testing.T) {
	Convey("Subject: Module registry requests while the upstream is unavailable", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}
		router, mockService := setupModuleRouter(t, user, "p, test-user, modules, create, hashicorp/*, allow")
		unavailable := fmt.Errorf("%w: upstream down", services.ErrUpstreamUnavailable)

		Convey("When the versions are listed", func() {
			mockService.On("Get", "hashicorp", "dir", "template", true).Return(nil, unavailable)
			w := serve(router, httptest.NewRequest(http.MethodGet, "/v1/modules/hashicorp/dir/template/versions", nil))

			Convey("Then it should be a bad gateway", func() {
				So(w.Code, ShouldEqual, http.StatusBadGateway)
			})
		})

		Convey("When the download location is requested", func() {
			mockService.On("GetVersionURL", "hashicorp", "dir", "template", "1.0.2", true).Return(nil, unavailable)
			w := serve(router, httptest.NewRequest(http.MethodGet, "/v1/modules/hashicorp/dir/template/1.0.2/download", nil))

			Convey("Then it should be a bad gateway", func() {
				So(w.Code, ShouldEqual, http.StatusBadGateway)
			})
		})
	})
}
