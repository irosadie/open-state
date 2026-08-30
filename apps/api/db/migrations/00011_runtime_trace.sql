-- +goose Up
-- Product-level, append-only runtime trace projection (PRD 142-143, 170).
-- Provider systems remain independent; only sanitized application-observed
-- metadata is persisted here.
CREATE TABLE runtime_trace_entries (
  id                    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id             UUID         NOT NULL,
  workflow_instance_id  UUID         NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
  turn_id               VARCHAR(255),
  sequence              BIGSERIAL    NOT NULL,
  stage                 VARCHAR(64)  NOT NULL,
  source                VARCHAR(32)  NOT NULL,
  status                VARCHAR(32)  NOT NULL,
  occurred_at           TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  correlation_id        VARCHAR(255),
  duration_ms           BIGINT,
  reason_code            VARCHAR(128),
  error_code             VARCHAR(128),
  provider_alias         VARCHAR(255),
  provider_reference     VARCHAR(255),
  summary                VARCHAR(1000),
  attributes             JSONB        NOT NULL DEFAULT '{}'
);

CREATE INDEX runtime_trace_entries_tenant_instance_sequence_idx
  ON runtime_trace_entries (tenant_id, workflow_instance_id, sequence);
CREATE INDEX runtime_trace_entries_tenant_instance_turn_sequence_idx
  ON runtime_trace_entries (tenant_id, workflow_instance_id, turn_id, sequence);

-- +goose Down
DROP INDEX IF EXISTS runtime_trace_entries_tenant_instance_turn_sequence_idx;
DROP INDEX IF EXISTS runtime_trace_entries_tenant_instance_sequence_idx;
DROP TABLE IF EXISTS runtime_trace_entries;
