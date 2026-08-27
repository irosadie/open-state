## Purpose

Provides persistent, tenant+project-isolated storage of running workflow instances and their state instances in PostgreSQL behind the `IInstanceRepository` interface, so runtime execution survives restarts (PRD 128), is reproducible against a pinned version (PRD 58), and is concurrency-safe via optimistic locking (PRD 31).

## ADDED Requirements

### Requirement: Workflow instance root table
The system SHALL persist runtime workflow instances in a `workflow_instances` table.

- `workflow_instances.id` SHALL be a UUID primary key.
- `workflow_instances.tenant_id` SHALL be required (PRD 4).
- `workflow_instances.workflow_version_id` SHALL reference `workflow_versions(id)` and SHALL be required (PRD 58).
- `workflow_instances.status` SHALL be a VARCHAR in the set `CREATED`, `RUNNING`, `WAITING`, `COMPLETED`, `CANCELLED`, `FAILED`, `EXPIRED`, `ABORTED`, `SUSPENDED` (PRD 10, 42-43).
- `workflow_instances.version` SHALL be an optimistic-lock integer (PRD 31).

#### Scenario: Create a workflow instance
- **WHEN** an instance is created for a tenant against a workflow version
- **THEN** a row is inserted with `status='CREATED'`, `version=0`, and the pinned `workflow_version_id`.

#### Scenario: Instance status update with optimistic lock
- **WHEN** an update targets a `workflow_instances` row with the correct expected `version`
- **THEN** the update SHALL succeed and increment `version`.

#### Scenario: Instance status update with stale version
- **WHEN** an update targets a `workflow_instances` row with a stale expected `version`
- **THEN** the update SHALL match zero rows and be reported as a conflict.

### Requirement: State instance table
The system SHALL persist runtime state occurrences in a `state_instances` table.

- `state_instances.workflow_instance_id` SHALL reference `workflow_instances(id)` with `ON DELETE CASCADE`.
- `state_instances.state_key` SHALL reference a `states.key` of the pinned version.
- `state_instances.status` SHALL be a VARCHAR in the set `ENTERING`, `ACTIVE`, `WAITING`, `EXITING`, `COMPLETED`, `FAILED`, `EXPIRED`, `CANCELLED` (PRD 11).
- `state_instances.version` SHALL be an optimistic-lock integer (PRD 31).
- `state_instances.retry_count` SHALL persist the retry counter (PRD 48).
- `state_instances.entered_at`, `expires_at`, `exited_at` SHALL be present for timeout and lifecycle (PRD 25, 3.6).

#### Scenario: Create a state instance
- **WHEN** a state is entered for a workflow instance
- **THEN** a `state_instances` row is inserted with `status='ENTERING'`, `version=0`, `retry_count=0`, and `entered_at` set.

#### Scenario: Atomic transition persists new + old state with version bump
- **WHEN** a transition is executed within one transaction
- **THEN** the old `state_instances` row SHALL be updated (exited) and the new state SHALL be inserted, and the parent `workflow_instances.version` SHALL be incremented atomically (PRD 69).

#### Scenario: Retry count is persisted
- **WHEN** a state is retried
- **THEN** `retry_count` SHALL be incremented and persisted (PRD 48).

### Requirement: Current state resolution
The system SHALL track the active state instance of a workflow instance.

- `workflow_instances.current_state_instance_id` SHALL reference the current `state_instances` row and SHALL be nullable.

#### Scenario: Query current state of an instance
- **WHEN** the current state of a workflow instance is requested
- **THEN** the `state_instances` row referenced by `current_state_instance_id` SHALL be returned.

### Requirement: Repository interface
The system SHALL expose instance persistence through an `IInstanceRepository` interface.

- Every method SHALL require an explicit `tenantID string` (PRD 4, 96).
- Methods SHALL operate on domain entities, not sqlc rows (ADR-001).
- A method SHALL perform the atomic transition (state exit + state enter + parent version bump) in one transaction.

#### Scenario: Repository methods are tenant-scoped
- **WHEN** any `IInstanceRepository` method is called with a `tenantID`
- **THEN** the returned rows SHALL all belong to that tenant.

### Requirement: PostgreSQL adapter
The system SHALL implement `IInstanceRepository` with a pgx adapter in `apps/api/internal/infrastructure/database/pgx_instance_repository.go` using sqlc-generated queries.

#### Scenario: Adapter maps rows to entities and reports not-found
- **WHEN** a query returns no row
- **THEN** the adapter SHALL return a `NOT_FOUND` DomainError.
- **WHEN** an optimistic-lock update matches zero rows
- **THEN** the adapter SHALL return a `CONFLICT` DomainError.

### Requirement: goose migration + sqlc
The schema SHALL be a goose migration (`00003_workflow_instances.sql`) with Up/Down blocks (including the `ALTER TABLE` for the circular `current_state_instance_id` FK), and queries SHALL be sqlc-generated.

#### Scenario: Migration applies and rolls back
- **WHEN** `goose up` and `goose down` are run
- **THEN** the two tables are created (with the circular FK) and then dropped cleanly.

#### Scenario: sqlc regenerates after query change
- **WHEN** `sqlc generate` is run after a query change
- **THEN** the generated Go in `internal/infrastructure/db/` is updated.
