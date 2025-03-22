package postgresql

import (
	"fmt"
)

// Config implements database.Configurator interface and
// handles the configuration parameters of the postgresql database.
type Config struct {
	// The account username.
	Username string
	// The account password.
	Password string
	// The hostname where the mysql server is hosted.
	Hostname string
	// The port on which the server can be accessed.
	Port int
	// The database name.
	Name string

	// The database URL can be used to establish the connection without specifying
	// other credentials
	URL string
}

func (t *Config) SetDefaults() {}

func (t *Config) Validate() error {
	connectionWithParts := !(t.Username == "" || t.Password == "" || t.Hostname == "" || t.Port == 0 || t.Name == "")
	connectionWithURL := !(t.URL == "")

	if !connectionWithParts && !connectionWithURL {
		return fmt.Errorf("no method for connection was provided")
	}

	return nil
}

func (t *Config) DSN() string {
	if t.URL != "" {
		return t.URL
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		t.Username,
		t.Password,
		t.Hostname,
		t.Port,
		t.Name,
	)
}
