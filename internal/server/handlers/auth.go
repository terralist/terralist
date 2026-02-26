package handlers

import (
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"terralist/internal/server/services"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/rbac"
	"terralist/pkg/session"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var (
	ErrMissing          = errors.New("missing definitive scope")
	ErrInvalidFormat    = errors.New("invalid format")
	ErrInvalidValue     = errors.New("token either expired or inexistent")
	ErrUnexpectedOrigin = errors.New("unexpected source of authentication")
)

type Authentication struct {
	ApiKeyService services.ApiKeyService
	JWT           jwt.JWT
	Store         session.Store
}

// parseTerraformCLI parses a request context, and, if the user is authenticated
// from the Terraform CLI it returns the user.
// If the user is authenticated, but not from TerraformCLI, it returns an
// ErrUnexpectedOrigin error.
func (a *Authentication) parseTerraformCLI(c *gin.Context) (*auth.User, error) {
	header := c.GetHeader("Authorization")
	if header == "" {
		return nil, fmt.Errorf("%w: Authorization header not set", ErrMissing)
	}

	if !strings.HasPrefix(header, "Bearer ") {
		return nil, fmt.Errorf("%w: Authorization header not properly set", ErrInvalidFormat)
	}

	var bearerToken string
	_, err := fmt.Sscanf(header, "Bearer %s", &bearerToken)
	if err != nil {
		return nil, fmt.Errorf("%w: no bearer token found", ErrInvalidFormat)
	}

	if strings.HasPrefix(bearerToken, "x-api-key:") {
		return nil, fmt.Errorf("%w: api-key", ErrUnexpectedOrigin)
	}

	user, err := a.JWT.Extract(bearerToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidValue, err)
	}

	return user, nil
}

// parseApiKey parses a request context, and, if the user is authenticated
// from an API Key it returns the user.
// If the user is authenticated, but not from an API Key, it returns an
// ErrUnexpectedOrigin error.
func (a *Authentication) parseApiKey(c *gin.Context) (*auth.User, error) {
	authorizationHeader := c.GetHeader("Authorization")
	apiKeyHeader := c.GetHeader("X-API-Key")
	if authorizationHeader == "" && apiKeyHeader == "" {
		return nil, fmt.Errorf("%w: neither Authorization nor X-API-Key headers are set", ErrMissing)
	}

	apiKey := ""
	if authorizationHeader != "" {
		if !strings.HasPrefix(authorizationHeader, "Bearer ") {
			return nil, fmt.Errorf("%w: Authorization header not properly set", ErrInvalidFormat)
		}

		var bearerToken string
		_, err := fmt.Sscanf(authorizationHeader, "Bearer %s", &bearerToken)
		if err != nil {
			return nil, fmt.Errorf("%w: Authorization header cannot be parsed", ErrInvalidFormat)
		}

		if !strings.HasPrefix(bearerToken, "x-api-key:") {
			return nil, fmt.Errorf("%w: terraform-cli", ErrUnexpectedOrigin)
		}

		if _, err := fmt.Sscanf(bearerToken, "x-api-key:%s", &apiKey); err != nil {
			return nil, fmt.Errorf("%w: no api-key found", ErrInvalidFormat)
		}
	} else {
		apiKey = apiKeyHeader
	}

	user, err := a.ApiKeyService.GetUserDetails(apiKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidValue, err)
	}

	return user, nil
}

// parseActiveSession parses a request context, and, if the user is authenticated
// with an active session it returns the user.
// If the user is authenticated, but not with an active session, it returns an
// ErrUnexpectedOrigin error.
func (a *Authentication) parseActiveSession(c *gin.Context) (*auth.User, error) {
	sess, err := a.Store.Get(c.Request)
	if err != nil {
		return nil, fmt.Errorf("%w: no session", ErrMissing)
	}

	user, ok := sess.Get("user")
	if !ok || user == nil {
		return nil, fmt.Errorf("%w: no user in session", ErrMissing)
	}

	authUser, ok := user.(*auth.User)
	if !ok {
		sess.Set("user", nil)
		if err := a.Store.Save(c.Request, c.Writer, sess); err != nil {
			return nil, fmt.Errorf("while trying to extract user from session: %w, could not save existing session: %v", ErrInvalidValue, err)
		}

		return nil, fmt.Errorf("%w: unknown user format in session", ErrInvalidValue)
	}

	return authUser, nil
}

// parseUser iteratively check all possible authentication methods and selects
// the first one that validates the user.
func (a *Authentication) parseUser(c *gin.Context) (*auth.User, []error) {
	users := make([]*auth.User, 3)
	errs := make([]error, 3)

	users[0], errs[0] = a.parseTerraformCLI(c)
	users[1], errs[1] = a.parseApiKey(c)
	users[2], errs[2] = a.parseActiveSession(c)

	user := lo.Reduce(users, func(acc *auth.User, cur *auth.User, _ int) *auth.User {
		if acc != nil {
			return acc
		}

		if cur != nil {
			return cur
		}

		return acc
	}, nil)

	if user == nil {
		log.Debug().
			Ctx(c).
			Any("users", users).
			Errs("errors", errs).
			Msg("Cannot find any authenticated user.")
		return nil, errs
	}

	return user, nil
}

// AttemptAuthentication is a gin handler that attempts to get the authenticated user.
func (a *Authentication) AttemptAuthentication() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, errs := a.parseUser(ctx)
		if errs != nil {
			ctx.Set("authErrors", lo.Map(errs, func(err error, _ int) string { return err.Error() }))
			return
		}

		ctx.Set("user", user)
		ctx.Set("userName", user.Name)
		ctx.Set("userEmail", user.Email)

		if user.Authority != "" {
			ctx.Set("authorityName", user.Authority)
		}

		if user.AuthorityID != "" {
			ctx.Set("authorityID", user.AuthorityID)
		}
	}
}

// RequireAuthentication is a gin handler that ensures the user is authenticated.
// Any other handler that is executed after this one should query the context to
// retrieve the authenticated user.
func (a *Authentication) RequireAuthentication() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if _, ok := ctx.Get("user"); !ok {
			if authErrors, ok := ctx.Get("authErrors"); ok {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"errors": authErrors,
				})
			} else {
				ctx.AbortWithStatus(http.StatusUnauthorized)
			}

			return
		}
	}
}

// Authorization ...
type Authorization struct {
	Enforcer         *rbac.Enforcer
	AuthorityService services.AuthorityService
}

// CanPerform checks if a given subject can perform an action on a specified object
// from a given resource API group.
func (a *Authorization) CanPerform(subject auth.User, resource, action, object string) bool {
	logger := log.With().
		Str("user", subject.Name).
		Str("authority", subject.Authority).
		Str("authorityID", subject.AuthorityID).
		Str("resource", resource).
		Str("action", action).
		Logger()

	// Enforce authority isolation for API key authenticated users.
	// If user has an authority (from API key), they can only access their own authority's resources.
	if subject.AuthorityID != "" && slices.Contains([]string{rbac.ResourceModules, rbac.ResourceProviders}, resource) {
		if parts := strings.Split(object, "/"); len(parts) > 0 {
			requestedNamespace := parts[0]
			if !strings.EqualFold(requestedNamespace, subject.Authority) {
				logger.Debug().
					Str("requestedNamespace", requestedNamespace).
					Msg("API key denied access to different authority")
				return false
			}
		}
	}

	if slices.Contains([]string{rbac.ResourceModules, rbac.ResourceProviders}, resource) && action == rbac.ActionGet {
		if parts := strings.Split(object, "/"); len(parts) >= 0 {
			authorityName := parts[0]
			// TODO: This should be cached server-side with a small TTL - a couple of minutes.
			if authority, err := a.AuthorityService.GetByName(authorityName); err != nil {
				logger.Error().
					Str("resourceAuthority", authorityName).
					Err(err).
					Msg("Could not fetch authority by name.")
			} else {
				if authority.Public {
					logger.Debug().
						Str("resourceAuthority", authorityName).
						Msg("Authorizing request as authority is marked as public.")

					return true
				}
			}
		}
	}

	if err := a.Enforcer.Protect(subject, resource, action, object); err != nil {
		logger.Debug().
			Err(err).
			Msg("User not authorized")

		return false
	}

	return true
}

// RequireAuthorization is a wrapper function that returns a custom function that can
// be used to generate middlewares to handle authorization for a specific action.
func (a *Authorization) RequireAuthorization(resource string) func(action string, objectFn func(c *gin.Context) string) gin.HandlerFunc {
	return func(action string, objectFn func(c *gin.Context) string) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			user, err := GetFromContext[auth.User](ctx, "user")
			if err != nil {
				// If the user is not found in the context, we're going to authenticate them as
				// anonymous. It is possible that some resource groups to expose publicly some
				// objects.

				user = &auth.User{
					Name: rbac.SubjectAnonymous,
				}
			}

			object := objectFn(ctx)

			if !a.CanPerform(*user, resource, action, object) {
				ctx.AbortWithStatus(http.StatusForbidden)
				return
			}
		}
	}
}

func RequireAuthority() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Get("authorityID"); !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"errors": []string{"An API key is required to perform this operation."},
			})
			return
		}
	}
}
