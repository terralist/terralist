package find

import (
	"fmt"
	"terralist/pkg/storage/batch"
	"terralist/pkg/storage/resolver"
)

type BatchInput struct {
	Key string
}

type BatchOutput struct {
	URLs []string
}

type Creator struct{}

func (Creator) New(r resolver.Resolver) batch.Batch {
	return &findBatch{
		resolver: r,
		inputs:   []*BatchInput{},
	}
}

// FindBatch implements a convenient way to handle multiple keys when calling
// the Find method
type findBatch struct {
	resolver resolver.Resolver
	inputs   []*BatchInput
}

func (b *findBatch) Add(input batch.Input) batch.Batch {
	b.inputs = append(b.inputs, input.(*BatchInput))

	return b
}

func (b *findBatch) Commit() (batch.Output, error) {
	var urls []string

	for _, i := range b.inputs {
		url, err := b.resolver.Find(i.Key)
		if err != nil {
			return nil, fmt.Errorf("could not find key %v: %v", i.Key, err)
		}
		urls = append(urls, url)
	}

	return &BatchOutput{
		URLs: urls,
	}, nil
}
