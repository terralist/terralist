package proxy

import (
	"terralist/pkg/storage"
)

// The proxy resolver will only forward the download URL received

// Resolver is the concrete implementation of storage.Resolver
type Resolver struct{}

func (r *Resolver) Store(in *storage.StoreInput) (string, error) {
	return string(in.Content), nil
}

func (r *Resolver) Find(key string) (string, error) {
	return key, nil
}

func (r *Resolver) Purge(_ string) error {
	return nil
}
