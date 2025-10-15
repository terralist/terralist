package controllers

import (
	"net/http"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/authority"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth"
	"terralist/pkg/rbac"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

const (
	authorityApiBase = "/api/authorities"
)

// AuthorityController registers the endpoints to control authorities.
type AuthorityController interface {
	api.RestController
}

// DefaultAuthorityController is a concrete implementation of
// AuthorityController.
type DefaultAuthorityController struct {
	AuthorityService services.AuthorityService
	ApiKeyService    services.ApiKeyService

	Authentication *handlers.Authentication
	Authorization  *handlers.Authorization
}

func (c *DefaultAuthorityController) Paths() []string {
	return []string{authorityApiBase}
}

func (c *DefaultAuthorityController) Subscribe(apis ...*gin.RouterGroup) {
	requireAuthorization := c.Authorization.RequireAuthorization(rbac.ResourceAuthorities)
	authorityComposer := func(ctx *gin.Context) string {
		id, err := uuid.Parse(ctx.Param("id"))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"errors": []string{err.Error()},
			})

			return ""
		}

		authority, err := c.AuthorityService.GetByID(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"errors": []string{err.Error()},
			})

			return ""
		}

		ctx.Set("authority", authority)

		return authority.Name
	}

	api := apis[0]

	api.Use(c.Authentication.AttemptAuthentication())

	// This is a protected endpoint, every request should be authenticated.
	api.Use(c.Authentication.RequireAuthentication())

	api.GET(
		"/",
		func(ctx *gin.Context) {
			authorities, err := c.AuthorityService.GetAll()
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			dtos := lo.Map(
				authorities,
				func(a *authority.Authority, _ int) authority.AuthorityDTO {
					return a.ToDTO()
				})

			// Filter only authorities where the user has access
			// The user key should be preset by the RequireAuthentication middleware.
			user := handlers.MustGetFromContext[auth.User](ctx, "user")

			dtos = lo.Filter(dtos, func(dto authority.AuthorityDTO, idx int) bool {
				return c.Authorization.CanPerform(*user, rbac.ResourceAuthorities, rbac.ActionGet, dto.Name)
			})

			ctx.JSON(http.StatusOK, dtos)
		},
	)

	api.GET(
		"/:id",
		requireAuthorization(rbac.ActionGet, authorityComposer),
		func(ctx *gin.Context) {
			authority := handlers.MustGetFromContext[authority.Authority](ctx, "authority")
			ctx.JSON(http.StatusOK, authority.ToDTO())
		},
	)

	api.POST(
		"/",
		requireAuthorization(rbac.ActionCreate, func(ctx *gin.Context) string {
			var body authority.AuthorityCreateDTO
			if err := ctx.BindJSON(&body); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})

				return ""
			}

			ctx.Set("body", &body)

			return body.Name
		}),
		func(ctx *gin.Context) {
			body := handlers.MustGetFromContext[authority.AuthorityCreateDTO](ctx, "body")

			authority, err := c.AuthorityService.Create(*body)
			if err != nil {
				ctx.JSON(http.StatusConflict, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusCreated, authority)
		},
	)

	api.PATCH(
		"/:id",
		requireAuthorization(rbac.ActionUpdate, authorityComposer),
		func(ctx *gin.Context) {
			id := handlers.MustGetFromContext[authority.Authority](ctx, "authority").ID

			var body authority.AuthorityDTO
			if err := ctx.BindJSON(&body); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			authority, err := c.AuthorityService.Update(id, body)
			if err != nil {
				ctx.JSON(http.StatusConflict, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, authority)
		},
	)

	api.DELETE(
		"/:id",
		requireAuthorization(rbac.ActionDelete, authorityComposer),
		func(ctx *gin.Context) {
			id := handlers.MustGetFromContext[authority.Authority](ctx, "authority").ID

			if err := c.AuthorityService.Delete(id); err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, true)
		},
	)

	api.POST(
		"/:id/keys",
		requireAuthorization(rbac.ActionUpdate, authorityComposer),
		func(ctx *gin.Context) {
			authorityId := handlers.MustGetFromContext[authority.Authority](ctx, "authority").ID

			var body authority.KeyDTO
			if err := ctx.BindJSON(&body); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			dto, err := c.AuthorityService.AddKey(authorityId, body)
			if err != nil {
				ctx.JSON(http.StatusConflict, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, dto)
		},
	)

	api.DELETE(
		"/:id/keys/:keyId",
		requireAuthorization(rbac.ActionDelete, authorityComposer),
		func(ctx *gin.Context) {
			authorityId := handlers.MustGetFromContext[authority.Authority](ctx, "authority").ID

			id, err := uuid.Parse(ctx.Param("keyId"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			if err := c.AuthorityService.RemoveKey(authorityId, id); err != nil {
				ctx.JSON(http.StatusConflict, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, true)
		},
	)

	api.POST(
		"/:id/api-keys",
		requireAuthorization(rbac.ActionUpdate, authorityComposer),
		func(ctx *gin.Context) {
			authorityId := handlers.MustGetFromContext[authority.Authority](ctx, "authority").ID

			var body authority.ApiKeyDTO
			if err := ctx.BindJSON(&body); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			apiKey, err := c.ApiKeyService.Grant(authorityId, body.Name, 0)
			if err != nil {
				ctx.JSON(http.StatusConflict, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{
				"id":   apiKey,
				"name": body.Name,
			})
		},
	)

	api.DELETE(
		"/:id/api-keys/:apiKey",
		requireAuthorization(rbac.ActionUpdate, authorityComposer),
		func(ctx *gin.Context) {
			id, err := uuid.Parse(ctx.Param("apiKey"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			if err := c.ApiKeyService.Revoke(id.String()); err != nil {
				ctx.JSON(http.StatusConflict, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, true)
		},
	)
}
