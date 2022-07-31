package cookie

import "github.com/gorilla/sessions"

// Session is a concrete implementation of session.Session
// and is a wrapper over gorilla's session
type Session struct {
	session *sessions.Session
}

func (s *Session) Get(key any) (any, bool) {
	if v, ok := s.session.Values[key]; ok {
		return v, true
	}

	return nil, false
}

func (s *Session) Set(key any, value any) {
	s.session.Values[key] = value
}

func (s *Session) Unset(key any) {
	delete(s.session.Values, key)
}
