# Context: API Code Review

## Target Folders

```
apps/api/internal/interfaces/http/     → routes, controllers, middleware
apps/api/internal/application/         → services, use-cases, DTOs
apps/api/internal/domain/              → entities and repository interfaces
apps/api/internal/infrastructure/      → database / external implementations
apps/api/internal/infrastructure/db/   → sqlc generated (DO NOT EDIT)
packages/go-shared/domain/             → DomainError (shared)
packages/types/                        → shared response types
docs/openapi/                          → split OpenAPI source of truth
```

## Key Patterns

- routes must not jump straight to a repository
- controllers must not hold business logic
- use cases must not know about HTTP or Echo concerns
- use cases return `*domain.DomainError` — never `echo.HTTPError`
- DTO/response shape changes trigger a contract drift audit
- OpenAPI and shared types are part of the review when endpoint behavior changes
- sqlc-generated files must never be edited manually

## Active Surface Examples

- Routes: `apps/api/internal/interfaces/http/routes/routes.go`
- App assembly: `apps/api/internal/interfaces/http/create_app.go`
- Error handler: `apps/api/internal/interfaces/http/middleware/error_handler.go`
- Active OpenAPI: `docs/openapi/base.json`
