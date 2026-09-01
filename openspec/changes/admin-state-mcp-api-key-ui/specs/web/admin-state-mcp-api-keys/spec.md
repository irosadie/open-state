## Purpose

Provide a permission-aware Admin Console surface for managing the tenant-scoped
machine credentials used by State MCP clients.

## ADDED Requirements

### Requirement: API-key management has an Admin Console entry point

The web application SHALL provide `/admin/api-keys` and SHALL expose it from
the `/admin` overview and shared Admin Console navigation only when the current
user has `api_key:read`.

#### Scenario: Authorized operator opens API keys

- **WHEN** a user with `api_key:read` opens `/admin`
- **THEN** the overview and sidebar show the State MCP API Keys entry point
- **AND** selecting it opens `/admin/api-keys`.

#### Scenario: Unauthorized user opens API keys

- **WHEN** a user without `api_key:read` opens `/admin/api-keys`
- **THEN** the page renders an unauthorized state
- **AND** it does not issue a list request.

### Requirement: Operators can create scoped State MCP keys

The page SHALL allow a user with `api_key:create` to submit a validated key
name, one or more existing project IDs, an optional default project, one or more
supported MCP scopes, and an optional future expiration to the existing API.

#### Scenario: Key creation succeeds

- **WHEN** an authorized operator submits valid key details
- **THEN** the browser sends the tenant header and validated payload to
  `POST /api/api-keys`
- **AND** refreshes the metadata list after success
- **AND** displays the returned raw `osk_...` secret exactly once.

#### Scenario: Invalid key details are rejected locally

- **WHEN** the name, project list, default project, scopes, or expiration is
  invalid
- **THEN** the page shows a field-level validation message
- **AND** does not issue the create request.

### Requirement: Raw secrets are handled as one-time credentials

The page SHALL clearly warn that the raw key is shown once, SHALL provide a
copy action, and SHALL never reconstruct or display a raw key from list data.

#### Scenario: Operator dismisses the new secret

- **WHEN** the operator copies or dismisses a newly created key
- **THEN** the page clears the raw secret from rendered state
- **AND** the metadata list continues to show only the public prefix.

### Requirement: Operators can inspect and revoke key metadata

The page SHALL list safe metadata including name, prefix, projects, scopes,
expiration, last use, and active/revoked status. A user with `api_key:revoke`
SHALL be able to revoke an active key only after explicit confirmation.

#### Scenario: Operator revokes an active key

- **WHEN** an authorized operator confirms revocation
- **THEN** the browser calls `POST /api/api-keys/{id}/revoke`
- **AND** refreshes the list after success
- **AND** renders the key as revoked without exposing its secret.

### Requirement: API failures remain recoverable and tenant-scoped

The page SHALL use the existing authenticated BFF path, send the configured
`X-Tenant-ID`, validate successful responses, and show retryable error feedback
for failed list/create/revoke operations. It SHALL not send `MCP_API_KEY_PEPPER`
or any provider credential to the browser.

#### Scenario: Metadata loading fails

- **WHEN** the API rejects the list request
- **THEN** the page shows the server error and a retry action
- **AND** it does not present the failed response as valid key metadata.
