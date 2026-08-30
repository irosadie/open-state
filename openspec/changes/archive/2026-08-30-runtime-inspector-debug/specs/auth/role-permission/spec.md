## MODIFIED Requirements

### Requirement: Role-permission matrix

The platform SHALL define a static role-permission matrix mapping each role to
the set of permissions it grants (PRD §80).

| Role      | Permissions |
|-----------|-------------|
| `OWNER`   | `workflow:*`, `capability:*`, `binding:*`, `user:*`, `audit:*`, `tenant:*`, `instance:*`, `debug:*` |
| `ADMIN`   | `workflow:*`, `capability:*`, `binding:*`, `audit:*`, `instance:*`, `debug:*` |
| `EDITOR`  | `workflow:read`, `workflow:create`, `workflow:update`, `workflow:publish`, `workflow:simulate` |
| `OPERATOR`| `instance:read`, `instance:retry`, `instance:suspend`, `instance:resume`, `debug:read` |
| `VIEWER`  | `workflow:read`, `capability:read`, `binding:read`, `instance:read`, `audit:read` |

- The matrix SHALL live in the domain layer as a Go map (`map[UserRole][]Permission`).
- `OWNER` SHALL be the only role granted `user:*` and `tenant:*` permissions.
- Permissions SHALL be `:<verb>` suffixed or `*` wildcards and matched
  wildcard-first (e.g. `workflow:*` matches `workflow:read`, `workflow:publish`).
- `debug:read` SHALL be required for Runtime Inspector Debug View queries; an
  `instance:read` permission alone SHALL not grant access to trace metadata.

#### Scenario: Role grants its matrix permissions

- **WHEN** the platform resolves permissions for `OWNER`
- **THEN** it returns every permission in the `OWNER` row of the matrix

#### Scenario: Wildcard matches a concrete permission

- **WHEN** a role holds `workflow:*` and a request asks for `workflow:publish`
- **THEN** the platform SHALL authorize the request

#### Scenario: Operator can inspect a debug trace

- **WHEN** an `OPERATOR` requests a Debug View in their tenant
- **THEN** the platform SHALL authorize the request through `debug:read`

#### Scenario: No permission for a role

- **WHEN** a `VIEWER` requests a `debug:read` permission
- **THEN** the platform SHALL deny the request
