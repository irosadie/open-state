-- name: CreateMCPConnection :one
INSERT INTO mcp_connections (
  tenant_id, project_id, name, alias, transport, endpoint, stdio_profile,
  stdio_args, auth_type, credential_reference, credential_status, status,
  created_by, updated_by
)
VALUES (
  $1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), $8, $9,
  NULLIF($10, ''), $11, $12, $13, $13
)
RETURNING id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at;

-- name: FindMCPConnectionByID :one
SELECT id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at
FROM mcp_connections
WHERE id = $1 AND tenant_id = $2 AND project_id = $3
LIMIT 1;

-- name: ListMCPConnectionsByProject :many
SELECT id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at
FROM mcp_connections
WHERE tenant_id = $1 AND project_id = $2
ORDER BY created_at DESC, alias ASC;

-- name: UpdateMCPConnection :one
UPDATE mcp_connections
SET name = $4,
    alias = $5,
    transport = $6,
    endpoint = NULLIF($7, ''),
    stdio_profile = NULLIF($8, ''),
    stdio_args = $9,
    auth_type = $10,
    credential_reference = CASE WHEN $11 IS NULL THEN NULL WHEN $11 = '' THEN credential_reference ELSE $11 END,
    credential_status = $12,
    updated_by = $13,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND project_id = $3
RETURNING id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at;

-- name: DeleteMCPConnection :exec
DELETE FROM mcp_connections
WHERE id = $1 AND tenant_id = $2 AND project_id = $3;

-- name: UpdateMCPConnectionStatus :one
UPDATE mcp_connections
SET status = $4,
    last_test_status = CASE WHEN $4 = 'disabled' THEN 'disabled' ELSE last_test_status END,
    updated_by = $5,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND project_id = $3
RETURNING id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at;

-- name: RecordMCPConnectionTest :one
UPDATE mcp_connections
SET last_test_status = $4,
    last_test_error_code = NULLIF($5, ''),
    last_tested_at = NOW(),
    updated_by = $6,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND project_id = $3
RETURNING id, tenant_id, project_id, name, alias, transport, endpoint,
  stdio_profile, stdio_args, auth_type, credential_reference, credential_status,
  status, last_test_status, last_test_error_code, last_tested_at, created_by,
  updated_by, created_at, updated_at;
