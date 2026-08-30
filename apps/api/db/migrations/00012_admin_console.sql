-- +goose Up
-- Tenant profile persistence for the Admin Console. Existing role assignments
-- already carry tenant ids; this table supplies the editable profile.
CREATE TABLE tenants (
  id          UUID         PRIMARY KEY,
  name        VARCHAR(255) NOT NULL,
  slug        VARCHAR(255) NOT NULL UNIQUE,
  description TEXT         NOT NULL DEFAULT '',
  created_at  TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMP(6) NOT NULL DEFAULT NOW()
);

-- Keep local development usable without introducing tenant-creation APIs.
INSERT INTO tenants (id, name, slug, description)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'Demo tenant',
  'demo',
  'Local development tenant'
)
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM tenants
WHERE id = '00000000-0000-0000-0000-000000000001';
DROP TABLE IF EXISTS tenants;
