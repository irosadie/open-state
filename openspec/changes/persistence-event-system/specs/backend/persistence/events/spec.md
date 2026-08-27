## Purpose

Provides durable, tenant-isolated event persistence — append-only history, inbound dedup inbox, reliable delivery outbox, and idempotency records — behind the `IEventRepository` interface and a pgx adapter (PRD 27-32, 65-66, 128).

## ADDED Requirements

### Requirement: Append-only event history
The system SHALL persist an immutable event history in an `events` table.

- `events.event_id` SHALL be a logical id, unique per tenant.
- `events.type`, `events.source`, `events.aggregate_id`, `events.workflow_instance_id`, `events.timestamp`, `events.payload`, `events.correlation_id`, `events.causation_id`, `events.idempotency_key` SHALL store the PRD 27 event model.
- `events.source` SHALL be a VARCHAR in `USER`, `LLM`, `MCP`, `WEBHOOK`, `SYSTEM`, `SCHEDULER`, `ADMIN`, `API` (PRD 28).
- Events SHALL be append-only and SHALL NOT be modified/deleted during normal operation (PRD 51).

#### Scenario: Append an event
- **WHEN** an event is appended
- **THEN** a row is inserted into `events` with the full event model and a monotonic `sequence`.

#### Scenario: Replay event history
- **WHEN** event history for a workflow instance is requested
- **THEN** rows SHALL be returned in `sequence` order (PRD 32, 52).

### Requirement: Event ordering per instance
Events for the same workflow instance SHALL be ordered deterministically.

- The `events` table SHALL have an index on `(workflow_instance_id, sequence)` (PRD 32).

#### Scenario: Ordering query returns deterministic order
- **WHEN** events are queried per `workflow_instance_id`
- **THEN** they SHALL be ordered by `sequence`.

### Requirement: Inbound inbox with dedup
The system SHALL persist inbound external events in an `event_inbox` table for deduplication before processing (PRD 66).

- `event_inbox.idempotency_key` SHALL be unique per tenant.
- `event_inbox.status` SHALL be a VARCHAR in `RECEIVED`, `PROCESSING`, `PROCESSED`, `FAILED`.

#### Scenario: Deduplicate identical inbound event
- **WHEN** an inbound event with a duplicate `idempotency_key` is inserted
- **THEN** the insert SHALL be rejected (unique constraint), preventing double processing.

#### Scenario: Claim inbox entries for processing
- **WHEN** a worker claims inbound events
- **THEN** it SHALL atomically transition a batch from `RECEIVED` to `PROCESSING`.

### Requirement: Reliable outbox
The system SHALL persist events to be published in an `event_outbox` table (PRD 65).

- `event_outbox.status` SHALL be a VARCHAR in `PENDING`, `PUBLISHED`, `FAILED`.
- An outbox event SHALL be written atomically with the DB state change it accompanies (PRD 65, 69).

#### Scenario: Enqueue outbox event atomically
- **WHEN** a DB state change and its outbox event are committed
- **THEN** both SHALL be persisted in the same transaction.

#### Scenario: Publish pending outbox events
- **WHEN** a publisher processes pending outbox events
- **THEN** it SHALL transition them from `PENDING` to `PUBLISHED`.

### Requirement: Idempotency ledger
The system SHALL persist processed-event dedup outcomes in `idempotency_records` (PRD 30).

- `idempotency_records.idempotency_key` SHALL be unique per tenant.
- `idempotency_records.result_status` SHALL be in `PROCESSED`, `IGNORED`, `FAILED`.

#### Scenario: Record an idempotent outcome
- **WHEN** an event is processed
- **THEN** an `idempotency_records` row SHALL record the outcome keyed by `idempotency_key`.

#### Scenario: Repeated delivery is ignored
- **WHEN** an event with an already-recorded `idempotency_key` arrives
- **THEN** the system SHALL consult the ledger and skip re-processing.

### Requirement: Repository interface
The system SHALL expose event persistence through an `IEventRepository` interface.

- Every method SHALL require an explicit `tenantID string` (PRD 4, 96).
- Methods SHALL operate on domain entities, not sqlc rows (ADR-001).
- An outbox-enqueue operation SHALL run in a transaction with the accompanying state change.

#### Scenario: Repository methods are tenant-scoped
- **WHEN** any `IEventRepository` method is called with a `tenantID`
- **THEN** the returned rows SHALL all belong to that tenant.

### Requirement: PostgreSQL adapter
The system SHALL implement `IEventRepository` with a pgx adapter in `apps/api/internal/infrastructure/database/pgx_event_repository.go`.

#### Scenario: Adapter maps rows to entities and reports not-found
- **WHEN** a query returns no row
- **THEN** the adapter SHALL return a `NOT_FOUND` DomainError.
- **WHEN** a dedup unique violation occurs
- **THEN** the adapter SHALL surface a conflict (DomainError `CONFLICT`).

### Requirement: goose migration + sqlc
The schema SHALL be a goose migration (`00004_events.sql`) with Up/Down blocks, and queries SHALL be sqlc-generated.

#### Scenario: Migration applies and rolls back
- **WHEN** `goose up` and `goose down` are run
- **THEN** the four tables are created and then dropped cleanly.

#### Scenario: sqlc regenerates after query change
- **WHEN** `sqlc generate` is run after a query change
- **THEN** the generated Go in `internal/infrastructure/db/` is updated.
