package gcs

import (
	"context"
	"fmt"
	"io"
	"time"

	"terralist/pkg/storage"

	gcs "cloud.google.com/go/storage"
)

// The GCS resolver will download files from the given URL then
// uploads them to an GCS bucket, generating a public download URL.

type Resolver struct {
	BucketName   string
	BucketPrefix string
	LinkExpire   int

	Client *gcs.Client
}

// Resolver is the concrete implementation of storage.Resolver.
func (r *Resolver) Store(in *storage.StoreInput) (string, error) {
	key := fmt.Sprintf("%s/%s", in.KeyPrefix, in.FileName)

	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Minute*2)
	defer cancel()

	wc := r.Client.Bucket(r.BucketName).Object(key).NewWriter(ctx)

	if _, err := io.Copy(wc, in.Reader); err != nil {
		return "", fmt.Errorf("could not upload archive: %v", err)
	}

	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("could not close the archive: %v", err)
	}

	return key, nil
}

func (r *Resolver) Find(key string) (string, error) {
	opts := &gcs.SignedURLOptions{
		Scheme:  gcs.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(time.Duration(r.LinkExpire) * time.Minute),
	}

	url, err := r.Client.Bucket(r.BucketName).SignedURL(key, opts)
	if err != nil {
		return "", fmt.Errorf("could not generate URL for %v: %v", key, err)
	}

	return url, nil
}

func (r *Resolver) Purge(key string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute*2)
	defer cancel()

	if err := r.Client.Bucket(r.BucketName).Object(key).Delete(ctx); err != nil {
		return fmt.Errorf("could not purge object: %v", err)
	}

	return nil
}
