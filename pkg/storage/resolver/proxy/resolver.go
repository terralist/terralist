package proxy

// The proxy resolver will only forward the download URL received

// Resolver is the concrete implementation of storage.Resolver
type Resolver struct{}

func (r *Resolver) Store(url string, archive bool) (string, error) {
	return url, nil
}

func (r *Resolver) Find(key string) (string, error) {
	return key, nil
}

func (r *Resolver) Purge(_ string) error {
	return nil
}
