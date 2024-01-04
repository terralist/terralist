package file

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

// InMemoryFile holds a file in-memory
type InMemoryFile struct {
	Name     string
	FileInfo fs.FileInfo
	Content  []byte
}

// Archive archives a slice of InMemoryFiles and returns the
// archive as an InMemoryFile
func Archive(name string, files []*InMemoryFile) (*InMemoryFile, error) {
	buffer := new(bytes.Buffer)

	writer := zip.NewWriter(buffer)

	for _, f := range files {
		// fetch file info and set file path to relative one to preserve directories
		hdr, err := zip.FileInfoHeader(f.FileInfo)
		hdr.Name = f.Name
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrArchiveFailure, err)
		}

		w, err := writer.CreateHeader(hdr)
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

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrArchiveFailure, err)
	}

	return &InMemoryFile{
		Name:     name,
		FileInfo: nil,
		Content:  buffer.Bytes(),
	}, nil
}

// ContentType returns the http-compliant content-type
// of an InMemoryFile
func ContentType(f *InMemoryFile) string {
	return http.DetectContentType(f.Content)
}
