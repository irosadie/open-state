## Why

State MCP currently accepts tenant and project identifiers from tool arguments without authenticating the HTTP connection. An external MCP client therefore needs a server-authenticated principal before it can safely read state or invoke an authorized capability.

## What Changes

- Add opaque, revocable API keys for machine-to-machine State MCP access; show the raw secret only at key creation and retain only a secure verifier thereafter.
- Authenticate `/mcp` using `Authorization: Bearer <api-key>` before MCP protocol handling.
- Bind an authenticated key to exactly one tenant, an optional default project, an allowlist of projects, and explicit scopes.
- Derive tenant scope from the authenticated key; validate any requested project against the key's allowed projects and remove tenant identity as a client-controlled trust input.
- Provide authenticated key lifecycle operations for creation, listing metadata, revocation, and project/scope configuration, plus audit events for key lifecycle and denied MCP requests.
- **BREAKING**: State MCP clients must send a bearer API key; direct unauthenticated `/mcp` calls and tenant impersonation through tool arguments are rejected.

## Capabilities

### New Capabilities

- `mcp/state-mcp-api-authentication`: Authenticate State MCP HTTP connections and enforce tenant, project, and scope authorization for every MCP tool call.
- `api/api-key-management`: Create, inspect, revoke, and audit scoped machine API keys used by State MCP clients.

### Modified Capabilities

- None.

## Impact

- `apps/api`: database schema, sqlc queries, domain entities/repositories, application services, State MCP HTTP middleware/tool contracts, audit logging, HTTP administration endpoints, and tests.
- `docs/openapi`: API-key lifecycle contract and State MCP client configuration documentation.
- MCP clients: configure `Authorization: Bearer osk_...`; stop passing a trusted tenant identifier with each call.

## Non-goals

- Replacing the existing human JWT/SSO authentication used by the admin HTTP application.
- OAuth dynamic client registration, delegated end-user identity, or externally managed secrets.
- Persisting third-party provider credentials in the State MCP API key.
