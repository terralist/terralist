package repositories

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	ErrDatabaseFailure = errors.New("database failure")
	ErrNotFound        = errors.New("not found")
	ErrAlreadyExists   = errors.New("already exists")
)

// writeError reports a write rejected by a unique index as ErrAlreadyExists.
func writeError(err error) error {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return fmt.Errorf("%w: %v", ErrAlreadyExists, err)
	}

	return err
}
