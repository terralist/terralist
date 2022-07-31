package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"terralist/internal/server/models/oauth"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/session"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

const (
	authTerraformApiBase = "/v1/auth"
	authDefaultApiBase   = "/v1/api/auth"

	authorizeRoute = "/authorization"
	tokenRoute     = "/token"
	redirectRoute  = "/redirect"
)

// LoginController registers the endpoints required to handle the OAUTH 2.0
// authentication
type LoginController interface {
	api.RestController

	// AuthorizationRoute returns the endpoint where Terraform can call for
	// initiating the authorization process
	AuthorizationRoute() string

	// TokenRoute returns the endpoint where Terraform can call to validate
	// the code components and obtain the authorization token
	TokenRoute() string
}

// DefaultLoginController is a concrete implementation of LoginController
type DefaultLoginController struct {
	Store        session.Store
	LoginService services.LoginService

	HostURL *url.URL

	EncryptSalt string
}

func (c *DefaultLoginController) Paths() []string {
	return []string{
		authTerraformApiBase,
		authDefaultApiBase,
	}
}

func (c *DefaultLoginController) AuthorizationRoute() string {
	return fmt.Sprintf("%s%s", authTerraformApiBase, authorizeRoute)
}

func (c *DefaultLoginController) TokenRoute() string {
	return fmt.Sprintf("%s%s", authTerraformApiBase, tokenRoute)
}

func (c *DefaultLoginController) Subscribe(apis ...*gin.RouterGroup) {
	// tfApi should be compliant with the Terraform Registry Protocol for
	// authentication
	// Docs: https://www.terraform.io/docs/internals/login-protocol.html
	tfApi := apis[0]

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

	// api holds the routes that are not described by the Terraform protocol
	api := apis[1]

	api.GET(redirectRoute, func(ctx *gin.Context) {
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

		codeComponents, erro := c.LoginService.UnpackCode(code, &r)
		if erro != nil {
			ctx.Redirect(http.StatusFound, c.redirectWithError(r.RedirectURI, r.State, erro))
			return
		}

		uri, err := url.Parse(r.RedirectURI)
		if err != nil {
			log.Warn().
				AnErr("Error", err).
				Str("RedirectURI", r.RedirectURI).
				Msg("An invalid redirect URI was detected during the OAUTH callback.")

			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// Check if the call was made from this origin
		if uri.Host == c.HostURL.Host {
			// There's no need in validating the request, if we made this call
			// Save user session and redirect back
			sess, err := c.Store.Get(ctx.Request)
			if err != nil {
				ctx.Redirect(
					http.StatusFound,
					c.redirectWithError(
						uri.String(),
						"",
						oauth.WrapError(
							fmt.Errorf("could not fetch the session"),
							oauth.ServerError,
						),
					),
				)
			}

			sess.Set("user", &auth.User{
				Name:  codeComponents.UserName,
				Email: codeComponents.UserEmail,
			})

			if err := c.Store.Save(ctx.Request, ctx.Writer, sess); err != nil {
				ctx.Redirect(
					http.StatusFound,
					c.redirectWithError(
						uri.String(),
						"",
						oauth.WrapError(
							fmt.Errorf("could not save session"),
							oauth.ServerError,
						),
					),
				)
			}

			// Redirect back
			ctx.Redirect(http.StatusFound, uri.String())
			return
		}

		redirectURL, erro := c.LoginService.Redirect(codeComponents, &r)
		if erro != nil {
			ctx.Redirect(http.StatusFound, c.redirectWithError(r.RedirectURI, r.State, erro))
			return
		}

		ctx.Redirect(http.StatusFound, redirectURL)
	})

}

func (c *DefaultLoginController) redirectWithError(
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
