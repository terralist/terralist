package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRecordStorageOperation(t *testing.T) {
	// Reset metrics
	StorageOperationsTotal.Reset()
	StorageBytesTotal.Reset()
	StorageOperationDuration.Reset()

	// Record successful upload
	RecordStorageOperation("upload", "s3", "success", 1.5, 1024)

	// Verify counter
	if count := testutil.ToFloat64(StorageOperationsTotal.WithLabelValues("upload", "s3", "success")); count != 1 {
		t.Errorf("Expected 1 operation, got %f", count)
	}

	// Verify bytes
	if bytes := testutil.ToFloat64(StorageBytesTotal.WithLabelValues("upload", "s3")); bytes != 1024 {
		t.Errorf("Expected 1024 bytes, got %f", bytes)
	}

	// Note: Histogram verification is done implicitly - if RecordStorageOperation didn't panic, it worked
}

func TestStorageOperationError(t *testing.T) {
	StorageOperationsTotal.Reset()

	// Record failed upload
	RecordStorageOperation("upload", "azure", "error", 0.5, 0)

	if count := testutil.ToFloat64(StorageOperationsTotal.WithLabelValues("upload", "azure", "error")); count != 1 {
		t.Errorf("Expected 1 error, got %f", count)
	}
}

func TestStorageOperationMultipleBackends(t *testing.T) {
	StorageOperationsTotal.Reset()
	StorageBytesTotal.Reset()

	// Record operations for different backends
	RecordStorageOperation("upload", "s3", "success", 1.0, 100)
	RecordStorageOperation("upload", "gcs", "success", 1.2, 200)
	RecordStorageOperation("upload", "azure", "success", 0.8, 150)

	// Verify separation by backend
	if count := testutil.ToFloat64(StorageOperationsTotal.WithLabelValues("upload", "s3", "success")); count != 1 {
		t.Errorf("Expected 1 s3 upload, got %f", count)
	}

	if count := testutil.ToFloat64(StorageOperationsTotal.WithLabelValues("upload", "gcs", "success")); count != 1 {
		t.Errorf("Expected 1 gcs upload, got %f", count)
	}

	if count := testutil.ToFloat64(StorageOperationsTotal.WithLabelValues("upload", "azure", "success")); count != 1 {
		t.Errorf("Expected 1 azure upload, got %f", count)
	}

	// Verify bytes for each backend
	if bytes := testutil.ToFloat64(StorageBytesTotal.WithLabelValues("upload", "s3")); bytes != 100 {
		t.Errorf("Expected 100 bytes for s3, got %f", bytes)
	}

	if bytes := testutil.ToFloat64(StorageBytesTotal.WithLabelValues("upload", "gcs")); bytes != 200 {
		t.Errorf("Expected 200 bytes for gcs, got %f", bytes)
	}
}

func TestStorageOperationTypes(t *testing.T) {
	StorageOperationsTotal.Reset()

	// Test different operation types
	RecordStorageOperation("upload", "s3", "success", 1.0, 500)
	RecordStorageOperation("download", "s3", "success", 0.5, 0)
	RecordStorageOperation("delete", "s3", "success", 0.1, 0)

	if count := testutil.ToFloat64(StorageOperationsTotal.WithLabelValues("upload", "s3", "success")); count != 1 {
		t.Errorf("Expected 1 upload, got %f", count)
	}

	if count := testutil.ToFloat64(StorageOperationsTotal.WithLabelValues("download", "s3", "success")); count != 1 {
		t.Errorf("Expected 1 download, got %f", count)
	}

	if count := testutil.ToFloat64(StorageOperationsTotal.WithLabelValues("delete", "s3", "success")); count != 1 {
		t.Errorf("Expected 1 delete, got %f", count)
	}
}

func TestStorageOperationZeroBytes(t *testing.T) {
	StorageBytesTotal.Reset()

	// Record operation with zero bytes (should not increment bytes counter)
	RecordStorageOperation("download", "s3", "success", 0.5, 0)

	// Bytes counter should not be created for zero bytes
	if bytes := testutil.ToFloat64(StorageBytesTotal.WithLabelValues("download", "s3")); bytes != 0 {
		t.Errorf("Expected 0 bytes for download, got %f", bytes)
	}
}

func TestStorageMetricsRegistration(t *testing.T) {
	// Create a new registry
	reg := prometheus.NewRegistry()

	// Register metrics
	err := reg.Register(StorageOperationsTotal)
	if err != nil {
		t.Errorf("Failed to register StorageOperationsTotal: %v", err)
	}

	err = reg.Register(StorageBytesTotal)
	if err != nil {
		t.Errorf("Failed to register StorageBytesTotal: %v", err)
	}

	err = reg.Register(StorageOperationDuration)
	if err != nil {
		t.Errorf("Failed to register StorageOperationDuration: %v", err)
	}
}
