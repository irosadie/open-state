-- name: CreateAPIKey :one
INSERT INTO auth_api_keys (tenant_id, name, key_prefix, key_verifier, default_project_id, expires_at, created_by)
VALUES ($1, $2, $3, $4, sqlc.narg('default_project_id'), sqlc.narg('expires_at'), $5)
RETURNING id, tenant_id, name, key_prefix, key_verifier, default_project_id, expires_at, revoked_at, last_used_at, created_by, created_at, updated_at;

-- name: AddAPIKeyProject :exec
INSERT INTO auth_api_key_projects (api_key_id, project_id)
VALUES ($1, $2);

-- name: AddAPIKeyScope :exec
INSERT INTO auth_api_key_scopes (api_key_id, scope)
VALUES ($1, $2);

-- name: FindAPIKeyByPrefix :one
SELECT id, tenant_id, name, key_prefix, key_verifier, default_project_id, expires_at, revoked_at, last_used_at, created_by, created_at, updated_at
FROM auth_api_keys
WHERE key_prefix = $1
LIMIT 1;

-- name: ListAPIKeysByTenant :many
SELECT id, tenant_id, name, key_prefix, key_verifier, default_project_id, expires_at, revoked_at, last_used_at, created_by, created_at, updated_at
FROM auth_api_keys
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: ListAPIKeyProjects :many
SELECT project_id
FROM auth_api_key_projects
WHERE api_key_id = $1
ORDER BY project_id ASC;

-- name: ListAPIKeyScopes :many
SELECT scope
FROM auth_api_key_scopes
WHERE api_key_id = $1
ORDER BY scope ASC;

-- name: RevokeAPIKey :one
UPDATE auth_api_keys
SET revoked_at = NOW(), updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND revoked_at IS NULL
RETURNING id, tenant_id, name, key_prefix, key_verifier, default_project_id, expires_at, revoked_at, last_used_at, created_by, created_at, updated_at;

-- name: TouchAPIKeyLastUsed :exec
UPDATE auth_api_keys
SET last_used_at = NOW(), updated_at = NOW()
WHERE id = $1 AND revoked_at IS NULL;
