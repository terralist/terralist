package proxy

import (
	"fmt"

	"terralist/pkg/storage"
)

// The proxy resolver will only forward the download URL received

// Resolver is the concrete implementation of storage.Resolver
type Resolver struct{}

func (r *Resolver) Store(in *storage.StoreInput) (string, error) {
	if in.Content != nil {
		return "", fmt.Errorf("proxy does not support storing content")
	}

	return in.URL, nil
}

func (r *Resolver) Find(key string) (string, error) {
	return key, nil
}

func (r *Resolver) Purge(_ string) error {
	return nil
}
