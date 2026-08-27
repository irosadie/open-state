# Checklist: Web API Integrated

- [ ] Read `.agents/settings.json`
- [ ] Read `.agents/guides/ARCHITECTURE.md` (apps/web section)
- [ ] Read `references/context.md`
- [ ] Read the relevant API contract (OpenSpec specs, Postman, or Swagger)
- [ ] Zod schema created in `packages/schemas/` (pattern: constants + labels + helper + schema + type)
- [ ] Response type created in `packages/types/` as a flat file per domain + re-export from root if needed
- [ ] URL registered in `constants/api-routers.ts` (`:id` path variables)
- [ ] Query key registered in `constants/query-keys.ts` (flat strings, format: `{domain}{OperationPascalCase}`)
- [ ] `use-data-table.ts` hook created (react-query useQuery)
- [ ] `use-get-one.ts` hook created
- [ ] `use-insert-one.ts` hook created
- [ ] `use-update-one.ts` hook created
- [ ] `use-delete-one.ts` hook created
- [ ] `index.ts` re-exports all hooks
- [ ] Every hook: default export + named export alias
- [ ] No direct `axios`/`fetch` in components
- [ ] No `any`
- [ ] Every file ends with a newline (EOF)
