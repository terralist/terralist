package server

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
	"github.com/valentindeaconu/terralist/database"
	"github.com/valentindeaconu/terralist/oauth"
	"github.com/valentindeaconu/terralist/oauth/provider"
	"github.com/valentindeaconu/terralist/routes"
)

// Server represents the Terralist server
type Server struct {
	Version       string
	Port          int
	Logger        *logrus.Logger
	Router        *gin.Engine
	OAuthProvider oauth.OAuthProvider
	Database      database.DB
}

// Config holds the server configuration that isn't configurable by the user
type Config struct {
	Version string
}

func NewServer(userConfig UserConfig, config Config) (*Server, error) {
	logger := logrus.New()

	if err := database.Connect(); err != nil {
		return nil, err
	}

	if userConfig.OAuthProvider == "github" {
		if err := provider.InitProvider("github"); err != nil {
			return nil, err
		}
	}

	router := gin.New()
	router.Use(ginlogrus.Logger(logger), gin.Recovery())

	return &Server{
		Version:       config.Version,
		Port:          userConfig.Port,
		Router:        router,
		Logger:        logger,
		OAuthProvider: provider.Handler(),
		Database:      database.Handler(),
	}, nil
}

// Start initializes the routes and starts serving
func (s *Server) Start() error {
	// Entry
	routes.InitEntryRoutes(s.Router)

	// Login
	// https://www.terraform.io/docs/internals/login-protocol.html
	routes.InitLoginRoutes(s.Router)

	// Modules
	// https://www.terraform.io/docs/internals/module-registry-protocol.html
	routes.InitModuleRoutes(s.Router)

	// Providers
	// https://www.terraform.io/docs/internals/provider-registry-protocol.html
	routes.InitProviderRoutes(s.Router)

	// Ensure server gracefully drains connections when stopped
	stop := make(chan os.Signal, 1)

	go func() {
		s.Logger.Info("Terralist started, listening on port %v", s.Port)

		err := s.Router.Run(fmt.Sprintf(":%d", s.Port))

		if err != nil {
			s.Logger.Error(err.Error())
		}
	}()
	<-stop

	s.Logger.Warn("Received intrerrupt signal, waiting for in-progress operations to complete")
	s.waitForDrain()

	return nil
}

// waitForDrain blocks the process until draining is complete
func (s *Server) waitForDrain() {
	drainComplete := make(chan bool, 1)

	go func() {
		// TODO: Make Drainer when necessary
		drainComplete <- true
	}()

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-drainComplete:
			s.Logger.Info("All in-progress operations completed, shutting down")
		case <-ticker.C:
			s.Logger.Info("Waiting for in-progress operations to complete")
		}
	}
}
