package file

import "errors"

var (
	ErrSystemFailure = errors.New("system failure")

	ErrInvalidArguments = errors.New("invalid arguments")

	ErrDownloadFailure   = errors.New("download failure")
	ErrDownloadInterrupt = errors.New("download interrupt")

	ErrArchiveFailure = errors.New("archive failure")
)
