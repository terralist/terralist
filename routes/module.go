package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/valentindeaconu/terralist/controllers"
	"github.com/valentindeaconu/terralist/middlewares"
	"github.com/valentindeaconu/terralist/settings"
)

func InitModuleRoutes(r *gin.Engine) {
	// https://www.terraform.io/docs/internals/module-registry-protocol.html#list-available-versions-for-a-specific-module
	r.GET(
		fmt.Sprintf(
			"%s/:namespace/:name/:provider/versions",
			settings.ServiceDiscovery.ModuleEndpoint,
		),
		middlewares.Authorize(),
		middlewares.AuditLogging(),
		controllers.ModuleGet,
	)

	// https://www.terraform.io/docs/internals/module-registry-protocol.html#download-source-code-for-a-specific-module-version
	r.GET(
		fmt.Sprintf(
			"%s/:namespace/:name/:provider/:version/download",
			settings.ServiceDiscovery.ModuleEndpoint,
		),
		middlewares.Authorize(),
		middlewares.AuditLogging(),
		controllers.ModuleGetVersion,
	)

	// Upload a new module version
	r.POST(
		fmt.Sprintf(
			"%s/:namespace/:name/:provider/:version/upload",
			settings.ServiceDiscovery.ModuleEndpoint,
		),
		middlewares.Authorize(),
		middlewares.AuditLogging(),
		controllers.ModuleUpload,
	)

	// Delete a module
	r.DELETE(
		fmt.Sprintf(
			"%s/:namespace/:name/:provider/remove",
			settings.ServiceDiscovery.ModuleEndpoint,
		),
		middlewares.Authorize(),
		middlewares.AuditLogging(),
		controllers.ModuleDelete,
	)

	// Delete a module version
	r.DELETE(
		fmt.Sprintf(
			"%s/:namespace/:name/:provider/:version/remove",
			settings.ServiceDiscovery.ModuleEndpoint,
		),
		middlewares.Authorize(),
		middlewares.AuditLogging(),
		controllers.ModuleVersionDelete,
	)
}
