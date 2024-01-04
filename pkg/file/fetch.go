package file

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/go-getter"
	urlhelper "github.com/hashicorp/go-getter/helper/url"
)

const (
	tempDirPattern = "tl-fetch"
)

// fetch downloads a file/directory from a given URL and loads them into the memory
func fetch(name string, url string, checksum string, kind int) (*InMemoryFile, error) {
	tempDir, err := os.MkdirTemp("", tempDirPattern)
	if err != nil {
		return nil, fmt.Errorf("%w: could not create temp dir: %v", ErrSystemFailure, err)
	}
	defer os.RemoveAll(tempDir)

	dst := tempDir
	if kind == file {
		dst = path.Join(tempDir, name)
	}

	u, err := urlhelper.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSystemFailure, err)
	}

	// Set extra arguments
	q := u.Query()
	if kind == file {
		// Force go-getter to avoid decompressing
		q.Add("archive", "false")
	}
	if checksum != "" {
		// Add a checksum to be checked if we have one
		q.Add("checksum", checksum)
	}
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithCancel(context.Background())
	client := &getter.Client{
		Ctx: ctx,
		Src: u.String(),
		Dst: dst,
		Pwd: tempDir,
	}

	if kind == file {
		client.Mode = getter.ClientModeFile
	} else if kind == dir {
		client.Mode = getter.ClientModeDir
	} else if kind == any {
		client.Mode = getter.ClientModeAny
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

		// If we know the time, just parse it
		if kind == file || kind == dir {
			return parseResult(name, dst, kind)
		}

		// We need to find out what we have downloaded
		inf, err := os.Stat(dst)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSystemFailure, err)
		}

		if inf.IsDir() {
			return parseResult(name, dst, dir)
		}

		return parseResult(name, dst, file)
	}
}

// parseResult parses the download result and returns an
// InMemoryFile
func parseResult(name, src string, kind int) (*InMemoryFile, error) {
	if kind == file {
		return readFile(name, src)
	} else if kind == dir {
		return archiveDir(name, src)
	}

	return nil, fmt.Errorf("%w: unknown file type", ErrSystemFailure)
}

// readFile reads a file from the disk and returns it
// as an InMemoryFile
func readFile(name, src string) (*InMemoryFile, error) {
	// Open the file
	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("%w: cannot read downloaded file: %v", ErrSystemFailure, err)
	}
	defer f.Close()

	// Fetch file headers
	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("%w: cannot stat downloaded file: %v", ErrSystemFailure, err)
	}

	buffer := new(bytes.Buffer)

	// Copy the file to the buffer
	if _, err := io.Copy(buffer, f); err != nil {
		return nil, err
	}

	return &InMemoryFile{
		Name:     name,
		FileInfo: fi,
		Content:  buffer.Bytes(),
	}, nil
}

// archiveDir reads a directory from the disk, archives it
// and returns the archive file as an InMemoryFile
func archiveDir(name, src string) (*InMemoryFile, error) {
	dirFiles := []*InMemoryFile{}

	// Walk recursively through the given directory
	if err := filepath.Walk(src, func(fpath string, info os.FileInfo, err error) error {
		// If there's an error on other files, propagate the error and return
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(src, fpath)
		relPath = filepath.Clean(relPath)
		relPath = strings.Replace(relPath, "\\", "/", -1)

		file, err := readFile(relPath, fpath)
		if err != nil {
			return fmt.Errorf("%w: cannot read downloaded file: %v", ErrSystemFailure, err)
		}

		dirFiles = append(dirFiles, &InMemoryFile{
			Name:     relPath,
			FileInfo: file.FileInfo,
			Content:  file.Content,
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
