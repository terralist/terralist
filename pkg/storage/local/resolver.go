package local

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"terralist/pkg/auth/jwt"
	"terralist/pkg/file"
	"terralist/pkg/storage"
)

// The local resolver will download files to a given path on the disk
// and will generate a public URL from which they can be downloaded.

var (
	ErrFileNotFound = errors.New("file not found")
)

// Resolver is the concrete implementation of storage.Resolver.
type Resolver struct {
	RegistryDir string
	LinkExpire  int
	URLFormat   string

	JWT jwt.JWT
}

type downloadTokenPayload struct {
	Key string `json:"key"`
}

func (r *Resolver) Store(in *storage.StoreInput) (string, error) {
	fileKey := path.Join(in.KeyPrefix, in.FileName)
	filePath := path.Join(r.RegistryDir, fileKey)

	_, err := os.Stat(filePath)

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	// If the file already exists, remove it first so it can be overwritten
	if err == nil {
		os.Remove(filePath)
	}

	if err := os.MkdirAll(path.Dir(filePath), 0700); err != nil {
		return "", fmt.Errorf("could not create parent directories: %w", err)
	}

	if _, err := in.Reader.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("could not rewind input reader: %w", err)
	}

	content, err := io.ReadAll(in.Reader)
	if err != nil {
		return "", fmt.Errorf("could not read input content: %w", err)
	}

	if err := os.WriteFile(filePath, content, 0700); err != nil {
		return "", fmt.Errorf("could not store file: %w", err)
	}

	return fileKey, nil
}

// ObjectExists checks if a given path exists on the disk.
func (r *Resolver) ObjectExists(key string) (string, error) {
	filePath := path.Join(r.RegistryDir, key)

	_, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrFileNotFound
		}

		return "", err
	}

	return filePath, nil
}

// GetObject reads a file from disk and returns it as a file.File abstraction.
func (r *Resolver) GetObject(key string) (file.File, error) {
	filePath, err := r.ObjectExists(key)
	if err != nil {
		return nil, err
	}

	return file.LoadFromDisk(path.Base(filePath), filePath)
}

func (r *Resolver) Find(key string) (string, error) {
	_, err := r.ObjectExists(key)
	if err != nil {
		return "", fmt.Errorf("could not generate URL for %v: %w", key, err)
	}

	token, err := r.JWT.Build(downloadTokenPayload{Key: key}, r.LinkExpire)
	if err != nil {
		return "", fmt.Errorf("could not generate a temporarily token: %w", err)
	}

	return fmt.Sprintf(r.URLFormat, key, token), nil
}

func (r *Resolver) Purge(key string) error {
	filePath, err := r.ObjectExists(key)
	if err != nil {
		if !errors.Is(err, ErrFileNotFound) {
			return err
		}

		return nil
	}

	return os.Remove(filePath)
}
