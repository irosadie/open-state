## Context

The product hierarchy is `Tenant → Project → Intent → Workflow → State`.
Workflow and Intent APIs already accept an optional project scope, while the
State Builder API already supports `X-Project-ID`. The missing piece is the UI
route and the propagation of that scope between the pages.

## Decisions

### 1. Use a dedicated read-only Project inventory

Add `/admin/projects` using the existing `GET /api/projects` contract. The page
lists projects owned by the current tenant and provides explicit links into
Intents and Workflows. The API remains read-only and does not create a default
project as a side effect.

### 2. Use `projectId` as the shareable scope

Selecting a project navigates to `/admin/intents?projectId={id}` or
`/admin/workflows?projectId={id}`. Those pages pass the value to their existing
React Query hooks. Builder links preserve the same query parameter, and the
Builder store sends it as `X-Project-ID` for load, save, version, and publish
operations.

This keeps scope explicit, bookmarkable, and safe across refreshes without
introducing a second global project state or tenant switcher.

### 3. Reuse `workflow:read` for project visibility

Projects are the scope for workflow authoring and intent routing, so the
existing `workflow:read` permission is used for the inventory route and API.
Final tenant and project authorization remains enforced by the backend.

### 4. Keep State Builder ownership intact

The Admin Console only selects a project and links to Builder. It does not
duplicate state authoring, save, publish, or version controls.

## Risks and Mitigations

- A stale or invalid `projectId` is rejected by the existing tenant-scoped API
  services; the UI displays the server error and offers retry.
- A direct Builder URL without `projectId` keeps the existing default-project
  behavior for backwards compatibility.
- Existing API-key users retain access because their roles already include the
  workflow read permission; the API-key page still validates the selected IDs
  against the tenant-owned project list.
