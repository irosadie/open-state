## Context

See `proposal.md` for motivation. The provider mock already has scenario-selected MCP tool discovery and a padel-only in-memory booking store. Food-order and doctor tools are static fixture responses.

## Goals / Non-Goals

**Goals:**

- Provide deterministic, testable write lifecycles for all three provider scenarios.
- Preserve English scenario payloads and MCP JSON-schema input validation.
- Keep all data process-local and fixture-driven.

**Non-Goals:**

- Model every production edge case or provide durable storage.
- Change the OpenState capability authorization layer.

## Decisions

### Use domain-specific in-memory stores behind common MCP execution

Padel, food-order, and doctor operations have different lifecycles, so each scenario has a focused in-memory store while the MCP server continues to register tools from the active scenario. Tool metadata identifies the requested operation and delegates to the relevant store.

Alternative considered: encode all mutations in generic JSON-path operations. Rejected because lifecycle invariants and business errors would be opaque and difficult to verify.

### Require explicit references for later lifecycle operations

Cart, booking, order, reservation, and payment identifiers are accepted as tool inputs. The mock creates deterministic sequential references and validates ownership/existence before mutation.

Alternative considered: infer the latest resource implicitly. Rejected because it hides client behavior and cannot model invalid references.

### Validate writes at the MCP boundary

Each write tool declares its required JSON-schema input; domain stores enforce availability and transition rules after schema validation. The curl smoke test covers each write-then-read path using only MCP HTTP requests.

## Risks / Trade-offs

- [Fixture complexity grows] → Keep initial state small and deterministic, and test one full lifecycle per domain.
- [Static tools can drift from dynamic state] → Route all write-adjacent reads through the domain stores.
- [Test server cleanup fails] → Retain bounded local smoke processes and shell traps.

## Migration Plan

1. Extend scenario operations and data schemas for food-order and doctor state.
2. Implement domain stores and switch affected tools from static results to stateful execution.
3. Update fixtures, SDK tests, and curl smoke flows.
4. Validate resets and MCP-visible failures.
