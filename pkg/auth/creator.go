package auth

const (
	GITHUB = iota
	BITBUCKET
)

type Backend = int

// Creator creates the database
type Creator interface {
	New(config Configurator) (Provider, error)
}
