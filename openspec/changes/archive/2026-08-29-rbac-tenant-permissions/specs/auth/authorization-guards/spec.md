# auth/authorization-guards Specification

## Purpose

Define how the tenant-scoped RBAC model (see `auth/role-permission`) is enforced
across the HTTP API through a `RequirePermission` middleware, wired to every
protected route group (workflows, capabilities, bindings, builder), with
default-deny semantics and per-route permission requirements.

## ADDED Requirements

### Requirement: RequirePermission middleware

The platform SHALL provide a `RequirePermission` Echo middleware that checks a
required permission against the authenticated user's role for the request
tenant (PRD §80, §81).

- The middleware SHALL read the authenticated user id from the request context
  (set by the existing JWT middleware).
- The middleware SHALL read the tenant id from the `X-Tenant-ID` header.
- The middleware SHALL resolve the user's role for that tenant via the
  authorization service (see `auth/role-permission`).
- If the required permission is not granted, the middleware SHALL return a
  403 Forbidden error; if the user is unauthenticated or the tenant header is
  missing, it SHALL return 401/400 as appropriate.
- The middleware SHALL be composable with the existing `JWT` and `AuthSession`
  middleware, and SHALL run after authentication.

#### Scenario: Granted permission passes

- **WHEN** an authenticated `EDITOR` calls a route requiring `workflow:create`
  in their tenant
- **THEN** the request proceeds to the handler

#### Scenario: Missing permission is forbidden

- **WHEN** a `VIEWER` calls a route requiring `workflow:create`
- **THEN** the middleware SHALL return 403 Forbidden

#### Scenario: Missing tenant header

- **WHEN** an authenticated user calls a protected route without `X-Tenant-ID`
- **THEN** the middleware SHALL reject the request without invoking the handler

### Requirement: Role isolation across tenants

The platform SHALL enforce that a user's authorization in one tenant does not
grant access in another tenant (PRD §81).

- Permission checks SHALL always use the role assignment for the tenant in the
  request header.
- A user authorized as `OWNER` in tenant A but not assigned in tenant B SHALL
  be denied in tenant B.

#### Scenario: No cross-tenant escalation

- **WHEN** a user with `OWNER` in tenant A calls a tenant-B route
- **THEN** the middleware SHALL deny the request based on the user's tenant-B
  role (or absence thereof)

### Requirement: Per-route permission wiring

The platform SHALL wire `RequirePermission` to every protected route group with
an explicit, least-privilege permission per operation (PRD §80, §146).

| Route group | Operation | Required permission |
|-------------|-----------|---------------------|
| `/api/workflows` | GET list / get | `workflow:read` |
| `/api/workflows` | POST create | `workflow:create` |
| `/api/workflows/:id` | PATCH update | `workflow:update` |
| `/api/workflows/:id/publish` | POST publish | `workflow:publish` |
| `/api/workflows/:id/versions` | GET list versions | `workflow:read` |
| `/api/capabilities` | GET list / get | `capability:read` |
| `/api/capabilities` | POST create | `capability:create` |
| `/api/capabilities/:id` | PATCH update | `capability:update` |
| `/api/capabilities/:id` | DELETE delete | `capability:delete` |
| `/api/capabilities/:id/test` | POST test invoke | `capability:invoke` |
| `/api/capabilities/:id/bindings` | GET list | `binding:read` |
| `/api/capabilities/:id/bindings` | POST create | `binding:create` |
| `/api/bindings/:id` | DELETE delete | `binding:delete` |
| `/api/builder/...` (builder routes) | all mutating ops | per-operation (`workflow:*` family) |
| `/api/audit` | GET (Phase 2) | `audit:read` |

- The builder route group SHALL be gated with the same permission set as the
  workflow group.
- Route registration SHALL apply `RequirePermission(<permission>)` after
  authentication middleware.

#### Scenario: Mutating routes require elevated permissions

- **WHEN** a route that mutates state (e.g. publish workflow) is called with a
  read-only role
- **THEN** the middleware SHALL deny it (403)

#### Scenario: Read routes allow read roles

- **WHEN** a `VIEWER` calls a read-only route (e.g. list workflows)
- **THEN** the middleware SHALL allow it

### Requirement: Default deny

The platform SHALL treat any unlisted permission or unknown route guard as
denied by default.

- The role-permission matrix resolves to an empty set when a role is not
  assigned or unknown.
- New endpoints SHALL opt-in to a permission; there SHALL be no implicit allow.

#### Scenario: Unknown permission is denied

- **WHEN** a guard checks a permission not present in any role's matrix
- **THEN** the request SHALL be denied

### Requirement: Current user endpoint exposes roles and permissions

The platform SHALL extend the current-user response (`GET /api/auth/me`) to
expose the user's effective role and granted permissions for the request tenant.

- The `UserDTO` SHALL include a `role` and `permissions` array resolved for the
  tenant in the `X-Tenant-ID` header.
- The frontend SHALL use this to gate UI actions (see `web/audit-ui` for
  related patterns).

#### Scenario: Me returns role and permissions

- **WHEN** an authenticated `EDITOR` requests `/api/auth/me` for their tenant
- **THEN** the response includes `role: "EDITOR"` and the `EDITOR` permission
  set

#### Scenario: Me is tenant-scoped

- **WHEN** the same user requests `/api/auth/me` for a tenant where they hold a
  different role
- **THEN** the response reflects that tenant's role and permissions

### Requirement: Authorization errors map to HTTP

The platform SHALL map authorization denials to a consistent HTTP error.

- Permission denial SHALL produce a 403 with a stable machine-readable code
  (e.g. `FORBIDDEN` / `permission.denied`).
- Authentication failures SHALL remain 401 (distinct from authorization).

#### Scenario: Forbidden vs unauthorized

- **WHEN** a request is authenticated but lacks permission
- **THEN** the status is 403
- **WHEN** a request is unauthenticated
- **THEN** the status is 401
