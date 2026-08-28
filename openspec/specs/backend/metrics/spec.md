# backend/metrics Specification

## Purpose
Define a Prometheus-compatible metrics endpoint (`GET /metrics`) and the runtime
and application metrics that enable monitoring of API and worker health (PRD
§84). It establishes metric collection, HTTP exposition, and baseline metrics
for requests, errors, durations, and audit volume.

## Requirements

### Requirement: Metrics endpoint

The platform SHALL expose a Prometheus text-format metrics endpoint at
`GET /metrics` (PRD §84).

- The endpoint SHALL be served on the API process.
- It SHALL be accessible without authentication (or via an operator-scoped
  network/basic-auth) so Prometheus can scrape it.
- The response SHALL be in Prometheus text exposition format.

#### Scenario: Prometheus scrapes metrics

- **WHEN** Prometheus requests `GET /metrics`
- **THEN** it receives a text-format metrics payload

### Requirement: HTTP request metrics

The platform SHALL expose standard RED metrics for HTTP requests.

- Metrics SHALL include request rate, error rate, and latency histogram,
  labeled by route/method/status.
- Metric names SHALL follow Prometheus conventions (e.g.
  `http_requests_total`, `http_request_duration_seconds`).

#### Scenario: Request metrics recorded

- **WHEN** requests are served
- **THEN** counters/histograms are updated with route, method, and status labels

### Requirement: Runtime metrics

The platform SHALL expose Go runtime metrics.

- Standard Go runtime/process metrics (goroutines, heap, GC, CPU, etc.) SHALL be
  exported.
- The Go collector SHALL be registered with the metrics registry.

#### Scenario: Runtime metrics exported

- **WHEN** `/metrics` is scraped
- **THEN** Go runtime and process metrics are present

### Requirement: Application metrics

The platform SHALL expose business-level metrics where valuable.

- An audit-volume counter labeled by action SHALL be incremented on each audit
  write.
- A capability-invocation counter labeled by result (success/denied) SHALL be
  incremented.
- These SHALL be optional but recommended for operational insight.

#### Scenario: Audit volume metric

- **WHEN** an audit entry is written
- **THEN** an `audit_entries_total` counter with the action label is incremented

### Requirement: Worker metrics

The platform SHALL expose worker job metrics (PRD §84).

- A counter of jobs processed labeled by type and result (success/failure) SHALL
  be exposed.
- The worker SHALL expose its metrics on a `/metrics` endpoint or via the same
  registry when co-hosted.

#### Scenario: Job metrics recorded

- **WHEN** a worker completes a job
- **THEN** the job counter is incremented with type and result labels

### Requirement: Metrics configuration

Metrics SHALL be enabled/disabled via configuration.

- An env var (e.g. `METRICS_ENABLED`) SHALL control the `/metrics` endpoint.
- The scrape path SHALL be `GET /metrics` and SHALL NOT collide with API routes.

#### Scenario: Metrics toggle

- **WHEN** `METRICS_ENABLED` is true
- **THEN** `/metrics` is served
- **WHEN** it is false
- **THEN** `/metrics` is not exposed
