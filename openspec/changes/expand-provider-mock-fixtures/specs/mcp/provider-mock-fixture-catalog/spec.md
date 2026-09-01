## Purpose

Provide representative padel, food-order, and doctor provider catalogs as configurable MCP scenarios so client integrations can exercise the real provider protocol locally.

## ADDED Requirements

### Requirement: Domain fixture scenarios are available

The provider mock SHALL provide separate selectable scenarios for the migrated padel, food-order, and doctor fixture catalogs. Each scenario SHALL declare only its own MCP tools and SHALL preserve the corresponding fixture response data.

#### Scenario: Select the food-order scenario

- **WHEN** the provider mock starts with the food-order scenario selected
- **THEN** MCP tool discovery returns the food-order tools and no padel or doctor tools

#### Scenario: Invoke a migrated doctor tool

- **WHEN** an MCP client invokes a declared doctor tool with an input that satisfies its schema
- **THEN** the tool result contains the corresponding doctor fixture response data

### Requirement: Migrated provider fixtures have one runtime source

The API JSON fixture catalog SHALL no longer retain fixture entries migrated to the provider mock scenarios. Unrelated API fixtures SHALL remain available to existing API tests.

#### Scenario: Resolve a migrated provider capability

- **WHEN** a test needs a migrated padel, food-order, or doctor provider response
- **THEN** it obtains the response by invoking the matching tool on the provider mock

### Requirement: MCP protocol smoke testing is runnable with curl

The repository SHALL provide a repeatable curl-based smoke test that starts the provider mock, initializes an MCP session, discovers tools, and invokes representative tools from padel, food-order, and doctor scenarios.

#### Scenario: Run protocol smoke checks

- **WHEN** a developer runs the provider mock curl smoke-test command
- **THEN** each scenario completes MCP initialization, tool discovery, and a successful representative `tools/call` request, and the temporary mock process is cleaned up
