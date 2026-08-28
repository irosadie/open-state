-- Tenant-scoped RBAC queries (PRD 80, 81). Every query takes an explicit
-- tenant_id (PRD 4, 96) and filters by it so cross-tenant access is impossible
-- at the data-access layer.

-- name: AssignRole :one
INSERT INTO role_assignments (user_id, tenant_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, tenant_id) DO UPDATE SET role = EXCLUDED.role, updated_at = NOW()
RETURNING id, user_id, tenant_id, role, created_at, updated_at;

-- name: FindRoleByUserAndTenant :one
SELECT id, user_id, tenant_id, role, created_at, updated_at
FROM role_assignments
WHERE user_id = $1 AND tenant_id = $2
LIMIT 1;

-- name: ListRolesByTenant :many
SELECT id, user_id, tenant_id, role, created_at, updated_at
FROM role_assignments
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: RemoveRoleAssignment :exec
DELETE FROM role_assignments
WHERE user_id = $1 AND tenant_id = $2;
