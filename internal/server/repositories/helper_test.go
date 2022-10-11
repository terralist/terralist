package repositories

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
	"terralist/pkg/database"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type mockEngine struct {
	h *gorm.DB
	m sqlmock.Sqlmock
}

func (d *mockEngine) WithMigration(_ database.Migrator) error {
	return nil
}

func (d *mockEngine) Handler() *gorm.DB {
	return d.h
}

type mockDatabaseCreator struct{}

func (c *mockDatabaseCreator) New(_ database.Configurator) (database.Engine, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	gdb, err := gorm.Open(
		postgres.New(
			postgres.Config{
				DSN:                  "sqlmock_db_0",
				DriverName:           "postgres",
				Conn:                 db,
				PreferSimpleProtocol: true,
			},
		),
		&gorm.Config{
			// Logger: logger.Default.LogMode(logger.Silent),
		},
	)

	if err != nil {
		return nil, err
	}

	return &mockEngine{
		h: gdb,
		m: mock,
	}, nil
}

// newMockDatabase creates a new sqlmock instance
func newMockDatabase() (database.Engine, sqlmock.Sqlmock, error) {
	creator := &mockDatabaseCreator{}
	engine, err := creator.New(nil)
	if err != nil {
		return nil, nil, err
	}

	mockEngine := engine.(*mockEngine)

	return mockEngine, mockEngine.m, nil
}

type QueryConstructorFunc = func(string) string

// QueryConstructor returns a function that interpolates
// the table name in all queries to avoid repetition
// of the table name in tests
func newQueryConstructor(tableName string) QueryConstructorFunc {
	return func(query string) string {
		occ := strings.Count(query, "%s")
		names := make([]any, occ)
		for i := 0; i < occ; i++ {
			names[i] = tableName
		}

		return regexp.QuoteMeta(fmt.Sprintf(query, names...))
	}
}

// newRows creates a Rows object with entity.Entity columns
// injected
func newRows(columns []string) *sqlmock.Rows {
	return sqlmock.NewRows(
		append([]string{"id", "created_at", "updated_at"}, columns...),
	)
}

var (
	// AnyTime matches with any given time
	AnyTime = &_anyTime{}

	// AnyInt matches with any given int
	AnyInt = &_anyInt{}

	// AnyString matches with any given string
	AnyString = &_anyString{}

	// AnyID matches with any given ID
	AnyID = &_anyID{}

	// Any matches everything
	Any = &_any{}
)

type _anyTime struct{}

func (a _anyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

type _anyInt struct{}

func (a _anyInt) Match(v driver.Value) bool {
	_, ok := v.(int)
	return ok
}

type _anyString struct{}

func (a _anyString) Match(v driver.Value) bool {
	_, ok := v.(string)
	return ok
}

type _anyID struct{}

func (a _anyID) Match(v driver.Value) bool {
	_, ok := v.(uuid.UUID)
	return ok
}

type _any struct{}

func (a _any) Match(v driver.Value) bool {
	return true
}
