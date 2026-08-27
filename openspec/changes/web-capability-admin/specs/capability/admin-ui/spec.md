# capability/admin-ui Specification

## Purpose

Define the tenant-scoped admin UI for managing the Capability Registry: list, create,
view, edit, and disable capabilities, consuming the capability admin API.

## ADDED Requirements

### Requirement: List capabilities

The admin UI SHALL display the tenant's capability registry with filtering.

#### Scenario: View registry

- **WHEN** an admin opens the capability list page
- **THEN** the UI shows capabilities with name, description, provider type, status, and
  version
- **AND** lets the admin filter by provider type and status

### Requirement: Create capability

The admin UI SHALL let an admin create a capability through a validated form.

#### Scenario: Submit create form

- **WHEN** an admin fills the create form and submits
- **THEN** the UI validates the payload with the shared schema
- **AND** calls the create endpoint and shows the created capability
- **AND** surfaces validation or conflict errors inline

### Requirement: View and edit capability

The admin UI SHALL let an admin view and edit a capability's details.

#### Scenario: View detail

- **WHEN** an admin opens a capability detail page
- **THEN** the UI shows description, provider, schemas, status, version, and the
  credential reference (not the secret)

#### Scenario: Edit capability

- **WHEN** an admin edits and saves capability fields
- **THEN** the UI validates and calls the update endpoint
- **AND** refreshes the detail view

### Requirement: Disable capability

The admin UI SHALL let an admin delete or disable a capability.

#### Scenario: Disable

- **WHEN** an admin confirms disabling a capability
- **THEN** the UI calls the delete/disable endpoint
- **AND** updates the list accordingly

### Requirement: Secret safety

The admin UI SHALL only ever display or collect the `credential_reference` string.

#### Scenario: No secret input

- **WHEN** the UI renders capability forms or details
- **THEN** only the credential reference is shown or collected
- **AND** never the resolved secret value
