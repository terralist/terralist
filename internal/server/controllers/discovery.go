package controllers

import (
	"net/http"
	"terralist/pkg/api"

	"github.com/gin-gonic/gin"
)

var (
	terraformPorts = []int{10000, 10010}
)

// ServiceDiscoveryController registers the endpoints described by the
// service discovery protocol
type ServiceDiscoveryController interface {
	api.RestController
}

// DefaultServiceDiscoveryController is a concrete implementation of
// ServiceDiscoveryController
type DefaultServiceDiscoveryController struct {
	AuthorizationEndpoint string
	TokenEndpoint         string
	ModuleEndpoint        string
	ProviderEndpoint      string
}

func (c *DefaultServiceDiscoveryController) Paths() []string {
	return []string{
		"/.well-known", // Terraform API route
	}
}

func (c *DefaultServiceDiscoveryController) Subscribe(apis ...*gin.RouterGroup) {
	// tfApi should be compliant with the Terraform Registry Protocol for
	// service discovery
	// Docs: https://www.terraform.io/internals/remote-service-discovery
	tfApi := apis[0]

	// Terraform Service Discovery API
	tfApi.GET(
		"/terraform.json",
		func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"login.v1": gin.H{
					"client":      "terraform-cli",
					"grant_types": []string{"authz_code"},
					"authz":       c.AuthorizationEndpoint,
					"token":       c.TokenEndpoint,
					"ports":       terraformPorts,
				},
				"modules.v1":   c.ModuleEndpoint,
				"providers.v1": c.ProviderEndpoint,
			})
		},
	)
}
