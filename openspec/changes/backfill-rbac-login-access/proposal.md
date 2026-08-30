## Why

Users created before tenant-scoped RBAC can authenticate successfully but receive no permissions, which leaves them without an accessible application area. The RBAC rollout needs a safe upgrade path so existing accounts retain access after the new authorization model is applied.

## What Changes

- Add a database migration that backfills default-tenant role assignments for existing users from their legacy global role.
- Preserve any tenant role assignment that already exists, rather than replacing it during the backfill.
- Verify the local login path after the migration so authenticated legacy users receive a resolvable application area.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `auth/role-permission`: Existing users must receive a default-tenant role assignment during the RBAC migration without overriding explicit tenant roles.

## Impact

- `apps/api/db/migrations` receives an additive Goose migration.
- Existing authorization resolution through `/auth/me` gains valid role-derived permissions for migrated accounts.
- Local API migration and login verification are required; no public API contract changes are expected.

## Non-goals

- Creating or resetting user accounts that do not exist in the target database.
- Changing the role-permission matrix or adding a new login flow.
- Replacing explicit tenant role assignments.
