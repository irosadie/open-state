---
name: api-feature
description: Implement new backend features following Clean Architecture. Use for tasks involving new endpoints, use cases, entities, repositories, or application-layer changes.
---

# Skill: API Feature

## Context (Required)
- Folder scope + code examples: `references/context.md`
- Execution checklist: `templates/checklist.md`

Implement backend features professionally, following Clean Architecture.

## Workflow

1. Read the API contract or requirement provided.

2. Read the guide for **each folder** before creating files there:
   - `.agents/guides/api-dto.md`
   - `.agents/guides/api-entity.md`
   - `.agents/guides/api-repository.md`
   - `.agents/guides/api-usecase.md`
   - `.agents/guides/api-service.md`
   - `.agents/guides/api-db-repository.md`
   - `.agents/guides/api-controller.md`
   - `.agents/guides/api-route.md`
   - `.agents/guides/api-error.md`

3. Create files **in order** by layer dependency:
   ```
   1. internal/application/dtos/{domain}.go
   2. internal/domain/entities/{domain}.go
   3. internal/domain/repositories/i_{domain}_repository.go
   4. internal/domain/use-cases/{verb}_{domain}.go   (one per operation)
   5. internal/application/services/{domain}_service.go
   6. internal/infrastructure/database/pgx_{domain}_repository.go
   7. internal/interfaces/http/controllers/{domain}_controller.go
   8. internal/interfaces/http/routes/routes.go       (register new routes)
   9. db/queries/{domain}.sql                         (sqlc query file)
   10. Run: sqlc generate
   ```

4. Error handling — do not catch in Service or Controller:
   ```
   UseCase returns DomainError → Echo ErrorHandler middleware
   ```

## Prohibitions

- **NEVER** use `interface{}` or `any` without explicit justification.
- **NEVER** put business logic in the Controller.
- **NEVER** access the database directly in a Use Case — go through repository interface.
- **NEVER** return `echo.HTTPError` from a Use Case — use `domain.DomainError` from `packages/go-shared`.
- **NEVER** change files unrelated to the task.
- **NEVER** use plain `string` for fields with a fixed value set — define as typed Go constant.
- **NEVER** edit sqlc-generated files in `internal/infrastructure/db/` — regenerate instead.

## Pre-Completion Checklist

- [ ] DTO created
- [ ] Entity created
- [ ] Repository interface created
- [ ] Use case(s) created (one per operation)
- [ ] Service created
- [ ] sqlc query file created and `sqlc generate` run
- [ ] pgx repository implementation created
- [ ] Controller created
- [ ] Route registered in `routes.go`
- [ ] No `interface{}` / `any` without justification
- [ ] No business logic in Controller
- [ ] No DB access in Use Case
- [ ] `go build ./...` passes
- [ ] All files end with a newline (EOF)
