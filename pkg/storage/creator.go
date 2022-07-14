package storage

const (
	PROXY = iota
	LOCAL
	S3
)

type Backend = int

// Creator creates the database
type Creator interface {
	New(config Configurator) (Resolver, error)
}
