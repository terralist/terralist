package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type EntryController struct {
	AuthorizationEndpoint string
	TokenEndpoint         string
	ModuleEndpoint        string
	ProviderEndpoint      string
	TerraformPorts        []int
}

func (e *EntryController) ServiceDiscovery() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"login.v1": gin.H{
				"client":      "terraform-cli",
				"grant_types": []string{"authz_code"},
				"authz":       e.AuthorizationEndpoint,
				"token":       e.TokenEndpoint,
				"ports":       e.TerraformPorts,
			},
			"modules.v1":   e.ModuleEndpoint,
			"providers.v1": e.ProviderEndpoint,
		})
	}
}
