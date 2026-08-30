## Purpose

Provide a tenant-scoped administration API for the current tenant's profile, memberships, and tenant roles while preserving RBAC, tenant isolation, and auditability.

## ADDED Requirements

### Requirement: Current tenant profile is managed within the authenticated tenant

The system SHALL return and update only the tenant resolved from the authenticated request context. It SHALL NOT accept a client-supplied tenant identifier to select an administration target.

#### Scenario: Owner updates the current tenant profile

- **WHEN** an authenticated Owner submits valid permitted changes to the current tenant profile
- **THEN** the system updates that tenant only
- **AND** returns the updated tenant representation
- **AND** records an audit entry for the mutation.

#### Scenario: Caller lacks tenant-update permission

- **WHEN** a caller without `tenant:update` requests a tenant profile mutation
- **THEN** the system rejects the request without changing tenant data
- **AND** does not disclose another tenant's data.

### Requirement: Membership reads are tenant-scoped

The system SHALL list memberships and their tenant roles only for the authenticated tenant, using stable pagination and safe filters where supported.

#### Scenario: Owner lists tenant memberships

- **WHEN** an authenticated caller with `user:read` requests memberships
- **THEN** the response contains only memberships assigned to the caller's tenant
- **AND** each result identifies the member and current tenant role without exposing memberships from other tenants.

#### Scenario: Cross-tenant membership identifier is requested

- **WHEN** a caller supplies an identifier that belongs to a membership in another tenant
- **THEN** the system returns the standard tenant-scoped not-found or forbidden response
- **AND** does not reveal whether that membership exists elsewhere.

### Requirement: Membership role changes preserve tenant ownership

The system SHALL permit an authorized caller to assign, replace, or remove a tenant membership role only within the authenticated tenant. The system SHALL validate the requested role and SHALL prevent an operation that would leave the tenant with no Owner.

#### Scenario: Owner replaces a member role

- **WHEN** an authenticated caller with `user:update` assigns a valid tenant role to an existing membership in the current tenant
- **THEN** the system persists the role assignment atomically
- **AND** records an audit entry with actor, tenant, member, previous role, and new role.

#### Scenario: Attempt to remove the final Owner

- **WHEN** a role replacement or membership removal would leave the tenant without an Owner
- **THEN** the system rejects the operation with a stable conflict response
- **AND** leaves the membership and role assignments unchanged
- **AND** records the rejected operation when audit policy requires it.

### Requirement: Identity administration mutations are authorized and auditable

The system SHALL enforce `tenant:*` permissions for tenant profile operations and `user:*` permissions for membership and role operations on the server. Every successful identity-administration mutation SHALL produce an existing-format audit record.

#### Scenario: Unauthorized membership mutation

- **WHEN** a caller without the required `user:update` permission attempts to change or remove a membership
- **THEN** the system rejects the request before any role-assignment write
- **AND** returns a standard authorization error.
