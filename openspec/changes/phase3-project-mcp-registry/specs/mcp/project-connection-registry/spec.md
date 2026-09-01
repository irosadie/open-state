## Purpose

Provide a tenant-isolated, project-owned registry for external MCP connections so
workflow dependencies can be configured without embedding endpoints or secrets.

## ADDED Requirements

### Requirement: Project-owned MCP connections

The platform SHALL allow an authorized operator to create, list, update, disable,
and delete MCP connections owned by the active project. A connection SHALL not be
visible or usable from another project, including another project in the same tenant.
Each connection SHALL have a non-empty display name and an alias unique within its
project.

#### Scenario: Create a project connection

- **WHEN** an authorized operator creates an MCP connection with a valid unique alias
- **THEN** the platform stores it in the active tenant and project
- **AND** returns its safe metadata without credentials.

#### Scenario: Alias conflict

- **WHEN** an operator creates or renames a connection to an alias already used in the project
- **THEN** the platform rejects the request with a conflict error
- **AND** leaves the existing connection unchanged.

#### Scenario: Cross-project access

- **WHEN** a request targets an MCP connection outside the active project
- **THEN** the platform returns not found and does not disclose its metadata.

### Requirement: Supported connection configuration

The platform SHALL support Streamable HTTP, legacy SSE, and STDIO transport records.
Remote transport records SHALL require a valid endpoint URL. STDIO records SHALL
identify a configured executable and arguments but SHALL not expose an unrestricted
host shell to ordinary project operators. A connection SHALL declare one of `none`,
`bearer`, or `oauth` authentication modes.

#### Scenario: Create an HTTP connection

- **WHEN** an operator submits a valid Streamable HTTP connection URL and authentication mode
- **THEN** the platform creates a remote MCP connection record
- **AND** keeps the URL separate from workflow definitions.

#### Scenario: Invalid transport configuration

- **WHEN** an operator omits a required endpoint or supplies fields incompatible with the selected transport
- **THEN** the platform rejects the request with field-level validation feedback.

### Requirement: Secret-safe authentication configuration

The platform SHALL store bearer credentials and OAuth client/token material as protected
secret references. API and UI responses SHALL expose only the authentication mode and
safe credential status; they SHALL NOT expose token values, authorization headers,
client secrets, refresh tokens, or secret-reference internals.

#### Scenario: Read a bearer-authenticated connection

- **WHEN** an operator reads a connection configured with bearer authentication
- **THEN** the platform returns that bearer authentication is configured
- **AND** does not return the bearer token or secret value.

### Requirement: Connection lifecycle and verification

The platform SHALL expose enabled and disabled connection states and a deliberate
connection-test operation. The test SHALL verify that the configured transport can
establish an MCP session without executing any provider business tool. The result
SHALL record a safe status, timestamp, and classified error when applicable.

#### Scenario: Successful connection test

- **WHEN** an operator tests an enabled reachable connection
- **THEN** the platform records and returns a successful verification result
- **AND** does not execute a provider business tool.

#### Scenario: Disabled connection test

- **WHEN** an operator tests a disabled connection
- **THEN** the platform rejects the action without opening an external connection.

### Requirement: Authorization and auditability

The platform SHALL enforce project-scoped permissions for connection management and
write a redacted audit record for create, update, disable, delete, and test actions.

#### Scenario: Unauthorized mutation

- **WHEN** a user without project MCP management permission attempts a mutation
- **THEN** the platform denies the mutation
- **AND** does not contact the configured external endpoint.
