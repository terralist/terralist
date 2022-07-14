package version

import (
	"regexp"
)

var (
	// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
	versionRegEx = regexp.MustCompile(`(?m)^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
)

// Version is a wrapper over the string primitive used to extract
// parts from a semantic version
type Version string

// Valid returns true if a version respect the semantic versioning
// standard
func (v Version) Valid() bool {
	return versionRegEx.Match([]byte(v))
}

// Major returns the first part of a semantic version
func (v Version) Major() string {
	return *v.part(0)
}

// Minor returns the second part of a semantic version
func (v Version) Minor() string {
	return *v.part(1)
}

// Patch returns the third part of a semantic version
func (v Version) Patch() string {
	return *v.part(2)
}

// PreRelease returns the fourth part of a semantic version
// If missing, the string will be empty
func (v Version) PreRelease() string {
	p := v.part(3)
	if p == nil {
		return ""
	}

	return *p
}

// BuildMetadata returns the fifth part of a semantic version
// If missing, the string will be empty
func (v Version) BuildMetadata() string {
	p := v.part(4)
	if p == nil {
		return ""
	}

	return *p
}

func (v Version) part(id int) *string {
	matches := versionRegEx.FindAllStringSubmatch(string(v), -1)

	if matches == nil || len(matches) != 1 {
		return nil
	}

	return &matches[0][id+1]
}
