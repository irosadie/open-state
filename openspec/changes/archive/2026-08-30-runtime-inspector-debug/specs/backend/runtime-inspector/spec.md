## Purpose

Provide a tenant-isolated, operator-facing read model for inspecting the current
and historical execution state of persisted workflow instances (PRD 142, 144).

## ADDED Requirements

### Requirement: Tenant-scoped runtime instance discovery

The platform SHALL allow a user with `instance:read` to list workflow instances
within their tenant and filter by lifecycle status, workflow, or correlation key.
Each list item SHALL identify the workflow, pinned version, current state,
lifecycle status, last activity time, and correlation id when one exists.

#### Scenario: Operator lists active instances

- **WHEN** an authorized operator requests runtime instances for their tenant
- **THEN** the platform returns only instances belonging to that tenant
- **AND** each item includes its current state and pinned workflow version

#### Scenario: Cross-tenant instance is not exposed

- **WHEN** a user requests an instance that belongs to another tenant
- **THEN** the platform SHALL NOT disclose the instance or its metadata

### Requirement: Runtime instance detail

The platform SHALL provide an authorized instance detail view containing the
workflow/version, lifecycle status, current state, available and missing context,
and an ordered timeline of state and event activity.

- Timeline entries SHALL be ordered by their persisted sequence or occurrence
  time and SHALL preserve the correlation id when available.
- Sensitive context values SHALL be redacted before they leave the backend.

#### Scenario: Inspect a running workflow

- **WHEN** an authorized operator opens a running instance
- **THEN** the platform returns the current state, sanitized context, and ordered
  history required to explain how that instance reached its current state

#### Scenario: Sensitive context is protected

- **WHEN** inspected context contains a value classified as sensitive
- **THEN** the returned detail SHALL replace that value with a redacted marker
- **AND** SHALL NOT return the original value

### Requirement: Deterministic decision evidence

The runtime detail SHALL expose structured reason codes for recorded event,
guard, and transition outcomes when that evidence exists.

- Evidence SHALL identify the rule or transition evaluated, its outcome, and a
  machine-readable reason code such as `GUARD_FAILED`.
- Evidence SHALL NOT contain LLM chain-of-thought or unstructured model reasoning.

#### Scenario: Rejected transition is explained

- **WHEN** an event is rejected because its guard does not pass
- **THEN** the detail view returns the guard outcome and `GUARD_FAILED` reason
  code
- **AND** does not return model reasoning as an explanation
