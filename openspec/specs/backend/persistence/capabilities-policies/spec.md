# backend/persistence/capabilities-policies Specification

## Purpose
Provides tenant-isolated persistence of the Capability Registry, scoped capability bindings, and policies behind the `ICapabilityRepository` interface and a pgx adapter (PRD 59-64).

## Requirements

### Requirement: Capability registry
The system SHALL persist capabilities in a `capabilities` table (PRD 59).

- `capabilities.name` SHALL be unique per tenant.
- `capabilities.provider_type` SHALL be a VARCHAR in `MCP`, `INTERNAL`, `HTTP`, `FUTURE`.
- `capabilities.input_schema` and `capabilities.output_schema` SHALL be JSONB.
- `capabilities.status` SHALL be a VARCHAR in `ACTIVE`, `INACTIVE`, `DISABLED`.
- `capabilities.credential_reference` SHALL store a reference, never secrets (PRD 61).

#### Scenario: Register a capability
- **WHEN** a capability is registered for a tenant
- **THEN** a row is inserted into `capabilities` with the registry shape.

#### Scenario: Capability name uniqueness per tenant
- **WHEN** two capabilities share the same `name` for a tenant
- **THEN** the second insert SHALL be rejected.

### Requirement: Scoped capability bindings
The system SHALL persist capability bindings in a `capability_bindings` table (PRD 60).

- `capability_bindings.scope_type` SHALL be a VARCHAR in `TENANT`, `WORKFLOW`, `STATE`.
- `capability_bindings.permission` SHALL be `ALLOW` or `DENY`.
- `capability_bindings.capability_id` SHALL reference `capabilities(id)`.
- Bindings SHALL be unique per `(tenant_id, capability_id, scope_type, scope_id)`.

#### Scenario: Bind a capability to a scope
- **WHEN** a capability is bound to a tenant/workflow/state scope
- **THEN** a `capability_bindings` row is inserted with the permission.

#### Scenario: List bindings for resolution
- **WHEN** all bindings for a capability are requested
- **THEN** the most-restrictive-wins resolution inputs SHALL be returned (PRD 60).

### Requirement: Scoped policies
The system SHALL persist policies in a `policies` table (PRD 3.13, 12).

- `policies.scope_type` SHALL be in `TENANT`, `WORKFLOW`, `STATE`.
- `policies.type` SHALL name the policy kind (e.g., `timeout`, `retry`, `human_handoff`, `workflow`).
- `policies.content` SHALL be JSONB.
- Policies SHALL be unique per `(tenant_id, scope_type, scope_id, type)`.

#### Scenario: Persist a policy
- **WHEN** a policy is defined for a scope
- **THEN** a `policies` row is inserted with its JSONB content.

### Requirement: Repository interface
The system SHALL expose capability/policy persistence through an `ICapabilityRepository` interface.

- Every method SHALL require an explicit `tenantID string` (PRD 4, 96).
- Methods SHALL operate on domain entities, not sqlc rows (ADR-001).

#### Scenario: Repository methods are tenant-scoped
- **WHEN** any `ICapabilityRepository` method is called with a `tenantID`
- **THEN** the returned rows SHALL all belong to that tenant.

### Requirement: PostgreSQL adapter
The system SHALL implement `ICapabilityRepository` with a pgx adapter in `apps/api/internal/infrastructure/database/pgx_capability_repository.go`.

#### Scenario: Adapter maps rows to entities and reports not-found
- **WHEN** a query returns no row
- **THEN** the adapter SHALL return a `NOT_FOUND` DomainError.
- **WHEN** a unique violation occurs
- **THEN** the adapter SHALL return a `CONFLICT` DomainError.

### Requirement: goose migration + sqlc
The schema SHALL be a goose migration (`00006_capabilities.sql`) with Up/Down blocks, and queries SHALL be sqlc-generated.

#### Scenario: Migration applies and rolls back
- **WHEN** `goose up` and `goose down` are run
- **THEN** the three tables are created and then dropped cleanly.

#### Scenario: sqlc regenerates after query change
- **WHEN** `sqlc generate` is run after a query change
- **THEN** the generated Go in `internal/infrastructure/db/` is updated.
