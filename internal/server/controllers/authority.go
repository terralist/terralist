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

// AuthorityController registers the endpoints to control authorities
type AuthorityController interface {
	api.RestController
}

// DefaultAuthorityController is a concrete implementation of
// AuthorityController
type DefaultAuthorityController struct {
	AuthorityService services.AuthorityService

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
			// TODO: Should return all authorities, not the ones owned by someone
			owner, _ := ctx.Get("userEmail")
			ownerStr, _ := owner.(string)

			authorities, err := c.AuthorityService.GetAll(ownerStr)
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

			authority, err := c.AuthorityService.Get(id)
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
			// TODO
			ctx.JSON(http.StatusCreated, gin.H{})
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

			// TODO
			_ = id

			ctx.JSON(http.StatusOK, gin.H{})
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
}
