package api

import "github.com/gin-gonic/gin"

// RestController is an interface implemented by a REST controller.
type RestController interface {
	// Paths returns a list of relative paths to be passed to the router
	// groups.
	Paths() []string

	// Subscribe is the method called by the router, passing the router
	// groups to let the controller register its methods.
	// The length of the groups received is equal with the length of the
	// relative paths returned by Paths.
	Subscribe(apis ...*gin.RouterGroup)
}
