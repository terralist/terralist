package services

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/url"
	"testing"

	"terralist/internal/server/models/oauth"

	mockAuth "terralist/mocks/pkg/auth"
	mockJWT "terralist/mocks/pkg/auth/jwt"

	"github.com/mazen160/go-random"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestAuthorize(t *testing.T) {
	Convey("Subject: Compute an authorize URL", t, func() {
		mockProvider := mockAuth.NewProvider(t)

		loginService := &DefaultLoginService{
			Provider: mockProvider,
		}

		Convey("Given a state", func() {
			state, _ := random.String(16)

			mockProvider.
				On("GetAuthorizeUrl", state).
				Return("")

			Convey("When the service is queried", func() {
				url, err := loginService.Authorize(oauth.Payload(state))

				Convey("Should return an URL", func() {
					So(url, ShouldNotBeNil)
					So(err, ShouldBeNil)
				})
			})
		})
	})
}

func TestUnpackCode(t *testing.T) {
	Convey("Subject: Unpack received code", t, func() {
		mockProvider := mockAuth.NewProvider(t)

		loginService := &DefaultLoginService{
			Provider: mockProvider,
		}

		Convey("Given a code during a request", func() {
			code, _ := random.String(16)

			Convey("If the code validates the user access", func() {
				mockProvider.
					On("GetUserDetails", code, mock.AnythingOfType("*auth.User")).
					Return(nil)

				Convey("When the service is queried", func() {
					cc, err := loginService.UnpackCode(code, &oauth.Request{})

					Convey("The code components should be returned", func() {
						So(cc, ShouldNotBeNil)
						So(err, ShouldBeNil)
					})
				})
			})

			Convey("If the code does not validate the user access", func() {
				mockProvider.
					On("GetUserDetails", code, mock.AnythingOfType("*auth.User")).
					Return(errors.New(""))

				Convey("When the service is queried", func() {
					cc, err := loginService.UnpackCode(code, &oauth.Request{})

					Convey("An access denied error should be returned", func() {
						So(cc, ShouldBeNil)
						So(err.Kind(), ShouldEqual, oauth.AccessDenied)
					})
				})
			})
		})
	})
}

func TestRedirect(t *testing.T) {
	Convey("Subject: Compute a redirect URL", t, func() {
		salt, _ := random.String(16)

		loginService := &DefaultLoginService{
			EncryptSalt: salt,
		}

		Convey("Given the code components and the request", func() {
			cc := oauth.CodeComponents{}
			payload, err := cc.ToPayload(salt)

			So(err, ShouldBeNil)

			state, _ := random.String(16)

			req := oauth.Request{
				RedirectURI: "http://localhost",
				State:       state,
			}

			Convey("When the service is queried", func() {
				redirectURL, oauthError := loginService.Redirect(&cc, &req)

				Convey("A redirect URL should be returned", func() {
					So(redirectURL, ShouldNotBeNil)
					So(oauthError, ShouldBeNil)

					parsedURL, err := url.Parse(redirectURL)
					So(err, ShouldBeNil)
					So(parsedURL.Host, ShouldEqual, "localhost")

					query := parsedURL.Query()
					So(query.Get("state"), ShouldEqual, state)
					So(query.Get("code"), ShouldEqual, payload)
				})
			})
		})
	})
}

func TestValidateToken(t *testing.T) {
	Convey("Subject: Validate a token", t, func() {
		mockJWT := mockJWT.NewJWT(t)

		loginService := &DefaultLoginService{
			JWT: mockJWT,
		}

		Convey("Given the code components and the code verifier", func() {
			cc := oauth.CodeComponents{}
			verifier, _ := random.String(16)

			Convey("If the challenge method is known (S256)", func() {
				cc.CodeChallengeMethod = "S256"

				Convey("If the code verifier is correct", func() {
					hash := sha256.Sum256([]byte(verifier))
					cv := base64.RawURLEncoding.EncodeToString(hash[:])

					cc.CodeChallenge = cv

					token, _ := random.String(16)
					mockJWT.
						On("Build", mock.AnythingOfType("auth.User"), mock.AnythingOfType("int")).
						Return(token, nil)

					Convey("When the service is queried", func() {
						tr, err := loginService.ValidateToken(&cc, verifier)

						Convey("A token should be generated", func() {
							So(err, ShouldBeNil)
							So(tr.AccessToken, ShouldEqual, token)
							So(tr.TokenType, ShouldEqual, "bearer")
							So(tr.RefreshToken, ShouldEqual, "")
						})
					})
				})

				Convey("If the code verifier is not correct", func() {
					cc.CodeChallenge = "100%-not-the-same-with-the-verifier"

					Convey("When the service is queried", func() {
						tr, err := loginService.ValidateToken(&cc, verifier)

						Convey("An invalid request error should be returned", func() {
							So(tr, ShouldBeNil)
							So(err, ShouldNotBeNil)
							So(err.Kind(), ShouldEqual, oauth.InvalidRequest)
						})
					})
				})
			})

			Convey("If the challenge method is unknown", func() {
				cc.CodeChallengeMethod = "100%-unknown-method"

				Convey("When the service is queried", func() {
					tr, err := loginService.ValidateToken(&cc, verifier)

					Convey("An unsupported response error should be returned", func() {
						So(tr, ShouldBeNil)
						So(err, ShouldNotBeNil)
						So(err.Kind(), ShouldEqual, oauth.UnsupportedResponseType)
					})
				})
			})
		})
	})
}
