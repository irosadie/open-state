## Purpose

Allow tenant administrators to issue and revoke scoped machine credentials for State MCP clients while keeping raw credentials secret and lifecycle changes auditable.

## ADDED Requirements

### Requirement: Administrators can create a scoped API key

An authorized tenant administrator SHALL be able to create an API key with a name, tenant-bound project access, optional default project, explicit MCP scopes, and optional expiration. The raw key SHALL be returned exactly once at creation and SHALL NOT be retrievable afterward.

#### Scenario: Create a project-scoped key

- **WHEN** a tenant administrator creates a key for an allowed project and selected MCP scopes
- **THEN** the system returns the raw key once with non-secret metadata including key id, prefix, tenant, project scope, scopes, and expiration

#### Scenario: Reject an invalid project scope

- **WHEN** an administrator includes a project belonging to another tenant
- **THEN** key creation is rejected and no API key is created

### Requirement: Key metadata can be listed and revoked

An authorized tenant administrator SHALL be able to list non-secret metadata for tenant API keys and revoke an active key. Revocation SHALL take effect for subsequent State MCP requests.

#### Scenario: Revoke an active key

- **WHEN** an administrator revokes an active API key
- **THEN** the key is shown as revoked and later State MCP authentication with it is denied

### Requirement: API key secrets are not retained or exposed

The system SHALL retain only a secure verifier and non-secret identifying prefix for an API key. API responses, audit records, application logs, and list endpoints SHALL NOT expose the raw key after its one-time creation response.

#### Scenario: List keys after creation

- **WHEN** an administrator lists tenant API keys after creating one
- **THEN** the response includes metadata and status but no raw API key secret

### Requirement: API-key lifecycle is authorized and audited

Key creation, revocation, and denied management requests SHALL require the tenant's existing human authorization controls and SHALL create audit records with actor and key metadata where applicable.

#### Scenario: Deny a non-administrator

- **WHEN** a user without the required tenant permission attempts to revoke a key
- **THEN** the request is denied and an authorization audit record is created
