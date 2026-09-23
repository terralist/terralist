package controllers

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"terralist/internal/server/models/oauth"
	"terralist/internal/server/services"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

var (
	loginTestStateKey = []byte("test-state-key")
	errAccessDenied   = errors.New("access denied")
)

func setupLoginRouter(t *testing.T) (*gin.Engine, *services.MockLoginService) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	mockService := services.NewMockLoginService(t)

	hostURL, err := url.Parse("https://registry.example.com")
	if err != nil {
		t.Fatalf("failed to parse host URL: %v", err)
	}

	controller := &DefaultLoginController{
		LoginService: mockService,
		HostURL:      hostURL,
		StateKey:     loginTestStateKey,
	}

	router := gin.New()
	group := router.Group("/v1")
	paths := controller.Paths()
	groups := make([]*gin.RouterGroup, len(paths))
	for i, p := range paths {
		groups[i] = group.Group(p)
	}

	controller.Subscribe(groups...)

	return router, mockService
}

func signedState(t *testing.T, r oauth.Request) string {
	t.Helper()

	payload, err := r.ToPayload(loginTestStateKey)
	if err != nil {
		t.Fatalf("failed to sign state: %v", err)
	}

	return payload.String()
}

func forgedState(t *testing.T, r oauth.Request) string {
	t.Helper()

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	fakeSignature := bytes.Repeat([]byte{0}, sha256.Size)

	return base64.StdEncoding.EncodeToString(append(fakeSignature, data...))
}

func TestLoginController_Authorize(t *testing.T) {
	Convey("Subject: Starting the authorization flow", t, func() {
		router, mockService := setupLoginRouter(t)

		authorize := func(redirectURI string) *httptest.ResponseRecorder {
			query := url.Values{}
			query.Set("client_id", "terraform-cli")
			query.Set("code_challenge", "challenge")
			query.Set("code_challenge_method", "S256")
			query.Set("redirect_uri", redirectURI)
			query.Set("response_type", "code")
			query.Set("state", "client-state")

			req := httptest.NewRequest(http.MethodGet, "/v1/auth/authorization?"+query.Encode(), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			return w
		}

		Convey("Given a redirect URI on a Terraform loopback port", func() {
			var receivedState oauth.Payload
			mockService.On("Authorize", mock.Anything).
				Run(func(args mock.Arguments) {
					receivedState = args.Get(0).(oauth.Payload) //nolint:forcetypeassert
				}).
				Return("https://idp.example.com/authorize", nil)

			w := authorize("http://localhost:10000/login")

			Convey("Then the client is redirected to the identity provider", func() {
				So(w.Code, ShouldEqual, http.StatusFound)
				So(w.Header().Get("Location"), ShouldEqual, "https://idp.example.com/authorize")
			})

			Convey("Then the state handed to the provider verifies and carries the request", func() {
				r, err := receivedState.ToRequest(loginTestStateKey)
				So(err, ShouldBeNil)
				So(r.RedirectURI, ShouldEqual, "http://localhost:10000/login")
				So(r.CodeChallenge, ShouldEqual, "challenge")
				So(r.State, ShouldEqual, "client-state")
			})
		})

		Convey("Given a redirect URI on a foreign host", func() {
			w := authorize("https://evil.example.com/steal")

			Convey("Then the request is rejected without redirecting", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Header().Get("Location"), ShouldBeEmpty)
				mockService.AssertNotCalled(t, "Authorize", mock.Anything)
			})
		})

		Convey("Given a loopback redirect URI outside the Terraform port range", func() {
			w := authorize("http://127.0.0.1:65000/login")

			Convey("Then the request is rejected without redirecting", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Header().Get("Location"), ShouldBeEmpty)
			})
		})
	})
}

func TestLoginController_Redirect(t *testing.T) {
	Convey("Subject: Completing the authorization flow on the provider callback", t, func() {
		router, mockService := setupLoginRouter(t)

		callback := func(state string) *httptest.ResponseRecorder {
			query := url.Values{}
			query.Set("code", "provider-code")
			query.Set("state", state)

			req := httptest.NewRequest(http.MethodGet, "/v1/api/auth/redirect?"+query.Encode(), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			return w
		}

		request := oauth.Request{
			ClientID:            "terraform-cli",
			CodeChallenge:       "challenge",
			CodeChallengeMethod: "S256",
			RedirectURI:         "http://localhost:10000/login",
			ResponseType:        "code",
			State:               "client-state",
		}

		Convey("Given a state signed by the server", func() {
			components := &oauth.CodeComponents{
				CodeChallenge:       request.CodeChallenge,
				CodeChallengeMethod: request.CodeChallengeMethod,
				UserName:            "Test User",
				UserEmail:           "test@example.com",
			}

			mockService.On("UnpackCode", "provider-code", mock.Anything).Return(components, nil)
			mockService.On("Redirect", components, mock.Anything).
				Return("http://localhost:10000/login?state=client-state&code=opaque", nil)

			w := callback(signedState(t, request))

			Convey("Then the client is redirected to its callback with the code", func() {
				So(w.Code, ShouldEqual, http.StatusFound)
				So(w.Header().Get("Location"), ShouldEqual, "http://localhost:10000/login?state=client-state&code=opaque")
			})
		})

		Convey("Given a signed state whose code cannot be unpacked", func() {
			mockService.On("UnpackCode", "provider-code", mock.Anything).
				Return(nil, oauth.WrapError(errAccessDenied, oauth.AccessDenied))

			w := callback(signedState(t, request))

			Convey("Then the error is delivered to the allowed callback", func() {
				So(w.Code, ShouldEqual, http.StatusFound)
				So(w.Header().Get("Location"), ShouldStartWith, "http://localhost:10000/login?error=access_denied")
			})
		})

		Convey("Given a state forged without the server key", func() {
			forged := request
			forged.RedirectURI = "https://evil.example.com/steal"

			w := callback(forgedState(t, forged))

			Convey("Then the request is rejected without redirecting", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Header().Get("Location"), ShouldBeEmpty)
				mockService.AssertNotCalled(t, "UnpackCode", mock.Anything, mock.Anything)
			})
		})

		Convey("Given a state shorter than a signature", func() {
			w := callback("QUFBQQ==")

			Convey("Then the request is rejected without redirecting", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Header().Get("Location"), ShouldBeEmpty)
			})
		})

		Convey("Given a signed state pointing to a foreign host", func() {
			foreign := request
			foreign.RedirectURI = "https://evil.example.com/steal"

			w := callback(signedState(t, foreign))

			Convey("Then the request is rejected without redirecting", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(w.Header().Get("Location"), ShouldBeEmpty)
				mockService.AssertNotCalled(t, "UnpackCode", mock.Anything, mock.Anything)
			})
		})

		Convey("Given no state at all", func() {
			w := callback("")

			Convey("Then the request is rejected", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}
