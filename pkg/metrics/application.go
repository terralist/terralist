package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	startTime = time.Now()

	// BuildInfo provides build information as labels.
	BuildInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "terralist_build_info",
			Help: "Build information (version, commit, timestamp)",
		},
		[]string{"version", "commit_hash", "build_timestamp"},
	)

	// Uptime tracks application uptime in seconds.
	Uptime = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "terralist_uptime_seconds",
			Help: "Application uptime in seconds",
		},
		func() float64 {
			return time.Since(startTime).Seconds()
		},
	)

	// ApplicationErrors counts application-level errors.
	ApplicationErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "terralist_application_errors_total",
			Help: "Total number of application errors",
		},
		[]string{"component", "severity"},
	)
)

// SetBuildInfo sets the build information.
func SetBuildInfo(version, commitHash, buildTimestamp string) {
	BuildInfo.WithLabelValues(version, commitHash, buildTimestamp).Set(1)
}

// RecordError records an application error.
func RecordError(component, severity string) {
	ApplicationErrors.WithLabelValues(component, severity).Inc()
}
