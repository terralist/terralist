package api

import "github.com/gin-gonic/gin"

const (
	RouterMinPriority = 255
	RouterMaxPriority = 0
)

// Router is an abstraction over gin-gonic routers
type Router interface {
	// Prefix returns the base prefix for a router
	Prefix() string

	// Priority returns the priority of the router
	// the lower the number is, the higher priority
	Priority() uint8

	// Register subscribes a controller to the router
	Register(RestController)

	// Router returns the router
	Router() *gin.Engine
}

type RouterOptions struct {
	Prefix   string
	Priority uint8
}

func NewRouter(opts RouterOptions) Router {
	r := gin.New()
	g := r.Group(opts.Prefix)

	return &defaultRouter{
		router:   r,
		group:    g,
		prefix:   opts.Prefix,
		priority: opts.Priority,
	}
}

// defaultRouter is a concrete implementation of Router
type defaultRouter struct {
	// router is the gin-gonic router
	router *gin.Engine

	// group is the router group configured with the given prefix
	group *gin.RouterGroup

	// prefix is the router prefix
	prefix string

	// priority is the router priority
	priority uint8
}

func (r *defaultRouter) Prefix() string {
	return r.prefix
}

func (r *defaultRouter) Priority() uint8 {
	return r.priority
}

func (r *defaultRouter) Register(c RestController) {
	var groups []*gin.RouterGroup

	paths := c.Paths()

	for _, p := range paths {
		groups = append(groups, r.group.Group(p))
	}

	c.Subscribe(groups...)
}

func (r *defaultRouter) Router() *gin.Engine {
	return r.router
}
