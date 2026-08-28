## MODIFIED Requirements

### Requirement: Current state tool

The MCP server SHALL expose a `get_current_state` tool that returns the active state
of a workflow instance.

#### Scenario: Read current state

- **WHEN** a client invokes `get_current_state` with a workflow instance id
- **THEN** the server returns the state id, purpose, instructions, and the allowed
  events/transitions (PRD 12, 14, 33-34)
- **AND** returns a not-found error if the instance is not in the tenant's scope

#### Scenario: Purpose and instructions

- **WHEN** a client invokes `get_current_state` for an instance
- **THEN** the server returns the current node's `purpose` (description) and
  `instructions` from the pinned workflow definition
- **AND** returns the node's `requiredContext` fields

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

#### Scenario: Engine-reproduced state

- **WHEN** the engine replays the instance's recorded events
- **THEN** the events are re-driven through the engine in deterministic sequence order
- **AND** the tool returns the reproduced resulting context/state without persisting
