## Context

See `proposal.md` and the Phase 2 state-provider contract. Phase 2 represents a
provider as free-text alias/tool metadata, so it cannot establish project ownership,
test a real provider, or protect connection credentials. This phase introduces the
control-plane resource only; it does not load tools or forward calls.

## Goals / Non-Goals

**Goals:**

- Make an external MCP connection a first-class project resource.
- Keep endpoints and credentials out of workflow definitions and MCP responses.
- Provide a thin, testable administrative API and UI using existing Clean Architecture
  and shared frontend contract conventions.

**Non-Goals:**

- Tool discovery, tool selection, provider invocation, tenant-shared connection reuse,
  and full OAuth authorization are deferred to later phases.

## Decisions

### D1 — Project owns the connection; tenant remains the isolation boundary

Create `mcp_connections` with tenant ID, project ID, display name, normalized alias,
transport, safe endpoint/STDIO profile fields, authentication mode, credential
reference, enabled state, latest test metadata, and audit metadata. Enforce a unique
`(project_id, alias)` constraint and derive tenant scope through the project relation.

This supports the requested `tenant → project → MCP connections` hierarchy. A future
tenant-shared resource can be attached to projects without weakening the initial
project isolation.

**Alternative considered:** store connections at tenant scope and use a project list.
Rejected for v1 because accidental cross-project access is easier and State Builder
needs a single unambiguous project catalog.

### D2 — Store a connection alias and endpoint separately from capabilities

Capability/workflow definitions will later reference an owned connection ID through a
binding. The connection's alias is human-readable and stable for State MCP projections;
the endpoint stays inside the connection resource. No API accepts an endpoint as part
of a workflow or capability mutation.

### D3 — Use an authentication descriptor and secret reference

The connection record stores `auth_type` and a write-only credential payload resolved
to a protected `credential_reference`. Bearer tokens are accepted only on create or
explicit replacement; UI/API reads return `configured`, `missing`, or `action_required`.
OAuth in this phase stores connection configuration/status but defers interactive grant
and refresh operations to Phase 7.

**Alternative considered:** store encrypted token columns directly in the business
table. Rejected because it couples persistence to one vault/encryption deployment and
makes redaction/error handling harder.

### D4 — Test means handshake, not discovery

The test service uses a transport adapter to establish an MCP session and verify the
server identity/protocol response. It records a redacted `ready`, `failed`, or
`disabled` health result but does not call `tools/list`; discovery is separately
auditable in Phase 4.

### D5 — Explicit permissions and audit actions

Add project-scoped `mcp_connection:read` and `mcp_connection:manage` permissions (or
map them to the established project capability-admin role during migration). HTTP
handlers derive tenant/project solely from authenticated request context. Service
commands emit redacted audit events after successful mutations/tests.

### D6 — One project-context admin page

Add a project MCP Connections page under Admin Console navigation. Backend DTOs,
Zod/types/constants/hooks follow the repository's API-integration pattern. The form
uses transport-dependent inputs, write-only secret fields, server-side validation,
and confirmation for disable/delete.

## Risks / Trade-offs

- **[SSRF or unsafe STDIO configuration in a registry]** → Phase 3 validates shape
  and keeps STDIO behind trusted profiles; complete egress/process enforcement is
  mandatory in Phase 7 before production gateway mode.
- **[OAuth appears selectable before full grant support]** → show explicit
  `action_required` status and do not report a connection ready until authorization is
  completed in Phase 7.
- **[Existing free-text provider mappings]** → retain them as legacy compatibility
  metadata; do not auto-convert a mapping to a connection.

## Migration Plan

1. Add connection, credential-reference, and safe health/audit persistence with
   tenant/project foreign keys and alias uniqueness.
2. Add repository/application/HTTP contracts and RBAC permissions.
3. Add the Admin Console page and project-aware API hooks.
4. Backfill no connection automatically; existing provider mappings retain their
   current unavailable/missing status until an operator creates a connection.
5. Roll back by disabling new routes and leaving additive records unused; no workflow
   definition is changed in this phase.
