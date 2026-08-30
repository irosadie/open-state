## Context

The project has tenant-scoped RBAC and audit logging, plus standalone Audit and Capabilities administration pages. It lacks a unified console for the remaining administrative work. The Builder lifecycle and Runtime Inspector are separate active changes; duplicating their detail views here would create competing contracts.

This change therefore makes the Admin Console an access and operations layer. It owns management of the current tenant and its memberships, console navigation, instance commands, and event browsing. It links to the owning feature for workflow and runtime detail.

## Goals / Non-Goals

### Goals

- Let an Owner manage the current tenant's profile, memberships, and tenant-scoped role assignments.
- Let roles already authorized by RBAC issue safe instance `suspend`, `resume`, and `retry` commands in their own tenant.
- Let authorized users browse tenant events without altering event history.
- Provide one permission-aware Admin Console that incorporates existing Audit and Capabilities surfaces and routes to Builder and Runtime Inspector detail.
- Preserve server-enforced tenant isolation and a durable audit trail for every mutation.

### Non-Goals

- Tenant creation, organization switching, invitations, SCIM, or global user lifecycle management.
- Any cross-tenant query or mutation.
- Direct workflow editing/publishing/version management, simulation, or implementation of Runtime Inspector detail.
- Event mutation, event replay, deletion, or direct third-party AI/provider connections.

## Decisions

### 1. Separate identity administration from runtime operations

The backend exposes two bounded services:

- **Identity administration** serves only the authenticated request's tenant. It returns the current tenant profile and its memberships, and permits a caller with existing `tenant:*` and `user:*` authority to update tenant settings or change a membership role.
- **Runtime operations** validates tenant ownership and the existing exact instance permission (`instance:suspend`, `instance:resume`, or `instance:retry`) before delegating a lifecycle command to the runtime/application service. Event endpoints are read-only and are restricted to the current tenant.

This keeps role/membership policy separate from runtime state transitions and allows each service to use the correct authorization and audit vocabulary.

### 2. Reuse existing role semantics; do not introduce a second console role model

The console derives both visibility and action availability from the established tenant-scoped RBAC permissions. The server remains authoritative: hidden controls are a usability aid, never an authorization boundary.

Tenant profile and membership changes require the existing Owner-level tenant/user permissions. Runtime commands use the existing operator permissions. Users without a relevant permission can neither receive another tenant's data nor submit that action through the API.

### 3. Keep events immutable and queryable

The event browser supports paginated/filterable list and detail reads scoped by tenant, with correlation identifiers and source metadata where already persisted. It exposes no edit, delete, replay, or ad-hoc event injection operation. The browser can link an event to a runtime instance or audit record when an identifier is available, but it does not redefine the event contract.

### 4. Integrate rather than duplicate adjacent product areas

The Admin Console has a shared `/admin` layout and navigation. It provides:

- tenant settings and members/roles;
- a workflow inventory that routes to Builder lifecycle/version views;
- an instance list with permitted operational controls and links to Runtime Inspector detail;
- a read-only event browser;
- links to the existing Audit and Capabilities pages.

Workflow authoring/version flows are owned by `complete-builder-lifecycle`. Instance state, context, timeline, trace intent, and sanitized provider reference metadata are owned by `runtime-inspector-debug`. This console must consume those routes/contracts rather than recreate them.

### 5. Treat mutations as deliberate, auditable operations

The UI requires a clear confirmation before changing tenant settings, changing/removing a membership role, or submitting an instance command. APIs validate the current state and target tenant, reject invalid transitions, record actor/tenant/target/action/outcome/correlation context in audit storage, and return a stable error that the UI can display. Command controls refresh affected console and Runtime Inspector queries after success; they do not assume an optimistic state transition.

## API Shape

Routes are illustrative; final route naming follows the project's existing HTTP conventions.

| Area | Read | Mutation | Authorization |
| --- | --- | --- | --- |
| Current tenant | `GET /api/admin/tenant` | `PATCH /api/admin/tenant` | `tenant:read` / `tenant:update` |
| Memberships | `GET /api/admin/members` | `PUT /api/admin/members/{userId}/role`, `DELETE /api/admin/members/{userId}` | `user:read` / `user:update` |
| Instances | existing tenant-scoped list/detail contracts | `POST /api/admin/instances/{id}/suspend`, `/resume`, `/retry` | matching `instance:*` permission |
| Events | `GET /api/admin/events`, `GET /api/admin/events/{eventId}` | none | `instance:read` |

All list/detail routes obtain tenant identity from the authenticated session/context, never from a client-selected tenant identifier. The membership update must validate that the requested role is a valid tenant role and must prevent an invalid removal or role change that would leave the tenant without an Owner.

## Data and Audit

Existing tenant, user, role-assignment, workflow, instance, event, and audit persistence are reused. New tenant-filtered queries may be added and generated through the repository's standard data-access workflow; schema migration is not expected unless discovery finds a missing necessary persisted attribute.

Each successful or rejected mutation records an audit entry according to existing audit conventions, including actor, tenant, target resource, action, outcome, correlation identifier when present, and safe contextual metadata. Event reads do not write audit entries unless existing platform audit policy already requires read access logging.

## Risks / Trade-offs

- A member list can accidentally become a cross-tenant data leak. Every query must constrain by authenticated tenant before pagination/filtering and tests must cover mismatched resource identifiers.
- Concurrent runtime commands can race with orchestration. The command service must use existing state/version validation and return a conflict instead of silently applying an invalid transition.
- A member mutation can lock out the tenant. The service must preserve at least one Owner and reject a caller's action when this invariant would be violated.
- Builder and Runtime Inspector changes may land on a different cadence. Console links and integration hooks should be feature-compatible without reimplementing the dependent feature.

## Migration Plan

1. Add the tenant-filtered query, service, controller, and UI contracts behind current RBAC checks.
2. Deliver console navigation and pages with links to existing and dependent management surfaces.
3. Verify tenant-boundary, role invariant, audit, and runtime transition tests before enabling the routes.

Rollback removes the new routes and console entries. No destructive data migration is planned; audit records already written remain intentionally immutable.

## Open Questions

- Whether the existing role-assignment persistence already guarantees the last-Owner invariant or needs a transactional guard in the new service.
- Exact destination paths exported by `complete-builder-lifecycle` and `runtime-inspector-debug`; this change will use their published route contracts when they are finalized.
