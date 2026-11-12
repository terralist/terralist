package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// ArtifactsUploadedTotal counts the total number of uploaded artifacts.
	ArtifactsUploadedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "terralist_artifacts_uploaded_total",
			Help: "Total number of artifacts uploaded",
		},
		[]string{"type", "authority", "name"},
	)

	// ArtifactsDownloadedTotal counts the total number of downloaded artifacts.
	ArtifactsDownloadedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "terralist_artifacts_downloaded_total",
			Help: "Total number of artifacts downloaded",
		},
		[]string{"type", "authority", "name"},
	)

	// ArtifactsDeletedTotal counts the total number of deleted artifacts.
	ArtifactsDeletedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "terralist_artifacts_deleted_total",
			Help: "Total number of artifacts deleted",
		},
		[]string{"type", "authority", "name"},
	)

	// ArtifactsTotal tracks the current number of artifacts.
	ArtifactsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "terralist_artifacts_total",
			Help: "Current total number of artifacts",
		},
		[]string{"type", "authority", "name"},
	)

	// RequestsByAuthorityTotal counts requests by authority, operation, and artifact name.
	RequestsByAuthorityTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "terralist_requests_by_authority_total",
			Help: "Total number of requests by authority, operation type, and artifact name",
		},
		[]string{"authority", "operation", "name"},
	)

	// ApiKeysTotal tracks the number of API keys by scope and status.
	ApiKeysTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "terralist_api_keys_total",
			Help: "Current number of API keys by scope and status",
		},
		[]string{"scope", "status"},
	)
)

// RecordArtifactUpload records an artifact upload.
func RecordArtifactUpload(artifactType, authority, name string) {
	ArtifactsUploadedTotal.WithLabelValues(artifactType, authority, name).Inc()
	ArtifactsTotal.WithLabelValues(artifactType, authority, name).Inc()
}

// RecordArtifactDownload records an artifact download.
func RecordArtifactDownload(artifactType, authority, name string) {
	ArtifactsDownloadedTotal.WithLabelValues(artifactType, authority, name).Inc()
}

// RecordArtifactDeletion records an artifact deletion.
func RecordArtifactDeletion(artifactType, authority, name string) {
	ArtifactsDeletedTotal.WithLabelValues(artifactType, authority, name).Inc()
	ArtifactsTotal.WithLabelValues(artifactType, authority, name).Dec()
}

// RecordRequest records a request by authority, operation, and artifact name.
func RecordRequest(authority, operation, name string) {
	RequestsByAuthorityTotal.WithLabelValues(authority, operation, name).Inc()
}

// SetApiKeysCount sets the number of API keys for a scope with a specific status.
func SetApiKeysCount(scope, status string, count float64) {
	ApiKeysTotal.WithLabelValues(scope, status).Set(count)
}
