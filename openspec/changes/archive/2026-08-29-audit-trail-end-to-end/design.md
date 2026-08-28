## Context

Epic #6 Phase 2 wires the audit trail end-to-end. `persistence-audit-adapter` already provides `audit_logs` and `IAuditRepository` (append-only, tenant-isolated). This change adds real audit writers, a query API, and a UI.

## Goals / Non-Goals

**Goals:**
- `AuditWriter` service that records important operations (best-effort).
- Filtered + paginated audit queries.
- `GET /api/audit` behind `audit:read`.
- Frontend audit page.

**Non-Goals:**
- Audit archival/retention.
- Structured logging / OTel.

## Decisions

### D1: Best-effort AuditWriter
`AuditWriter.Write(ctx, tenantID, actor, action, resourceType, resourceID, before, after, correlationID)` appends an entry and logs (not fails) on write error, so an audit failure never breaks the originating business operation (PRD §50). Actor and tenant are always derived from the authenticated context, never the body.

### D2: Actor threading
Controllers read the authenticated user id from the JWT context and pass it into service methods (`Publish`, `Bind`, `Unbind`, `TestInvoke`) so audit entries carry the real actor.

### D3: Filtered + paginated queries
`ListAuditFiltered`/`CountAuditFiltered` use optional parameters (via `sqlc.narg`) for action/resourceType/resourceId/actor/correlationId/from/to and `LIMIT/OFFSET` pagination. `IAuditRepository` exposes `ListFiltered`/`CountFiltered` with an `AuditFilter` value; `AuditService` clamps page size (max 100) and computes `hasNext`.

### D4: Query API
`AuditController.List` parses filters/page from query params and delegates to `AuditService`. Route `GET /api/audit` is gated by `RequirePermission(authz, "audit:read", audit)`.

### D5: Audit UI
`web/audit-ui` provides `/admin/audit` (server page → client content), a `useAuditList` react-query hook, Zod schema (`packages/schemas/audit.ts`), response types (`packages/types/audit-response.ts`), and constants. The page renders a table with action/actor/resource/time columns, filters, and pagination.
