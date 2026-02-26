package auth

import (
	"encoding/gob"
	"fmt"
	"strings"
)

// User holds the user authorized user data.
type User struct {
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Authority   string   `json:"authority"`
	AuthorityID string   `json:"authority_id"`
	Groups      []string `json:"groups"`
}

func (u User) String() string {
	return fmt.Sprintf(
		"User{Name: %v, Email: %v, Authority: %v, AuthorityID: %v, Groups: %v}",
		u.Name,
		u.Email,
		u.Authority,
		u.AuthorityID,
		strings.Join(u.Groups, ","),
	)
}

func init() {
	// Register the user interface.
	gob.Register(&User{})
}
