# web/audit-ui Specification

## Purpose
Define the frontend Audit page that displays the tenant-scoped audit trail
(PRD §50) consumed from the `GET /api/audit` endpoint (see
`backend/audit-api`), gated by the user's `audit:read` permission (see
`auth/authorization-guards`). It gives operators visibility into what happened,
by whom, and when.

## Requirements

### Requirement: Audit page

The platform SHALL provide an Audit page in the frontend that lists audit
entries for the current tenant (PRD §50).

- The page SHALL be accessible to users who hold the `audit:read` permission.
- The page SHALL render each entry's action, actor, resource, occurred-at time,
  and expandable `before`/`after` payload.
- The page SHALL load data from `GET /api/audit` via a react-query hook.

#### Scenario: Authorized user views audit log

- **WHEN** a user with `audit:read` opens the Audit page
- **THEN** the page lists the tenant's audit entries, newest first

#### Scenario: Unauthorized user is gated

- **WHEN** a user without `audit:read` opens the Audit page
- **THEN** the page SHALL not render audit data (hide/redirect based on the
  user's permissions)

### Requirement: Audit query hook

The platform SHALL provide a react-query hook encapsulating the audit list API.

- The hook SHALL accept filters (action, resourceType, resourceId, actor) and a
  page/pageSize.
- The hook SHALL map to `GET /api/audit` with the tenant header.
- The hook SHALL expose loading, error, data, and pagination state.

#### Scenario: Hook returns typed data

- **WHEN** the hook is used with filters
- **THEN** it returns the typed, paginated audit entries

### Requirement: Filtering and search

The platform SHALL support filtering the audit list from the UI.

- The page SHALL provide controls to filter by action, resource type, resource
  id, and actor.
- Filters SHALL be reflected in the query and reset the pagination.

#### Scenario: User filters the audit log

- **WHEN** a user selects an action filter
- **THEN** the page reloads entries matching that action

### Requirement: Pagination

The platform SHALL support paging through the audit log from the UI.

- The page SHALL render previous/next (or load-more) controls using the API's
  pagination envelope.

#### Scenario: User pages through entries

- **WHEN** a user requests the next page
- **THEN** the page shows the next set of entries

### Requirement: Empty and error states

The platform SHALL handle empty and error states on the Audit page.

- An empty result SHALL show a clear "no audit entries" message.
- An API error SHALL show a recoverable error state with a retry action.

#### Scenario: Empty state

- **WHEN** a filter yields no entries
- **THEN** the page SHALL show an empty-state message

#### Scenario: Error state with retry

- **WHEN** the audit list request fails
- **THEN** the page SHALL show the error and a retry control

### Requirement: Zod schema + typed contract

The frontend SHALL define Zod schemas and TypeScript types for the audit DTO
matching the backend contract (repo convention: shared types + Zod).

- An `AuditEntry` Zod schema and the paginated response schema SHALL be defined.
- The frontend types SHALL stay in sync with the backend DTO and OpenAPI
  contract.

#### Scenario: Schemas validate API responses

- **WHEN** the audit list API responds
- **THEN** the response SHALL be parsed against the shared Zod schema
