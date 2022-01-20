package provider

import (
	"fmt"

	"github.com/valentindeaconu/terralist/oauth"
	"github.com/valentindeaconu/terralist/oauth/providers"
)

var (
	provider oauth.OAuthProvider
)

var (
	clientID     string = "to_be_added"
	clientSecret string = "to_be_added"
)

func InitProvider(p string) error {
	switch p {
	case "github":
		provider = providers.NewGitHubOAuthProvider(clientID, clientSecret, "")
	default:
		return fmt.Errorf("oauth provider not supported")
	}

	return nil
}

func Handler() oauth.OAuthProvider {
	return provider
}
