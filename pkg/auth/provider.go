package auth

// Provider handles the OAuth provider and operations
type Provider interface {
	GetAuthorizeUrl(state string) string
	GetUserDetails(code string, user *User) error
}
