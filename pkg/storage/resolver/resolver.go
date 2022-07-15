package resolver

// Resolver handles the storage and resolve operations
type Resolver interface {
	// Store downloads a file from a given url and stores it somewhere, then
	// return a unique key to identify the file
	Store(url string, archive bool) (string, error)

	// Find receives a key and returns a URL from which the file can be
	// downloaded
	Find(key string) (string, error)

	// Purge removes the document stored at a given key
	// If the given key does not exist, it will not return an error
	Purge(string) error
}
