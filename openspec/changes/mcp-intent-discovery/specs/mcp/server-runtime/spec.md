## MODIFIED Requirements

### Requirement: Intent resolution tool

The MCP server SHALL expose an intent-resolution tool that accepts a canonical intent identifier within a tenant and project and returns the classified intent and its mapped workflow to the LLM. Runtime state SHALL continue to be obtained through the active-workflow or context tools.

#### Scenario: Resolve intent

- **WHEN** the LLM calls the intent-resolution tool with a canonical intent such as `BOOKING_PADEL`, tenant, and project
- **THEN** the server resolves the canonical intent through the tenant/project-scoped intent catalog
- **AND** returns the canonical intent and mapped workflow

#### Scenario: Unknown intent

- **WHEN** the LLM calls the intent-resolution tool with an identifier that is not routable in the requested tenant/project
- **THEN** the server returns a not-found error
- **AND** does not resolve a workflow by arbitrary workflow ID or slug
