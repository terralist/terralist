package auth

import (
	"encoding/gob"
	"encoding/json"
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

func (u User) MarshalJSON() ([]byte, error) {
	type userAlias User
	return json.Marshal(userAlias(u))
}

func (u *User) UnmarshalJSON(data []byte) error {
	type userAlias User
	var aux userAlias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*u = User(aux)
	return nil
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
