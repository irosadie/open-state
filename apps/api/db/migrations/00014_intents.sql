-- +goose Up
-- Canonical, tenant/project-scoped intent catalog for LLM routing.
CREATE TABLE intents (
  id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id    UUID         NOT NULL,
  project_id   UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  workflow_id  UUID         NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
  intent_key   VARCHAR(128) NOT NULL,
  name         VARCHAR(255) NOT NULL,
  description  TEXT         NOT NULL DEFAULT '',
  examples     JSONB        NOT NULL DEFAULT '[]',
  created_at   TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT intents_tenant_project_key_unique UNIQUE (tenant_id, project_id, intent_key)
);

CREATE INDEX intents_tenant_project_idx ON intents (tenant_id, project_id);
CREATE INDEX intents_workflow_id_idx ON intents (workflow_id);

-- +goose Down
DROP TABLE IF EXISTS intents;
