## ADDED Requirements

### Requirement: Secure gateway tool surface

When secure gateway mode is enabled, the State MCP server SHALL advertise only the
state-control and gateway invocation operations needed by the authenticated project.
It SHALL not expose a raw provider endpoint, provider credential, or unrestricted
provider-tool registry.

#### Scenario: Initialize secure gateway mode

- **WHEN** a client initializes State MCP in secure gateway mode
- **THEN** the server declares the state-control and gateway tool surface
- **AND** omits external provider connection secrets and endpoint details.
