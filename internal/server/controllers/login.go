package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"terralist/internal/server/models/oauth"
	"terralist/internal/server/services"

	"github.com/gin-gonic/gin"
)

const (
	authorizeRoute = "/authorization"
	tokenRoute     = "/token"
	redirectRoute  = "/redirect"
)

type LoginController struct {
	LoginService services.LoginService

	EncryptSalt string
}

func (c *LoginController) TerraformApiBase() string {
	return "/v1/auth"
}

func (c *LoginController) ApiBase() string {
	return "/v1/api/auth"
}

func (c *LoginController) AuthorizationRoute() string {
	return fmt.Sprintf("%s%s", c.TerraformApiBase(), authorizeRoute)
}

func (c *LoginController) TokenRoute() string {
	return fmt.Sprintf("%s%s", c.TerraformApiBase(), tokenRoute)
}

func (c *LoginController) RedirectRoute() string {
	return fmt.Sprintf("%s%s", c.TerraformApiBase(), redirectRoute)
}

func (c *LoginController) Subscribe(tfApi *gin.RouterGroup, api *gin.RouterGroup) {
	// tfApi should be compliant with the Terraform Registry Protocol for
	// authentication
	// Docs: https://www.terraform.io/docs/internals/login-protocol.html

	tfApi.GET(authorizeRoute, func(ctx *gin.Context) {
		r := &oauth.Request{
			ClientID:            ctx.Query("client_id"),
			CodeChallenge:       ctx.Query("code_challenge"),
			CodeChallengeMethod: ctx.Query("code_challenge_method"),
			RedirectURI:         ctx.Query("redirect_uri"),
			ResponseType:        ctx.Query("response_type"),
			State:               ctx.Query("state"),
		}

		state, err := r.ToPayload(c.EncryptSalt)
		if err != nil {
			ctx.Redirect(
				http.StatusFound,
				c.redirectWithError(r.RedirectURI, r.State, oauth.WrapError(err, oauth.ServerError)),
			)
			return
		}

		authorizeURL, erro := c.LoginService.Authorize(state)
		if erro != nil {
			ctx.Redirect(http.StatusFound, c.redirectWithError(r.RedirectURI, r.State, erro))
			return
		}

		ctx.Redirect(http.StatusFound, authorizeURL)
	})

	tfApi.GET(redirectRoute, func(ctx *gin.Context) {
		code := ctx.Query("code")
		state := ctx.Query("state")

		r, err := oauth.Payload(state).ToRequest(c.EncryptSalt)
		if err != nil {
			ctx.Redirect(
				http.StatusFound,
				c.redirectWithError(r.RedirectURI, r.State, oauth.WrapError(err, oauth.InvalidRequest)),
			)
			return
		}

		redirectURL, erro := c.LoginService.Redirect(code, &r)
		if erro != nil {
			ctx.Redirect(http.StatusFound, c.redirectWithError(r.RedirectURI, r.State, erro))
			return
		}

		ctx.Redirect(http.StatusFound, redirectURL)
	})

	tfApi.POST(tokenRoute, func(ctx *gin.Context) {
		var r oauth.TokenValidationRequest
		if err := ctx.Bind(&r); err != nil {
			// if we catch an error, we don't know where to redirect, just exit the routine
			return
		}

		if r.GrantType != "authorization_code" {
			ctx.Redirect(
				http.StatusFound,
				c.redirectWithError(
					r.RedirectURI,
					"",
					oauth.WrapError(fmt.Errorf("requested grant type is not supported"), oauth.AccessDenied),
				),
			)
			return
		}

		codeComponents, err := oauth.Payload(r.Code).ToCodeComponents(c.EncryptSalt)
		if err != nil {
			ctx.Redirect(
				http.StatusFound,
				c.redirectWithError(r.RedirectURI, "", oauth.WrapError(err, oauth.InvalidRequest)),
			)
			return
		}

		resp, erro := c.LoginService.ValidateToken(&codeComponents, r.CodeVerifier)
		if erro != nil {
			ctx.Redirect(http.StatusFound, c.redirectWithError(r.RedirectURI, "", erro))
			return
		}

		ctx.JSON(http.StatusOK, resp)
	})
}

func (c *LoginController) redirectWithError(
	uri string,
	state string,
	err oauth.Error,
) string {
	stateQuery := ""
	if state != "" {
		stateQuery = fmt.Sprintf("&state=%s", state)
	}

	return fmt.Sprintf(
		"%s?error=%s&error_description=%s%s",
		uri,
		err.Kind(),
		url.QueryEscape(err.Error()),
		stateQuery,
	)
}
