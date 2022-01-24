package oauth

import (
	"fmt"

	models "github.com/valentindeaconu/terralist/internal/server/models/oauth"
	"github.com/valentindeaconu/terralist/internal/server/oauth/config"
	"github.com/valentindeaconu/terralist/internal/server/oauth/github"
)

// Engine handles the OAuth provider and operations
type Engine interface {
	GetAuthorizeUrl(state string) string
	GetUserDetails(code string, user *models.UserDetails) error
}

// ProviderCreator creates the OAuth provider
type ProviderCreator interface {
	NewProvider(provider string, config config.OAuthConfig) (Engine, error)
}

// DefaultProviderCreator is the concrete implementation of ProviderCreator
type DefaultProviderCreator struct{}

func (p *DefaultProviderCreator) NewProvider(provider string, config config.OAuthConfig) (Engine, error) {
	switch provider {
	case "github":
		return github.NewProvider(config)
	default:
		return nil, fmt.Errorf("there's no default oauth provider")
	}
}
