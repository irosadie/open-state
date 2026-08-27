## Context

`apps/api` is currently a Hono/TypeScript backend following Clean Architecture (entity → use-case → repository → controller → route). `apps/worker` is a BullMQ/TypeScript background worker. Both share `packages/schemas` Zod types with the frontend. The migration replaces both Go apps while keeping the frontend (`apps/web`) and shared packages untouched except for schema ownership clarification.

See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Replace `apps/api` with an Echo/Go HTTP server implementing the same auth domain
- Replace `apps/worker` with an asynq/Go worker with equivalent scaffold
- Establish sqlc + goose as the database access and migration pattern
- Clarify that `packages/schemas` is frontend-only going forward
- Keep the monorepo structure intact — Go apps live under `apps/`

**Non-Goals:**
- Changing any frontend code in `apps/web`
- Changing `packages/types` or `packages/utils`
- Adding new business features beyond what exists in the current TypeScript backend
- Setting up CI/CD pipeline changes (out of scope for this change)
- Migrating existing production data

## Decisions

### D1: Echo over Gin
Echo's centralized `HTTPErrorHandler` maps directly to the existing Hono `onError` pattern. The middleware chain API (`e.Use(...)`) is cleaner for the existing middleware stack (logger, CORS, auth). Gin requires manual `c.Abort()` calls which is more error-prone.

### D2: sqlc over GORM/Ent
Production-grade choice. sqlc generates type-safe Go from raw SQL — no runtime reflection, no hidden queries, no magic. Paired with goose for explicit migration management. GORM auto-migration is a production risk. Ent is closer to Prisma but adds unnecessary abstraction over SQL.

### D3: asynq over machinery/goworker
asynq is the most actively maintained Redis-based job queue for Go (backed by Hibiken, widely used). It has a simple API close to BullMQ semantics. Redis is already a dependency for the existing worker.

### D4: pgx/v5 as PostgreSQL driver
pgx/v5 is the most performant and featureful PostgreSQL driver for Go. It is the recommended driver for sqlc. `database/sql` adapter available if needed.

### D5: Monorepo layout — Go modules per app
Each Go app (`apps/api`, `apps/worker`) has its own `go.mod`. They do not share a Go module. This keeps Go dependency trees isolated and avoids workspace complexity. Turbo pipeline excludes Go apps from TypeScript tasks.

### D6: Folder structure mirrors Clean Architecture
```
internal/
  domain/         → entities, repository interfaces, DomainError
  application/    → use cases, app services
  infrastructure/ → sqlc DB, goose, external service impls
  interfaces/     → Echo handlers, routes, middleware
```
This preserves the same layering as the current TypeScript backend, making the migration reviewable layer-by-layer.

## Risks / Trade-offs

- **Risk: packages/types drift from Go structs** → Mitigation: OpenAPI spec (docs/openapi) is the contract. After migration, Go structs are the source of truth; `packages/types` must be manually kept in sync or generated from OpenAPI.
- **Risk: sqlc query verbosity slows initial development** → Mitigation: Accepted trade-off for production correctness. Simple CRUD queries are fast to write.
- **Risk: asynq Redis dependency** → Mitigation: Worker already depends on Redis via BullMQ. No new infrastructure required.
- **Risk: turbo.json misconfiguration breaks CI** → Mitigation: Go apps are excluded from TypeScript pipeline entirely. They are built via `go build ./...` in their own directory.

## Migration Plan

1. Create feature branch `feature/migrate-backend-to-golang`
2. Implement `apps/api` in Go (new files, do not delete TypeScript until complete)
3. Implement `apps/worker` in Go
4. Validate Go API passes existing integration test scenarios
5. Remove TypeScript `apps/api` and `apps/worker` source + dependencies
6. Update `turbo.json`, root `package.json`, `docker-compose.yml`
7. Update `packages/schemas` to remove backend-only exports
8. PR review → merge

**Rollback**: Feature branch — TypeScript `main` is untouched until merge.

## Open Questions

- Should `apps/api` and `apps/worker` share a Go workspace (`go.work`) for shared internal utilities (e.g. DomainError)? Currently scoped to separate modules — shared code would require a `packages/go-shared` module or duplication.
