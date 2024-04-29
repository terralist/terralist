package storage

const (
	PROXY = iota
	LOCAL
	S3
	AZURE
)

type Backend = int

// Creator creates the resolver
type Creator interface {
	New(config Configurator) (Resolver, error)
}
