-- +goose Up
-- Capability Registry and Policy persistence (PRD §3.11, §3.13, §59-64).
-- Hierarchy: Tenant → Capability → scoped bindings; Policies scoped to tenant/workflow/state.

-- Capability Registry: logical operations referenced by states, mapped to a
-- provider (MCP/INTERNAL/HTTP/FUTURE) at runtime (PRD §59). Secrets are never
-- stored here — only a credential_reference (PRD §61).
CREATE TABLE capabilities (
  id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id          UUID         NOT NULL,
  name               VARCHAR(255) NOT NULL,              -- logical capability, e.g. payment.create
  description        TEXT,
  provider_type      VARCHAR(32)  NOT NULL,              -- MCP/INTERNAL/HTTP/FUTURE (PRD §59)
  provider_id        VARCHAR(255),
  input_schema       JSONB        NOT NULL DEFAULT '{}',
  output_schema      JSONB        NOT NULL DEFAULT '{}',
  status             VARCHAR(32)  NOT NULL DEFAULT 'ACTIVE',  -- ACTIVE/INACTIVE/DISABLED
  version            INT          NOT NULL DEFAULT 1,
  credential_reference VARCHAR(255),                     -- PRD §61, never secrets
  created_at         TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT capabilities_tenant_name_unique UNIQUE (tenant_id, name)
);

-- Scoped capability bindings: scope availability of a capability at
-- tenant/workflow/state level with most-restrictive-wins resolution (PRD §60).
CREATE TABLE capability_bindings (
  id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id     UUID         NOT NULL,
  capability_id UUID         NOT NULL REFERENCES capabilities(id) ON DELETE CASCADE,
  scope_type    VARCHAR(32)  NOT NULL,                   -- TENANT/WORKFLOW/STATE (PRD §60)
  scope_id      VARCHAR(255) NOT NULL,                   -- tenant/workflow/state id
  permission    VARCHAR(16)  NOT NULL DEFAULT 'ALLOW',   -- ALLOW/DENY
  created_at    TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT capability_bindings_tenant_capability_scope_unique UNIQUE (tenant_id, capability_id, scope_type, scope_id)
);

-- Policies: runtime/security/business constraints scoped to a
-- tenant/workflow/state (PRD §3.13, §12). content is a JSONB policy document.
CREATE TABLE policies (
  id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id  UUID         NOT NULL,
  scope_type VARCHAR(32)  NOT NULL,                      -- TENANT/WORKFLOW/STATE
  scope_id   VARCHAR(255) NOT NULL,                      -- tenant/workflow/state id
  type       VARCHAR(64)  NOT NULL,                      -- e.g. timeout/retry/human_handoff/workflow
  content    JSONB        NOT NULL DEFAULT '{}',
  created_at TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT policies_tenant_scope_type_unique UNIQUE (tenant_id, scope_type, scope_id, type)
);

-- Indexes for tenant/capability/scope-scoped lookups (PRD §4, §96).
CREATE INDEX capabilities_tenant_id_idx ON capabilities (tenant_id);
CREATE INDEX capability_bindings_tenant_capability_idx ON capability_bindings (tenant_id, capability_id);
CREATE INDEX capability_bindings_tenant_scope_idx ON capability_bindings (tenant_id, scope_type, scope_id);
CREATE INDEX policies_tenant_scope_idx ON policies (tenant_id, scope_type, scope_id);

-- +goose Down
DROP TABLE IF EXISTS policies;
DROP TABLE IF EXISTS capability_bindings;
DROP TABLE IF EXISTS capabilities;
