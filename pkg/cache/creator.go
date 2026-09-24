package cache

const (
	MEMORY = iota
)

type Backend = int

// Creator creates the cache.
type Creator interface {
	New(config Configurator) (Cache, error)
}
