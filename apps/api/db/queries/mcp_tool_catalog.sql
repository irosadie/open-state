-- name: ListMCPDiscoveredTools :many
SELECT id, tenant_id, project_id, connection_id, tool_name, title, description,
  input_schema, annotations, fingerprint, enabled, availability, drift_status,
  first_seen_at, last_seen_at, removed_at, discovery_run_id, created_at, updated_at
FROM mcp_discovered_tools
WHERE tenant_id = $1 AND project_id = $2 AND connection_id = $3
ORDER BY availability ASC, tool_name ASC;

-- name: GetLatestMCPDiscoveryRun :one
SELECT id, tenant_id, project_id, connection_id, status, tool_count,
  catalog_fingerprint, error_code, started_at, completed_at, created_by
FROM mcp_discovery_runs
WHERE tenant_id = $1 AND project_id = $2 AND connection_id = $3
ORDER BY completed_at DESC
LIMIT 1;

-- name: GetLastSuccessfulMCPDiscoveryRun :one
SELECT id, tenant_id, project_id, connection_id, status, tool_count,
  catalog_fingerprint, error_code, started_at, completed_at, created_by
FROM mcp_discovery_runs
WHERE tenant_id = $1 AND project_id = $2 AND connection_id = $3
  AND status = 'succeeded'
ORDER BY completed_at DESC
LIMIT 1;

-- name: CreateMCPDiscoveryRun :one
INSERT INTO mcp_discovery_runs (
  tenant_id, project_id, connection_id, status, tool_count,
  catalog_fingerprint, error_code, started_at, completed_at, created_by
)
VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), $8, $9, $10)
RETURNING id, tenant_id, project_id, connection_id, status, tool_count,
  catalog_fingerprint, error_code, started_at, completed_at, created_by;

-- name: MarkMCPDiscoveredToolsRemoved :exec
UPDATE mcp_discovered_tools
SET availability = 'removed',
    drift_status = 'removed',
    removed_at = COALESCE(removed_at, NOW()),
    updated_at = NOW()
WHERE tenant_id = $1 AND project_id = $2 AND connection_id = $3
  AND availability <> 'removed';

-- name: UpsertMCPDiscoveredTool :one
INSERT INTO mcp_discovered_tools (
  tenant_id, project_id, connection_id, tool_name, title, description,
  input_schema, annotations, fingerprint, availability, drift_status,
  last_seen_at, removed_at, discovery_run_id
)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9, 'present', 'new', NOW(), NULL, $10)
ON CONFLICT (connection_id, tool_name) DO UPDATE
SET title = EXCLUDED.title,
    description = EXCLUDED.description,
    input_schema = EXCLUDED.input_schema,
    annotations = EXCLUDED.annotations,
    drift_status = CASE
      WHEN mcp_discovered_tools.fingerprint = EXCLUDED.fingerprint THEN 'unchanged'
      ELSE 'changed'
    END,
    fingerprint = EXCLUDED.fingerprint,
    availability = 'present',
    last_seen_at = NOW(),
    removed_at = NULL,
    discovery_run_id = EXCLUDED.discovery_run_id,
    updated_at = NOW()
RETURNING id, tenant_id, project_id, connection_id, tool_name, title, description,
  input_schema, annotations, fingerprint, enabled, availability, drift_status,
  first_seen_at, last_seen_at, removed_at, discovery_run_id, created_at, updated_at;

-- name: SetMCPDiscoveredToolEnabled :one
UPDATE mcp_discovered_tools
SET enabled = $4,
    updated_at = NOW()
WHERE tenant_id = $1 AND project_id = $2 AND connection_id = $3 AND tool_name = $5
RETURNING id, tenant_id, project_id, connection_id, tool_name, title, description,
  input_schema, annotations, fingerprint, enabled, availability, drift_status,
  first_seen_at, last_seen_at, removed_at, discovery_run_id, created_at, updated_at;
