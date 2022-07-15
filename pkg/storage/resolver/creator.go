package resolver

const (
	PROXY = iota
	LOCAL
	S3
)

type Backend = int

// Creator creates the resolver
type Creator interface {
	New(config Configurator) (Resolver, error)
}
