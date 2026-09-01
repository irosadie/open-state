-- name: ListEligibleMCPToolOptions :many
SELECT
  conn.id AS connection_id,
  conn.name AS connection_name,
  conn.alias AS connection_alias,
  conn.status AS connection_status,
  tool.id AS tool_id,
  tool.tool_name AS tool_name,
  tool.title AS tool_title,
  tool.description AS tool_description,
  tool.input_schema AS tool_input_schema,
  tool.fingerprint AS tool_fingerprint
FROM mcp_connections conn
JOIN mcp_discovered_tools tool
  ON tool.tenant_id = conn.tenant_id
 AND tool.project_id = conn.project_id
 AND tool.connection_id = conn.id
WHERE conn.tenant_id = $1
  AND conn.project_id = $2
  AND conn.status = 'enabled'
  AND tool.enabled = TRUE
  AND tool.availability = 'present'
ORDER BY conn.alias ASC, tool.tool_name ASC;

-- name: ListProjectCapabilityMCPBindings :many
SELECT
  binding.id,
  binding.tenant_id,
  binding.project_id,
  binding.capability_id,
  cap.name AS capability_name,
  cap.description AS capability_description,
  binding.mcp_connection_id AS connection_id,
  conn.name AS connection_name,
  conn.alias AS connection_alias,
  conn.status AS connection_status,
  binding.mcp_discovered_tool_id AS tool_id,
  tool.tool_name,
  tool.title AS tool_title,
  tool.description AS tool_description,
  binding.bound_tool_fingerprint,
  tool.fingerprint AS current_tool_fingerprint,
  tool.enabled AS tool_enabled,
  tool.availability AS tool_availability,
  tool.drift_status AS tool_drift_status,
  CASE
    WHEN conn.status <> 'enabled' THEN 'CONNECTION_DISABLED'
    WHEN tool.enabled = FALSE THEN 'TOOL_DISABLED'
    WHEN tool.availability <> 'present' THEN 'TOOL_REMOVED'
    WHEN binding.bound_tool_fingerprint <> tool.fingerprint THEN 'STALE'
    ELSE 'ACTIVE'
  END AS health,
  CASE
    WHEN conn.status <> 'enabled' THEN 'Provider connection is disabled'
    WHEN tool.enabled = FALSE THEN 'Provider tool is disabled'
    WHEN tool.availability <> 'present' THEN 'Provider tool was removed from the latest catalog'
    WHEN binding.bound_tool_fingerprint <> tool.fingerprint THEN 'Provider tool catalog changed; rebind required'
    ELSE ''
  END AS health_reason,
  binding.created_at,
  binding.updated_at
FROM project_capability_mcp_bindings binding
JOIN capabilities cap
  ON cap.tenant_id = binding.tenant_id
 AND cap.id = binding.capability_id
JOIN mcp_connections conn
  ON conn.tenant_id = binding.tenant_id
 AND conn.project_id = binding.project_id
 AND conn.id = binding.mcp_connection_id
JOIN mcp_discovered_tools tool
  ON tool.tenant_id = binding.tenant_id
 AND tool.project_id = binding.project_id
 AND tool.connection_id = binding.mcp_connection_id
 AND tool.id = binding.mcp_discovered_tool_id
WHERE binding.tenant_id = $1
  AND binding.project_id = $2
ORDER BY cap.name ASC;

-- name: FindProjectCapabilityMCPBinding :one
SELECT
  binding.id,
  binding.tenant_id,
  binding.project_id,
  binding.capability_id,
  cap.name AS capability_name,
  cap.description AS capability_description,
  binding.mcp_connection_id AS connection_id,
  conn.name AS connection_name,
  conn.alias AS connection_alias,
  conn.status AS connection_status,
  binding.mcp_discovered_tool_id AS tool_id,
  tool.tool_name,
  tool.title AS tool_title,
  tool.description AS tool_description,
  binding.bound_tool_fingerprint,
  tool.fingerprint AS current_tool_fingerprint,
  tool.enabled AS tool_enabled,
  tool.availability AS tool_availability,
  tool.drift_status AS tool_drift_status,
  CASE
    WHEN conn.status <> 'enabled' THEN 'CONNECTION_DISABLED'
    WHEN tool.enabled = FALSE THEN 'TOOL_DISABLED'
    WHEN tool.availability <> 'present' THEN 'TOOL_REMOVED'
    WHEN binding.bound_tool_fingerprint <> tool.fingerprint THEN 'STALE'
    ELSE 'ACTIVE'
  END AS health,
  CASE
    WHEN conn.status <> 'enabled' THEN 'Provider connection is disabled'
    WHEN tool.enabled = FALSE THEN 'Provider tool is disabled'
    WHEN tool.availability <> 'present' THEN 'Provider tool was removed from the latest catalog'
    WHEN binding.bound_tool_fingerprint <> tool.fingerprint THEN 'Provider tool catalog changed; rebind required'
    ELSE ''
  END AS health_reason,
  binding.created_at,
  binding.updated_at
FROM project_capability_mcp_bindings binding
JOIN capabilities cap
  ON cap.tenant_id = binding.tenant_id
 AND cap.id = binding.capability_id
JOIN mcp_connections conn
  ON conn.tenant_id = binding.tenant_id
 AND conn.project_id = binding.project_id
 AND conn.id = binding.mcp_connection_id
JOIN mcp_discovered_tools tool
  ON tool.tenant_id = binding.tenant_id
 AND tool.project_id = binding.project_id
 AND tool.connection_id = binding.mcp_connection_id
 AND tool.id = binding.mcp_discovered_tool_id
WHERE binding.tenant_id = $1
  AND binding.project_id = $2
  AND binding.capability_id = $3
LIMIT 1;

-- name: UpsertProjectCapabilityMCPBinding :exec
INSERT INTO project_capability_mcp_bindings (
  tenant_id, project_id, capability_id, mcp_connection_id,
  mcp_discovered_tool_id, bound_tool_fingerprint
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (project_id, capability_id) DO UPDATE
SET mcp_connection_id = EXCLUDED.mcp_connection_id,
    mcp_discovered_tool_id = EXCLUDED.mcp_discovered_tool_id,
    bound_tool_fingerprint = EXCLUDED.bound_tool_fingerprint,
    updated_at = NOW();

-- name: DeleteProjectCapabilityMCPBinding :exec
DELETE FROM project_capability_mcp_bindings
WHERE tenant_id = $1
  AND project_id = $2
  AND capability_id = $3;
