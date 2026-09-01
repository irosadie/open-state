## MODIFIED Requirements

### Requirement: Admin Console navigation reflects existing permissions

The web application SHALL provide a shared `/admin` layout and landing view. It
SHALL derive visible navigation sections and enabled actions from the authenticated
user’s existing tenant-scoped permissions while relying on APIs for final
authorization. The navigation SHALL include State MCP API Keys for users with
`api_key:read`.

#### Scenario: Viewer opens the Admin Console

- **WHEN** a Viewer opens `/admin`
- **THEN** the console shows only sections the Viewer can read, including
  permitted workflow, instance, event, audit, capability, or binding views
- **AND** it does not present tenant/member mutation controls.

#### Scenario: Owner opens the Admin Console

- **WHEN** an Owner opens `/admin`
- **THEN** the console provides the existing tenant, member, workflow, runtime,
  audit, and capability sections
- **AND** provides State MCP API Keys when `api_key:read` is granted.
