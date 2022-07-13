package auth

const (
	GITHUB = iota
)

type Backend = int

// Creator creates the database
type Creator interface {
	New(config Configurator) (Provider, error)
}
