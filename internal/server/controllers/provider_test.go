package controllers

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
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
	"terralist/pkg/database/entity"
	"terralist/pkg/rbac"
	"terralist/pkg/session/cookie"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

// setupProviderRouter creates a gin test router with the provider controller
// registered under /v1. When user is non-nil, a middleware injects the user
// into the context to simulate an authenticated request.
func setupProviderRouter(t *testing.T, user *auth.User, policyCSV string) (*gin.Engine, *services.MockProviderService, *services.MockAuthorityService) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	mockService := services.NewMockProviderService(t)
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

	controller := &DefaultProviderController{
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

	return router, mockService, mockAuthorityService
}

// packagesUpload builds a multipart request body for the package upload
// endpoint from the given files and form values.
func packagesUpload(files map[string][]namedContent, values map[string]string) (*bytes.Buffer, string) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for field, parts := range files {
		for _, part := range parts {
			w, _ := writer.CreateFormFile(field, part.name)
			_, _ = w.Write(part.content)
		}
	}

	for field, value := range values {
		_ = writer.WriteField(field, value)
	}

	_ = writer.Close()

	return &body, writer.FormDataContentType()
}

type namedContent struct {
	name    string
	content []byte
}

func TestProviderController_UploadPackages(t *testing.T) {
	Convey("Subject: Uploading the packages of a provider version", t, func() {
		user := &auth.User{Name: "test-user", Email: "test@example.com"}
		url := "/v1/api/providers/hashicorp/null/3.2.4/upload-files"
		authorityID, _ := uuid.NewRandom()
		metadata := []byte(`{"archives":{"linux_amd64":{"url":"terraform-provider-null_3.2.4_linux_amd64.zip","hashes":["h1:abc"]}}}`)
		archives := map[string][]namedContent{
			"metadata": {{name: "3.2.4.json", content: metadata}},
			"archives": {{name: "terraform-provider-null_3.2.4_linux_amd64.zip", content: []byte("zip-content")}},
		}

		post := func(router *gin.Engine, body *bytes.Buffer, contentType string) *httptest.ResponseRecorder {
			req := httptest.NewRequest(http.MethodPost, url, body)
			req.Header.Set("Content-Type", contentType)

			return serve(router, req)
		}

		Convey("Given no authenticated user", func() {
			router, _, _ := setupProviderRouter(t, nil, "")

			Convey("When an upload is requested", func() {
				body, contentType := packagesUpload(archives, nil)
				w := post(router, body, contentType)

				Convey("Then it should be unauthorized", func() {
					So(w.Code, ShouldEqual, http.StatusUnauthorized)
				})
			})
		})

		Convey("Given a user without create permission", func() {
			router, _, _ := setupProviderRouter(t, user, "")

			Convey("When an upload is requested", func() {
				body, contentType := packagesUpload(archives, nil)
				w := post(router, body, contentType)

				Convey("Then it should be forbidden", func() {
					So(w.Code, ShouldEqual, http.StatusForbidden)
				})
			})
		})

		Convey("Given a user with create permission", func() {
			router, mockService, mockAuthorityService := setupProviderRouter(t, user, "p, test-user, providers, create, hashicorp/*, allow")
			mockAuthorityService.On("GetByName", "hashicorp").Return(&authority.Authority{Entity: entity.Entity{ID: authorityID}, Name: "hashicorp"}, nil).Maybe()

			Convey("When a valid upload without signature material is requested", func() {
				var uploaded *provider.PackagesUploadDTO
				mockService.
					On("UploadPackages", mock.AnythingOfType("*provider.PackagesUploadDTO")).
					Run(func(args mock.Arguments) {
						uploaded, _ = args.Get(0).(*provider.PackagesUploadDTO)
					}).
					Return(nil)

				body, contentType := packagesUpload(archives, nil)
				w := post(router, body, contentType)

				Convey("Then it should succeed and pass the parsed upload to the service", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					So(uploaded.AuthorityID, ShouldEqual, authorityID)
					So(uploaded.Name, ShouldEqual, "null")
					So(uploaded.Version, ShouldEqual, "3.2.4")
					So(uploaded.Protocols, ShouldBeEmpty)
					So(uploaded.Metadata.Archives["linux_amd64"].Hashes, ShouldResemble, []string{"h1:abc"})
					So(uploaded.ShaSums, ShouldBeNil)
					So(uploaded.ShaSumsSignature, ShouldBeNil)
					So(len(uploaded.Archives), ShouldEqual, 1)

					content, _ := io.ReadAll(uploaded.Archives[0])
					So(uploaded.Archives[0].Name(), ShouldEqual, "terraform-provider-null_3.2.4_linux_amd64.zip")
					So(string(content), ShouldEqual, "zip-content")
				})
			})

			Convey("When a valid upload with signature material is requested", func() {
				var uploaded *provider.PackagesUploadDTO
				mockService.
					On("UploadPackages", mock.AnythingOfType("*provider.PackagesUploadDTO")).
					Run(func(args mock.Arguments) {
						uploaded, _ = args.Get(0).(*provider.PackagesUploadDTO)
					}).
					Return(nil)

				signed := map[string][]namedContent{
					"metadata":          archives["metadata"],
					"archives":          archives["archives"],
					"shasums":           {{name: "SHA256SUMS", content: []byte("sums")}},
					"shasums_signature": {{name: "SHA256SUMS.sig", content: []byte("sig")}},
				}
				body, contentType := packagesUpload(signed, map[string]string{"protocols": "5.0, 6.0"})
				w := post(router, body, contentType)

				Convey("Then it should pass the signature files and the protocols to the service", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
					So(uploaded.Protocols, ShouldResemble, []string{"5.0", "6.0"})
					So(uploaded.ShaSums.Name(), ShouldEqual, "SHA256SUMS")
					So(uploaded.ShaSumsSignature.Name(), ShouldEqual, "SHA256SUMS.sig")
				})
			})

			Convey("When the service rejects the upload", func() {
				mockService.
					On("UploadPackages", mock.Anything).
					Return(errors.New("version 3.2.4 already exists"))

				body, contentType := packagesUpload(archives, nil)
				w := post(router, body, contentType)

				Convey("Then it should return a conflict with the error", func() {
					So(w.Code, ShouldEqual, http.StatusConflict)
					So(w.Body.String(), ShouldContainSubstring, "already exists")
				})
			})

			Convey("When the metadata part is missing", func() {
				body, contentType := packagesUpload(map[string][]namedContent{"archives": archives["archives"]}, nil)
				w := post(router, body, contentType)

				Convey("Then it should be a bad request", func() {
					So(w.Code, ShouldEqual, http.StatusBadRequest)
				})
			})

			Convey("When the metadata is not valid JSON", func() {
				body, contentType := packagesUpload(map[string][]namedContent{
					"metadata": {{name: "3.2.4.json", content: []byte("not-json")}},
					"archives": archives["archives"],
				}, nil)
				w := post(router, body, contentType)

				Convey("Then it should be a bad request", func() {
					So(w.Code, ShouldEqual, http.StatusBadRequest)
				})
			})

			Convey("When no package is attached", func() {
				body, contentType := packagesUpload(map[string][]namedContent{"metadata": archives["metadata"]}, nil)
				w := post(router, body, contentType)

				Convey("Then it should be a bad request", func() {
					So(w.Code, ShouldEqual, http.StatusBadRequest)
				})
			})

			Convey("When the request is not multipart", func() {
				w := post(router, bytes.NewBufferString("{}"), "application/json")

				Convey("Then it should be a bad request", func() {
					So(w.Code, ShouldEqual, http.StatusBadRequest)
				})
			})

			Convey("When the authority does not exist", func() {
				router, _, mockAuthorityService := setupProviderRouter(t, user, "p, test-user, providers, create, unknown/*, allow")
				mockAuthorityService.On("GetByName", "unknown").Return(nil, errors.New("no authority found"))

				body, contentType := packagesUpload(archives, nil)
				req := httptest.NewRequest(http.MethodPost, "/v1/api/providers/unknown/null/3.2.4/upload-files", body)
				req.Header.Set("Content-Type", contentType)
				w := serve(router, req)

				Convey("Then it should be not found", func() {
					So(w.Code, ShouldEqual, http.StatusNotFound)
				})
			})
		})
	})
}
