package file

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"
)

// BufferFileInfo implements fs.FileInfo for a bytes.Buffer
type BufferFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
}

// Implement the FileInfo interface
func (fi *BufferFileInfo) Name() string {
	return fi.name
}

func (fi *BufferFileInfo) Size() int64 {
	return fi.size
}

func (fi *BufferFileInfo) Mode() fs.FileMode {
	return fi.mode
}

func (fi *BufferFileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi *BufferFileInfo) IsDir() bool {
	return false
}

func (fi *BufferFileInfo) Sys() interface{} {
	return nil
}

// NewBufferFileInfo creates a FileInfo for a given bytes.Buffer
func NewBufferFileInfo(buffer *bytes.Buffer, name string) fs.FileInfo {
	return &BufferFileInfo{
		name:    name,
		size:    int64(buffer.Len()),
		mode:    0644,
		modTime: time.Now(),
	}
}

// BufferReadSeekCloser wraps a bytes.Buffer and implements io.ReadSeekCloser
type BufferReadSeekCloser struct {
	buf *bytes.Reader
}

func (b *BufferReadSeekCloser) Read(p []byte) (int, error) {
	return b.buf.Read(p)
}

func (b *BufferReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	return b.buf.Seek(offset, whence)
}

func (b *BufferReadSeekCloser) Close() error {
	// No real resources to close for a bytes.Buffer, so return nil
	return nil
}

// NewBufferReadSeekCloser is a constructor that takes a bytes.Buffer
// and returns a ReadSeekCloser implementation
func NewBufferReadSeekCloser(buffer *bytes.Buffer) io.ReadSeekCloser {
	return &BufferReadSeekCloser{
		buf: bytes.NewReader(buffer.Bytes()),
	}
}

// Archive archives a slice of Files and returns the
// archive as a StreamingFile
func Archive(name string, files []File) (File, error) {
	buffer := new(bytes.Buffer)

	writer := zip.NewWriter(buffer)

	for _, f := range files {
		// fetch file info and set file path to relative one to preserve directories
		hdr, err := zip.FileInfoHeader(f.Metadata())
		hdr.Name = f.Name()

		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrArchiveFailure, err)
		}

		w, err := writer.CreateHeader(hdr)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrArchiveFailure, err)
		}

		if _, err := io.Copy(w, f); err != nil {
			return nil, err
		}
	}

	if !strings.HasSuffix(name, ".zip") {
		name = fmt.Sprintf("%s.zip", name)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrArchiveFailure, err)
	}

	return &StreamingFile{
		name:     name,
		fileInfo: NewBufferFileInfo(buffer, name),
		reader:   NewBufferReadSeekCloser(buffer),
	}, nil
}
