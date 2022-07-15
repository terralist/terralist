package batch

import (
	"terralist/pkg/storage/resolver"
)

const (
	STORE = iota
	FIND
	PURGE
)

type Kind = int

// Creator creates the batch
type Creator interface {
	New(resolver.Resolver) Batch
}
