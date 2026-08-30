-- Tenant profile and membership queries for the Admin Console. All lookups are
-- explicitly tenant-scoped; the caller never supplies a tenant through a body.

-- name: FindTenantByID :one
SELECT id, name, slug, description, created_at, updated_at
FROM tenants
WHERE id = $1
LIMIT 1;

-- name: UpdateTenantProfile :one
UPDATE tenants
SET name = $2, slug = $3, description = $4, updated_at = NOW()
WHERE id = $1
RETURNING id, name, slug, description, created_at, updated_at;

-- name: ListAdminMemberships :many
SELECT
  ra.id AS role_assignment_id,
  ra.user_id,
  ra.tenant_id,
  ra.role,
  ra.created_at AS role_created_at,
  ra.updated_at AS role_updated_at,
  u.email,
  u.name,
  u.status,
  u.photo
FROM role_assignments AS ra
JOIN users AS u ON u.id = ra.user_id
WHERE ra.tenant_id = @tenant_id
  AND (
    sqlc.narg('search')::text IS NULL
    OR u.name ILIKE '%' || sqlc.narg('search') || '%'
    OR u.email ILIKE '%' || sqlc.narg('search') || '%'
  )
ORDER BY u.name ASC, ra.created_at DESC
LIMIT @page_size OFFSET @page_offset;

-- name: CountAdminMemberships :one
SELECT COUNT(*)
FROM role_assignments AS ra
JOIN users AS u ON u.id = ra.user_id
WHERE ra.tenant_id = @tenant_id
  AND (
    sqlc.narg('search')::text IS NULL
    OR u.name ILIKE '%' || sqlc.narg('search') || '%'
    OR u.email ILIKE '%' || sqlc.narg('search') || '%'
  );

-- name: FindAdminMembership :one
SELECT
  ra.id AS role_assignment_id,
  ra.user_id,
  ra.tenant_id,
  ra.role,
  ra.created_at AS role_created_at,
  ra.updated_at AS role_updated_at,
  u.email,
  u.name,
  u.status,
  u.photo
FROM role_assignments AS ra
JOIN users AS u ON u.id = ra.user_id
WHERE ra.tenant_id = $1 AND ra.user_id = $2
LIMIT 1;

-- name: CountAdminOwners :one
SELECT COUNT(*)
FROM role_assignments
WHERE tenant_id = $1 AND role = 'OWNER';

-- name: UpsertAdminMembershipRole :one
INSERT INTO role_assignments (user_id, tenant_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, tenant_id)
DO UPDATE SET role = EXCLUDED.role, updated_at = NOW()
RETURNING id, user_id, tenant_id, role, created_at, updated_at;

-- name: RemoveAdminMembership :exec
DELETE FROM role_assignments
WHERE tenant_id = $1 AND user_id = $2;
