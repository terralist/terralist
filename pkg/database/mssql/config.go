package mssql

import (
	"fmt"
	"net/url"
)

// Config implements database.Configurator interface and
// handles the configuration parameters of the MSSQL database
type Config struct {
	Username string // the account username
	Password string // the account password
	Hostname string // the hostname where the MSSQL server is hosted
	Port     int    // the port on which the server can be accessed
	Name     string // the database name

	// The database URL can be used to establish the connection without specifying
	// other credentials
	URL string

	// URL in parsed form
	parsedURL *url.URL
}

func (t *Config) SetDefaults() {}

func (t *Config) Validate() error {
	connectionWithParts := !(t.Username == "" || t.Password == "" || t.Hostname == "" || t.Port == 0 || t.Name == "")
	connectionWithURL := !(t.URL == "")

	if !connectionWithParts && !connectionWithURL {
		return fmt.Errorf("no method for connection was provided")
	}

	if connectionWithURL {
		pu, err := url.Parse(t.URL)
		if err != nil {
			return fmt.Errorf("cannot parse connection url: %w", err)
		}

		t.parsedURL = pu
	}

	return nil
}

func (t *Config) DSN() string {
	if t.parsedURL != nil {
		return t.parsedURL.String()
	}

	// MSSQL DSN format: sqlserver://username:password@host:port?database=dbname
	return fmt.Sprintf(
		"sqlserver://%s:%s@%s:%d?database=%s",
		t.Username,
		t.Password,
		t.Hostname,
		t.Port,
		t.Name,
	)
}
