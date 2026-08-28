## Why

Epic **#6 (Security & Ops)** requires tenant-scoped RBAC (PRD §80, §81). Today the platform has only a single global `users.role` column (USER/ADMIN), no permission model, and no enforcement. This change introduces the RBAC foundation: a tenant-scoped role-assignment store, the PRD §80 role set (OWNER/ADMIN/EDITOR/OPERATOR/VIEWER), a domain role-permission matrix, and an HTTP `RequirePermission` guard wired to protected routes.

## What Changes

- **NEW** — `apps/api/db/migrations/00008_rbac.sql` goose migration creating `role_assignments` (user_id, tenant_id, role) with a unique `(user_id, tenant_id)` constraint.
- **NEW** — `apps/api/db/queries/rbac.sql` sqlc query file (AssignRole, FindRoleByUserAndTenant, ListRolesByTenant, RemoveRoleAssignment).
- **MODIFIED** — Domain entity `User` role set replaced with PRD §80 roles; new `RoleAssignment` entity.
- **NEW** — Repository interface `IRoleAssignmentRepository` + pgx adapter `PgxRoleAssignmentRepository`, composed into `PostgresAdapter`.
- **NEW** — Domain role-permission matrix + authorization helpers (`domain/services/authorization_service.go`).
- **NEW** — Application `AuthorizationService` (role resolution + permission checks, default-deny).
- **NEW** — HTTP `RequirePermission` middleware wired to workflows, capabilities, and bindings routes.
- **MODIFIED** — `GET /api/auth/me` returns the tenant role + granted permissions.
- **NEW** — RBAC audit actions (role_assigned/updated/removed, authorization.denied) + denial audit in the middleware.
- Uses **`db-sqlc-schema`** and **`api-feature`** skills.

## Capabilities

### New Capabilities

- `auth/role-permission`: tenant-scoped role-assignment model + role-permission matrix.
- `auth/authorization-guards`: `RequirePermission` middleware wired to protected routes.
- `auth/rbac-audit`: audit of role mutations and authorization denials.

## Impact

- **`apps/api/db/migrations/`** — add `00008_rbac.sql`.
- **`apps/api/db/queries/`** — add `rbac.sql`.
- **`apps/api/internal/domain/entities/`** — update `user.go`, add `RoleAssignment`.
- **`apps/api/internal/domain/repositories/`** — add `role_assignment_repository.go`.
- **`apps/api/internal/domain/services/`** — add `authorization_service.go` (matrix + helpers).
- **`apps/api/internal/application/services/`** — add `authorization_service.go`.
- **`apps/api/internal/infrastructure/database/`** — add `pgx_role_assignment_repository.go`; update `postgres_adapter.go`.
- **`apps/api/internal/interfaces/http/middleware/`** — add `authorization.go`.
- **`apps/api/internal/interfaces/http/routes/`** — wire `RequirePermission`; update auth `me`.
- **No** changes to worker, shared packages.

## Non-Goals

- Role-management CRUD endpoints (assign/remove roles via UI/API) — future slice.
- SSO/OIDC account linking — separate epic change.
- Fine-grained ABAC — explicitly future (PRD §80).

## Dependencies

- Phase 1 `rbac-tenant-permissions` depends on the existing auth (JWT) and the `PostgresAdapter` composition (ADR-001).
- Related to audit slices (`audit-trail-end-to-end`) which consume the RBAC audit actions.
