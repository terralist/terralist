package storage

type Configurator interface {
	SetDefaults()
	Validate() error
}
