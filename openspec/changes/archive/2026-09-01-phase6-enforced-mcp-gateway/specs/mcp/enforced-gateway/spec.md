## Purpose

Provide a secure execution mode in which OpenState forwards only state-authorized
calls to registered provider MCP tools, preventing an LLM from bypassing the state
gatekeeper.

## ADDED Requirements

### Requirement: State-authorized gateway invocation

The platform SHALL expose a gateway invocation operation that executes a provider
tool only when the authenticated tenant/project, workflow instance, current state,
required capability, and verified project MCP binding authorize it. The caller SHALL
not select an arbitrary provider endpoint, connection alias, or tool as the authority.

#### Scenario: Invoke an authorized provider action

- **WHEN** an MCP client requests a required capability authorized by the current state
- **THEN** the gateway resolves its registered project binding and invokes the mapped provider tool
- **AND** returns a normalized result or classified failure.

#### Scenario: Attempt to bypass the required capability

- **WHEN** a client requests a provider action not authorized by the current state
- **THEN** the gateway rejects the request before contacting any provider.

### Requirement: Evidence-gated transition

The gateway SHALL validate provider input and normalized output, persist accepted
execution evidence with correlation and idempotency metadata, and allow a subsequent
state transition only when all required evidence is satisfied.

#### Scenario: Successful execution enables transition

- **WHEN** a required gateway invocation succeeds with valid normalized output
- **THEN** the platform stores accepted evidence for the current state requirement
- **AND** a subsequent eligible transition can proceed.

#### Scenario: Provider failure leaves state unchanged

- **WHEN** a provider invocation fails, times out, or returns invalid output
- **THEN** the platform records a classified result when auditability requires it
- **AND** does not satisfy the requirement or transition the workflow.

### Requirement: Secure and advisory deployment modes

The platform SHALL document and distinguish a secure gateway mode, where the LLM sees
only OpenState, from direct-two-MCP advisory mode. Direct provider calls in advisory
mode SHALL not be documented or represented as enforced state control.

#### Scenario: Secure mode discovery

- **WHEN** an LLM initializes the State MCP in secure gateway mode
- **THEN** it receives the state-authorized provider execution surface
- **AND** it does not receive provider endpoints or credentials.
