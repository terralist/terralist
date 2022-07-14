package local

import "fmt"

// The local resolver will download files to a given path on the disk
// and will generate a public URL from which they can be downloaded

// Resolver is the concrete implementation of storage.Resolver
type Resolver struct {
	DataStorePath string
}

func (r *Resolver) Store(url string) (string, error) {
	return "", fmt.Errorf("not yet implemented")
}
