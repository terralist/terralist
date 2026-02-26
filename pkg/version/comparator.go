package version

import "strings"

// Compare compares two versions and returns:
//
//   - -1  if lhs < rhs
//   - 0   if lhs = rhs
//   - 1   if lhs > rhs
func Compare(lhs Version, rhs Version) int {
	major := strings.Compare(lhs.Major(), rhs.Major())
	if major != 0 {
		return major
	}

	minor := strings.Compare(lhs.Minor(), rhs.Minor())
	if minor != 0 {
		return minor
	}

	patch := strings.Compare(lhs.Patch(), rhs.Patch())
	if patch != 0 {
		return patch
	}

	preRelease := strings.Compare(lhs.PreRelease(), rhs.PreRelease())
	if preRelease != 0 {
		return preRelease
	}

	buildMetadata := strings.Compare(lhs.BuildMetadata(), rhs.BuildMetadata())
	if buildMetadata != 0 {
		return buildMetadata
	}

	return 0
}
