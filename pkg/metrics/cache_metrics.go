package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// CacheOperationsTotal counts cache operations by backend and result.
	CacheOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "terralist_cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "backend", "result"},
	)
)

// RecordCacheOperation records a completed cache operation.
// operation: "get", "set", "delete"
// backend: "memory"
// result: "hit", "miss", "success", "error"
func RecordCacheOperation(operation, backend, result string) {
	CacheOperationsTotal.WithLabelValues(operation, backend, result).Inc()
}
