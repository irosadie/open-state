## MODIFIED Requirements

### Requirement: Admin Console navigation reflects existing permissions

The web application SHALL provide a shared `/admin` layout and landing view. It
SHALL derive visible navigation sections and enabled actions from the authenticated
user's existing tenant-scoped permissions while relying on APIs for final
authorization. It SHALL include the project-scoped MCP Connections destination for
users permitted to read or manage project MCP connections.

#### Scenario: Viewer opens the Admin Console

- **WHEN** a Viewer opens `/admin`
- **THEN** the console shows only sections the Viewer can read, including permitted workflow, instance, event, audit, capability, binding, or project MCP connection views
- **AND** it does not present tenant/member mutation controls.

#### Scenario: Owner opens the Admin Console

- **WHEN** an Owner opens `/admin`
- **THEN** the console provides tenant settings and members/roles navigation in addition to the sections allowed by that Owner's permissions
- **AND** it provides the project MCP Connections destination.
