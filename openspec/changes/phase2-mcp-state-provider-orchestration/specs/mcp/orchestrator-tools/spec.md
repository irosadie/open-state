## MODIFIED Requirements

### Requirement: Current state tool

The MCP server SHALL expose a `get_current_state` tool that returns the active state of a workflow instance, including its purpose, instructions, required context, allowed transitions, and structured provider capabilities required before each applicable transition.

#### Scenario: Read current state

- **WHEN** a client invokes `get_current_state` with a workflow instance id
- **THEN** the server returns the state id, purpose, instructions, required context, and the allowed events/transitions
- **AND** returns the provider capabilities required by the current state
- **AND** returns a not-found error if the instance is not in the tenant's scope

#### Scenario: Purpose and instructions

- **WHEN** a client invokes `get_current_state` for an instance
- **THEN** the server returns the current node's `purpose` (description) and `instructions` from the pinned workflow definition
- **AND** returns the node's `requiredContext` fields
- **AND** identifies which required capabilities are pending, satisfied, or failed for the current state

### Requirement: Propose event tool

The MCP server SHALL expose a `propose_event` tool so a client can suggest an event without mutating state directly. The engine SHALL validate that all required capability evidence for the current state and proposed transition has been accepted before applying a transition.

#### Scenario: Propose and transition

- **WHEN** a client proposes an event for a workflow instance
- **THEN** the engine validates the event against the current state and executes the resulting transition
- **AND** verifies the required capability evidence for the current state before applying the transition
- **AND** returns the updated state/transition result, or a validation error if the event is not allowed

#### Scenario: Required provider capability is incomplete

- **WHEN** a client proposes an event whose transition requires a provider capability without successful evidence
- **THEN** the State MCP rejects the proposal with a deterministic requirement-not-satisfied error
- **AND** does not update the workflow instance, state, or event history

### Requirement: Authorized capabilities tool

The MCP server SHALL expose a `get_allowed_capabilities` tool listing the capabilities authorized for a state/context, including the provider MCP server and concrete tool mapping for each returned capability when configured.

#### Scenario: List authorized capabilities

- **WHEN** a client invokes `get_allowed_capabilities` for a state/context
- **THEN** the server resolves capabilities via the security chain and returns those authorized
- **AND** includes safe provider/tool metadata for each capability
