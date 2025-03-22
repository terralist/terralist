package controllers

import (
	"net/http"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/authority"
	"terralist/internal/server/services"
	"terralist/pkg/api"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ssoroka/slice"
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

	Authorization *handlers.Authorization
}

func (c *DefaultAuthorityController) Paths() []string {
	return []string{authorityApiBase}
}

func (c *DefaultAuthorityController) Subscribe(apis ...*gin.RouterGroup) {
	api := apis[0]
	api.Use(c.Authorization.AnyAuthentication())

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

			dtos := slice.Map[*authority.Authority, authority.AuthorityDTO](
				authorities,
				func(a *authority.Authority) authority.AuthorityDTO {
					return a.ToDTO()
				})

			ctx.JSON(http.StatusOK, dtos)
		},
	)

	api.GET(
		"/:id",
		func(ctx *gin.Context) {
			id, err := uuid.Parse(ctx.Param("id"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			authority, err := c.AuthorityService.GetByID(id)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			ctx.JSON(http.StatusOK, authority.ToDTO())
		},
	)

	api.POST(
		"/",
		func(ctx *gin.Context) {
			var body authority.AuthorityCreateDTO
			if err := ctx.BindJSON(&body); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

			authority, err := c.AuthorityService.Create(body)
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
		func(ctx *gin.Context) {
			id, err := uuid.Parse(ctx.Param("id"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

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
		func(ctx *gin.Context) {
			id, err := uuid.Parse(ctx.Param("id"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

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
		func(ctx *gin.Context) {
			authorityId, err := uuid.Parse(ctx.Param("id"))

			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

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
		func(ctx *gin.Context) {
			authorityId, err := uuid.Parse(ctx.Param("id"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

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
		func(ctx *gin.Context) {
			authorityId, err := uuid.Parse(ctx.Param("id"))

			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

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
		func(ctx *gin.Context) {
			// We don't really need the authority ID to identify the API key
			// but we will keep it like this only to keep the API aligned
			_, err := uuid.Parse(ctx.Param("id"))
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"errors": []string{err.Error()},
				})
				return
			}

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
