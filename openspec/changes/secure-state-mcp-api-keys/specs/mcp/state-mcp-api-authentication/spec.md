## Purpose

Secure State MCP connections with a machine principal whose tenant, project access, and permissions are established by the server rather than asserted by an LLM tool call.

## ADDED Requirements

### Requirement: State MCP authenticates bearer API keys

The State MCP `/mcp` endpoint SHALL require an opaque API key in the `Authorization: Bearer <key>` HTTP header before processing MCP protocol requests. The endpoint SHALL reject missing, malformed, unknown, expired, or revoked keys without exposing tenant or project data.

#### Scenario: Authenticate an active API key

- **WHEN** a client sends a valid active API key in the bearer authorization header
- **THEN** the server associates the MCP connection with that key's machine principal before serving MCP requests

#### Scenario: Reject an unauthenticated MCP request

- **WHEN** a client sends an MCP request without a valid bearer API key
- **THEN** the server rejects the HTTP request and does not execute an MCP tool

### Requirement: MCP tenant and project scope are server-derived

The server SHALL derive tenant identity exclusively from the authenticated API key. MCP tools SHALL NOT accept tenant identity as a client-controlled authorization input. A project requested by a tool SHALL be accepted only when it belongs to the key's tenant and is in the key's allowed project set; when omitted, the key's default project SHALL be used.

#### Scenario: Resolve an intent using the default project

- **WHEN** a key with a default project invokes an intent tool without a project argument
- **THEN** the tool resolves the intent within the key's tenant and default project

#### Scenario: Reject a project outside the key scope

- **WHEN** a client requests a project that is not allowed by its key
- **THEN** the tool returns an authorization error and does not read or mutate that project

### Requirement: MCP scopes gate tool actions

The server SHALL enforce explicit API-key scopes for read-only state and intent tools, state-changing lifecycle tools, and capability invocation tools. A tool call without the necessary scope SHALL be denied before application-layer work begins.

#### Scenario: Deny an unauthorized state mutation

- **WHEN** a read-only key invokes a state-changing MCP tool
- **THEN** the tool is denied and workflow state remains unchanged

### Requirement: Auth decisions are auditable

The server SHALL record auditable metadata for successful key use and denied MCP authorization requests without recording the raw API key or authorization header.

#### Scenario: Record a denied request safely

- **WHEN** a revoked key attempts an MCP request
- **THEN** an audit record identifies the key metadata and denial reason without containing the secret
