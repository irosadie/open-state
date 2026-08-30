## 1. RBAC migration repair

- [x] 1.1 Add an additive Goose migration after default-tenant provisioning that maps legacy `USER` and `ADMIN` accounts to `VIEWER` and `OWNER` role assignments.
- [x] 1.2 Preserve pre-existing default-tenant assignments and record only migration-created rows so the Down migration is targeted and safe.
- [x] 1.3 Add a regression test or migration verification covering the legacy mappings, generated permissions, and preservation of an explicit assignment.

## 2. Local environment verification

- [x] 2.1 Apply pending database migrations to the local PostgreSQL environment and confirm the expected schema version.
- [x] 2.2 Restart the local API from the current source, then verify its health endpoint and authorization response with an available account or fixture.
- [x] 2.3 Confirm a fresh login session resolves an accessible application area; document if the reported email is absent from the target database.

## 3. Quality checks

- [x] 3.1 Run the affected backend tests and static checks.
- [x] 3.2 Run `openspec validate backfill-rbac-login-access --strict` and update task completion status with the evidence.
