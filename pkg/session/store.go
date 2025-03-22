package session

import "net/http"

// Store is an interface that describes how a session
// store should look like.
// Interface is a copy of gorilla/session.Store, but uses
// our own Session implementation.
type Store interface {
	// Get should return a cached session.
	Get(r *http.Request) (Session, error)

	// New should create and return a new session.
	//
	// Note that New should never return a nil session, even in the case of
	// an error if using the Registry infrastructure to cache the session.
	New(r *http.Request) (Session, error)

	// Save should persist session to the underlying store implementation.
	Save(r *http.Request, w http.ResponseWriter, s Session) error
}
