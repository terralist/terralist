package file

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/rs/zerolog/log"
)

// File abstracts a type of file.
type File interface {
	io.Reader
	io.Seeker
	io.Closer

	// Name returns the name of the file.
	Name() string

	// Metadata returns info headers for the files.
	Metadata() fs.FileInfo
}

// InMemoryFile holds a file in-memory.
type InMemoryFile struct {
	name     string
	fileInfo fs.FileInfo
	content  []byte
}

func (f *InMemoryFile) Name() string {
	return f.name
}

func (f *InMemoryFile) Metadata() fs.FileInfo {
	return f.fileInfo
}

func (f *InMemoryFile) Read(p []byte) (n int, err error) {
	return bytes.NewReader(f.content).Read(p)
}

func (f *InMemoryFile) Seek(offset int64, whence int) (int64, error) {
	return bytes.NewReader(f.content).Seek(offset, whence)
}

func (f *InMemoryFile) Close() error {
	return nil
}

// StreamingFile holds a streaming file.
type StreamingFile struct {
	name     string
	fileInfo fs.FileInfo
	reader   io.ReadSeekCloser
}

func (f *StreamingFile) Name() string {
	return f.name
}

func (f *StreamingFile) Metadata() fs.FileInfo {
	return f.fileInfo
}

func (f *StreamingFile) Read(p []byte) (n int, err error) {
	return f.reader.Read(p)
}

func (f *StreamingFile) Seek(offset int64, whence int) (int64, error) {
	return f.reader.Seek(offset, whence)
}

func (f *StreamingFile) Close() error {
	return f.reader.Close()
}

// OnDiskFile is a File wrapper which is also stored on disk.
type OnDiskFile struct {
	name string
	path string
}

func (f *OnDiskFile) Name() string {
	return f.name
}

func (f *OnDiskFile) Metadata() fs.FileInfo {
	fi, err := os.Stat(f.path)
	if err != nil {
		return nil
	}

	return fi
}

func (f *OnDiskFile) Read(p []byte) (int, error) {
	file, err := os.Open(f.path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return file.Read(p)
}

func (f *OnDiskFile) Seek(offset int64, whence int) (int64, error) {
	return 0, fmt.Errorf("seek operation not supported")
}

func (f *OnDiskFile) Close() error {
	return nil
}

func (f *OnDiskFile) ToStreamingFile() (*StreamingFile, error) {
	file, err := os.Open(f.path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}

	return &StreamingFile{
		name:     f.name,
		fileInfo: f.Metadata(),
		reader:   file,
	}, nil
}

// Path returns the path to the file on the disk.
func (f *OnDiskFile) Path() string {
	return f.path
}

// Remove removes the file from the disk.
func (f *OnDiskFile) Remove() error {
	return os.Remove(f.path)
}

// ContentType returns the http-compliant content-type of a File.
func ContentType(f File) string {
	data, err := bufio.NewReader(f).Peek(512)
	if err != nil {
		return "application/octet-stream"
	}

	// Rewind the reader
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		log.Error().
			Err(err).
			Str("name", f.Name()).
			Msg("could not rewind the file")
	}

	return http.DetectContentType(data)
}

// SaveToTemp writes a file to the disk, in a temp file.
func SaveToTemp(f File) (*OnDiskFile, error) {
	file, err := os.CreateTemp("", "terralist.tmp.*")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err := io.Copy(file, f); err != nil {
		return nil, err
	}

	return &OnDiskFile{
		name: f.Name(),
		path: file.Name(),
	}, nil
}

// SaveToDisk writes a file to the disk, in a given destination.
func SaveToDisk(f File, dst string) (*OnDiskFile, error) {
	filePath := path.Join(dst, f.Name())
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err := io.Copy(file, f); err != nil {
		return nil, err
	}

	return &OnDiskFile{
		name: f.Name(),
		path: filePath,
	}, nil
}

// LoadFromDisk loads a file from the disk.
func LoadFromDisk(name string, src string) (File, error) {
	file := &OnDiskFile{name: name, path: src}
	return file.ToStreamingFile()
}

// NewFromMultipartFileHeader returns an empty file using metadata from a multipart.FileHeader.
func NewFromMultipartFileHeader(h *multipart.FileHeader) File {
	return &InMemoryFile{
		name: h.Filename,
		fileInfo: &BufferFileInfo{
			name:    h.Filename,
			size:    h.Size,
			mode:    0644,
			modTime: time.Now(),
		},
		content: []byte{},
	}
}

// NewEmptyFile returns an empty file.
func NewEmptyFile(name string) File {
	return &InMemoryFile{
		name: name,
		fileInfo: &BufferFileInfo{
			name:    name,
			size:    0,
			mode:    0644,
			modTime: time.Now(),
		},
		content: []byte{},
	}
}
