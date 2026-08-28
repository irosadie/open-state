## Context

Epic #6 (Security & Ops) Phase 1 delivers tenant-scoped RBAC. Current auth has a global `users.role` enum (USER/ADMIN) and no permission model. PRD §80 defines the five-role set; PRD §81 requires per-tenant roles ("a user may be ADMIN in tenant A and VIEWER in tenant B").

## Goals / Non-Goals

**Goals:**
- Tenant-scoped `role_assignments` store (source of truth for a user's role per tenant).
- PRD §80 role set (OWNER/ADMIN/EDITOR/OPERATOR/VIEWER).
- Domain role-permission matrix with wildcard support.
- `RequirePermission` middleware wired to protected routes.
- `GET /api/auth/me` returns tenant role + permissions.
- RBAC audit actions + denial auditing.

**Non-Goals:**
- Role-management CRUD endpoints.
- SSO/OIDC account linking.
- Fine-grained ABAC.

## Decisions

### D1: `role_assignments` as the source of truth
A new `role_assignments(user_id, tenant_id, role)` table with a unique `(user_id, tenant_id)` constraint stores the effective role per tenant (PRD §81). The legacy `users.role` enum is retained (deprecated) for backward compatibility; authorization reads only from `role_assignments`. `role` is `VARCHAR` with Go typed constants (per `db-sqlc-schema` prohibition on PostgreSQL ENUMs).

### D2: Five-role set (PRD §80)
`UserRole` declares OWNER/ADMIN/EDITOR/OPERATOR/VIEWER. Registration writes the legacy `USER` value to the deprecated `users.role` column; the effective role defaults to `VIEWER` when no `role_assignments` row exists (least privilege). A `UserRoleLegacy` constant keeps the deprecated column valid.

### D3: Domain role-permission matrix
`domain/services/authorization_service.go` holds `RolePermissionMatrix map[UserRole][]Permission` and `HasPermission` with wildcard-first matching (e.g. `workflow:*` matches `workflow:publish`). OWNER is the only role granted `user:*`/`tenant:*`. Unknown/absent roles yield an empty set (default deny).

### D4: Application AuthorizationService
`application/services/authorization_service.go` resolves a user's role for a tenant via `IRoleAssignmentRepository`, derives permissions from the matrix, and exposes `Authorize`/`Require` (FORBIDDEN on denial).

### D5: HTTP RequirePermission middleware
`middleware/authorization.go` reads the user id (set by JWT) and the tenant header, calls `authz.Require`, returns 401 (unauthenticated/missing tenant) vs 403 (denied). Wired after auth on workflows/capabilities/bindings routes. Also audits `authorization.denied` entries.

### D6: `/me` enrichment
`GET /api/auth/me` resolves the tenant role + permissions via the AuthorizationService and returns them on `UserDTO`.
