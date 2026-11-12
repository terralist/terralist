package metrics

import (
	"fmt"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRecordArtifactUpload(t *testing.T) {
	// Reset metrics
	ArtifactsUploadedTotal.Reset()
	ArtifactsTotal.Reset()

	// Record upload
	RecordArtifactUpload("module", "test-scope", "test-module")

	// Check uploaded counter
	if count := testutil.ToFloat64(ArtifactsUploadedTotal.WithLabelValues("module", "test-scope", "test-module")); count != 1 {
		t.Errorf("Expected ArtifactsUploadedTotal to be 1, got %f", count)
	}

	// Check total gauge
	if count := testutil.ToFloat64(ArtifactsTotal.WithLabelValues("module", "test-scope", "test-module")); count != 1 {
		t.Errorf("Expected ArtifactsTotal to be 1, got %f", count)
	}

	// Record another upload
	RecordArtifactUpload("module", "test-scope", "test-module")

	if count := testutil.ToFloat64(ArtifactsUploadedTotal.WithLabelValues("module", "test-scope", "test-module")); count != 2 {
		t.Errorf("Expected ArtifactsUploadedTotal to be 2, got %f", count)
	}

	if count := testutil.ToFloat64(ArtifactsTotal.WithLabelValues("module", "test-scope", "test-module")); count != 2 {
		t.Errorf("Expected ArtifactsTotal to be 2, got %f", count)
	}
}

func TestRecordArtifactDownload(t *testing.T) {
	// Reset metrics
	ArtifactsDownloadedTotal.Reset()

	// Record download
	RecordArtifactDownload("provider", "test-scope", "test-provider")

	if count := testutil.ToFloat64(ArtifactsDownloadedTotal.WithLabelValues("provider", "test-scope", "test-provider")); count != 1 {
		t.Errorf("Expected ArtifactsDownloadedTotal to be 1, got %f", count)
	}

	// Record multiple downloads
	RecordArtifactDownload("provider", "test-scope", "test-provider")
	RecordArtifactDownload("provider", "test-scope", "test-provider")

	if count := testutil.ToFloat64(ArtifactsDownloadedTotal.WithLabelValues("provider", "test-scope", "test-provider")); count != 3 {
		t.Errorf("Expected ArtifactsDownloadedTotal to be 3, got %f", count)
	}
}

func TestRecordArtifactDeletion(t *testing.T) {
	// Reset and setup
	ArtifactsDeletedTotal.Reset()
	ArtifactsTotal.Reset()

	// Set initial state
	ArtifactsTotal.WithLabelValues("module", "test-scope", "test-module").Set(5)

	// Record deletion
	RecordArtifactDeletion("module", "test-scope", "test-module")

	// Check deleted counter
	if count := testutil.ToFloat64(ArtifactsDeletedTotal.WithLabelValues("module", "test-scope", "test-module")); count != 1 {
		t.Errorf("Expected ArtifactsDeletedTotal to be 1, got %f", count)
	}

	// Check total gauge decreased
	if count := testutil.ToFloat64(ArtifactsTotal.WithLabelValues("module", "test-scope", "test-module")); count != 4 {
		t.Errorf("Expected ArtifactsTotal to be 4, got %f", count)
	}
}

func TestRecordRequest(t *testing.T) {
	// Reset metrics
	RequestsByAuthorityTotal.Reset()

	// Record different operations
	RecordRequest("authority1", "upload", "test-module")
	RecordRequest("authority1", "upload", "test-module")
	RecordRequest("authority1", "download", "test-provider")
	RecordRequest("authority2", "list", "test-module")

	// Check counts
	if count := testutil.ToFloat64(RequestsByAuthorityTotal.WithLabelValues("authority1", "upload", "test-module")); count != 2 {
		t.Errorf("Expected authority1/upload to be 2, got %f", count)
	}

	if count := testutil.ToFloat64(RequestsByAuthorityTotal.WithLabelValues("authority1", "download", "test-provider")); count != 1 {
		t.Errorf("Expected authority1/download to be 1, got %f", count)
	}

	if count := testutil.ToFloat64(RequestsByAuthorityTotal.WithLabelValues("authority2", "list", "test-module")); count != 1 {
		t.Errorf("Expected authority2/list to be 1, got %f", count)
	}
}

func TestSetApiKeysCount(t *testing.T) {
	// Reset metrics
	ApiKeysTotal.Reset()

	// Set counts
	SetApiKeysCount("test-scope", "active", 5)
	SetApiKeysCount("test-scope", "expired", 2)

	// Check values
	if count := testutil.ToFloat64(ApiKeysTotal.WithLabelValues("test-scope", "active")); count != 5 {
		t.Errorf("Expected active keys to be 5, got %f", count)
	}

	if count := testutil.ToFloat64(ApiKeysTotal.WithLabelValues("test-scope", "expired")); count != 2 {
		t.Errorf("Expected expired keys to be 2, got %f", count)
	}

	// Update counts
	SetApiKeysCount("test-scope", "active", 6)

	if count := testutil.ToFloat64(ApiKeysTotal.WithLabelValues("test-scope", "active")); count != 6 {
		t.Errorf("Expected active keys to be 6, got %f", count)
	}
}

func TestMultipleAuthoritiesMetrics(t *testing.T) {
	// Reset all metrics
	ArtifactsUploadedTotal.Reset()
	ArtifactsTotal.Reset()
	RequestsByAuthorityTotal.Reset()

	// Record for multiple authorities
	RecordArtifactUpload("module", "authority-a", "test-module")
	RecordArtifactUpload("module", "authority-b", "test-module")
	RecordArtifactUpload("provider", "authority-a", "test-provider")

	RecordRequest("authority-a", "upload", "test-module")
	RecordRequest("authority-b", "upload", "test-module")

	// Verify separation
	if count := testutil.ToFloat64(ArtifactsUploadedTotal.WithLabelValues("module", "authority-a", "test-module")); count != 1 {
		t.Errorf("Expected authority-a module uploads to be 1, got %f", count)
	}

	if count := testutil.ToFloat64(ArtifactsUploadedTotal.WithLabelValues("module", "authority-b", "test-module")); count != 1 {
		t.Errorf("Expected authority-b module uploads to be 1, got %f", count)
	}

	if count := testutil.ToFloat64(ArtifactsUploadedTotal.WithLabelValues("provider", "authority-a", "test-provider")); count != 1 {
		t.Errorf("Expected authority-a provider uploads to be 1, got %f", count)
	}
}

func TestMetricsSeparateArtifactNames(t *testing.T) {
	for _, artifactType := range []string{"module", "provider"} {
		t.Run(artifactType, func(t *testing.T) {
			ArtifactsUploadedTotal.Reset()
			ArtifactsDownloadedTotal.Reset()
			ArtifactsDeletedTotal.Reset()
			ArtifactsTotal.Reset()
			RequestsByAuthorityTotal.Reset()
			t.Cleanup(func() {
				ArtifactsUploadedTotal.Reset()
				ArtifactsDownloadedTotal.Reset()
				ArtifactsDeletedTotal.Reset()
				ArtifactsTotal.Reset()
				RequestsByAuthorityTotal.Reset()
			})

			samples := []struct {
				authority string
				name      string
				uploads   int
				downloads int
			}{
				{authority: "authority-a", name: "first", uploads: 3, downloads: 2},
				{authority: "authority-a", name: "second", uploads: 1, downloads: 1},
				{authority: "authority-b", name: "first", uploads: 1, downloads: 1},
			}
			for _, sample := range samples {
				for range sample.uploads {
					RecordArtifactUpload(artifactType, sample.authority, sample.name)
				}
				for range sample.downloads {
					RecordArtifactDownload(artifactType, sample.authority, sample.name)
					RecordRequest(sample.authority, "download", sample.name)
				}
				RecordArtifactDeletion(artifactType, sample.authority, sample.name)
			}

			tests := []struct {
				name      string
				help      string
				kind      string
				collector prometheus.Collector
				values    []int
			}{
				{"terralist_artifacts_uploaded_total", "Total number of artifacts uploaded", "counter", ArtifactsUploadedTotal, []int{3, 1, 1}},
				{"terralist_artifacts_downloaded_total", "Total number of artifacts downloaded", "counter", ArtifactsDownloadedTotal, []int{2, 1, 1}},
				{"terralist_artifacts_deleted_total", "Total number of artifacts deleted", "counter", ArtifactsDeletedTotal, []int{1, 1, 1}},
				{"terralist_artifacts_total", "Current total number of artifacts", "gauge", ArtifactsTotal, []int{2, 0, 0}},
				{"terralist_requests_by_authority_total", "Total number of requests by authority, operation type, and artifact name", "counter", RequestsByAuthorityTotal, []int{2, 1, 1}},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					var expected strings.Builder
					fmt.Fprintf(&expected, "# HELP %s %s\n# TYPE %s %s\n", tt.name, tt.help, tt.name, tt.kind)
					for i, sample := range samples {
						if tt.collector == RequestsByAuthorityTotal {
							fmt.Fprintf(&expected, "%s{authority=%q,name=%q,operation=\"download\"} %d\n", tt.name, sample.authority, sample.name, tt.values[i])
						} else {
							fmt.Fprintf(&expected, "%s{authority=%q,name=%q,type=%q} %d\n", tt.name, sample.authority, sample.name, artifactType, tt.values[i])
						}
					}
					if err := testutil.CollectAndCompare(tt.collector, strings.NewReader(expected.String()), tt.name); err != nil {
						t.Fatal(err)
					}
				})
			}
		})
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
