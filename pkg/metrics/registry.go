package metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// RegistryConfig holds configuration for creating a metrics registry
type RegistryConfig struct {
	// SqlDB is optional - if provided, database connection pool metrics will be registered
	SqlDB *sql.DB
}

// NewRegistry creates and configures a new Prometheus registry with all metrics
func NewRegistry(cfg *RegistryConfig) *prometheus.Registry {
	reg := prometheus.NewRegistry()

	// Register Go runtime metrics
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Register HTTP metrics
	reg.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		HTTPRequestSize,
		HTTPResponseSize,
		HTTPRequestsInFlight,
	)

	// Register application metrics
	reg.MustRegister(
		BuildInfo,
		Uptime,
		ApplicationErrors,
	)

	// Register database metrics if SQL DB is provided
	if cfg != nil && cfg.SqlDB != nil {
		dbCollectors := NewDatabaseCollectors(cfg.SqlDB)
		for _, collector := range dbCollectors {
			reg.MustRegister(collector)
		}
	}

	return reg
}
