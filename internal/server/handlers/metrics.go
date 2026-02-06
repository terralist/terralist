package handlers

import (
	"strconv"
	"time"

	"terralist/pkg/metrics"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetrics is a middleware that collects HTTP metrics.
func PrometheusMetrics(registry *prometheus.Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics for /metrics endpoint itself to avoid recursion
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		method := c.Request.Method
		path := c.FullPath()

		// If route is not registered, use the raw path
		if path == "" {
			path = c.Request.URL.Path
		}

		// Track request size
		if c.Request.ContentLength > 0 {
			metrics.HTTPRequestSize.WithLabelValues(method, path).Observe(float64(c.Request.ContentLength))
		}

		// Increment in-flight requests
		metrics.HTTPRequestsInFlight.WithLabelValues(method).Inc()
		defer metrics.HTTPRequestsInFlight.WithLabelValues(method).Dec()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// Record error metrics for 5xx responses
		if c.Writer.Status() >= 500 {
			metrics.RecordError("http", "server_error")
		}

		// Record metrics
		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
		metrics.HTTPResponseSize.WithLabelValues(method, path).Observe(float64(c.Writer.Size()))
	}
}
