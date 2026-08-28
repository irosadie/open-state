# backend/audit-api Specification

## Purpose

Define the tenant-scoped HTTP endpoint(s) to query the audit trail (PRD §50),
protected by the RBAC guard for `audit:read` (see `auth/authorization-guards`).
It enables the audit UI (see `web/audit-ui`) and operators to read, filter, and
paginate audit records without cross-tenant access.

## ADDED Requirements

### Requirement: List audit entries

The platform SHALL expose a tenant-scoped endpoint to list audit entries:
`GET /api/audit` (PRD §50).

- The endpoint SHALL require authentication.
- The endpoint SHALL require the `audit:read` permission for the request tenant
  (via `RequirePermission`).
- The tenant SHALL be read from the `X-Tenant-ID` header; results SHALL be
  limited to that tenant (PRD §4, §96).
- The endpoint SHALL return a paginated list of audit entries ordered by
  `occurred_at` descending.

#### Scenario: Read authorized audit entries

- **WHEN** an authorized user (with `audit:read`) requests `/api/audit` for a
  tenant
- **THEN** the platform returns the tenant's audit entries, newest first,
  paginated

#### Scenario: Unauthorized read is forbidden

- **WHEN** a user without `audit:read` requests `/api/audit`
- **THEN** the platform SHALL return 403

#### Scenario: Cross-tenant entries are excluded

- **WHEN** a user requests audit for tenant A
- **THEN** the response SHALL NOT include entries belonging to other tenants

### Requirement: Filter audit entries

The platform SHALL support filtering audit entries by action, resource type,
resource id, and actor (PRD §50).

- `GET /api/audit` SHALL accept optional query parameters `action`,
  `resourceType`, `resourceId`, `actor`, `correlationId`, and `from`/`to`
  (occurred_at range).
- Multiple filters SHALL be combinable (AND).
- An unknown/invalid filter value SHALL produce a 400 validation error.

#### Scenario: Filter by action

- **WHEN** a request includes `action=workflow.published`
- **THEN** the response SHALL contain only entries with that action

#### Scenario: Filter by resource

- **WHEN** a request includes `resourceType=capability&resourceId=<id>`
- **THEN** the response SHALL contain only matching entries

#### Scenario: Combined filters

- **WHEN** a request includes `action` and `resourceType` filters
- **THEN** the response SHALL contain only entries matching both

### Requirement: Paginate audit entries

The platform SHALL support cursor or offset pagination for audit reads (PRD §50).

- The response SHALL include a pagination envelope (`data`, `page`,
  `pageSize`, `total`, and `next`/`hasNext`).
- Default `pageSize` SHALL be capped at a maximum (e.g. 100) to bound response
  size.

#### Scenario: Paginate results

- **WHEN** an audit listing exceeds the page size
- **THEN** the response SHALL return the first page with a valid `next` cursor
  or page indicator

#### Scenario: Page size is bounded

- **WHEN** a request asks for a page size above the cap
- **THEN** the platform SHALL clamp it to the maximum

### Requirement: Audit entry DTO

The platform SHALL expose an audit entry DTO that is safe for external
consumption.

- The DTO SHALL include `id`, `actor`, `action`, `resourceType`,
  `resourceId`, `before`, `after`, `correlationId`, and `occurredAt`.
- The DTO SHALL NOT include internal fields or raw secret data.

#### Scenario: DTO is serializable and secret-safe

- **WHEN** an audit entry is serialized in the API response
- **THEN** it SHALL contain the public fields and SHALL NOT leak secrets

### Requirement: Sync DTO/OpenAPI contract

The audit query API SHALL be reflected in the shared API contract and OpenAPI
documentation (repo convention: sync DTO, OpenAPI, shared types).

- The response DTOs and query parameters SHALL be added to the API contract
  docs.
- Any contract drift introduced by this endpoint SHALL be reconciled.

#### Scenario: Contract is documented

- **WHEN** the audit endpoint is added
- **THEN** it SHALL be documented in the shared contract/OpenAPI with its DTO
  and query parameters
