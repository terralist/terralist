# Monitoring and Observability

Terralist exposes Prometheus metrics to provide deep insights into registry operations, storage backend performance, and system health.

## Metrics Endpoint

Metrics are available at:

```
GET /metrics
```

This endpoint returns metrics in Prometheus format and can be scraped by Prometheus or compatible monitoring systems.

## Available Metrics

### Application Metrics

#### Build Information

```
terralist_build_info{version="...", commit="...", build_time="..."}
```

Static metadata about the running Terralist instance.

#### Uptime

```
terralist_uptime_seconds
```

Total seconds since Terralist started.

#### Errors

```
terralist_errors_total{type="..."}
```

Total count of errors by type.

---

### Artifacts Metrics

Track module and provider operations across authorities.

#### Uploads

```
terralist_artifacts_uploaded_total{type="module|provider", authority="..."}
```

Total number of uploaded artifacts.

**Example queries:**
```promql
# Upload rate per minute
rate(terralist_artifacts_uploaded_total[5m])

# Uploads by authority
sum by (authority) (terralist_artifacts_uploaded_total)
```

#### Downloads

```
terralist_artifacts_downloaded_total{type="module|provider", authority="..."}
```

Total number of downloaded artifacts.

**Example queries:**
```promql
# Most downloaded modules
topk(10, sum by (authority) (terralist_artifacts_downloaded_total{type="module"}))
```

#### Deletions

```
terralist_artifacts_deleted_total{type="module|provider", authority="..."}
```

Total number of deleted artifacts.

#### Current Artifact Count

```
terralist_artifacts_total{type="module|provider", authority="..."}
```

Current number of artifacts (gauge that increases with uploads, decreases with deletions).

**Example queries:**
```promql
# Total artifacts in registry
sum(terralist_artifacts_total)

# Artifacts per authority
sum by (authority, type) (terralist_artifacts_total)
```

---

### Request Metrics

#### Requests by Authority

```
terralist_requests_by_authority_total{authority="...", operation="upload|download|list"}
```

Total requests grouped by authority and operation type.

**Example queries:**
```promql
# Request rate per authority
rate(terralist_requests_by_authority_total[5m])

# Most active authorities
topk(5, sum by (authority) (rate(terralist_requests_by_authority_total[5m])))
```

---

### API Keys Metrics

```
terralist_api_keys_total{authority="...", status="active|expired"}
```

Current number of API keys by authority and status.

**Example queries:**
```promql
# Total active API keys
sum(terralist_api_keys_total{status="active"})

# Expired keys that need cleanup
terralist_api_keys_total{status="expired"} > 0
```

---

### Storage Backend Metrics

Monitor the performance and health of storage backends (S3, Azure, GCS, Local).

#### Operations

```
terralist_storage_operations_total{operation="upload|download|delete", backend="s3|azure|gcs|local", status="success|error"}
```

Total storage operations by type, backend, and status.

**Example queries:**
```promql
# Error rate by backend
rate(terralist_storage_operations_total{status="error"}[5m])

# Success rate percentage
sum(rate(terralist_storage_operations_total{status="success"}[5m])) 
/ 
sum(rate(terralist_storage_operations_total[5m])) * 100
```

#### Data Transfer

```
terralist_storage_bytes_total{operation="upload|download", backend="..."}
```

Total bytes transferred through storage operations.

**Example queries:**
```promql
# Upload throughput (bytes/sec)
rate(terralist_storage_bytes_total{operation="upload"}[5m])

# Total data uploaded per backend
sum by (backend) (terralist_storage_bytes_total{operation="upload"})
```

#### Operation Duration

```
terralist_storage_operation_duration_seconds{operation="...", backend="..."}
```

Histogram of storage operation durations.

**Example queries:**
```promql
# P95 upload latency
histogram_quantile(0.95, sum(rate(terralist_storage_operation_duration_seconds_bucket{operation="upload"}[5m])) by (le, backend))

# P50 download latency by backend
histogram_quantile(0.50, sum(rate(terralist_storage_operation_duration_seconds_bucket{operation="download"}[5m])) by (le, backend))

# Slow operations (>5s)
terralist_storage_operation_duration_seconds_bucket{le="5.0"} - terralist_storage_operation_duration_seconds_bucket{le="2.5"}
```

---

### HTTP Metrics

Standard HTTP metrics provided by Prometheus middleware.

#### Request Duration

```
terralist_http_request_duration_seconds{method="GET|POST|PUT|DELETE", path="...", status="200|404|500"}
```

Histogram of HTTP request durations.

#### Requests Total

```
terralist_http_requests_total{method="...", path="...", status="..."}
```

Total HTTP requests.

#### Request Size

```
terralist_http_request_size_bytes{method="...", path="..."}
```

Histogram of HTTP request body sizes.

#### Response Size

```
terralist_http_response_size_bytes{method="...", path="..."}
```

Histogram of HTTP response sizes.

**Example queries:**
```promql
# Request rate per endpoint
sum by (path) (rate(terralist_http_requests_total[5m]))

# Error rate (4xx + 5xx)
sum(rate(terralist_http_requests_total{status=~"4..|5.."}[5m]))

# P99 response time
histogram_quantile(0.99, sum(rate(terralist_http_request_duration_seconds_bucket[5m])) by (le))
```

---

### Database Metrics

Connection pool and query performance metrics.

#### Active Connections

```
terralist_database_connections_active
```

Current number of active database connections.

#### Idle Connections

```
terralist_database_connections_idle
```

Current number of idle database connections in the pool.

#### Connections in Use

```
terralist_database_connections_in_use
```

Current number of connections actively executing queries.

#### Wait Count

```
terralist_database_connections_wait_count_total
```

Total number of times a connection had to wait.

#### Wait Duration

```
terralist_database_connections_wait_duration_seconds_total
```

Total time spent waiting for connections.

**Example queries:**
```promql
# Connection pool utilization %
(terralist_database_connections_in_use / terralist_database_connections_active) * 100

# Average wait time
rate(terralist_database_connections_wait_duration_seconds_total[5m]) 
/ 
rate(terralist_database_connections_wait_count_total[5m])
```

---

## Prometheus Configuration

Add Terralist to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'terralist'
    scrape_interval: 15s
    static_configs:
      - targets: ['terralist:5758']
    metrics_path: /metrics
```

---

## Alerting Examples

> **Note:** These are example alerts to get you started. Thresholds, severity levels, and time windows should be adjusted based on your specific workload, SLA requirements, and operational experience. Start with conservative thresholds and refine them based on actual production behavior to avoid alert fatigue.

### Storage Backend Alerts

#### High Storage Error Rate

```yaml
groups:
  - name: terralist_storage
    rules:
      - alert: HighStorageErrorRate
        expr: |
          (
            sum by (backend) (rate(terralist_storage_operations_total{status="error"}[5m]))
            /
            sum by (backend) (rate(terralist_storage_operations_total[5m]))
          ) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High storage error rate on {{ $labels.backend }} ({{ $value | humanizePercentage }})"
          description: "Storage backend {{ $labels.backend }} has error rate above 5% for 5 minutes"
```

#### Slow Storage Operations

```yaml
      - alert: SlowStorageOperations
        expr: |
          histogram_quantile(0.95,
            sum by (le, backend, operation) (rate(terralist_storage_operation_duration_seconds_bucket[5m]))
          ) > 5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Slow {{ $labels.operation }} on {{ $labels.backend }} (P95: {{ $value | humanizeDuration }})"
          description: "Storage {{ $labels.operation }} P95 latency exceeds 5s on {{ $labels.backend }}"

      - alert: CriticallySlowStorageOperations
        expr: |
          histogram_quantile(0.95,
            sum by (le, backend, operation) (rate(terralist_storage_operation_duration_seconds_bucket[5m]))
          ) > 30
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Critical storage latency on {{ $labels.backend }} (P95: {{ $value | humanizeDuration }})"
          description: "Storage operations critically slow, may impact user experience"
```

#### Storage Backend Down

```yaml
      - alert: StorageBackendNoActivity
        expr: |
          (time() - max by (backend) (terralist_storage_operations_total)) > 3600
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "No storage activity on {{ $labels.backend }} for over 1 hour"
          description: "Storage backend {{ $labels.backend }} may be unreachable or experiencing issues"
```

---

### HTTP and API Alerts

#### High HTTP Error Rate

```yaml
  - name: terralist_http
    rules:
      - alert: HighHTTPErrorRate
        expr: |
          (
            sum(rate(terralist_http_requests_total{status=~"5.."}[5m]))
            /
            sum(rate(terralist_http_requests_total[5m]))
          ) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High HTTP 5xx error rate ({{ $value | humanizePercentage }})"
          description: "More than 5% of requests returning 5xx errors"

      - alert: HighHTTPClientErrorRate
        expr: |
          (
            sum(rate(terralist_http_requests_total{status=~"4.."}[5m]))
            /
            sum(rate(terralist_http_requests_total[5m]))
          ) > 0.20
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High HTTP 4xx error rate ({{ $value | humanizePercentage }})"
          description: "More than 20% of requests returning 4xx errors, check authentication"
```

#### Slow HTTP Responses

```yaml
      - alert: SlowHTTPResponses
        expr: |
          histogram_quantile(0.95,
            sum by (le, path) (rate(terralist_http_request_duration_seconds_bucket[5m]))
          ) > 5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Slow responses on {{ $labels.path }} (P95: {{ $value | humanizeDuration }})"
          description: "API endpoint {{ $labels.path }} responding slowly"
```

---

### Database Alerts

#### Connection Pool Exhaustion

```yaml
  - name: terralist_database
    rules:
      - alert: DatabaseConnectionPoolNearLimit
        expr: |
          (terralist_database_connections_in_use / terralist_database_connections_active) > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Database connection pool {{ $value | humanizePercentage }} utilized"
          description: "Connection pool usage above 80%, consider increasing pool size"

      - alert: DatabaseConnectionPoolExhausted
        expr: |
          (terralist_database_connections_in_use / terralist_database_connections_active) > 0.95
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Database connection pool nearly exhausted ({{ $value | humanizePercentage }})"
          description: "Immediate action required - connection pool at capacity"
```

#### High Connection Wait Times

```yaml
      - alert: HighDatabaseConnectionWaitTime
        expr: |
          rate(terralist_database_connections_wait_duration_seconds_total[5m])
          /
          rate(terralist_database_connections_wait_count_total[5m])
          > 0.5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High database connection wait time ({{ $value | humanizeDuration }} avg)"
          description: "Applications waiting too long for database connections"
```

---

### System Health Alerts

#### Service Down

```yaml
  - name: terralist_health
    rules:
      - alert: TerraListDown
        expr: up{job="terralist"} == 0
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Terralist service is down"
          description: "Terralist instance unreachable for 2 minutes"
```

#### High Error Count

```yaml
      - alert: HighApplicationErrorCount
        expr: |
          increase(terralist_errors_total[5m]) > 50
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High application error count ({{ $value }} in 5m)"
          description: "Application experiencing elevated error rates"
```

