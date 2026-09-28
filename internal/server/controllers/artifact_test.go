package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/artifact"
	"terralist/internal/server/models/authority"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/rbac"
	"terralist/pkg/session/cookie"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"
)

func setupArtifactRouter(t *testing.T, policyCSV string, upstreamEnabled bool) (*gin.Engine, *services.MockProviderService, *services.MockModuleService) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	mockProviderService := services.NewMockProviderService(t)
	mockModuleService := services.NewMockModuleService(t)
	mockAuthorityService := services.NewMockAuthorityService(t)
	mockAuthorityService.
		On("GetByName", "hashicorp").
		Return(&authority.Authority{Name: "hashicorp", UpstreamEnabled: upstreamEnabled}, nil).
		Maybe()

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

	controller := &DefaultArtifactController{
		AuthorityService: mockAuthorityService,
		ProviderService:  mockProviderService,
		ModuleService:    mockModuleService,
		Authentication:   &handlers.Authentication{JWT: jwtManager, Store: store},
		Authorization:    &handlers.Authorization{Enforcer: enforcer, AuthorityService: mockAuthorityService},
	}

	user := &auth.User{Name: "test-user", Email: "test@example.com"}
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		ctx.Set("user", user)
		ctx.Set("userName", user.Name)
		ctx.Set("userEmail", user.Email)
	})

	api.NewRouterGroup(router, &api.RouterGroupOptions{Prefix: "/v1"}).Register(controller)

	return router, mockProviderService, mockModuleService
}

func decodeVersions(w *httptest.ResponseRecorder) artifact.Versions {
	var body artifact.Versions
	So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)

	return body
}

func TestArtifactController_ProviderVersions(t *testing.T) {
	Convey("Subject: Listing the versions of a provider for the web UI", t, func() {
		url := "/v1/api/artifacts/hashicorp/null/version"
		versions := []artifact.VersionDetails{
			{Version: "3.2.4", Origin: "manual"},
			{Version: "3.2.5", Origin: "upstream", MirrorOnly: true},
		}

		Convey("Given a user who may only read the provider", func() {
			router, mockProviderService, _ := setupArtifactRouter(t, "", true)
			mockProviderService.On("ListVersions", "hashicorp", "null").Return(versions, nil)

			w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

			Convey("Then the versions should come with their details and no actions", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				body := decodeVersions(w)
				So(body.Versions, ShouldResemble, versions)
				So(body.CanDelete, ShouldBeFalse)
				So(body.CanFetch, ShouldBeFalse)
			})
		})

		Convey("Given a user who may delete and create the provider of an authority pulling through", func() {
			router, mockProviderService, _ := setupArtifactRouter(t, "p, test-user, providers, delete, hashicorp/*, allow\np, test-user, providers, create, hashicorp/*, allow", true)
			mockProviderService.On("ListVersions", "hashicorp", "null").Return(versions, nil)

			body := decodeVersions(serve(router, httptest.NewRequest(http.MethodGet, url, nil)))

			Convey("Then both actions should be offered", func() {
				So(body.CanDelete, ShouldBeTrue)
				So(body.CanFetch, ShouldBeTrue)
			})
		})

		Convey("Given a user who may delete the provider and update its authority pulling through", func() {
			router, mockProviderService, _ := setupArtifactRouter(t, "p, test-user, providers, delete, hashicorp/*, allow\np, test-user, authorities, update, hashicorp, allow", true)
			mockProviderService.On("ListVersions", "hashicorp", "null").Return(versions, nil)

			body := decodeVersions(serve(router, httptest.NewRequest(http.MethodGet, url, nil)))

			Convey("Then blocking pulled versions should be offered", func() {
				So(body.CanBlock, ShouldBeTrue)
			})
		})

		Convey("Given a user who may delete the provider but not update its authority", func() {
			router, mockProviderService, _ := setupArtifactRouter(t, "p, test-user, providers, delete, hashicorp/*, allow", true)
			mockProviderService.On("ListVersions", "hashicorp", "null").Return(versions, nil)

			body := decodeVersions(serve(router, httptest.NewRequest(http.MethodGet, url, nil)))

			Convey("Then blocking should not be offered", func() {
				So(body.CanDelete, ShouldBeTrue)
				So(body.CanBlock, ShouldBeFalse)
			})
		})

		Convey("Given a user who may delete and update an authority not pulling through", func() {
			router, mockProviderService, _ := setupArtifactRouter(t, "p, test-user, providers, delete, hashicorp/*, allow\np, test-user, authorities, update, hashicorp, allow", false)
			mockProviderService.On("ListVersions", "hashicorp", "null").Return(versions, nil)

			body := decodeVersions(serve(router, httptest.NewRequest(http.MethodGet, url, nil)))

			Convey("Then blocking should not be offered", func() {
				So(body.CanBlock, ShouldBeFalse)
			})
		})

		Convey("Given a user who may create the provider of an authority not pulling through", func() {
			router, mockProviderService, _ := setupArtifactRouter(t, "p, test-user, providers, create, hashicorp/*, allow", false)
			mockProviderService.On("ListVersions", "hashicorp", "null").Return(versions, nil)

			body := decodeVersions(serve(router, httptest.NewRequest(http.MethodGet, url, nil)))

			Convey("Then fetching should not be offered", func() {
				So(body.CanFetch, ShouldBeFalse)
			})
		})
	})
}

func TestArtifactController_ModuleVersions(t *testing.T) {
	Convey("Subject: Listing the versions of a module for the web UI", t, func() {
		url := "/v1/api/artifacts/hashicorp/dir/template/version"
		versions := []artifact.VersionDetails{
			{Version: "1.0.0", Origin: "manual"},
			{Version: "1.0.2", Origin: "upstream"},
		}

		Convey("Given a user who may delete and create the module of an authority pulling through", func() {
			router, _, mockModuleService := setupArtifactRouter(t, "p, test-user, modules, delete, hashicorp/*, allow\np, test-user, modules, create, hashicorp/*, allow", true)
			mockModuleService.On("ListVersions", "hashicorp", "dir", "template").Return(versions, nil)

			w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

			Convey("Then the versions should come with their details and both actions", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				body := decodeVersions(w)
				So(body.Versions, ShouldResemble, versions)
				So(body.CanDelete, ShouldBeTrue)
				So(body.CanFetch, ShouldBeTrue)
			})
		})
	})
}
