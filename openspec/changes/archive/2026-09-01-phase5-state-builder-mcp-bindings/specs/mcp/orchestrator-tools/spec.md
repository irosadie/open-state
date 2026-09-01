## ADDED Requirements

### Requirement: Verified provider requirement projection

When a current state or authorized capability requires an external provider, the MCP
server SHALL project the logical capability, project-scoped provider alias, concrete
discovered tool name, purpose, and binding availability. It SHALL not project the
provider endpoint or credentials.

#### Scenario: Return an active provider requirement

- **WHEN** a client reads a state that requires a valid MCP-bound capability
- **THEN** the server returns the logical capability, verified provider alias, and tool name
- **AND** does not return a provider URL or credential.

#### Scenario: Return an unavailable provider requirement

- **WHEN** a state requires a capability whose project binding is unavailable
- **THEN** the server identifies the requirement as unavailable with a safe reason
- **AND** does not describe an arbitrary alternate provider target.
