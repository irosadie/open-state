## Purpose

Record a safe, append-only per-turn trace for application-owned orchestration
stages while representing LLM, RAG, and MCP provider activity only as sanitized
integration metadata (PRD 143, 170).

## ADDED Requirements

### Requirement: Append-only per-turn trace

The platform SHALL persist an ordered, tenant-scoped debug trace for each
observable runtime turn. A trace entry SHALL contain the workflow instance,
optional turn id, stage, source, status, occurrence time, correlation id when
available, and sanitized structured attributes.

- Supported stages SHALL cover intent resolution, active workflow lookup, current
  state lookup, context resolution, RAG integration, MCP capability activity, LLM
  integration, event handling, guard evaluation, and transition selection.
- Trace entries SHALL be append-only and ordered within an instance/turn.

#### Scenario: Application turn produces a trace

- **WHEN** OpenState processes a runtime turn through its orchestration boundary
- **THEN** the platform appends ordered trace entries for the stages it observes
- **AND** each entry is scoped to the turn's tenant and instance

#### Scenario: Partial trace remains useful

- **WHEN** only some stages are observed for a turn
- **THEN** the platform returns the recorded stages in order
- **AND** marks unavailable stages as not recorded rather than fabricating data

### Requirement: External-provider boundary

The platform SHALL treat LLM, RAG, and externally hosted MCP providers as
independent systems.

- A trace entry for an external provider SHALL contain only application-observed
  metadata: provider alias, operation/reference id when supplied, status, duration,
  correlation id, and sanitized summaries.
- The Runtime Inspector and trace APIs SHALL NOT call, poll, authenticate to, or
  retrieve data from an external provider to populate a trace.

#### Scenario: External LLM activity is displayed safely

- **WHEN** an integration supplies an LLM operation reference and completion
  metadata to OpenState
- **THEN** the trace exposes that metadata as an external-provider entry
- **AND** the inspector does not issue a request to the LLM provider

### Requirement: Trace redaction and access control

The platform SHALL require `debug:read` for debug-trace queries and SHALL redact
secrets, credentials, sensitive PII, raw prompts, raw model responses, and raw RAG
documents before storing or returning trace attributes.

#### Scenario: Viewer is denied Debug View

- **WHEN** a user without `debug:read` requests a runtime debug trace
- **THEN** the platform denies the request
- **AND** does not reveal trace metadata

#### Scenario: Sensitive external payload is not retained

- **WHEN** an integration attempts to record a raw prompt, credential, or RAG
  document in trace attributes
- **THEN** the platform redacts or omits the sensitive value before persistence
- **AND** no response can recover the original value
