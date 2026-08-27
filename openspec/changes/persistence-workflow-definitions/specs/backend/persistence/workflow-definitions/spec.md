## Purpose

Provides persistent, versioned, tenant-isolated storage of workflow definitions in PostgreSQL behind the `IWorkflowRepository` interface, so workflow authoring data survives restarts (PRD 128) and remains the source of truth (ADR-001).

## ADDED Requirements

### Requirement: Workflow root table
The system SHALL persist workflow definition roots in a `workflows` table with tenant isolation.

- `workflows.id` SHALL be a UUID primary key.
- `workflows.tenant_id` SHALL be a required UUID identifying the owning tenant (PRD 4).
- `workflows.slug` SHALL be required and SHALL be unique per tenant (PRD 5).
- `workflows.status` SHALL be a VARCHAR storing one of `DRAFT`, `VALIDATING`, `VALID`, `PUBLISHED`, `ARCHIVED` (PRD 9).
- `workflows.version` SHALL be a non-negative integer used for optimistic locking (PRD 31).
- `workflows.current_version` SHALL track the current published version number.

#### Scenario: Create a draft workflow
- **WHEN** a workflow definition is created for a tenant
- **THEN** a row is inserted into `workflows` with `status='DRAFT'`, `version=0`, and a unique `(tenant_id, slug)`.

#### Scenario: Slug uniqueness within a tenant
- **WHEN** two workflows share the same `slug` for the same `tenant_id`
- **THEN** the second insert SHALL be rejected.

#### Scenario: Optimistic-lock update succeeds
- **WHEN** an update targets `workflows` with the correct expected `version`
- **THEN** the update SHALL succeed and increment `version` by 1.

#### Scenario: Optimistic-lock update conflicts
- **WHEN** an update targets `workflows` with a stale expected `version`
- **THEN** the update SHALL match zero rows and be reported as a conflict.

### Requirement: Immutable workflow versions
The system SHALL store workflow definitions as immutable versioned snapshots in a `workflow_versions` table.

- `workflow_versions.version_no` SHALL be unique per `workflow_id`.
- `workflow_versions.definition` SHALL store the full `WorkflowDefinition` as JSONB.
- Published versions SHALL NOT be modified (PRD 3.3, 9, 55); edits create a new version.

#### Scenario: Publish creates a new immutable version
- **WHEN** a workflow is published
- **THEN** a new `workflow_versions` row is created with an incremented `version_no`, and `is_current` is set on it.

#### Scenario: Query current published version
- **WHEN** the current version of a workflow is requested
- **THEN** the row with `is_current=true` SHALL be returned.

### Requirement: Relational states
The system SHALL persist workflow states relationally, scoped to a workflow version.

- `states.key` SHALL be a stable node key (e.g. `PAYMENT`) unique per `workflow_version_id`.
- `states.kind` SHALL be a VARCHAR in the set `START`, `STATE`, `DECISION`, `WAIT`, `END`, `EVENT` (PRD 14).
- `states.required_context`, `states.capabilities`, `states.policy`, and `states.position` SHALL be JSONB.

#### Scenario: States are version-scoped
- **WHEN** states for a workflow version are queried
- **THEN** only rows referencing that `workflow_version_id` SHALL be returned.

### Requirement: Relational transitions
The system SHALL persist transitions and guards relationally.

- `transitions.event` SHALL store the triggering event type.
- `transitions.priority` SHALL be an integer where lower value is evaluated first (PRD 34).
- `transitions.source_state_id` and `target_state_id` SHALL reference `states`.
- `transition_guards.logic` SHALL be `AND` or `OR` and `conditions` SHALL be JSONB (PRD 35).

#### Scenario: Transitions are version-scoped
- **WHEN** transitions for a workflow version are queried
- **THEN** only rows referencing that `workflow_version_id` SHALL be returned.

#### Scenario: Guards belong to a transition
- **WHEN** guards for a transition are queried
- **THEN** only rows referencing that `transition_id` SHALL be returned.

### Requirement: Repository interface
The system SHALL expose workflow-definition persistence through an `IWorkflowRepository` interface in `internal/domain/repositories/`.

- Every query method SHALL require an explicit `tenantID string` parameter (PRD 4, 96).
- The interface SHALL declare methods to create/find a workflow, create a version, list versions, and query states/transitions/guards.
- The interface SHALL operate on domain entities (DB-agnostic) rather than sqlc rows (ADR-001).

#### Scenario: Repository methods are tenant-scoped
- **WHEN** any `IWorkflowRepository` method is called with a `tenantID`
- **THEN** the returned rows SHALL all belong to that tenant.

### Requirement: PostgreSQL adapter
The system SHALL implement `IWorkflowRepository` with a pgx adapter backed by sqlc-generated queries in `apps/api/internal/infrastructure/database/pgx_workflow_repository.go`.

#### Scenario: Adapter maps sqlc rows to entities
- **WHEN** a repository method returns
- **THEN** sqlc rows SHALL be mapped to domain entities, with UUID columns converted to string IDs.

### Requirement: goose migration + sqlc
The schema SHALL be delivered as a goose migration (`00002_workflows.sql`) with `-- +goose Up` and `-- +goose Down` blocks, and queries SHALL be generated by sqlc (never hand-edited).

#### Scenario: Migration applies and rolls back
- **WHEN** `goose up` and then `goose down` are run
- **THEN** the five tables are created and then dropped cleanly.

#### Scenario: sqlc regenerates after query change
- **WHEN** `sqlc generate` is run after a query change
- **THEN** the generated Go in `internal/infrastructure/db/` is updated.
