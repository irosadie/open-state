## Why

A saved MCP endpoint is not enough for an operator to configure a state safely: OpenState needs a verified, project-scoped catalog of the provider's actual tools and schemas. Without discovery, aliases and tool names remain free text and drift silently when a provider changes.

## What Changes

- Add a discovery action that initializes a registered MCP connection and reads its `tools/list` result.
- Persist a sanitized snapshot of each discovered tool: name, description, input schema, annotations, discovery timestamp, and catalog fingerprint.
- Show connection/discovery state, last successful check, classified errors, and tool enablement controls in the MCP Connections page.
- Let operators refresh tools manually without overwriting an active workflow definition; changed or removed tools are marked for follow-up.
- Prevent discovery from executing provider business tools or exposing credentials in errors, responses, logs, or audits.

## Capabilities

### New Capabilities

- `mcp/project-tool-catalog`: Safe discovery, persistence, refresh, and lifecycle visibility for tools exposed by a project MCP connection.
- `web/project-mcp-tool-catalog`: Operator experience for refreshing, inspecting, and enabling a project's discovered MCP tools.

### Modified Capabilities

- None.

## Impact

- MCP client/transport adapters, connection test/discovery services, database tables, APIs, audit events, and admin UI.
- Depends on Phase 3's project MCP connection registry.
- Produces the verified tool catalog consumed by State Builder in Phase 5.

## Non-goals

- Invoking business tools, writing provider data, or proxying traffic.
- Automatically changing existing capability mappings after a provider refresh.
- OAuth token lifecycle and production gateway resiliency (Phase 7).
