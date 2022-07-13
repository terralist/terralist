package mysql

import (
	"fmt"
	"strconv"
	"strings"
)

// Config implements database.Configurator interface and
// handles the configuration parameters of the sqlite database
type Config struct {
	Username string // the mysql account username
	Password string // the mysql account password
	Hostname string // the hostname where the mysql server is hosted
	Port     int    // the port on which the server can be accessed
	Name     string // the database name

	// The database URL can be used to establish the connection without specifying
	// other credentials
	URL string

	// MySQL arguments
	Charset   string
	ParseTime bool
}

func (t *Config) SetDefaults() {
	if t.Charset == "" {
		t.Charset = "utf8"
	}

	if !t.ParseTime {
		t.ParseTime = true
	}
}

func (t *Config) Validate() error {
	connectionWithParts := !(t.Username == "" || t.Password == "" || t.Hostname == "" || t.Port == 0 || t.Name == "")
	connectionWithURL := !(t.URL == "")

	if !connectionWithParts && !connectionWithURL {
		return fmt.Errorf("no method for connection was provided")
	}

	return nil
}

func (t *Config) DSN() string {
	args := t.dbArguments()

	if t.URL == "" {
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?%s",
			t.Username,
			t.Password,
			t.Hostname,
			t.Port,
			t.Name,
			args,
		)
	}

	return fmt.Sprintf("%s?%s", t.URL, args)
}

func (t *Config) dbArguments() string {
	args := []string{
		fmt.Sprintf("charset=%s", t.Charset),
		fmt.Sprintf("parseTime=%s", strconv.FormatBool(t.ParseTime)),
	}

	return strings.Join(args, "&")
}
