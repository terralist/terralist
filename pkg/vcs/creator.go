package vcs

const (
	GITHUB = iota
)

type Backend = int

// Creator creates the resolver.
type Creator interface {
	New(config Configurator) (Provider, error)
}
