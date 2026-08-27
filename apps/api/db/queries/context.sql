-- Scoped runtime context (PRD §23, §36, §131) and persistent memory references
-- (PRD §24, §43.2). All queries are tenant-scoped (PRD §4, §96).

-- name: UpsertContext :one
INSERT INTO context_records (tenant_id, scope_type, scope_id, key, value)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (tenant_id, scope_type, scope_id, key)
DO UPDATE SET
  value      = EXCLUDED.value,
  version    = context_records.version + 1,
  updated_at = NOW()
WHERE context_records.version = $6
RETURNING id, tenant_id, scope_type, scope_id, key, value, version, created_at, updated_at;

-- name: FindContextByScope :one
SELECT id, tenant_id, scope_type, scope_id, key, value, version, created_at, updated_at
FROM context_records
WHERE tenant_id = $1 AND scope_type = $2 AND scope_id = $3 AND key = $4
LIMIT 1;

-- name: ListContextByScope :many
SELECT id, tenant_id, scope_type, scope_id, key, value, version, created_at, updated_at
FROM context_records
WHERE tenant_id = $1 AND scope_type = $2 AND scope_id = $3
ORDER BY created_at ASC;

-- name: DeleteContext :exec
DELETE FROM context_records
WHERE tenant_id = $1 AND scope_type = $2 AND scope_id = $3 AND key = $4;

-- Persistent memory references (PRD §24, §43.2).

-- name: UpsertMemoryReference :one
INSERT INTO memory_references (tenant_id, owner_type, owner_id, name, value, source_workflow_instance_id)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (tenant_id, owner_type, owner_id, name)
DO UPDATE SET
  value                       = EXCLUDED.value,
  source_workflow_instance_id = EXCLUDED.source_workflow_instance_id,
  updated_at                  = NOW()
RETURNING id, tenant_id, owner_type, owner_id, name, value, source_workflow_instance_id, created_at, updated_at;

-- name: FindMemoryReference :one
SELECT id, tenant_id, owner_type, owner_id, name, value, source_workflow_instance_id, created_at, updated_at
FROM memory_references
WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3 AND name = $4
LIMIT 1;

-- name: ListMemoryByOwner :many
SELECT id, tenant_id, owner_type, owner_id, name, value, source_workflow_instance_id, created_at, updated_at
FROM memory_references
WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3
ORDER BY created_at ASC;

-- name: DeleteMemoryReference :exec
DELETE FROM memory_references
WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3 AND name = $4;
