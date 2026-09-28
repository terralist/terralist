package artifact

const (
	TypeModule   = "module"
	TypeProvider = "provider"
)

type Version struct {
	Tag           string `json:"tag"`
	Documentation string `json:"documentation"`
}

// VersionDetails describes a version of an artifact held by Terralist: where
// it came from, and, for a provider, whether only the network mirror serves
// it, for lack of a signed SHA256SUMS file.
type VersionDetails struct {
	Version    string `json:"version"`
	Origin     string `json:"origin"`
	MirrorOnly bool   `json:"mirror_only,omitempty"`
}

// Versions lists the versions of an artifact, with what the caller may do
// with the artifact: delete its versions, fetch versions from the upstream
// registry of its authority, and block a pulled version, which denies it by a
// rule of the authority and deletes it.
type Versions struct {
	Versions  []VersionDetails `json:"versions"`
	CanDelete bool             `json:"can_delete"`
	CanFetch  bool             `json:"can_fetch"`
	CanBlock  bool             `json:"can_block"`
}

type Artifact struct {
	ID        string   `json:"id"`
	FullName  string   `json:"full_name"`
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Provider  string   `json:"provider"`
	Type      string   `json:"type"`
	Versions  []string `json:"versions"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}
