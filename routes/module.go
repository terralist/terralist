package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/valentindeaconu/terralist/controllers"
	"github.com/valentindeaconu/terralist/settings"
)

func InitModuleRoutes(r *gin.Engine) {
	// https://www.terraform.io/docs/internals/module-registry-protocol.html#list-available-versions-for-a-specific-module
	r.GET(
		fmt.Sprintf(
			"%s/:namespace/:name/:provider/versions",
			settings.ServiceDiscovery.ModuleEndpoint,
		),
		controllers.ModuleGet,
	)

	// https://www.terraform.io/docs/internals/module-registry-protocol.html#download-source-code-for-a-specific-module-version
	r.GET(
		fmt.Sprintf(
			"%s/:namespace/:name/:provider/:version/download",
			settings.ServiceDiscovery.ModuleEndpoint,
		),
		controllers.ModuleGetVersion,
	)

	// Upload a new module
	r.POST(
		fmt.Sprintf(
			"%s/:namespace/:name/:provider/:version/upload",
			settings.ServiceDiscovery.ModuleEndpoint,
		),
		controllers.ModuleCreate,
	)

	// Upload a new version to a module
	r.PUT(
		fmt.Sprintf(
			"%s/:namespace/:name/:provider/:version/upload",
			settings.ServiceDiscovery.ModuleEndpoint,
		),
		controllers.ModuleAddVersion,
	)

	// Delete a module
	r.DELETE(
		fmt.Sprintf(
			"%s/:namespace/:name/:provider/remove",
			settings.ServiceDiscovery.ModuleEndpoint,
		),
		controllers.ModuleDelete,
	)

	// Delete a version from a module
	r.DELETE(
		fmt.Sprintf(
			"%s/:namespace/:name/:provider/:version/remove",
			settings.ServiceDiscovery.ModuleEndpoint,
		),
		controllers.ModuleDeleteVersion,
	)
}
