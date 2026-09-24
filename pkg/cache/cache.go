package cache

import (
	"context"
	"time"
)

// Cache stores byte values under string keys for a bounded time. Callers
// encode what they store; the cache never inspects it.
type Cache interface {
	// Get returns the value stored under the key and whether it was found.
	Get(ctx context.Context, key string) ([]byte, bool, error)

	// Set stores the value under the key and keeps it for the retention
	// duration, replacing any previous value.
	Set(ctx context.Context, key string, value []byte, retention time.Duration) error

	// Delete removes the value stored under the key, if any.
	Delete(ctx context.Context, key string) error
}
