## MODIFIED Requirements

### Requirement: Capability credential handling

The platform SHALL keep capability provider credentials as references, never as
plaintext in API responses (PRD §61, §91). MCP connection bearer credentials and
OAuth client, access, and refresh artifacts SHALL be managed as protected references
with scoped rotation, revocation, and access controls.

- The capability registry and project MCP connection registry SHALL store a
  `credential_reference` (a key/vault path) rather than the secret value.
- Resolving the actual secret SHALL happen in infrastructure, scoped and
  access-controlled.
- Secret values and secret-reference internals SHALL not be returned by API, MCP,
  audit, logging, workflow-definition, or browser responses.

#### Scenario: Registry stores references only

- **WHEN** a capability or project MCP connection is created/read
- **THEN** only safe credential configuration status is returned, never the secret value
- **AND** secret-reference internals are not disclosed.

#### Scenario: Revoke a connection credential

- **WHEN** an authorized operator revokes a project MCP connection credential
- **THEN** the platform prevents subsequent gateway use of that credential
- **AND** records a redacted audit event without exposing the revoked secret.
