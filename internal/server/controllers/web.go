package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"terralist/pkg/api"
	"terralist/pkg/builders"
	"terralist/pkg/session"
	"terralist/pkg/webui"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// WebController registers the endpoints for web interface
type WebController interface {
	api.RestController
}

// DefaultWebController is a concrete implementation of
// WebController
type DefaultWebController struct {
	Store     session.Store
	UIManager webui.Manager

	ProviderName          string
	HostURL               *url.URL
	AuthorizationEndpoint string
}

func (c *DefaultWebController) Paths() []string {
	return []string{
		"",       // home
		"/error", // errors
	}
}

func (c *DefaultWebController) Subscribe(apis ...*gin.RouterGroup) {
	homeGroup := apis[0]

	loginKey, _ := c.UIManager.Register(
		builders.
			NewSliceBuilder[string]().
			Add("layout.html.tpl").
			Add("login.html.tpl").
			Build(),
	)

	homeGroup.GET("/",
		c.checkSession(false),
		func(ctx *gin.Context) {
			authError := ctx.Query("error")
			authErrorDescription := ctx.Query("error_description")

			if err := c.UIManager.Render(ctx.Writer, loginKey, &map[string]string{
				"Provider":              c.ProviderName,
				"AuthorizationEndpoint": c.AuthorizationEndpoint,
				"HostURL":               c.HostURL.String(),

				// Handle oauth response errors
				"Error":            authError,
				"ErrorDescription": authErrorDescription,
			}); err != nil {
				log.Debug().AnErr("Error", err).Msg("Cannot render login view")
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		},
	)

	homeKey, _ := c.UIManager.Register(
		builders.
			NewSliceBuilder[string]().
			Add("layout.html.tpl").
			Add("home.html.tpl").
			Build(),
	)

	homeGroup.GET(
		"/home",
		c.checkSession(true),
		func(ctx *gin.Context) {
			user, _ := ctx.Get("user")
			if err := c.UIManager.Render(ctx.Writer, homeKey, &map[string]any{
				"User": user,
			}); err != nil {
				log.Debug().AnErr("Error", err).Msg("Cannot render home view")
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		},
	)

	homeGroup.GET(
		"/logout",
		c.checkSession(true),
		func(ctx *gin.Context) {
			// Session must be valid, since the checkSession handler passed
			sess, _ := c.Store.Get(ctx.Request)
			sess.Set("user", nil)
			c.Store.Save(ctx.Request, ctx.Writer, sess)

			// Redirect to login page
			ctx.Redirect(http.StatusFound, "/")
		},
	)

	errorGroup := apis[1]

	errorKey, _ := c.UIManager.Register(
		builders.
			NewSliceBuilder[string]().
			Add("layout.html.tpl").
			Add("error.html.tpl").
			Build(),
	)

	errorGroup.GET("/error", func(ctx *gin.Context) {
		if err := c.UIManager.Render(ctx.Writer, errorKey, &map[string]string{
			"Error":       ctx.Query("error"),
			"Description": ctx.Query("error_description"),
		}); err != nil {
			ctx.AbortWithError(
				http.StatusServiceUnavailable,
				fmt.Errorf("service unavailable"),
			)
		}
	})
}

func (c *DefaultWebController) checkSession(mustBe bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionActive := false
		sess, err := c.Store.Get(ctx.Request)
		if err == nil {
			user, ok := sess.Get("user")
			if ok && user != nil {
				// Pass user details to request
				ctx.Set("user", user)
				sessionActive = true
			}
		}

		// If session is active and it should not be active
		// redirect user to home page (authenticated)
		if sessionActive && !mustBe {
			ctx.Redirect(http.StatusFound, "/home")
			return
		}

		// If session is not active and it should be active
		// redirect user to login page (non-authenticated)
		if !sessionActive && mustBe {
			ctx.Redirect(
				http.StatusFound,
				fmt.Sprintf(
					"/?error=%s&error_description=%s",
					url.QueryEscape("access_denied"),
					url.QueryEscape("you must be authenticated to access this resource"),
				),
			)
			return
		}

		ctx.Next()
	}
}
