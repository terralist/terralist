package database

// Migrator migrates the entities to the physical database
type Migrator interface {
	Migrate(*DB) error
}
