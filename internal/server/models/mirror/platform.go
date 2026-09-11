package mirror

import (
	"fmt"
	"strings"

	"terralist/pkg/database/entity"

	"github.com/google/uuid"
)

// Platform is a mirrored provider package built for a specific os and
// architecture.
type Platform struct {
	entity.Entity
	VersionID    uuid.UUID
	Version      Version
	System       string `gorm:"not null"`
	Architecture string `gorm:"not null"`
	Location     string `gorm:"not null"`
	Hashes       string `gorm:"not null"`
}

func (Platform) TableName() string {
	return "mirror_platforms"
}

// String returns the platform key used by the network mirror protocol, in the
// form of os_arch.
func (p Platform) String() string {
	return fmt.Sprintf("%s_%s", p.System, p.Architecture)
}

// FileName returns the file name of the provider package as produced by
// `terraform providers mirror`.
func (p Platform) FileName() string {
	return fmt.Sprintf(
		"terraform-provider-%s_%s_%s_%s.zip",
		p.Version.Provider.Name,
		p.Version.Version,
		p.System,
		p.Architecture,
	)
}

// ToArchiveDTO maps the platform to its network mirror protocol archive entry.
// The url is the location from where the package can be downloaded.
func (p Platform) ToArchiveDTO(url string) ArchiveDTO {
	return ArchiveDTO{
		URL:    url,
		Hashes: strings.Split(p.Hashes, ","),
	}
}

// ArchivesDTO is the network mirror protocol "List Available Installation
// Packages" document. It is both the format served to Terraform and the format
// produced by `terraform providers mirror` for a version.
type ArchivesDTO struct {
	Archives map[string]ArchiveDTO `json:"archives"`
}

// ArchiveDTO describes a single provider package inside an ArchivesDTO.
type ArchiveDTO struct {
	URL    string   `json:"url"`
	Hashes []string `json:"hashes"`
}

// ToPlatform maps an archive entry to a platform. The key is the os_arch
// platform key under which the archive is listed.
func (d ArchiveDTO) ToPlatform(key string) (Platform, error) {
	system, architecture, ok := strings.Cut(key, "_")
	if !ok || system == "" || architecture == "" {
		return Platform{}, fmt.Errorf("invalid platform key %q, expected os_arch", key)
	}

	return Platform{
		System:       system,
		Architecture: architecture,
		Hashes:       strings.Join(d.Hashes, ","),
	}, nil
}
