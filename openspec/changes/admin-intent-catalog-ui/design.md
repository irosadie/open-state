## Context

The MCP intent-discovery change already persists canonical intents and exposes
`IntentService.ListIntents` for tenant/project-scoped, published mappings. The
HTTP API currently has no adapter for that service, while the Admin Console has a
four-step guide (`Tenant → Project → Workflow → Builder`) and no intent route.
Workflow listing already accepts an optional project scope and falls back to the
tenant's default project.

See `proposal.md` and the delta specs for the user-visible contract.

## Goals / Non-Goals

**Goals:**

- Reuse the persisted, published-only intent catalog for a safe Admin Console
  read endpoint.
- Make the intent-to-workflow relationship visible without duplicating Builder
  authoring behavior.
- Keep the current automatic default-project behavior understandable and
  consistent across the guide, intent page, and workflow page.
- Preserve the existing RBAC model and MCP contract.

**Non-Goals:**

- Do not add a second intent repository, a new intent permission family, or a
  separate project-management subsystem.
- Do not make a read request create or mutate a project/catalog record.
- Do not classify user messages in the API or move workflow ownership into the
  Admin Console.

## Decisions

### 1. Add a thin HTTP adapter over the existing intent catalog

Add an application/controller/route path for `GET /api/intents` that calls the
existing intent repository through an application service. The service resolves
an explicit `projectId` within the tenant or resolves the existing `default`
project when the parameter is absent, then delegates to the same published-only
catalog query used by MCP. The read path must not create a missing default
project.

Alternative considered: call the MCP endpoint from the browser. Rejected because
the Admin Console uses the authenticated API/BFF contract, and the MCP server is
an external LLM integration boundary rather than a browser data API.

### 2. Reuse `workflow:read` for catalog visibility

Protect the endpoint and UI with the existing `workflow:read` permission. Intent
records are routing metadata for workflows, and introducing `intent:read` would
require a new RBAC matrix, role migration, and permission synchronization for no
additional user value in this read-only slice.

### 3. Return a UI-ready projection

The HTTP response will wrap a `data` array and expose the catalog fields needed by
the page: `id`, `key`, `name`, `description`, `examples`, `tenantId`, `projectId`,
`workflowId`, and `workflowSlug`. `workflowId` is included so the page can link to
the existing `/state-builder/{workflowId}` route without guessing from a slug.
The MCP projection remains unchanged.

### 4. Add a dedicated Admin Console route

Create `/admin/intents` with the established App Router shape: a thin Suspense
`page.tsx` and a client `intents-page-content.tsx`. The page uses a React Query
hook through the existing axios/BFF layer, shows a compact table/card list, and
has explicit loading, error/retry, and empty states. It renders no mutation
controls. The current project is displayed as `Default Project (automatic)` and
the optional project query contract remains available for future project
selection.

Alternative considered: append intents as a panel on `/admin/workflows`. Rejected
because it keeps Intent hidden in navigation and makes the hierarchy appear to
skip a domain object.

### 5. Make the shared guide the single hierarchy source

Extend `AdminFlowGuide` to five typed steps: Tenant, Project, Intent, Workflow,
Builder. Tenant, Intent, Workflow, and Builder remain links to existing routes;
Project remains a non-link context card until project management exists. Render
the guide's `intent` current state on the new page and keep the same guide on the
overview, tenant, and workflow pages. Add the intent page to overview cards and
permission-aware shell navigation.

### 6. Document and test the contract at both boundaries

Document `GET /intents` in the repository's existing split JSON OpenAPI files,
reusing the established error schemas and adding one intent response schema.
Backend tests cover default/explicit project scope, published filtering, and
authorization. Frontend tests cover navigation visibility, the five-step guide,
catalog rendering, Builder links, and error/empty states. The API response is
validated in the frontend before it is rendered.

## Risks / Trade-offs

- **[Risk] Project selection is still unavailable.** → **Mitigation:** label the
  active project in every relevant surface and keep Project as an explicit
  context step; reserve project selector/CRUD for a separate change.
- **[Risk] A catalog mapping becomes stale when its workflow is unpublished.** →
  **Mitigation:** retain the existing SQL join/filter so only active-project,
  published-workflow mappings are returned.
- **[Risk] HTTP and MCP projections drift.** → **Mitigation:** map both from the
  same domain entity, keep the MCP contract untouched, and add response-shape
  tests for the canonical `BOOKING_PADEL` mapping.
- **[Risk] The UI can imply that listing intents edits routing.** →
  **Mitigation:** use read-only copy and provide only a link to Builder for the
  mapped workflow.

## Migration Plan

1. Add the authenticated HTTP read endpoint and OpenAPI contract using the
   already-applied `intents` table and existing catalog queries.
2. Add the frontend schema, type, hook, route, navigation entry, and hierarchy
   guide update.
3. Verify the seeded `BOOKING_PADEL` row appears with its examples and Builder
   link, then run backend/frontend tests and type/lint/build checks.

Rollback is additive: remove the Admin Console entry point and HTTP route from a
future release if needed. Do not remove the `intents` table or alter MCP routing,
because those belong to the earlier intent-discovery change.
