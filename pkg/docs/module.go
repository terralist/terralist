package docs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"

	"terralist/pkg/file"
)

const (
	tfEntrypointFile   = "main.tf"
	docsEntrypointFile = "README.md"
)

var (
	ErrNoEntrypointFound = errors.New("could not find an entrypoint")
)

// findModuleRoot finds the top-level directory in a FS.
// It returns the relative path to this dir.
func findModuleRoot(moduleFS *file.FS) (string, error) {
	var results []string = []string{}

	if err := moduleFS.Walk("./", func(p string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path.Base(p) == tfEntrypointFile || path.Base(p) == docsEntrypointFile {
			results = append(results, p)
		}

		return nil
	}); err != nil {
		return "", fmt.Errorf("could not traverse the module: %w", err)
	}

	slices.SortFunc(results, func(lhs, rhs string) int {
		lhsParts := strings.Split(lhs, "/")
		rhsParts := strings.Split(rhs, "/")

		if len(lhsParts) < len(rhsParts) {
			return -1
		} else if len(lhsParts) > len(rhsParts) {
			return 1
		}

		// We don't really care if the folder depth is equal.
		// Whichever comes first goes first.
		return 0
	})

	if len(results) == 0 {
		return ".", nil
	}

	return path.Dir(results[0]), nil
}

// generateModuleDocumentation traverse a module FS and analyze module files.
// It returns a generic Markdown documentation for the module.
func generateModuleDocumentation(moduleFS *file.FS, rootRelativePath string) (string, error) {
	module, diags := tfconfig.LoadModuleFromFilesystem(tfconfig.WrapFS(moduleFS), rootRelativePath)
	if diags.HasErrors() {
		return "", diags.Err()
	}

	buf := new(bytes.Buffer)
	bw := bufio.NewWriter(buf)

	if err := tfconfig.RenderMarkdown(bw, module); err != nil {
		return "", err
	}

	if err := bw.Flush(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GetModuleDocumentation traverse a module FS and returns a markdown documentation for it.
// If the relative path to the module is known (relative within the FS), it can be passed as
// the second argument, otherwise, the function will try to detect the top-level directory
// which contains either a README.md file, or a main.tf file.
// If the README.md file is found, the content of this file is returned. If a main.tf file is
// found, the module is automatically analyzed and a generic documentation will be generated
// for it.
func GetModuleDocumentation(moduleFS *file.FS, rootRelativePath string) (string, error) {
	if rootRelativePath == "" {
		// If there is no rootRelativePath, we need to find it
		var err error
		rootRelativePath, err = findModuleRoot(moduleFS)
		if err != nil {
			return "", fmt.Errorf("could not find module root: %w", err)
		}
	}

	// We need to search the rootRelativePath and find all possible entrypoints
	foundEntrypoints := []string{}
	if err := moduleFS.Walk(rootRelativePath, func(filepath string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If we are searching in a subdirectory, we can skip it
		if path.Dir(filepath) != rootRelativePath {
			return file.WalkSkipDir
		}

		// Save any entrypoint found
		if path.Base(filepath) == tfEntrypointFile || path.Base(filepath) == docsEntrypointFile {
			foundEntrypoints = append(foundEntrypoints, filepath)
		}

		return nil
	}); err != nil {
		return "", fmt.Errorf("could not find module entrypoint: %w", err)
	}

	if len(foundEntrypoints) == 0 {
		return "", ErrNoEntrypointFound
	}

	// Sort the entrypoints to make sure the docs entrypoint goes first
	slices.SortFunc(foundEntrypoints, func(lhs, rhs string) int {
		if path.Base(lhs) == tfEntrypointFile && path.Base(rhs) == docsEntrypointFile {
			return 1
		} else if path.Base(lhs) == docsEntrypointFile && path.Base(rhs) == tfEntrypointFile {
			return -1
		}

		return 0
	})

	entrypoint := foundEntrypoints[0]

	// If we found a docs entrypoint, read it and return the content as the documentation
	if path.Base(entrypoint) == docsEntrypointFile {
		file, err := moduleFS.Open(entrypoint)
		if err != nil {
			return "", fmt.Errorf("could not open entrypoint: %w", err)
		}

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, file); err != nil {
			return "", fmt.Errorf("could not read file: %w", err)
		}

		return buf.String(), nil
	}

	// Otherwise, analyze the module and generate documentation for it
	return generateModuleDocumentation(moduleFS, rootRelativePath)
}
