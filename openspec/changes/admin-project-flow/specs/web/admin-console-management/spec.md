## MODIFIED Requirements

### Requirement: Admin Console navigation reflects existing permissions

The web application SHALL provide a shared `/admin` layout and landing view. It
SHALL derive visible navigation sections and enabled actions from the
authenticated user's existing tenant-scoped permissions while relying on APIs
for final authorization. The navigation SHALL include Projects, Intents, and
Workflows for users with `workflow:read`.

#### Scenario: Viewer opens the Admin Console

- **WHEN** a Viewer opens `/admin`
- **THEN** the console shows Projects, Intents, and Workflows when the Viewer
  has `workflow:read`
- **AND** it does not present tenant/member mutation controls.

#### Scenario: Owner opens the Admin Console

- **WHEN** an Owner opens `/admin`
- **THEN** the console provides tenant settings and members/roles navigation in
  addition to Projects, Intents, and Workflows allowed by that Owner's
  permissions.

### Requirement: Admin Console makes the tenant-to-state hierarchy explicit

The shared Admin Console setup guide SHALL show the ordered hierarchy
`Tenant → Project → Intent → Workflow → State` wherever the console explains
how a workflow is created or operated. The Project step SHALL be navigable to
the project inventory, and a selected project SHALL be shown as the scope for
downstream steps.

#### Scenario: Workflow inventory shows project as an explicit step

- **WHEN** an authorized user opens the workflow inventory
- **THEN** the setup guide shows Project followed by Intent followed by
  Workflow and State
- **AND** the page states which project scope owns the listed workflows.

#### Scenario: User selects a project from the guide

- **WHEN** an authorized user selects the Project step
- **THEN** the application opens the project inventory
- **AND** the user can continue to an Intent or Workflow view scoped to that
  project.
