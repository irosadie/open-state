## Purpose

Give project operators one safe admin page for registering and operating the external
MCP servers available to the selected project's workflows.

## ADDED Requirements

### Requirement: Project MCP Connections page

The web application SHALL provide a project-scoped MCP Connections page reachable
from the Admin Console. The page SHALL show only the active project's connections,
their aliases, transport, authentication status, enablement, and latest safe test
result.

#### Scenario: Open the connections page

- **WHEN** an authorized user opens MCP Connections for a project
- **THEN** the page displays the project's connections and their safe operational status
- **AND** does not display secret values or connections from another project.

### Requirement: Safe connection form

The page SHALL let authorized users create and edit a connection with name, alias,
transport, URL or trusted STDIO configuration, and authentication mode. It SHALL
validate required fields before submission and make bearer/OAuth values write-only.

#### Scenario: Save a bearer connection

- **WHEN** an operator enters a valid bearer token and saves a connection
- **THEN** the page submits it only to the protected write operation
- **AND** renders a configured status rather than the token after save.

### Requirement: Lifecycle controls and feedback

The page SHALL provide deliberate controls to test, enable, disable, and delete a
connection, with confirmation for destructive or availability-affecting actions and
clear server validation or authorization feedback.

#### Scenario: Disable a connection

- **WHEN** an operator confirms disabling a connection
- **THEN** the page refreshes the connection state after server success
- **AND** presents the reason if the server rejects the operation.
