## Context

The API already resolves a user's tenant-scoped role and permissions through `GET /api/auth/me`, and protected APIs use `RequirePermission`. The web session stores identity and tokens but not effective permissions. The frontend proxy currently only matches the login route, `getRoleRedirectPath` always returns `/`, and standalone admin pages issue data queries without a shared authorization gate.

The configured tenant header remains the current tenant-selection mechanism. This design does not change its scope; it makes every UI authorization decision explicitly use the same tenant context as API requests.

## Goals / Non-Goals

### Goals

- Establish one frontend source of effective role and permissions per active tenant.
- Make authentication, route access, navigation, data-query enablement, and action visibility follow one permission policy.
- Preserve safe callback URLs while sending users to an authorized landing when no valid callback is available.
- Distinguish session expiry from permission denial and keep server enforcement authoritative.

### Non-Goals

- Persisting permissions into the NextAuth JWT/session or relying on login payload roles for authorization.
- Replacing API authorization, adding a new backend authorization endpoint, or allowing the browser to choose a tenant outside the existing tenant context.
- Deciding the visual design of Phase 4 Admin Console, Builder lifecycle, or Runtime Inspector pages.

## Decisions

### 1. Use `/api/auth/me` as the authorization snapshot

The authenticated client obtains a typed effective role and permissions snapshot from the existing current-user endpoint, passing the same `X-Tenant-ID` header used by protected resource hooks. A provider exposes explicit `loading`, `ready`, `unauthenticated`, `forbidden`, and recoverable-error states plus a refresh operation.

Permissions are not copied to the NextAuth session or trusted from the login response. Tenant membership can change while a session remains valid; a separately cached/refetchable snapshot avoids presenting stale privileges as authoritative.

### 2. Centralize permission semantics and route policy

A pure permission utility implements the same exact and resource-wildcard semantics as the server matrix (`workflow:*` grants `workflow:read` and `workflow:publish`). It does not use a second role-to-permission matrix.

A single typed route-policy registry declares each protected browser route, its required permission or explicit any-of access set, and its ordered fallback landing eligibility. It initially covers current Audit, Capability, and state-builder routes and reserves entries for the Admin Console, workflow inventory, instance/events, and Runtime Inspector routes delivered by active dependent changes. Navigation items, page guards, and action gates consume this registry instead of independently comparing role names.

### 3. Layer authentication and authorization guards

The Next.js proxy protects every non-public application page by requiring a valid session token, preserving only same-origin callback paths. It does not attempt permission resolution at the edge.

After authentication, an application-level authorization provider resolves `/auth/me`. A route guard waits for the snapshot before rendering a protected page or enabling its data hooks. A denied route shows a stable access-denied screen and never mounts protected data content. Public login and registration routes remain excluded from this behavior.

This division keeps the proxy fast and prevents route redirects based on stale token claims, while still avoiding unauthenticated page flashes.

### 4. Make post-login redirect permission-aware

After credential login, registration login, or a visit to the login page with an existing session, the app first resolves the active tenant authorization snapshot. It uses a sanitized same-origin callback only when the route policy permits that user; otherwise it chooses the first authorized landing from the registry. The initial landing priority is management for users with an authorized Admin Console entry, Builder for workflow users, Runtime for instance users, then another permitted read surface. If no route is permitted, the user receives an explicit no-access state rather than an endless redirect loop.

### 5. Separate page access from action access

Page routes require their read or section permission. Inside an allowed page, a reusable action gate controls each mutation independently. By default, unavailable sensitive actions are not rendered; a deliberate disabled/fallback state is allowed only when no event handler can be invoked. Queries and mutations use `enabled`/guarded execution so unauthorized actions do not call the API merely to determine visibility.

On an API `403`, the client refreshes the authorization snapshot and presents the forbidden state for the affected surface. On `401`, existing session-clear/login behavior remains. This handles role changes between initial UI rendering and a mutation request without conflating authorization failure with logout.

## Route Policy Baseline

| Surface | Route access | Action access |
| --- | --- | --- |
| Audit | `audit:read` | read-only |
| Capabilities | `capability:read` | create/update/delete/test and binding actions use their matching capability/binding permission |
| State Builder / workflow inventory | `workflow:read` | create/update/publish/simulate use matching workflow permission |
| Admin tenant and members | `tenant:read` / `user:read` | settings and membership changes use matching tenant/user permission |
| Runtime instances and events | `instance:read` | suspend/resume/retry and debug trace use their matching instance/debug permission |

Routes that are not explicitly classified as public or protected are denied by policy until they are classified. Dependent changes own their route names; they must register their published routes and actions in this baseline rather than create local role checks.

## Risks / Trade-offs

- [Authorization snapshot fetch delays first protected render] → Render an intentional loading boundary and do not mount sensitive content before it resolves.
- [Client/server policy can drift] → The client consumes server-provided permission strings, uses one shared matcher, and adds matrix/route/action tests for every registered policy.
- [Role changes during an active session] → Refresh the snapshot after role/membership mutations and after a `403`; the API remains the final default-deny guard.
- [Fallback configuration can cause redirect loops] → Require every fallback candidate to be policy-authorized and render a terminal no-access state when no candidate matches.
- [Tenant header configuration is not yet a tenant switcher] → Scope this change to the active configured tenant and keep the tenant source explicit in the provider/query key.

## Migration Plan

1. Introduce the typed authorization snapshot, matcher, and route-policy registry with unit tests before routing current pages through it.
2. Expand the proxy matcher and add the shared protected-page/action guard boundary, retaining public auth routes.
3. Migrate current Audit, Capability, and Builder surfaces; dependent Admin Console and Runtime Inspector work registers its routes/actions against the same policy.
4. Release behind the existing API guards, verify 401/403 and tenant-boundary behavior, then remove obsolete generic role redirect and duplicate menu permission logic.

Rollback restores the current protected-route matcher and generic landing behavior. No database migration or persistent authorization data change is involved.
