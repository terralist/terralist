package cache

import (
	"context"
	"time"

	"terralist/pkg/metrics"
)

// MetricsCache wraps a Cache and records metrics for all operations.
type MetricsCache struct {
	Cache   Cache
	Backend string
}

// Get reads a value and records a hit, a miss or an error.
func (m *MetricsCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	value, ok, err := m.Cache.Get(ctx, key)

	switch {
	case err != nil:
		metrics.RecordCacheOperation("get", m.Backend, "error")
	case ok:
		metrics.RecordCacheOperation("get", m.Backend, "hit")
	default:
		metrics.RecordCacheOperation("get", m.Backend, "miss")
	}

	return value, ok, err
}

// Set stores a value and records the outcome.
func (m *MetricsCache) Set(ctx context.Context, key string, value []byte, retention time.Duration) error {
	err := m.Cache.Set(ctx, key, value, retention)
	metrics.RecordCacheOperation("set", m.Backend, outcome(err))

	return err
}

// Delete removes a value and records the outcome.
func (m *MetricsCache) Delete(ctx context.Context, key string) error {
	err := m.Cache.Delete(ctx, key)
	metrics.RecordCacheOperation("delete", m.Backend, outcome(err))

	return err
}

func outcome(err error) string {
	if err != nil {
		return "error"
	}

	return "success"
}
