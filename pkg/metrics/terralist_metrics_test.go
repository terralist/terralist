package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRecordArtifactUpload(t *testing.T) {
	// Reset metrics
	ArtifactsUploadedTotal.Reset()
	ArtifactsTotal.Reset()

	// Record upload
	RecordArtifactUpload("module", "test-authority")

	// Check uploaded counter
	if count := testutil.ToFloat64(ArtifactsUploadedTotal.WithLabelValues("module", "test-authority")); count != 1 {
		t.Errorf("Expected ArtifactsUploadedTotal to be 1, got %f", count)
	}

	// Check total gauge
	if count := testutil.ToFloat64(ArtifactsTotal.WithLabelValues("module", "test-authority")); count != 1 {
		t.Errorf("Expected ArtifactsTotal to be 1, got %f", count)
	}

	// Record another upload
	RecordArtifactUpload("module", "test-authority")

	if count := testutil.ToFloat64(ArtifactsUploadedTotal.WithLabelValues("module", "test-authority")); count != 2 {
		t.Errorf("Expected ArtifactsUploadedTotal to be 2, got %f", count)
	}

	if count := testutil.ToFloat64(ArtifactsTotal.WithLabelValues("module", "test-authority")); count != 2 {
		t.Errorf("Expected ArtifactsTotal to be 2, got %f", count)
	}
}

func TestRecordArtifactDownload(t *testing.T) {
	// Reset metrics
	ArtifactsDownloadedTotal.Reset()

	// Record download
	RecordArtifactDownload("provider", "test-authority")

	if count := testutil.ToFloat64(ArtifactsDownloadedTotal.WithLabelValues("provider", "test-authority")); count != 1 {
		t.Errorf("Expected ArtifactsDownloadedTotal to be 1, got %f", count)
	}

	// Record multiple downloads
	RecordArtifactDownload("provider", "test-authority")
	RecordArtifactDownload("provider", "test-authority")

	if count := testutil.ToFloat64(ArtifactsDownloadedTotal.WithLabelValues("provider", "test-authority")); count != 3 {
		t.Errorf("Expected ArtifactsDownloadedTotal to be 3, got %f", count)
	}
}

func TestRecordArtifactDeletion(t *testing.T) {
	// Reset and setup
	ArtifactsDeletedTotal.Reset()
	ArtifactsTotal.Reset()

	// Set initial state
	ArtifactsTotal.WithLabelValues("module", "test-authority").Set(5)

	// Record deletion
	RecordArtifactDeletion("module", "test-authority")

	// Check deleted counter
	if count := testutil.ToFloat64(ArtifactsDeletedTotal.WithLabelValues("module", "test-authority")); count != 1 {
		t.Errorf("Expected ArtifactsDeletedTotal to be 1, got %f", count)
	}

	// Check total gauge decreased
	if count := testutil.ToFloat64(ArtifactsTotal.WithLabelValues("module", "test-authority")); count != 4 {
		t.Errorf("Expected ArtifactsTotal to be 4, got %f", count)
	}
}

func TestRecordRequest(t *testing.T) {
	// Reset metrics
	RequestsByAuthorityTotal.Reset()

	// Record different operations
	RecordRequest("authority1", "upload")
	RecordRequest("authority1", "upload")
	RecordRequest("authority1", "download")
	RecordRequest("authority2", "list")

	// Check counts
	if count := testutil.ToFloat64(RequestsByAuthorityTotal.WithLabelValues("authority1", "upload")); count != 2 {
		t.Errorf("Expected authority1/upload to be 2, got %f", count)
	}

	if count := testutil.ToFloat64(RequestsByAuthorityTotal.WithLabelValues("authority1", "download")); count != 1 {
		t.Errorf("Expected authority1/download to be 1, got %f", count)
	}

	if count := testutil.ToFloat64(RequestsByAuthorityTotal.WithLabelValues("authority2", "list")); count != 1 {
		t.Errorf("Expected authority2/list to be 1, got %f", count)
	}
}

func TestSetApiKeysCount(t *testing.T) {
	// Reset metrics
	ApiKeysTotal.Reset()

	// Set counts
	SetApiKeysCount("test-authority", "active", 5)
	SetApiKeysCount("test-authority", "expired", 2)

	// Check values
	if count := testutil.ToFloat64(ApiKeysTotal.WithLabelValues("test-authority", "active")); count != 5 {
		t.Errorf("Expected active keys to be 5, got %f", count)
	}

	if count := testutil.ToFloat64(ApiKeysTotal.WithLabelValues("test-authority", "expired")); count != 2 {
		t.Errorf("Expected expired keys to be 2, got %f", count)
	}

	// Update counts
	SetApiKeysCount("test-authority", "active", 6)

	if count := testutil.ToFloat64(ApiKeysTotal.WithLabelValues("test-authority", "active")); count != 6 {
		t.Errorf("Expected active keys to be 6, got %f", count)
	}
}

func TestMultipleAuthoritiesMetrics(t *testing.T) {
	// Reset all metrics
	ArtifactsUploadedTotal.Reset()
	ArtifactsTotal.Reset()
	RequestsByAuthorityTotal.Reset()

	// Record for multiple authorities
	RecordArtifactUpload("module", "authority-a")
	RecordArtifactUpload("module", "authority-b")
	RecordArtifactUpload("provider", "authority-a")

	RecordRequest("authority-a", "upload")
	RecordRequest("authority-b", "upload")

	// Verify separation
	if count := testutil.ToFloat64(ArtifactsUploadedTotal.WithLabelValues("module", "authority-a")); count != 1 {
		t.Errorf("Expected authority-a module uploads to be 1, got %f", count)
	}

	if count := testutil.ToFloat64(ArtifactsUploadedTotal.WithLabelValues("module", "authority-b")); count != 1 {
		t.Errorf("Expected authority-b module uploads to be 1, got %f", count)
	}

	if count := testutil.ToFloat64(ArtifactsUploadedTotal.WithLabelValues("provider", "authority-a")); count != 1 {
		t.Errorf("Expected authority-a provider uploads to be 1, got %f", count)
	}
}

func TestMetricsRegistration(t *testing.T) {
	// Create a new registry
	reg := prometheus.NewRegistry()

	// Register metrics
	err := reg.Register(ArtifactsUploadedTotal)
	if err != nil {
		t.Errorf("Failed to register ArtifactsUploadedTotal: %v", err)
	}

	err = reg.Register(ArtifactsDownloadedTotal)
	if err != nil {
		t.Errorf("Failed to register ArtifactsDownloadedTotal: %v", err)
	}

	err = reg.Register(ArtifactsDeletedTotal)
	if err != nil {
		t.Errorf("Failed to register ArtifactsDeletedTotal: %v", err)
	}

	err = reg.Register(ArtifactsTotal)
	if err != nil {
		t.Errorf("Failed to register ArtifactsTotal: %v", err)
	}

	err = reg.Register(RequestsByAuthorityTotal)
	if err != nil {
		t.Errorf("Failed to register RequestsByAuthorityTotal: %v", err)
	}

	err = reg.Register(ApiKeysTotal)
	if err != nil {
		t.Errorf("Failed to register ApiKeysTotal: %v", err)
	}
}
