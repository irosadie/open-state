## Why

The MCP server currently exposes `resolve_intent`, but the backend does not expose a durable, tenant/project-scoped catalog of canonical intents. That leaves the LLM without the names, descriptions, and example utterances it needs to map a message such as “saya mau order lapangan” to `BOOKING_PADEL` reliably.

## What Changes

- Add a persisted intent catalog owned by a project, with a canonical ID, description, example utterances, and an explicit workflow mapping.
- Add a read-only MCP `list_intents` tool that returns the catalog for the requested tenant and project.
- Make `resolve_intent` resolve canonical catalog IDs such as `BOOKING_PADEL` through the persisted mapping and return the mapped workflow; runtime state remains available through the existing workflow/context tools.
- Seed the demo projects with canonical intent records, including examples for padel court booking, food ordering, and doctor ordering.
- Add automated coverage for tenant/project isolation, intent discovery, canonical resolution, and the live MCP tool contract.

## Capabilities

### New Capabilities

- `mcp/intent-discovery`: Provide the LLM with a scoped intent catalog and natural-language examples for choosing a canonical intent before workflow execution.

### Modified Capabilities

- `mcp/server-runtime`: Update intent resolution to use the canonical, tenant/project-scoped intent catalog rather than treating an arbitrary workflow ID or slug as the intent.

## Impact

- Backend database: new intent persistence and migration/query code.
- Backend domain/application layers: intent entity, repository contract, catalog service, and workflow mapping lookup.
- MCP interface: one new read-only tool and an updated `resolve_intent` contract.
- Demo seed data and backend tests.
- No frontend behavior is required for the first slice; the MCP catalog becomes the backend source of truth for LLM routing.

## Non-goals

- Performing LLM classification inside OpenState; the LLM still chooses from the catalog and the State Engine remains authoritative.
- Replacing workflow/state authorization or exposing the global capability registry.
- Adding project-management UI in this change.
