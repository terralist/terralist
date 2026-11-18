package auth

import "errors"

// ErrUpstreamUnavailable reports that the provider could not be reached, as
// opposed to the provider refusing a grant.
var ErrUpstreamUnavailable = errors.New("upstream provider is unavailable")

// Provider handles the OAuth provider and operations.
type Provider interface {
	Name() string
	GetAuthorizeUrl(state string) string
	GetUserDetails(code string, user *User) error
}

// Grant is the upstream credential kept from a login, which lets the provider
// check the user again later.
type Grant struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// Refresher is implemented by providers that can check a user again after
// the login has completed.
type Refresher interface {
	// ExchangeCode works like GetUserDetails, and also returns the grant to
	// keep for refreshing. It returns a nil grant when the provider gave none
	// for this login.
	ExchangeCode(code string, user *User) (*Grant, error)

	// Refresh fetches the user details again with a kept grant, and returns
	// the grant to keep for the next refresh. Providers that rotate their
	// refresh tokens return the new one.
	Refresh(grant Grant, user *User) (*Grant, error)
}
