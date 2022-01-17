package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/valentindeaconu/terralist/controllers"
	"github.com/valentindeaconu/terralist/settings"
)

func InitProviderRoutes(r *gin.Engine) {
	// https://www.terraform.io/docs/internals/provider-registry-protocol.html#list-available-versions
	r.GET(
		fmt.Sprintf(
			"%s/:namespace/:name/versions",
			settings.ServiceDiscovery.ProviderEndpoint,
		),
		controllers.ProviderGet,
	)

	// https://www.terraform.io/docs/internals/provider-registry-protocol.html#find-a-provider-package
	r.GET(
		fmt.Sprintf(
			"%s/:namespace/:name/:version/download/:os/:arch",
			settings.ServiceDiscovery.ProviderEndpoint,
		),
		controllers.ProviderGetVersion,
	)

	// Upload a new provider version
	r.POST(
		fmt.Sprintf(
			"%s/:namespace/:name/:version/upload",
			settings.ServiceDiscovery.ProviderEndpoint,
		),
		controllers.ProviderUpload,
	)

	// Delete a provider
	r.DELETE(
		fmt.Sprintf(
			"%s/:namespace/:name/remove",
			settings.ServiceDiscovery.ProviderEndpoint,
		),
		controllers.ProviderDelete,
	)

	// Delete a provider version
	r.DELETE(
		fmt.Sprintf(
			"%s/:namespace/:name/:version/remove",
			settings.ServiceDiscovery.ProviderEndpoint,
		),
		controllers.ProviderVersionDelete,
	)
}
