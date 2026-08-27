## 1. DB schema — audit log (Skill: db-sqlc-schema)

- [ ] 1.1 Read `.agents/settings.json` and `apps/api/sqlc.yaml` before any change
- [ ] 1.2 Create `apps/api/db/migrations/00007_audit.sql` with `audit_logs` table (id, tenant_id, actor, action, resource_type, resource_id, before JSONB, after JSONB, correlation_id, occurred_at, created_at) + indexes `(tenant_id, action)`, `(tenant_id, resource_type, resource_id)`, `(tenant_id, occurred_at)`
- [ ] 1.3 Append-only: no UPDATE/DELETE queries in normal operation
- [ ] 1.4 Verify: standard columns, snake_case, Up + Down blocks

## 2. sqlc queries + regen (Skill: db-sqlc-schema)

- [ ] 2.1 Create `apps/api/db/queries/audit.sql`: AppendAuditLog, ListAuditByTenant, ListAuditByAction, ListAuditByResource
- [ ] 2.2 Every query filters by `tenant_id` (PRD 4, 96)
- [ ] 2.3 Run `sqlc generate` from `apps/api` — no errors; do NOT edit generated files

## 3. Domain entity (Skill: api-feature)

- [ ] 3.1 Read `.agents/guides/api-entity.md`
- [ ] 3.2 Create `internal/domain/entities/audit_log.go` — `AuditLog` struct + `AuditAction` typed constants (workflow.published / state.entered / transition.executed / guard.failed / capability.invoked / capability.denied / workflow.suspended / workflow.resumed / human_handoff.created)
- [ ] 3.3 Verify: no `interface{}`/`any`; typed Go constants

## 4. Repository interface (Skill: api-feature)

- [ ] 4.1 Read `.agents/guides/api-repository.md`
- [ ] 4.2 Create `internal/domain/repositories/audit_repository.go` — `IAuditRepository`
- [ ] 4.3 All methods take explicit `ctx` + `tenantID string`
- [ ] 4.4 Methods: Append, ListByTenant, ListByAction, ListByResource
- [ ] 4.5 Operates on entities only; returns DomainError `NOT_FOUND`/`CONFLICT` where applicable

## 5. PostgreSQL adapter — audit (Skill: api-feature)

- [ ] 5.1 Read `.agents/guides/api-db-repository.md`
- [ ] 5.2 Create `internal/infrastructure/database/pgx_audit_repository.go` implementing `IAuditRepository`
- [ ] 5.3 Constructor `NewPgxAuditRepository(pool *pgxpool.Pool) repositories.IAuditRepository`

## 6. Shared PostgresAdapter + tenant helper (Skill: api-feature)

- [ ] 6.1 Create `internal/infrastructure/database/tenant.go` — centralized tenant-scoping helper/convention (PRD 96)
- [ ] 6.2 Create `internal/infrastructure/database/postgres_adapter.go` — `PostgresAdapter` composing the six pgx repositories (workflow, instance, event, context, capability, audit) with typed getters
- [ ] 6.3 Add `WithTx(func(*PostgresAdapter) error)` for atomic multi-table operations (PRD 65, 69)
- [ ] 6.4 Constructor `NewPostgresAdapter(pool *pgxpool.Pool) *PostgresAdapter`
- [ ] 6.5 Verify: `PostgresAdapter` is the only composition point importing pgx/sqlc (portability seam, ADR-001)

## 7. Wiring (Skill: api-feature)

- [ ] 7.1 Update `cmd/server` dependency wiring to construct `NewPostgresAdapter` and expose repository interfaces to application services (mirrors existing auth wiring)
- [ ] 7.2 Ensure existing auth wiring still compiles/works (no regression)

## 8. Quality gate (Skill: api-code-review)

- [ ] 8.1 Run `go build ./...`, `go vet ./...`, `go test ./...` in `apps/api`
- [ ] 8.2 Smoke: `goose up`; append audit entry; query per action/resource; verify adapter exposes all six repos; verify `WithTx` atomicity
- [ ] 8.3 `sqlc generate` idempotent
- [ ] 8.4 `api-code-review`: tenant-scoped everywhere, append-only audit, no business logic in adapter, adapter is sole pgx/sqlc import point, no edited sqlc files
- [ ] 8.5 All files end with a newline (EOF)
