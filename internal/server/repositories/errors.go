package repositories

import "errors"

var (
	ErrDatabaseFailure = errors.New("database failure")
	ErrNotFound        = errors.New("not found")
)
