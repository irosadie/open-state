-- +goose Up
-- Scoped runtime context and persistent memory (PRD §3.4, §23-24, §43.2, §131).
-- Distinct concerns: context_records is scoped, versioned runtime context; memory_references
-- is long-lived user/customer memory that MUST survive workflow expiry (PRD §24).

-- Runtime context keyed to a scope (tenant/conversation/workflow instance/state instance).
-- Values are JSONB (PRD §131) with an optimistic-lock version for change tracking (PRD §31).
CREATE TABLE context_records (
  id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  UUID         NOT NULL,
  scope_type VARCHAR(32)  NOT NULL,   -- TENANT/CONVERSATION/WORKFLOW_INSTANCE/STATE_INSTANCE (PRD §23)
  scope_id   VARCHAR(255) NOT NULL,   -- tenant/conversation/instance/state-instance id
  key        VARCHAR(255) NOT NULL,   -- e.g. booking.time_start
  value      JSONB        NOT NULL,   -- typed value / snapshot (PRD §131)
  version    INT          NOT NULL DEFAULT 0,  -- optimistic lock (PRD §31)
  created_at TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT context_records_tenant_scope_key_unique UNIQUE (tenant_id, scope_type, scope_id, key)
);

-- Persistent memory references (PRD §24, §43.2). Distinct from workflow data: deleting a
-- workflow instance must NEVER cascade-delete user memory (PRD §24). source_workflow_instance_id
-- is therefore a plain UUID provenance reference — NOT a hard FK — so memory survives instance
-- deletion. Do not turn it into a hard FK.
CREATE TABLE memory_references (
  id                          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id                   UUID         NOT NULL,
  owner_type                  VARCHAR(32)  NOT NULL,   -- e.g. CUSTOMER/USER
  owner_id                    VARCHAR(255) NOT NULL,
  name                        VARCHAR(255) NOT NULL,   -- e.g. address/preferences
  value                       JSONB        NOT NULL,
  source_workflow_instance_id UUID,                    -- optional provenance, soft reference (PRD §24)
  created_at                  TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at                  TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT memory_references_tenant_owner_name_unique UNIQUE (tenant_id, owner_type, owner_id, name)
);

-- Indexes for tenant/scope-scoped lookups (PRD §4, §96).
CREATE INDEX context_records_tenant_scope_idx  ON context_records (tenant_id, scope_type, scope_id);
CREATE INDEX memory_references_tenant_owner_idx ON memory_references (tenant_id, owner_type, owner_id);

-- +goose Down
DROP INDEX IF EXISTS memory_references_tenant_owner_idx;
DROP INDEX IF EXISTS context_records_tenant_scope_idx;
DROP TABLE IF EXISTS memory_references;
DROP TABLE IF EXISTS context_records;
