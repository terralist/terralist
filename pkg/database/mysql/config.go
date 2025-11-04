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

	// URL in parsed form.
	parsedURL *url.URL
}

func (t *Config) SetDefaults() {}

func (t *Config) Validate() error {
	connectionWithParts := t.Username != "" && t.Password != "" && t.Hostname != "" && t.Port != 0 && t.Name != ""
	connectionWithURL := t.URL != ""

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
	// MySQL DSN needs to have parseTime=true (https://github.com/go-sql-driver/mysql#timetime-support)

	if t.parsedURL != nil {
		pu := *t.parsedURL
		values := pu.Query()
		values.Set("parseTime", "true")
		// Ensure full Unicode support (emojis)
		if values.Get("charset") == "" {
			values.Set("charset", "utf8mb4")
		}
		if values.Get("collation") == "" {
			values.Set("collation", "utf8mb4_unicode_ci")
		}
		pu.RawQuery = values.Encode()

		return pu.String()
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		t.Username,
		t.Password,
		t.Hostname,
		t.Port,
		t.Name,
	)
}
