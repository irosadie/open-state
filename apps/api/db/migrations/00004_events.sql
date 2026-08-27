-- +goose Up
-- Event system persistence (PRD §3.3, §27-32, §51-52, §65-66, §128) — the durable
-- driver of state transitions. Builds on the workflow runtime schema (00003):
-- events.workflow_instance_id references workflow_instances. Four tenant-isolated
-- tables: immutable event history, inbound dedup inbox, reliable emit outbox, and
-- the idempotency ledger.

-- Append-only immutable event history (PRD §27, §51). Never updated/deleted during
-- normal operation; supports replay in deterministic sequence order (PRD §32, §52).
CREATE TABLE events (
  id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id            UUID         NOT NULL,
  event_id             VARCHAR(255) NOT NULL,          -- logical id, unique per tenant (PRD §27)
  type                 VARCHAR(255) NOT NULL,          -- e.g. payment.success
  source               VARCHAR(32)  NOT NULL,          -- USER/LLM/MCP/WEBHOOK/SYSTEM/SCHEDULER/ADMIN/API (PRD §28)
  aggregate_id         VARCHAR(255),
  workflow_instance_id UUID         REFERENCES workflow_instances(id) ON DELETE CASCADE,
  sequence             BIGSERIAL    NOT NULL,          -- monotonic per-tenant auto-increment
  timestamp            TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  payload              JSONB        NOT NULL DEFAULT '{}',
  correlation_id       VARCHAR(255),
  causation_id         VARCHAR(255),
  idempotency_key      VARCHAR(255),
  created_at           TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT events_tenant_event_id_key UNIQUE (tenant_id, event_id)
);

-- Inbound external-event dedup/queue (PRD §66). Unique idempotency_key per tenant
-- prevents double processing (PRD §30, §66).
CREATE TABLE event_inbox (
  id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id        UUID         NOT NULL,
  idempotency_key  VARCHAR(255) NOT NULL,
  event_type       VARCHAR(255) NOT NULL,
  source           VARCHAR(32)  NOT NULL,              -- USER/LLM/MCP/WEBHOOK/SYSTEM/SCHEDULER/ADMIN/API
  payload          JSONB        NOT NULL DEFAULT '{}',
  status           VARCHAR(32)  NOT NULL DEFAULT 'RECEIVED', -- RECEIVED/PROCESSING/PROCESSED/FAILED
  attempt_count    INT          NOT NULL DEFAULT 0,
  received_at      TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  processed_at     TIMESTAMP(6),
  created_at       TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT event_inbox_tenant_idempotency_key UNIQUE (tenant_id, idempotency_key)
);

-- Outbound reliable delivery queue (PRD §65). Written atomically with the DB state
-- change it accompanies; publisher transitions PENDING → PUBLISHED (→ FAILED on retry).
CREATE TABLE event_outbox (
  id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id     UUID         NOT NULL,
  event_id      UUID         REFERENCES events(id) ON DELETE CASCADE,
  payload       JSONB        NOT NULL DEFAULT '{}',
  topic         VARCHAR(255) NOT NULL,
  status        VARCHAR(32)  NOT NULL DEFAULT 'PENDING', -- PENDING/PUBLISHED/FAILED
  attempt_count INT          NOT NULL DEFAULT 0,
  published_at  TIMESTAMP(6),
  created_at    TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMP(6) NOT NULL DEFAULT NOW()
);

-- Idempotency ledger (PRD §30): durable record of processed-event outcomes keyed by
-- idempotency_key so repeated external deliveries are skipped.
CREATE TABLE idempotency_records (
  id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id      UUID         NOT NULL,
  idempotency_key VARCHAR(255) NOT NULL,
  scope          VARCHAR(64)  NOT NULL DEFAULT 'event',
  result_status  VARCHAR(32)  NOT NULL,                -- PROCESSED/IGNORED/FAILED
  payload        JSONB,
  created_at     TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT idempotency_records_tenant_key UNIQUE (tenant_id, idempotency_key)
);

-- Indexes for deterministic replay order (PRD §32) and fast worker scans (PRD §65-66).
CREATE INDEX events_workflow_instance_id_sequence_idx ON events (workflow_instance_id, sequence);
CREATE INDEX events_tenant_id_idx ON events (tenant_id);
CREATE INDEX event_inbox_tenant_id_status_idx ON event_inbox (tenant_id, status);
CREATE INDEX event_outbox_tenant_id_status_idx ON event_outbox (tenant_id, status);
CREATE INDEX idempotency_records_tenant_id_idx ON idempotency_records (tenant_id);

-- +goose Down
DROP INDEX IF EXISTS idempotency_records_tenant_id_idx;
DROP INDEX IF EXISTS event_outbox_tenant_id_status_idx;
DROP INDEX IF EXISTS event_inbox_tenant_id_status_idx;
DROP INDEX IF EXISTS events_tenant_id_idx;
DROP INDEX IF EXISTS events_workflow_instance_id_sequence_idx;
DROP TABLE IF EXISTS idempotency_records;
DROP TABLE IF EXISTS event_outbox;
DROP TABLE IF EXISTS event_inbox;
DROP TABLE IF EXISTS events;
