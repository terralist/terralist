package metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

// NewDatabaseCollectors creates Prometheus collectors for database metrics
// Returns collectors that should be registered with the registry
func NewDatabaseCollectors(sqlDB *sql.DB) []prometheus.Collector {
	if sqlDB == nil {
		return nil
	}

	return []prometheus.Collector{
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "terralist_db_connections_active",
				Help: "Number of active database connections",
			},
			func() float64 {
				stats := sqlDB.Stats()
				return float64(stats.OpenConnections)
			},
		),
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "terralist_db_connections_idle",
				Help: "Number of idle database connections in the pool",
			},
			func() float64 {
				stats := sqlDB.Stats()
				return float64(stats.Idle)
			},
		),
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "terralist_db_connections_in_use",
				Help: "Number of database connections currently in use",
			},
			func() float64 {
				stats := sqlDB.Stats()
				return float64(stats.InUse)
			},
		),
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "terralist_db_connections_max_open",
				Help: "Maximum number of open database connections allowed",
			},
			func() float64 {
				stats := sqlDB.Stats()
				return float64(stats.MaxOpenConnections)
			},
		),
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "terralist_db_connections_wait_count_total",
				Help: "Total number of connections waited for",
			},
			func() float64 {
				stats := sqlDB.Stats()
				return float64(stats.WaitCount)
			},
		),
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "terralist_db_connections_wait_duration_seconds_total",
				Help: "Total time blocked waiting for new connections",
			},
			func() float64 {
				stats := sqlDB.Stats()
				return stats.WaitDuration.Seconds()
			},
		),
	}
}
