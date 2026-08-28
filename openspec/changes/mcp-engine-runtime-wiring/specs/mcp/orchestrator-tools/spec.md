## MODIFIED Requirements

### Requirement: Propose event tool

The MCP server SHALL expose a `propose_event` tool so a client can suggest an event
without mutating state directly.

#### Scenario: Propose and transition

- **WHEN** a client proposes an event for a workflow instance
- **THEN** the engine validates the event against the current state and executes the
  resulting transition (PRD 38)
- **AND** returns the updated state/transition result, or a validation error if the
  event is not allowed

#### Scenario: Persisted transition

- **WHEN** the engine executes a transition on a proposed event
- **THEN** the resulting workflow-instance current state is persisted via the repository
- **AND** a subsequent `get_current_state` returns the updated state

#### Scenario: Rejected proposal leaves state unchanged

- **WHEN** a proposed event is disallowed from the current state or no guard passes
- **THEN** the tool returns a structured rejection
- **AND** the workflow instance state is unchanged
