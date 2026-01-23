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
		[]string{"type", "authority"},
	)

	// ArtifactsDownloadedTotal counts the total number of downloaded artifacts.
	ArtifactsDownloadedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "terralist_artifacts_downloaded_total",
			Help: "Total number of artifacts downloaded",
		},
		[]string{"type", "authority"},
	)

	// ArtifactsDeletedTotal counts the total number of deleted artifacts.
	ArtifactsDeletedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "terralist_artifacts_deleted_total",
			Help: "Total number of artifacts deleted",
		},
		[]string{"type", "authority"},
	)

	// ArtifactsTotal tracks the current number of artifacts.
	ArtifactsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "terralist_artifacts_total",
			Help: "Current total number of artifacts",
		},
		[]string{"type", "authority"},
	)

	// RequestsByAuthorityTotal counts requests by authority and operation.
	RequestsByAuthorityTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "terralist_requests_by_authority_total",
			Help: "Total number of requests by authority and operation type",
		},
		[]string{"authority", "operation"},
	)

	// ApiKeysTotal tracks the number of API keys by authority and status.
	ApiKeysTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "terralist_api_keys_total",
			Help: "Current number of API keys by authority and status",
		},
		[]string{"authority", "status"},
	)
)

// RecordArtifactUpload records an artifact upload.
func RecordArtifactUpload(artifactType, authority string) {
	ArtifactsUploadedTotal.WithLabelValues(artifactType, authority).Inc()
	ArtifactsTotal.WithLabelValues(artifactType, authority).Inc()
}

// RecordArtifactDownload records an artifact download.
func RecordArtifactDownload(artifactType, authority string) {
	ArtifactsDownloadedTotal.WithLabelValues(artifactType, authority).Inc()
}

// RecordArtifactDeletion records an artifact deletion.
func RecordArtifactDeletion(artifactType, authority string) {
	ArtifactsDeletedTotal.WithLabelValues(artifactType, authority).Inc()
	ArtifactsTotal.WithLabelValues(artifactType, authority).Dec()
}

// RecordRequest records a request by authority and operation.
func RecordRequest(authority, operation string) {
	RequestsByAuthorityTotal.WithLabelValues(authority, operation).Inc()
}

// SetApiKeysCount sets the number of API keys for an authority with a specific status.
func SetApiKeysCount(authority, status string, count float64) {
	ApiKeysTotal.WithLabelValues(authority, status).Set(count)
}
