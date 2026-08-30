-- +goose Up
-- Preserve default-tenant access for users created before tenant-scoped RBAC.
-- Explicit assignments remain authoritative: only absent pairs are inserted.
CREATE TABLE auth_role_backfill_records (
  id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  role_assignment_id UUID        NOT NULL UNIQUE REFERENCES role_assignments(id) ON DELETE CASCADE,
  initial_role       VARCHAR(32) NOT NULL,
  created_at         TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMP(6) NOT NULL DEFAULT NOW()
);

CREATE INDEX auth_role_backfill_records_assignment_id_idx
  ON auth_role_backfill_records (role_assignment_id);

WITH inserted_assignments AS (
  INSERT INTO role_assignments (user_id, tenant_id, role)
  SELECT
    users.id,
    '00000000-0000-0000-0000-000000000001'::uuid,
    CASE users.role::text
      WHEN 'ADMIN' THEN 'OWNER'
      WHEN 'USER' THEN 'VIEWER'
    END
  FROM users
  WHERE users.role IN ('USER', 'ADMIN')
  ON CONFLICT (user_id, tenant_id) DO NOTHING
  RETURNING id, role
)
INSERT INTO auth_role_backfill_records (role_assignment_id, initial_role)
SELECT id, role
FROM inserted_assignments;

-- +goose Down
-- Do not remove assignments that an administrator changed after the backfill.
DELETE FROM role_assignments AS assignments
USING auth_role_backfill_records AS backfills
WHERE assignments.id = backfills.role_assignment_id
  AND assignments.role = backfills.initial_role;

DROP INDEX IF EXISTS auth_role_backfill_records_assignment_id_idx;
DROP TABLE IF EXISTS auth_role_backfill_records;
