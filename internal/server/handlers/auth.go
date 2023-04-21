package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"terralist/internal/server/services"
	"terralist/pkg/auth"
	"terralist/pkg/auth/jwt"
	"terralist/pkg/session"

	"github.com/gin-gonic/gin"
)

type authFn = func(c *gin.Context) (*auth.User, error)

var (
	ErrMissing       = errors.New("missing")
	ErrInvalidFormat = errors.New("invalid format")
	ErrInvalidValue  = errors.New("token either expired or inexistent")
)

type Authorization struct {
	ApiKeyService services.ApiKeyService
	JWT           jwt.JWT
	Store         session.Store
}

func (a *Authorization) hasAuthorizationHeader(c *gin.Context) (*auth.User, error) {
	header := c.GetHeader("Authorization")
	if header == "" {
		return nil, fmt.Errorf("Authorization: %w", ErrMissing)
	}

	if !strings.HasPrefix(header, "Bearer ") {
		return nil, fmt.Errorf("Authorization: %w", ErrInvalidFormat)
	}

	var bearerToken string
	_, err := fmt.Sscanf(header, "Bearer %s", &bearerToken)
	if err != nil {
		return nil, fmt.Errorf("Authorization: %w", ErrInvalidFormat)
	}

	var user *auth.User

	if !strings.HasPrefix(bearerToken, "x-api-key:") {
		user, err = a.JWT.Extract(bearerToken)
		if err != nil {
			return nil, fmt.Errorf("Authorization: %w", ErrInvalidValue)
		}
	} else {
		var apiKey string
		if _, err := fmt.Sscanf(bearerToken, "x-api-key:%s", &apiKey); err != nil {
			return nil, fmt.Errorf("Authorization: %w", ErrInvalidFormat)
		}

		user, err = a.ApiKeyService.GetUserDetails(apiKey)
		if err != nil {
			return nil, fmt.Errorf("Authorization: %w", ErrInvalidValue)
		}
	}

	return user, nil
}

func (a *Authorization) hasAPIKeyHeader(c *gin.Context) (*auth.User, error) {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		return nil, fmt.Errorf("X-API-Key: %w", ErrMissing)
	}

	user, err := a.ApiKeyService.GetUserDetails(apiKey)
	if err != nil {
		return nil, fmt.Errorf("X-API-Key: %w", ErrInvalidValue)
	}

	return user, nil
}

func (a *Authorization) hasActiveSession(c *gin.Context) (*auth.User, error) {
	sess, err := a.Store.Get(c.Request)
	if err != nil {
		return nil, fmt.Errorf("session: %w", ErrMissing)
	}

	user, ok := sess.Get("user")
	if !ok || user == nil {
		return nil, fmt.Errorf("session: %w", ErrMissing)
	}

	authUser, ok := user.(*auth.User)
	if !ok {
		sess.Set("user", nil)
		a.Store.Save(c.Request, c.Writer, sess)
		return nil, fmt.Errorf("session: %w", ErrInvalidValue)
	}

	return authUser, nil
}

func (a *Authorization) verifyRules(c *gin.Context, fns []authFn) {
	users := make([]*auth.User, 0, len(fns))
	errs := make([]string, 0, len(fns))
	for _, f := range fns {
		user, err := f(c)

		if user != nil {
			users = append(users, user)
		} else {
			errs = append(errs, err.Error())
		}
	}

	if len(users) == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"errors": errs,
		})
		return
	}

	// The precedence is given by the fns slice
	user := users[0]

	c.Set("userName", user.Name)
	c.Set("userEmail", user.Email)

	if user.AuthorityID != "" {
		c.Set("authority", user.AuthorityID)
	}

	c.Next()
}

func (a *Authorization) ApiAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		fns := []authFn{
			a.hasAuthorizationHeader,
			a.hasAPIKeyHeader,
		}

		a.verifyRules(c, fns)
	}
}

func (a *Authorization) SessionAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		fns := []authFn{
			a.hasActiveSession,
		}

		a.verifyRules(c, fns)
	}
}

func (a *Authorization) AnyAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		fns := []authFn{
			a.hasAuthorizationHeader,
			a.hasAPIKeyHeader,
			a.hasActiveSession,
		}

		a.verifyRules(c, fns)
	}
}

func RequireAuthority() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Get("authority"); !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errors": []string{"An API key is required to perform this operation."},
			})
			return
		}

		c.Next()
	}
}
