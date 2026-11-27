package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"terralist/internal/server/models/oauth"

	. "github.com/smartystreets/goconvey/convey"
)

func postForm(router http.Handler, path string, form url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

func TestLoginController_RefreshToken(t *testing.T) {
	Convey("Subject: Refreshing an access token", t, func() {
		router, mockService := setupLoginRouterWithRefresh(t, true)

		form := url.Values{
			"grant_type":    []string{"refresh_token"},
			"refresh_token": []string{"refresh-token"},
			"client_id":     []string{"oauth-client"},
		}

		Convey("When the service refreshes the session", func() {
			mockService.
				On("RefreshToken", "refresh-token", "oauth-client").
				Return(&oauth.TokenResponse{
					AccessToken:  "access-token",
					TokenType:    "bearer",
					RefreshToken: "next-refresh-token",
					ExpiresIn:    3600,
				}, nil)

			w := postForm(router, "/v1/auth/token", form)

			Convey("Then the new tokens should be returned", func() {
				So(w.Code, ShouldEqual, http.StatusOK)

				var resp oauth.TokenResponse
				So(json.Unmarshal(w.Body.Bytes(), &resp), ShouldBeNil)
				So(resp.AccessToken, ShouldEqual, "access-token")
				So(resp.RefreshToken, ShouldEqual, "next-refresh-token")
			})
		})

		Convey("When the refresh token is refused", func() {
			mockService.
				On("RefreshToken", "refresh-token", "oauth-client").
				Return(nil, oauth.WrapError(errors.New("refresh token is invalid or expired"), oauth.InvalidGrant))

			w := postForm(router, "/v1/auth/token", form)

			Convey("Then an invalid_grant error should be returned as JSON", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Header().Get("Location"), ShouldBeEmpty)
				So(w.Header().Get("Cache-Control"), ShouldEqual, "no-store")

				var body map[string]string
				So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)
				So(body["error"], ShouldEqual, oauth.InvalidGrant)
			})
		})

		Convey("When the upstream provider is unavailable", func() {
			mockService.
				On("RefreshToken", "refresh-token", "oauth-client").
				Return(nil, oauth.WrapError(errors.New("dial tcp: connection refused"), oauth.TemporarilyUnavailable))

			w := postForm(router, "/v1/auth/token", form)

			Convey("Then a 503 should be returned without the upstream details", func() {
				So(w.Code, ShouldEqual, http.StatusServiceUnavailable)
				So(w.Body.String(), ShouldNotContainSubstring, "dial tcp")

				var body map[string]string
				So(json.Unmarshal(w.Body.Bytes(), &body), ShouldBeNil)
				So(body["error"], ShouldEqual, oauth.TemporarilyUnavailable)
			})
		})
	})
}

func TestLoginController_RevokeToken(t *testing.T) {
	Convey("Subject: Revoking a refresh token", t, func() {
		form := url.Values{
			"token":     []string{"refresh-token"},
			"client_id": []string{"oauth-client"},
		}

		Convey("Given refresh tokens are enabled", func() {
			router, mockService := setupLoginRouterWithRefresh(t, true)

			Convey("When a token is revoked", func() {
				mockService.On("RevokeToken", "refresh-token", "oauth-client").Return(nil)

				w := postForm(router, "/v1/auth/revoke", form)

				Convey("Then the request should succeed", func() {
					So(w.Code, ShouldEqual, http.StatusOK)
				})
			})

			Convey("When no token is given", func() {
				w := postForm(router, "/v1/auth/revoke", url.Values{})

				Convey("Then an invalid_request error should be returned", func() {
					So(w.Code, ShouldEqual, http.StatusBadRequest)
				})
			})
		})

		Convey("Given refresh tokens are disabled", func() {
			router, _ := setupLoginRouterWithRefresh(t, false)

			w := postForm(router, "/v1/auth/revoke", form)

			Convey("Then the endpoint should not exist", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})
	})
}
