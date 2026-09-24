package provider

import "fmt"

// MirrorVersionListDTO is the network mirror protocol "List Available
// Versions" document. Each version maps to an empty object, as the protocol
// requires.
type MirrorVersionListDTO struct {
	Versions map[string]struct{} `json:"versions"`
}

func (p Provider) ToMirrorVersionListDTO() MirrorVersionListDTO {
	versions := make(map[string]struct{}, len(p.Versions))
	for _, v := range p.Versions {
		versions[v.Version] = struct{}{}
	}

	return MirrorVersionListDTO{
		Versions: versions,
	}
}

// MirrorArchivesDTO is the network mirror protocol "List Available
// Installation Packages" document, keyed by os_arch platform.
type MirrorArchivesDTO struct {
	Archives map[string]MirrorArchiveDTO `json:"archives"`
}

// MirrorArchiveDTO describes a single provider package inside a
// MirrorArchivesDTO.
type MirrorArchiveDTO struct {
	URL    string   `json:"url"`
	Hashes []string `json:"hashes"`
}

func (v Version) ToMirrorArchivesDTO() MirrorArchivesDTO {
	archives := make(map[string]MirrorArchiveDTO, len(v.Platforms))
	for _, p := range v.Platforms {
		archives[p.String()] = p.ToMirrorArchiveDTO()
	}

	return MirrorArchivesDTO{
		Archives: archives,
	}
}

// ToMirrorArchiveDTO maps the platform to its network mirror protocol archive
// entry. The h1 hash is listed when known, followed by the zh hash, which is
// the sha256 of the package as listed in the SHA256SUMS file.
func (p Platform) ToMirrorArchiveDTO() MirrorArchiveDTO {
	hashes := make([]string, 0, 2)
	if p.H1 != "" {
		hashes = append(hashes, p.H1)
	}
	hashes = append(hashes, fmt.Sprintf("zh:%s", p.ShaSum))

	return MirrorArchiveDTO{
		URL:    p.Location,
		Hashes: hashes,
	}
}
