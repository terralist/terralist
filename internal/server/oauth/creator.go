package oauth

import (
	models "github.com/valentindeaconu/terralist/internal/server/models/oauth"
	"github.com/valentindeaconu/terralist/internal/server/oauth/github"
)

const (
	Github = iota
)

// Engine handles the OAuth provider and operations
type Engine interface {
	GetAuthorizeUrl(state string) string
	GetUserDetails(code string, user *models.UserDetails) error
}

// ProviderCreator creates the OAuth provider
type ProviderCreator interface {
	NewProvider() (Engine, error)
}

// DefaultProviderCreator is the concrete implementation of ProviderCreator
type DefaultProviderCreator struct{}

func (p *DefaultProviderCreator) NewProvider() (Engine, error) {
	return github.NewProvider()
}
