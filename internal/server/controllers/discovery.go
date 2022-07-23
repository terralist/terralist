package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	terraformPorts = []int{10000, 10010}
)

type ServiceDiscoveryController struct {
	AuthorizationEndpoint string
	TokenEndpoint         string
	ModuleEndpoint        string
	ProviderEndpoint      string
}

func (c *ServiceDiscoveryController) TerraformApiBase() string {
	return "/.well-known"
}

func (c *ServiceDiscoveryController) ApiBase() string {
	return "/.not-used"
}

func (c *ServiceDiscoveryController) Subscribe(tfApi *gin.RouterGroup, _ *gin.RouterGroup) {
	// tfApi should be compliant with the Terraform Registry Protocol for
	// service discovery
	// Docs: https://www.terraform.io/internals/remote-service-discovery

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
