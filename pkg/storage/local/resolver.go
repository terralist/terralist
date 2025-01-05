package local

import (
	"bytes"
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
// and will generate a public URL from which they can be downloaded

var (
	ErrFileNotFound = errors.New("file not found")
)

// Resolver is the concrete implementation of storage.Resolver
type Resolver struct {
	RegistryDir string
	LinkExpire  int
	URLFormat   string

	JWT jwt.JWT
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

	if err := os.WriteFile(filePath, in.Content, 0700); err != nil {
		return "", fmt.Errorf("could not store file: %w", err)
	}

	return fileKey, nil
}

// ObjectExists checks if a given path exists on the disk
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

// GetObject reads a file from the disk and returns it as an InMemoryFile abstraction
func (r *Resolver) GetObject(key string) (*file.InMemoryFile, error) {
	filePath, err := r.ObjectExists(key)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("cannot stat file: %w", err)
	}

	buffer := new(bytes.Buffer)
	if _, err := io.Copy(buffer, f); err != nil {
		return nil, err
	}

	return &file.InMemoryFile{
		Name:     fi.Name(),
		FileInfo: fi,
		Content:  buffer.Bytes(),
	}, nil
}

func (r *Resolver) Find(key string) (string, error) {
	_, err := r.ObjectExists(key)
	if err != nil {
		return "", fmt.Errorf("could not generate URL for %v: %w", key, err)
	}

	token, err := r.JWT.Build(nil, r.LinkExpire)
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
