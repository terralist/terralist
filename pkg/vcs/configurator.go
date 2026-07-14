package vcs

type Configurator interface {
	SetDefaults()
	Validate() error
}
