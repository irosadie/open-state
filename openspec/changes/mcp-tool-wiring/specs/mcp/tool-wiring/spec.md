# mcp/tool-wiring Specification

## Purpose

Complete the wiring of the MCP tools that are currently stubs or partially wired so a
3rd-party LLM/RAG client receives truthful, enforceable results: active workflow lookup,
authorized + validated capability invocation, real intent resolution, and workflow
replay (PRD 52, 170).

## ADDED Requirements

### Requirement: Active workflow lookup is real

The `get_active_workflow` tool SHALL resolve the active workflow instance for a
conversation from persisted runtime data, not a stub.

#### Scenario: Resolve active instance

- **WHEN** a client invokes `get_active_workflow` with a conversation id
- **THEN** the server returns the active workflow instance for that conversation (its
  id, workflow id, status, current state) or reports none if no instance is active
  (PRD 10, 142).

### Requirement: Capability invocation is authorized and validated

The `invoke_capability` / `execute_capability` tool SHALL run through the capability
resolver (authorization) and the schema validator before any provider call (PRD 59-62).

#### Scenario: Authorized invocation

- **WHEN** a client invokes a capability that is not authorized for the context
- **THEN** the server returns an authorization error and does NOT call a provider
  (PRD 59-62).

#### Scenario: Schema validation

- **WHEN** the invocation payload fails the capability's input schema validation
- **THEN** the server returns a validation error and does NOT call a provider
  (PRD 62).

### Requirement: Intent resolution is real

The `resolve_intent` tool SHALL resolve an intent from the real intent registry +
workflow definition, not a dummy list.

#### Scenario: Resolve intent to workflow

- **WHEN** a client invokes `resolve_intent` with an intent id + project
- **THEN** the server resolves the intent to its workflow definition + entry state
  (PRD 38, 171).

### Requirement: Workflow replay tool

The platform SHALL expose a `replay_workflow` tool that replays a workflow instance's
recorded event history to reproduce its resulting state (PRD 52).

#### Scenario: Replay reproduces state

- **WHEN** a client invokes `replay_workflow` with a tenant + workflow instance id
- **THEN** the server replays the instance's events in deterministic sequence order
  and returns the reproduced state, or a not-found error for instances outside the
  tenant's scope.
