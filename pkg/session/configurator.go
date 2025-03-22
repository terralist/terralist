package session

// Configurator is a interface that describes how
// a session store configurator should look like.
type Configurator interface {
	SetDefaults()
	Validate() error
}
