# backend/persistence/audit-writer Specification

## Purpose
Define the application-layer audit writer that records important operations into
the append-only, tenant-isolated `audit_logs` table (PRD §50). The existing
`IAuditRepository` and `00007_audit.sql` already provide the persistence seam
(see `backend/persistence/audit-adapter`); this change wires real business
operations to emit audit entries so the trail is queryable (see
`backend/audit-api`) and visible (see `web/audit-ui`).

## Requirements

### Requirement: Audit writer service

The platform SHALL provide an `AuditWriter` application service that appends
audit entries for important operations (PRD §50).

- `AuditWriter` SHALL depend on `IAuditRepository`.
- Each `Record`/`Write` call SHALL accept an explicit `tenantID`, `actor`,
  `action`, `resourceType`, `resourceID`, optional `before`/`after`, and
  optional `correlationID`.
- `AuditWriter` SHALL be injected into the services that perform auditable
  operations (workflow publish, capability invoke, binding create/delete).
- A write failure SHALL NOT fail the originating business operation by default
  (best-effort) unless the operation is itself an audit-critical mutation.

#### Scenario: Service records an audit entry

- **WHEN** an important operation completes
- **THEN** `AuditWriter` appends a matching `audit_logs` row

#### Scenario: Best-effort failure

- **WHEN** writing an audit entry fails during a normal operation
- **THEN** the originating operation SHALL still succeed and the failure SHALL
  be logged

### Requirement: Audit workflow publish

The platform SHALL record an audit entry when a workflow is published (PRD §50).

- On `workflow.publish`, `AuditWriter` SHALL append action
  `workflow.published`, resource_type `workflow`, resource_id = the workflow id.
- The `after` SHALL include the published version.

#### Scenario: Publish is audited

- **WHEN** an authorized user publishes a workflow
- **THEN** an audit row with action `workflow.published` and the new version is
  appended

### Requirement: Audit capability invocation

The platform SHALL record audit entries when a capability is invoked and when it
is denied (PRD §50, §62).

- On successful invocation, `AuditWriter` SHALL append action
  `capability.invoked`, resource_type `capability`, resource_id = the capability
  id.
- On a denied invocation (authorization, rate limit, validation), `AuditWriter`
  SHALL append action `capability.denied` with the reason in `after`.

#### Scenario: Successful invocation is audited

- **WHEN** a capability is invoked successfully
- **THEN** an audit row with action `capability.invoked` is appended

#### Scenario: Denied invocation is audited

- **WHEN** a capability invocation is denied
- **THEN** an audit row with action `capability.denied` is appended, including
  the denial reason

### Requirement: Audit binding create/delete

The platform SHALL record audit entries when a capability binding is created or
deleted (PRD §50, §60).

- On create, `AuditWriter` SHALL append action `binding.created`,
  resource_type `binding`, resource_id = the binding id, with `after` = the
  binding.
- On delete, `AuditWriter` SHALL append action `binding.deleted`, with `before`
  = the deleted binding.

#### Scenario: Binding creation is audited

- **WHEN** a binding is created
- **THEN** an audit row with action `binding.created` and the binding `after`
  is appended

#### Scenario: Binding deletion is audited

- **WHEN** a binding is deleted
- **THEN** an audit row with action `binding.deleted` and the binding `before`
  is appended

### Requirement: Audit action constants

The platform SHALL extend the typed audit action set with the operation actions
used by `AuditWriter` (PRD §50).

- Constants `workflow.published`, `capability.invoked`, `capability.denied`,
  `binding.created`, and `binding.deleted` SHALL be declared.
- The audit action set SHALL also include the RBAC actions defined in
  `auth/rbac-audit`.

#### Scenario: Action values are constrained

- **WHEN** an audit entry is written
- **THEN** its `action` SHALL be one of the declared typed constants

### Requirement: Actor and tenant derivation

The platform SHALL derive the audit actor and tenant from the authenticated
request context, never from the request body (PRD §50, §4, §96).

- The actor SHALL be the authenticated user id (or a system marker for
  non-user-driven operations).
- The tenant SHALL be the request tenant.

#### Scenario: Actor is the authenticated user

- **WHEN** an authenticated user performs an auditable operation
- **THEN** the audit row's actor SHALL be that user id

### Requirement: Correlation id propagation

The platform SHALL propagate an optional `correlation_id` into audit entries
when a business correlation is available (PRD §50).

- A correlation id carried through the operation context SHALL be written to
  `audit_logs.correlation_id`.

#### Scenario: Correlation id recorded

- **WHEN** an operation runs within a correlated business context
- **THEN** the resulting audit row SHALL include the correlation id
