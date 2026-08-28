# auth/role-permission Specification

## Purpose
Define the tenant-scoped RBAC data model, domain role-permission matrix, and the
authorization service that the HTTP guards (see `auth/authorization-guards`)
depend on. It establishes how a user's role is stored per tenant (PRD §80, §81),
replacing the single global `users.role` column with multi-tenant role
assignments.

## Requirements

### Requirement: Tenant-scoped role assignments

The platform SHALL model a user's role per tenant through a dedicated
`role_assignments` table (PRD §81), so that a user may hold different roles in
different tenants (e.g. `OWNER` in tenant A and `VIEWER` in tenant B) without
cross-tenant access.

- A goose migration SHALL create `role_assignments` with columns
  `user_id` (FK to `users`), `tenant_id` (UUID), and `role`.
- A unique constraint SHALL prevent a user from holding more than one role in
  the same tenant (`user_id`, `tenant_id`).
- `role_assignments.tenant_id` SHALL be scoped explicitly; every query SHALL
  filter by `tenant_id` (PRD §4, §96).
- The single global `users.role` column SHALL be deprecated; the effective role
  for authorization SHALL come from `role_assignments`.

#### Scenario: User holds different roles in different tenants

- **WHEN** a user is assigned `OWNER` in tenant A and `VIEWER` in tenant B
- **THEN** the platform SHALL authorize the user as `OWNER` for tenant A and as
  `VIEWER` for tenant B
- **AND** SHALL NOT grant tenant-A privileges inside tenant B

#### Scenario: Duplicate role assignment rejected

- **WHEN** a role assignment is created for a `(user_id, tenant_id)` pair that
  already exists
- **THEN** the platform SHALL reject it with a conflict error

### Requirement: Role set per PRD §80

The platform SHALL define the role set as `OWNER`, `ADMIN`, `EDITOR`,
`OPERATOR`, and `VIEWER` (PRD §80), replacing the existing `USER`/`ADMIN`
enum values.

- The domain `UserRole` type SHALL declare the five constants
  `UserRoleOwner`, `UserRoleAdmin`, `UserRoleEditor`, `UserRoleOperator`,
  `UserRoleViewer`.
- Existing rows using `USER`/`ADMIN` SHALL be migrated: `USER` → `VIEWER` and
  `ADMIN` → `OWNER` on the default/current tenant of each user.
- New registrations SHALL default to the least-privilege `VIEWER` role
  (PRD §80: least-privilege).

#### Scenario: Legacy roles migrated

- **WHEN** the migration runs on a database containing `USER` and `ADMIN` roles
- **THEN** `USER` rows become `VIEWER` and `ADMIN` rows become `OWNER`
- **AND** existing `USER`/`ADMIN` enum values SHALL be removed from the type

#### Scenario: New user gets least privilege

- **WHEN** a user registers
- **THEN** a `role_assignments` row is created with role `VIEWER` for the user's
  tenant

### Requirement: Role-permission matrix

The platform SHALL define a static role-permission matrix mapping each role to
the set of permissions it grants (PRD §80).

| Role      | Permissions |
|-----------|-------------|
| `OWNER`   | `workflow:*`, `capability:*`, `binding:*`, `user:*`, `audit:*`, `tenant:*` |
| `ADMIN`   | `workflow:*`, `capability:*`, `binding:*`, `audit:*` |
| `EDITOR`  | `workflow:read`, `workflow:create`, `workflow:update`, `workflow:publish`, `workflow:simulate` |
| `OPERATOR`| `instance:read`, `instance:retry`, `instance:suspend`, `instance:resume` |
| `VIEWER`  | `workflow:read`, `capability:read`, `binding:read`, `instance:read`, `audit:read` |

- The matrix SHALL live in the domain layer as a Go map (`map[UserRole][]Permission`).
- `OWNER` SHALL be the only role granted `user:*` and `tenant:*` permissions.
- Permissions SHALL be `:<verb>` suffixed or `*` wildcards and matched
  wildcard-first (e.g. `workflow:*` matches `workflow:read`, `workflow:publish`).

#### Scenario: Role grants its matrix permissions

- **WHEN** the platform resolves permissions for `OWNER`
- **THEN** it returns every permission in the `OWNER` row of the matrix

#### Scenario: Wildcard matches a concrete permission

- **WHEN** a role holds `workflow:*` and a request asks for `workflow:publish`
- **THEN** the platform SHALL authorize the request

#### Scenario: No permission for a role

- **WHEN** a `VIEWER` requests a `user:manage` permission
- **THEN** the platform SHALL deny the request

### Requirement: Authorization service

The platform SHALL expose an authorization service that, given a user, a
tenant, and a required permission, returns whether the user is authorized.

- The service SHALL load the user's role for the tenant from
  `role_assignments` via an `IRoleAssignmentRepository`.
- The service SHALL resolve the effective permissions from the domain matrix.
- An absent role assignment SHALL resolve to an empty permission set (default
  deny).
- The service SHALL be usable by the HTTP `RequirePermission` middleware (see
  `auth/authorization-guards`) and by the capability invocation chain.

#### Scenario: Authorize granted permission

- **WHEN** an `ADMIN` user requests `workflow:publish` in their tenant
- **THEN** the service returns authorized

#### Scenario: Deny absent role assignment

- **WHEN** a user with no `role_assignments` row in a tenant requests any
  permission
- **THEN** the service returns denied

### Requirement: Repository interface for role assignments

The platform SHALL expose a `IRoleAssignmentRepository` interface with
tenant-scoped methods.

- `Assign(ctx, userID, tenantID, role)` SHALL create or replace the role
  assignment for the `(user_id, tenant_id)` pair.
- `FindRoleByUserAndTenant(ctx, userID, tenantID) (UserRole, error)` SHALL
  return the user's role for a tenant.
- `ListByTenant(ctx, tenantID) ([]RoleAssignment, error)` SHALL return all role
  assignments in a tenant.
- `Remove(ctx, userID, tenantID) error` SHALL delete a role assignment.
- Every method SHALL take an explicit `tenantID` and filter by it (PRD §4, §96).

#### Scenario: Assign and read back

- **WHEN** a role is assigned and then looked up by user and tenant
- **THEN** the platform returns the assigned role

#### Scenario: Tenant isolation on reads

- **WHEN** `ListByTenant` is called with a `tenantID`
- **THEN** the returned assignments SHALL all belong to that tenant

### Requirement: goose migration + sqlc

The RBAC schema SHALL be delivered as a goose migration (`00008_rbac.sql`) with
Up/Down blocks, and queries SHALL be sqlc-generated.

- The migration SHALL create `role_assignments`, migrate legacy roles, and alter
  the `user_role` enum type.
- `sqlc generate` SHALL regenerate the RBAC queries in
  `internal/infrastructure/db/`.

#### Scenario: Migration applies and rolls back

- **WHEN** `goose up` then `goose down` are run
- **THEN** `role_assignments` is created and then dropped cleanly, and the
  legacy role data is restored

#### Scenario: sqlc regenerates after query change

- **WHEN** `sqlc generate` runs after an RBAC query change
- **THEN** the generated Go in `internal/infrastructure/db/` is updated
