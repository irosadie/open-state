## Why

The Admin Console shows `Project` in the tenant-to-state path, but the step is
currently a non-clickable placeholder. Operators therefore cannot choose the
project that owns the next Intent, Workflow, and State Builder views.

## What Changes

- Add a tenant-scoped Project inventory at `/admin/projects`.
- Make the Project step clickable from the shared Admin Console flow guide,
  navigation, and overview.
- Let a selected project scope Intent, Workflow, and State Builder routes using
  an explicit `projectId` URL parameter.
- Reuse the existing project discovery API as a read-only flow boundary.

## Capabilities

### New Capabilities

- `web/admin-project-flow`: Provide a clickable tenant-to-project-to-state flow
  and a permission-aware project inventory.

### Modified Capabilities

- `web/admin-console-management`: Make Project a real step in the shared flow
  and preserve the selected project as the scope for downstream pages.

## Impact

- Frontend Admin Console routes, navigation, flow guide, project inventory,
  query-string scope propagation, and Builder persistence headers.
- The existing `GET /api/projects` read endpoint becomes a shared flow inventory
  endpoint protected by `workflow:read`.
- No project CRUD, tenant switching, or database migration is introduced.

## Non-goals

- No create, edit, archive, or delete project UI.
- No new permission family; project discovery uses `workflow:read`.
- No change to MCP tool names, state definitions, or provider mock behavior.
