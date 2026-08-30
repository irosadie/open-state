## MODIFIED Requirements

### Requirement: Create a workflow draft
The system SHALL allow an authenticated operator to create a workflow definition draft
within a project.

- The request SHALL be authenticated (JWT + auth session).
- The tenant SHALL be derived from the `X-Tenant-ID` header (PRD §74, §96), never from the body.
- The project SHALL be supplied by the caller or resolved to the tenant's default
  project, and SHALL be scoped to that tenant.
- The request SHALL include a complete initial workflow definition graph.
- The new workflow SHALL have `status = DRAFT` and persist that graph as its
  editable draft definition.
- The response SHALL include the persisted workflow (id, slug, name, status,
  currentVersion, version, and definition).

#### Scenario: Create a draft workflow
- **WHEN** an authenticated request creates a workflow with a valid projectId, slug, name, and definition
- **THEN** the workflow and its complete draft graph are persisted with `status=DRAFT` and its `id` is returned.

#### Scenario: Duplicate slug rejected
- **WHEN** a workflow is created with a slug that already exists for the same (tenant, project)
- **THEN** the request is rejected with a conflict error (409).

#### Scenario: Missing tenant header
- **WHEN** a request does not include `X-Tenant-ID`
- **THEN** it is rejected with an unauthorized error (401).

### Requirement: Get a workflow draft
The system SHALL allow an authenticated operator to fetch a single workflow definition
by id, scoped to their tenant and project.

- The response SHALL include the current editable draft definition even after one
  or more versions have been published.

#### Scenario: Fetch existing workflow
- **WHEN** an authenticated request fetches a workflow by id within their tenant and project
- **THEN** the workflow definition and its current draft graph are returned.

#### Scenario: Fetch missing workflow
- **WHEN** a workflow id does not exist for the caller's (tenant, project)
- **THEN** a not-found error (404) is returned.

### Requirement: Update a workflow draft
The system SHALL allow an authenticated operator to update mutable fields and the
complete editable graph of a workflow definition draft using optimistic concurrency
(PRD §31).

- The update SHALL require the caller's current `version`; a stale `version` SHALL be rejected as a conflict (409).
- The update SHALL atomically persist the supplied definition, name, and
  description; absent optional metadata SHALL retain its existing value.
- The update SHALL be allowed for an editable workflow, including one that already
  has published versions, and SHALL NOT alter immutable versions.

#### Scenario: Update draft with current version
- **WHEN** an authenticated request updates an editable workflow with a complete definition and the correct expected `version`
- **THEN** the update succeeds, `version` increments by 1, and the updated workflow including its definition is returned.

#### Scenario: Update draft with stale version
- **WHEN** an authenticated request updates a workflow with a stale expected `version`
- **THEN** the request is rejected with a conflict error (409) and its persisted draft remains unchanged.

### Requirement: Publish a workflow version
The system SHALL allow an authenticated operator to publish the workflow's persisted
draft definition, creating an immutable, current version snapshot (PRD §3.3, §9,
§55, §65, §69).

- Publishing SHALL atomically validate the server-side draft, insert a new
  `workflow_versions` row from that draft, mark it `is_current`, and bump
  `workflow.current_version` plus optimistic `version` (PRD §58).
- The client SHALL supply its expected workflow version but SHALL NOT be able to
  publish a graph different from the current persisted draft.
- Invalid definitions SHALL be rejected with a validation error that identifies
  the graph rule violations; a client-side validation result is not authoritative.
- A published snapshot SHALL remain immutable while later edits continue on the
  workflow's draft head.

#### Scenario: Publish creates an immutable current version
- **WHEN** an authenticated request publishes a workflow whose saved draft is valid and whose version is current
- **THEN** a new immutable version is created from that saved draft, marked current, and returned.

#### Scenario: Publish with stale version
- **WHEN** an authenticated request publishes a workflow with a stale expected `version`
- **THEN** the request is rejected with a conflict error (409).

#### Scenario: Publish invalid saved draft
- **WHEN** an authenticated request publishes a workflow whose saved draft violates workflow graph validation
- **THEN** no version is created and the request is rejected with a validation error.

### Requirement: List workflow versions
The system SHALL allow an authenticated operator to list the immutable versions of a
workflow, newest first.

#### Scenario: List versions
- **WHEN** an authenticated request lists versions for a workflow
- **THEN** all versions (versionNo, status, isCurrent, createdAt, updatedAt) are returned newest first.

## ADDED Requirements

### Requirement: Retrieve an immutable workflow version
The system SHALL allow an authenticated operator to retrieve the complete graph
definition for one published version of a workflow within their tenant and project.

#### Scenario: Retrieve a published version
- **WHEN** an authenticated request requests an existing workflow version number
- **THEN** the response contains that immutable version's metadata and complete
  definition.

#### Scenario: Retrieve an absent or cross-scope version
- **WHEN** a requested version does not belong to the workflow within the caller's tenant and project
- **THEN** the request is rejected with a not-found error (404).

### Requirement: Compare two immutable workflow versions
The system SHALL allow an authenticated operator to compare two distinct published
versions of the same workflow within their tenant and project.

- The response SHALL identify the base and target version numbers.
- The response SHALL report nodes and transitions added, removed, and changed,
  keyed by their stable graph ids.
- A changed node or transition SHALL identify which top-level fields changed,
  without modifying either version.
- Both requested versions SHALL belong to the specified workflow and scope.

#### Scenario: Compare consecutive versions
- **WHEN** an authenticated request compares two existing published versions of the same workflow
- **THEN** it receives a deterministic graph diff between the requested base and target versions.

#### Scenario: Reject invalid comparison pair
- **WHEN** a request compares identical versions or a version from another workflow
- **THEN** it is rejected without returning a diff.
