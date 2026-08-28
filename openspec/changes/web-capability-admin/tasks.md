## 1. Shared contracts (Skill: web-api-integrated)

- [x] 1.1 Read `.agents/settings.json`, `.agents/guides/shared-schema.md`, `.agents/guides/web-type.md`
- [x] 1.2 Create `packages/schemas/capability.ts` — Zod schemas: create/update capability, binding, test-invocation; enums for provider_type, status, scope_type, permission (shared `as const` + `z.enum`)
- [x] 1.3 Create `packages/types/capability-response.ts` — Capability, CapabilityBinding, InvocationResult, CapabilityError response types
- [x] 1.4 Add `packages/schemas/capability.test.ts` — schema validation unit tests

## 2. Constants (Skill: web-api-integrated)

- [x] 2.1 Add capability routes to `apps/web/constants/api-routers.ts` (list/get/create/update/delete/bindings/test)
- [x] 2.2 Add capability query keys to `apps/web/constants/query-keys.ts`

## 3. Data hooks (Skill: web-api-integrated)

- [x] 3.1 Read `.agents/examples/web-api-integrated/hooks/use-examples/` hook pattern and `.agents/guides/web-hook.md`
- [x] 3.2 Create `apps/web/hooks/transactions/use-capability/use-list-capabilities.ts` (useQuery, tenant list + filter)
- [x] 3.3 Create `use-get-capability.ts` (useQuery by id)
- [x] 3.4 Create `use-create-capability.ts`, `use-update-capability.ts`, `use-delete-capability.ts` (useMutation; invalidate list on success)
- [x] 3.5 Create `use-list-bindings.ts`, `use-create-binding.ts`, `use-delete-binding.ts`
- [x] 3.6 Create `use-test-capability.ts` (useMutation → test endpoint)
- [x] 3.7 Export all from `apps/web/hooks/transactions/use-capability/index.ts`

## 4. List & create UI (Skill: web-slicing)

- [x] 4.1 Read `.agents/guides/web-page.md`, `.agents/guides/web-component.md`
- [x] 4.2 Create `apps/web/app/admin/capabilities/page.tsx` (thin Suspense wrapper) + `capabilities-page-content.tsx` (Client orchestrator: list, filter, table)
- [x] 4.3 Create `_components/capability-table.tsx` (columns: name, provider, status, version, actions) + `_components/capability-form.tsx` (create dialog/drawer)
- [x] 4.4 Wire create mutation + inline validation/conflict errors; disable action with confirmation

## 5. Detail, bindings & test UI (Skill: web-slicing)

- [x] 5.1 Create `apps/web/app/admin/capabilities/[id]/page.tsx` + content (detail, bindings panel, test panel)
- [x] 5.2 Create `_components/capability-detail.tsx` (read-only fields; show only `credential_reference`, never secrets — PRD §61)
- [x] 5.3 Create `_components/bindings-panel.tsx` (list + create + delete bindings; scope type/id + permission)
- [x] 5.4 Create `_components/test-invocation-panel.tsx` (payload form → result + sandbox/mock badge; on failure show classified error kind/code, never raw provider error — PRD §63, §2064)
- [x] 5.5 Create `_components/capability-edit-form.tsx` (edit via update mutation)

## 6. Integration & polish (Skills: web-slicing, web-api-integrated)

- [x] 6.1 Wire all mutations/query hooks to the sliced UI components (no direct axios/fetch in components)
- [x] 6.2 Add route navigation/link from capability list to detail/edit
- [x] 6.3 Empty/loading/error states for list, detail, bindings, test panels
- [x] 6.4 Run `bun run test` in `apps/web`; add/adjust smoke tests if the repo has them

## 7. Quality gate (Skill: web-code-review)

- [x] 7.1 `bun run test` and `bun run build` pass in `apps/web`
- [x] 7.2 Biome clean: no `any`, no `console.*`, no unused imports/vars, 2-space indent, double quotes, no semicolons (read root `biome.json`)
- [x] 7.3 No JSX calling axios/fetch directly (all data via hooks)
- [x] 7.4 `web-code-review` passes on the capability admin pages/hooks
- [x] 7.5 All files end with a newline (EOF)
