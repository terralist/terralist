package storage

// StoreInput holds the inputs for the Store method
type StoreInput struct {
	// Content stores the data itself that will be stored
	Content []byte

	// KeyPrefix stores any custom key prefix that will be
	// applied to the resulted key
	// Also represents the dirname of the datastore path
	KeyPrefix string

	// FileName represents the name of the file and will be
	// applied to the resulted key
	// Also represents the basename of the datastore path
	FileName string
}

// Resolver handles the storage and resolve operations
type Resolver interface {
	// Store uploads a file to the resolver datastore and returns
	// a unique key
	Store(*StoreInput) (string, error)

	// Find receives a key and returns a URL from where the
	// stored file can be downloaded
	Find(keys string) (string, error)

	// Purge removes the document stored at a given key
	// If the given key does not exist, it will not return an error
	Purge(string) error
}
