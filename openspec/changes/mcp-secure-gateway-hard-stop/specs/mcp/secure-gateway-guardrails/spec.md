## Purpose

Provide deterministic secure-gateway failure semantics so an LLM cannot replace a state-declared provider capability or continue a workflow after the gateway blocks execution.

## ADDED Requirements

### Requirement: Secure gateway failures are hard stops

The secure State MCP `invoke_capability` response MUST identify every failed invocation as a hard stop, include a stable machine-readable code, and instruct the client to stop before proposing a transition or selecting another capability.

#### Scenario: Project mapping is missing

- **WHEN** the active state declares an MCP capability without an active project binding
- **THEN** `invoke_capability` returns `ok: false`, `invoked: false`, `hardStop: true`, and a mapping-specific code
- **AND** the response does not expose a provider endpoint, credential, or alternate target

#### Scenario: Provider invocation fails

- **WHEN** an active project binding reaches a provider and the provider returns a timeout, transport, validation, or business failure
- **THEN** `invoke_capability` returns `ok: false`, `invoked: false`, `hardStop: true`, and the normalized failure code
- **AND** the client MUST NOT invoke another capability or propose a transition based on the failed call

### Requirement: Secure clients cannot use broader scopes as a fallback

Secure MCP instructions and tool descriptions MUST state that only capabilities declared by the current state may be invoked. Workflow-, tenant-, and context-level discovery MUST NOT authorize a substitute capability for the active state.

#### Scenario: Tenant capability is not declared by the state

- **WHEN** the client discovers a similarly named capability at tenant or workflow scope after a current-state invocation fails
- **THEN** the client does not invoke the broader-scope capability
- **AND** the gateway continues to reject any attempted call that is not declared by the current state

### Requirement: State-changing MCP retries are idempotent

Secure workflow start and event proposal calls MUST accept a caller-supplied idempotency key. Repeating the same operation with the same tenant and key MUST return the original outcome without creating another workflow instance or event.

#### Scenario: Repeated event proposal

- **WHEN** a client retries the same event proposal with the same idempotency key
- **THEN** the server returns the original event outcome and does not append a second event

#### Scenario: Repeated workflow start

- **WHEN** a client retries the same workflow start with the same idempotency key
- **THEN** the server returns the original workflow instance outcome and does not create a second instance
