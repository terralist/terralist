package mysql

import (
	"fmt"
	"net/url"
)

// Config implements database.Configurator interface and
// handles the configuration parameters of the mysql database.
type Config struct {
	// Username is the account username.
	Username string
	// The account password.
	Password string
	// The hostname where the mysql server is hosted
	Hostname string
	// The port on which the server can be accessed.
	Port int
	// The database name.
	Name string

	// The database URL can be used to establish the connection without specifying
	// other credentials.
	URL string
}

func (t *Config) SetDefaults() {}

func (t *Config) Validate() error {
	connectionWithParts := t.Username != "" && t.Password != "" && t.Hostname != "" && t.Port != 0 && t.Name != ""
	connectionWithURL := t.URL != ""

	if !connectionWithParts && !connectionWithURL {
		return fmt.Errorf("no method for connection was provided")
	}

	if connectionWithURL {
		if _, err := url.Parse(t.URL); err != nil {
			return fmt.Errorf("cannot parse connection url: %w", err)
		}
	}

	return nil
}

func (t *Config) DSN() string {
	if t.URL != "" {
		return t.URL
	}

	// Build the connection string according to the MySQL DSN schema:
	// https://github.com/go-sql-driver/mysql#dsn-data-source-name

	userInfo := url.UserPassword(t.Username, t.Password)
	url := &url.URL{
		Opaque: fmt.Sprintf("%s@tcp(%s:%d)/%s", userInfo.String(), t.Hostname, t.Port, t.Name),
		Host:   fmt.Sprintf("%s:%d", t.Hostname, t.Port),
		Path:   t.Name,
	}

	q := url.Query()

	// MySQL DSN needs to have parseTime=true (https://github.com/go-sql-driver/mysql#timetime-support)
	q.Set("parseTime", "true")

	q.Set("charset", "utf8mb4")
	q.Set("collation", "utf8mb4_unicode_ci")

	url.RawQuery = q.Encode()

	return url.String()
}
