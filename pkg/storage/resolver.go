package storage

// StoreInput holds the inputs for the Store method
type StoreInput struct {
	// URL represents the url source of the data that will be stored
	// Conflicts with Content
	URL string

	// Content stores the data itself that will be stored
	// Conflicts with URL
	Content []byte

	// KeyPrefix stores any custom key prefix that will be
	// applied to the resulted key
	// Also represents the dirname of the datastore path
	KeyPrefix string

	// FileName represents the name of the file and will be
	// applied to the resulted key
	// Also represents the basename of the datastore path
	FileName string

	// Archive controls the archiving process, if enabled,
	// the data received will be stored in an archive
	// Does not apply for Content
	Archive bool
}

// Resolver handles the storage and resolve operations
type Resolver interface {
	// Store resolve a source and uploads it to the resolver datastore
	// then return a unique key to access it later
	Store(*StoreInput) (string, error)

	// Find receives a key and returns a URL from which the file can be
	// downloaded
	Find(key string) (string, error)

	// Purge removes the document stored at a given key
	// If the given key does not exist, it will not return an error
	Purge(string) error
}
