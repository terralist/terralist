package file

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
)

var (
	ErrDuplicateFile = errors.New("could not index duplicate files")
)

// FS implements io.FS and represents an FS based on
// the File interface
type FS struct {
	files map[string]File
}

// NewFS takes a list of files and creates a FS
func NewFS(files []File) (*FS, error) {
	filesMap := map[string]File{}

	for _, file := range files {
		name := file.Name()

		if _, ok := filesMap[name]; ok {
			return nil, fmt.Errorf("%w: %v", ErrDuplicateFile, name)
		}

		filesMap[file.Name()] = file
	}

	return &FS{
		files: filesMap,
	}, nil
}

// MustNewFS takes a list of files and creates a FS
// It panics if it cannot create the FS
func MustNewFS(files []File) *FS {
	fs, err := NewFS(files)
	if err != nil {
		panic(err)
	}

	return fs
}

func (f *FS) Open(name string) (fs.File, error) {
	if file, ok := f.files[name]; ok {
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("could not rewind the file: %w", err)
		}

		return file, nil
	}

	return nil, fs.ErrNotExist
}

func (f *FS) ReadFile(name string) ([]byte, error) {
	if file, ok := f.files[name]; ok {
		// Rewind the file
		file.Seek(0, io.SeekStart)

		// Read to the buffer
		body := make([]byte, file.Metadata().Size())
		_, err := file.Read(body)

		return body, err
	}

	return nil, fs.ErrNotExist
}

func (f *FS) ReadDir(dirname string) ([]fs.DirEntry, error) {
	contents := []fs.DirEntry{}

	for name, file := range f.files {
		if dirname != path.Dir(name) {
			continue
		}

		dirEntry, err := FileInfoToDirEntry(file, dirname)
		if err != nil {
			return nil, fmt.Errorf("could not compute relative stat for file %v: %w", name, err)
		}

		contents = append(contents, dirEntry)
	}

	return contents, nil
}

var (
	WalkSkipAll = errors.New("skip all files")
	WalkSkipDir = errors.New("skip current directory")
)

type WalkFunc func(path string, fi fs.FileInfo, err error) error

func (f *FS) Walk(relativePath string, walkFn WalkFunc) error {
	if path.Dir(relativePath) == "." {
		relativePath = ""
	}

	skippedDirs := map[string]struct{}{}

	var err error
	for name, file := range f.files {
		if err != nil && errors.Is(err, WalkSkipAll) {
			err = nil
			break
		}

		dirPrefix := path.Dir(name)
		if _, ok := skippedDirs[dirPrefix]; ok {
			continue
		}

		if strings.HasPrefix(name, relativePath) {
			err = walkFn(name, file.Metadata(), err)
			if err != nil && errors.Is(err, WalkSkipDir) {
				skippedDirs[dirPrefix] = struct{}{}
				err = nil
			}
		}
	}

	return err
}
