## Why

Epic **#3 (Data & Persistence)** requires every important operation to be auditable (PRD 50) and the persistence layer to be delivered as a single, cohesive **PostgresAdapter** behind the repository interfaces (ADR-001), with tenant isolation enforced consistently across every query (PRD 4, 96). This is the final persistence slice: the **audit log** and the **shared PostgresAdapter** that composes all six repository interfaces, centralizes tenant scoping, and defines the adapter boundary (the portability seam for future MySQL/SQLite/Mongo adapters, ADR-001).

It ties together the slices delivered in `persistence-workflow-definitions`, `persistence-runtime-instances`, `persistence-event-system`, `persistence-context-memory`, and `persistence-capabilities-policies`.

## What Changes

- **NEW** — `apps/api/db/migrations/00007_audit.sql` goose migration creating:
  - `audit_logs` — append-only audit trail (PRD 50) with actor, action, resource, before/after, correlation_id.
- **NEW** — `apps/api/db/queries/audit.sql` sqlc query file.
- **NEW** — Domain entity `AuditLog` in `apps/api/internal/domain/entities/`.
- **NEW** — Repository interface `apps/api/internal/domain/repositories/audit_repository.go` (`IAuditRepository`).
- **NEW** — pgx adapter `apps/api/internal/infrastructure/database/pgx_audit_repository.go`.
- **NEW** — Shared adapter `apps/api/internal/infrastructure/database/postgres_adapter.go` (`PostgresAdapter`) composing the six pgx repositories (workflow, instance, event, context, capability, audit) and exposing them under one port.
- **NEW** — Tenant-scoping helper `apps/api/internal/infrastructure/database/tenant.go` (centralized tenant-aware key/query convention per PRD 96).
- **REGENERATED** — sqlc-generated Go under `apps/api/internal/infrastructure/db/`.
- Uses **`db-sqlc-schema`**, **`api-feature`**, and **`api-code-review`** skills. `docs-openapi` not touched (no public endpoint).

## Capabilities

### New Capabilities

- `backend/persistence/audit`: append-only, tenant-isolated audit trail behind `IAuditRepository`.
- `backend/persistence/postgres-adapter`: the composed `PostgresAdapter` exposing all repository interfaces under one pgx-backed port, with centralized tenant scoping (ADR-001).

### Modified Capabilities

- `backend/persistence/workflow-definitions`, `workflow-instances`, `events`, `context-memory`, `capabilities-policies`: composed into `PostgresAdapter` and, where applicable, wrapped with the centralized tenant helper.

## Impact

- **`apps/api/db/migrations/`** — add `00007_audit.sql`.
- **`apps/api/db/queries/`** — add `audit.sql`.
- **`apps/api/internal/infrastructure/db/`** — sqlc-generated code regenerated.
- **`apps/api/internal/domain/entities/`** — add `audit_log.go`.
- **`apps/api/internal/domain/repositories/`** — add `audit_repository.go`.
- **`apps/api/internal/infrastructure/database/`** — add `pgx_audit_repository.go`, `postgres_adapter.go`, `tenant.go`.
- **No** changes to web, worker, shared packages, OpenAPI, docker.

## Non-Goals

- The Audit consumer/query endpoints (admin console) — separate frontend/backend epic.
- Enforcing tenant auth at the HTTP/middleware layer — repository-level scoping is delivered here (PRD 4/96); authorization middleware is a separate concern.
- Event bus / outbox publisher — separate.
- Any HTTP layer.

## Dependencies

- All five prior persistence slices (their repositories are composed here).
- Epic #3.
