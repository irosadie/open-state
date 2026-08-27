# capability/registry-admin Specification

## Purpose
Define the HTTP administrative contract for the Capability Registry and its bindings,
tenant-scoped and secret-safe, consumed by the admin frontend and operators.

## Requirements

### Requirement: Register capability

The platform SHALL expose an HTTP endpoint to register a capability in the registry.

#### Scenario: Create capability

- **WHEN** an authenticated operator submits a capability with name, description,
  provider type, provider id, input/output schema, status, and version
- **THEN** the platform creates it scoped to the tenant
- **AND** returns the created capability

#### Scenario: Duplicate name

- **WHEN** a capability with the same name already exists for the tenant
- **THEN** the platform rejects the request with a conflict error

### Requirement: Read and list capabilities

The platform SHALL expose HTTP endpoints to read and list capabilities for a tenant.

#### Scenario: List registry

- **WHEN** an operator lists capabilities
- **THEN** the platform returns the tenant-scoped registry, optionally filtered by
  provider type and status

#### Scenario: Read one capability

- **WHEN** an operator requests a specific capability
- **THEN** the platform returns it, or a not-found error if it does not exist for the
  tenant

### Requirement: Update capability

The platform SHALL expose an HTTP endpoint to update a capability.

#### Scenario: Update fields

- **WHEN** an operator updates description, provider, schema, status, or version
- **THEN** the platform applies the update for the tenant
- **AND** returns the updated capability

#### Scenario: Unknown capability

- **WHEN** the target capability does not exist for the tenant
- **THEN** the platform returns a not-found error

### Requirement: Delete capability

The platform SHALL expose an HTTP endpoint to delete or disable a capability.

#### Scenario: Delete capability

- **WHEN** an operator deletes a capability
- **THEN** the platform disables or removes it and its bindings for the tenant

### Requirement: Manage bindings

The platform SHALL expose HTTP endpoints to list, create, and delete capability
bindings scoped to tenant, workflow, or state.

#### Scenario: Create binding

- **WHEN** an operator binds a capability to a scope with an allow/deny permission
- **THEN** the platform stores the binding for the tenant
- **AND** rejects duplicates or invalid scopes with a conflict/validation error

#### Scenario: List and delete bindings

- **WHEN** an operator lists bindings for a capability
- **THEN** the platform returns them scoped to the tenant
- **AND** an operator may delete a binding

### Requirement: Secret safety

The platform SHALL NOT expose secret values through the admin API.

#### Scenario: No secret leakage

- **WHEN** a capability is created, read, or listed
- **THEN** only the `credential_reference` string is returned
- **AND** never the resolved credential or token
