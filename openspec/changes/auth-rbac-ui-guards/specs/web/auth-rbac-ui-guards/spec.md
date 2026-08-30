## Purpose

Provide tenant-aware browser route and action guards that consistently reflect effective RBAC permissions while keeping server-side authorization authoritative.

## ADDED Requirements

### Requirement: UI authorization uses the current tenant's effective permissions

The web application SHALL resolve the authenticated user's effective role and permissions from the current-user contract for the active tenant before making a protected UI authorization decision. It SHALL use the same tenant context as protected API requests and SHALL treat a missing, unknown, or empty permission set as default deny.

#### Scenario: Tenant-scoped authorization snapshot resolves

- **WHEN** an authenticated user opens a protected application route for the active tenant
- **THEN** the application obtains that tenant's effective role and permissions before rendering protected content
- **AND** authorization for another tenant is not inferred from the user's role in the active tenant.

#### Scenario: User has no assignment in the active tenant

- **WHEN** the current-user authorization response has no effective permissions for the active tenant
- **THEN** the application denies protected route and action access
- **AND** renders an explicit no-access state without loading protected resource data.

### Requirement: Permission matching follows server wildcard semantics

The web application SHALL evaluate permission checks using exact matches and resource wildcards consistent with the server RBAC matrix. A granted `resource:*` permission SHALL satisfy a requested `resource:verb`; an unknown or unmatched permission SHALL not satisfy any route or action requirement.

#### Scenario: Wildcard grants a workflow action

- **WHEN** the effective permission set contains `workflow:*`
- **AND** a workflow action requires `workflow:publish`
- **THEN** the application treats the action as available.

#### Scenario: Unknown permission is requested

- **WHEN** a route or action requirement has no matching effective permission
- **THEN** the application treats it as denied
- **AND** does not render an enabled protected control.

### Requirement: Protected routes are authenticated and permission-gated

The web application SHALL classify browser routes as public or protected. Every protected route SHALL require a valid browser session and a route-policy permission decision before protected page content or protected data queries render. A user who directly opens a route without its required permission SHALL receive a stable access-denied experience without protected data exposure.

#### Scenario: Unauthenticated user opens a protected route

- **WHEN** a user without a valid session opens a protected application route
- **THEN** the application redirects to login with a sanitized same-origin callback path
- **AND** does not render the protected page.

#### Scenario: Authenticated Viewer opens a tenant-management route

- **WHEN** a Viewer opens a route requiring tenant or user management permission
- **THEN** the application renders the access-denied state
- **AND** does not mount that route's protected data query or mutation controls.

### Requirement: Post-login landing is safe and permission-aware

The web application SHALL resolve post-login navigation from the active tenant's effective permissions. It SHALL retain a sanitized callback only when that route is authorized; otherwise it SHALL choose an authorized landing defined by route policy. If no landing is authorized, it SHALL render a terminal no-access state and SHALL NOT loop between routes.

#### Scenario: Authorized callback is retained after login

- **WHEN** an Editor logs in with a same-origin callback to a route the Editor may access
- **THEN** the application navigates to that callback after authorization resolves.

#### Scenario: Callback requires a denied permission

- **WHEN** an Operator logs in with a callback to a capability-management route requiring `capability:read`
- **THEN** the application does not navigate to the denied callback
- **AND** navigates to the Operator's first authorized landing instead.

### Requirement: Navigation and actions reflect their exact permissions

The web application SHALL derive protected navigation visibility and individual action availability from the shared route/action policy. A permitted page does not imply that every mutation on that page is permitted. Sensitive unavailable actions SHALL not render an invokable control, and a protected query or mutation SHALL not be issued solely to discover client-side visibility.

#### Scenario: Editor views capability inventory without management permission

- **WHEN** an Editor can open a permitted navigation surface but lacks `capability:create`
- **THEN** the application does not render an invokable create-capability control
- **AND** does not issue a create request.

#### Scenario: Operator opens an instance

- **WHEN** an Operator with `instance:read` opens an instance surface
- **THEN** the application renders permitted instance data
- **AND** renders suspend, resume, retry, and debug controls only when their exact instance or debug permission is granted.

### Requirement: Permission changes and API denials have clear recovery behavior

The web application SHALL distinguish unauthenticated and unauthorized responses. A `401` SHALL follow session-expiry login recovery. A `403` SHALL retain the session, refresh the effective authorization snapshot, and present forbidden feedback for the affected route or action without exposing protected data as authorized.

#### Scenario: Membership role changes during an active session

- **WHEN** a user's permission is removed after the UI initially rendered an allowed action
- **AND** the user submits that action and the API returns `403`
- **THEN** the application retains the user's session
- **AND** refreshes authorization state
- **AND** replaces the affected action or route with forbidden feedback.

#### Scenario: Access token expires

- **WHEN** a protected API request returns `401` after token refresh cannot recover the session
- **THEN** the application clears the expired session path and directs the user to login
- **AND** does not present the response as a permission denial.
