## Context

Epic #3 ends with: (a) an append-only audit trail (PRD 50), and (b) the composed **PostgresAdapter** that fulfills ADR-001 — the engine talks only to repository interfaces, PostgreSQL is the primary adapter, and tenant isolation is enforced consistently (PRD 4, 96). This change is the final slice.

## Goals / Non-Goals

**Goals:**
- Schema for `audit_logs`.
- Domain entity + `IAuditRepository`.
- pgx audit adapter.
- Composed `PostgresAdapter` exposing all six repository interfaces.
- Centralized tenant-scoping helper + adapter boundary (portability seam).

**Non-Goals:**
- Audit query/admin UI (separate epic).
- HTTP/auth middleware enforcement (separate concern).
- Outbox publisher / event bus (separate).
- Any HTTP layer.

## Decisions

### D1: Append-only audit log
`audit_logs` stores PRD 50 fields: `actor`, `timestamp`, `action`, `resource`, `before` (JSONB), `after` (JSONB), `correlation_id`, plus `tenant_id`. Append-only; rows are never updated/deleted in normal operation.

### D2: Single composed PostgresAdapter
`PostgresAdapter` in `internal/infrastructure/database/postgres_adapter.go` owns the `*pgxpool.Pool` and the sqlc `*Queries`, and composes the six pgx repositories (workflow, instance, event, context, capability, audit) into a single struct with getters. The engine/application layer depends on the interfaces, not on the concrete adapter (ADR-001). Constructor `NewPostgresAdapter(pool *pgxpool.Pool) *PostgresAdapter`.

### D3: Centralized tenant scoping
A `tenant.go` helper centralizes the tenant-aware conventions used across slices (PRD 96): e.g. `TenantScope(tenantID string)` builders and a documented rule that every repository method takes an explicit `tenantID`. This is the enforcement seam — reviewers can verify no query crosses tenant boundaries. Each slice's repositories already filter by `tenant_id`; the helper documents and unifies the convention rather than duplicating it.

### D4: Adapter boundary as portability seam
All non-standard SQL (JSONB, `BIGSERIAL`, optimistic-lock `RETURNING`, `ON CONFLICT`) is encapsulated inside the pgx adapters. The domain interfaces expose only portable operations (ADR-001). `PostgresAdapter` is the only place that imports pgx/sqlc, keeping the seam clean for future MySQL/SQLite/Mongo adapters.

### D5: Audit action/actor as VARCHAR + Go constants
`audit_logs.action` uses VARCHAR + Go typed constants for the PRD 50 audit event set (e.g. `workflow.published`, `state.entered`, `transition.executed`, `guard.failed`, `capability.invoked`, `capability.denied`, `workflow.suspended`, `workflow.resumed`, `human_handoff.created`).

## Schema Outline

```
audit_logs
  id            UUID PK
  tenant_id     UUID NOT NULL
  actor         VARCHAR NOT NULL            -- user/system id
  action        VARCHAR NOT NULL            -- PRD 50 audit event set (Go constants)
  resource_type VARCHAR NOT NULL            -- workflow / instance / state / event / capability / ...
  resource_id   VARCHAR NOT NULL
  before        JSONB
  after         JSONB
  correlation_id VARCHAR
  occurred_at   TIMESTAMP NOT NULL DEFAULT NOW()
  created_at
  INDEX(tenant_id, action), INDEX(tenant_id, resource_type, resource_id), INDEX(tenant_id, occurred_at)
```

Append-only: no UPDATE/DELETE queries in normal operation.

## Risks / Trade-offs

- **Risk: audit table grows unbounded** → Mitigation: indexed by `(tenant_id, occurred_at)`; archival/purge is an ops concern (out of scope). Append-only preserves the audit trail (PRD 50).
- **Risk: single `PostgresAdapter` becomes a god object** → Mitigation: it only composes/wires existing repositories and exposes typed getters; no business logic. Keeps the engine's dependency surface small and the portability seam explicit.
- **Risk: tenant helper is convention-only, not enforced by compiler** → Mitigation: every interface signature already requires `tenantID string` (enforced by the compiler), and `tenant.go` codifies the naming/convention for review. HTTP authorization middleware (later) enforces at the boundary.

## Migration Plan

1. Branch `feature/epic3-persistence-audit-adapter`.
2. Add migration `00007_audit.sql` (Up + Down).
3. Add `db/queries/audit.sql`.
4. Run `sqlc generate`.
5. Add `AuditLog` entity + `IAuditRepository`.
6. Implement `pgx_audit_repository.go`.
7. Add `tenant.go` helper + `postgres_adapter.go` composing all six repositories.
8. Wire `NewPostgresAdapter` into `cmd/server`/dependency wiring (replacing/augmenting per-repo construction).
9. `go build ./...`, `go vet ./...`, `go test ./...`.
10. Optional smoke: `goose up`; write an audit entry; verify all repositories accessible via the adapter.
11. PR → review → merge.

**Rollback**: migration `Down` drops `audit_logs`; adapter composition is additive (existing repos unchanged); feature branch isolates changes.

## Open Questions

- Whether `PostgresAdapter` should also own transaction orchestration (a `WithTx(func(repos) error)` helper) used by the engine. Decision: yes, expose a `WithTx` method so the outbox/transition atomicity (PRD 65/69) is enforced at the adapter boundary; individual repo methods may also expose tx-aware variants.
