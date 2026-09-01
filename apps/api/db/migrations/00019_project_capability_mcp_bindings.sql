-- +goose Up
-- Project-scoped, explicit bindings from a logical MCP capability to one
-- enabled external connection and one enabled discovered tool.

-- Composite unique indexes let the binding table enforce tenant/project
-- ownership at the database boundary, even though the legacy tables use
-- single-column UUID primary keys.
CREATE UNIQUE INDEX IF NOT EXISTS capabilities_tenant_id_id_unique
  ON capabilities (tenant_id, id);

CREATE UNIQUE INDEX IF NOT EXISTS mcp_connections_tenant_project_id_unique
  ON mcp_connections (tenant_id, project_id, id);

CREATE UNIQUE INDEX IF NOT EXISTS mcp_discovered_tools_tenant_project_connection_id_unique
  ON mcp_discovered_tools (tenant_id, project_id, connection_id, id);

CREATE TABLE IF NOT EXISTS project_capability_mcp_bindings (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id             UUID NOT NULL,
  project_id            UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  capability_id         UUID NOT NULL,
  mcp_connection_id     UUID NOT NULL,
  mcp_discovered_tool_id UUID NOT NULL,
  bound_tool_fingerprint VARCHAR(64) NOT NULL,
  created_at            TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT project_capability_mcp_bindings_project_capability_unique
    UNIQUE (project_id, capability_id),
  CONSTRAINT project_capability_mcp_bindings_capability_fk
    FOREIGN KEY (tenant_id, capability_id)
    REFERENCES capabilities (tenant_id, id) ON DELETE CASCADE,
  CONSTRAINT project_capability_mcp_bindings_connection_fk
    FOREIGN KEY (tenant_id, project_id, mcp_connection_id)
    REFERENCES mcp_connections (tenant_id, project_id, id) ON DELETE CASCADE,
  CONSTRAINT project_capability_mcp_bindings_tool_fk
    FOREIGN KEY (tenant_id, project_id, mcp_connection_id, mcp_discovered_tool_id)
    REFERENCES mcp_discovered_tools (tenant_id, project_id, connection_id, id)
    ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS project_capability_mcp_bindings_tenant_project_idx
  ON project_capability_mcp_bindings (tenant_id, project_id, capability_id);

-- Migrate only deterministic legacy MCP metadata. A legacy alias/tool pair is
-- backfilled when exactly one project has an enabled connection and one
-- present+enabled discovered tool matching it. Ambiguous or stale metadata
-- intentionally remains an explicit MISSING_MAPPING case for the UI/runtime.
WITH exact_matches AS (
  SELECT cap.tenant_id,
         cap.id AS capability_id,
         conn.project_id,
         conn.id AS mcp_connection_id,
         tool.id AS mcp_discovered_tool_id,
         tool.fingerprint AS bound_tool_fingerprint
  FROM capabilities cap
  JOIN mcp_connections conn
    ON conn.tenant_id = cap.tenant_id
   AND conn.alias = cap.provider_id
   AND conn.status = 'enabled'
  JOIN mcp_discovered_tools tool
    ON tool.tenant_id = conn.tenant_id
   AND tool.project_id = conn.project_id
   AND tool.connection_id = conn.id
   AND tool.tool_name = cap.provider_tool
   AND tool.enabled = TRUE
   AND tool.availability = 'present'
  WHERE cap.provider_type = 'MCP'
    AND cap.status = 'ACTIVE'
    AND NULLIF(cap.provider_id, '') IS NOT NULL
    AND NULLIF(cap.provider_tool, '') IS NOT NULL
), unambiguous AS (
  SELECT tenant_id, capability_id
  FROM exact_matches
  GROUP BY tenant_id, capability_id
  HAVING COUNT(*) = 1
)
INSERT INTO project_capability_mcp_bindings (
  tenant_id, project_id, capability_id, mcp_connection_id,
  mcp_discovered_tool_id, bound_tool_fingerprint
)
SELECT match.tenant_id,
       match.project_id,
       match.capability_id,
       match.mcp_connection_id,
       match.mcp_discovered_tool_id,
       match.bound_tool_fingerprint
FROM exact_matches match
JOIN unambiguous
  ON unambiguous.tenant_id = match.tenant_id
 AND unambiguous.capability_id = match.capability_id
ON CONFLICT (project_id, capability_id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS project_capability_mcp_bindings;
DROP INDEX IF EXISTS mcp_discovered_tools_tenant_project_connection_id_unique;
DROP INDEX IF EXISTS mcp_connections_tenant_project_id_unique;
DROP INDEX IF EXISTS capabilities_tenant_id_id_unique;
