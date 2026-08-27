## 1. Monorepo & Go Workspace Setup

- [x] 1.1 Create `go.work` at repo root referencing `apps/api`, `apps/worker`, and `packages/go-shared`
- [x] 1.2 Create `packages/go-shared/` Go module (`module github.com/vibecoding-starter/go-shared`) with `go.mod`
- [x] 1.3 Define `DomainError` struct + error codes + constructor helpers in `packages/go-shared/domain/errors.go`
- [x] 1.4 Update `turbo.json` to exclude Go apps from TypeScript pipeline tasks
- [x] 1.5 Update root `package.json` scripts to add `go:build` and `go:test` commands for both apps

## 2. Database Layer (sqlc + goose)

- [x] 2.1 Create `apps/api/db/migrations/` and write initial goose migration SQL for `users` and `auth_sessions` tables (matching current Prisma schema)
- [x] 2.2 Create `apps/api/db/queries/` with sqlc `.sql` query files for all auth operations (find user by email, create user, create session, find session by token hash, delete session)
- [x] 2.3 Create `apps/api/sqlc.yaml` config pointing to `db/queries/` and `db/migrations/` schema
- [x] 2.4 Run `sqlc generate` and commit generated Go code to `apps/api/internal/infrastructure/db/`
- [x] 2.5 Implement `apps/api/internal/infrastructure/config/env.go` to load `DATABASE_URL`, `PORT`, `JWT_SECRET` from environment with fail-fast on missing required vars

## 3. Domain Layer (apps/api)

- [x] 3.1 Define `User` and `AuthSession` entity structs in `apps/api/internal/domain/entities/`
- [x] 3.2 Define `IAuthRepository` interface in `apps/api/internal/domain/repositories/`
- [x] 3.3 Define `TokenService` and `StorageService` interfaces in `apps/api/internal/domain/services/`

## 4. Infrastructure Layer (apps/api)

- [x] 4.1 Implement `PrismaAuthRepository` → `PgxAuthRepository` in `apps/api/internal/infrastructure/database/` using sqlc-generated queries
- [x] 4.2 Implement `JwtTokenService` in `apps/api/internal/infrastructure/services/` using `golang-jwt/jwt/v5`
- [x] 4.3 Implement `LocalStorageService` in `apps/api/internal/infrastructure/services/`
- [x] 4.4 Implement `apps/api/internal/infrastructure/config/database.go` for pgx connection pool setup

## 5. Application Layer (apps/api)

- [x] 5.1 Implement `RegisterUser` use case in `apps/api/internal/application/use-cases/`
- [x] 5.2 Implement `LoginUser` use case
- [x] 5.3 Implement `LogoutUser` use case
- [x] 5.4 Implement `GetCurrentUser` use case
- [x] 5.5 Implement `GetHealth` and `GetAppInfo` use cases
- [x] 5.6 Implement `AuthService` in `apps/api/internal/application/services/` orchestrating use cases and transforming entities to DTOs
- [x] 5.7 Define auth DTOs (request/response structs) in `apps/api/internal/application/dtos/`

## 6. Interfaces Layer (apps/api)

- [x] 6.1 Implement Echo error handler in `apps/api/internal/interfaces/http/middleware/error_handler.go` mapping DomainError codes to HTTP status (import from `packages/go-shared`)
- [x] 6.2 Implement JWT auth middleware in `apps/api/internal/interfaces/http/middleware/jwt.go`
- [x] 6.3 Implement auth session middleware in `apps/api/internal/interfaces/http/middleware/auth_session.go`
- [x] 6.4 Implement `AuthController` in `apps/api/internal/interfaces/http/controllers/`
- [x] 6.5 Implement `SystemController` in `apps/api/internal/interfaces/http/controllers/`
- [x] 6.6 Register auth routes (`/api/auth/register`, `/api/auth/login`, `/api/auth/logout`, `/api/auth/me`) in `apps/api/internal/interfaces/http/routes/`
- [x] 6.7 Register health and root routes
- [x] 6.8 Implement `create_app.go` wiring Echo instance with all middleware and routes

## 7. Entrypoint & go.mod (apps/api)

- [x] 7.1 Create `apps/api/go.mod` with module name and required dependencies (echo/v4, pgx/v5, golang-jwt/jwt/v5, goose/v3, bcrypt)
- [x] 7.2 Create `apps/api/cmd/server/main.go` entrypoint: load config, connect DB, wire dependencies, start Echo server
- [x] 7.3 Run `go mod tidy` in `apps/api`
- [x] 7.4 Verify `go build ./...` passes with no errors

## 8. Worker (apps/worker)

- [x] 8.1 Create `apps/worker/go.mod` with module name and asynq dependency
- [x] 8.2 Implement asynq server setup in `apps/worker/internal/infrastructure/queue/server.go` reading `REDIS_URL`
- [x] 8.3 Implement runtime summary job handler scaffold in `apps/worker/internal/application/use-cases/`
- [x] 8.4 Create `apps/worker/cmd/worker/main.go` entrypoint: load config, register handlers, start asynq server
- [x] 8.5 Run `go mod tidy` in `apps/worker`
- [x] 8.6 Verify `go build ./...` passes with no errors

## 9. Cleanup TypeScript Backend

- [x] 9.1 Remove `apps/api/src/` TypeScript source files
- [x] 9.2 Remove Prisma-related files: `apps/api/prisma/`, `apps/api/prisma.config.ts`, `apps/api/scalar.config.json`
- [x] 9.3 Remove TypeScript deps from `apps/api/package.json` (hono, prisma, @prisma/client, etc.) or delete `package.json` entirely if Go app has no Node deps
- [x] 9.4 Remove `apps/worker/src/` TypeScript source files
- [x] 9.5 Remove BullMQ deps from `apps/worker/package.json` or delete if no Node deps remain
- [x] 9.6 Update `packages/schemas` — remove any backend-only Zod schemas, keep only frontend-used ones

## 10. Docker & Infrastructure

- [x] 10.1 Update `apps/api` Dockerfile to Go multi-stage build (build stage: `golang:1.23-alpine`, runtime: `alpine:3.20`)
- [x] 10.2 Update `apps/worker` Dockerfile to Go multi-stage build
- [x] 10.3 Verify `docker-compose.yml` service configs still work with new Go images

## 11. Verification

- [x] 11.1 Write unit tests for all 4 auth use cases in `apps/api/internal/domain/use-cases/`
- [x] 11.2 Write integration smoke test hitting `GET /health` and `POST /api/auth/register` → `POST /api/auth/login` → `GET /api/auth/me` → `POST /api/auth/logout`
- [x] 11.3 Run `go test ./...` in `apps/api` — all tests pass
- [x] 11.4 Run `go test ./...` in `apps/worker` — all tests pass
- [x] 11.5 Run `bun run build` from repo root — frontend and packages build cleanly
