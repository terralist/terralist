package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

const (
	guestEndpoint         = "/"
	authenticatedEndpoint = "/home"
	logoutEndpoint        = "/logout"

	authorityEndpoint       = "/authority"
	authorityCreateEndpoint = authorityEndpoint + "/create"
	authorityRemoveEndpoint = authorityEndpoint + "/:id/remove"

	authorityKeyBase           = authorityEndpoint + "/:id/keys"
	authorityKeyCreateEndpoint = authorityKeyBase + "/add"
	authorityKeyRemoveEndpoint = authorityKeyBase + "/:kid/remove"

	authorityApiKeyBase           = authorityEndpoint + "/:id/apikeys"
	authorityApiKeyCreateEndpoint = authorityApiKeyBase + "/add"
	authorityApiKeyRemoveEndpoint = authorityApiKeyBase + "/:kid/remove"
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
		"", // empty string so we can map on `/`
	}
}

func (c *DefaultWebController) Subscribe(apis ...*gin.RouterGroup) {
	// Set base templates
	_ = c.UIManager.AddBase("layout.html.tpl")

	// Router group
	r := apis[0]

	loginKey, _ := c.UIManager.Register("login.html.tpl")
	r.GET("/",
		c.checkSession(false),
		func(ctx *gin.Context) {
			c.render(ctx, loginKey, map[string]any{
				"Provider": c.ProviderName,
				"HostURL":  c.HostURL.String(),
				"Endpoints": &map[string]any{
					"Authorization": c.AuthorizationEndpoint,
				},
			})
		},
	)

	homeKey, _ := c.UIManager.Register("home.html.tpl")
	r.GET(
		authenticatedEndpoint,
		c.checkSession(true),
		func(ctx *gin.Context) {
			u, _ := ctx.Get("user")
			user := u.(*auth.User)

			authorities, err := c.AuthorityService.GetAll(user.Email)
			if err != nil {
				log.Error().
					AnErr("Error", err).
					Msg("Cannot fetch user authorities.")
			}

			c.render(ctx, homeKey, map[string]any{
				"User":        user,
				"Authorities": authorities,
				"Endpoints": &map[string]any{
					"Logout":          logoutEndpoint,
					"CreateAuthority": authorityCreateEndpoint,
					"RemoveAuthority": authorityRemoveEndpoint,
					"CreateKey":       authorityKeyCreateEndpoint,
					"RemoveKey":       authorityKeyRemoveEndpoint,
					"CreateApiKey":    authorityApiKeyCreateEndpoint,
					"RemoveApiKey":    authorityApiKeyRemoveEndpoint,
				},
			})
		},
	)

	r.GET(
		logoutEndpoint,
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

	// Authority group
	authorityCreateKey, _ := c.UIManager.Register("authority/create.html.tpl")
	r.GET(
		authorityCreateEndpoint,
		c.checkSession(true),
		func(ctx *gin.Context) {
			c.render(ctx, authorityCreateKey, map[string]any{
				"Title": "Create Authority",
			})
		},
	)

	r.POST(
		authorityCreateEndpoint,
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
				c.returnWithErr(ctx, authorityCreateEndpoint, err)
			}

			ctx.Redirect(http.StatusFound, authenticatedEndpoint)
		},
	)

	r.GET(
		authorityRemoveEndpoint,
		c.checkSession(true),
		func(ctx *gin.Context) {
			id := ctx.Param("id")

			authorityID, err := uuid.Parse(id)
			if err != nil {
				c.returnWithErr(ctx, authenticatedEndpoint, fmt.Errorf("invalid authority ID: %v", err))
			}

			if err := c.AuthorityService.Delete(authorityID); err != nil {
				c.returnWithErr(ctx, authenticatedEndpoint, err)
			}

			ctx.Redirect(http.StatusFound, authenticatedEndpoint)
		},
	)

	authorityKeyAddKey, _ := c.UIManager.Register("authority/key/add.html.tpl")
	r.GET(
		authorityKeyCreateEndpoint,
		c.checkSession(true),
		func(ctx *gin.Context) {
			c.render(ctx, authorityKeyAddKey, map[string]any{
				"Title": "Add Authority Key",
			})
		},
	)

	r.POST(
		authorityKeyCreateEndpoint,
		c.checkSession(true),
		func(ctx *gin.Context) {
			id := ctx.Param("id")

			authorityID, err := uuid.Parse(id)
			if err != nil {
				c.returnWithErr(ctx, authenticatedEndpoint, fmt.Errorf("invalid authority ID: %v", err))
			}

			keyId := ctx.PostForm("key_id")
			asciiArmor := ctx.PostForm("ascii_armor")
			trustSignature := ctx.PostForm("trust_signature")

			if err := c.AuthorityService.AddKey(authorityID, authority.KeyDTO{
				KeyId:          keyId,
				AsciiArmor:     asciiArmor,
				TrustSignature: trustSignature,
			}); err != nil {
				c.returnWithErr(ctx, strings.Replace(authorityKeyCreateEndpoint, ":id", id, 1), err)
			}

			ctx.Redirect(http.StatusFound, authenticatedEndpoint)
		},
	)

	r.GET(
		authorityKeyRemoveEndpoint,
		c.checkSession(true),
		func(ctx *gin.Context) {
			aID := ctx.Param("id")
			kID := ctx.Param("kid")

			authorityID, err1 := uuid.Parse(aID)
			keyID, err2 := uuid.Parse(kID)
			if err1 != nil || err2 != nil {
				c.returnWithErr(ctx, authenticatedEndpoint, fmt.Errorf("invalid authority ID or key ID"))
			}

			if err := c.AuthorityService.RemoveKey(authorityID, keyID); err != nil {
				c.returnWithErr(ctx, authenticatedEndpoint, err)
			}

			ctx.Redirect(http.StatusFound, authenticatedEndpoint)
		},
	)

	r.GET(
		authorityApiKeyCreateEndpoint,
		c.checkSession(true),
		func(ctx *gin.Context) {
			id := ctx.Param("id")

			authorityID, err := uuid.Parse(id)
			if err != nil {
				c.returnWithErr(ctx, authenticatedEndpoint, fmt.Errorf("invalid authority ID"))
			}

			if _, err := c.ApiKeyService.Grant(authorityID, 0); err != nil {
				c.returnWithErr(ctx, authenticatedEndpoint, err)
			}

			ctx.Redirect(http.StatusFound, authenticatedEndpoint)
		},
	)

	r.GET(
		authorityApiKeyRemoveEndpoint,
		c.checkSession(true),
		func(ctx *gin.Context) {
			apiKey := ctx.Param("kid")

			err := c.ApiKeyService.Revoke(apiKey)
			if err != nil {
				c.returnWithErr(ctx, authenticatedEndpoint, err)
			}

			ctx.Redirect(http.StatusFound, authenticatedEndpoint)
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
			ctx.Redirect(http.StatusFound, authenticatedEndpoint)
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

func (c *DefaultWebController) render(ctx *gin.Context, key string, values map[string]any) {
	qErr := ctx.Query("error")
	qErrDesc := ctx.Query("error_description")

	if qErr != "" || qErrDesc != "" {
		values["Error"] = &map[string]string{
			"Name":        ctx.Query("error"),
			"Description": ctx.Query("error_description"),
		}
	}

	if err := c.UIManager.Render(ctx.Writer, key, values); err != nil {
		log.Error().
			AnErr("Error", err).
			Str("Template Key", key).
			Msg("Cannot render view")

		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
}

func (c *DefaultWebController) returnWithErr(ctx *gin.Context, endpoint string, err error) {
	ctx.Redirect(
		http.StatusFound,
		fmt.Sprintf(
			"%s?error=%s&error_description=%s",
			endpoint,
			url.QueryEscape("Something Went Wrong"),
			url.QueryEscape(err.Error()),
		),
	)

	ctx.Abort()
}
