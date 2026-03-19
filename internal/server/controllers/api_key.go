package controllers

import (
	"net/http"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/apikey"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/rbac"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

const (
	apiKeyApiBase = "/api/api-keys"
)

// ApiKeyController registers the endpoints to manage standalone API keys.
type ApiKeyController interface {
	api.RestController
}

// DefaultApiKeyController is a concrete implementation of ApiKeyController.
type DefaultApiKeyController struct {
	Service services.StandaloneApiKeyService

	Authentication *handlers.Authentication
	Authorization  *handlers.Authorization
}

func (c *DefaultApiKeyController) Paths() []string {
	return []string{apiKeyApiBase}
}

func (c *DefaultApiKeyController) Subscribe(apis ...*gin.RouterGroup) {
	requireAuthorization := c.Authorization.RequireAuthorization(rbac.ResourceApiKeys)

	api := apis[0]

	api.Use(c.Authentication.AttemptAuthentication())
	api.Use(c.Authentication.RequireAuthentication())

	api.GET(
		"/",
		func(ctx *gin.Context) {
			keys, err := c.Service.List()
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			user := handlers.MustGetFromContext[auth.User](ctx, "user")

			keys = lo.Filter(keys, func(dto apikey.ApiKeyDTO, _ int) bool {
				return c.Authorization.CanPerform(*user, rbac.ResourceApiKeys, rbac.ActionGet, dto.ID)
			})

			ctx.JSON(http.StatusOK, keys)
		},
	)

	api.POST(
		"/",
		requireAuthorization(rbac.ActionCreate, func(ctx *gin.Context) string {
			var body apikey.CreateApiKeyDTO
			if err := ctx.BindJSON(&body); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return ""
			}

			ctx.Set("body", &body)
			return "*"
		}),
		func(ctx *gin.Context) {
			body := handlers.MustGetFromContext[apikey.CreateApiKeyDTO](ctx, "body")
			user := handlers.MustGetFromContext[auth.User](ctx, "user")

			policies := lo.Map(body.Policies, func(p apikey.CreatePolicyDTO, _ int) apikey.Policy {
				return p.ToModel()
			})

			id, err := c.Service.Create(body.Name, user.Email, body.ExpireIn, policies)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusCreated, gin.H{
				"id":   id,
				"name": body.Name,
			})
		},
	)

	api.DELETE(
		"/:id",
		requireAuthorization(rbac.ActionDelete, func(_ *gin.Context) string {
			return "*"
		}),
		func(ctx *gin.Context) {
			id := ctx.Param("id")

			if err := c.Service.Delete(id); err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, true)
		},
	)
}
