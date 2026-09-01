## Purpose

Protect registered external MCP connections and their credentials against secret
exposure, unsafe egress, and unbounded local process execution in production.

## ADDED Requirements

### Requirement: MCP credential lifecycle

The platform SHALL support protected bearer credential rotation/revocation and OAuth
authorization, refresh, disconnect, and expiry handling for project MCP connections.
Plaintext credentials and OAuth artifacts SHALL never be returned by APIs, MCP tools,
logs, audits, workflow definitions, or browser state.

#### Scenario: Rotate a bearer credential

- **WHEN** an authorized operator replaces a connection's bearer credential
- **THEN** future provider calls use the replacement credential
- **AND** no API response or audit event reveals either secret value.

#### Scenario: OAuth token expires

- **WHEN** an OAuth-backed provider access token expires
- **THEN** the platform refreshes or marks the connection action-required according to its configured authorization capability
- **AND** returns a safe classified status without exposing OAuth artifacts.

### Requirement: Safe outbound provider access

The platform SHALL validate external MCP endpoint resolution and execution against
configured egress policy, including HTTPS requirements where configured, host/network
allowlists, redirect restrictions, DNS revalidation, request limits, and denial of
private or otherwise prohibited destinations.

#### Scenario: Block prohibited endpoint

- **WHEN** an operator registers or execution resolves an endpoint prohibited by egress policy
- **THEN** the platform rejects or blocks the external request
- **AND** records a safe security event without attempting the connection.

### Requirement: Bounded STDIO execution

The platform SHALL allow STDIO MCP execution only through reviewed runtime profiles
with an executable allowlist, fixed argument policy, isolated environment, resource
limits, and no inherited secret values beyond the connection's explicit secret
references.

#### Scenario: Unapproved STDIO command

- **WHEN** a connection requests an executable outside the approved STDIO profile
- **THEN** the platform refuses to start the process
- **AND** reports a safe configuration error.
