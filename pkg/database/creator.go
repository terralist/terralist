package database

const (
	SQLITE = iota
	MYSQL
)

type Backend = int

// Creator creates the database
type Creator interface {
	New(config Configurator, migrator Migrator) (Engine, error)
}
