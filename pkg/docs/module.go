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

// findTopLevelFile finds the shallowest occurrence of a given filename in a file system.
func findTopLevelFile(moduleFS *file.FS, fileName string) (string, error) {
	var paths []string
	if err := moduleFS.Walk("./", func(p string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path.Base(p) == fileName {
			paths = append(paths, p)
		}

		return nil
	}); err != nil {
		return "", fmt.Errorf("could not search for %s: %w", fileName, err)
	}

	if len(paths) == 0 {
		return "", ErrNoEntrypointFound
	}

	slices.SortFunc(paths, func(lhs, rhs string) int {
		return strings.Count(lhs, "/") - strings.Count(rhs, "/")
	})

	return paths[0], nil
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

// GetModuleDocumentation traverses a module's file system to find its documentation.
// It prioritizes finding the shallowest `main.tf` to locate the module root. It then
// checks for a `README.md` at the same level or higher. If a suitable `README.md` is
// found, its content is returned. Otherwise, documentation is generated from the .tf files.
// If no `main.tf` is found, it falls back to the shallowest `README.md`.
func GetModuleDocumentation(moduleFS *file.FS, relativePath string) (string, error) {
	mainTfPath, mainTfErr := findTopLevelFile(moduleFS, tfEntrypointFile)
	readmePath, readmeErr := findTopLevelFile(moduleFS, docsEntrypointFile)

	// If a specific subdirectory is requested, prioritize that.
	if relativePath != "" {
		readmeInSubdir := path.Join(relativePath, docsEntrypointFile)
		readmeFile, err := moduleFS.Open(readmeInSubdir)

		if err == nil {
			defer readmeFile.Close()
			buf := new(bytes.Buffer)

			if _, err := io.Copy(buf, readmeFile); err != nil {
				return "", fmt.Errorf("could not read README.md in subdir: %w", err)
			}

			return buf.String(), nil
		}

		return generateModuleDocumentation(moduleFS, relativePath)
	}

	if mainTfErr != nil && readmeErr != nil {
		return "", ErrNoEntrypointFound
	}

	// Logic to decide which entrypoint to use
	var entrypointPath string
	var useReadme bool

	if mainTfErr == nil && readmeErr == nil {
		// Both files exist. Prefer the one that is at a shallower directory level.
		// If README is at the same or higher level as main.tf, use README.
		if strings.Count(readmePath, "/") <= strings.Count(mainTfPath, "/") {
			entrypointPath = readmePath
			useReadme = true
		} else {
			// main.tf is shallower, check for a README in its directory.
			moduleRoot := path.Dir(mainTfPath)
			readmeInModuleRoot := path.Join(moduleRoot, docsEntrypointFile)
			readmeFile, err := moduleFS.Open(readmeInModuleRoot)

			if err == nil {
				readmeFile.Close()
				entrypointPath = readmeInModuleRoot
				useReadme = true
			} else {
				entrypointPath = mainTfPath
				useReadme = false
			}
		}
	} else if readmeErr == nil {
		entrypointPath = readmePath
		useReadme = true
	} else {
		entrypointPath = mainTfPath
		useReadme = false
	}

	if useReadme {
		file, err := moduleFS.Open(entrypointPath)

		if err != nil {
			return "", fmt.Errorf("could not open entrypoint: %w", err)
		}

		defer file.Close()
		buf := new(bytes.Buffer)

		if _, err := io.Copy(buf, file); err != nil {
			return "", fmt.Errorf("could not read file: %w", err)
		}

		return buf.String(), nil
	}

	// Otherwise, analyze the module and generate documentation for it
	return generateModuleDocumentation(moduleFS, path.Dir(entrypointPath))
}
