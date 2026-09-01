-- +goose Up
-- Provider MCP tool mapping and State MCP execution evidence.

ALTER TABLE capabilities ADD COLUMN IF NOT EXISTS provider_tool VARCHAR(255);

CREATE TABLE IF NOT EXISTS capability_execution_evidence (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id             UUID NOT NULL,
  project_id            UUID NOT NULL,
  workflow_instance_id  UUID NOT NULL,
  state_id              VARCHAR(255) NOT NULL,
  capability_id         UUID NOT NULL,
  capability_name       VARCHAR(255) NOT NULL,
  provider_server        VARCHAR(255) NOT NULL,
  provider_tool         VARCHAR(255) NOT NULL,
  correlation_id        VARCHAR(255),
  idempotency_key        VARCHAR(255) NOT NULL,
  status                VARCHAR(32) NOT NULL,
  result                JSONB NOT NULL DEFAULT '{}',
  error                 JSONB,
  created_at            TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT capability_evidence_instance_key_unique
    UNIQUE (tenant_id, workflow_instance_id, state_id, capability_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS capability_evidence_instance_idx
  ON capability_execution_evidence (tenant_id, project_id, workflow_instance_id, state_id);

-- +goose Down
DROP INDEX IF EXISTS capability_evidence_instance_idx;
DROP TABLE IF EXISTS capability_execution_evidence;
ALTER TABLE capabilities DROP COLUMN IF EXISTS provider_tool;
