package controllers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	models "github.com/valentindeaconu/terralist/internal/server/models/oauth"
	"github.com/valentindeaconu/terralist/internal/server/oauth"
	"github.com/valentindeaconu/terralist/internal/server/oauth/token"
	"github.com/valentindeaconu/terralist/settings"
)

type LoginController struct {
	Provider oauth.Engine
}

func (l *LoginController) Authorize() func(c *gin.Context) {
	return func(c *gin.Context) {
		request := models.AuthorizationRequest{
			ClientID:            c.Query("client_id"),
			CodeChallenge:       c.Query("code_challenge"),
			CodeChallengeMethod: c.Query("code_challenge_method"),
			RedirectUri:         c.Query("redirect_uri"),
			ResponseType:        c.Query("response_type"),
			State:               c.Query("state"),
		}

		state, err := request.ToPayload()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": []string{err.Error()},
			})
			return
		}

		url := l.Provider.GetAuthorizeUrl(state)

		c.Redirect(
			http.StatusFound,
			url,
		)
	}
}

func (l *LoginController) Redirect() func(c *gin.Context) {
	return func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")

		request, err := models.AuthorizationRequestFromPayload(state)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{err.Error()},
			})
			return
		}

		var userDetails models.UserDetails
		if err := l.Provider.GetUserDetails(code, &userDetails); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{err.Error()},
			})
		}

		codeComponents := models.AuthorizationCodeComponents{
			Key:                 settings.CodeExchangeKey,
			CodeChallenge:       request.CodeChallenge,
			CodeChallengeMethod: request.CodeChallengeMethod,
			UserName:            userDetails.Name,
			UserEmail:           userDetails.Email,
		}

		payload, err := codeComponents.ToPayload()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{err.Error()},
			})
			return
		}

		c.Redirect(http.StatusFound, fmt.Sprintf("%s?state=%s&code=%s", request.RedirectUri, request.State, payload))
	}
}

func (l *LoginController) TokenValidate() func(c *gin.Context) {
	return func(c *gin.Context) {
		var request models.OAuthTokenValidationRequest
		if err := c.Bind(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{err.Error()},
			})
			return
		}

		if request.GrantType != "authorization_code" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errors": "requested grant type is not supported",
			})
		}

		codeComponents, err := models.AuthorizationCodeFromPayload(request.Code)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": []string{err.Error()},
			})
			return
		}

		if codeComponents.CodeChallengeMethod != "S256" {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"errors": "code challenge method unsupported",
			})
			return
		}

		codeVerifierHash := sha256.Sum256([]byte(request.CodeVerifier))
		codeVerifierDecoded := base64.RawURLEncoding.EncodeToString(codeVerifierHash[:])

		if string(codeVerifierDecoded) != codeComponents.CodeChallenge {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": "code verification failed",
			})
			return
		}

		t, err := token.Generate(models.UserDetails{
			Name:  codeComponents.UserName,
			Email: codeComponents.UserEmail,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"errors": err.Error(),
			})

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
