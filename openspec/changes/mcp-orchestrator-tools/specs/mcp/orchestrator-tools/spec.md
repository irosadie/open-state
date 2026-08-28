# mcp/orchestrator-tools Specification

## Purpose

Define the remaining MCP orchestrator tools that let a 3rd-party LLM/RAG client read
state, list authorized capabilities, propose events, drive workflow lifecycle, and
replay history — completing the stable MCP tool contract (PRD 170).

## ADDED Requirements

### Requirement: Current state tool

The MCP server SHALL expose a `get_current_state` tool that returns the active state
of a workflow instance.

#### Scenario: Read current state

- **WHEN** a client invokes `get_current_state` with a workflow instance id
- **THEN** the server returns the state id, purpose, instructions, and the allowed
  events/transitions (PRD 12, 14, 33-34)
- **AND** returns a not-found error if the instance is not in the tenant's scope

### Requirement: Authorized capabilities tool

The MCP server SHALL expose a `get_allowed_capabilities` tool listing the capabilities
authorized for a state/context.

#### Scenario: List authorized capabilities

- **WHEN** a client invokes `get_allowed_capabilities` for a state/context
- **THEN** the server resolves capabilities via the security chain and returns those
  authorized (PRD 59-62)

### Requirement: Propose event tool

The MCP server SHALL expose a `propose_event` tool so a client can suggest an event
without mutating state directly.

#### Scenario: Propose and transition

- **WHEN** a client proposes an event for a workflow instance
- **THEN** the engine validates the event against the current state and executes the
  resulting transition (PRD 38)
- **AND** returns the updated state/transition result, or a validation error if the
  event is not allowed

### Requirement: Lifecycle tools

The MCP server SHALL expose `start_workflow`, `suspend_workflow`, `resume_workflow`, and
`cancel_workflow` tools.

#### Scenario: Start a workflow

- **WHEN** a client invokes `start_workflow` with a workflow id (and optional context)
- **THEN** the engine creates and runs the workflow instance (PRD 25)

#### Scenario: Suspend / resume / cancel

- **WHEN** a client suspends, resumes, or cancels a workflow instance
- **THEN** the engine updates the instance lifecycle status accordingly (PRD 42-43)
- **AND** returns a not-found error for instances outside the tenant's scope

### Requirement: Instances and history tools

The MCP server SHALL expose `get_workflow_instances`, `get_history`, and
`replay_workflow` tools.

#### Scenario: List instances

- **WHEN** a client invokes `get_workflow_instances` (filtered by tenant/status)
- **THEN** the server returns the tenant's workflow instances (PRD 142)

#### Scenario: Read history

- **WHEN** a client invokes `get_history` for a workflow instance
- **THEN** the server returns the event/state history in deterministic sequence order
  (PRD 52)

#### Scenario: Replay workflow

- **WHEN** a client invokes `replay_workflow` for a workflow instance
- **THEN** the engine replays the recorded events to reproduce the resulting state
  (PRD 52)

### Requirement: Tenant scoping

All orchestrator tools SHALL be tenant-scoped (PRD 4, 96).

#### Scenario: Cross-tenant access is impossible

- **WHEN** any orchestrator tool targets a workflow instance
- **THEN** the server SHALL resolve and enforce the tenant id from the authenticated
  context and return not-found for resources outside it.
