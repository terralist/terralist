package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// StorageOperationsTotal counts storage operations by backend and status.
	StorageOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "terralist_storage_operations_total",
			Help: "Total number of storage operations",
		},
		[]string{"operation", "backend", "status"},
	)

	// StorageBytesTotal tracks total bytes transferred in storage operations.
	StorageBytesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "terralist_storage_bytes_total",
			Help: "Total bytes transferred in storage operations",
		},
		[]string{"operation", "backend"},
	)

	// StorageOperationDuration tracks the duration of storage operations.
	StorageOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "terralist_storage_operation_duration_seconds",
			Help:    "Duration of storage operations in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"operation", "backend"},
	)
)

// RecordStorageOperation records a completed storage operation with its metrics.
// operation: "upload", "download", "delete"
// backend: "s3", "azure", "gcs", "local"
// status: "success", "error"
// durationSeconds: time taken for the operation
// bytes: number of bytes transferred (use 0 if not applicable)
func RecordStorageOperation(operation, backend, status string, durationSeconds float64, bytes int64) {
	StorageOperationsTotal.WithLabelValues(operation, backend, status).Inc()
	StorageOperationDuration.WithLabelValues(operation, backend).Observe(durationSeconds)

	if bytes > 0 {
		StorageBytesTotal.WithLabelValues(operation, backend).Add(float64(bytes))
	}
}
