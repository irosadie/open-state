## Purpose

Provide a permission-aware Admin Console that centralizes tenant operations and links to the established Builder, Runtime Inspector, Audit, and Capabilities experiences.

## ADDED Requirements

### Requirement: Admin Console navigation reflects existing permissions

The web application SHALL provide a shared `/admin` layout and landing view. It SHALL derive visible navigation sections and enabled actions from the authenticated user's existing tenant-scoped permissions while relying on APIs for final authorization.

#### Scenario: Viewer opens the Admin Console

- **WHEN** a Viewer opens `/admin`
- **THEN** the console shows only sections the Viewer can read, including permitted workflow, instance, event, audit, capability, or binding views
- **AND** it does not present tenant/member mutation controls.

#### Scenario: Owner opens the Admin Console

- **WHEN** an Owner opens `/admin`
- **THEN** the console provides tenant settings and members/roles navigation in addition to the sections allowed by that Owner's permissions.

### Requirement: Tenant and membership management requires deliberate confirmation

The web application SHALL provide tenant settings and members/roles management pages for authorized users. It SHALL require an explicit confirmation before submitting a tenant update, role assignment/replacement, or membership removal and SHALL render API authorization, validation, and last-Owner conflict errors clearly.

#### Scenario: Owner changes a membership role

- **WHEN** an Owner selects a valid new role and confirms the change
- **THEN** the application submits the tenant-scoped role mutation
- **AND** refreshes affected membership data after server success
- **AND** displays the server error without local state mutation if the request fails.

### Requirement: Workflow inventory routes to Builder lifecycle ownership

The web application SHALL provide a tenant-scoped workflow inventory in the Admin Console and SHALL route workflow authoring and version-detail actions to the destinations owned by the Builder lifecycle contract. It SHALL NOT duplicate workflow authoring, publishing, or version management UI.

#### Scenario: Administrator opens a workflow from inventory

- **WHEN** a user with workflow read access selects a workflow in the console inventory
- **THEN** the application navigates to the Builder lifecycle or version-detail destination
- **AND** does not render a competing authoring flow in the console page.

### Requirement: Instance management reuses Runtime Inspector detail

The web application SHALL provide an instance list and SHALL present only lifecycle controls permitted to the current user. It SHALL require confirmation before submitting suspend, resume, or retry and SHALL route instance detail to the Runtime Inspector contract.

#### Scenario: Operator retries a failed instance

- **WHEN** an Operator with `instance:retry` confirms retry for an eligible tenant instance
- **THEN** the application submits the retry request
- **AND** refreshes affected instance and Runtime Inspector queries only after success
- **AND** renders conflict feedback when the server rejects the command.

#### Scenario: User opens instance detail

- **WHEN** a user selects an instance in the Admin Console
- **THEN** the application navigates to the Runtime Inspector detail destination
- **AND** state, context, timeline, and debug presentation remain owned by that destination.

### Requirement: Event browser is read-only

The web application SHALL provide a tenant-scoped, paginated event browser with safe filters and event detail. It SHALL link related instance or audit context when identifiers are available and SHALL not render edit, delete, replay, or injection controls.

#### Scenario: User filters events by correlation identifier

- **WHEN** a user applies a supported correlation identifier filter
- **THEN** the application displays the matching event page for the current tenant
- **AND** keeps event data read-only.

### Requirement: Existing audit and capability administration are integrated

The web application SHALL include the established Audit and Capabilities pages in Admin Console navigation without replacing or regressing their existing behavior.

#### Scenario: Administrator navigates to audit from the console

- **WHEN** a user selects Audit from a permitted Admin Console navigation item
- **THEN** the application renders the existing audit experience under the shared console navigation
- **AND** retains its existing authorization and filtering behavior.
