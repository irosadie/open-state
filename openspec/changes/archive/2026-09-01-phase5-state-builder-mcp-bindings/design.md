## Context

Phase 2 currently projects `providerServer` and `tool` from capability metadata.
Phases 3–4 introduce project connections and trusted tool records. A logical
capability can be reused across projects, so binding metadata must not make a global
capability point at one project's endpoint.

## Goals / Non-Goals

**Goals:**

- Replace free-text provider targets with project-validated bindings.
- Keep the logical capability name stable in workflow state definitions.
- Make invalid/provider-drift conditions visible before publish and runtime.

**Non-Goals:**

- Executing the provider or changing provider credentials from State Builder.

## Decisions

### D1 — Use a separate project capability binding

Create `project_capability_mcp_bindings` keyed by `(project_id, capability_id)` and
referencing one `mcp_connection_id` and `mcp_discovered_tool_id`. The service verifies
that all referenced records belong to the same tenant/project and that the connection
and tool are active. State definitions retain the logical capability identifier.

This permits the same capability name to map to different providers in different
projects without duplicating capability definitions.

### D2 — Treat existing alias/tool fields as legacy fallback only

Keep the Phase 2 `provider_id`/`provider_tool` columns for compatibility and migration,
but a new or updated project binding is authoritative. Missing bindings become explicit
`MISSING_MAPPING` status; no heuristic uses an alias or tool text to select a connection.

### D3 — Validate at authoring, publishing, and projection time

Builder APIs load only eligible connection/tool pairs. Save validates referential scope;
publish rejects required unavailable bindings. State MCP re-resolves the binding when
projecting requirements, returning safe availability status rather than endpoints.

### D4 — Keep State Builder API integration layer clean

Add project MCP catalog schemas/types/constants and React Query hooks. Capability form
controls consume hooks rather than direct HTTP, and use connection/tool selectors that
cannot create arbitrary provider targets.

## Risks / Trade-offs

- **[A tool removal blocks a workflow publish]** → this is intentional; operators
  receive the exact capability/connection alias and can refresh/rebind.
- **[Existing definitions are incomplete]** → migration marks them unavailable; seed
  data and a targeted operator backfill provide a clear path rather than guessed data.
- **[Catalog refresh changes a tool schema]** → fingerprint/status surfaces the drift;
  schema compatibility is rechecked when publishing and before Phase 6 execution.

## Migration Plan

1. Add project capability binding persistence and scope constraints.
2. Expose binding validation through capability and Builder APIs.
3. Migrate State Builder form to selectors and add preview/status display.
4. Update State MCP provider projection to prefer the verified binding.
5. Backfill demo data only where an exact project connection/tool match exists; leave
   all ambiguous mappings visible as missing.
