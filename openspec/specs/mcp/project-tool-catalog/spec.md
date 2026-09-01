## Purpose

Maintain a verified, safe, project-scoped catalog of the tools an external MCP
connection currently exposes, so later bindings never rely on guessed tool names.

## Requirements

### Requirement: Explicit tool discovery

The platform SHALL let an authorized operator refresh an enabled registered MCP
connection's tool catalog by initializing the MCP session and requesting its tool
list. Discovery SHALL not invoke any provider business tool.

#### Scenario: Discover provider tools

- **WHEN** an operator refreshes tools for a reachable enabled connection
- **THEN** the platform records the tools returned by the MCP server
- **AND** does not invoke a provider business tool.

#### Scenario: Discovery fails

- **WHEN** the provider cannot initialize or return its tool list
- **THEN** the platform preserves the last successful catalog
- **AND** returns a classified, redacted discovery failure.

### Requirement: Sanitized tool snapshots

For every discovered tool, the platform SHALL retain the connection scope, tool name,
description, input schema, safe annotations, discovery timestamp, and catalog
fingerprint. Tool names SHALL be unique within one connection catalog.

#### Scenario: Read a tool catalog

- **WHEN** an authorized user reads a project's connection catalog
- **THEN** the platform returns the latest sanitized tool records and their verification status
- **AND** does not return connection credentials or raw transport headers.

### Requirement: Refresh drift detection

The platform SHALL compare a successful refresh with the previous catalog and mark
new, changed, and removed tools. Refreshing SHALL not automatically rewrite existing
workflow or capability bindings.

#### Scenario: Provider removes a tool

- **WHEN** a refreshed catalog no longer contains a previously discovered tool
- **THEN** the platform marks that tool unavailable for new bindings
- **AND** retains enough safe status to identify existing affected bindings.

### Requirement: Tool enablement

The platform SHALL allow an authorized operator to enable or disable individual
discovered tools. Disabled and unavailable tools SHALL remain inspectable but SHALL
not be eligible for new state bindings or gateway execution.

#### Scenario: Disable a discovered tool

- **WHEN** an operator disables a discovered tool
- **THEN** the platform marks the tool unavailable for new bindings and execution
- **AND** records a redacted audit event.
