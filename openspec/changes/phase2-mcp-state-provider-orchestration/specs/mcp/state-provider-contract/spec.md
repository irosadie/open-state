## Purpose

Define the runtime contract that connects an OpenState intent/state requirement to a pre-configured third-party MCP provider and its concrete tool, so an LLM can execute the correct data operation without guessing.

## ADDED Requirements

### Requirement: State-declared provider dependency

The platform SHALL represent every state-required external operation as a structured capability requirement containing a stable capability identifier, provider MCP server identifier, tool name, purpose, required flag, and input/output mapping metadata when applicable.

#### Scenario: State requires a provider tool

- **WHEN** a workflow state requires external data or an external side effect
- **THEN** the State MCP response SHALL identify the logical capability, provider MCP server, concrete tool, and reason for the requirement
- **AND** the response SHALL indicate whether the capability is mandatory before a named transition

#### Scenario: State has no external dependency

- **WHEN** a workflow state can complete using only state context and user input
- **THEN** the State MCP response SHALL return an empty provider requirement list
- **AND** the LLM SHALL not be instructed to call an unrelated provider tool

### Requirement: Stable provider server identity

Provider requirements SHALL reference a logical provider MCP server name or alias and a concrete tool name as the authoritative target. The MCP client/LLM host SHALL resolve that alias to a pre-configured connection. Provider endpoints SHALL NOT be stored in workflow definitions or treated as instructions for State MCP to forward or open a provider connection.

#### Scenario: Provider alias mapping

- **WHEN** the workflow declares provider server alias `padel-provider` and tool `padel.check_available`
- **THEN** the LLM host SHALL resolve `padel-provider` to its already-configured MCP connection
- **AND** changing the provider endpoint SHALL not require changing the workflow state definition

#### Scenario: Provider alias is not connected

- **WHEN** a state requirement references a provider server alias that is not available in the connected MCP sessions
- **THEN** the LLM host SHALL report a structured provider-unavailable result to State MCP
- **AND** the workflow SHALL not advance through a transition that requires that provider

### Requirement: Provider execution evidence

The platform SHALL record an authorized, normalized result for every required provider capability before considering the corresponding state requirement satisfied. A direct provider tool response without an accepted State MCP execution record SHALL NOT by itself authorize a state transition.

#### Scenario: Required provider result is accepted

- **WHEN** a required provider capability returns a valid result through the configured execution path
- **THEN** the platform SHALL persist execution evidence against the workflow instance and current state
- **AND** the evidence SHALL include the logical capability, provider server alias, tool name, outcome, and correlation identifier

#### Scenario: Required provider result is missing

- **WHEN** an LLM proposes a transition while a required provider capability has no successful execution evidence for the current state
- **THEN** the State MCP SHALL reject the transition with a deterministic requirement-not-satisfied error
- **AND** the current state SHALL remain unchanged

### Requirement: Capability result reporting

The State MCP SHALL expose a capability-result reporting operation that accepts a workflow instance, state requirement, provider server alias, concrete tool name, correlation identifier, and normalized provider result or failure. The operation SHALL validate that the report matches a currently declared requirement before persisting evidence.

#### Scenario: LLM reports a provider result

- **WHEN** the LLM reports the result of a provider tool it invoked on a connected MCP server
- **THEN** State MCP SHALL verify the provider alias and tool against the current state's declared requirement
- **AND** SHALL validate the result against the declared output contract when one exists
- **AND** SHALL persist accepted evidence for the workflow instance

#### Scenario: LLM reports an undeclared tool

- **WHEN** the LLM reports a provider server or tool that is not required or authorized for the current state
- **THEN** State MCP SHALL reject the report
- **AND** SHALL not mark any capability requirement as satisfied

#### Scenario: LLM reports a provider failure

- **WHEN** the LLM reports a timeout, unavailable provider, validation failure, or business failure
- **THEN** State MCP SHALL persist the classified failure status when auditability requires it
- **AND** SHALL keep the corresponding required capability unsatisfied

### Requirement: Secret-safe provider metadata

Provider requirement metadata SHALL be safe to expose to an LLM and SHALL NOT contain provider credentials, bearer tokens, connection secrets, or resolved credential values.

#### Scenario: Provider metadata is returned

- **WHEN** the State MCP returns provider requirements
- **THEN** it SHALL include only logical server/tool metadata and safe execution instructions
- **AND** it SHALL omit credentials and authorization headers
