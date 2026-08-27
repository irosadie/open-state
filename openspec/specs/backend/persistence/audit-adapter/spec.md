# backend/persistence/audit-adapter Specification

## Purpose
Provides an append-only, tenant-isolated audit trail and the composed **PostgresAdapter** that exposes all six persistence repository interfaces behind one pgx-backed port with centralized tenant scoping (PRD 50, ADR-001, PRD 4/96).

## Requirements

### Requirement: Append-only audit log
The system SHALL persist an audit trail in an `audit_logs` table (PRD 50).

- `audit_logs.actor`, `action`, `resource_type`, `resource_id` SHALL be recorded per entry.
- `audit_logs.before` and `audit_logs.after` SHALL be JSONB.
- `audit_logs.correlation_id` SHALL be optional.
- Audit rows SHALL be append-only; SHALL NOT be updated/deleted during normal operation (PRD 50).

#### Scenario: Write an audit entry
- **WHEN** an important operation occurs
- **THEN** an `audit_logs` row is appended with actor, action, resource, before/after, and correlation_id.

#### Scenario: Audit trail is immutable
- **WHEN** a stored audit entry is queried
- **THEN** its fields SHALL match the originally appended values (append-only).

### Requirement: Audit action set as typed constants
The system SHALL represent audit actions with typed Go constants for the PRD 50 audit event set (e.g. `workflow.published`, `state.entered`, `transition.executed`, `guard.failed`, `capability.invoked`, `capability.denied`, `workflow.suspended`, `workflow.resumed`, `human_handoff.created`).

#### Scenario: Action values are constrained
- **WHEN** an audit entry is written
- **THEN** its `action` SHALL be one of the declared typed constants.

### Requirement: Audit repository interface
The system SHALL expose audit persistence through an `IAuditRepository` interface.

- `Append` SHALL take an explicit `tenantID string` (PRD 4, 96) and the audit entry.
- `ListByTenant`, `ListByAction`, `ListByResource` SHALL be tenant-scoped query methods.

#### Scenario: Audit methods are tenant-scoped
- **WHEN** any `IAuditRepository` query is called with a `tenantID`
- **THEN** the returned rows SHALL all belong to that tenant.

### Requirement: Composed PostgresAdapter
The system SHALL expose all persistence repository interfaces through a single `PostgresAdapter`.

- `PostgresAdapter` SHALL compose the workflow, instance, event, context, capability, and audit pgx repositories.
- `PostgresAdapter` SHALL expose typed getters (e.g. `Workflows()`, `Instances()`, `Events()`, `Context()`, `Capabilities()`, `Audit()`).
- `PostgresAdapter` SHALL expose a `WithTx` method so atomic multi-table operations (outbox emit, instance transition) run in one transaction (PRD 65, 69).

#### Scenario: All repositories reachable via the adapter
- **WHEN** an engine depends on `PostgresAdapter`
- **THEN** it SHALL access each repository through the adapter's getters.

#### Scenario: Atomic transaction helper
- **WHEN** a multi-repository operation requires atomicity
- **THEN** it SHALL run within `PostgresAdapter.WithTx` and commit/roll back together.

### Requirement: Centralized tenant scoping
The system SHALL centralize tenant-scoping conventions in a helper (`internal/infrastructure/database/tenant.go`).

- Every repository interface method SHALL require an explicit `tenantID string` (PRD 4, 96) — enforced by the compiler at the signature level.
- The helper SHALL document/unify the tenant-aware conventions and be used to build tenant-scoped queries.

#### Scenario: Cross-tenant access is impossible at the data layer
- **WHEN** a repository method is called with a tenant id
- **THEN** the SQL SHALL filter by that `tenant_id` on every query (PRD 4, 96).

### Requirement: PostgreSQL adapter boundary (portability seam)
The system SHALL encapsulate all PostgreSQL-specific SQL inside the pgx adapters.

- Non-standard SQL (JSONB, `BIGSERIAL`, `RETURNING`, `ON CONFLICT`, optimistic locks) SHALL NOT leak into domain interfaces (ADR-001).
- `PostgresAdapter` SHALL be the only composition point importing pgx/sqlc.

#### Scenario: Domain stays DB-agnostic
- **WHEN** a future MySQL/SQLite/Mongo adapter is added
- **THEN** the domain/application layer SHALL not need changes because it depends only on repository interfaces (ADR-001).

### Requirement: goose migration + sqlc
The schema SHALL be a goose migration (`00007_audit.sql`) with Up/Down blocks, and queries SHALL be sqlc-generated.

#### Scenario: Migration applies and rolls back
- **WHEN** `goose up` and `goose down` are run
- **THEN** `audit_logs` is created and then dropped cleanly.

#### Scenario: sqlc regenerates after query change
- **WHEN** `sqlc generate` is run after a query change
- **THEN** the generated Go in `internal/infrastructure/db/` is updated.
