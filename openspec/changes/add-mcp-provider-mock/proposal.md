## Why

OpenState needs a deterministic, networked stand-in for third-party data MCP
providers so the state-controlled capability path can be exercised without
connecting to live systems. The existing in-process JSON fixture provider is
useful for engine tests, but it does not validate MCP discovery, Streamable HTTP
transport, mapped tool calls, or provider-level failures.

## What Changes

- Add `apps/mcp-provider-mock`, a development and test-only Streamable HTTP MCP
  server that simulates an external read/write data provider.
- Drive the mock server from validated JSON scenario files that declare tools,
  input schemas, deterministic responses, and classified failure behaviors.
- Provide a padel fixture with `padel.cek_available` and a write-capable booking
  tool so state-controlled capability execution has a representative integration
  target.
- Expose MCP tool discovery plus live and ready health endpoints for local
  orchestration and integration tests.
- Add integration coverage proving the existing OpenState MCP client adapter can
  initialize and invoke a mapped provider tool, while a standard MCP client can
  discover its exposed tool contract and exercise representative failure paths.

## Non-goals

- Replacing the existing in-process `FIXTURE_FILE` provider used by fast unit and
  sandbox tests.
- Connecting a production LLM directly to an external data MCP or issuing a
  production state-execution grant for that topology.
- Storing real provider credentials, customer data, or production integrations in
  the mock server.
- Changing workflow definitions, capability bindings, or the public State MCP
  tool contract.

## Capabilities

### New Capabilities

- `mcp/provider-mock`: A configuration-driven MCP provider simulator for local
  development and integration tests, including tool discovery, deterministic
  responses, and classified provider failures.

### Modified Capabilities

- None.

## Impact

- New Bun workspace app under `apps/mcp-provider-mock` and its JSON fixtures,
  tests, and local development scripts.
- Root workspace/Turbo development commands will gain an opt-in provider-mock
  process rather than making test data services part of production runtime.
- `apps/api/internal/infrastructure/capability/MCPProvider` integration tests gain
  a real MCP transport target; core state-engine and capability contracts remain
  unchanged.
- Aligns with MAIN_PRD §§59-64, 170, and 172: the State Engine authorizes the
  logical capability, while a separate MCP provider executes the data operation.
