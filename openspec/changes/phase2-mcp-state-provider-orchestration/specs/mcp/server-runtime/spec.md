## MODIFIED Requirements

### Requirement: MCP server startup

The platform SHALL run a standalone MCP server reachable over Streamable HTTP. During MCP initialization, the server SHALL provide concise instructions that identify OpenState as the authoritative controller for registered intents, workflow states, and transitions, and explain that external provider capabilities may be required before a transition is allowed.

#### Scenario: Server exposes tools

- **WHEN** the MCP server starts
- **THEN** it connects and declares tools for intent resolution, active workflow, context retrieval, capability-result reporting, and authorized capability invocation
- **AND** the MCP initialize result includes the OpenState state-control instructions

#### Scenario: Server instructions define the control boundary

- **WHEN** an LLM initializes a State MCP session
- **THEN** the instructions SHALL tell the LLM to resolve matching intents through OpenState, read the current state, use declared provider requirements, and submit results before proposing a transition
- **AND** the instructions SHALL state that the LLM is not the source of truth for workflow state

### Requirement: Intent resolution tool

The MCP server SHALL expose an intent-resolution tool that accepts a canonical intent identifier within the authenticated tenant/project scope and returns the classified intent, its mapped workflow, the state-control requirement, and the provider capabilities required by the resolved entry state when known.

#### Scenario: Resolve intent

- **WHEN** the LLM calls the intent-resolution tool with a canonical intent
- **THEN** the server resolves the canonical intent through the tenant/project-scoped intent catalog
- **AND** returns the mapped workflow and state-controller metadata
- **AND** returns structured provider requirements for the resolved workflow entry state when those requirements exist

#### Scenario: Unknown intent

- **WHEN** the LLM calls the intent-resolution tool with an identifier that is not routable in the requested tenant/project
- **THEN** the server returns a not-found error
- **AND** does not resolve a workflow by arbitrary workflow ID or slug

#### Scenario: Provider endpoint is not embedded as authority

- **WHEN** the intent response identifies a required provider MCP server
- **THEN** it SHALL use the configured logical provider server name/alias and concrete tool name
- **AND** it SHALL not instruct State MCP or the LLM to connect to an arbitrary endpoint
