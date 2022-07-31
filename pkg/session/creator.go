package session

const (
	COOKIE = iota
)

type Backend = int

// Creator creates the resolver
type Creator interface {
	New(config Configurator) (Store, error)
}
