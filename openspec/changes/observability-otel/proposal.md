## Why

Epic **#6 (Security & Ops)** Phase 4 delivers observability (PRD §84). The API used ad-hoc `log.Printf`, no structured logging, no distributed tracing, and no metrics. This change adds structured logging (`log/slog`), OpenTelemetry tracing (HTTP + DB + worker), and a Prometheus metrics endpoint.

## What Changes

- **NEW** — `infrastructure/logging` (`log/slog`, JSON default, text for dev) + config (`LOG_FORMAT`, `LOG_LEVEL`).
- **MODIFIED** — Replaced `log.Printf`/`log.Fatalf` in `cmd/server`, `cmd/mcp-server`, `cmd/seed`, and worker with slog.
- **NEW** — `middleware.RequestLogger` (slog request logs with method/path/status/duration/request_id/user/tenant).
- **NEW** — `infrastructure/tracing`: OTel TracerProvider (OTLP/no-op), `HTTPTrace` Echo middleware (server span + traceparent extraction), DB span in `PostgresAdapter.WithTx`.
- **NEW** — Worker tracing (`worker/infrastructure/tracing` + job span).
- **MODIFIED** — `AuditWriter` records the OTel trace id as the audit `correlation_id`.
- **NEW** — `infrastructure/metrics`: Prometheus registry (RED + runtime + app metrics), `middleware.Metrics`, `GET /metrics` (config `METRICS_ENABLED`).
- Uses **`api-feature`** skill.

## Capabilities

### New Capabilities

- `backend/structured-logging`: slog logger + standard request-log fields.
- `backend/distributed-tracing`: OTel setup + HTTP/DB/worker spans + correlation to audit.
- `backend/metrics`: Prometheus `/metrics` with RED + runtime + app metrics.

## Impact

- **`apps/api/internal/infrastructure/`** — new `logging/`, `tracing/`, `metrics/`.
- **`apps/api/internal/interfaces/http/middleware/`** — add `request_logger.go`, `metrics.go`.
- **`apps/api/internal/infrastructure/database/postgres_adapter.go`** — DB span in `WithTx`.
- **`apps/api/internal/application/services/audit_writer.go`** — trace-id correlation + audit metric.
- **`apps/api/cmd/server/main.go`** — slog + OTel setup + metrics wiring.
- **`apps/worker/`** — OTel setup + job spans + slog.
- **`apps/api/go.mod`, `apps/worker/go.mod`** — OTel, Prometheus dependencies.

## Non-Goals

- Trace/observability UI — collector (e.g. Jaeger/Tempo) is external.
- Alerting/alert rules — external (Prometheus/Alertmanager).

## Dependencies

- Phase 1 `rbac-tenant-permissions` (user_id/tenant_id in logs, audit:read).
