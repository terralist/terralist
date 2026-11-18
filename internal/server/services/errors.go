package services

import "errors"

// ErrStorageFailure is returned when an artifact cannot be written to the
// storage resolver.
var ErrStorageFailure = errors.New("storage failure")
