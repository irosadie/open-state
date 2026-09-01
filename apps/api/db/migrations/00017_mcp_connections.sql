-- +goose Up
-- Project-owned external MCP connection registry. Secrets are represented only
-- by a protected reference; this table never stores bearer/OAuth values.
CREATE TABLE IF NOT EXISTS mcp_connections (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id             UUID NOT NULL,
  project_id            UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name                  VARCHAR(255) NOT NULL,
  alias                 VARCHAR(128) NOT NULL,
  transport             VARCHAR(32) NOT NULL,
  endpoint              TEXT,
  stdio_profile         VARCHAR(128),
  stdio_args            JSONB NOT NULL DEFAULT '[]'::jsonb,
  auth_type             VARCHAR(32) NOT NULL DEFAULT 'none',
  credential_reference  VARCHAR(255),
  credential_status     VARCHAR(32) NOT NULL DEFAULT 'missing',
  status                VARCHAR(32) NOT NULL DEFAULT 'enabled',
  last_test_status      VARCHAR(32) NOT NULL DEFAULT 'never',
  last_test_error_code  VARCHAR(64),
  last_tested_at        TIMESTAMP(6),
  created_by            VARCHAR(255) NOT NULL,
  updated_by            VARCHAR(255) NOT NULL,
  created_at            TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT mcp_connections_alias_unique UNIQUE (project_id, alias),
  CONSTRAINT mcp_connections_transport_check CHECK (transport IN ('streamable_http', 'sse', 'stdio')),
  CONSTRAINT mcp_connections_auth_type_check CHECK (auth_type IN ('none', 'bearer', 'oauth')),
  CONSTRAINT mcp_connections_credential_status_check CHECK (credential_status IN ('configured', 'missing', 'action_required')),
  CONSTRAINT mcp_connections_status_check CHECK (status IN ('enabled', 'disabled')),
  CONSTRAINT mcp_connections_last_test_status_check CHECK (last_test_status IN ('never', 'ready', 'failed', 'disabled')),
  CONSTRAINT mcp_connections_remote_endpoint_check CHECK (
    (transport = 'stdio' AND endpoint IS NULL) OR
    (transport IN ('streamable_http', 'sse') AND endpoint IS NOT NULL)
  ),
  CONSTRAINT mcp_connections_stdio_profile_check CHECK (
    (transport = 'stdio' AND stdio_profile IS NOT NULL) OR
    (transport IN ('streamable_http', 'sse') AND stdio_profile IS NULL)
  )
);

CREATE INDEX IF NOT EXISTS mcp_connections_project_idx
  ON mcp_connections (tenant_id, project_id, status, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS mcp_connections_project_idx;
DROP TABLE IF EXISTS mcp_connections;
