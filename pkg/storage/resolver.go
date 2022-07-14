package storage

// Resolver handles the storage and resolve operations
type Resolver interface {
	// Store downloads a file from a given url and stores it somewhere, then
	// return a Terraform-compliant download URL
	Store(string) (string, error)

	// Purge removes the document stored at a given URL, if the URL is valid
	Purge(string) error
}
