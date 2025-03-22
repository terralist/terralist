package database

const (
	SQLITE = iota
	POSTGRESQL
	MYSQL
)

type Backend = int

// Creator creates the database.
type Creator interface {
	New(config Configurator) (Engine, error)
}
