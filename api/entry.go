package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/valentindeaconu/terralist/settings"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}

func ServiceDiscovery(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"login.v1":     fmt.Sprintf("%s/", settings.ServiceDiscovery.LoginEndpoint),
		"modules.v1":   fmt.Sprintf("%s/", settings.ServiceDiscovery.ModuleEndpoint),
		"providers.v1": fmt.Sprintf("%s/", settings.ServiceDiscovery.ProviderEndpoint),
	})
}
