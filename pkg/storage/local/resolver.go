package local

import (
	"fmt"

	"terralist/pkg/storage"
)

// The local resolver will download files to a given path on the disk
// and will generate a public URL from which they can be downloaded

// Resolver is the concrete implementation of storage.Resolver
type Resolver struct {
	RegistryDir string
}

func (r *Resolver) Store(_ *storage.StoreInput) (string, error) {
	return "", fmt.Errorf("not yet implemented")
}

func (r *Resolver) Find(_ string) (string, error) {
	return "", fmt.Errorf("not yet implemented")
}

func (r *Resolver) Purge(_ string) error {
	return nil
}
