package controllers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/valentindeaconu/terralist/models/authorization"
	"github.com/valentindeaconu/terralist/oauth"
	"github.com/valentindeaconu/terralist/oauth/provider"
	"github.com/valentindeaconu/terralist/oauth/token"
	"github.com/valentindeaconu/terralist/settings"
)

type OAuthAccessResponse struct {
	AccessToken string `json:"access_token"`
}

type OAuthTokenValidationRequest struct {
	ClientID     string `form:"client_id"`
	Code         string `form:"code"`
	CodeVerifier string `form:"code_verifier"`
	GrantType    string `form:"grant_type"`
	RedirectUri  string `form:"redirect_uri"`
}

var (
	clientID     string = "97f782ace0a77ca03799"
	clientSecret string = "bff2b051b17144502622d7d7502f10422e6b6e8c"
)

func Authorize(c *gin.Context) {
	request := authorization.AuthorizationRequest{
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

	url := provider.Handler().GetAuthorizeUrl(state)

	c.Redirect(
		http.StatusFound,
		url,
	)
}

func Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	request, err := authorization.AuthorizationRequestFromPayload(state)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": []string{err.Error()},
		})
		return
	}

	var userDetails oauth.UserDetails
	if err := provider.Handler().GetUserDetails(code, &userDetails); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": []string{err.Error()},
		})
	}

	codeComponents := authorization.AuthorizationCodeComponents{
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

func TokenValidate(c *gin.Context) {
	var request OAuthTokenValidationRequest
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

	codeComponents, err := authorization.AuthorizationCodeFromPayload(request.Code)

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

	t, err := token.Generate(oauth.UserDetails{
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
