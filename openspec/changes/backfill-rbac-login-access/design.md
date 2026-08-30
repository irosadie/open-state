## Context

See [proposal.md](proposal.md) for the motivation. `role_assignments` became
the source of effective permissions, but the initial RBAC migration did not
create assignments for users that predate it. The authorization service
correctly defaults an absent assignment to no permissions, so those users can
authenticate but cannot be routed to an application area.

The default tenant is created by a later migration than the RBAC table. The
repair must therefore run after both migrations and be safe on databases where
an operator has already assigned tenant roles manually.

## Goals / Non-Goals

**Goals:**

- Restore default-tenant access for legacy `USER` and `ADMIN` accounts.
- Preserve the tenant-specific role as the authorization source of truth.
- Make rollback remove only assignments inserted by this repair.

**Non-Goals:**

- Altering the role-permission matrix or the `/auth/me` response contract.
- Guessing a tenant for users beyond the established default tenant.
- Creating an account that is absent from the database being verified.

## Decisions

### Use a follow-up additive migration

Add a new Goose migration after default-tenant creation instead of modifying
the historical RBAC migration. Existing environments may already have applied
that migration, while a new database will apply the complete sequence in order.

Alternative considered: changing the original RBAC migration. This would leave
already-upgraded environments broken and creates divergent migration history.

### Map only legacy global roles and never overwrite an assignment

The migration will translate `ADMIN` to `OWNER` and `USER` to `VIEWER` for the
default tenant. It will insert only when the `(user_id, tenant_id)` pair has no
role assignment. Any explicit assignment remains authoritative.

Alternative considered: upserting the mapped role. This would silently replace
an operator's tenant-specific access decision.

### Record inserted rows for safe rollback

The migration will retain a small migration-owned record of role assignments it
inserted. Its Down block will use that record to remove only these rows before
removing the record itself.

Alternative considered: deleting every default-tenant assignment for legacy
users on rollback. This could delete assignments created after the migration
ran, so it is unsafe.

## Risks / Trade-offs

- [Legacy user does not belong to the default tenant] → The repair follows the
  existing single-default-tenant deployment model; multi-tenant imports remain
  an explicit administrative migration.
- [Migration is applied to a partially provisioned database] → It runs after
  default tenant provisioning and uses a fixed, existing default tenant ID.
- [Local browser keeps an old authorization snapshot] → Verify with a fresh
  login/session after migration and API restart.

## Migration Plan

1. Add and review the additive migration, including its targeted rollback.
2. Apply it to the local PostgreSQL database after the existing RBAC and tenant
   migrations.
3. Restart the local API so its schema state and running binary are aligned.
4. Verify health plus an authenticated legacy-user authorization snapshot where
   a fixture account is available.
5. Roll back the new migration only if needed; the Down block removes solely
   rows recorded by the migration.

## Open Questions

- The reported email is not present in the local database, so its credentials
  cannot be used for local end-to-end verification. A fresh or correctly
  targeted account/session will be required after implementation.
