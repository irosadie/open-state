## 1. Authorization contract and policy foundation

- [x] 1.1 Map current public/protected browser routes, existing API permission requirements, and all current navigation/action controls into a route-action policy inventory.
- [x] 1.2 Add shared frontend role/permission schema and typed authorization snapshot contract for the existing tenant-scoped current-user response.
- [x] 1.3 Add a pure permission matcher that implements exact and `resource:*` wildcard behavior aligned with the server RBAC matrix.
- [x] 1.4 Add a typed route-policy registry with public routes, protected route requirements, action requirements, and ordered authorized landing candidates.
- [x] 1.5 Add unit tests for wildcard semantics, unknown/default-deny permissions, route matching, action requirements, and fallback selection.

## 2. Authentication and authorization boundaries

- [x] 2.1 Update the current-user query to send the active tenant header and include tenant identity in its cache key.
- [x] 2.2 Add an authorization provider/hook that exposes session-aware loading, ready, unauthenticated, forbidden, error, and refresh states from the current-user snapshot.
- [x] 2.3 Expand the Next.js proxy matcher to authenticate every non-public application route and preserve only sanitized same-origin callback paths.
- [x] 2.4 Add shared protected-route and access-denied boundaries that wait for authorization before mounting protected content or its data queries.
- [x] 2.5 Add a shared permission/action gate whose unavailable sensitive controls cannot invoke a query or mutation handler.
- [x] 2.6 Add tests for unauthenticated redirect, permission-loading boundary, denied direct navigation, no-access terminal state, and protected-content non-rendering.

## 3. Role-aware session navigation and error recovery

- [x] 3.1 Replace the generic role redirect helper with the policy-driven safe-callback-or-authorized-landing resolver.
- [x] 3.2 Update credential login, registration login, and already-authenticated login-page behavior to resolve authorization before redirecting.
- [x] 3.3 Handle no authorized landing as an explicit no-access state and prevent redirect loops.
- [x] 3.4 Standardize authenticated API error handling so `401` clears/returns to login and `403` refreshes authorization then surfaces forbidden feedback without logging out.
- [x] 3.5 Add tests for permitted callback retention, denied callback fallback, role-specific landing selection, no-access behavior, and 401 versus 403 handling.

## 4. Migrate existing and planned product surfaces

- [x] 4.1 Guard the Audit page and its query with `audit:read`, preserving its current authorized behavior and avoiding data load for denied users.
- [x] 4.2 Guard Capability routes/data and gate create, update, delete, test, and binding controls by their exact capability/binding permissions.
- [x] 4.3 Guard state-builder/workflow routes and gate create, update, publish, and simulate actions by their exact workflow permissions.
- [x] 4.4 Register Admin Console tenant/member/workflow/instance/event routes and controls from `admin-console-management` with their declared permissions; do not add local role checks.
- [x] 4.5 Register Runtime Inspector and Debug View routes/actions from `runtime-inspector-debug` with `instance:read` and `debug:read` independently.
- [x] 4.6 Replace or adapt existing menu filtering to use the shared matcher and route-policy registry, removing duplicate incompatible wildcard logic.
- [x] 4.7 Add integration/page tests for each role's visible navigation, allowed actions, hidden/disabled disallowed actions, guarded queries, and direct-route denial.

## 5. Verification and documentation

- [x] 5.1 Document the frontend route/action policy registration convention and how future features declare their permissions.
- [x] 5.2 Run frontend unit, component, and route tests plus relevant API authorization tests across Owner, Admin, Editor, Operator, Viewer, and absent-role cases.
- [x] 5.3 Validate the OpenSpec change strictly and resolve all validation issues.
