package controllers

import (
	"io/fs"
	"net/http"

	"terralist/pkg/api"
	"terralist/web"

	"github.com/gin-gonic/gin"
)

// WebController registers the endpoints for web interface
type WebController interface {
	api.RestController
}

// DefaultWebController is a concrete implementation of
// WebController
type DefaultWebController struct{}

func (c *DefaultWebController) Paths() []string {
	return []string{
		"", // bind to the router default '/'
	}
}

func (c *DefaultWebController) Subscribe(apis ...*gin.RouterGroup) {
	// Router group
	r := apis[0]

	distFS, _ := fs.Sub(web.FS, "dist")
	r.StaticFS("/", http.FS(distFS))
}
