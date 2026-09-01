# project-mcp-binding Specification

## Purpose

Bind logical state capabilities to verified MCP tools in the owning project so a
workflow can declare its external dependency without storing endpoints or secrets.

## Requirements

### Requirement: Validated project MCP binding

The platform SHALL allow a logical capability to bind to exactly one active project
MCP connection and one enabled, currently discovered tool from that connection. The
binding SHALL retain the logical capability identity while resolving the connection
alias and tool name at runtime.

#### Scenario: Create a valid binding

- **WHEN** an operator selects an active project connection and enabled discovered tool
- **THEN** the platform saves the capability binding
- **AND** associates it only with the workflow's project.

#### Scenario: Reject raw provider target

- **WHEN** a request attempts to bind a capability using a raw URL or unverified tool name
- **THEN** the platform rejects the request
- **AND** does not create a provider binding.

### Requirement: Binding availability validation

The platform SHALL identify a binding as unavailable when its connection is disabled,
its tool is disabled or removed, or its discovery state is no longer valid. A workflow
with a required unavailable binding SHALL not be published until corrected.

#### Scenario: Publish with removed tool

- **WHEN** an operator attempts to publish a workflow whose required capability references a removed provider tool
- **THEN** the platform rejects publication with the affected capability and connection alias
- **AND** leaves the existing published version unchanged.

### Requirement: Project isolation

The platform SHALL reject a capability binding that references a connection or tool
outside the workflow's project, even when both projects belong to the same tenant.

#### Scenario: Bind from another project

- **WHEN** an operator submits a connection identifier owned by a different project
- **THEN** the platform returns not found or validation failure
- **AND** does not disclose the other project's connection metadata.
