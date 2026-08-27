-- +goose Up
-- Workflow definitions (PRD §3.2, §5, §9) — tenant+project-isolated definition roots.
-- Hierarchy: Tenant → Project → Intent → Workflow → State.

-- Projects are business areas owned by a tenant (e.g. resto, padel, dokter).
CREATE TABLE projects (
  id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  UUID         NOT NULL,
  name       VARCHAR(255) NOT NULL,
  slug       VARCHAR(255) NOT NULL,
  status     VARCHAR(32)  NOT NULL DEFAULT 'ACTIVE',   -- ACTIVE/ARCHIVED
  created_at TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT projects_tenant_slug_unique UNIQUE (tenant_id, slug)
);

CREATE TABLE workflows (
  id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id        UUID         NOT NULL,
  project_id       UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  slug             VARCHAR(255) NOT NULL,
  name             VARCHAR(255) NOT NULL,
  description      TEXT,
  status           VARCHAR(32)  NOT NULL DEFAULT 'DRAFT',   -- DRAFT/VALIDATING/VALID/PUBLISHED/ARCHIVED
  current_version  INT          NOT NULL DEFAULT 0,
  version          INT          NOT NULL DEFAULT 0,          -- optimistic lock (PRD §31)
  created_at       TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT workflows_tenant_project_slug_unique UNIQUE (tenant_id, project_id, slug)
);

-- Immutable published snapshots of a workflow definition (PRD §3.3, §9, §55, §58).
CREATE TABLE workflow_versions (
  id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  workflow_id   UUID         NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
  tenant_id     UUID         NOT NULL,                        -- denormalized for scoped reads
  project_id    UUID         NOT NULL,                        -- denormalized for scoped reads
  version_no    INT          NOT NULL,
  definition    JSONB        NOT NULL,                        -- full WorkflowDefinition envelope
  status        VARCHAR(32)  NOT NULL DEFAULT 'DRAFT',
  is_current    BOOLEAN      NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT workflow_versions_workflow_version_unique UNIQUE (workflow_id, version_no)
);

-- Relational, queryable projection of workflow definition nodes (PRD §12, §14).
CREATE TABLE states (
  id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  workflow_version_id UUID        NOT NULL REFERENCES workflow_versions(id) ON DELETE CASCADE,
  key                VARCHAR(255) NOT NULL,                   -- stable node key, e.g. PAYMENT
  kind               VARCHAR(32)  NOT NULL,                   -- START/STATE/DECISION/WAIT/END/EVENT
  name               VARCHAR(255) NOT NULL,
  description        TEXT,
  instructions       TEXT,
  required_context   JSONB        NOT NULL DEFAULT '[]',
  capabilities       JSONB        NOT NULL DEFAULT '[]',
  policy             JSONB        NOT NULL DEFAULT '{}',
  is_terminal        BOOLEAN      NOT NULL DEFAULT FALSE,
  position           JSONB        NOT NULL DEFAULT '{}',      -- x/y for builder
  created_at         TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT states_workflow_version_key_unique UNIQUE (workflow_version_id, key)
);

-- Relational transitions (PRD §33, §34).
CREATE TABLE transitions (
  id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  workflow_version_id UUID        NOT NULL REFERENCES workflow_versions(id) ON DELETE CASCADE,
  key                VARCHAR(255) NOT NULL,                   -- stable transition id
  source_state_id    UUID         NOT NULL REFERENCES states(id) ON DELETE CASCADE,
  target_state_id    UUID         NOT NULL REFERENCES states(id) ON DELETE CASCADE,
  event              VARCHAR(255) NOT NULL,
  priority           INT          NOT NULL DEFAULT 1,         -- lower = evaluated first (PRD §34)
  is_active          BOOLEAN      NOT NULL DEFAULT TRUE,
  created_at         TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT transitions_workflow_version_key_unique UNIQUE (workflow_version_id, key)
);

-- Relational guards per transition (PRD §35).
CREATE TABLE transition_guards (
  id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  transition_id      UUID         NOT NULL REFERENCES transitions(id) ON DELETE CASCADE,
  workflow_version_id UUID        NOT NULL REFERENCES workflow_versions(id) ON DELETE CASCADE,
  logic              VARCHAR(8)   NOT NULL DEFAULT 'AND',     -- AND/OR
  conditions         JSONB        NOT NULL DEFAULT '[]',
  created_at         TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMP(6) NOT NULL DEFAULT NOW()
);

-- Indexes for FK and tenant/project-scoped lookups (PRD §4, §96).
CREATE INDEX projects_tenant_id_idx          ON projects (tenant_id);
CREATE INDEX workflows_tenant_id_idx         ON workflows (tenant_id);
CREATE INDEX workflows_project_id_idx        ON workflows (project_id);
CREATE INDEX workflow_versions_workflow_id_idx ON workflow_versions (workflow_id);
CREATE INDEX workflow_versions_tenant_id_idx ON workflow_versions (tenant_id);
CREATE INDEX workflow_versions_project_id_idx ON workflow_versions (project_id);
CREATE INDEX states_workflow_version_id_idx   ON states (workflow_version_id);
CREATE INDEX transitions_workflow_version_id_idx ON transitions (workflow_version_id);
CREATE INDEX transitions_source_state_id_idx  ON transitions (source_state_id);
CREATE INDEX transitions_target_state_id_idx  ON transitions (target_state_id);
CREATE INDEX transition_guards_transition_id_idx     ON transition_guards (transition_id);
CREATE INDEX transition_guards_workflow_version_id_idx ON transition_guards (workflow_version_id);

-- +goose Down
DROP TABLE IF EXISTS transition_guards;
DROP TABLE IF EXISTS transitions;
DROP TABLE IF EXISTS states;
DROP TABLE IF EXISTS workflow_versions;
DROP TABLE IF EXISTS workflows;
DROP TABLE IF EXISTS projects;
