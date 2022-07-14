package proxy

// The proxy resolver will only forward the download URL received

// Resolver is the concrete implementation of storage.Resolver
type Resolver struct{}

func (r *Resolver) Store(url string) (string, error) {
	return url, nil
}
