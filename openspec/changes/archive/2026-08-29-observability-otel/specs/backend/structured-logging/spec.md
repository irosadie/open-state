# backend/structured-logging Specification

## Purpose

Define structured logging for the API and worker using the standard library
`log/slog`, replacing ad-hoc `log.Printf` and improving machine-parseable,
correlation-aware log output (PRD §84). It establishes a single logging entry
point and consistent log fields across the service.

## ADDED Requirements

### Requirement: Structured logger

The platform SHALL provide structured logging via `log/slog` (Go standard
library).

- A single logger SHALL be constructed once in the composition root
  (e.g. `infrastructure/logging`) and injected into services/middleware.
- The default handler SHALL emit JSON logs (suitable for production ingestion).
- A debug/text handler MAY be enabled via config (`LOG_FORMAT=text`).

#### Scenario: JSON log output

- **WHEN** the service runs with the default config
- **THEN** logs are emitted as JSON objects with key/value fields

#### Scenario: Text output for local dev

- **WHEN** `LOG_FORMAT=text` is set
- **THEN** logs are emitted in a human-readable text format

### Requirement: Standard log fields

The platform SHALL attach consistent fields to log records.

- Every request log SHALL include: `method`, `path`, `status`, `duration_ms`,
  `request_id`, `user_id` (when authenticated), and `tenant_id` (when present).
- Every worker log SHALL include the job type and relevant job context.
- A `level` (slog level) SHALL be set on each record.

#### Scenario: Request log with correlation fields

- **WHEN** a request is logged
- **THEN** the record includes method, path, status, duration, request_id, and
  (when available) user_id and tenant_id

### Requirement: Replace unstructured logging

The platform SHALL replace existing `log.Printf`/`log.Fatalf` calls in the API
with the structured logger where feasible.

- The HTTP Echo logger middleware SHALL emit structured request logs.
- Audit-failure and rate-limiter-failure logs SHALL use the structured logger.
- Startup/shutdown logs SHALL use the structured logger.

#### Scenario: Echo requests logged structurally

- **WHEN** the Echo HTTP logger runs
- **THEN** request logs carry the standard fields above instead of plain text

### Requirement: Correlation id in logs

The platform SHALL include a correlation/request id in log records (see
`backend/distributed-tracing` for propagation).

- The `request_id` set by the existing Echo `RequestID` middleware SHALL be
  attached to request log records.
- Logs within a single request SHALL share the same request_id to enable
  correlation.

#### Scenario: Shared request id

- **WHEN** multiple log records are emitted during one request
- **THEN** they all carry the same request_id

### Requirement: Worker structured logging

The platform SHALL emit structured logs from the worker process.

- Job handlers SHALL log start/completion/error with job type and id.
- The worker logger SHALL use the same structured handler as the API.

#### Scenario: Job lifecycle logged

- **WHEN** a worker processes a job
- **THEN** structured records are emitted for start, success, and failure with
  job context
