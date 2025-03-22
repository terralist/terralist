package auth

// Provider handles the OAuth provider and operations.
type Provider interface {
	Name() string
	GetAuthorizeUrl(state string) string
	GetUserDetails(code string, user *User) error
}
