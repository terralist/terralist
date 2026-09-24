package memory

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Cache keeps entries in process memory. Entries past their retention are
// never served and are removed by a periodic sweep.
type Cache struct {
	mu      sync.Mutex
	entries map[string]entry
	now     func() time.Time
}

type entry struct {
	value     []byte
	expiresAt time.Time
}

func (c *Cache) Get(_ context.Context, key string) ([]byte, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.entries[key]
	if !ok {
		return nil, false, nil
	}

	if !c.now().Before(e.expiresAt) {
		delete(c.entries, key)
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

	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = entry{
		value:     stored,
		expiresAt: c.now().Add(retention),
	}

	return nil
}

func (c *Cache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)

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

// sweep removes every expired entry.
func (c *Cache) sweep() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	for key, e := range c.entries {
		if !now.Before(e.expiresAt) {
			delete(c.entries, key)
		}
	}
}
