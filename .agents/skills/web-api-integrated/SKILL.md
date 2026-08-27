---
name: web-api-integrated
description: Integrate API endpoints into the frontend with Zod schemas, response types, constants, and react-query hooks. Use when implementing endpoints from an API contract, Postman, or Swagger.
---

# Skill: Web API Integrated

## Context (Required)
- Folder scope + code examples: `references/context.md`
- Execution checklist: `templates/checklist.md`

Integrate APIs into the frontend with a consistent architecture.

## Workflow

### 0. Read the API Contract
- If the user provides an API contract (Postman, Swagger, OpenSpec specs), read and understand every endpoint: method, URL, payload, and response.
- Check `openspec/changes/{slug}/specs/` for OpenSpec-based contracts.

### 1. Create Zod Schema (`packages/schemas/`)
- Follow the pattern: type constants + labels array + `get{Type}Label()` + Zod schema + type alias
- Read guide: `.agents/guides/shared-schema.md`

### 2. Create Response Types (`packages/types/`)
- One file per response type, e.g. `packages/types/payment-method-response.ts`
- Import from frontend via `@vibecoding-starter/types/payment-method-response`
- Re-export from `packages/types/index.ts` if needed from the root package

### 3. Register Constants
- Add URL to `apps/web/constants/api-routers.ts` (use `:id` path variables)
- Add query key to `apps/web/constants/query-keys.ts` (flat strings)
- Read guide: `.agents/guides/web-constant.md`

### 4. Create Custom Hooks (`apps/web/hooks/transactions/use-{domain}/`)
- One folder per domain, one file per operation
- Hooks call axios **directly** — no service function required
- File order: `use-data-table.ts`, `use-get-one.ts`, `use-insert-one.ts`, `use-update-one.ts`, `use-delete-one.ts`, `index.ts`
- Read guide: `.agents/guides/web-hook.md`
- See example: `.agents/examples/web-api-integrated/hooks/use-examples/`

## Prohibitions

- **FORBIDDEN** to call `axios`/`fetch` directly inside JSX/TSX components.
- **FORBIDDEN** to use `any` as a type.
- **FORBIDDEN** to skip Zod schema — every payload must be validated.
- **FORBIDDEN** to hardcode API URLs — use `apiRouters` from constants.
- **FORBIDDEN** to use functions for URL building — use `:id` path variable + `pathVariable()`.

## Pre-Completion Checklist

- [ ] Zod schema created in `packages/schemas/` (full pattern: constants + labels + helper + schema + type)
- [ ] Response type created in `packages/types/` as a flat file per domain + re-export from root if needed
- [ ] URL registered in `constants/api-routers.ts` (`:id` path variables)
- [ ] Query key registered in `constants/query-keys.ts` (flat strings)
- [ ] Custom hooks created: `useDataTable`, `useGetOne`, `useInsertOne`, `useUpdateOne`, `useDeleteOne`
- [ ] `index.ts` re-exports all hooks
- [ ] No direct `axios`/`fetch` in components
- [ ] No `any`
- [ ] Every file ends with a newline (EOF)
