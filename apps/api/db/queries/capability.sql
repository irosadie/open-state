-- name: CreateCapability :one
INSERT INTO capabilities (tenant_id, name, description, provider_type, provider_id, provider_tool, input_schema, output_schema, status, version, credential_reference)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, tenant_id, name, description, provider_type, provider_id, provider_tool, input_schema, output_schema, status, version, credential_reference, created_at, updated_at;

-- name: FindCapabilityByID :one
SELECT id, tenant_id, name, description, provider_type, provider_id, provider_tool, input_schema, output_schema, status, version, credential_reference, created_at, updated_at
FROM capabilities
WHERE id = $1 AND tenant_id = $2
LIMIT 1;

-- name: FindCapabilityByName :one
SELECT id, tenant_id, name, description, provider_type, provider_id, provider_tool, input_schema, output_schema, status, version, credential_reference, created_at, updated_at
FROM capabilities
WHERE name = $1 AND tenant_id = $2
LIMIT 1;

-- name: ListCapabilitiesByTenant :many
SELECT id, tenant_id, name, description, provider_type, provider_id, provider_tool, input_schema, output_schema, status, version, credential_reference, created_at, updated_at
FROM capabilities
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: ListCapabilitiesByTenantFiltered :many
SELECT id, tenant_id, name, description, provider_type, provider_id, provider_tool, input_schema, output_schema, status, version, credential_reference, created_at, updated_at
FROM capabilities
WHERE tenant_id = $1
  AND ($2::VARCHAR = '' OR provider_type = $2)
  AND ($3::VARCHAR = '' OR status = $3)
ORDER BY created_at DESC;

-- name: UpdateCapability :one
UPDATE capabilities
SET description = $3,
    provider_type = $4,
    provider_id = $5,
    provider_tool = $6,
    input_schema = $7,
    output_schema = $8,
    status = $9,
    version = $10,
    credential_reference = $11,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING id, tenant_id, name, description, provider_type, provider_id, provider_tool, input_schema, output_schema, status, version, credential_reference, created_at, updated_at;

-- name: DisableCapability :one
UPDATE capabilities
SET status = 'DISABLED', updated_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING id, tenant_id, name, description, provider_type, provider_id, provider_tool, input_schema, output_schema, status, version, credential_reference, created_at, updated_at;

-- name: UpdateCapabilityStatus :one
UPDATE capabilities
SET status = $3, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2
RETURNING id, tenant_id, name, description, provider_type, provider_id, provider_tool, input_schema, output_schema, status, version, credential_reference, created_at, updated_at;

-- name: BindCapability :one
INSERT INTO capability_bindings (tenant_id, capability_id, scope_type, scope_id, permission)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, capability_id, scope_type, scope_id, permission, created_at, updated_at;

-- name: ListBindingsByCapability :many
SELECT id, tenant_id, capability_id, scope_type, scope_id, permission, created_at, updated_at
FROM capability_bindings
WHERE capability_id = $1 AND tenant_id = $2
ORDER BY scope_type ASC, scope_id ASC;

-- name: ListBindingsByScope :many
SELECT id, tenant_id, capability_id, scope_type, scope_id, permission, created_at, updated_at
FROM capability_bindings
WHERE scope_type = $1 AND scope_id = $2 AND tenant_id = $3
ORDER BY capability_id ASC;

-- name: DeleteBinding :one
DELETE FROM capability_bindings
WHERE id = $1 AND tenant_id = $2
RETURNING id, tenant_id, capability_id, scope_type, scope_id, permission, created_at, updated_at;

-- name: UpsertPolicy :one
INSERT INTO policies (tenant_id, scope_type, scope_id, type, content)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (tenant_id, scope_type, scope_id, type)
DO UPDATE SET content = EXCLUDED.content, updated_at = NOW()
RETURNING id, tenant_id, scope_type, scope_id, type, content, created_at, updated_at;

-- name: FindPolicyByType :one
SELECT id, tenant_id, scope_type, scope_id, type, content, created_at, updated_at
FROM policies
WHERE tenant_id = $1 AND scope_type = $2 AND scope_id = $3 AND type = $4
LIMIT 1;

-- name: ListPoliciesByScope :many
SELECT id, tenant_id, scope_type, scope_id, type, content, created_at, updated_at
FROM policies
WHERE tenant_id = $1 AND scope_type = $2 AND scope_id = $3
ORDER BY type ASC;
