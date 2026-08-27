# Context: DB sqlc Schema

## Target Folders

```
apps/api/
├── db/
│   ├── migrations/   → goose SQL migration files (000XX_description.sql)
│   └── queries/      → sqlc .sql query files ({domain}.sql)
├── internal/
│   └── infrastructure/
│       └── db/       → sqlc generated Go code (DO NOT EDIT)
└── sqlc.yaml         → sqlc configuration
```

## Migration File Pattern

```sql
-- +goose Up
CREATE TABLE master_categories (
  id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  name       VARCHAR(100) NOT NULL,
  is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP(6) NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP(6) NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS master_categories;
```

## sqlc Query File Pattern

```sql
-- name: FindCategoryByID :one
SELECT id, name, is_active, created_at, updated_at
FROM master_categories
WHERE id = $1;

-- name: CreateCategory :one
INSERT INTO master_categories (name, is_active)
VALUES ($1, $2)
RETURNING id, name, is_active, created_at, updated_at;
```

## Commands

```bash
# Run from apps/api/
sqlc generate

# Run migrations (from apps/api/)
goose -dir db/migrations postgres "$DATABASE_URL" up
goose -dir db/migrations postgres "$DATABASE_URL" down
goose -dir db/migrations postgres "$DATABASE_URL" status
```

## sqlc.yaml Reference

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "db/queries/"
    schema: "db/migrations/"
    gen:
      go:
        package: "db"
        out: "internal/infrastructure/db"
        emit_json_tags: true
```
