package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/mirror"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/file"
	"terralist/pkg/rbac"
	"terralist/pkg/session/cookie"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

// setupMirrorRouter creates a gin test router with the mirror controller
// registered at the root. When user is non-nil, a middleware injects the user
// into the context to simulate an authenticated request.
func setupMirrorRouter(
	t *testing.T,
	user *auth.User,
	policyCSV string,
	anonymousRead bool,
) (*gin.Engine, *services.MockMirrorService) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	mockService := services.NewMockMirrorService(t)
	mockAuthorityService := services.NewMockAuthorityService(t)

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

	store, err := (&cookie.Creator{}).New(&cookie.Config{
		Name:   "test-session",
		Secret: "test-secret",
	})
	if err != nil {
		t.Fatalf("failed to create session store: %v", err)
	}

	controller := &DefaultMirrorController{
		MirrorService: mockService,
		Authentication: &handlers.Authentication{
			JWT:   jwtManager,
			Store: store,
		},
		Authorization: &handlers.Authorization{
			Enforcer:         enforcer,
			AuthorityService: mockAuthorityService,
		},
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

	return router, mockService
}

func serve(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

// multipartUpload builds a multipart request body with a metadata part and the
// given archive parts.
func multipartUpload(metadata []byte, archives map[string][]byte) (*bytes.Buffer, string) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if metadata != nil {
		part, _ := writer.CreateFormFile("metadata", "3.2.4.json")
		_, _ = part.Write(metadata)
	}

	for name, content := range archives {
		part, _ := writer.CreateFormFile("archives", name)
		_, _ = part.Write(content)
	}

	_ = writer.Close()

	return &body, writer.FormDataContentType()
}

func TestMirrorController_ListVersions(t *testing.T) {
	Convey("Subject: Listing the versions of a mirrored provider", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}
		url := "/providers/registry.terraform.io/hashicorp/null/index.json"

		Convey("Given anonymous read is disabled and no user", func() {
			router, _ := setupMirrorRouter(t, nil, "", false)

			Convey("When the versions are requested", func() {
				w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

				Convey("Then it should be forbidden", func() {
					So(w.Code, ShouldEqual, http.StatusForbidden)
				})
			})
		})

		Convey("Given anonymous read is enabled and no user", func() {
			router, mockService := setupMirrorRouter(t, nil, "", true)
			mockService.
				On("ListVersions", "registry.terraform.io", "hashicorp", "null").
				Return(&mirror.VersionListDTO{Versions: map[string]struct{}{"3.2.4": {}}}, nil)

			Convey("When the versions are requested", func() {
				w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

				Convey("Then it should return the versions", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					So(w.Body.String(), ShouldEqual, `{"versions":{"3.2.4":{}}}`)
				})
			})
		})

		Convey("Given a readonly user", func() {
			router, mockService := setupMirrorRouter(t, user, "", false)

			Convey("If the provider exists", func() {
				mockService.
					On("ListVersions", "registry.terraform.io", "hashicorp", "null").
					Return(&mirror.VersionListDTO{Versions: map[string]struct{}{"3.2.4": {}}}, nil)

				Convey("When the versions are requested", func() {
					w := serve(router, httptest.NewRequest(http.MethodGet, url, nil))

					Convey("Then it should return the versions", func() {
						So(w.Code, ShouldEqual, http.StatusOK)
						So(w.Body.String(), ShouldEqual, `{"versions":{"3.2.4":{}}}`)
					})
				})
			})

			Convey("If the provider does not exist", func() {
				mockService.
					On("ListVersions", "registry.terraform.io", "hashicorp", "null").
					Return(nil, errors.New("not found"))

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
		})
	})
}

func TestMirrorController_GetVersion(t *testing.T) {
	Convey("Subject: Listing the installation packages of a mirrored provider version", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}
		router, mockService := setupMirrorRouter(t, user, "", false)

		Convey("If the version exists", func() {
			mockService.
				On("GetVersion", "registry.terraform.io", "hashicorp", "null", "3.2.4").
				Return(&mirror.ArchivesDTO{
					Archives: map[string]mirror.ArchiveDTO{
						"linux_amd64": {URL: "https://storage.example.com/pkg.zip", Hashes: []string{"h1:abc"}},
					},
				}, nil)

			Convey("When the version document is requested", func() {
				w := serve(router, httptest.NewRequest(http.MethodGet, "/providers/registry.terraform.io/hashicorp/null/3.2.4.json", nil))

				Convey("Then it should return the archives", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					So(w.Body.String(), ShouldEqual, `{"archives":{"linux_amd64":{"url":"https://storage.example.com/pkg.zip","hashes":["h1:abc"]}}}`)
				})
			})
		})

		Convey("When the version is requested without the json extension", func() {
			w := serve(router, httptest.NewRequest(http.MethodGet, "/providers/registry.terraform.io/hashicorp/null/3.2.4", nil))

			Convey("Then it should be not found and the service not called", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
				mockService.AssertNotCalled(t, "GetVersion", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			})
		})

		Convey("If the version does not exist", func() {
			mockService.
				On("GetVersion", "registry.terraform.io", "hashicorp", "null", "9.9.9").
				Return(nil, errors.New("not found"))

			Convey("When the version document is requested", func() {
				w := serve(router, httptest.NewRequest(http.MethodGet, "/providers/registry.terraform.io/hashicorp/null/9.9.9.json", nil))

				Convey("Then it should return not found", func() {
					So(w.Code, ShouldEqual, http.StatusNotFound)
				})
			})
		})
	})
}

func TestMirrorController_Upload(t *testing.T) {
	Convey("Subject: Uploading mirrored provider packages", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}
		url := "/v1/api/mirror/registry.terraform.io/hashicorp/null/3.2.4/upload"
		metadata := []byte(`{"archives":{"linux_amd64":{"url":"terraform-provider-null_3.2.4_linux_amd64.zip","hashes":["h1:abc"]}}}`)
		archives := map[string][]byte{"terraform-provider-null_3.2.4_linux_amd64.zip": []byte("zip-content")}

		Convey("Given no authenticated user", func() {
			router, _ := setupMirrorRouter(t, nil, "", true)

			Convey("When an upload is requested", func() {
				body, contentType := multipartUpload(metadata, archives)
				req := httptest.NewRequest(http.MethodPost, url, body)
				req.Header.Set("Content-Type", contentType)
				w := serve(router, req)

				Convey("Then it should be unauthorized", func() {
					So(w.Code, ShouldEqual, http.StatusUnauthorized)
				})
			})
		})

		Convey("Given a user without create permission", func() {
			router, _ := setupMirrorRouter(t, user, "", false)

			Convey("When an upload is requested", func() {
				body, contentType := multipartUpload(metadata, archives)
				req := httptest.NewRequest(http.MethodPost, url, body)
				req.Header.Set("Content-Type", contentType)
				w := serve(router, req)

				Convey("Then it should be forbidden", func() {
					So(w.Code, ShouldEqual, http.StatusForbidden)
				})
			})
		})

		Convey("Given a user with create permission", func() {
			router, mockService := setupMirrorRouter(t, user, "p, test-user, mirror, create, registry.terraform.io/*, allow", false)

			Convey("When a valid upload is requested", func() {
				var uploaded []string
				mockService.
					On("Upload", "registry.terraform.io", "hashicorp", "null", "3.2.4", mock.AnythingOfType("mirror.ArchivesDTO"), mock.Anything).
					Run(func(args mock.Arguments) {
						files, _ := args.Get(5).([]file.File)
						for _, f := range files {
							content, _ := io.ReadAll(f)
							uploaded = append(uploaded, f.Name()+"="+string(content))
						}
					}).
					Return(nil)

				body, contentType := multipartUpload(metadata, archives)
				req := httptest.NewRequest(http.MethodPost, url, body)
				req.Header.Set("Content-Type", contentType)
				w := serve(router, req)

				Convey("Then it should succeed and pass the parsed metadata and archives to the service", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					So(uploaded, ShouldResemble, []string{"terraform-provider-null_3.2.4_linux_amd64.zip=zip-content"})

					call := mockService.Calls[0]
					dto, _ := call.Arguments.Get(4).(mirror.ArchivesDTO)
					So(dto.Archives["linux_amd64"].Hashes, ShouldResemble, []string{"h1:abc"})
				})
			})

			Convey("When the service rejects the upload", func() {
				mockService.
					On("Upload", "registry.terraform.io", "hashicorp", "null", "3.2.4", mock.Anything, mock.Anything).
					Return(errors.New("platform linux_amd64 already exists for version 3.2.4"))

				body, contentType := multipartUpload(metadata, archives)
				req := httptest.NewRequest(http.MethodPost, url, body)
				req.Header.Set("Content-Type", contentType)
				w := serve(router, req)

				Convey("Then it should return a conflict with the error", func() {
					So(w.Code, ShouldEqual, http.StatusConflict)
					So(w.Body.String(), ShouldContainSubstring, "already exists")
				})
			})

			Convey("When the metadata part is missing", func() {
				body, contentType := multipartUpload(nil, archives)
				req := httptest.NewRequest(http.MethodPost, url, body)
				req.Header.Set("Content-Type", contentType)
				w := serve(router, req)

				Convey("Then it should be a bad request", func() {
					So(w.Code, ShouldEqual, http.StatusBadRequest)
				})
			})

			Convey("When the metadata is not valid JSON", func() {
				body, contentType := multipartUpload([]byte("not-json"), archives)
				req := httptest.NewRequest(http.MethodPost, url, body)
				req.Header.Set("Content-Type", contentType)
				w := serve(router, req)

				Convey("Then it should be a bad request", func() {
					So(w.Code, ShouldEqual, http.StatusBadRequest)
				})
			})

			Convey("When no archive is attached", func() {
				body, contentType := multipartUpload(metadata, nil)
				req := httptest.NewRequest(http.MethodPost, url, body)
				req.Header.Set("Content-Type", contentType)
				w := serve(router, req)

				Convey("Then it should be a bad request", func() {
					So(w.Code, ShouldEqual, http.StatusBadRequest)
				})
			})

			Convey("When the request is not multipart", func() {
				req := httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString("{}"))
				req.Header.Set("Content-Type", "application/json")
				w := serve(router, req)

				Convey("Then it should be a bad request", func() {
					So(w.Code, ShouldEqual, http.StatusBadRequest)
				})
			})
		})
	})
}

func TestMirrorController_Delete(t *testing.T) {
	Convey("Subject: Deleting mirrored providers", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}

		Convey("Given no authenticated user", func() {
			router, _ := setupMirrorRouter(t, nil, "", true)

			Convey("When a deletion is requested", func() {
				w := serve(router, httptest.NewRequest(http.MethodDelete, "/v1/api/mirror/registry.terraform.io", nil))

				Convey("Then it should be unauthorized", func() {
					So(w.Code, ShouldEqual, http.StatusUnauthorized)
				})
			})
		})

		Convey("Given a user without delete permission", func() {
			router, _ := setupMirrorRouter(t, user, "", false)

			Convey("When a deletion is requested", func() {
				w := serve(router, httptest.NewRequest(http.MethodDelete, "/v1/api/mirror/registry.terraform.io/hashicorp/null/3.2.4", nil))

				Convey("Then it should be forbidden", func() {
					So(w.Code, ShouldEqual, http.StatusForbidden)
				})
			})
		})

		Convey("Given a user allowed to delete under one hostname only", func() {
			router, mockService := setupMirrorRouter(t, user, "p, test-user, mirror, delete, registry.terraform.io*, allow", false)

			Convey("When a version deletion is requested", func() {
				mockService.On("DeleteVersion", "registry.terraform.io", "hashicorp", "null", "3.2.4").Return(nil)

				w := serve(router, httptest.NewRequest(http.MethodDelete, "/v1/api/mirror/registry.terraform.io/hashicorp/null/3.2.4", nil))

				Convey("Then it should succeed", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})

			Convey("When a provider deletion is requested", func() {
				mockService.On("Delete", "registry.terraform.io", "hashicorp", "null").Return(nil)

				w := serve(router, httptest.NewRequest(http.MethodDelete, "/v1/api/mirror/registry.terraform.io/hashicorp/null", nil))

				Convey("Then it should succeed", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})

			Convey("When a namespace deletion is requested", func() {
				mockService.On("DeleteNamespace", "registry.terraform.io", "hashicorp").Return(nil)

				w := serve(router, httptest.NewRequest(http.MethodDelete, "/v1/api/mirror/registry.terraform.io/hashicorp", nil))

				Convey("Then it should succeed", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})

			Convey("When a hostname deletion is requested", func() {
				mockService.On("DeleteHostname", "registry.terraform.io").Return(nil)

				w := serve(router, httptest.NewRequest(http.MethodDelete, "/v1/api/mirror/registry.terraform.io", nil))

				Convey("Then it should succeed", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})

			Convey("When the target does not exist", func() {
				mockService.On("Delete", "registry.terraform.io", "hashicorp", "missing").Return(errors.New("not found"))

				w := serve(router, httptest.NewRequest(http.MethodDelete, "/v1/api/mirror/registry.terraform.io/hashicorp/missing", nil))

				Convey("Then it should be not found", func() {
					So(w.Code, ShouldEqual, http.StatusNotFound)
				})
			})

			Convey("When a deletion under another hostname is requested", func() {
				w := serve(router, httptest.NewRequest(http.MethodDelete, "/v1/api/mirror/registry.opentofu.org/hashicorp/null", nil))

				Convey("Then it should be forbidden and the service not called", func() {
					So(w.Code, ShouldEqual, http.StatusForbidden)
					mockService.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
				})
			})
		})
	})
}
