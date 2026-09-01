## Purpose

Provide authenticated Admin Console clients with a safe, tenant/project-scoped
read model of the canonical intent mappings already used by MCP routing.

## ADDED Requirements

### Requirement: Admin clients can list routable intents

The backend SHALL expose a read-only intent catalog endpoint at `GET
/api/intents`. The endpoint SHALL accept an optional `projectId` query parameter;
when it is omitted, the request SHALL use the tenant's current `Default Project`
resolution used by workflow APIs.

#### Scenario: List intents for the default project

- **WHEN** an authenticated user with `workflow:read` requests `GET /api/intents`
  with a valid tenant scope
- **THEN** the response contains the canonical intents for that tenant's
  default project
- **AND** each returned intent includes its canonical key, display metadata,
  example utterances, project id, and mapped workflow identity.

#### Scenario: List intents for an explicit project

- **WHEN** an authenticated user with `workflow:read` requests `GET
  /api/intents?projectId={projectId}`
- **THEN** the response contains only intents owned by that project and tenant
- **AND** a project belonging to another tenant is not observable through the
  endpoint.

#### Scenario: Valid project has no routable intents

- **WHEN** an authorized user requests a valid active project with no published
  intent mappings
- **THEN** the endpoint returns a successful response with an empty data array
- **AND** it does not substitute intents from another project.

### Requirement: Catalog reads expose only routable mappings

The endpoint SHALL return only intent records whose project is active and whose
mapped workflow is published. Each response item SHALL include `id`, `key`,
`name`, `description`, `examples`, `tenantId`, `projectId`, `workflowId`, and
`workflowSlug`.

#### Scenario: Draft workflow mapping is hidden

- **WHEN** an intent is mapped to a workflow that is not published
- **THEN** that intent is excluded from the HTTP catalog
- **AND** the endpoint does not expose the draft workflow definition or its
  unpublished mapping.

#### Scenario: Published mapping is returned with examples

- **WHEN** an intent is mapped to a published workflow
- **THEN** the response includes its canonical key such as `BOOKING_PADEL`
- **AND** includes the stored example utterances such as `saya mau order
  lapangan`
- **AND** includes the workflow id and slug needed to open the owning workflow.

### Requirement: Catalog access preserves existing authorization boundaries

The endpoint SHALL require authentication, a tenant scope, and the existing
`workflow:read` permission. It SHALL not provide create, update, delete, or
workflow-remapping operations.

#### Scenario: Missing authentication or tenant scope

- **WHEN** a request is unauthenticated or does not provide the required tenant
  scope
- **THEN** the endpoint rejects the request using the existing authentication
  error contract
- **AND** it does not query or return catalog data.

#### Scenario: User lacks workflow read access

- **WHEN** an authenticated user without `workflow:read` requests the catalog
- **THEN** the endpoint rejects the request using the existing authorization
  error contract
- **AND** no intent records are returned.
