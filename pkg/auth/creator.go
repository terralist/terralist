package auth

const (
	GITHUB = iota
	BITBUCKET
	GITLAB
	OIDC
	SAML
)

type Backend = int

// Creator creates the database.
type Creator interface {
	New(config Configurator) (Provider, error)
}
