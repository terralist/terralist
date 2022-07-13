package database

const (
	SQLITE = iota
	POSTGRESQL
)

type Backend = int

// Creator creates the database
type Creator interface {
	New(config Configurator) (Engine, error)
}
