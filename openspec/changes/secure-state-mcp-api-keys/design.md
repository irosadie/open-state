## Context

See `proposal.md` for motivation. State MCP is currently exposed as a standard-library HTTP `/mcp` handler and individual tool calls receive `tenant` and sometimes `project` arguments. Existing human JWT/RBAC protects the admin HTTP application, but there is no machine-key entity, verifier, or State MCP HTTP authentication layer.

## Goals / Non-Goals

**Goals:**

- Make State MCP safe to expose to independently configured MCP clients.
- Bind every request to a tenant-scoped machine principal before protocol handling.
- Support least-privilege project and scope grants with operational lifecycle controls.
- Preserve existing human JWT/RBAC as the authority for managing machine keys.

**Non-Goals:**

- Implement OAuth, user-delegated identity, or provider credential storage.
- Reuse browser JWTs as machine credentials.
- Add an unauthenticated production compatibility mode.

## Decisions

### Store opaque keys as verifier plus metadata

Keys use a recognizable public prefix and a high-entropy secret returned only once. The database stores key metadata and a keyed cryptographic verifier, never the raw secret. A server configuration pepper participates in verifier computation, and comparisons are constant-time.

Alternative considered: encrypted raw key storage. Rejected because raw-key recovery is unnecessary and increases breach impact.

### Model key scope explicitly

An API key belongs to one tenant and has an allowlist of project IDs, one optional default project, named MCP scopes, expiration, revocation state, and usage metadata. Empty project scope is invalid; the default project, if present, must be allowed.

Alternative considered: a key per project only. Rejected because a controlled multi-project key is needed for cross-project integrations and can still be least-privilege.

### Authenticate before the Streamable MCP handler

A standard HTTP middleware validates the Bearer key for `/mcp`, constructs an immutable principal in request context, and passes authenticated requests to the existing MCP transport. Tool handlers read tenant/project/scope from that principal instead of trusting a tenant tool argument.

Alternative considered: validate the key in each tool handler. Rejected because initialization and future tools could be exposed accidentally, and policy would be duplicated.

### Separate key management from State MCP

Key lifecycle is provided through existing JWT/RBAC-protected admin HTTP routes. State MCP consumes keys only and never returns key secrets.

## Risks / Trade-offs

- [Existing MCP clients break when tenant arguments are removed] → Publish migration documentation, return explicit authentication/scope errors, and cover the new client configuration with protocol tests.
- [Key verifier pepper is lost or rotated incorrectly] → Make the pepper required configuration, document controlled rotation, and treat replacement as a key reissue operation in the first release.
- [Long-lived keys are leaked by a client] → Support expiration, immediate revocation, prefix-only logging, and last-used metadata for detection.
- [Tool scope drift] → Maintain a single server-side mapping from MCP tool names to scopes and test every registered tool.

## Migration Plan

1. Add key persistence, verifier configuration, and admin lifecycle operations behind existing human authorization.
2. Add State MCP middleware and principal context, then migrate tool contracts to server-derived tenant/project scope.
3. Update MCP docs/client examples and seed or test fixtures to issue development keys explicitly.
4. Release with no unauthenticated State MCP mode; revoke/reissue keys to rotate access.
