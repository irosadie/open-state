## Purpose

Make Project a navigable, tenant-scoped step between Tenant and Intent, and
carry that project scope through Workflow and State Builder experiences.

## ADDED Requirements

### Requirement: Project inventory is available to workflow readers

The web application SHALL provide `/admin/projects` for users with
`workflow:read`. It SHALL load and display only projects returned for the
current tenant and SHALL provide a clear unauthorized, loading, error, and
empty state.

#### Scenario: User opens the project inventory

- **WHEN** a user with `workflow:read` opens `/admin/projects`
- **THEN** the page displays the current tenant and its available projects
- **AND** each project exposes a way to continue to its Intents or Workflows.

#### Scenario: User without workflow read access opens the project inventory

- **WHEN** a user without `workflow:read` opens `/admin/projects`
- **THEN** the route is denied by the existing authorization policy
- **AND** the page does not request project data.

### Requirement: Project is a clickable flow step

The shared Admin Console guide SHALL render the ordered path
`Tenant → Project → Intent → Workflow → State`. The Project step SHALL link to
`/admin/projects`, and the guide SHALL identify the selected project or the
existing automatic default scope.

#### Scenario: User follows the Project step

- **WHEN** a user selects Project in the Admin Console guide
- **THEN** the application navigates to `/admin/projects`.

### Requirement: Downstream views preserve project scope

The web application SHALL pass a selected `projectId` from project links to
Intent, Workflow, and State Builder data requests and destinations.

#### Scenario: User opens a project workflow

- **WHEN** a user selects Workflows for project `project-1`
- **THEN** the workflow inventory requests `projectId=project-1`
- **AND** each State Builder destination preserves `projectId=project-1`.

#### Scenario: User saves a scoped workflow

- **WHEN** the State Builder is opened with `projectId=project-1`
- **THEN** load, create, update, version, and publish requests carry the
  `X-Project-ID: project-1` scope.
