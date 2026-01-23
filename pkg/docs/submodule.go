package docs

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	"terralist/pkg/file"
)

const (
	// submodulesDir is the conventional directory name used by the Terraform
	// community for organizing reusable module components. This is the
	// recommended pattern according to the Terraform module structure guide.
	// See: https://developer.hashicorp.com/terraform/language/modules/develop/structure
	submodulesDir = "modules"

	// submodulesAlt is an alternative directory name sometimes used for
	// organizing module components. While "modules" is the recommended
	// convention, some projects use "submodules" for clarity or to avoid
	// confusion with the root "modules" directory in multi-module repositories.
	submodulesAlt = "submodules"
)

// SubmoduleInfo contains information about a discovered submodule.
type SubmoduleInfo struct {
	Path          string
	Documentation string
}

// dirInfo stores metadata about a directory discovered during the walk.
type dirInfo struct {
	hasMainTf bool
	depth     int
}

// FindSubmodules scans the module filesystem for subdirectories containing submodules.
// It looks for directories named "modules" or "submodules" at the root level,
// then identifies each subdirectory as a potential submodule.
func FindSubmodules(moduleFS *file.FS) ([]SubmoduleInfo, error) {
	var submodules []SubmoduleInfo
	dirs := make(map[string]*dirInfo)

	// Single pass: collect all directories and main.tf locations
	if err := moduleFS.Walk("./", func(p string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		normalizedPath := strings.TrimPrefix(p, "./")
		if normalizedPath == "" || normalizedPath == "." {
			return nil
		}

		parts := strings.Split(normalizedPath, "/")

		// Check if this path is within a "modules" or "submodules" directory
		if len(parts) < 2 {
			return nil
		}

		if parts[0] != submodulesDir && parts[0] != submodulesAlt {
			return nil
		}

		// If this is main.tf, mark the parent directory
		if path.Base(p) == tfEntrypointFile {
			dir := path.Dir(normalizedPath)
			if dirs[dir] == nil {
				dirs[dir] = &dirInfo{depth: strings.Count(dir, "/")}
			}
			dirs[dir].hasMainTf = true
		}

		// Record all directory levels in the path
		// For example, "modules/networking/vpc/main.tf" records:
		// - modules/networking
		// - modules/networking/vpc
		for i := 2; i <= len(parts); i++ {
			dirPath := strings.Join(parts[:i], "/")
			// Only add if it's a directory (not the file itself)
			if i < len(parts) {
				if dirs[dirPath] == nil {
					dirs[dirPath] = &dirInfo{depth: strings.Count(dirPath, "/")}
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// Convert map to sorted slice for deterministic output
	var submoduleDirs []string
	for dir := range dirs {
		submoduleDirs = append(submoduleDirs, dir)
	}

	// Sort by depth (shallowest first), then by name
	slices.SortFunc(submoduleDirs, func(lhs, rhs string) int {
		if dirs[lhs].depth != dirs[rhs].depth {
			return dirs[lhs].depth - dirs[rhs].depth
		}
		return strings.Compare(lhs, rhs)
	})

	// Filter to keep only leaf directories or directories with main.tf
	// A directory is considered a submodule if:
	// 1. It has no subdirectories that contain main.tf, OR
	// 2. It directly contains main.tf
	filteredDirs := []string{}
	for _, dir := range submoduleDirs {
		info := dirs[dir]

		// Check if any subdirectory has main.tf
		hasSubdirWithMainTf := false
		for otherDir, otherInfo := range dirs {
			if otherInfo.hasMainTf && otherDir != dir && strings.HasPrefix(otherDir, dir+"/") {
				hasSubdirWithMainTf = true
				break
			}
		}

		// Include this directory if it has main.tf or has no subdirs with main.tf
		if info.hasMainTf || !hasSubdirWithMainTf {
			filteredDirs = append(filteredDirs, dir)
		}
	}

	// Generate documentation for each submodule
	for _, submodulePath := range filteredDirs {
		doc, err := GetModuleDocumentation(moduleFS, submodulePath)
		if err != nil {
			// If we can't generate docs, provide a helpful message instead of empty string
			// This helps users understand why documentation is missing
			if errors.Is(err, ErrNoEntrypointFound) {
				doc = "# Documentation Not Available\n\nNo README.md or main.tf file found in this submodule."
			} else {
				doc = fmt.Sprintf("# Documentation Not Available\n\nFailed to generate documentation: %s", err.Error())
			}
		}

		submodules = append(submodules, SubmoduleInfo{
			Path:          submodulePath,
			Documentation: doc,
		})
	}

	return submodules, nil
}
