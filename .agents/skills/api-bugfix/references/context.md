# Context: API Bugfix

## Target Folders

```
apps/api/internal/interfaces/http/     → routes, controllers, middleware
apps/api/internal/application/         → services, use-cases, DTOs
apps/api/internal/domain/              → entities and repository interfaces
apps/api/internal/infrastructure/      → database / external implementations
apps/api/db/queries/                   → sqlc .sql source files
apps/api/internal/infrastructure/db/   → sqlc generated (DO NOT EDIT)
packages/types/                        → shared response types
docs/openapi/                          → split OpenAPI source of truth
docs/openapi.json                      → generated merged spec
```

## Impact Map

Check in this order:
1. Is the bug in request binding or middleware?
2. Is the bug in orchestration / response mapping?
3. Is the bug in a use-case business rule?
4. Is the bug in repository / side effect or sqlc query?
5. Does the user-visible endpoint behavior change?

## Key Patterns

- Layer boundaries must stay clean during a bugfix
- Minimal touch beats broad refactor
- Response or error contract changes trigger an audit of `packages/types` and OpenAPI
- sqlc query changes require `sqlc generate` — never edit generated files
- Edit OpenAPI in split files under `docs/openapi/`

## Active Surface Examples

- `apps/api/internal/interfaces/http/routes/routes.go`
- `apps/api/internal/application/services/`
- `apps/api/internal/application/use-cases/`
- `apps/api/internal/interfaces/http/middleware/error_handler.go`
- `docs/openapi/base.json`
