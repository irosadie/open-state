## MODIFIED Requirements

### Requirement: Workflow inventory routes to Builder lifecycle ownership

The web application SHALL provide a tenant-scoped workflow inventory in the Admin
Console and SHALL route workflow authoring and version-detail actions to the
destinations owned by the Builder lifecycle contract. It SHALL NOT duplicate
workflow authoring, publishing, or version management UI.

For users with `workflow:create` permission, the inventory page SHALL also provide
an entry point to create a new workflow draft. After successful creation the
application SHALL navigate to the Builder destination for the new workflow.

#### Scenario: Administrator opens a workflow from inventory

- **WHEN** a user with workflow read access selects a workflow in the console inventory
- **THEN** the application navigates to the Builder lifecycle or version-detail destination
- **AND** does not render a competing authoring flow in the console page.

#### Scenario: Authorized user creates a new workflow from inventory

- **WHEN** a user with `workflow:create` permission opens the workflow inventory
- **THEN** a "New Workflow" action SHALL be visible on the inventory page
- **AND** activating it opens a creation form requesting at minimum a slug and a name
- **AND** submitting valid input creates a workflow draft via the backend API
- **AND** the application navigates to the Builder destination for the newly created workflow.

#### Scenario: Unauthorized user cannot see the create action

- **WHEN** a user without `workflow:create` permission opens the workflow inventory
- **THEN** no workflow creation action is visible or accessible.

#### Scenario: Create form validation rejects invalid input

- **WHEN** a user submits the creation form with a missing slug or name
- **THEN** the form SHALL display field-level validation errors
- **AND** SHALL NOT submit the request to the backend.

#### Scenario: Backend error is surfaced to the user

- **WHEN** the backend returns an error for the create request
- **THEN** the form SHALL display the error message
- **AND** SHALL remain open so the user can correct and retry.

## ADDED Requirements

### Requirement: Workflow creation navigates to Builder on success

The web application SHALL, upon successful workflow creation from the Admin Console
inventory, immediately navigate the user to the Builder destination for the new
workflow ID returned by the API. It SHALL NOT remain on the inventory page after
a successful creation.

#### Scenario: Successful creation redirects to Builder

- **WHEN** a workflow draft is successfully created
- **THEN** the application navigates to `/state-builder/{id}` where `{id}` is the
  new workflow's identifier returned by the API.

### Requirement: Admin Console explains tenant and project scope

The Admin Console SHALL present the operator flow as `Tenant → Project →
Workflow → Builder`. The workflow inventory SHALL identify the current tenant
scope and SHALL state that workflows created without an explicit project are
placed in the tenant's `Default Project`.

#### Scenario: Operator opens the Admin Console

- **WHEN** an authenticated operator opens the Admin Console overview, tenant
  settings, or workflow inventory
- **THEN** the application shows the ordered setup path from tenant to Builder
- **AND** the Project step explains that `Default Project` is currently used
  automatically

#### Scenario: Operator opens workflow inventory

- **WHEN** an operator opens the workflow inventory
- **THEN** the page identifies the current tenant/project scope
- **AND** the page provides a clear next action to create a workflow or open an
  existing workflow in Builder

#### Scenario: Project management is not yet available

- **WHEN** an operator looks for project selection or project settings
- **THEN** the UI does not show a non-functional project selector
- **AND** the UI clearly explains that project switching/CRUD is not available
  yet and that workflow creation uses `Default Project`
