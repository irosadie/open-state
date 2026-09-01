## MODIFIED Requirements

### Requirement: Admin Console navigation reflects existing permissions

The web application SHALL provide a shared `/admin` layout and landing view. It
SHALL derive visible navigation sections and enabled actions from the authenticated
user's existing tenant-scoped permissions while relying on APIs for final
authorization. The navigation SHALL include the read-only Intents section for
users with `workflow:read`.

#### Scenario: Viewer opens the Admin Console

- **WHEN** a Viewer opens `/admin`
- **THEN** the console shows only sections the Viewer can read, including permitted workflow, intent, instance, event, audit, capability, or binding views
- **AND** it does not present tenant/member mutation controls.

#### Scenario: Owner opens the Admin Console

- **WHEN** an Owner opens `/admin`
- **THEN** the console provides tenant settings and members/roles navigation in addition to the sections allowed by that Owner's permissions
- **AND** provides the read-only Intents navigation alongside Workflows when `workflow:read` is granted.

## ADDED Requirements

### Requirement: Admin Console makes the tenant-to-builder hierarchy explicit

The shared Admin Console setup guide SHALL show the ordered hierarchy `Tenant →
Project → Intent → Workflow → Builder` wherever the console explains how a
workflow is created or operated. It SHALL identify `Default Project` as the
current automatic project scope until project management is available.

#### Scenario: Workflow inventory shows intent as an explicit step

- **WHEN** an authorized user opens the workflow inventory
- **THEN** the setup guide shows Project followed by Intent followed by Workflow
  and Builder
- **AND** the page states which project scope owns the listed workflows.

#### Scenario: Existing Admin Console pages keep the same hierarchy

- **WHEN** an authorized user opens the Admin Console overview, tenant settings,
  or intent catalog
- **THEN** the same five-step hierarchy is shown in the same order
- **AND** selecting the Intent step leads to the intent catalog when the user has
  workflow read access.
