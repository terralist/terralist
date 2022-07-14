package s3

import "fmt"

// The S3 resolver will download files from the given URL then
// uploads them to an S3 bucket, generating a public download URL

// Resolver is the concrete implementation of storage.Resolver
type Resolver struct{}

func (r *Resolver) Store(url string) (string, error) {
	return "", fmt.Errorf("not yet implemented")
}

func (r *Resolver) Purge(_ string) error {
	return nil
}
