# Checklist: DB sqlc Schema

- [ ] Read `.agents/settings.json`
- [ ] Read `references/context.md`
- [ ] Read `apps/api/sqlc.yaml` before any change
- [ ] Migration file created with sequential filename (`000XX_description.sql`)
- [ ] `-- +goose Up` and `-- +goose Down` blocks present
- [ ] Table naming: prefix matches category (`master_`, `business_`, `member_`, etc.)
- [ ] All columns in snake_case
- [ ] Boolean fields prefixed `is_` or `has_`
- [ ] Standard columns present: `id` (UUID), `created_at`, `updated_at`
- [ ] FK constraints use `REFERENCES` with appropriate `ON DELETE` action
- [ ] Indexes created for FK columns and frequently queried fields
- [ ] sqlc query file created/updated in `db/queries/`
- [ ] `sqlc generate` run — no errors
- [ ] Domain entity updated if shape changed
- [ ] Repository interface updated if new queries added
- [ ] pgx repository implementation updated
- [ ] `go build ./...` passes
- [ ] Every file ends with a newline (EOF)
