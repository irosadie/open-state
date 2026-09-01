## ADDED Requirements

### Requirement: State Builder MCP binding selection

The State Builder SHALL load the active project's verified MCP connections and enabled
discovered tools through its API hooks. It SHALL let an operator select a connection
and tool for a capability and persist only the validated project binding.

#### Scenario: Select a provider tool in State Builder

- **WHEN** an operator configures a capability in a project's State Builder
- **THEN** the builder presents only that project's active connections and enabled tools
- **AND** saves the selected connection/tool binding through the Builder API.

#### Scenario: No eligible provider tools

- **WHEN** the active project has no eligible discovered tools
- **THEN** the builder explains that an MCP connection must be registered and refreshed
- **AND** does not offer free-text endpoint or tool entry.

### Requirement: Binding health feedback

The State Builder SHALL render an actionable validation status for every capability
binding that is stale, disabled, removed, or missing before save or publish.

#### Scenario: Binding becomes unavailable

- **WHEN** a previously selected tool is disabled or removed
- **THEN** the builder identifies the affected capability and provider alias
- **AND** prevents publication until the binding is corrected.
