package cookie

import (
	"fmt"
	"net/http"

	"terralist/pkg/session"

	"github.com/gorilla/sessions"
)

// Store is a concrete implementation of session.Store and uses
// gorilla CookieStore as backend.
type Store struct {
	name string

	store *sessions.CookieStore
}

func (s *Store) Get(r *http.Request) (session.Session, error) {
	sess, err := s.store.Get(r, s.name)
	if err != nil {
		return nil, err
	}

	return &Session{
		session: sess,
	}, nil
}

func (s *Store) New(r *http.Request) (session.Session, error) {
	sess, err := s.store.New(r, s.name)
	if err != nil {
		return nil, err
	}

	return &Session{
		session: sess,
	}, nil
}

func (s *Store) Save(r *http.Request, w http.ResponseWriter, sess session.Session) error {
	if impl, ok := sess.(*Session); ok {
		return s.store.Save(r, w, impl.session)
	}

	return fmt.Errorf("unsupported session type")
}
