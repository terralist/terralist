package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// UpstreamRequestsTotal counts requests to upstream registries by hostname,
	// operation and result.
	UpstreamRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "terralist_upstream_requests_total",
			Help: "Total number of requests to upstream registries",
		},
		[]string{"hostname", "operation", "result"},
	)

	// UpstreamFetchDuration tracks how long fetching a package from an upstream
	// registry into storage takes.
	UpstreamFetchDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "terralist_upstream_fetch_duration_seconds",
			Help:    "Duration of package fetches from upstream registries in seconds",
			Buckets: []float64{0.5, 1, 2.5, 5, 10, 30, 60, 120, 300},
		},
		[]string{"hostname"},
	)
)

// RecordUpstreamRequest records a request to an upstream registry.
// operation: "versions", "version", "package"
// result: "success", "stale", "error"
func RecordUpstreamRequest(hostname, operation, result string) {
	UpstreamRequestsTotal.WithLabelValues(hostname, operation, result).Inc()
}

func RecordUpstreamFetch(hostname string, durationSeconds float64) {
	UpstreamFetchDuration.WithLabelValues(hostname).Observe(durationSeconds)
}
