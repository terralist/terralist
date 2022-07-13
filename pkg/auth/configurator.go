package auth

type Configurator interface {
	SetDefaults()
	Validate() error
}
