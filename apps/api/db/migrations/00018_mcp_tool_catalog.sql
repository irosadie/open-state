-- +goose Up
-- Verified, project-scoped snapshots returned by an external MCP connection.
CREATE TABLE IF NOT EXISTS mcp_discovery_runs (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id            UUID NOT NULL,
  project_id           UUID NOT NULL,
  connection_id        UUID NOT NULL REFERENCES mcp_connections(id) ON DELETE CASCADE,
  status               VARCHAR(32) NOT NULL,
  tool_count           INTEGER NOT NULL DEFAULT 0,
  catalog_fingerprint  VARCHAR(64),
  error_code           VARCHAR(64),
  started_at           TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  completed_at         TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  created_by           VARCHAR(255) NOT NULL,
  CONSTRAINT mcp_discovery_runs_status_check CHECK (status IN ('succeeded', 'failed')),
  CONSTRAINT mcp_discovery_runs_tool_count_check CHECK (tool_count >= 0)
);

CREATE INDEX IF NOT EXISTS mcp_discovery_runs_connection_idx
  ON mcp_discovery_runs (tenant_id, project_id, connection_id, completed_at DESC);

CREATE TABLE IF NOT EXISTS mcp_discovered_tools (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id         UUID NOT NULL,
  project_id        UUID NOT NULL,
  connection_id     UUID NOT NULL REFERENCES mcp_connections(id) ON DELETE CASCADE,
  tool_name         VARCHAR(255) NOT NULL,
  title             VARCHAR(255),
  description       TEXT NOT NULL DEFAULT '',
  input_schema      JSONB NOT NULL,
  annotations       JSONB NOT NULL DEFAULT '{}'::jsonb,
  fingerprint       VARCHAR(64) NOT NULL,
  enabled           BOOLEAN NOT NULL DEFAULT TRUE,
  availability      VARCHAR(32) NOT NULL DEFAULT 'present',
  drift_status      VARCHAR(32) NOT NULL DEFAULT 'new',
  first_seen_at     TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  last_seen_at      TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  removed_at        TIMESTAMP(6),
  discovery_run_id  UUID REFERENCES mcp_discovery_runs(id) ON DELETE SET NULL,
  created_at        TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT mcp_discovered_tools_connection_name_unique UNIQUE (connection_id, tool_name),
  CONSTRAINT mcp_discovered_tools_availability_check CHECK (availability IN ('present', 'removed')),
  CONSTRAINT mcp_discovered_tools_drift_status_check CHECK (drift_status IN ('new', 'unchanged', 'changed', 'removed'))
);

CREATE INDEX IF NOT EXISTS mcp_discovered_tools_project_idx
  ON mcp_discovered_tools (tenant_id, project_id, connection_id, availability, tool_name);

-- +goose Down
DROP INDEX IF EXISTS mcp_discovered_tools_project_idx;
DROP TABLE IF EXISTS mcp_discovered_tools;
DROP INDEX IF EXISTS mcp_discovery_runs_connection_idx;
DROP TABLE IF EXISTS mcp_discovery_runs;
