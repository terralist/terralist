package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"terralist/internal/server/models/authority"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/session"
	"terralist/pkg/webui"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// WebController registers the endpoints for web interface
type WebController interface {
	api.RestController
}

// DefaultWebController is a concrete implementation of
// WebController
type DefaultWebController struct {
	Store            session.Store
	UIManager        webui.Manager
	AuthorityService services.AuthorityService
	ApiKeyService    services.ApiKeyService

	ProviderName          string
	HostURL               *url.URL
	AuthorizationEndpoint string
}

func (c *DefaultWebController) Paths() []string {
	return []string{
		"",           // home
		"/error",     // errors
		"/authority", // authority
	}
}

func (c *DefaultWebController) Subscribe(apis ...*gin.RouterGroup) {
	homeGroup := apis[0]

	_ = c.UIManager.AddBase("layout.html.tpl")

	loginKey, _ := c.UIManager.Register("login.html.tpl")
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

	homeKey, _ := c.UIManager.Register("home.html.tpl")
	homeGroup.GET(
		"/home",
		c.checkSession(true),
		func(ctx *gin.Context) {
			u, _ := ctx.Get("user")
			user := u.(*auth.User)

			authorities, err := c.AuthorityService.GetAll(user.Email)
			if err != nil {
				log.Debug().
					AnErr("Error", err).
					Msg("Cannot fetch user authorities.")
			}

			if err := c.UIManager.Render(ctx.Writer, homeKey, &map[string]any{
				"User":        user,
				"Authorities": authorities,
			}); err != nil {
				log.Debug().AnErr("Error", err).Msg("Cannot render home view.")
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

	// Error group
	errorGroup := apis[1]

	errorKey, _ := c.UIManager.Register("error.html.tpl")
	errorGroup.GET("/", func(ctx *gin.Context) {
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

	// Authority group
	authorityGroup := apis[2]

	authorityCreateKey, _ := c.UIManager.Register("authority/create.html.tpl")
	authorityGroup.GET(
		"/create",
		c.checkSession(true),
		func(ctx *gin.Context) {
			if err := c.UIManager.Render(ctx.Writer, authorityCreateKey, &map[string]any{
				"Title": "Create Authority",
			}); err != nil {
				log.Debug().AnErr("Error", err).Msg("Cannot render authority create view.")
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		},
	)

	authorityGroup.POST(
		"/create",
		c.checkSession(true),
		func(ctx *gin.Context) {
			u, _ := ctx.Get("user")
			user := u.(*auth.User)

			name := ctx.PostForm("name")
			policyURL := ctx.PostForm("policy_url")

			if err := c.AuthorityService.Create(authority.AuthorityCreateDTO{
				Name:      name,
				PolicyURL: policyURL,
				Owner:     user.Email,
			}); err != nil {
				ctx.Redirect(
					http.StatusFound,
					fmt.Sprintf(
						"/authority/create?error=%s&error_description=%s",
						url.QueryEscape("something_went_wrong"),
						url.QueryEscape(err.Error()),
					),
				)

				return
			}

			ctx.Redirect(http.StatusFound, "/home")
		},
	)

	authorityGroup.GET(
		"/delete/:id",
		c.checkSession(true),
		func(ctx *gin.Context) {
			id := ctx.Param("id")

			authorityID, err := uuid.Parse(id)
			if err != nil {
				ctx.Redirect(
					http.StatusFound,
					fmt.Sprintf(
						"/home?error=%s&error_description=%s",
						url.QueryEscape("bad_request"),
						url.QueryEscape("invalid authority ID"),
					),
				)

				return
			}

			if err := c.AuthorityService.Delete(authorityID); err != nil {
				ctx.Redirect(
					http.StatusFound,
					fmt.Sprintf(
						"/home?error=%s&error_description=%s",
						url.QueryEscape("internal_server_error"),
						url.QueryEscape(err.Error()),
					),
				)

				return
			}

			ctx.Redirect(http.StatusFound, "/home")
		},
	)

	authorityKeyAddKey, _ := c.UIManager.Register("authority/key/add.html.tpl")
	authorityGroup.GET(
		":id/keys/add",
		c.checkSession(true),
		func(ctx *gin.Context) {
			if err := c.UIManager.Render(ctx.Writer, authorityKeyAddKey, &map[string]any{
				"Title": "Add Authority Key",
			}); err != nil {
				log.Debug().AnErr("Error", err).Msg("Cannot render authority authority key add view.")
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		},
	)

	authorityGroup.POST(
		":id/keys/add",
		c.checkSession(true),
		func(ctx *gin.Context) {
			id := ctx.Param("id")

			authorityID, err := uuid.Parse(id)
			if err != nil {
				ctx.Redirect(
					http.StatusFound,
					fmt.Sprintf(
						"/home?error=%s&error_description=%s",
						url.QueryEscape("bad_request"),
						url.QueryEscape("invalid authority ID"),
					),
				)

				return
			}

			keyId := ctx.PostForm("key_id")
			asciiArmor := ctx.PostForm("ascii_armor")
			trustSignature := ctx.PostForm("trust_signature")

			if err := c.AuthorityService.AddKey(authorityID, authority.KeyDTO{
				KeyId:          keyId,
				AsciiArmor:     asciiArmor,
				TrustSignature: trustSignature,
			}); err != nil {
				ctx.Redirect(
					http.StatusFound,
					fmt.Sprintf(
						"/authority/%s/keys/add?error=%s&error_description=%s",
						id,
						url.QueryEscape("something_went_wrong"),
						url.QueryEscape(err.Error()),
					),
				)

				return
			}

			ctx.Redirect(http.StatusFound, "/home")
		},
	)

	authorityGroup.GET(
		":id/keys/:kid/remove",
		c.checkSession(true),
		func(ctx *gin.Context) {
			aID := ctx.Param("id")
			kID := ctx.Param("kid")

			authorityID, err1 := uuid.Parse(aID)
			keyID, err2 := uuid.Parse(kID)
			if err1 != nil || err2 != nil {
				ctx.Redirect(
					http.StatusFound,
					fmt.Sprintf(
						"/home?error=%s&error_description=%s",
						url.QueryEscape("bad_request"),
						url.QueryEscape("invalid authority ID or key ID"),
					),
				)

				return
			}

			if err := c.AuthorityService.RemoveKey(authorityID, keyID); err != nil {
				ctx.Redirect(
					http.StatusFound,
					fmt.Sprintf(
						"/home?error=%s&error_description=%s",
						url.QueryEscape("internal_server_error"),
						url.QueryEscape(err.Error()),
					),
				)

				return
			}

			ctx.Redirect(http.StatusFound, "/home")
		},
	)

	authorityGroup.GET(
		":id/apikeys/add",
		c.checkSession(true),
		func(ctx *gin.Context) {
			id := ctx.Param("id")

			authorityID, err := uuid.Parse(id)
			if err != nil {
				ctx.Redirect(
					http.StatusFound,
					fmt.Sprintf(
						"/home?error=%s&error_description=%s",
						url.QueryEscape("bad_request"),
						url.QueryEscape("invalid authority ID"),
					),
				)

				return
			}

			if _, err := c.ApiKeyService.Grant(authorityID, 0); err != nil {
				ctx.Redirect(
					http.StatusFound,
					fmt.Sprintf(
						"/home?error=%s&error_description=%s",
						url.QueryEscape("internal_server_error"),
						url.QueryEscape(err.Error()),
					),
				)

				return
			}

			ctx.Redirect(http.StatusFound, "/home")
		},
	)

	authorityGroup.GET(
		"/apikeys/:id/remove",
		c.checkSession(true),
		func(ctx *gin.Context) {
			apiKey := ctx.Param("id")

			err := c.ApiKeyService.Revoke(apiKey)
			if err != nil {
				ctx.Redirect(
					http.StatusFound,
					fmt.Sprintf(
						"/home?error=%s&error_description=%s",
						url.QueryEscape("internal_server_error"),
						url.QueryEscape(err.Error()),
					),
				)

				return
			}

			ctx.Redirect(http.StatusFound, "/home")
		},
	)
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
