package file

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"

	getter "github.com/hashicorp/go-getter"
	urlhelper "github.com/hashicorp/go-getter/helper/url"
)

const (
	tempDirPattern = "tl-fetch"
)

// generateGetters returns the map of getters.
// Modified version of https://github.com/hashicorp/go-getter/blob/f7836fb97529673f24dac0aaa140762ee05c847f/get.go#L65
// to add support for custom http headers.
func generateGetters(header http.Header) map[string]getter.Getter {
	httpGetter := &getter.HttpGetter{
		Netrc:  true,
		Header: header,
	}

	return map[string]getter.Getter{
		"file":  new(getter.FileGetter),
		"git":   new(getter.GitGetter),
		"gcs":   new(getter.GCSGetter),
		"hg":    new(getter.HgGetter),
		"s3":    new(getter.S3Getter),
		"http":  httpGetter,
		"https": httpGetter,
	}
}

// fetch downloads a file/directory from a given URL and loads them.
func fetch(name string, url string, checksum string, kind int, header http.Header) (File, error) {
	tempDir, err := os.MkdirTemp("", tempDirPattern)
	if err != nil {
		return nil, fmt.Errorf("%w: could not create temp dir: %v", ErrSystemFailure, err)
	}
	// TODO: We need to keep the temp dir to read the files and automatically generate the module
	// documentation. This is not great, as over time we risk accumulating a lot of garbage.
	// However, we ask the OS to give us a temp directory, which means the OS would handle
	// its garbage collection. Unfortunately, this is not ideal, as for some OSes, the
	// garbage collection of the tmp dir differs, some uses crons, some only does it on reboot,
	// some never do it.
	// Terralist runs as a server, so a reboot to clean the disk space is not a solution.
	// Officially, Terralist is bundled on an Alpine distribution, and, in Alpine, by default
	// the tmp directory is never cleaned.
	// We can patch this by manually setting a cron to do it (also enable clean-up on boot might
	// help), but definitely, that is not a fix. If Terralist is bundled in another OS, the entire
	// process have to be handled by the user.
	// We should implement an asynchronous worker that automatically cleans those files for
	// Terralist, once in a while, so we get the same behavior on any OS we run.
	// defer os.RemoveAll(tempDir)

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
		Ctx:     ctx,
		Src:     u.String(),
		Dst:     dst,
		Pwd:     tempDir,
		Getters: generateGetters(header),
	}

	switch kind {
	case file:
		client.Mode = getter.ClientModeFile
	case dir:
		client.Mode = getter.ClientModeDir
	case unknown:
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
// File.
func parseResult(name, src string, kind int) (File, error) {
	switch kind {
	case file:
		return readFile(name, src)
	case dir:
		return archiveDir(name, src)
	}

	return nil, fmt.Errorf("%w: unknown file type", ErrSystemFailure)
}

// readFile reads a file from the disk and returns it
// as an File.
func readFile(name, src string) (File, error) {
	return LoadFromDisk(name, src)
}

// archiveDir reads a directory from the disk, archives it
// and returns the archive file.
func archiveDir(name, src string) (File, error) {
	dirFiles := []File{}
	var rootDir string

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
		relPath = strings.ReplaceAll(relPath, "\\", "/")

		// Detect if all files are under a single root directory
		// (common in GitHub archive downloads like terraform-aws-eks-21.15.1/)
		if rootDir == "" {
			parts := strings.Split(relPath, "/")
			if len(parts) > 1 {
				rootDir = parts[0]
			}
		}

		file, err := readFile(relPath, fpath)
		if err != nil {
			return fmt.Errorf("%w: cannot read downloaded file: %v", ErrSystemFailure, err)
		}

		dirFiles = append(dirFiles, file)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("%w: could not parse downloaded dir: %v", ErrSystemFailure, err)
	}

	// If all files share a common root directory, strip it to normalize the archive structure
	// This handles GitHub archive downloads which include a version-specific root folder
	if rootDir != "" {
		allFilesInRoot := true
		for _, f := range dirFiles {
			if !strings.HasPrefix(f.Name(), rootDir+"/") {
				allFilesInRoot = false
				break
			}
		}

		if allFilesInRoot {
			// Strip the root directory from all file paths
			strippedFiles := make([]File, len(dirFiles))
			for i, f := range dirFiles {
				originalName := f.Name()
				strippedName := strings.TrimPrefix(originalName, rootDir+"/")

				// Read the content from the original file
				var buf bytes.Buffer
				if _, err := io.Copy(&buf, f); err != nil {
					return nil, fmt.Errorf("%w: failed to copy file content: %v", ErrSystemFailure, err)
				}

				// Create a new file with the stripped name
				strippedFiles[i] = NewInMemoryFile(strippedName, buf.Bytes())
			}
			dirFiles = strippedFiles
		}
	}

	archive, err := Archive(name, dirFiles)
	if err != nil {
		return nil, err
	}

	return archive, nil
}
