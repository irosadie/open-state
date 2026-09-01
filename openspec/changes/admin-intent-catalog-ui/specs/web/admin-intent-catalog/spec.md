## Purpose

Give Admin Console operators a clear, read-only view of the canonical intents
that connect user language to published workflows within the current project.

## ADDED Requirements

### Requirement: Admin Console exposes an intent catalog entry point

The web application SHALL provide an `/admin/intents` page and a visible
`Intents` navigation item for users who can read workflows. Users without
`workflow:read` SHALL not see the navigation item or receive an intent catalog
view.

#### Scenario: Authorized user opens the intent catalog

- **WHEN** a user with `workflow:read` opens the Admin Console
- **THEN** the console provides an `Intents` navigation item
- **AND** selecting it opens `/admin/intents`.

#### Scenario: Unauthorized user opens the Admin Console

- **WHEN** a user without `workflow:read` opens the Admin Console
- **THEN** the console does not present the `Intents` navigation item
- **AND** direct access to `/admin/intents` renders the existing unauthorized
  state instead of catalog data.

### Requirement: Intent catalog shows routing metadata and scope

The intent page SHALL load the current tenant/project catalog and show each
intent's canonical key, name, description, example utterances, and mapped
workflow. The page SHALL state that the current scope is the tenant's
`Default Project` when no project selector is available.

#### Scenario: Operator reviews a booking intent

- **WHEN** the catalog contains `BOOKING_PADEL`
- **THEN** the page shows the canonical key and its natural-language examples
- **AND** shows the workflow slug mapped to that intent
- **AND** provides an `Open Builder` destination for the mapped workflow.

#### Scenario: Catalog is empty

- **WHEN** the selected project has no routable intents
- **THEN** the page shows a clear empty state explaining that no published intent
  mappings are available
- **AND** it does not show a misleading workflow list.

### Requirement: Intent page handles loading and read failures clearly

The page SHALL show a loading state while the catalog is being fetched and a
recoverable error state when the API request fails. The page SHALL keep the
catalog read-only and SHALL not render intent mutation controls.

#### Scenario: Catalog request fails

- **WHEN** the intent catalog request returns an API error
- **THEN** the page displays the server error using the Admin Console's existing
  error presentation
- **AND** provides a retry action
- **AND** does not render stale or cross-scope intent rows as current data.

### Requirement: Admin Console setup guide shows the complete hierarchy

The shared setup guide SHALL present the ordered path `Tenant → Project → Intent
→ Workflow → Builder`. Tenant, Intent, Workflow, and Builder steps SHALL link to
their existing destinations when available; Project SHALL remain a context step
while project management is not available.

#### Scenario: Operator follows the setup guide

- **WHEN** an operator views the Admin Console overview, tenant settings, intent
  catalog, or workflow inventory
- **THEN** the guide displays five numbered steps in the order Tenant, Project,
  Intent, Workflow, Builder
- **AND** the current tenant and current `Default Project` scope are visible
- **AND** the guide identifies Intent as the routing choice that leads to a
  mapped Workflow.
