package store

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"terralist/pkg/storage/batch"
	"terralist/pkg/storage/resolver"
)

type BatchInput struct {
	URL     string
	Archive bool
}

type BatchOutput struct {
	Keys []string
}

type Creator struct{}

func (Creator) New(r resolver.Resolver) batch.Batch {
	return &storeBatch{
		resolver: r,
		inputs:   []*BatchInput{},
	}
}

// StoreBatch implements a convenient way to handle multiple URLs when calling
// the Store method
type storeBatch struct {
	resolver resolver.Resolver
	inputs   []*BatchInput
}

func (b *storeBatch) Add(input batch.Input) batch.Batch {
	b.inputs = append(b.inputs, input.(*BatchInput))

	return b
}

func (b *storeBatch) Commit() (batch.Output, error) {
	var keys []string
	var e error

	for _, i := range b.inputs {
		key, err := b.resolver.Store(i.URL, i.Archive)
		if err != nil {
			e = fmt.Errorf("could commit batch: %v", err)
		}
		keys = append(keys, key)
	}

	if len(keys) != len(b.inputs) {
		for _, k := range keys {
			if err := b.resolver.Purge(k); err != nil {
				log.Error().
					Str("Key", k).
					AnErr("Error", err).
					Msg("Error while purging an aborted batch file. Require manual clean-up.")
			}
		}

		return nil, e
	}

	return &BatchOutput{
		Keys: keys,
	}, nil
}
