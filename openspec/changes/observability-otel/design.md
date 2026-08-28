## Context

Epic #6 Phase 4 adds observability (PRD §84): structured logging, distributed tracing, and metrics. The service previously used `log.Printf` and Echo's plain-text logger with no traces or metrics.

## Goals / Non-Goals

**Goals:**
- `log/slog` with JSON default + standard request-log fields.
- OTel tracing (HTTP server span, DB span, worker job span, traceparent propagation).
- Correlation id (trace id) written to audit entries.
- Prometheus `/metrics` with RED + runtime + application metrics.
- Config toggles for log format/level and metrics.

**Non-Goals:**
- External observability UI/collector setup.
- Alerting rules.

## Decisions

### D1: slog logger
`infrastructure/logging.New(cfg)` returns a `*slog.Logger` (JSON default, text if `LOG_FORMAT=text`, level from `LOG_LEVEL`). `cmd/server` sets it as the global default so all packages log structurally. A dedicated `middleware.RequestLogger` emits one structured record per request with method/path/status/duration_ms/request_id/user_id/tenant_id.

### D2: OTel tracing
`infrastructure/tracing.Setup` builds a TracerProvider with an OTLP/HTTP exporter when `OTEL_EXPORTER_OTLP_ENDPOINT` is set, else a no-op (service always starts). `HTTPTrace` Echo middleware extracts `traceparent`, starts a server span, and sets the span context on the request so DB/worker spans become children. `PostgresAdapter.WithTx` starts a DB span. Worker module has its own tracing setup + per-job span.

### D3: Correlation to audit
`AuditWriter.Write` uses the caller-provided correlation id, else the active OTel `trace_id`, so audit entries link to distributed traces (PRD §50, §84).

### D4: Prometheus metrics
`infrastructure/metrics.New` registers Go runtime/process collectors and app vectors (`http_requests_total`, `http_request_duration_seconds`, `audit_entries_total`, `capability_invocations_total`, `auth_failures_total`). `middleware.Metrics` records RED per request; `GET /metrics` (promhttp) is registered when `METRICS_ENABLED=true`. `AuditWriter` increments the audit-volume metric.
