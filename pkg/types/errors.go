package types

import "errors"

var (
	ErrInvalidType = errors.New("invalid type")
	ErrUnparsable  = errors.New("cannot parse")
)
