## Purpose

Provide tenant-scoped Admin Console operations for safe runtime lifecycle commands and read-only event inspection, while retaining the existing event and RBAC contracts.

## ADDED Requirements

### Requirement: Instance commands require exact existing runtime permissions

The system SHALL accept `suspend`, `resume`, and `retry` requests only for an instance in the authenticated tenant. It SHALL require `instance:suspend`, `instance:resume`, or `instance:retry` respectively and SHALL delegate valid requests to the runtime orchestration service.

#### Scenario: Authorized Operator suspends a tenant instance

- **WHEN** a caller with `instance:suspend` submits a suspend request for an instance in the caller's tenant that can be suspended
- **THEN** the system submits the runtime lifecycle command
- **AND** returns the resulting accepted or updated instance representation
- **AND** records an audit entry for the command.

#### Scenario: Caller attempts a command without its exact permission

- **WHEN** a caller lacks the exact permission required by a requested lifecycle command
- **THEN** the system rejects the command without invoking runtime orchestration
- **AND** returns a standard authorization error.

### Requirement: Runtime commands validate tenant ownership and lifecycle state

The system SHALL resolve an instance within the authenticated tenant before applying a command. It SHALL reject unavailable or invalid lifecycle transitions without silently overwriting newer runtime state.

#### Scenario: Command targets an instance from another tenant

- **WHEN** a caller submits a lifecycle command for an instance outside the caller's tenant
- **THEN** the system returns the standard tenant-scoped not-found or forbidden response
- **AND** does not invoke a command for that instance.

#### Scenario: Resume is invalid for the current state

- **WHEN** a caller requests resume for an instance whose current lifecycle state cannot be resumed
- **THEN** the system returns a stable conflict response with no state change
- **AND** preserves the current runtime state.

### Requirement: Event browsing is immutable and tenant-scoped

The system SHALL provide authorized users with paginated, filterable event list and detail reads for the authenticated tenant. It SHALL expose no Admin Console endpoint that edits, deletes, replays, or injects persisted events.

#### Scenario: Authorized user browses tenant events

- **WHEN** a caller with `instance:read` requests a filtered event list
- **THEN** the system returns only events belonging to the caller's tenant
- **AND** includes persisted event identity, source, timestamps, correlation metadata, and safe payload representation according to the event contract.

#### Scenario: Client attempts an event mutation route

- **WHEN** a client attempts to edit, delete, replay, or inject an event through Admin Console APIs
- **THEN** the system exposes no supported mutation operation
- **AND** stored event history remains unchanged.

### Requirement: Runtime lifecycle mutations are auditable

The system SHALL record each lifecycle command outcome using the existing audit format with actor, tenant, target instance, action, outcome, and available correlation context.

#### Scenario: Retry command is rejected by lifecycle validation

- **WHEN** a retry command fails lifecycle validation or orchestration conflict checks
- **THEN** the system returns the appropriate stable error
- **AND** records the rejected command outcome when audit policy requires it.
