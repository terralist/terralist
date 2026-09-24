package provider

import "strings"

const (
	packagePrefix = "terraform-provider-"
	packageSuffix = ".zip"
)

// Package identifies a provider package by the coordinates carried in its
// file name.
type Package struct {
	Name         string
	Version      string
	System       string
	Architecture string
}

// ParsePackageFileName reads the coordinates out of a package file name of the
// form terraform-provider-<name>_<version>_<os>_<arch>.zip.
func ParsePackageFileName(fileName string) (Package, bool) {
	if !strings.HasPrefix(fileName, packagePrefix) || !strings.HasSuffix(fileName, packageSuffix) {
		return Package{}, false
	}

	parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(fileName, packagePrefix), packageSuffix), "_")
	if len(parts) != 4 {
		return Package{}, false
	}

	for _, part := range parts {
		if part == "" {
			return Package{}, false
		}
	}

	return Package{
		Name:         parts[0],
		Version:      parts[1],
		System:       parts[2],
		Architecture: parts[3],
	}, true
}
