package controllers

import (
	"net/http"
	"sync/atomic"

	"terralist/pkg/api"

	"github.com/gin-gonic/gin"
)

// ProbeController registers the routes that handles the probes.
type ProbeController interface {
	api.RestController
}

// DefaultProbeController is a concrete implementation of ProbeController.
type DefaultProbeController struct {
	Ready *atomic.Bool
}

func (c *DefaultProbeController) Paths() []string {
	return []string{""} // bind to router default
}

func (c *DefaultProbeController) Subscribe(apis ...*gin.RouterGroup) {
	api := apis[0]

	api.GET("/healthz", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	api.GET("/readyz", func(ctx *gin.Context) {
		if c.Ready == nil || !c.Ready.Load() {
			ctx.AbortWithStatus(http.StatusServiceUnavailable)
		}

		ctx.Status(http.StatusOK)
	})
}
