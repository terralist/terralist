package database

// Session is a concrete implementation of session.Session
// backed by a database record.
type Session struct {
	id     string
	values map[any]any
	isNew  bool
}

func newSession(id string, isNew bool) *Session {
	return &Session{
		id:     id,
		values: make(map[any]any),
		isNew:  isNew,
	}
}

func (s *Session) Get(key any) (any, bool) {
	v, ok := s.values[key]
	return v, ok
}

func (s *Session) Set(key any, value any) {
	s.values[key] = value
}

func (s *Session) Unset(key any) {
	delete(s.values, key)
}
