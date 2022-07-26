package auth

// User holds the user authorized user data
type User struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	AuthorityID string `json:"authority_id"`
}
