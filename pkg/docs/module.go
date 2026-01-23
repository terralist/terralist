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

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"terralist/pkg/file"
)

const (
	tfEntrypointFile   = "main.tf"
	docsEntrypointFile = "README.md"
)

var (
	ErrNoEntrypointFound    = errors.New("could not find an entrypoint")
	ErrNoReadmeFound        = errors.New("no README.md file found")
	ErrNoMainTfFound        = errors.New("no main.tf file found")
	ErrNoDocumentationFound = errors.New("no documentation available")
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

			normalized, nerr := normalizeToUTF8(buf.Bytes())
			if nerr != nil {
				return "", nerr
			}

			return normalized, nil
		}

		return generateModuleDocumentation(moduleFS, relativePath)
	}

	if mainTfErr != nil && readmeErr != nil {
		// Neither main.tf nor README.md found - no way to generate docs
		return "", fmt.Errorf("%w: searched for %s and %s",
			ErrNoEntrypointFound, tfEntrypointFile, docsEntrypointFile)
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

		normalized, nerr := normalizeToUTF8(buf.Bytes())
		if nerr != nil {
			return "", nerr
		}

		return normalized, nil
	}

	// Otherwise, analyze the module and generate documentation for it
	return generateModuleDocumentation(moduleFS, path.Dir(entrypointPath))
}

// normalizeToUTF8 ensures the provided data is returned as a valid UTF-8 string.
// It strips UTF-8 BOM and decodes UTF-16 (LE/BE) content when a BOM is present.
func normalizeToUTF8(data []byte) (string, error) {
	// UTF-8 BOM
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return string(data[3:]), nil
	}

	// UTF-16 LE BOM
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		if ((len(data) - 2) % 2) != 0 {
			return "", fmt.Errorf("could not decode utf-16le: odd number of bytes")
		}
		rdr := transform.NewReader(bytes.NewReader(data[2:]), unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder())
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, rdr); err != nil {
			return "", fmt.Errorf("could not decode utf-16le: %w", err)
		}
		return buf.String(), nil
	}

	// UTF-16 BE BOM
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		if ((len(data) - 2) % 2) != 0 {
			return "", fmt.Errorf("could not decode utf-16be: odd number of bytes")
		}
		rdr := transform.NewReader(bytes.NewReader(data[2:]), unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder())
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, rdr); err != nil {
			return "", fmt.Errorf("could not decode utf-16be: %w", err)
		}
		return buf.String(), nil
	}

	// Assume input is UTF-8 (common for README.md); return as-is
	return string(data), nil
}
