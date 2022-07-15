package zipper

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/mazen160/go-random"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	errorSystemFailure = errors.New("system failure")
)

var (
	ErrSystemFailure = func() error { return errorSystemFailure }()
)

// Zip archives all files from a given source directory and stores
// the resulting archive in a given destination directory, with a
// random generated name
// It will return the full path to the resulting archive
func Zip(src string, destination string) (string, error) {
	log.Info().
		Str("Source", src).
		Msg("Archiving directory")

	archiveName := nextArchiveName(filepath.Base(src))
	archivePath := filepath.Clean(path.Join(destination, archiveName))

	archive, err := os.Create(archivePath)
	if err != nil {
		return "", fmt.Errorf("%w: cannot create archive: %v", ErrSystemFailure, err)
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

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

		// Skip the archive file
		if filePath == archivePath {
			return nil
		}

		// Open the file
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		relPath, _ := filepath.Rel(src, filePath)

		// Create an associate in the given archive
		w, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// Copy the file to the archive
		if _, err := io.Copy(w, f); err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Error().
			Str("Source", src).
			Str("Archive", archiveName).
			AnErr("Error", err).
			Msg("Could not create archive")

		return "", fmt.Errorf("%w: %v", ErrSystemFailure, err)
	}

	log.Info().
		Str("Source", src).
		Str("Archive", archiveName).
		Msg("Directory archived")

	return archivePath, nil
}

func nextArchiveName(basename string) string {
	timestamp := time.Now().Format("20170907170606")
	name, _ := random.String(8)

	return fmt.Sprintf("%s_%s_%s.zip", basename, name, timestamp)
}
