## Why

The provider mock currently returns static fixture payloads for most actions that are writes, so an MCP client cannot exercise the real effects and failure cases of cart, order, appointment, booking, and payment flows. Stateful writes are needed to make the mock a credible local third-party provider boundary.

## What Changes

- Make every padel, food-order, and doctor MCP write tool mutate isolated in-memory scenario state.
- Make the corresponding read tools return the state created by prior calls in the same provider process.
- Define deterministic identifiers, valid state transitions, and MCP tool errors for invalid write requests and duplicate or invalid transitions.
- Extend SDK and curl protocol coverage to verify write-then-read flows and process-reset behavior.

## Capabilities

### New Capabilities

- `mcp/provider-mock-stateful-writes`: Provider mock scenarios support deterministic stateful write operations for padel booking/payment, food cart/order/payment, and doctor reservation/appointment/payment lifecycles.

### Modified Capabilities

- None.

## Impact

- `apps/mcp-provider-mock`: scenario data contract, in-memory stores, MCP tool execution, input schemas, fixtures, unit tests, and curl smoke tests.
- No production database or OpenState State MCP authorization behavior changes.

## Non-goals

- Persistence across provider mock restarts or cross-process sharing.
- Production payment, inventory, medical scheduling, or real provider credential integration.
- Bypassing OpenState state authorization; the provider mock only executes a call it receives.
