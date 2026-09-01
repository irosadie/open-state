## Context

See `proposal.md` for the motivation. OpenState already separates logical
capabilities from provider execution through `CapabilityProvider`, and its MCP
client adapter can invoke a mapped tool over Streamable HTTP. The in-process
`FIXTURE_FILE` provider remains useful for fast engine tests but cannot verify a
separate MCP server's initialization, tool discovery, transport behavior, or
stateful write semantics.

## Goals / Non-Goals

**Goals:**

- Add a small, standalone data-plane MCP provider that behaves like a third-party
  provider in local development and integration tests.
- Make provider behavior repeatable through versioned JSON scenarios, including
  successful reads, stateful booking writes, validation failures, business failures,
  and latency.
- Exercise both a generic MCP client (tool discovery) and OpenState's existing MCP
  client adapter (mapped invocation) against the same server.

**Non-Goals:**

- Making this provider part of the default production topology or sharing OpenState's
  database.
- Replacing the capability resolver, provider binding, or in-process fixture provider.
- Defining the direct LLM-to-provider authorization-grant protocol. A later change
  must add signed, short-lived grants and provider-side verification before that
  topology can be treated as enforced.

## Decisions

### Standalone Bun/TypeScript workspace app

Create `apps/mcp-provider-mock` as a Bun workspace app using the official TypeScript
MCP SDK and Streamable HTTP transport. It will have independent `dev`, `build`, and
`test` scripts and will be started explicitly (for example through a filtered root
script), rather than being added to the default `bun run dev` stack.

This keeps the mock visibly separate from the OpenState MCP control plane and makes
its network contract representative of a third-party provider. Reusing the Go API
binary would blur that boundary; a REST JSON server would not exercise MCP
initialization or tool discovery.

### JSON scenarios are configuration, not the server protocol

The app will load one JSON scenario selected by environment variable or command-line
configuration. A scenario declares provider identity, tools, JSON input schemas,
initial in-memory data, and per-tool outcomes. Startup validates the scenario before
exposing readiness.

The JSON file is intentionally a data source; clients still interact only through
MCP. This permits new provider examples without embedding them in server code while
preserving a real MCP integration boundary.

### Stateful padel reference fixture

Ship a padel fixture containing `padel.cek_available` and `padel.create_booking`.
Availability reads from scenario state; a successful booking reserves one slot in
that process's memory and returns a deterministic reference. Restarting the process
reinitializes the fixture, which gives each test run a clean baseline without a
database or cleanup endpoint.

This is more representative than static response fixtures while remaining
deterministic. A full generic workflow engine inside the mock is rejected because
workflow decisions belong to OpenState, not the provider.

### Explicit, testable failure modes

Scenario outcomes will support successful data, a returned MCP tool error, and a
bounded delay. Tests can combine the delay with a caller timeout to validate timeout
handling; a separate unreachable-endpoint test remains the correct way to validate
connection-unavailable behavior in the OpenState adapter.

This avoids inventing nonstandard MCP transport failure semantics just to simulate
every OpenState error category.

### Tests at two protocol boundaries

Provider-app tests will verify fixture validation, MCP tool discovery, schema
rejection, padel read/write behavior, and reset-on-restart semantics. API integration
tests will run the provider mock as a process and use the existing `MCPProvider`
adapter to invoke mapped tools and normalize results.

The two layers keep provider protocol failures distinguishable from OpenState's
capability-resolution and state-authorization tests, which remain in the API domain.

## Risks / Trade-offs

- [Mock diverges from production provider behavior] → Keep fixtures intentionally
  small, validate them, and cover MCP discovery/invocation with a standard client.
- [Stateful tests leak bookings across cases] → Start each test with a fresh process
  and scenario state; no mutable fixture files are written.
- [Tests become timing-sensitive] → Use bounded configured delays and caller-side
  deadlines; do not assert wall-clock durations beyond timeout behavior.
- [Direct LLM access bypasses state policy] → Treat provider mock as an integration
  target only until a separate signed-grant protocol and provider verification are
  specified and implemented.

## Migration Plan

1. Add the workspace app, fixture contract, and padel scenario without changing any
   production API path.
2. Add an opt-in local command and provider-app tests.
3. Add API integration tests that target the mock on a test port.
4. Roll back by removing the opt-in workspace app and tests; existing in-process
   fixture and production capability paths are unaffected.
