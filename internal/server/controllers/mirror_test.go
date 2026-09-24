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
	"terralist/internal/server/repositories"
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
	autoCreate ...string,
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
		AutoCreate:    autoCreate,
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
			mockService.On("ListMirrorVersions", "hashicorp", "null", false).Return(versions, nil)

			Convey("When the versions are requested", func() {
				w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

				Convey("Then it should return the versions", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})
		})

		Convey("Given anonymous read is enabled and no user", func() {
			router, mockService, _ := setupMirrorRouter(t, nil, true, false)
			mockService.On("ListMirrorVersions", "hashicorp", "null", false).Return(versions, nil)

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
				mockService.On("ListMirrorVersions", "hashicorp", "null", false).Return(versions, nil)

				Convey("When the versions are requested", func() {
					w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

					Convey("Then it should return the versions", func() {
						So(w.Code, ShouldEqual, http.StatusOK)
						So(w.Body.String(), ShouldEqual, `{"versions":{"3.2.4":{}}}`)
					})
				})
			})

			Convey("If the provider does not exist", func() {
				mockService.On("ListMirrorVersions", "hashicorp", "null", false).Return(nil, errors.New("not found"))

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
					mockService.AssertNotCalled(t, "ListMirrorVersions", mock.Anything, mock.Anything, mock.Anything)
				})
			})

			Convey("When the versions are requested under an upstream hostname an authority stands for", func() {
				mockAuthorityService.
					On("GetByUpstream", "registry.terraform.io", "hashicorp").
					Return(&authority.Authority{Name: "hashicorp-mirror"}, nil)
				mockAuthorityService.On("GetByName", "hashicorp-mirror").Return(&authority.Authority{}, nil).Maybe()
				mockService.On("ListMirrorVersions", "hashicorp-mirror", "null", false).Return(versions, nil)

				w := serve(router, httptest.NewRequest(http.MethodGet, "/providers/registry.terraform.io/hashicorp/null/index.json", nil))

				Convey("Then it should return the versions of the authority's provider", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					So(w.Body.String(), ShouldEqual, `{"versions":{"3.2.4":{}}}`)
				})
			})

			Convey("When the hostname differs only in case", func() {
				mockService.On("ListMirrorVersions", "hashicorp", "null", false).Return(versions, nil)

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
				On("ListMirrorArchives", "hashicorp", "null", "3.2.4", false).
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
				mockService.AssertNotCalled(t, "ListMirrorArchives", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
		})

		Convey("If the version does not exist", func() {
			mockService.
				On("ListMirrorArchives", "hashicorp", "null", "9.9.9", false).
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
				On("ListMirrorArchives", "hashicorp-mirror", "null", "3.2.4", false).
				Return(&provider.MirrorArchivesDTO{Archives: map[string]provider.MirrorArchiveDTO{}}, nil)

			w := serve(router, httptest.NewRequest(http.MethodGet, "/providers/registry.terraform.io/hashicorp/null/3.2.4.json", nil))

			Convey("Then it should return the archives of the authority's provider", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}

func TestMirrorController_DownloadArchive(t *testing.T) {
	Convey("Subject: Downloading a provider package through the network mirror", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}
		base := "/providers/" + mirrorTestHostname + "/hashicorp/null/"

		Convey("Given a readonly user", func() {
			router, mockService, _ := setupMirrorRouter(t, user, false, false)

			Convey("When a stored package is requested", func() {
				mockService.On("Download", "hashicorp", "null", "3.2.4", "linux", "amd64", false).Return("https://storage.example.com/pkg.zip", nil)

				w := serve(router, httptest.NewRequest(http.MethodGet, base+"terraform-provider-null_3.2.4_linux_amd64.zip", nil))

				Convey("Then it should redirect to the storage", func() {
					So(w.Code, ShouldEqual, http.StatusFound)
					So(w.Header().Get("Location"), ShouldEqual, "https://storage.example.com/pkg.zip")
				})
			})

			Convey("When a package that would need fetching is requested", func() {
				mockService.On("Download", "hashicorp", "null", "3.2.4", "darwin", "arm64", false).Return("", services.ErrFetchRequiresCreate)

				w := serve(router, httptest.NewRequest(http.MethodGet, base+"terraform-provider-null_3.2.4_darwin_arm64.zip", nil))

				Convey("Then it should be forbidden", func() {
					So(w.Code, ShouldEqual, http.StatusForbidden)
				})
			})

			Convey("When an unknown package is requested", func() {
				mockService.On("Download", "hashicorp", "null", "9.9.9", "linux", "amd64", false).Return("", repositories.ErrNotFound)

				w := serve(router, httptest.NewRequest(http.MethodGet, base+"terraform-provider-null_9.9.9_linux_amd64.zip", nil))

				Convey("Then it should be not found", func() {
					So(w.Code, ShouldEqual, http.StatusNotFound)
				})
			})

			Convey("When the file name is not a provider package", func() {
				w := serve(router, httptest.NewRequest(http.MethodGet, base+"terraform-provider-other_3.2.4_linux_amd64.zip", nil))

				Convey("Then it should be not found and the service not called", func() {
					So(w.Code, ShouldEqual, http.StatusNotFound)
					mockService.AssertNotCalled(t, "Download", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				})
			})
		})

		Convey("Given a user allowed to create the provider", func() {
			router, mockService, _ := setupMirrorRouterWithPolicy(t, user, "p, test-user, providers, create, hashicorp/*, allow")

			Convey("When a package that needs fetching is requested", func() {
				mockService.On("Download", "hashicorp", "null", "3.2.4", "darwin", "arm64", true).Return("https://storage.example.com/darwin.zip", nil)

				w := serve(router, httptest.NewRequest(http.MethodGet, base+"terraform-provider-null_3.2.4_darwin_arm64.zip", nil))

				Convey("Then it should redirect to the storage", func() {
					So(w.Code, ShouldEqual, http.StatusFound)
					So(w.Header().Get("Location"), ShouldEqual, "https://storage.example.com/darwin.zip")
				})
			})

			Convey("When the upstream fetch fails", func() {
				mockService.On("Download", "hashicorp", "null", "3.2.4", "darwin", "arm64", true).Return("", errors.New("upstream down"))

				w := serve(router, httptest.NewRequest(http.MethodGet, base+"terraform-provider-null_3.2.4_darwin_arm64.zip", nil))

				Convey("Then it should report a bad gateway with the reason", func() {
					So(w.Code, ShouldEqual, http.StatusBadGateway)
					So(w.Body.String(), ShouldContainSubstring, "upstream down")
				})
			})

			Convey("When the version document is requested", func() {
				mockService.On("ListMirrorArchives", "hashicorp", "null", "3.2.4", true).Return(&provider.MirrorArchivesDTO{Archives: map[string]provider.MirrorArchiveDTO{}}, nil)

				w := serve(router, httptest.NewRequest(http.MethodGet, base+"3.2.4.json", nil))

				Convey("Then the upstream should be included", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})
		})
	})
}

func TestMirrorController_AutoCreate(t *testing.T) {
	Convey("Subject: Creating an authority on the first request for an allowlisted upstream", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}
		url := "/providers/registry.terraform.io/integrations/github/index.json"
		versions := &provider.MirrorVersionListDTO{Versions: map[string]struct{}{"6.0.0": {}}}

		Convey("Given the hostname is allowlisted and the user is authenticated", func() {
			router, mockService, mockAuthorityService := setupMirrorRouter(t, user, false, false, "registry.terraform.io")
			mockAuthorityService.On("GetByUpstream", "registry.terraform.io", "integrations").Return(nil, errors.New("no authority found"))
			mockAuthorityService.On("GetByName", "integrations").Return(&authority.Authority{}, nil).Maybe()

			var created authority.AuthorityCreateDTO
			mockAuthorityService.
				On("Create", mock.AnythingOfType("authority.AuthorityCreateDTO")).
				Run(func(args mock.Arguments) { created, _ = args.Get(0).(authority.AuthorityCreateDTO) }).
				Return(&authority.AuthorityDTO{Name: "integrations"}, nil)
			mockService.On("ListMirrorVersions", "integrations", "github", false).Return(versions, nil)

			w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

			Convey("Then the authority should be created for the user and the request served", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				So(created.Name, ShouldEqual, "integrations")
				So(created.UpstreamHostname, ShouldEqual, "registry.terraform.io")
				So(created.UpstreamNamespace, ShouldEqual, "integrations")
				So(created.UpstreamEnabled, ShouldBeTrue)
				So(created.Owner, ShouldEqual, "test@example.com")
			})
		})

		Convey("Given the hostname is not allowlisted", func() {
			router, mockService, mockAuthorityService := setupMirrorRouter(t, user, false, false, "registry.opentofu.org")
			mockAuthorityService.On("GetByUpstream", "registry.terraform.io", "integrations").Return(nil, errors.New("no authority found"))

			w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

			Convey("Then it should be not found and nothing created", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
				mockAuthorityService.AssertNotCalled(t, "Create", mock.Anything)
				mockService.AssertNotCalled(t, "ListMirrorVersions", mock.Anything, mock.Anything, mock.Anything)
			})
		})

		Convey("Given the hostname is allowlisted but the request is anonymous", func() {
			router, _, mockAuthorityService := setupMirrorRouter(t, nil, true, false, "registry.terraform.io")
			mockAuthorityService.On("GetByUpstream", "registry.terraform.io", "integrations").Return(nil, errors.New("no authority found"))

			w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

			Convey("Then it should be not found and nothing created", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
				mockAuthorityService.AssertNotCalled(t, "Create", mock.Anything)
			})
		})
	})
}

// setupMirrorRouterWithPolicy is setupMirrorRouter for an authenticated user
// with the given RBAC policy.
func setupMirrorRouterWithPolicy(t *testing.T, user *auth.User, policyCSV string) (*gin.Engine, *services.MockProviderService, *services.MockAuthorityService) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	mockService := services.NewMockProviderService(t)
	mockAuthorityService := services.NewMockAuthorityService(t)
	mockAuthorityService.On("GetByName", "hashicorp").Return(&authority.Authority{}, nil).Maybe()

	enforcer, err := rbac.NewEnforcerFromString(policyCSV, "readonly")
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

	controller := &DefaultMirrorController{
		ProviderService:  mockService,
		AuthorityService: mockAuthorityService,
		Authentication:   &handlers.Authentication{JWT: jwtManager, Store: store},
		Authorization:    &handlers.Authorization{Enforcer: enforcer, AuthorityService: mockAuthorityService},
		Hostname:         mirrorTestHostname,
	}

	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		ctx.Set("user", user)
		ctx.Set("userName", user.Name)
		ctx.Set("userEmail", user.Email)
	})

	api.NewRouterGroup(router, &api.RouterGroupOptions{Prefix: ""}).Register(controller)

	return router, mockService, mockAuthorityService
}
