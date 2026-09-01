## Context

See Phase 3 for connection ownership. A connection test proves only that a session can
be established; it intentionally does not provide an authoritative provider tool list.

## Goals / Non-Goals

**Goals:**

- Capture a verified catalog that later phases can bind and execute safely.
- Detect provider drift without silently changing authored workflows.
- Make refresh safe, observable, and free of business-tool side effects.

**Non-Goals:**

- Running provider tools, automatic capability remapping, or remote endpoint policy.

## Decisions

### D1 — Persist discovered tools as catalog snapshots

Add `mcp_discovered_tools` scoped to a project connection with tool name, description,
input schema JSON, safe annotations, enabled state, availability status, last seen
timestamp, and stable schema fingerprint. A `(connection_id, tool_name)` key maintains
one current record per provider tool while a discovery-run/audit record preserves the
outcome and hash for diagnostics.

### D2 — Discover with MCP initialization plus `tools/list` only

The discovery adapter uses the Phase 3 connection configuration, initializes the MCP
session, then sends `tools/list`. It never calls `tools/call`; a provider that requires
side effects to list tools is considered incompatible. All raw transport errors flow
through a redactor/classifier before persistence or UI output.

### D3 — Refresh is an atomic reconciliation

On successful discovery, upsert returned tools, update changed fingerprints, and mark
previously known but absent tools `removed`. Keep removed records and existing bindings
intact for validation/audit; new bindings can select only enabled, present tools. On
failure, do not mutate the prior catalog or its fingerprint.

**Alternative considered:** delete and recreate the catalog on every refresh. Rejected
because it loses binding identity and cannot distinguish provider drift from an outage.

### D4 — Extend the Phase 3 page compositionally

Keep connection lifecycle controls in Phase 3 and add a catalog panel/detail route
using dedicated discovery endpoints/hooks. Refresh and enablement operations are
explicit actions; no page load contacts the provider.

## Risks / Trade-offs

- **[Large or adversarial schemas]** → impose payload/schema-size limits and retain a
  safe validation error rather than arbitrary provider content.
- **[Tool descriptions contain sensitive data]** → treat the catalog as untrusted
  provider metadata and redact configured patterns before persistence/display.
- **[A provider changes tools between refreshes]** → Phase 5 validates binding health
  on save/publish and Phase 6 revalidates before execution.

## Migration Plan

1. Add catalog, discovery-run, fingerprint, and status persistence.
2. Implement transport-backed discovery service and redacted error classification.
3. Add list/refresh/enablement APIs and Admin UI catalog panel.
4. Roll back by disabling discovery actions; existing connection records remain valid.
