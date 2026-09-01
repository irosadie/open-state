## 1. API key persistence and security

- [x] 1.1 Add the API key persistence model, migrations, SQL queries, and generated data access for tenant ownership, project access, default project, scopes, expiration, revocation, and usage metadata.
- [x] 1.2 Implement opaque API key generation, secure verifier storage with a server-side pepper, and constant-time validation without persisting or logging raw secrets.
- [x] 1.3 Add the application service and repository contracts for creating, listing, revoking, and resolving API key principals, including tenant/project/scope invariants and audit events.

## 2. API key management API

- [x] 2.1 Add authenticated, RBAC-protected management endpoints and DTOs to create, list, and revoke API keys.
- [x] 2.2 Return the raw key only in a successful create response; expose only safe metadata in subsequent list, audit, and error responses.
- [x] 2.3 Document the management endpoints and request/response contracts in the split OpenAPI specification.
- [x] 2.4 Add API and application tests for tenant isolation, authorization, expiration, revocation, and one-time-secret behavior.

## 3. Authenticate and authorize State MCP

- [x] 3.1 Add HTTP Bearer authentication ahead of the State MCP transport and attach a resolved API key principal to the request context.
- [x] 3.2 Reject missing, malformed, unknown, expired, and revoked keys without exposing State MCP data or capabilities.
- [x] 3.3 Derive tenant identity from the authenticated principal, validate an explicitly requested project against the key allowlist, and use the key default project when appropriate.
- [x] 3.4 Map State MCP read, lifecycle-write, and capability-invocation operations to scopes; update tool contracts so tenant is no longer a trusted caller-supplied identity.
- [x] 3.5 Record safe API key usage and authorization-denial audit events, and update last-used metadata without storing credentials.

## 4. Migration guidance and verification

- [x] 4.1 Add configuration and client examples showing `Authorization: Bearer <api-key>`, key provisioning, project selection, and scope requirements.
- [x] 4.2 Add MCP protocol tests covering authenticated access, unauthorized access, scope denial, project denial, revocation, and tenant impersonation attempts.
- [x] 4.3 Run the relevant API, MCP, and documentation validation suites.
