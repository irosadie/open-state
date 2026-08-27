# Context: Web Bugfix

## Target Folder

```
apps/web/app/                          → page wrapper + main content component
apps/web/hooks/transactions/           → query / mutation hooks
apps/web/components/                   → reusable components when truly the bug source
apps/web/constants/                    → api routers + query keys
apps/web/services/axios/               → interceptor / response handling
packages/schemas/                      → shared payload schema
packages/types/                        → shared response types
```

## Impact Map

Check in this order:
1. Is the bug pure UI state?
2. Is the bug in a hook or request handling?
3. Is the bug from schema/type drift?
4. Is the bug actually a contract/backend issue?

## Key Patterns

- minimum touch beats cosmetic cleanup
- `page.tsx` stays thin
- data flow goes through transaction hooks
- update `packages/schemas` / `packages/types` only when actually impacted
- a pure frontend bug does not automatically require an OpenAPI update

## Active Surface Examples

- `apps/web/app/`
- `apps/web/hooks/transactions/`
- `apps/web/constants/api-routers.ts`
- `packages/types/error-response.ts`
