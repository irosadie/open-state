-- +goose Up
-- Phase 7: protected OAuth metadata, safe health state, and bounded execution
-- policy. Secret values and OAuth token artifacts remain outside PostgreSQL.
ALTER TABLE mcp_connections
  ADD COLUMN IF NOT EXISTS oauth_authorization_endpoint TEXT,
  ADD COLUMN IF NOT EXISTS oauth_token_endpoint TEXT,
  ADD COLUMN IF NOT EXISTS oauth_client_id VARCHAR(255),
  ADD COLUMN IF NOT EXISTS oauth_client_secret_reference VARCHAR(255),
  ADD COLUMN IF NOT EXISTS oauth_scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS oauth_redirect_uri TEXT,
  ADD COLUMN IF NOT EXISTS oauth_access_token_reference VARCHAR(255),
  ADD COLUMN IF NOT EXISTS oauth_refresh_token_reference VARCHAR(255),
  ADD COLUMN IF NOT EXISTS oauth_expires_at TIMESTAMP(6),
  ADD COLUMN IF NOT EXISTS oauth_status VARCHAR(32) NOT NULL DEFAULT 'disconnected',
  ADD COLUMN IF NOT EXISTS health_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
  ADD COLUMN IF NOT EXISTS health_reason VARCHAR(128),
  ADD COLUMN IF NOT EXISTS last_success_at TIMESTAMP(6),
  ADD COLUMN IF NOT EXISTS consecutive_failures INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS circuit_opened_at TIMESTAMP(6),
  ADD COLUMN IF NOT EXISTS timeout_ms INTEGER NOT NULL DEFAULT 10000,
  ADD COLUMN IF NOT EXISTS max_concurrency INTEGER NOT NULL DEFAULT 4,
  ADD COLUMN IF NOT EXISTS rate_limit_per_second DOUBLE PRECISION NOT NULL DEFAULT 10,
  ADD COLUMN IF NOT EXISTS rate_limit_burst INTEGER NOT NULL DEFAULT 20,
  ADD COLUMN IF NOT EXISTS retry_max INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS circuit_failure_threshold INTEGER NOT NULL DEFAULT 5,
  ADD COLUMN IF NOT EXISTS circuit_recovery_seconds INTEGER NOT NULL DEFAULT 30;

ALTER TABLE mcp_connections
  ADD CONSTRAINT mcp_connections_oauth_status_check CHECK (oauth_status IN ('connected', 'expired', 'disconnected', 'action_required')),
  ADD CONSTRAINT mcp_connections_health_status_check CHECK (health_status IN ('unknown', 'healthy', 'degraded', 'unavailable', 'action_required', 'circuit_open')),
  ADD CONSTRAINT mcp_connections_consecutive_failures_check CHECK (consecutive_failures >= 0),
  ADD CONSTRAINT mcp_connections_timeout_check CHECK (timeout_ms BETWEEN 100 AND 300000),
  ADD CONSTRAINT mcp_connections_concurrency_check CHECK (max_concurrency BETWEEN 1 AND 256),
  ADD CONSTRAINT mcp_connections_rate_check CHECK (rate_limit_per_second > 0 AND rate_limit_per_second <= 10000),
  ADD CONSTRAINT mcp_connections_rate_burst_check CHECK (rate_limit_burst BETWEEN 1 AND 10000),
  ADD CONSTRAINT mcp_connections_retry_check CHECK (retry_max BETWEEN 0 AND 5),
  ADD CONSTRAINT mcp_connections_circuit_threshold_check CHECK (circuit_failure_threshold BETWEEN 1 AND 100),
  ADD CONSTRAINT mcp_connections_circuit_recovery_check CHECK (circuit_recovery_seconds BETWEEN 1 AND 86400);

CREATE INDEX IF NOT EXISTS mcp_connections_health_idx
  ON mcp_connections (tenant_id, project_id, health_status, updated_at DESC);

-- +goose Down
DROP INDEX IF EXISTS mcp_connections_health_idx;
ALTER TABLE mcp_connections
  DROP CONSTRAINT IF EXISTS mcp_connections_oauth_status_check,
  DROP CONSTRAINT IF EXISTS mcp_connections_health_status_check,
  DROP CONSTRAINT IF EXISTS mcp_connections_consecutive_failures_check,
  DROP CONSTRAINT IF EXISTS mcp_connections_timeout_check,
  DROP CONSTRAINT IF EXISTS mcp_connections_concurrency_check,
  DROP CONSTRAINT IF EXISTS mcp_connections_rate_check,
  DROP CONSTRAINT IF EXISTS mcp_connections_rate_burst_check,
  DROP CONSTRAINT IF EXISTS mcp_connections_retry_check,
  DROP CONSTRAINT IF EXISTS mcp_connections_circuit_threshold_check,
  DROP CONSTRAINT IF EXISTS mcp_connections_circuit_recovery_check;
ALTER TABLE mcp_connections
  DROP COLUMN IF EXISTS oauth_authorization_endpoint,
  DROP COLUMN IF EXISTS oauth_token_endpoint,
  DROP COLUMN IF EXISTS oauth_client_id,
  DROP COLUMN IF EXISTS oauth_client_secret_reference,
  DROP COLUMN IF EXISTS oauth_scopes,
  DROP COLUMN IF EXISTS oauth_redirect_uri,
  DROP COLUMN IF EXISTS oauth_access_token_reference,
  DROP COLUMN IF EXISTS oauth_refresh_token_reference,
  DROP COLUMN IF EXISTS oauth_expires_at,
  DROP COLUMN IF EXISTS oauth_status,
  DROP COLUMN IF EXISTS health_status,
  DROP COLUMN IF EXISTS health_reason,
  DROP COLUMN IF EXISTS last_success_at,
  DROP COLUMN IF EXISTS consecutive_failures,
  DROP COLUMN IF EXISTS circuit_opened_at,
  DROP COLUMN IF EXISTS timeout_ms,
  DROP COLUMN IF EXISTS max_concurrency,
  DROP COLUMN IF EXISTS rate_limit_per_second,
  DROP COLUMN IF EXISTS rate_limit_burst,
  DROP COLUMN IF EXISTS retry_max,
  DROP COLUMN IF EXISTS circuit_failure_threshold,
  DROP COLUMN IF EXISTS circuit_recovery_seconds;
