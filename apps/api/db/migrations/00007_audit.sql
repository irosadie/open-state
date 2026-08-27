-- +goose Up
-- Append-only audit trail (PRD 50). Immutable record of important operations,
-- tenant-isolated. Rows are never updated or deleted during normal operation;
-- they are only appended to preserve the audit trail.

CREATE TABLE audit_logs (
  id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id      UUID         NOT NULL,
  actor          VARCHAR(255) NOT NULL,               -- user/system id
  action         VARCHAR(255) NOT NULL,               -- PRD 50 audit event set (Go constants)
  resource_type  VARCHAR(255) NOT NULL,               -- workflow / instance / state / event / capability / ...
  resource_id    VARCHAR(255) NOT NULL,
  before         JSONB,                               -- state before the operation
  after          JSONB,                               -- state after the operation
  correlation_id VARCHAR(255),                        -- conversation/business correlation
  occurred_at    TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  created_at     TIMESTAMP(6) NOT NULL DEFAULT NOW()
);

-- Indexes for tenant-scoped reads (PRD 4, 96) and audit queries by action/resource.
CREATE INDEX audit_logs_tenant_action_idx       ON audit_logs (tenant_id, action);
CREATE INDEX audit_logs_tenant_resource_idx     ON audit_logs (tenant_id, resource_type, resource_id);
CREATE INDEX audit_logs_tenant_occurred_at_idx  ON audit_logs (tenant_id, occurred_at);

-- +goose Down
DROP INDEX IF EXISTS audit_logs_tenant_occurred_at_idx;
DROP INDEX IF EXISTS audit_logs_tenant_resource_idx;
DROP INDEX IF EXISTS audit_logs_tenant_action_idx;
DROP TABLE IF EXISTS audit_logs;
