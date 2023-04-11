package controllers

import (
	"net/http"

	"terralist/internal/server/handlers"
	"terralist/internal/server/models/authority"
	"terralist/internal/server/services"
	"terralist/pkg/api"
	"terralist/pkg/auth/jwt"

	"github.com/gin-gonic/gin"
	"github.com/ssoroka/slice"
)

const (
	authorityApiBase = "/api/authority"
)

// AuthorityController registers the endpoints to control authorities
type AuthorityController interface {
	api.RestController
}

// DefaultAuthorityController is a concrete implementation of
// AuthorityController
type DefaultAuthorityController struct {
	AuthorityService services.AuthorityService
	ApiKeyService    services.ApiKeyService

	JWT jwt.JWT
}

func (c *DefaultAuthorityController) Paths() []string {
	return []string{authorityApiBase}
}

func (c *DefaultAuthorityController) Subscribe(apis ...*gin.RouterGroup) {
	api := apis[0]
	api.Use(handlers.Authorize(c.JWT, c.ApiKeyService))

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
				func(a *authority.Authority,
				) authority.AuthorityDTO {
					return a.ToDTO()
				})

			ctx.JSON(http.StatusOK, dtos)
		},
	)
}
