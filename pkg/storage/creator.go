package storage

const (
	PROXY = iota
	LOCAL
	S3
	AZURE
	GCS
)

type Backend = int

// Creator creates the resolver.
type Creator interface {
	New(config Configurator) (Resolver, error)
}
