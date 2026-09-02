## MODIFIED Requirements

### Requirement: Propose event tool

The MCP server SHALL expose a `propose_event` tool so a client can suggest an event without mutating state directly. In secure mode, the tool SHALL require a stable idempotency key and correlation id, and repeated proposals with the same tenant and key SHALL return the original event outcome without appending another event.

#### Scenario: Propose and transition

- **WHEN** a client proposes an event for a workflow instance with an idempotency key
- **THEN** the engine validates the event against the current state and executes the resulting transition
- **AND** the client uses the exact `requiredContext` key names returned by `get_current_state`, including dotted names
- **AND** returns the updated state/transition result, or a validation error if the event is not allowed

#### Scenario: Retry the same proposal

- **WHEN** a client retries the same event proposal with the same tenant and idempotency key
- **THEN** the server returns the original event outcome
- **AND** does not append a second event or execute the transition again

### Requirement: Lifecycle tools

The MCP server SHALL expose `start_workflow`, `suspend_workflow`, `resume_workflow`, and `cancel_workflow` tools. In secure mode, `start_workflow` SHALL require a stable idempotency key so a retry cannot create a second workflow instance.

#### Scenario: Start a workflow

- **WHEN** a client invokes `start_workflow` with a workflow id and an idempotency key
- **THEN** the engine creates and runs the workflow instance
- **AND** returns the created instance outcome

#### Scenario: Retry the same workflow start

- **WHEN** a client retries the same workflow start with the same tenant and idempotency key
- **THEN** the server returns the original workflow instance outcome
- **AND** does not create a second instance

#### Scenario: Suspend / resume / cancel

- **WHEN** a client suspends, resumes, or cancels a workflow instance
- **THEN** the engine updates the instance lifecycle status accordingly (PRD 42-43)
- **AND** returns a not-found error for instances outside the tenant's scope

### Requirement: Gateway capability execution tool

The MCP server SHALL expose a state-gated provider execution tool that accepts only the workflow context and capability input required by the current state. The tool SHALL use the gateway authorization path and return normalized, redacted results. A failed secure invocation SHALL be returned as an explicit hard stop.

#### Scenario: Execute a current-state requirement

- **WHEN** a client invokes the gateway execution tool for a capability required by the current state
- **THEN** the server forwards the authorized call through the gateway
- **AND** returns the normalized execution result and safe evidence status

#### Scenario: Execute a non-current requirement

- **WHEN** a client invokes the gateway execution tool for a capability not authorized by the current state
- **THEN** the server rejects the call without contacting the provider
- **AND** returns `ok: false`, `invoked: false`, `hardStop: true`, and `nextAction: STOP`

#### Scenario: Provider or mapping failure

- **WHEN** the gateway cannot resolve a project binding, provider connection, catalog, tool, or provider result
- **THEN** the server returns `ok: false`, `invoked: false`, `hardStop: true`, and a stable safe failure code
- **AND** the client MUST NOT select another capability, scope, provider, or tool or propose a transition
