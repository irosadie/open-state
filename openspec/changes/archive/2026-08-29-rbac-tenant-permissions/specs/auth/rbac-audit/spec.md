# auth/rbac-audit Specification

## Purpose

Define how role-assignment changes and authorization denials are recorded in
the audit trail (PRD §50, §80, §81). RBAC mutations are security-sensitive and
SHALL be append-only auditable, tenant-scoped, and traceable to an actor.

## ADDED Requirements

### Requirement: Audit role-assignment mutations

The platform SHALL append an audit entry for every role-assignment mutation
(assign, update, remove) (PRD §50).

- The audit `action` SHALL use the `rbac.role_assigned`,
  `rbac.role_updated`, and `rbac.role_removed` typed constants.
- The audit `resource_type` SHALL be `role_assignment`; `resource_id` SHALL be
  the target `user_id`.
- The audit `actor` SHALL be the authenticated user performing the mutation.
- The audit `before`/`after` SHALL capture the previous and new role values as
  JSON (e.g. `{"role":"VIEWER"}` → `{"role":"EDITOR"}`).
- The audit `tenant_id` SHALL be the tenant in which the assignment changed.

#### Scenario: Role assignment is audited

- **WHEN** an `OWNER` assigns `EDITOR` to a user in a tenant
- **THEN** an `audit_logs` row is appended with action `rbac.role_assigned`,
  actor = the owner, before = prior role (or null), after = `{"role":"EDITOR"}`
- **AND** the row belongs to that tenant

#### Scenario: Role removal is audited

- **WHEN** a user's role assignment is removed
- **THEN** an audit row with action `rbac.role_removed` is appended

### Requirement: Audit authorization denials

The platform SHALL append an audit entry when an authenticated user is denied a
permission by the `RequirePermission` middleware (PRD §50, §80).

- The audit `action` SHALL be `authorization.denied`.
- The audit `resource_type` SHALL be the requested route/resource
  (e.g. `workflow`, `capability`); `resource_id` SHALL be the requested
  permission.
- The audit SHALL NOT record sensitive data (no tokens, no passwords).

#### Scenario: Denied request is audited

- **WHEN** a `VIEWER` is denied `workflow:create` on a route
- **THEN** an audit row with action `authorization.denied`, actor = the user,
  resource_type = `workflow`, resource_id = `workflow:create` is appended

### Requirement: Audit is append-only and tenant-scoped

RBAC audit entries SHALL respect the append-only and tenant-scoping rules of
the audit trail (PRD §50, §4, §96).

- Entries SHALL be written to `audit_logs` and never updated/deleted.
- Every write SHALL carry the explicit `tenant_id`.
- The actor SHALL be derived from the authenticated request, never from the
  request body.

#### Scenario: Audit row is immutable

- **WHEN** an RBAC audit entry is queried
- **THEN** its fields SHALL match the originally appended values

### Requirement: Audit constants registered

The platform SHALL register the RBAC audit actions as typed Go constants in the
audit action set (see `backend/persistence/audit-adapter`).

- Constants `rbac.role_assigned`, `rbac.role_updated`, `rbac.role_removed`, and
  `authorization.denied` SHALL be declared.
- An unknown/unregistered action SHALL be rejected at write time.

#### Scenario: Registered actions are constrained

- **WHEN** an RBAC audit entry is written
- **THEN** its `action` SHALL be one of the declared RBAC constants
