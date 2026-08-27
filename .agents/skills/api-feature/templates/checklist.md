# Checklist: API Feature

- [ ] Read `.agents/settings.json`
- [ ] Read `.agents/guides/ARCHITECTURE.md` (apps/api section)
- [ ] Read `references/context.md`
- [ ] Read the guide for each folder before writing code
- [ ] DTO created in `internal/application/dtos/`
- [ ] Entity created in `internal/domain/entities/`
- [ ] Repository interface created in `internal/domain/repositories/`
- [ ] Use case(s) created in `internal/application/use-cases/` (one file per operation)
- [ ] Service created in `internal/application/services/`
- [ ] sqlc query file created in `db/queries/` and `sqlc generate` run
- [ ] pgx repository created in `internal/infrastructure/database/`
- [ ] Controller created in `internal/interfaces/http/controllers/`
- [ ] Route registered in `internal/interfaces/http/routes/routes.go`
- [ ] No `interface{}` / `any` without justification
- [ ] No business logic in Controller
- [ ] No DB/sqlc access in Use Case
- [ ] `go build ./...` passes
- [ ] All files end with a newline (EOF)
