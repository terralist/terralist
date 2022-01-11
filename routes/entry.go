package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/valentindeaconu/terralist/controllers"
)

func InitEntryRoutes(r *gin.Engine) {
	// Health Check API
	r.GET("/health", controllers.HealthCheck)
	r.GET("/.well-known/terraform.json", controllers.ServiceDiscovery)

}
