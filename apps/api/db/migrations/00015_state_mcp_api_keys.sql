-- +goose Up
-- Machine credentials for the State MCP endpoint. The raw key is never stored;
-- only a keyed verifier and a safe identifying prefix are persisted.
CREATE TABLE auth_api_keys (
  id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id          UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  name               VARCHAR(255) NOT NULL,
  key_prefix         VARCHAR(32)  NOT NULL UNIQUE,
  key_verifier       BYTEA        NOT NULL,
  default_project_id UUID         REFERENCES projects(id) ON DELETE RESTRICT,
  expires_at         TIMESTAMP(6),
  revoked_at         TIMESTAMP(6),
  last_used_at       TIMESTAMP(6),
  created_by         VARCHAR(255) NOT NULL,
  created_at         TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMP(6) NOT NULL DEFAULT NOW()
);

CREATE TABLE auth_api_key_projects (
  api_key_id UUID        NOT NULL REFERENCES auth_api_keys(id) ON DELETE CASCADE,
  project_id UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  created_at TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  PRIMARY KEY (api_key_id, project_id)
);

CREATE TABLE auth_api_key_scopes (
  api_key_id UUID         NOT NULL REFERENCES auth_api_keys(id) ON DELETE CASCADE,
  scope      VARCHAR(100) NOT NULL,
  created_at TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  PRIMARY KEY (api_key_id, scope)
);

CREATE INDEX auth_api_keys_tenant_id_idx ON auth_api_keys (tenant_id, created_at DESC);
CREATE INDEX auth_api_keys_active_prefix_idx ON auth_api_keys (key_prefix) WHERE revoked_at IS NULL;
CREATE INDEX auth_api_key_projects_project_id_idx ON auth_api_key_projects (project_id);

-- +goose Down
DROP TABLE IF EXISTS auth_api_key_scopes;
DROP TABLE IF EXISTS auth_api_key_projects;
DROP TABLE IF EXISTS auth_api_keys;
