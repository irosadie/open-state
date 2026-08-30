## 1. Backend: tenant and membership administration

- [x] 1.1 Map the existing tenant, user, role-assignment, RBAC, and audit services, then define tenant-filtered request/response DTOs for the console.
- [x] 1.2 Add repository queries and application services to read/update the authenticated tenant profile and list its memberships without cross-tenant leakage.
- [x] 1.3 Add a transactional membership-role service that validates tenant roles and preserves at least one Owner when roles or memberships change.
- [x] 1.4 Add authenticated admin routes/controllers for tenant profile and membership reads/mutations, enforcing existing `tenant:*` and `user:*` permissions.
- [x] 1.5 Emit existing-format audit entries for tenant settings, membership, and role mutations, including rejected invariant or authorization outcomes where policy requires.
- [x] 1.6 Add unit/integration tests for tenant isolation, authorization, role validation, last-Owner protection, response contracts, and audit records.

## 2. Backend: runtime operations and events

- [x] 2.1 Map existing runtime orchestration and event persistence interfaces, then define tenant-scoped command and event-browser DTOs.
- [x] 2.2 Add instance suspend, resume, and retry command handlers that validate tenant ownership, current lifecycle state/version, and the matching existing instance permission.
- [x] 2.3 Add read-only tenant-scoped event list/detail queries with pagination and safe filters; do not add event mutation or replay endpoints.
- [x] 2.4 Add authenticated admin routes/controllers for the instance commands and event reads, returning stable conflict/not-found/forbidden errors without resource existence leakage.
- [x] 2.5 Audit runtime lifecycle commands with actor, tenant, target, action, outcome, and correlation context; retain events as immutable records.
- [x] 2.6 Add tests for command authorization, invalid transition conflicts, tenant isolation, event immutability, pagination/filter behavior, and audit entries.

## 3. Frontend: console data contracts

- [x] 3.1 Add validated API schemas, types, constants, and query/mutation hooks for current tenant settings and tenant memberships.
- [x] 3.2 Add validated API schemas, types, constants, and query/mutation hooks for runtime commands and read-only event browsing.
- [x] 3.3 Ensure mutation hooks invalidate affected tenant, member, instance, event, and Runtime Inspector queries only after server success.
- [x] 3.4 Add frontend contract and hook tests for authorization/error states, confirmation flows, and cache invalidation.

## 4. Frontend: Admin Console experience

- [x] 4.1 Add a shared `/admin` console layout, landing view, and navigation that derives visible sections and enabled actions from session permissions.
- [x] 4.2 Add tenant settings and members/roles pages with accessible forms, role selection, last-Owner-safe errors, and explicit confirmation before every mutation.
- [x] 4.3 Add a workflow inventory page that uses the Builder lifecycle contract and routes users to its authoring/version destinations without duplicating those flows.
- [x] 4.4 Add an instance management page with tenant-scoped list, permitted suspend/resume/retry controls, confirmations, conflict feedback, and links to Runtime Inspector detail.
- [x] 4.5 Add a read-only event browser with filters, pagination, detail view, and links to related instance/audit context when identifiers are available.
- [x] 4.6 Integrate the existing Audit and Capabilities pages into console navigation without regressing their established behavior.
- [x] 4.7 Add page-level tests for permission-aware navigation, forbidden states, confirmations, lifecycle-operation feedback, and no event mutation controls.

## 5. Verification and documentation

- [x] 5.1 Update relevant API/OpenAPI and operational documentation for the tenant, membership, runtime-command, and event-browser contracts.
- [x] 5.2 Run backend, frontend, and end-to-end checks covering tenant boundaries, RBAC, audit records, runtime conflicts, and console navigation.
- [x] 5.3 Validate the OpenSpec change strictly and resolve all validation issues.
