package mysql

import (
	"fmt"
	"net/url"
)

// Config implements database.Configurator interface and
// handles the configuration parameters of the mysql database
type Config struct {
	Username string // the account username
	Password string // the account password
	Hostname string // the hostname where the mysql server is hosted
	Port     int    // the port on which the server can be accessed
	Name     string // the database name

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

	if t.URL != "" {
		pu, err := url.Parse(t.URL)
		if err != nil {
			return fmt.Errorf("cannot parse connection url: %w", err)
		}

		// https://github.com/go-sql-driver/mysql#timetime-support
		if pu.Query().Get("parseTime") != "true" {
			return fmt.Errorf("mysql: connection url must include property parseTime=true")
		}
	}

	return nil
}

func (t *Config) DSN() string {
	if t.URL != "" {
		return t.URL
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true",
		t.Username,
		t.Password,
		t.Hostname,
		t.Port,
		t.Name,
	)
}
