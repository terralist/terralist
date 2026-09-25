package memory

import (
	"context"
	"fmt"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

// maxEntries bounds the number of entries the cache holds; storing more
// evicts the least recently used ones.
const maxEntries = 10000

// Cache keeps entries in process memory, up to maxEntries of them. Entries
// past their retention are never served and are removed by a periodic sweep.
type Cache struct {
	entries *lru.Cache[string, entry]
	now     func() time.Time
}

type entry struct {
	value     []byte
	expiresAt time.Time
}

func newCache(capacity int, now func() time.Time) *Cache {
	// lru.New only fails on a size that is not positive.
	entries, _ := lru.New[string, entry](capacity)

	return &Cache{
		entries: entries,
		now:     now,
	}
}

func (c *Cache) Get(_ context.Context, key string) ([]byte, bool, error) {
	e, ok := c.entries.Get(key)
	if !ok {
		return nil, false, nil
	}

	if !c.now().Before(e.expiresAt) {
		c.entries.Remove(key)
		return nil, false, nil
	}

	value := make([]byte, len(e.value))
	copy(value, e.value)

	return value, true, nil
}

func (c *Cache) Set(_ context.Context, key string, value []byte, retention time.Duration) error {
	if retention <= 0 {
		return fmt.Errorf("the retention of %q must be positive", key)
	}

	stored := make([]byte, len(value))
	copy(stored, value)

	c.entries.Add(key, entry{
		value:     stored,
		expiresAt: c.now().Add(retention),
	})

	return nil
}

func (c *Cache) Delete(_ context.Context, key string) error {
	c.entries.Remove(key)

	return nil
}

// run sweeps expired entries at every interval.
func (c *Cache) run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.sweep()
	}
}

// sweep removes every expired entry, without counting as a use of the others.
func (c *Cache) sweep() {
	now := c.now()
	for _, key := range c.entries.Keys() {
		if e, ok := c.entries.Peek(key); ok && !now.Before(e.expiresAt) {
			c.entries.Remove(key)
		}
	}
}
