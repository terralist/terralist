package artifact

const (
	TypeModule   = "module"
	TypeProvider = "provider"
)

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
