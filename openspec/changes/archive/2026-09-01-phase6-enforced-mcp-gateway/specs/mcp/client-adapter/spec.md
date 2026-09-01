## ADDED Requirements

### Requirement: Gateway-bound provider target resolution

The MCP client adapter SHALL resolve a provider invocation from a gateway-authorized
project connection and discovered tool binding. It SHALL reject an invocation whose
target does not match that authorization and SHALL not accept a raw external URL as
an invocation target.

#### Scenario: Resolve a verified provider target

- **WHEN** the gateway authorizes a capability invocation
- **THEN** the adapter uses the registered project connection and mapped discovered tool
- **AND** does not use caller-supplied endpoint or tool metadata.
