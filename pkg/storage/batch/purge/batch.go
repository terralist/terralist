package purge

import (
	"fmt"
	"terralist/pkg/storage/batch"
	"terralist/pkg/storage/resolver"
)

type BatchInput struct {
	Key string
}

type BatchOutput struct{}

type Creator struct{}

func (Creator) New(r resolver.Resolver) batch.Batch {
	return &purgeBatch{
		resolver: r,
		inputs:   []*BatchInput{},
	}
}

// purgeBatch implements a convenient way to handle multiple keys when calling
// the Purge method
type purgeBatch struct {
	resolver resolver.Resolver
	inputs   []*BatchInput
}

func (b *purgeBatch) Add(input batch.Input) batch.Batch {
	b.inputs = append(b.inputs, input.(*BatchInput))

	return b
}

func (b *purgeBatch) Commit() (batch.Output, error) {
	for _, i := range b.inputs {
		err := b.resolver.Purge(i.Key)
		if err != nil {
			return nil, fmt.Errorf("could not purge key %v: %v", i.Key, err)
		}
	}

	return &BatchOutput{}, nil
}
