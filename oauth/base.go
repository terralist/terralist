package oauth

type UserDetails struct {
	Name  string
	Email string
}

type OAuthProvider interface {
	GetAuthorizeUrl(state string) string
	GetUserDetails(code string, user *UserDetails) error
}
