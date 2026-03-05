package controllers

import (
	"net/http"
	"strings"

	"terralist/internal/server/handlers"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/rbac"

	"github.com/gin-gonic/gin"
)

const (
	settingsCapabilityApiBase = "/api/auth/capabilities"
	settingsCapabilityObject  = "page"
)

// SettingsCapabilityController registers settings capability endpoints.
type SettingsCapabilityController interface {
	api.RestController
}

// DefaultSettingsCapabilityController is a concrete implementation of SettingsCapabilityController.
type DefaultSettingsCapabilityController struct {
	Authentication  *handlers.Authentication
	Authorization   *handlers.Authorization
	AuthorizedUsers string
}

func (c *DefaultSettingsCapabilityController) Paths() []string {
	return []string{settingsCapabilityApiBase}
}

func (c *DefaultSettingsCapabilityController) Subscribe(apis ...*gin.RouterGroup) {
	api := apis[0]
	api.Use(c.Authentication.AttemptAuthentication())
	api.Use(c.Authentication.RequireAuthentication())

	api.GET("/settings", func(ctx *gin.Context) {
		user := handlers.MustGetFromContext[auth.User](ctx, "user")
		allowed := c.Authorization.CanPerform(*user, rbac.ResourceSettings, rbac.ActionGet, settingsCapabilityObject) ||
			canAccessByAuthorizedUsers(*user, c.AuthorizedUsers)
		ctx.JSON(http.StatusOK, gin.H{
			"allowed": allowed,
		})
	})
}

func canAccessByAuthorizedUsers(user auth.User, authorizedUsers string) bool {
	users := strings.Split(authorizedUsers, ",")
	return users[0] == "" || slicesContain(users, user.Name) || slicesContain(users, user.Email)
}

func slicesContain(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
