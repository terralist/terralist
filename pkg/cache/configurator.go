package cache

// Configurator describes the configuration of a cache backend.
type Configurator interface {
	SetDefaults()
	Validate() error
}
