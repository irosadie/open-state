## 1. Structured logging (Skill: api-feature)

- [x] 1.1 Create `internal/infrastructure/logging/logger.go` (`slog`, JSON/text, level)
- [x] 1.2 Add `LOG_FORMAT`/`LOG_LEVEL` config
- [x] 1.3 Replace `log.Printf`/`log.Fatalf` in `cmd/server`, `cmd/mcp-server`, `cmd/seed`
- [x] 1.4 Add `middleware.RequestLogger` (standard fields) and wire in `CreateApp`
- [x] 1.5 Replace worker `log.` with slog (main, runtime_summary, queue)

## 2. Distributed tracing (Skill: api-feature)

- [x] 2.1 Add OTel deps (otel, sdk, otlptracehttp, otelhttp)
- [x] 2.2 Create `infrastructure/tracing/tracer.go` (TracerProvider, no-op fallback, shutdown)
- [x] 2.3 Add `OTel` config (`OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_SERVICE_NAME`)
- [x] 2.4 Create `infrastructure/tracing/httptrace.go` (Echo server span + traceparent)
- [x] 2.5 Add DB span in `PostgresAdapter.WithTx`
- [x] 2.6 Worker tracing setup + job span
- [x] 2.7 `AuditWriter` records trace id as correlation_id

## 3. Metrics (Skill: api-feature)

- [x] 3.1 Add Prometheus deps (client_golang)
- [x] 3.2 Create `infrastructure/metrics/metrics.go` (registry + RED + runtime + app vectors)
- [x] 3.3 Add `middleware.Metrics` and wire in `CreateApp`
- [x] 3.4 Register `GET /metrics` (promhttp) behind `METRICS_ENABLED`
- [x] 3.5 Wire audit-volume metric into `AuditWriter`

## 4. Verify

- [x] 4.1 `go build ./...` + `go vet ./...` (api + worker) pass
- [x] 4.2 `go test ./...` (api) passes
- [x] 4.3 `go mod tidy` clean (api + worker)
- [x] 4.4 `gofmt` clean
