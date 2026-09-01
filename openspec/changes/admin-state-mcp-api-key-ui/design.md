## Context

The backend API-key lifecycle is already implemented at `GET`/`POST
/api/api-keys` and `POST /api/api-keys/{id}/revoke`. The browser uses the
Next.js BFF proxy and the configured tenant id, while the server derives the
authenticated human actor from the session. The raw machine key is returned
only by a successful create response.

## Goals / Non-Goals

**Goals:**

- Make local and hosted State MCP provisioning possible from the Admin Console.
- Keep the page consistent with existing permission-aware admin surfaces.
- Validate every API response and create payload before rendering or sending it.
- Make one-time secret handling explicit and recoverable only by creating a new
  key.

**Non-Goals:**

- Do not add project management in this slice. Project discovery is read-only.
- Do not expose verifier material, raw secrets from list responses, or
  `MCP_API_KEY_PEPPER` to the browser.

## Decisions

### 1. Dedicated `/admin/api-keys` route with `/admin` entry points

Use the established Admin Console route shape and add an overview card plus a
sidebar item. The route is separate from tenant settings because key lifecycle
is machine access management, not tenant profile editing.

### 2. Reuse API-key RBAC permissions

Gate route/list visibility with `api_key:read`, creation with
`api_key:create`, and revocation with `api_key:revoke`. The backend remains the
final authority; UI checks only control presentation and query/mutation startup.

### 3. Use a tenant-scoped project-ID selector

The existing API requires at least one allowed project. Add a read-only
`GET /api/projects` endpoint protected by `workflow:read`, using the current
tenant header and the existing `ListByTenant` repository method. The endpoint
is shared by the Admin Console project flow and the API-key form. The form uses
those results as a multi-select, and the default project is a single select
constrained to the selected projects. The backend still validates project
ownership during key creation.

### 4. One-time secret card

After creation, render the raw key in a dedicated warning panel with a copy
button. Keep it only in component state, clear it when dismissed, and never
populate it from list data. A user can create a replacement if it is lost.

### 5. Lightweight metadata table

List key prefix, project count, scopes, expiration, last-used time, and status.
Revoke requires confirmation and invalidates the list query after success.
Loading, empty, unauthorized, API error, and retry states follow existing
Admin Console patterns.

## Risks / Trade-offs

- **[Risk] A tenant has no projects.** → **Mitigation:** show an explicit empty
  state in the selector and keep API-key creation validation authoritative.
- **[Risk] A secret is copied into logs or screenshots.** → **Mitigation:** show
  it only once, label it as sensitive, avoid logging, and provide copy instead
  of repeated display.
- **[Risk] UI permissions drift from backend roles.** → **Mitigation:** use the
  exact permission strings already used by the backend and add route/action
  policy coverage.

## Migration Plan

1. Add frontend contract files and hooks for the already-deployed API.
2. Add the route, overview/navigation entry, and key management interactions.
3. Run schema tests, frontend tests, typecheck, lint, and production build.

Rollback is additive: remove the page and navigation entry without changing
existing API-key records or State MCP authentication.
