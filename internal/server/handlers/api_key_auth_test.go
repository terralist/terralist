package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"terralist/internal/server/services"
	"terralist/pkg/auth"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthentication_ParseApiKey(t *testing.T) {
	Convey("Subject: Authenticating requests with an API key", t, func() {
		gin.SetMode(gin.TestMode)

		mockApiKeyService := services.NewMockApiKeyService(t)

		authentication := &Authentication{
			ApiKeyService: mockApiKeyService,
			MasterApiKey:  "master-key",
		}

		requestWithKey := func(key string) *gin.Context {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			ctx.Request.Header.Set("X-API-Key", key)

			return ctx
		}

		Convey("Given the master API key", func() {
			user, err := authentication.parseApiKey(requestWithKey("master-key"))

			Convey("Then the admin user is returned", func() {
				So(err, ShouldBeNil)
				So(user.Name, ShouldEqual, "terralist-admin")
			})
		})

		Convey("Given an API key", func() {
			mockApiKeyService.On("Authenticate", "ci-key").
				Return(&auth.User{Name: "apikey:ci", Email: "ci@example.com"}, nil)

			user, err := authentication.parseApiKey(requestWithKey("ci-key"))

			Convey("Then the key's user is returned", func() {
				So(err, ShouldBeNil)
				So(user.Email, ShouldEqual, "ci@example.com")
			})
		})

		Convey("Given an unknown API key", func() {
			mockApiKeyService.On("Authenticate", "unknown-key").
				Return(nil, errors.New("invalid key"))

			user, err := authentication.parseApiKey(requestWithKey("unknown-key"))

			Convey("Then the request is rejected", func() {
				So(user, ShouldBeNil)
				So(errors.Is(err, ErrInvalidValue), ShouldBeTrue)
			})
		})
	})
}
