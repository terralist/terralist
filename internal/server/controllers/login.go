package controllers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"terralist/internal/server/models/oauth"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"

	"github.com/gin-gonic/gin"
)

type LoginController struct {
	Provider auth.Provider
	JWT      jwt.JWT

	EncryptSalt     string
	CodeExchangeKey string
}

func (l *LoginController) Authorize() func(c *gin.Context) {
	return func(c *gin.Context) {
		request := oauth.Request{
			ClientID:            c.Query("client_id"),
			CodeChallenge:       c.Query("code_challenge"),
			CodeChallengeMethod: c.Query("code_challenge_method"),
			RedirectUri:         c.Query("redirect_uri"),
			ResponseType:        c.Query("response_type"),
			State:               c.Query("state"),
		}

		state, err := request.ToPayload(l.EncryptSalt)

		if err != nil {
			c.Redirect(
				http.StatusFound,
				l.getUrlForRedirectWithError(
					request.RedirectUri,
					request.State,
					oauth.ServerError,
					err.Error(),
				),
			)
			return
		}

		authorizeURL := l.Provider.GetAuthorizeUrl(state.String())

		c.Redirect(
			http.StatusFound,
			authorizeURL,
		)
	}
}

func (l *LoginController) Redirect() func(c *gin.Context) {
	return func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")

		request, err := oauth.Payload(state).ToRequest(l.EncryptSalt)

		if err != nil {
			c.Redirect(
				http.StatusFound,
				l.getUrlForRedirectWithError(
					request.RedirectUri,
					request.State,
					oauth.InvalidRequest,
					err.Error(),
				),
			)
			return
		}

		var userDetails auth.User
		if err := l.Provider.GetUserDetails(code, &userDetails); err != nil {
			c.Redirect(
				http.StatusFound,
				l.getUrlForRedirectWithError(
					request.RedirectUri,
					request.State,
					oauth.AccessDenied,
					err.Error(),
				),
			)
			return
		}

		codeComponents := oauth.CodeComponents{
			Key:                 l.CodeExchangeKey,
			CodeChallenge:       request.CodeChallenge,
			CodeChallengeMethod: request.CodeChallengeMethod,
			UserName:            userDetails.Name,
			UserEmail:           userDetails.Email,
		}

		payload, err := codeComponents.ToPayload(l.EncryptSalt)
		if err != nil {
			c.Redirect(
				http.StatusFound,
				l.getUrlForRedirectWithError(
					request.RedirectUri,
					request.State,
					oauth.InvalidRequest,
					err.Error(),
				),
			)
			return
		}

		c.Redirect(
			http.StatusFound,
			fmt.Sprintf(
				"%s?state=%s&code=%s",
				request.RedirectUri,
				request.State,
				payload,
			),
		)
	}
}

func (l *LoginController) TokenValidate() func(c *gin.Context) {
	return func(c *gin.Context) {
		var request oauth.TokenValidationRequest
		if err := c.Bind(&request); err != nil {
			c.Redirect(
				http.StatusFound,
				l.getUrlForRedirectWithError(
					request.RedirectUri,
					"",
					oauth.InvalidRequest,
					err.Error(),
				),
			)
			return
		}

		if request.GrantType != "authorization_code" {
			c.Redirect(
				http.StatusFound,
				l.getUrlForRedirectWithError(
					request.RedirectUri,
					"",
					oauth.AccessDenied,
					"requested grant type is not supported",
				),
			)
		}

		codeComponents, err := oauth.Payload(request.Code).ToCodeComponents(l.EncryptSalt)

		if err != nil {
			c.Redirect(
				http.StatusFound,
				l.getUrlForRedirectWithError(
					request.RedirectUri,
					"",
					oauth.InvalidRequest,
					err.Error(),
				),
			)
			return
		}

		if codeComponents.CodeChallengeMethod != "S256" {
			c.Redirect(
				http.StatusFound,
				l.getUrlForRedirectWithError(
					request.RedirectUri,
					"",
					oauth.UnsupportedResponseType,
					"code challenge method unsupported",
				),
			)
			return
		}

		codeVerifierHash := sha256.Sum256([]byte(request.CodeVerifier))
		codeVerifierDecoded := base64.RawURLEncoding.EncodeToString(codeVerifierHash[:])

		if string(codeVerifierDecoded) != codeComponents.CodeChallenge {
			c.Redirect(
				http.StatusFound,
				l.getUrlForRedirectWithError(
					request.RedirectUri,
					"",
					oauth.InvalidRequest,
					"code verification failed",
				),
			)
			return
		}

		t, err := l.JWT.Build(auth.User{
			Name:  codeComponents.UserName,
			Email: codeComponents.UserEmail,
		})

		if err != nil {
			c.Redirect(
				http.StatusFound,
				l.getUrlForRedirectWithError(
					request.RedirectUri,
					"",
					oauth.InvalidRequest,
					err.Error(),
				),
			)

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  t,
			"token_type":    "bearer",
			"refresh_token": "",
			"expires_in":    0,
		})
	}
}

func (l *LoginController) getUrlForRedirectWithError(redirectUri string, state string, oauthError oauth.Error, description string) string {
	stateQuery := ""
	if state != "" {
		stateQuery = fmt.Sprintf("&state=%s", state)
	}

	return fmt.Sprintf(
		"%s?error=%s&error_description=%s%s",
		redirectUri,
		oauthError,
		url.QueryEscape(description),
		stateQuery,
	)
}
