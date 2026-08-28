## 1. DB schema — role_assignments (Skill: db-sqlc-schema)

- [x] 1.1 Read `.agents/settings.json` and `apps/api/sqlc.yaml`
- [x] 1.2 Create `apps/api/db/migrations/00008_rbac.sql` with `role_assignments` (id, user_id, tenant_id, role VARCHAR, created_at, updated_at) + unique `(user_id, tenant_id)`
- [x] 1.3 Add `apps/api/db/queries/rbac.sql`: AssignRole, FindRoleByUserAndTenant, ListRolesByTenant, RemoveRoleAssignment
- [x] 1.4 Run `sqlc generate` from `apps/api` — no errors; do NOT edit generated files

## 2. Domain entity (Skill: api-feature)

- [x] 2.1 Update `internal/domain/entities/user.go`: PRD §80 roles (OWNER/ADMIN/EDITOR/OPERATOR/VIEWER) + `UserRoleLegacy`; add `RoleAssignment` entity

## 3. Repository interface + pgx adapter (Skill: api-feature)

- [x] 3.1 Create `internal/domain/repositories/role_assignment_repository.go` (`IRoleAssignmentRepository`, tenant-scoped)
- [x] 3.2 Create `internal/infrastructure/database/pgx_role_assignment_repository.go`
- [x] 3.3 Compose into `PostgresAdapter` (field, getter, WithTx)

## 4. Domain permission matrix (Skill: api-feature)

- [x] 4.1 Create `internal/domain/services/authorization_service.go`: `RolePermissionMatrix`, `HasPermission` (wildcard-first), `PermissionsForRole`

## 5. Application AuthorizationService (Skill: api-feature)

- [x] 5.1 Create `internal/application/services/authorization_service.go`: `RoleForTenant`, `PermissionsForTenant`, `Authorize`, `Require`

## 6. HTTP guard (Skill: api-feature)

- [x] 6.1 Create `internal/interfaces/http/middleware/authorization.go`: `RequirePermission`
- [x] 6.2 Wire `RequirePermission` to workflows/capabilities/bindings routes
- [x] 6.3 Update `GET /api/auth/me` to return tenant role + permissions
- [x] 6.4 Audit `authorization.denied`

## 7. Verify

- [x] 7.1 `go build ./...` passes
- [x] 7.2 `go vet ./...` passes
- [x] 7.3 Tests pass (authorization matrix + service)
- [x] 7.4 `gofmt` clean on changed files
