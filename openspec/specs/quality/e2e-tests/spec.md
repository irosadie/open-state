# quality/e2e-tests Specification

## Purpose
Define an end-to-end test that drives the full integration path from an LLM/MCP mock
through the engine to a state transition, proving the platform works without a real
LLM (PRD 170).

## Requirements

### Requirement: End-to-end integration harness

The platform SHALL provide an E2E test that exercises the full path with a
deterministic LLM/MCP mock.

#### Scenario: Full path execution

- **GIVEN** a seeded example workflow and a mock MCP/LLM client
- **WHEN** the mock resolves an intent and proposes an event through the MCP tool
  surface
- **THEN** the engine validates and executes the state transition
- **AND** the test asserts the resulting state (PRD 126)

### Requirement: MCP tool surface driven by mock

The E2E SHALL drive the real MCP tool handlers (or server) rather than the engine in
isolation.

#### Scenario: Drive real tool handlers

- **WHEN** the E2E invokes `resolve_intent` / `get_active_workflow` and
  `propose_event` via the mock client
- **THEN** the real MCP tool handlers and the engine process the request
- **AND** no real LLM is called (PRD 170)

### Requirement: Deterministic mock

The mock client SHALL be deterministic so E2E results are repeatable.

#### Scenario: Repeatable outcomes

- **GIVEN** a fixed set of mock responses
- **WHEN** the E2E runs repeatedly
- **THEN** the resulting state transitions are identical each run

### Requirement: Persistent-state assertion

The E2E SHALL assert against persisted state via the repository layer.

#### Scenario: Persisted result

- **WHEN** a transition completes in the E2E
- **THEN** the resulting workflow instance/state is readable through the repository
- **AND** the test asserts it equals the expected terminal/current state
