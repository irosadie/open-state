# mcp/capability-execution Specification

## Purpose
Define how the platform safely and deterministically resolves and executes external
capabilities (MCP/INTERNAL/HTTP) behind a provider abstraction, with full security,
normalization, retry, timeout, and idempotency — independent of any concrete MCP SDK.

## Requirements

### Requirement: Capability provider abstraction

The platform SHALL execute capabilities only through a `CapabilityProvider` port so the
core engine remains agnostic to the concrete MCP implementation.

#### Scenario: Provider port invocation

- **WHEN** the platform needs to run a capability
- **THEN** it calls the bound provider's invoke operation with an invocation payload
- **AND** the provider returns a normalized result or a classified error

#### Scenario: Engine decoupling

- **WHEN** the capability execution layer is constructed
- **THEN** it depends only on the provider interface and not on any MCP SDK

### Requirement: Capability resolution

The platform SHALL resolve a logical capability to the correct provider and schema,
respecting tenant, workflow, and state bindings with most-restrictive-wins.

#### Scenario: Resolve with bindings

- **WHEN** a capability is resolved for a given tenant, workflow, and state
- **THEN** the most restrictive binding (DENY over ALLOW; state over workflow over
  tenant) determines availability

#### Scenario: Unknown or denied capability

- **WHEN** a capability is unknown or denied for the context
- **THEN** resolution SHALL return an authorization error and SHALL NOT invoke a provider

### Requirement: Execution security chain

The platform SHALL run the full security chain before invoking a provider.

#### Scenario: Full chain

- **WHEN** a capability is invoked
- **THEN** the platform authenticates, authorizes the tenant, workflow, and state,
  validates the input schema, applies rate limiting, and only then invokes the provider

#### Scenario: Schema validation failure

- **WHEN** the invocation payload fails input schema validation
- **THEN** invocation SHALL be rejected with a validation error and SHALL NOT reach the
  provider

### Requirement: Structured results and failures

The platform SHALL normalize provider outcomes into structured results or classified
errors, and map failures to deterministic capability events.

#### Scenario: Successful result

- **WHEN** a provider returns successfully
- **THEN** the platform normalizes the result and produces a capability result that may
  feed a transition event

#### Scenario: Failure classification

- **WHEN** a provider fails
- **THEN** the platform classifies the failure as one of timeout, unauthorized,
  validation, unavailable, or business error
- **AND** maps it to a corresponding capability event such as `capability.timeout` or
  `capability.unauthorized`

### Requirement: Retry and timeout policy

The platform SHALL retry only retryable errors with exponential backoff and jitter, and
SHALL enforce a configurable timeout per invocation.

#### Scenario: Retryable error

- **WHEN** a timeout or temporary network error occurs within the retry budget
- **THEN** the platform retries with exponential backoff and jitter

#### Scenario: Non-retryable error

- **WHEN** an authorization, validation, or business error occurs
- **THEN** the platform does NOT retry and returns the classified error

#### Scenario: Retry budget exhausted

- **WHEN** all retries are exhausted
- **THEN** the platform returns the final error according to state policy, for example
  transitioning to an error or human-handoff state

### Requirement: Idempotency

The platform SHALL support idempotency keys for capabilities that create external side
effects.

#### Scenario: Duplicate invocation suppressed

- **WHEN** an invocation carries an idempotency key equal to a previously completed
  invocation for the same workflow instance and action
- **THEN** the platform returns the prior result instead of re-invoking the provider

### Requirement: Sandbox and mock execution

The platform SHALL default to mocked or sandboxed provider execution.

#### Scenario: Default mock mode

- **WHEN** a capability is invoked without a real provider bound
- **THEN** the platform executes it through a mock provider instead of a live external
  call
- **AND** the platform indicates the result came from the mock provider

### Requirement: Deterministic testing

The capability execution layer SHALL be fully unit-testable without an HTTP, database,
or LLM dependency.

#### Scenario: No external dependency

- **WHEN** the execution layer is tested
- **THEN** it runs against in-memory fakes and mock providers
- **AND** tests cover resolver, security chain, retry, timeout, idempotency, and
  failure mapping with coverage above 80%
