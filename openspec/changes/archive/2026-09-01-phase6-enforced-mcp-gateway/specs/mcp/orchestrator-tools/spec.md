## ADDED Requirements

### Requirement: Gateway capability execution tool

The MCP server SHALL expose a state-gated provider execution tool that accepts only
the workflow context and capability input required by the current state. The tool
SHALL use the gateway authorization path and return normalized, redacted results.

#### Scenario: Execute a current-state requirement

- **WHEN** a client invokes the gateway execution tool for a capability required by the current state
- **THEN** the server forwards the authorized call through the gateway
- **AND** returns the normalized execution result and safe evidence status.

#### Scenario: Execute a non-current requirement

- **WHEN** a client invokes the gateway execution tool for a capability not authorized by the current state
- **THEN** the server rejects the call without contacting the provider.
