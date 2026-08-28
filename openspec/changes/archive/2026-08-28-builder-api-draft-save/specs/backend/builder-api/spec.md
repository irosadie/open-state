## Purpose

Provides the HTTP Builder API (PRD 146) that lets the State Builder and other
operator surfaces persist and manage workflow definitions — draft create/get/update,
publish to an immutable version, list, and list-versions — all tenant+project scoped
and behind authenticated auth.

## ADDED Requirements

### Requirement: Create a workflow draft
The system SHALL allow an authenticated operator to create a workflow definition draft
within a project.

- The request SHALL be authenticated (JWT + auth session).
- The tenant SHALL be derived from the `X-Tenant-ID` header (PRD §74, §96), never from the body.
- The project SHALL be supplied by the caller and validated to exist for the tenant.
- The new workflow SHALL have `status = DRAFT`.
- The response SHALL include the persisted workflow (id, slug, name, status, currentVersion, version).

#### Scenario: Create a draft workflow
- **WHEN** an authenticated request creates a workflow with a valid projectId, slug, and name
- **THEN** the workflow is persisted with `status=DRAFT` and its `id` is returned.

#### Scenario: Duplicate slug rejected
- **WHEN** a workflow is created with a slug that already exists for the same (tenant, project)
- **THEN** the request is rejected with a conflict error (409).

#### Scenario: Missing tenant header
- **WHEN** a request does not include `X-Tenant-ID`
- **THEN** it is rejected with an unauthorized error (401).

### Requirement: Get a workflow draft
The system SHALL allow an authenticated operator to fetch a single workflow definition
by id, scoped to their tenant and project.

#### Scenario: Fetch existing workflow
- **WHEN** an authenticated request fetches a workflow by id within their tenant and project
- **THEN** the workflow definition is returned.

#### Scenario: Fetch missing workflow
- **WHEN** a workflow id does not exist for the caller's (tenant, project)
- **THEN** a not-found error (404) is returned.

### Requirement: List workflows
The system SHALL allow an authenticated operator to list the workflow definitions
within a project for their tenant.

#### Scenario: List tenant/project workflows
- **WHEN** an authenticated request lists workflows for a project
- **THEN** all workflow definitions for that (tenant, project) are returned.

### Requirement: Update a workflow draft
The system SHALL allow an authenticated operator to update mutable fields of a workflow
definition draft (name, description) using optimistic concurrency (PRD §31).

- The update SHALL require the caller's current `version`; a stale `version` SHALL be rejected as a conflict (409).
- Only `DRAFT` workflows SHALL be editable.

#### Scenario: Update draft with current version
- **WHEN** an authenticated request updates a DRAFT workflow supplying the correct expected `version`
- **THEN** the update succeeds, `version` increments by 1, and the updated workflow is returned.

#### Scenario: Update draft with stale version
- **WHEN** an authenticated request updates a workflow with a stale expected `version`
- **THEN** the request is rejected with a conflict error (409).

### Requirement: Publish a workflow version
The system SHALL allow an authenticated operator to publish a workflow definition,
creating an immutable, current version snapshot (PRD §3.3, §9, §55, §65, §69).

- Publishing SHALL atomically insert a new `workflow_versions` row, mark it `is_current`, and bump `workflow.current_version` + optimistic `version` (PRD §58).
- Only a `VALID` (or validatable) workflow SHALL be published; a `DRAFT` without validation SHALL be rejected.
- The published `definition` SHALL store the full workflow definition JSON (PRD §68).

#### Scenario: Publish creates an immutable current version
- **WHEN** an authenticated request publishes a workflow with a valid definition
- **THEN** a new immutable version is created, marked current, and returned.

#### Scenario: Publish with stale version
- **WHEN** an authenticated request publishes a workflow with a stale expected `version`
- **THEN** the request is rejected with a conflict error (409).

### Requirement: List workflow versions
The system SHALL allow an authenticated operator to list the immutable versions of a
workflow, newest first.

#### Scenario: List versions
- **WHEN** an authenticated request lists versions for a workflow
- **THEN** all versions (versionNo, status, isCurrent, createdAt, updatedAt) are returned newest first.
