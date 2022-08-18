package file

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-getter"
)

const (
	tempDirPattern = "tl-fetch"
)

// FetchFile downloads a file from a given URL and returns it
// as an InMemoryFile
func FetchFile(name string, url string) (*InMemoryFile, error) {
	return fetch(name, url, true)
}

// FetchDir downloads all files from a given URL and returns
// them as an archive, stored in an InMemoryFile object
func FetchDir(name string, url string) (*InMemoryFile, error) {
	return fetch(name, url, false)
}

// fetch downloads a file/directory from a given URL and loads them into the memory
func fetch(name string, url string, isFile bool) (*InMemoryFile, error) {
	tempDir, err := os.MkdirTemp("", tempDirPattern)
	if err != nil {
		return nil, fmt.Errorf("%w: could not create temp dir: %v", ErrSystemFailure, err)
	}
	defer os.RemoveAll(tempDir)

	dst := tempDir
	if isFile {
		dst = path.Join(tempDir, name)
	}

	ctx, cancel := context.WithCancel(context.Background())
	client := &getter.Client{
		Ctx: ctx,
		Src: url,
		Dst: dst,
		Pwd: tempDir,
		Options: []getter.ClientOption{
			getter.WithInsecure(),
		},
	}

	if isFile {
		client.Mode = getter.ClientModeFile
	} else {
		client.Mode = getter.ClientModeDir
	}

	// Launch the download process
	wg := sync.WaitGroup{}
	wg.Add(1)
	ech := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()

		if err := client.Get(); err != nil {
			ech <- err
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)

	// Wait for the download process to finish
	select {
	case sig := <-sc:
		signal.Reset(os.Interrupt)
		cancel()
		wg.Wait()

		return nil, fmt.Errorf("%w: signal %v received", ErrDownloadInterrupt, sig.String())
	case err := <-ech:
		wg.Wait()

		return nil, fmt.Errorf("%w: %v", ErrDownloadFailure, err)
	case <-ctx.Done():
		wg.Wait()

		if isFile {
			return readFile(name, dst)
		}

		return archiveDir(name, dst)
	}
}

// readFile reads a file from the disk and returns it
// as an InMemoryFile
func readFile(name, src string) (*InMemoryFile, error) {
	content, err := ioutil.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("%w: cannot read downloaded file: %v", ErrSystemFailure, err)
	}

	return &InMemoryFile{
		Name:    name,
		Content: content,
	}, nil
}

// archiveDir reads a directory from the disk, archives it
// and returns the archive file as an InMemoryFile
func archiveDir(name, src string) (*InMemoryFile, error) {
	dirFiles := []*InMemoryFile{}

	// Walk recursively through the given directory
	if err := filepath.Walk(src, func(filePath string, info os.FileInfo, err error) error {
		// If there's an error on other files, propagate the error and return
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Open the file
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		relPath, _ := filepath.Rel(src, filePath)

		buffer := new(bytes.Buffer)

		// Copy the file to the buffer
		if _, err := io.Copy(buffer, f); err != nil {
			return err
		}

		dirFiles = append(dirFiles, &InMemoryFile{
			Name:    relPath,
			Content: buffer.Bytes(),
		})

		return nil
	}); err != nil {
		return nil, fmt.Errorf("%w: could not parse downloaded dir: %v", ErrSystemFailure, err)
	}

	archive, err := Archive(name, dirFiles)
	if err != nil {
		return nil, err
	}

	return archive, nil
}
