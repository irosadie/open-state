## Why

The provider mock currently demonstrates only a small padel scenario while the API fixture catalog still owns the representative padel, food-order, and doctor provider responses. This duplicates the integration boundary that clients need to evaluate: real MCP tool discovery and invocation against a configured third-party provider.

## What Changes

- Move the existing padel, food-order, and doctor fixture capability responses into domain-specific MCP provider scenarios under `apps/mcp-provider-mock/fixtures`.
- Expose every migrated capability as a discoverable MCP tool with the fixture response returned through `tools/call`.
- Retire the migrated entries from the API JSON fixture catalog so provider behavior has one source of truth.
- Add repeatable `curl`-based MCP protocol smoke tests that initialize, discover tools, and invoke representative padel, food-order, and doctor tools.

## Capabilities

### New Capabilities

- `mcp/provider-mock-fixture-catalog`: Configured provider mock scenarios expose the migrated padel, food-order, and doctor fixture catalogs through Streamable HTTP MCP and provide protocol-level curl verification.

### Modified Capabilities

- None.

## Impact

- `apps/mcp-provider-mock`: scenario fixtures, generic static fixture response handling, test coverage, and usage documentation.
- `apps/api/testdata/capability_fixtures.json`: migrated provider-domain entries are removed while unrelated generic fixtures remain.
- Root scripts or test helpers: a curl smoke-test command starts the mock, sends MCP JSON-RPC requests, and cleans up its local process.

## Non-goals

- This does not create a production provider, alter State MCP authorization, or change workflow state transitions.
- This does not make the provider mock stateful for every migrated domain; existing dynamic padel booking behavior remains its own scenario concern.
