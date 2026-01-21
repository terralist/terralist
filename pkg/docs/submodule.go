package docs

import (
	"io/fs"
	"path"
	"slices"
	"strings"

	"terralist/pkg/file"
)

const (
	// Common subdirectory names that contain Terraform submodules
	submodulesDir = "modules"
	submodulesAlt = "submodules"
)

// SubmoduleInfo contains information about a discovered submodule
type SubmoduleInfo struct {
	Path          string
	Documentation string
}

// FindSubmodules scans the module filesystem for subdirectories containing submodules.
// It looks for directories named "modules" or "submodules" at the root level,
// then identifies each subdirectory as a potential submodule.
func FindSubmodules(moduleFS *file.FS) ([]SubmoduleInfo, error) {
	var submodules []SubmoduleInfo
	submoduleDirsMap := make(map[string]bool)

	// Walk the filesystem to collect all file paths
	var allPaths []string
	if err := moduleFS.Walk("./", func(p string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		allPaths = append(allPaths, p)
		return nil
	}); err != nil {
		return nil, err
	}

	// Extract directories from file paths
	for _, p := range allPaths {
		normalizedPath := strings.TrimPrefix(p, "./")
		if normalizedPath == "" || normalizedPath == "." {
			continue
		}

		parts := strings.Split(normalizedPath, "/")

		// Check if this path is within a "modules" or "submodules" directory
		if len(parts) < 2 {
			continue
		}

		if parts[0] != submodulesDir && parts[0] != submodulesAlt {
			continue
		}

		// Extract all directory levels within modules/submodules
		// For example, "modules/vpc/main.tf" -> ["modules/vpc"]
		// "modules/networking/vpc/main.tf" -> ["modules/networking", "modules/networking/vpc"]
		for i := 2; i <= len(parts); i++ {
			dirPath := strings.Join(parts[:i], "/")
			// Only add if it's a directory (not the file itself)
			if i < len(parts) {
				submoduleDirsMap[dirPath] = true
			}
		}
	}

	// Convert map to slice
	var submoduleDirs []string
	for dir := range submoduleDirsMap {
		submoduleDirs = append(submoduleDirs, dir)
	}

	// Sort by depth (shallowest first)
	slices.SortFunc(submoduleDirs, func(lhs, rhs string) int {
		lhsDepth := strings.Count(lhs, "/")
		rhsDepth := strings.Count(rhs, "/")
		if lhsDepth != rhsDepth {
			return lhsDepth - rhsDepth
		}
		// Secondary sort by name for consistent ordering
		return strings.Compare(lhs, rhs)
	})

	// Filter to keep only leaf directories or directories with main.tf
	// A directory is considered a submodule if:
	// 1. It has no subdirectories that contain main.tf, OR
	// 2. It directly contains main.tf
	filteredDirs := []string{}
	for _, dir := range submoduleDirs {
		hasMainTf := false
		mainTfPath := path.Join(dir, tfEntrypointFile)
		if _, err := moduleFS.Open(mainTfPath); err == nil {
			hasMainTf = true
		}

		hasSubdirWithMainTf := false
		for _, otherDir := range submoduleDirs {
			if otherDir != dir && strings.HasPrefix(otherDir, dir+"/") {
				otherMainTfPath := path.Join(otherDir, tfEntrypointFile)
				if _, err := moduleFS.Open(otherMainTfPath); err == nil {
					hasSubdirWithMainTf = true
					break
				}
			}
		}

		// Include this directory if it has main.tf or has no subdirs with main.tf
		if hasMainTf || !hasSubdirWithMainTf {
			filteredDirs = append(filteredDirs, dir)
		}
	}

	// Generate documentation for each submodule
	for _, submodulePath := range filteredDirs {
		doc, err := GetModuleDocumentation(moduleFS, submodulePath)
		if err != nil {
			// If we can't generate docs, include the submodule anyway with empty docs
			doc = ""
		}

		submodules = append(submodules, SubmoduleInfo{
			Path:          submodulePath,
			Documentation: doc,
		})
	}

	return submodules, nil
}

// uniquePaths removes duplicate paths from a slice
func uniquePaths(paths []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, p := range paths {
		if !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
	}

	return result
}
