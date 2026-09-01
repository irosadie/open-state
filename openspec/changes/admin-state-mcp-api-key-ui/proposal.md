## Why

State MCP now requires a tenant-scoped `osk_...` bearer key, but the Admin
Console has no way for an operator to provision or revoke that credential. The
current workflow requires manual API calls and makes the one-time secret easy to
miss.

## What Changes

- Add a permission-aware `/admin/api-keys` page linked from the `/admin`
  overview and Admin Console navigation.
- Let authorized operators create a State MCP API key with a name, project
  allowlist, default project, scopes, and optional expiration.
- Show the raw bearer key exactly once after creation with a copy action and a
  clear warning that it cannot be recovered later.
- List safe key metadata and allow authorized operators to revoke active keys.
- Let the project selector discover projects already owned by the active tenant
  through a read-only API endpoint.
- Add frontend Zod schemas, response types, API constants, React Query hooks,
  route policy, and tests for the API-key and project-discovery endpoints.

## Capabilities

### New Capabilities

- `web/admin-state-mcp-api-keys`: Provide a tenant-scoped Admin Console surface
  for provisioning, inspecting, and revoking State MCP machine credentials.

### Modified Capabilities

- `web/admin-console-management`: Add the API Keys entry point and permission
  aware navigation for users granted `api_key:read`.

## Impact

- Frontend Admin Console route, navigation, overview card, route/action policy,
  shared schemas/types, API constants, and transaction hooks.
- A read-only tenant-scoped project-discovery endpoint backed by the existing
  project repository; no database schema change.
- No raw API key or server pepper is persisted in browser state or sent to the
  provider mock.

## Non-goals

- No project CRUD; the selector only discovers projects already owned by the
  current tenant.
- No API-key secret retrieval after creation.
- No OAuth, provider-mock authentication, or changes to State MCP tool scopes.
