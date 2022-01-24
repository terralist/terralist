package server

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
	"github.com/valentindeaconu/terralist/internal/server/controllers"
	"github.com/valentindeaconu/terralist/internal/server/database"
	"github.com/valentindeaconu/terralist/internal/server/handlers"
	"github.com/valentindeaconu/terralist/internal/server/mappers"
	"github.com/valentindeaconu/terralist/internal/server/oauth"
	"github.com/valentindeaconu/terralist/internal/server/services"
	"github.com/valentindeaconu/terralist/internal/server/utils"
)

const (
	OAuthAuthorizationEndpoint = "/oauth/authorization"
	OAuthTokenEndpoint         = "/oauth/token"
	OAuthRedirectEndpoint      = "/oauth/redirect"
	ModuleEndpoint             = "/v1/modules"
	ProviderEndpoint           = "/v1/providers"
)

var (
	TerraformPorts = []int{10000, 10010}
)

// Server represents the Terralist server
type Server struct {
	Version            string
	Port               int
	Logger             *logrus.Logger
	Router             *gin.Engine
	OAuthProvider      oauth.Engine
	Database           database.Engine
	EntryController    *controllers.EntryController
	LoginController    *controllers.LoginController
	ModuleController   *controllers.ModuleController
	ProviderController *controllers.ProviderController
	JWT                *utils.JWT
	Keychain           *utils.Keychain
}

// Config holds the server configuration that isn't configurable by the user
type Config struct {
	Version string
}

func NewServer(userConfig UserConfig, config Config) (*Server, error) {
	logger := logrus.New()

	databaseCreator := database.DefaultDatabaseCreator{}
	db, err := databaseCreator.NewDatabase(userConfig.DatabaseBackend, userConfig.ToDatabaseConfig())
	if err != nil {
		return nil, err
	}

	providerCreator := oauth.DefaultProviderCreator{}

	provider, err := providerCreator.NewProvider(userConfig.OAuthProvider, userConfig.ToOAuthProviderConfig())
	if err != nil {
		return nil, err
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(ginlogrus.Logger(logger), gin.Recovery())

	entryController := &controllers.EntryController{
		AuthorizationEndpoint: OAuthAuthorizationEndpoint,
		TokenEndpoint:         OAuthTokenEndpoint,
		ModuleEndpoint:        ModuleEndpoint + "/",
		ProviderEndpoint:      ProviderEndpoint + "/",
		TerraformPorts:        TerraformPorts,
	}

	keychain := utils.NewKeychain(userConfig.TokenSigningSecret)

	jwt := &utils.JWT{
		Keychain: keychain,
	}

	oauthMapper := &mappers.OAuthMapper{
		Keychain: keychain,
	}

	loginController := &controllers.LoginController{
		Provider:    provider,
		OAuthMapper: oauthMapper,
		Keychain:    keychain,
		JWT:         jwt,
	}

	moduleService := &services.ModuleService{
		Database: db,
	}

	moduleMapper := &mappers.ModuleMapper{}

	moduleController := &controllers.ModuleController{
		ModuleService: moduleService,
		ModuleMapper:  moduleMapper,
	}

	providerService := &services.ProviderService{
		Database: db,
	}

	providerMapper := &mappers.ProviderMapper{}

	providerController := &controllers.ProviderController{
		ProviderService: providerService,
		ProviderMapper:  providerMapper,
	}

	return &Server{
		Version:            config.Version,
		Port:               userConfig.Port,
		Router:             router,
		Logger:             logger,
		OAuthProvider:      provider,
		Database:           db,
		EntryController:    entryController,
		LoginController:    loginController,
		ModuleController:   moduleController,
		ProviderController: providerController,
		Keychain:           keychain,
		JWT:                jwt,
	}, nil
}

// Start initializes the routes and starts serving
func (s *Server) Start() error {
	// Entry routes (no security checks)
	// Health Check API
	s.Router.GET("/health", s.EntryController.HealthCheck())

	// Terraform Service Discovery API
	s.Router.GET("/.well-known/terraform.json", s.EntryController.ServiceDiscovery())

	// Login routes (no security checks)
	// https://www.terraform.io/docs/internals/login-protocol.html
	s.Router.GET(OAuthAuthorizationEndpoint, s.LoginController.Authorize())
	s.Router.GET(OAuthRedirectEndpoint, s.LoginController.Redirect())
	s.Router.POST(OAuthTokenEndpoint, s.LoginController.TokenValidate())

	// Modules routes (with security checks)
	// https://www.terraform.io/docs/internals/module-registry-protocol.html#list-available-versions-for-a-specific-module
	s.Router.GET(
		ModuleEndpoint+"/:namespace/:name/:provider/versions",
		handlers.Authorize(s.JWT),
		handlers.AuditLogging(s.Logger),
		s.ModuleController.Get(),
	)

	// https://www.terraform.io/docs/internals/module-registry-protocol.html#download-source-code-for-a-specific-module-version
	s.Router.GET(
		ModuleEndpoint+"/:namespace/:name/:provider/:version/download",
		handlers.Authorize(s.JWT),
		handlers.AuditLogging(s.Logger),
		s.ModuleController.GetVersion(),
	)

	// Upload a new module version
	s.Router.POST(
		ModuleEndpoint+"/:namespace/:name/:provider/:version/upload",
		handlers.Authorize(s.JWT),
		handlers.AuditLogging(s.Logger),
		s.ModuleController.Upload(),
	)

	// Delete a module
	s.Router.DELETE(
		ModuleEndpoint+"/:namespace/:name/:provider/remove",
		handlers.Authorize(s.JWT),
		handlers.AuditLogging(s.Logger),
		s.ModuleController.Delete(),
	)

	// Delete a module version
	s.Router.DELETE(
		ModuleEndpoint+"/:namespace/:name/:provider/:version/remove",
		handlers.Authorize(s.JWT),
		handlers.AuditLogging(s.Logger),
		s.ModuleController.DeleteVersion(),
	)

	// Providers
	// https://www.terraform.io/docs/internals/provider-registry-protocol.html#list-available-versions
	s.Router.GET(
		ProviderEndpoint+"/:namespace/:name/versions",
		handlers.Authorize(s.JWT),
		handlers.AuditLogging(s.Logger),
		s.ProviderController.Get(),
	)

	// https://www.terraform.io/docs/internals/provider-registry-protocol.html#find-a-provider-package
	s.Router.GET(
		ProviderEndpoint+"/:namespace/:name/:version/download/:os/:arch",
		handlers.Authorize(s.JWT),
		handlers.AuditLogging(s.Logger),
		s.ProviderController.GetVersion(),
	)

	// Upload a new provider version
	s.Router.POST(
		ProviderEndpoint+"/:namespace/:name/:version/upload",
		handlers.Authorize(s.JWT),
		handlers.AuditLogging(s.Logger),
		s.ProviderController.Upload(),
	)

	// Delete a provider
	s.Router.DELETE(
		ProviderEndpoint+"/:namespace/:name/remove",
		handlers.Authorize(s.JWT),
		handlers.AuditLogging(s.Logger),
		s.ProviderController.Delete(),
	)

	// Delete a provider version
	s.Router.DELETE(
		ProviderEndpoint+"/:namespace/:name/:version/remove",
		handlers.Authorize(s.JWT),
		handlers.AuditLogging(s.Logger),
		s.ProviderController.DeleteVersion(),
	)

	// Ensure server gracefully drains connections when stopped
	stop := make(chan os.Signal, 1)

	go func() {
		s.Logger.Infof("Terralist started, listening on port %v", s.Port)

		err := s.Router.Run(fmt.Sprintf(":%d", s.Port))

		if err != nil {
			s.Logger.Error(err.Error())
		}
	}()
	<-stop

	s.Logger.Warn("Received intrerupt signal, waiting for in-progress operations to complete")
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
