## Purpose

Let project operators inspect and govern the discovered MCP tools behind each
registered connection without ever executing those provider tools from the console.

## ADDED Requirements

### Requirement: Tool catalog visibility

The MCP Connections experience SHALL show a connection's discovered tools, names,
descriptions, input-schema summary, enabled state, discovery timestamp, and drift
status.

#### Scenario: Inspect discovered tools

- **WHEN** an authorized operator opens an MCP connection with a successful discovery
- **THEN** the page lists its sanitized discovered tools and verification details
- **AND** does not display an action that invokes a business tool.

### Requirement: Refresh and enablement controls

The page SHALL let authorized operators explicitly refresh a tool catalog and enable
or disable individual tools. It SHALL show classified discovery errors and retain the
last safe successful view when refresh fails.

#### Scenario: Refresh failure in the UI

- **WHEN** a tool refresh fails
- **THEN** the page shows the failure status and last successful discovery timestamp
- **AND** does not discard the previously displayed catalog.
