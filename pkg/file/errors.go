package file

import "errors"

var (
	ErrSystemFailure = errors.New("system failure")

	ErrInvalidArguments = errors.New("invalid arguments")

	ErrDownloadFailure   = errors.New("download failure")
	ErrDownloadInterrupt = errors.New("download interrupt")

	// ErrSourceRefused is a source Terralist will not fetch from: an
	// unsupported scheme or getter, or a non-public address.
	ErrSourceRefused = errors.New("source refused")

	// ErrChecksumMismatch is a fetched file not matching its expected checksum.
	ErrChecksumMismatch = errors.New("checksum mismatch")

	// ErrInvalidArchive is an archive that cannot be decompressed.
	ErrInvalidArchive = errors.New("invalid archive")

	ErrArchiveFailure = errors.New("archive failure")

	// errNonPublicAddress is a connection refused by the private address guard.
	errNonPublicAddress = errors.New("refusing to connect to non-public address")
)
