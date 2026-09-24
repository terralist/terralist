package provider

import (
	"fmt"
	"strings"

	"terralist/pkg/file"

	"github.com/google/uuid"
)

// PackagesUploadDTO holds the files of a provider version uploaded directly to
// Terralist, as produced by `terraform providers mirror`: the package archives,
// the version document listing them, and optionally the SHA256SUMS file with
// its signature.
type PackagesUploadDTO struct {
	AuthorityID      uuid.UUID
	Name             string
	Version          string
	Protocols        []string
	Metadata         MirrorArchivesDTO
	Archives         []file.File
	ShaSums          file.File
	ShaSumsSignature file.File
}

// Signed reports whether the upload carries the SHA256SUMS file and its
// signature.
func (d PackagesUploadDTO) Signed() bool {
	return d.ShaSums != nil && d.ShaSumsSignature != nil
}

// ToVersion maps the upload to a provider version without package locations.
func (d PackagesUploadDTO) ToVersion() Version {
	return Version{
		Version:   d.Version,
		Protocols: strings.Join(d.Protocols, ","),
	}
}

// FindArchive returns the platform key and the archive entry listed in the
// version document for the given package file name.
func (d MirrorArchivesDTO) FindArchive(fileName string) (string, MirrorArchiveDTO, bool) {
	for key, entry := range d.Archives {
		if entry.URL == fileName {
			return key, entry, true
		}
	}

	return "", MirrorArchiveDTO{}, false
}

// ToPlatform maps an archive entry to a platform without a location. The key
// is the os_arch platform key under which the archive is listed.
func (d MirrorArchiveDTO) ToPlatform(key string) (Platform, error) {
	system, architecture, ok := strings.Cut(key, "_")
	if !ok || system == "" || architecture == "" {
		return Platform{}, fmt.Errorf("invalid platform key %q, expected os_arch", key)
	}

	var h1 string
	for _, hash := range d.Hashes {
		if strings.HasPrefix(hash, "h1:") {
			h1 = hash
			break
		}
	}

	return Platform{
		System:       system,
		Architecture: architecture,
		H1:           h1,
	}, nil
}
