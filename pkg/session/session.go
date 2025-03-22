package session

// Session is an interface that describes how a session
// should look like.
type Session interface {
	// Get fetches a key and returns its value if it exists
	// else, it will return false as the second parameter.
	Get(key any) (any, bool)

	// Set sets a key's value to a given value.
	Set(key any, value any)

	// Unset removes a key from the session.
	Unset(key any)
}
