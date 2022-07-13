package database

type Configurator interface {
	SetDefaults()
	Validate() error
}
