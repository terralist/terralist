package api

import "github.com/gin-gonic/gin"

// RouterGroup is an abstraction over gin-gonic routers
type RouterGroup interface {
	// Prefix returns the base prefix for a router
	Prefix() string

	// Register subscribes a controller to the router
	Register(RestController)

	// RouterGroup returns the router
	RouterGroup() *gin.RouterGroup
}

type RouterGroupOptions struct {
	Prefix string
}

func NewRouterGroup(host *gin.Engine, opts *RouterGroupOptions) RouterGroup {
	r := host.Group(opts.Prefix)

	return &defaultRouterGroup{
		router: r,
		prefix: opts.Prefix,
	}
}

// defaultRouterGroup is a concrete implementation of RouterGroup
type defaultRouterGroup struct {
	// router is the router group configured with the given prefix
	router *gin.RouterGroup

	// prefix is the router prefix
	prefix string
}

func (r *defaultRouterGroup) Prefix() string {
	return r.prefix
}

func (r *defaultRouterGroup) Register(c RestController) {
	var groups []*gin.RouterGroup

	paths := c.Paths()

	for _, p := range paths {
		groups = append(groups, r.router.Group(p))
	}

	c.Subscribe(groups...)
}

func (r *defaultRouterGroup) RouterGroup() *gin.RouterGroup {
	return r.router
}
