-- +goose Up
-- Tenant-scoped RBAC (PRD 80, 81). role_assignments is the source of truth for
-- a user's effective role per tenant, replacing the single global users.role
-- column (which is kept for backward compatibility during the transition and
-- is deprecated). role is VARCHAR + Go typed constants (not PostgreSQL ENUM) —
-- see .agents/skills/db-sqlc-schema. The effective role for authorization is
-- read from role_assignments only.

CREATE TABLE role_assignments (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  tenant_id  UUID        NOT NULL,
  role       VARCHAR(32) NOT NULL,                -- OWNER / ADMIN / EDITOR / OPERATOR / VIEWER (PRD 80)
  created_at TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT role_assignments_user_tenant_unique UNIQUE (user_id, tenant_id)
);

CREATE INDEX role_assignments_tenant_id_idx  ON role_assignments (tenant_id);
CREATE INDEX role_assignments_user_id_idx    ON role_assignments (user_id);

-- +goose Down
DROP INDEX IF EXISTS role_assignments_user_id_idx;
DROP INDEX IF EXISTS role_assignments_tenant_id_idx;
DROP TABLE IF EXISTS role_assignments;
