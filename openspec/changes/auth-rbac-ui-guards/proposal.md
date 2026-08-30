# Change: Add tenant-aware Auth and RBAC UI guards

## Why

Login and browser sessions exist, and API routes already enforce tenant-scoped RBAC, but the web experience still redirects every role to the same home path and has no consistent route or action guard. Users can reach screens whose controls and data do not match their effective permission set, leaving the UI to fail only after an API request.

## What Changes

- Add a tenant-aware frontend authorization snapshot based on the existing authenticated current-user response, including effective role and permissions.
- Add a shared permission matcher and route-policy registry used by post-login redirects, protected routes, navigation, page data loading, and individual actions.
- Guard all non-public application routes at the browser boundary for authentication and gate protected pages after authorization resolves; unauthorized users see a deliberate access-denied state without protected data loading.
- Replace generic role redirect behavior with a safe callback-or-authorized-landing decision based on the user's effective permissions for the active tenant.
- Apply consistent permission-based visibility to existing Audit, Capability, workflow builder, and future Admin Console/Runtime Inspector controls, while retaining API enforcement as the security boundary.
- Handle role changes and server `401`/`403` responses consistently: expired sessions return to login, while permission denials refresh authorization state and remain a forbidden experience.

## Capabilities

### New Capabilities

- `web/auth-rbac-ui-guards`: Provides tenant-aware authorization state, permission matching, protected-route decisions, action guards, and authorized post-login landing behavior for the web application.

### Modified Capabilities

None.

## Impact

- Affected frontend areas: Next.js proxy/matcher, login and registration redirect behavior, authenticated API query setup, session/authorization provider, route layouts, menus, buttons/forms, and frontend tests.
- Reuses the existing `GET /api/auth/me` role/permission response and configured tenant header; no role matrix, API permission, or authentication-provider behavior is changed.
- Integrates with active `admin-console-management`, `complete-builder-lifecycle`, and `runtime-inspector-debug` changes by supplying their shared route/action guard foundation.

## Non-goals

- Changing the backend RBAC matrix, API `RequirePermission` middleware, token format, authentication providers, SSO, or tenant switching.
- Treating a client-side guard as authorization; all APIs remain responsible for server-side denial.
- Introducing ABAC, custom roles, or a global cross-tenant admin role.
