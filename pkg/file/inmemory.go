package file

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

// InMemoryFile holds a file in-memory
type InMemoryFile struct {
	Name    string
	Content []byte
}

// Archive archives a slice of InMemoryFiles and returns the
// archive as an InMemoryFile
func Archive(name string, files []*InMemoryFile) (*InMemoryFile, error) {
	buffer := new(bytes.Buffer)

	writer := zip.NewWriter(buffer)
	defer writer.Close()

	for _, f := range files {
		w, err := writer.Create(f.Name)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrArchiveFailure, err)
		}

		if _, err := w.Write(f.Content); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSystemFailure, err)
		}
	}

	if !strings.HasSuffix(name, ".zip") {
		name = fmt.Sprintf("%s.zip", name)
	}

	return &InMemoryFile{
		Name:    name,
		Content: buffer.Bytes(),
	}, nil
}

// ContentType returns the http-compliant content-type
// of an InMemoryFile
func ContentType(f *InMemoryFile) string {
	return http.DetectContentType(f.Content)
}
