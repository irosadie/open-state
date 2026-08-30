-- +goose Up

-- Durable editable graph head for the State Builder. Published snapshots remain
-- immutable in workflow_versions.definition.
ALTER TABLE workflows
  ADD COLUMN draft_definition JSONB NOT NULL DEFAULT '{}'::jsonb;

-- +goose Down
ALTER TABLE workflows
  DROP COLUMN IF EXISTS draft_definition;
