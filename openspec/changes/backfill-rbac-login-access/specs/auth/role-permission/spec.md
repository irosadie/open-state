## MODIFIED Requirements

### Requirement: Role set per PRD §80

The platform SHALL define the role set as `OWNER`, `ADMIN`, `EDITOR`,
`OPERATOR`, and `VIEWER` (PRD §80), replacing the existing `USER`/`ADMIN`
enum values.

- The domain `UserRole` type SHALL declare the five constants
  `UserRoleOwner`, `UserRoleAdmin`, `UserRoleEditor`, `UserRoleOperator`,
  `UserRoleViewer`.
- Existing rows using `USER`/`ADMIN` SHALL receive a role assignment on the
  default/current tenant of each user: `USER` → `VIEWER` and `ADMIN` → `OWNER`.
- The legacy-role backfill SHALL NOT replace an existing role assignment for
  the same user and tenant.
- New registrations SHALL default to the least-privilege `VIEWER` role
  (PRD §80: least-privilege).

#### Scenario: Legacy roles migrated

- **WHEN** the RBAC migration runs on a database containing a legacy `USER` or
  `ADMIN` account without a role assignment in the default tenant
- **THEN** the account SHALL receive `VIEWER` for `USER` or `OWNER` for `ADMIN`
  in that tenant
- **AND** the account's authorization snapshot SHALL include the permissions
  granted by that assigned role

#### Scenario: Existing explicit role is preserved

- **WHEN** the RBAC migration runs for a legacy account that already has a role
  assignment in the default tenant
- **THEN** the platform SHALL retain the existing assigned role

#### Scenario: New user gets least privilege

- **WHEN** a user registers
- **THEN** a `role_assignments` row is created with role `VIEWER` for the user's
  tenant
