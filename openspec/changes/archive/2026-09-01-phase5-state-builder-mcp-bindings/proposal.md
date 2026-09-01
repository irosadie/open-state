## Why

State Builder currently has no validated way to bind a logical capability to a real project MCP connection and discovered tool. Operators need a picker that makes the exact provider dependency visible in the state definition without entering raw URLs or guessing tool names.

## What Changes

- Add project MCP connection and discovered-tool selection to capability authoring in State Builder.
- Persist a capability binding as `capability → project MCP connection → tool`, while preserving the logical capability as the state definition's stable reference.
- Filter selections to active connections and enabled tools in the workflow's project.
- Validate bindings on save/publish and surface actionable status when the connection, tool, or catalog is stale, disabled, or removed.
- Display the provider alias, tool, purpose, and current verification status in state previews and runtime requirement projections.

## Capabilities

### New Capabilities

- `capability/project-mcp-binding`: Project-scoped binding and validation contract between a logical capability and a discovered MCP tool.

### Modified Capabilities

- `web/state-builder-api`: State Builder loads the project MCP catalog and persists validated provider-tool bindings.
- `mcp/orchestrator-tools`: State/runtime capability projections resolve a binding through a project MCP connection rather than an unverified free-text alias.

## Impact

- Capability registry/binding persistence, validation services, builder API, MCP projection, shared schemas/types, and State Builder UI.
- Depends on Phases 3–4.
- Existing capability rows need an explicit migration/backfill state so missing bindings fail visibly rather than pretending to be routable.

## Non-goals

- Executing a provider tool from Builder.
- Opening a provider connection from the LLM directly.
- Routing provider calls through an OpenState gateway (Phase 6).
