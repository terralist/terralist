package database

import (
	"fmt"
	"strings"
)

// CaseInsensitiveUniqueIndex is a unique index over columns whose values
// collide when they differ only in case. Rows holding a NULL in one of the
// columns never collide.
type CaseInsensitiveUniqueIndex struct {
	Table   string
	Name    string
	Columns []string
}

// Create creates the index, unless the table already has it.
func (i CaseInsensitiveUniqueIndex) Create(db *DB) error {
	if db.Migrator().HasIndex(i.Table, i.Name) {
		return nil
	}

	parts := i.lowered()
	if db.Name() == "mysql" {
		// MySQL takes an expression as a key part only within parentheses.
		for n, part := range parts {
			parts[n] = "(" + part + ")"
		}
	}

	return db.Exec(fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)", i.Name, i.Table, strings.Join(parts, ", "))).Error
}

// Duplicates returns the lowercased values that more than one row of the
// table holds, which keep the index from being created. A table without all
// the columns yet holds none.
func (i CaseInsensitiveUniqueIndex) Duplicates(db *DB) ([]map[string]any, error) {
	if !db.Migrator().HasTable(i.Table) {
		return nil, nil
	}

	for _, column := range i.Columns {
		if !db.Migrator().HasColumn(i.Table, column) {
			return nil, nil
		}
	}

	lowered := strings.Join(i.lowered(), ", ")
	selected := make([]string, len(i.Columns))
	notNull := make([]string, len(i.Columns))
	for n, column := range i.Columns {
		selected[n] = fmt.Sprintf("LOWER(%s) AS %s", column, column)
		notNull[n] = column + " IS NOT NULL"
	}

	var duplicates []map[string]any
	err := db.Table(i.Table).
		Select(strings.Join(selected, ", ")).
		Where(strings.Join(notNull, " AND ")).
		Group(lowered).
		Having("COUNT(*) > 1").
		Find(&duplicates).
		Error

	return duplicates, err
}

func (i CaseInsensitiveUniqueIndex) lowered() []string {
	parts := make([]string, len(i.Columns))
	for n, column := range i.Columns {
		parts[n] = "LOWER(" + column + ")"
	}

	return parts
}

// DropIndex drops an index of a table, if the table has it.
func DropIndex(db *DB, table, name string) error {
	if !db.Migrator().HasIndex(table, name) {
		return nil
	}

	// MySQL scopes index names to their table, the other databases to the
	// schema.
	if db.Name() == "mysql" {
		return db.Exec(fmt.Sprintf("DROP INDEX %s ON %s", name, table)).Error
	}

	return db.Exec(fmt.Sprintf("DROP INDEX %s", name)).Error
}
