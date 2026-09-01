## Why

The Admin Console currently jumps from Project directly to Workflow and does not
expose the canonical intent catalog that the MCP routing flow now uses. Operators
therefore cannot see which user-facing intents exist, which examples train the
LLM's choice, or which workflow each intent maps to.

## What Changes

- Add a tenant/project-scoped read-only HTTP endpoint for the existing published
  intent catalog.
- Add an Admin Console `Intents` page that lists each canonical intent, its
  example utterances, and its mapped workflow.
- Add `Intents` to permission-aware Admin Console navigation and overview entry
  points.
- Update the shared setup guide to show the complete path:
  `Tenant → Project → Intent → Workflow → Builder`.
- Make the current project scope explicit on the intent and workflow surfaces;
  continue resolving the configured `Default Project` automatically.

## Capabilities

### New Capabilities

- `backend/intent-catalog-api`: Expose the existing canonical intent catalog to
  authenticated, tenant/project-scoped Admin Console clients through a read-only
  HTTP endpoint.
- `web/admin-intent-catalog`: Provide a permission-aware Admin Console page for
  browsing canonical intents, examples, project scope, and workflow mappings.

### Modified Capabilities

- `web/admin-console-management`: Extend the console navigation and setup guide
  so Intent is a visible step between Project and Workflow.

## Impact

- Backend HTTP DTO, service/controller wiring, route registration, and API
  contract documentation; the existing intent repository and MCP contract remain
  the source of truth.
- Frontend Admin Console navigation, flow guide, overview cards, intent route,
  API schema/type/constants, and React Query hook.
- Existing workflow authoring and MCP routing behavior remain unchanged.
- No database migration is expected because the intent catalog already exists.

## Non-goals

- No intent create, edit, delete, or workflow remapping UI; this slice is
  read-only discovery.
- No new LLM classification or fuzzy intent matching in OpenState.
- No project CRUD, project selector, tenant switching, or changes to the current
  `Default Project` resolution behavior.
- No changes to MCP tool names, arguments, or response contracts.
- No duplicate workflow authoring or publishing UI in Admin Console; Builder
  remains the owner of those actions.
