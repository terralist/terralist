package services

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"terralist/internal/server/models/oauth"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"

	"github.com/mazen160/go-random"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestAuthorize(t *testing.T) {
	Convey("Subject: Compute an authorize URL", t, func() {
		mockProvider := auth.NewMockProvider(t)

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
		mockProvider := auth.NewMockProvider(t)

		loginService := &DefaultLoginService{
			Provider: mockProvider,
		}

		Convey("Given a code during a request", func() {
			code, _ := random.String(16)

			Convey("If the code validates the user access", func() {
				mockProvider.
					On("GetUserDetails", code, mock.AnythingOfType("*auth.User")).
					Run(func(args mock.Arguments) {
						user, ok := args.Get(1).(*auth.User)
						if !ok {
							t.Fatalf("expected *auth.User, got %T", args.Get(1))
						}
						user.Name = "alice"
						user.Email = "alice@example.com"
						user.Groups = []string{"engineering", "platform"}
						user.Authority = "example-org"
						user.AuthorityID = "authority-id-1"
					}).
					Return(nil)

				Convey("When the service is queried", func() {
					cc, err := loginService.UnpackCode(code, &oauth.Request{})

					Convey("The code components should be returned", func() {
						So(cc, ShouldNotBeNil)
						So(err, ShouldBeNil)
						So(cc.UserName, ShouldEqual, "alice")
						So(cc.UserEmail, ShouldEqual, "alice@example.com")
						So(cc.UserGroups, ShouldResemble, []string{"engineering", "platform"})
						So(cc.UserAuthority, ShouldEqual, "example-org")
						So(cc.UserAuthorityID, ShouldEqual, "authority-id-1")
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
					So(query.Get("code"), ShouldEqual, payload.String())
				})
			})
		})
	})
}

func TestParseTokenExpiration(t *testing.T) {
	Convey("Subject: Parse token expiration durations", t, func() {
		Convey("When parsing '1d'", func() {
			result := ParseTokenExpiration("1d")
			Convey("Should return 1 day in seconds", func() {
				So(result, ShouldEqual, 24*60*60)
			})
		})

		Convey("When parsing '1w'", func() {
			result := ParseTokenExpiration("1w")
			Convey("Should return 1 week in seconds", func() {
				So(result, ShouldEqual, 7*24*60*60)
			})
		})

		Convey("When parsing '1m'", func() {
			result := ParseTokenExpiration("1m")
			Convey("Should return 1 month (30 days) in seconds", func() {
				So(result, ShouldEqual, 30*24*60*60)
			})
		})

		Convey("When parsing '1y'", func() {
			result := ParseTokenExpiration("1y")
			Convey("Should return 1 year in seconds", func() {
				So(result, ShouldEqual, 365*24*60*60)
			})
		})

		Convey("When parsing 'never'", func() {
			result := ParseTokenExpiration("never")
			Convey("Should return 0 (no expiration)", func() {
				So(result, ShouldEqual, 0)
			})
		})

		Convey("When parsing an unknown value", func() {
			result := ParseTokenExpiration("unknown")
			Convey("Should return 1 day as default", func() {
				So(result, ShouldEqual, 24*60*60)
			})
		})
	})
}

func TestValidateToken(t *testing.T) {
	Convey("Subject: Validate a token", t, func() {
		mockJWT := jwt.NewMockJWT(t)

		loginService := &DefaultLoginService{
			JWT:                 mockJWT,
			TokenExpirationSecs: 24 * 60 * 60, // 1 day default
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
						On(
							"Build",
							mock.MatchedBy(func(u auth.User) bool {
								return u.Name == cc.UserName &&
									u.Email == cc.UserEmail &&
									u.Authority == cc.UserAuthority &&
									u.AuthorityID == cc.UserAuthorityID &&
									len(u.Groups) == len(cc.UserGroups)
							}),
							mock.AnythingOfType("int"),
						).
						Return(token, nil)

					cc.UserName = "alice"
					cc.UserEmail = "alice@example.com"
					cc.UserGroups = []string{"engineering", "platform"}
					cc.UserAuthority = "example-org"
					cc.UserAuthorityID = "authority-id-1"

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

func TestRedirect_UsesOpaqueCodeStore(t *testing.T) {
	salt, _ := random.String(16)
	state, _ := random.String(16)

	loginService := &DefaultLoginService{
		EncryptSalt: salt,
		CodeStore:   NewInMemoryOAuthCodeStore(2 * time.Minute),
	}

	cc := oauth.CodeComponents{
		UserName:  "alice",
		UserEmail: "alice@example.com",
	}
	req := oauth.Request{
		RedirectURI: "http://localhost/callback",
		State:       state,
	}

	redirectURL, oauthErr := loginService.Redirect(&cc, &req)
	if oauthErr != nil {
		t.Fatalf("expected redirect without error, got: %v", oauthErr)
	}

	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		t.Fatalf("failed to parse redirect url: %v", err)
	}

	code := parsedURL.Query().Get("code")
	if code == "" {
		t.Fatalf("expected opaque code in redirect url")
	}
	if strings.Contains(code, "{") {
		t.Fatalf("expected opaque code, got JSON-like payload")
	}

	resolved, resolveErr := loginService.ResolveCode(code)
	if resolveErr != nil {
		t.Fatalf("expected opaque code to resolve, got: %v", resolveErr)
	}

	if resolved.UserEmail != "alice@example.com" {
		t.Fatalf("unexpected resolved user email: %s", resolved.UserEmail)
	}
}

func TestResolveCode_FallsBackToPayloadDecode(t *testing.T) {
	salt, _ := random.String(16)
	loginService := &DefaultLoginService{
		EncryptSalt: salt,
		CodeStore:   NewInMemoryOAuthCodeStore(2 * time.Minute),
	}

	expected := oauth.CodeComponents{
		UserName:            "alice",
		UserEmail:           "alice@example.com",
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
	}

	payload, err := expected.ToPayload(salt)
	if err != nil {
		t.Fatalf("failed to create legacy payload: %v", err)
	}

	resolved, oauthErr := loginService.ResolveCode(payload.String())
	if oauthErr != nil {
		t.Fatalf("expected payload decode to still work, got: %v", oauthErr)
	}

	if resolved.UserName != expected.UserName || resolved.UserEmail != expected.UserEmail {
		t.Fatalf("resolved payload mismatch: %+v", resolved)
	}
}

func TestTerraformFlow_HighGroupCountClaimsArePreserved(t *testing.T) {
	mockProvider := auth.NewMockProvider(t)
	jwtManager, err := jwt.New("high-group-test-signing-secret")
	if err != nil {
		t.Fatalf("failed to create jwt manager: %v", err)
	}

	salt, _ := random.String(16)
	exchangeKey, _ := random.String(16)

	groupCount := 750
	groups := make([]string, 0, groupCount)
	for i := 0; i < groupCount; i++ {
		groups = append(groups, fmt.Sprintf("TEAM_%04d", i))
	}

	providerCode := "provider-code"
	verifier := "terraform-pkce-verifier"
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	mockProvider.
		On("GetUserDetails", providerCode, mock.AnythingOfType("*auth.User")).
		Run(func(args mock.Arguments) {
			user, ok := args.Get(1).(*auth.User)
			if !ok {
				t.Fatalf("expected *auth.User, got %T", args.Get(1))
			}
			user.Name = "alice"
			user.Email = "alice@example.com"
			user.Groups = groups
			user.Authority = "example-org"
			user.AuthorityID = "authority-id-1"
		}).
		Return(nil)

	loginService := &DefaultLoginService{
		Provider:            mockProvider,
		JWT:                 jwtManager,
		CodeStore:           NewInMemoryOAuthCodeStore(2 * time.Minute),
		EncryptSalt:         salt,
		CodeExchangeKey:     exchangeKey,
		TokenExpirationSecs: 24 * 60 * 60,
	}

	request := oauth.Request{
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		RedirectURI:         "http://localhost/callback",
		State:               "state-1",
	}

	codeComponents, oauthErr := loginService.UnpackCode(providerCode, &request)
	if oauthErr != nil {
		t.Fatalf("expected unpack code to succeed: %v", oauthErr)
	}
	if len(codeComponents.UserGroups) != groupCount {
		t.Fatalf("expected %d unpacked groups, got %d", groupCount, len(codeComponents.UserGroups))
	}

	redirectURL, oauthErr := loginService.Redirect(codeComponents, &request)
	if oauthErr != nil {
		t.Fatalf("expected redirect to succeed: %v", oauthErr)
	}

	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		t.Fatalf("failed to parse redirect URL: %v", err)
	}
	opaqueCode := parsedURL.Query().Get("code")
	if opaqueCode == "" {
		t.Fatalf("expected opaque code in redirect query")
	}
	if len(opaqueCode) > 100 {
		t.Fatalf("expected short opaque code, got length %d", len(opaqueCode))
	}

	resolvedComponents, oauthErr := loginService.ResolveCode(opaqueCode)
	if oauthErr != nil {
		t.Fatalf("expected opaque code to resolve: %v", oauthErr)
	}
	if len(resolvedComponents.UserGroups) != groupCount {
		t.Fatalf("expected %d resolved groups, got %d", groupCount, len(resolvedComponents.UserGroups))
	}

	tokenResponse, oauthErr := loginService.ValidateToken(resolvedComponents, verifier)
	if oauthErr != nil {
		t.Fatalf("expected token validation to succeed: %v", oauthErr)
	}

	serializer, err := jwtManager.Extract(tokenResponse.AccessToken)
	if err != nil {
		t.Fatalf("failed to extract token payload: %v", err)
	}

	user, ok := serializer.(*auth.User)
	if !ok {
		t.Fatalf("expected auth.User payload, got %T", serializer)
	}

	if user.Email != "alice@example.com" {
		t.Fatalf("unexpected email in jwt payload: %s", user.Email)
	}
	if len(user.Groups) != groupCount {
		t.Fatalf("expected %d groups in jwt payload, got %d", groupCount, len(user.Groups))
	}
}
