package auth

import (
	"encoding/gob"
	"encoding/json"
)

// User holds the user authorized user data
type User struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	AuthorityID string `json:"authority_id"`
}

func (u User) MarshalJSON() ([]byte, error) {
	return json.Marshal(u)
}

func (u User) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &u)
}

func init() {
	// Register the user interface
	gob.Register(&User{})
}
