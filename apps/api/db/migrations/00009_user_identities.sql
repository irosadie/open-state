-- +goose Up
-- External identity store (PRD 79): links an OIDC provider subject to a local
-- user. A user MAY have one identity per provider; the (provider, subject)
-- pair is globally unique. auto_provisioned marks users created on first login.

CREATE TABLE user_identities (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id           UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider          VARCHAR(64) NOT NULL,               -- google / github / entra
  subject_id        VARCHAR(255) NOT NULL,              -- provider sub claim
  auto_provisioned  BOOLEAN     NOT NULL DEFAULT FALSE,
  created_at        TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  CONSTRAINT user_identities_provider_subject_unique UNIQUE (provider, subject_id)
);

CREATE INDEX user_identities_user_id_idx   ON user_identities (user_id);
CREATE INDEX user_identities_provider_idx  ON user_identities (provider, subject_id);

-- +goose Down
DROP INDEX IF EXISTS user_identities_provider_idx;
DROP INDEX IF EXISTS user_identities_user_id_idx;
DROP TABLE IF EXISTS user_identities;
