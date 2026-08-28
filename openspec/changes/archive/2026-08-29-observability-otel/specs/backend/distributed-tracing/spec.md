# backend/distributed-tracing Specification

## Purpose

Define OpenTelemetry (OTel) tracing across the HTTP API, database access, and
the worker (asynq) so operations can be traced end-to-end and correlated with
logs (PRD §84). It establishes a TracerProvider, HTTP + DB + worker instrumentation,
and correlation-id propagation.

## ADDED Requirements

### Requirement: OpenTelemetry SDK setup

The platform SHALL initialize an OpenTelemetry TracerProvider in the
composition root (PRD §84).

- A `TracerProvider` SHALL be created from environment configuration
  (`OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SERVICE_NAME`, etc.).
- When no exporter endpoint is configured, the platform SHALL fall back to a
  no-op/console exporter so tracing never breaks startup in local dev.
- The SDK SHALL be shut down gracefully on process exit.

#### Scenario: OTLP exporter configured

- **WHEN** `OTEL_EXPORTER_OTLP_ENDPOINT` is set
- **THEN** spans are exported to that collector

#### Scenario: Tracing disabled gracefully

- **WHEN** no OTLP endpoint is configured
- **THEN** tracing is a no-op and the service still starts

### Requirement: HTTP tracing middleware

The platform SHALL instrument inbound HTTP requests with spans.

- An OTel HTTP middleware SHALL create a span per request with route, method,
  and status attributes.
- The span SHALL be linked to the request context so downstream spans (DB,
  services) become children.
- The incoming `traceparent` header SHALL be honored for distributed tracing.

#### Scenario: Request span hierarchy

- **WHEN** an HTTP request is traced
- **THEN** a root span is created and downstream operations are children of it

#### Scenario: Trace context propagated via headers

- **WHEN** an upstream service sends a `traceparent` header
- **THEN** the resulting trace is linked to the upstream trace

### Requirement: Database tracing

The platform SHALL instrument PostgreSQL queries with spans.

- sqlc/pgx query execution SHALL produce spans with the SQL and query duration.
- DB spans SHALL be children of the HTTP/worker span context.

#### Scenario: Query span under request

- **WHEN** a handler runs a query
- **THEN** a DB span is emitted as a child of the request span with query
  attributes

### Requirement: Worker tracing

The platform SHALL instrument asynq worker jobs with spans (PRD §84).

- Each job processing SHALL create a root span tagged with the job type and id.
- Any DB or external calls within the job SHALL be children of the job span.
- Trace context SHALL be propagated through the job payload when present.

#### Scenario: Job span hierarchy

- **WHEN** a worker processes a job
- **THEN** a job root span is created and DB spans are children

### Requirement: Correlation id propagation

The platform SHALL propagate a correlation/request id that links traces to logs.

- A `correlation_id` (or the OTel `trace_id`) SHALL be attached to request
  context and to audit entries (PRD §50).
- The `AuditWriter` SHALL record the current trace/correlation id on audit
  entries.
- HTTP responses MAY return the trace/correlation id in a header (e.g.
  `X-Request-Id` / `X-Correlation-Id`).

#### Scenario: Correlation id flows to audit

- **WHEN** an operation is traced and audited
- **THEN** the audit entry carries the correlation/trace id for linkage

### Requirement: Graceful lifecycle

Tracing SHALL be initialized once and shut down cleanly.

- The API and worker entrypoints SHALL call the OTel shutdown before exiting.

#### Scenario: Shutdown on exit

- **WHEN** the process receives a termination signal
- **THEN** pending spans are flushed before exit
