-- +goose Up
-- Runtime instances (PRD §3.6, §10, §11, §25, §31, §48, §58) — persisted execution
-- copies of published workflow versions, tenant-isolated. Builds on the workflow
-- definition schema (00002_workflows.sql): workflows / workflow_versions / states.

-- Runtime execution root: a running copy of a workflow version (PRD §10, §58).
CREATE TABLE workflow_instances (
  id                      UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id               UUID         NOT NULL,
  workflow_id             UUID         NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
  workflow_version_id     UUID         NOT NULL REFERENCES workflow_versions(id) ON DELETE CASCADE,
  correlation_key         VARCHAR(255),                              -- business/conversation correlation (PRD §6)
  status                  VARCHAR(32)  NOT NULL DEFAULT 'CREATED',   -- CREATED/RUNNING/WAITING/COMPLETED/CANCELLED/FAILED/EXPIRED/ABORTED/SUSPENDED (PRD §10, §42-43)
  version                 INT          NOT NULL DEFAULT 0,           -- optimistic lock (PRD §31)
  current_state_instance_id UUID,                                    -- circular FK to state_instances (added below, nullable)
  started_at              TIMESTAMP(6),
  completed_at            TIMESTAMP(6),
  expires_at              TIMESTAMP(6),                              -- workflow timeout (PRD §26)
  created_at              TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMP(6) NOT NULL DEFAULT NOW()
);

-- Runtime occurrence of a state inside an instance (PRD §11, §48).
CREATE TABLE state_instances (
  id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id            UUID         NOT NULL,
  workflow_instance_id UUID         NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  workflow_version_id  UUID         NOT NULL REFERENCES workflow_versions(id) ON DELETE CASCADE,
  state_key            VARCHAR(255) NOT NULL,                        -- references states.key of the pinned version
  state_id             UUID         REFERENCES states(id) ON DELETE SET NULL,
  status               VARCHAR(32)  NOT NULL DEFAULT 'ENTERING',     -- ENTERING/ACTIVE/WAITING/EXITING/COMPLETED/FAILED/EXPIRED/CANCELLED (PRD §11)
  version              INT          NOT NULL DEFAULT 0,               -- optimistic lock (PRD §31)
  retry_count          INT          NOT NULL DEFAULT 0,               -- retry counter (PRD §48)
  entered_at           TIMESTAMP(6) NOT NULL DEFAULT NOW(),          -- timeout scheduling (PRD §3.6, §25)
  expires_at           TIMESTAMP(6),
  exited_at            TIMESTAMP(6),
  created_at           TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMP(6) NOT NULL DEFAULT NOW()
);

-- Resolve the circular FK workflow_instances → state_instances now that both exist.
-- current_state_instance_id is nullable so there is no INSERT-time ordering problem (PRD §7).
ALTER TABLE workflow_instances
  ADD CONSTRAINT workflow_instances_current_state_instance_id_fkey
  FOREIGN KEY (current_state_instance_id) REFERENCES state_instances(id);

-- Indexes for FK and tenant-scoped lookups (PRD §4, §96).
CREATE INDEX workflow_instances_tenant_id_status_idx ON workflow_instances (tenant_id, status);
CREATE INDEX workflow_instances_tenant_id_correlation_key_idx ON workflow_instances (tenant_id, correlation_key);
CREATE INDEX workflow_instances_workflow_id_idx ON workflow_instances (workflow_id);
CREATE INDEX workflow_instances_workflow_version_id_idx ON workflow_instances (workflow_version_id);
CREATE INDEX state_instances_workflow_instance_id_idx ON state_instances (workflow_instance_id);
CREATE INDEX state_instances_tenant_id_status_idx ON state_instances (tenant_id, status);
CREATE INDEX state_instances_workflow_version_id_idx ON state_instances (workflow_version_id);

-- +goose Down
DROP INDEX IF EXISTS state_instances_workflow_version_id_idx;
DROP INDEX IF EXISTS state_instances_tenant_id_status_idx;
DROP INDEX IF EXISTS state_instances_workflow_instance_id_idx;
DROP INDEX IF EXISTS workflow_instances_workflow_version_id_idx;
DROP INDEX IF EXISTS workflow_instances_workflow_id_idx;
DROP INDEX IF EXISTS workflow_instances_tenant_id_correlation_key_idx;
DROP INDEX IF EXISTS workflow_instances_tenant_id_status_idx;
ALTER TABLE workflow_instances DROP CONSTRAINT IF EXISTS workflow_instances_current_state_instance_id_fkey;
DROP TABLE IF EXISTS state_instances;
DROP TABLE IF EXISTS workflow_instances;
