## Why

Epic **#6 (Security & Ops)** Phase 2 completes the audit trail end-to-end (PRD §50). The persistence layer (`persistence-audit-adapter`) provides `audit_logs` and `IAuditRepository`, but nothing writes meaningful entries, and there is no way to query or view the trail. This change wires an `AuditWriter` service into real operations, exposes a tenant-scoped query API, and ships an audit UI.

## What Changes

- **NEW** — `AuditWriter` application service (`internal/application/services/audit_writer.go`) that appends append-only, tenant-isolated audit entries; best-effort writes.
- **MODIFIED** — `AuditAction` constants extended: `workflow.published`, `capability.invoked`, `capability.denied`, `binding.created`, `binding.deleted`, `rbac.*`, `authorization.denied`.
- **MODIFIED** — `BuilderService.Publish` audits `workflow.published`; `CapabilityService` audits capability invoke/deny and binding create/delete.
- **NEW** — Filtered audit queries `ListAuditFiltered`/`CountAuditFiltered` (action/resourceType/resourceId/actor/correlationId/from/to + pagination) + `IAuditRepository.ListFiltered`/`CountFiltered`.
- **NEW** — `GET /api/audit` endpoint (`AuditController` + `AuditService`) behind `audit:read` RBAC.
- **NEW** — Audit DTOs + pagination envelope.
- **NEW** — Frontend audit page (`/admin/audit`), `useAuditList` hook, Zod schema, response types, constants.
- **NEW** — OpenAPI docs for the audit endpoint.
- Uses **`api-feature`**, **`db-sqlc-schema`**, **`web-api-integrated`**, and **`docs-openapi`** skills.

## Capabilities

### New Capabilities

- `backend/persistence/audit-writer`: application `AuditWriter` wired to real operations.
- `backend/audit-api`: `GET /api/audit` with filters + pagination.
- `web/audit-ui`: frontend audit page + hook + schema.

### Modified Capabilities

- `backend/persistence/audit-adapter`: extended query set + repository methods.

## Impact

- **`apps/api/db/queries/audit.sql`** — add `ListAuditFiltered`/`CountAuditFiltered`.
- **`apps/api/internal/domain/entities/audit_log.go`** — extend `AuditAction` constants.
- **`apps/api/internal/domain/repositories/audit_repository.go`** — add `ListFiltered`/`CountFiltered` + `AuditFilter`.
- **`apps/api/internal/application/services/`** — add `audit_writer.go`, `audit_service.go`; modify `builder_service.go`, `capability_service.go`.
- **`apps/api/internal/infrastructure/database/pgx_audit_repository.go`** — implement filtered queries.
- **`apps/api/internal/interfaces/http/controllers/`** — add `audit_controller.go`; update capability/workflow controllers (actor).
- **`apps/api/internal/interfaces/http/routes/`** — add `RegisterAuditRoutes`.
- **Frontend**: `packages/schemas/audit.ts`, `packages/types/audit-response.ts`, `apps/web/hooks/transactions/use-audit/`, `apps/web/app/admin/audit/`, constants.
- **Docs**: `docs/openapi/` audit path + schemas + merged `docs/openapi.json`.

## Non-Goals

- Audit archival/retention/export — ops concern (out of scope).
- Full role-management UI — separate slice.
- Structured logging / OTel — separate Phase 4.

## Dependencies

- Phase 1 `rbac-tenant-permissions` (audit:read permission, actor context).
- Phase 1 audit persistence (`persistence-audit-adapter`).
