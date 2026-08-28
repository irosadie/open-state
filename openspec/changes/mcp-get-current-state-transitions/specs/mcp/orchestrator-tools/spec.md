## MODIFIED Requirements

### Requirement: Current state tool

The MCP server SHALL expose a `get_current_state` tool that returns the active state
of a workflow instance.

#### Scenario: Read current state

- **WHEN** a client invokes `get_current_state` with a workflow instance id
- **THEN** the server returns the state id, purpose, instructions, and the allowed
  events/transitions (PRD 12, 14, 33-34)
- **AND** returns a not-found error if the instance is not in the tenant's scope

#### Scenario: Allowed events from current state

- **WHEN** a client invokes `get_current_state` for an instance whose workflow has
  transitions from the current state
- **THEN** the server returns an `allowedTransitions` array with each transition's
  `event`, `targetStateId`, and `priority`
- **AND** the transitions are derived from the instance's pinned workflow definition

#### Scenario: No transitions

- **WHEN** the current state has no outgoing transitions (e.g. a terminal state)
- **THEN** the server returns an empty `allowedTransitions` array
