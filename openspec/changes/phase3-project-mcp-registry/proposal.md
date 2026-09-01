## Why

OpenState currently stores a provider alias on a capability but has no project-owned record for the external MCP server behind that alias. Operators therefore cannot safely register, inspect, test, or govern the MCP connections that a project's workflows depend on.

## What Changes

- Add a project-scoped MCP Connection Registry under the tenant/project hierarchy.
- Provide an admin page for creating, editing, disabling, listing, and testing an external MCP connection.
- Support Streamable HTTP as the primary remote transport, legacy SSE compatibility, and trusted-local STDIO connections.
- Capture an authentication mode of none, bearer token, or OAuth configuration; credential values are stored as secret references and never returned to the browser or State MCP responses.
- Require a unique, stable connection alias within a project so workflows refer to an alias rather than an endpoint.
- Apply tenant and project RBAC to every registry operation.

## Capabilities

### New Capabilities

- `mcp/project-connection-registry`: Project-isolated lifecycle and connection-test contract for external MCP servers.
- `web/project-mcp-connections`: Admin experience for managing a project's external MCP connections.

### Modified Capabilities

- `web/admin-console-management`: Add the project-scoped MCP Connections destination to the admin navigation and project workflow.

## Impact

- PostgreSQL migration, sqlc queries, domain entities, repositories, services, HTTP API, RBAC, and audit records.
- Project admin UI, shared API schemas/types/hooks, and navigation.
- Existing capability provider aliases become resolvable against an owned project resource; no workflow endpoint is introduced.

## Non-goals

- Discovering or persisting the remote server's tools (Phase 4).
- Selecting a provider tool in State Builder (Phase 5).
- Calling providers through OpenState or enforcing gateway policy (Phase 6).
- Tenant-shared reusable connections; the first release is intentionally project-owned.
