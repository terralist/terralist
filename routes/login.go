package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/valentindeaconu/terralist/controllers"
	"github.com/valentindeaconu/terralist/settings"
)

func InitLoginRoutes(r *gin.Engine) {
	r.GET(settings.ServiceDiscovery.Login.AuthorizationEndpoint, controllers.Authorize)
	r.GET(settings.ServiceDiscovery.Login.TokenEndpoint, controllers.TokenValidate)
}
