-- +goose Up
-- Short-lived OAuth state/PKCE transactions. Only hashes and secret references
-- are persisted; OAuth codes and token artifacts never enter this table.
CREATE TABLE IF NOT EXISTS mcp_oauth_transactions (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id           UUID NOT NULL,
  project_id          UUID NOT NULL,
  connection_id       UUID NOT NULL REFERENCES mcp_connections(id) ON DELETE CASCADE,
  state_hash          BYTEA NOT NULL UNIQUE,
  verifier_reference  VARCHAR(255) NOT NULL,
  redirect_uri        TEXT NOT NULL,
  expires_at          TIMESTAMP(6) NOT NULL,
  status              VARCHAR(32) NOT NULL DEFAULT 'pending',
  created_at          TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT mcp_oauth_transactions_status_check CHECK (status IN ('pending', 'consumed', 'expired'))
);

CREATE INDEX IF NOT EXISTS mcp_oauth_transactions_lookup_idx
  ON mcp_oauth_transactions (tenant_id, project_id, connection_id, status, expires_at);

-- +goose Down
DROP INDEX IF EXISTS mcp_oauth_transactions_lookup_idx;
DROP TABLE IF EXISTS mcp_oauth_transactions;
