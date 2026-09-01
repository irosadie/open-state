# mcp/client-adapter Specification

## Purpose
Define the real MCP client implementation of the `CapabilityProvider` port, including
secure resolution of `credential_reference`, so production capability execution can run
against live MCP servers while keeping the core engine MCP-agnostic.

## Requirements

### Requirement: MCP provider implementation

The platform SHALL provide a real MCP client implementation of the `CapabilityProvider`
port.

#### Scenario: Invoke over MCP

- **WHEN** the execution layer invokes a capability bound to an MCP provider
- **THEN** the client adapter connects to the configured MCP server
- **AND** calls the mapped MCP tool with the invocation payload
- **AND** returns a normalized result or classified error

#### Scenario: Engine remains decoupled

- **WHEN** the adapter is used in production
- **THEN** the core engine still depends only on the `CapabilityProvider` port
- **AND** not on the MCP SDK directly

### Requirement: Credential resolution

The platform SHALL resolve `credential_reference` to the actual credential from secure
infrastructure without storing secrets in workflow definitions.

#### Scenario: Resolve reference

- **WHEN** a capability is invoked
- **THEN** the adapter resolves the referenced credential from a secure store
- **AND** uses it for the MCP call

#### Scenario: No secret leakage

- **WHEN** a credential is resolved or an MCP call is made
- **THEN** the adapter never logs the credential, token, or authorization header

### Requirement: Result normalization

The MCP client adapter SHALL normalize raw MCP tool results into the platform's
`InvocationResult` shape.

#### Scenario: Successful MCP call

- **WHEN** the MCP tool returns a success payload
- **THEN** the adapter normalizes it into an `InvocationResult` with data
- **AND** flags it as a real (non-mock) result

#### Scenario: MCP failure

- **WHEN** the MCP call fails or times out
- **THEN** the adapter maps the failure to a classified `CapabilityError`
- **AND** never exposes the raw error to callers

### Requirement: Gateway-bound provider target resolution

The MCP client adapter SHALL resolve a provider invocation from a gateway-authorized
project connection and discovered tool binding. It SHALL reject an invocation whose
target does not match that authorization and SHALL not accept a raw external URL as
an invocation target.

#### Scenario: Resolve a verified provider target

- **WHEN** the gateway authorizes a capability invocation
- **THEN** the adapter uses the registered project connection and mapped discovered tool
- **AND** does not use caller-supplied endpoint or tool metadata.
