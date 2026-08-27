# backend/persistence/context-memory Specification

## Purpose
Provides tenant-isolated persistence of scoped runtime context and persistent memory references behind the `IContextRepository` interface and a pgx adapter, preserving the PRD 24 distinction between workflow data and persistent memory.

## Requirements

### Requirement: Scoped context records
The system SHALL persist runtime context in a `context_records` table.

- `context_records.scope_type` SHALL be a VARCHAR in `TENANT`, `CONVERSATION`, `WORKFLOW_INSTANCE`, `STATE_INSTANCE` (PRD 23).
- `context_records.key` SHALL be unique per `(tenant_id, scope_type, scope_id)`.
- `context_records.value` SHALL be JSONB (PRD 131).
- `context_records.version` SHALL be an optimistic-lock integer (PRD 31).

#### Scenario: Upsert context for a scope
- **WHEN** a context value is written for a scope
- **THEN** it SHALL be stored keyed by `(tenant_id, scope_type, scope_id, key)`, inserting or updating with a version bump.

#### Scenario: Read full scope snapshot
- **WHEN** all context for a scope is requested
- **THEN** every `context_records` row for that scope SHALL be returned.

#### Scenario: Optimistic context update conflicts
- **WHEN** a context update targets a stale `version`
- **THEN** the update SHALL match zero rows and be reported as a conflict.

### Requirement: Persistent memory references
The system SHALL persist persistent user/customer memory in a `memory_references` table (PRD 24, 43.2).

- `memory_references.name` SHALL be unique per `(tenant_id, owner_type, owner_id)`.
- `memory_references.value` SHALL be JSONB.
- `memory_references.source_workflow_instance_id` SHALL be optional provenance and SHALL NOT cascade-delete memory when the instance is removed (PRD 24).

#### Scenario: Store persistent memory
- **WHEN** a memory reference is written for an owner
- **THEN** it SHALL be stored keyed by `(tenant_id, owner_type, owner_id, name)`.

#### Scenario: Memory survives workflow expiry/deletion
- **WHEN** a workflow instance is deleted
- **THEN** `memory_references` rows with that provenance SHALL remain (PRD 24).

#### Scenario: Resolve owner memory
- **WHEN** persistent memory for an owner is requested
- **THEN** all `memory_references` for that `(tenant_id, owner_type, owner_id)` SHALL be returned.

### Requirement: Repository interface
The system SHALL expose context/memory persistence through an `IContextRepository` interface.

- Every method SHALL require an explicit `tenantID string` (PRD 4, 96).
- Methods SHALL operate on domain entities, not sqlc rows (ADR-001).

#### Scenario: Repository methods are tenant-scoped
- **WHEN** any `IContextRepository` method is called with a `tenantID`
- **THEN** the returned rows SHALL all belong to that tenant.

### Requirement: PostgreSQL adapter
The system SHALL implement `IContextRepository` with a pgx adapter in `apps/api/internal/infrastructure/database/pgx_context_repository.go`.

#### Scenario: Adapter maps rows to entities and reports not-found
- **WHEN** a query returns no row
- **THEN** the adapter SHALL return a `NOT_FOUND` DomainError.
- **WHEN** an optimistic context update matches zero rows
- **THEN** the adapter SHALL return a `CONFLICT` DomainError.

### Requirement: goose migration + sqlc
The schema SHALL be a goose migration (`00005_context.sql`) with Up/Down blocks, and queries SHALL be sqlc-generated.

#### Scenario: Migration applies and rolls back
- **WHEN** `goose up` and `goose down` are run
- **THEN** the two tables are created and then dropped cleanly.

#### Scenario: sqlc regenerates after query change
- **WHEN** `sqlc generate` is run after a query change
- **THEN** the generated Go in `internal/infrastructure/db/` is updated.
