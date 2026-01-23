package storage

import (
	"time"

	"terralist/pkg/metrics"
)

// MetricsResolver wraps a Resolver and records metrics for all operations.
// It implements the Decorator pattern to add observability without modifying
// the underlying resolver implementations.
type MetricsResolver struct {
	Resolver Resolver
	Backend  string
}

// Store uploads a file and records metrics.
func (m *MetricsResolver) Store(in *StoreInput) (string, error) {
	start := time.Now()

	key, err := m.Resolver.Store(in)

	duration := time.Since(start).Seconds()
	status := "success"
	if err != nil {
		status = "error"
		metrics.RecordError("storage", "error")
	}

	metrics.RecordStorageOperation("upload", m.Backend, status, duration, in.Size)

	return key, err
}

// Find retrieves a URL for a stored file and records metrics.
func (m *MetricsResolver) Find(key string) (string, error) {
	start := time.Now()

	url, err := m.Resolver.Find(key)

	duration := time.Since(start).Seconds()
	status := "success"
	if err != nil {
		status = "error"
		metrics.RecordError("storage", "error")
	}

	metrics.RecordStorageOperation("download", m.Backend, status, duration, 0)

	return url, err
}

// Purge removes a stored file and records metrics.
func (m *MetricsResolver) Purge(key string) error {
	start := time.Now()

	err := m.Resolver.Purge(key)

	duration := time.Since(start).Seconds()
	status := "success"
	if err != nil {
		status = "error"
		metrics.RecordError("storage", "error")
	}

	metrics.RecordStorageOperation("delete", m.Backend, status, duration, 0)

	return err
}
