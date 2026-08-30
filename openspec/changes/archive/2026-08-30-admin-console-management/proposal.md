# Change: Add governed Admin Console management surfaces

## Why

The application already provides audit-log and capability administration pages, but it has no single governed console for operating a tenant. Tenant membership and roles, workflow inventory, runtime instances, and events still require fragmented or manual access. This makes routine operations hard to discover and increases the risk of bypassing tenant boundaries or audit controls.

## What Changes

- Add tenant-scoped APIs for an Owner to view and update the current tenant profile, list its memberships, and assign, replace, or remove a member's tenant role.
- Add tenant-scoped runtime administration APIs for authorized lifecycle commands (`suspend`, `resume`, and `retry`) and read-only event browsing.
- Add a permission-aware Admin Console shell with management pages for tenant settings, members and roles, workflows, instances, and events. The shell incorporates the existing Audit and Capabilities pages.
- Reuse the Builder lifecycle change for workflow authoring/version detail and the Runtime Inspector change for instance state, context, timeline, and debug detail; this change supplies console entry points and operational controls only.
- Require server-side authorization, tenant isolation, confirmation for mutating console actions, and audit records for every tenant, membership, role, and runtime lifecycle mutation.

## Capabilities

### New Capabilities

- `backend/admin-identity-management`: Provides current-tenant settings and membership/role administration without cross-tenant access.
- `backend/admin-runtime-operations`: Provides authorized instance lifecycle commands and immutable event browsing for the Admin Console.
- `web/admin-console-management`: Provides the permission-aware Admin Console navigation and management surfaces, integrating existing Audit and Capabilities pages.

## Impact

- Affected backend areas: authenticated HTTP routes/controllers, tenant and role-assignment queries/services, runtime orchestration commands, event queries, authorization, and audit logging.
- Affected frontend areas: Admin routes/layout, API schemas/types/hooks, workflow and runtime entry points, and existing Audit/Capabilities navigation.
- Dependencies: existing RBAC/audit/capability administration; the active `complete-builder-lifecycle` change for workflow lifecycle destinations; and the active `runtime-inspector-debug` change for runtime detail views.
- No third-party LLM, RAG, or MCP provider is called by this console. Provider-facing runtime detail remains governed by the Runtime Inspector contract.

## Non-goals

- Creating or provisioning tenants, cross-tenant administration, invitations, SCIM, global account deletion, or authentication/SSO lifecycle changes.
- Rebuilding workflow authoring, publishing, version comparison, simulation, or Runtime Inspector state/context/timeline/debug views.
- Editing, deleting, replaying, or otherwise mutating stored events.
- Exposing raw prompts, provider credentials, model responses, or RAG documents.
