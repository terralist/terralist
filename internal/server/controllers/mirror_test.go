package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/authority"
	"terralist/internal/server/models/provider"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/rbac"
	"terralist/pkg/session/cookie"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

const mirrorTestHostname = "terralist.example.com"

// setupMirrorRouter creates a gin test router with the mirror controller
// registered at the root. When user is non-nil, a middleware injects the user
// into the context to simulate an authenticated request.
func setupMirrorRouter(
	t *testing.T,
	user *auth.User,
	anonymousRead bool,
	public bool,
) (*gin.Engine, *services.MockProviderService, *services.MockAuthorityService) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	mockService := services.NewMockProviderService(t)
	mockAuthorityService := services.NewMockAuthorityService(t)
	mockAuthorityService.On("GetByName", "hashicorp").Return(&authority.Authority{Public: public}, nil).Maybe()

	enforcer, err := rbac.NewEnforcer("", "readonly")
	if err != nil {
		t.Fatalf("failed to create enforcer: %v", err)
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

	controller := &DefaultMirrorController{
		ProviderService:  mockService,
		AuthorityService: mockAuthorityService,
		Authentication: &handlers.Authentication{
			JWT:   jwtManager,
			Store: store,
		},
		Authorization: &handlers.Authorization{
			Enforcer:         enforcer,
			AuthorityService: mockAuthorityService,
		},
		Hostname:      mirrorTestHostname,
		AnonymousRead: anonymousRead,
	}

	router := gin.New()
	if user != nil {
		router.Use(func(ctx *gin.Context) {
			ctx.Set("user", user)
			ctx.Set("userName", user.Name)
			ctx.Set("userEmail", user.Email)
		})
	}

	api.NewRouterGroup(router, &api.RouterGroupOptions{Prefix: ""}).Register(controller)

	return router, mockService, mockAuthorityService
}

func serve(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

func TestMirrorController_ListVersions(t *testing.T) {
	Convey("Subject: Listing the versions of a provider through the network mirror", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}
		url := "/providers/" + mirrorTestHostname + "/hashicorp/null/index.json"
		versions := &provider.MirrorVersionListDTO{Versions: map[string]struct{}{"3.2.4": {}}}

		Convey("Given anonymous read is disabled and no user", func() {
			router, _, _ := setupMirrorRouter(t, nil, false, false)

			Convey("When the versions are requested", func() {
				w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

				Convey("Then it should be forbidden", func() {
					So(w.Code, ShouldEqual, http.StatusForbidden)
				})
			})
		})

		Convey("Given a public authority and no user", func() {
			router, mockService, _ := setupMirrorRouter(t, nil, false, true)
			mockService.On("ListMirrorVersions", "hashicorp", "null").Return(versions, nil)

			Convey("When the versions are requested", func() {
				w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

				Convey("Then it should return the versions", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})
		})

		Convey("Given anonymous read is enabled and no user", func() {
			router, mockService, _ := setupMirrorRouter(t, nil, true, false)
			mockService.On("ListMirrorVersions", "hashicorp", "null").Return(versions, nil)

			Convey("When the versions are requested", func() {
				w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

				Convey("Then it should return the versions", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					So(w.Body.String(), ShouldEqual, `{"versions":{"3.2.4":{}}}`)
				})
			})
		})

		Convey("Given a readonly user", func() {
			router, mockService, mockAuthorityService := setupMirrorRouter(t, user, false, false)

			Convey("If the provider exists", func() {
				mockService.On("ListMirrorVersions", "hashicorp", "null").Return(versions, nil)

				Convey("When the versions are requested", func() {
					w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

					Convey("Then it should return the versions", func() {
						So(w.Code, ShouldEqual, http.StatusOK)
						So(w.Body.String(), ShouldEqual, `{"versions":{"3.2.4":{}}}`)
					})
				})
			})

			Convey("If the provider does not exist", func() {
				mockService.On("ListMirrorVersions", "hashicorp", "null").Return(nil, errors.New("not found"))

				Convey("When the versions are requested", func() {
					w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

					Convey("Then it should return not found with errors", func() {
						So(w.Code, ShouldEqual, http.StatusNotFound)

						var body map[string][]string
						So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)
						So(body["errors"], ShouldResemble, []string{"not found"})
					})
				})
			})

			Convey("When the versions are requested under an upstream hostname nobody stands for", func() {
				mockAuthorityService.
					On("GetByUpstream", "registry.terraform.io", "hashicorp").
					Return(nil, errors.New("no authority found"))

				w := serve(router, httptest.NewRequest(http.MethodGet, "/providers/registry.terraform.io/hashicorp/null/index.json", nil))

				Convey("Then it should be not found and the service not called", func() {
					So(w.Code, ShouldEqual, http.StatusNotFound)
					mockService.AssertNotCalled(t, "ListMirrorVersions", mock.Anything, mock.Anything)
				})
			})

			Convey("When the versions are requested under an upstream hostname an authority stands for", func() {
				mockAuthorityService.
					On("GetByUpstream", "registry.terraform.io", "hashicorp").
					Return(&authority.Authority{Name: "hashicorp-mirror"}, nil)
				mockAuthorityService.On("GetByName", "hashicorp-mirror").Return(&authority.Authority{}, nil).Maybe()
				mockService.On("ListMirrorVersions", "hashicorp-mirror", "null").Return(versions, nil)

				w := serve(router, httptest.NewRequest(http.MethodGet, "/providers/registry.terraform.io/hashicorp/null/index.json", nil))

				Convey("Then it should return the versions of the authority's provider", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					So(w.Body.String(), ShouldEqual, `{"versions":{"3.2.4":{}}}`)
				})
			})

			Convey("When the hostname differs only in case", func() {
				mockService.On("ListMirrorVersions", "hashicorp", "null").Return(versions, nil)

				w := serve(router, httptest.NewRequest(http.MethodGet, "/providers/Terralist.Example.com/hashicorp/null/index.json", nil))

				Convey("Then it should return the versions", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})
		})
	})
}

func TestMirrorController_ListArchives(t *testing.T) {
	Convey("Subject: Listing the installation packages of a provider version through the network mirror", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}
		router, mockService, mockAuthorityService := setupMirrorRouter(t, user, false, false)
		base := "/providers/" + mirrorTestHostname + "/hashicorp/null/"

		Convey("If the version exists", func() {
			mockService.
				On("ListMirrorArchives", "hashicorp", "null", "3.2.4").
				Return(&provider.MirrorArchivesDTO{
					Archives: map[string]provider.MirrorArchiveDTO{
						"linux_amd64": {URL: "https://storage.example.com/pkg.zip", Hashes: []string{"zh:abc"}},
					},
				}, nil)

			Convey("When the version document is requested", func() {
				w := serve(router, httptest.NewRequest(http.MethodGet, base+"3.2.4.json", nil))

				Convey("Then it should return the archives", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					So(w.Body.String(), ShouldEqual, `{"archives":{"linux_amd64":{"url":"https://storage.example.com/pkg.zip","hashes":["zh:abc"]}}}`)
				})
			})
		})

		Convey("When the version is requested without the json extension", func() {
			w := serve(router, httptest.NewRequest(http.MethodGet, base+"3.2.4", nil))

			Convey("Then it should be not found and the service not called", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
				mockService.AssertNotCalled(t, "ListMirrorArchives", mock.Anything, mock.Anything, mock.Anything)
			})
		})

		Convey("If the version does not exist", func() {
			mockService.
				On("ListMirrorArchives", "hashicorp", "null", "9.9.9").
				Return(nil, errors.New("not found"))

			Convey("When the version document is requested", func() {
				w := serve(router, httptest.NewRequest(http.MethodGet, base+"9.9.9.json", nil))

				Convey("Then it should return not found", func() {
					So(w.Code, ShouldEqual, http.StatusNotFound)
				})
			})
		})

		Convey("When the version document is requested under an upstream hostname an authority stands for", func() {
			mockAuthorityService.
				On("GetByUpstream", "registry.terraform.io", "hashicorp").
				Return(&authority.Authority{Name: "hashicorp-mirror"}, nil)
			mockAuthorityService.On("GetByName", "hashicorp-mirror").Return(&authority.Authority{}, nil).Maybe()
			mockService.
				On("ListMirrorArchives", "hashicorp-mirror", "null", "3.2.4").
				Return(&provider.MirrorArchivesDTO{Archives: map[string]provider.MirrorArchiveDTO{}}, nil)

			w := serve(router, httptest.NewRequest(http.MethodGet, "/providers/registry.terraform.io/hashicorp/null/3.2.4.json", nil))

			Convey("Then it should return the archives of the authority's provider", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}
