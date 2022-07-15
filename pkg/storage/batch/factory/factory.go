package factory

import (
	"terralist/pkg/storage/batch"
	"terralist/pkg/storage/batch/find"
	"terralist/pkg/storage/batch/purge"
	"terralist/pkg/storage/batch/store"
	"terralist/pkg/storage/resolver"
)

func NewBatch(kind batch.Kind, resolver resolver.Resolver) batch.Batch {
	var creator batch.Creator

	switch kind {
	case batch.STORE:
		creator = &store.Creator{}
	case batch.FIND:
		creator = &find.Creator{}
	case batch.PURGE:
		creator = &purge.Creator{}
	default:
		return nil
	}

	return creator.New(resolver)
}
