package resolver

type Configurator interface {
	SetDefaults()
	Validate() error
}
