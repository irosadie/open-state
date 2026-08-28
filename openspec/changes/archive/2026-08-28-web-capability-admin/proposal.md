## Why

Epic **#4 (MCP & Integrasi)** needs a human-facing admin surface to manage the
Capability Registry and its bindings, and to test capability execution in sandbox mode.
The backend contract is provided by `capability-admin-api`. This fourth slice —
**web-capability-admin** — builds the admin UI in `apps/web`: list, create, edit,
delete capabilities, manage tenant/project/workflow/state bindings, and test-invoke a
capability in mock mode.

## What Changes

- **New admin route(s)** under `apps/web/app/admin/capabilities/`:
  - `capabilities/page.tsx` + `capabilities-page-content.tsx` — list + filter registry.
  - `capabilities/[id]/page.tsx` + content — detail: view capability, bindings,
    test-invocation panel.
  - `capabilities/new` (or a drawer/dialog) — create capability.
  - `capabilities/[id]/edit` — edit capability (description, provider, schema, status,
    version, credential_reference).
- **Shared contracts**:
  - `packages/schemas/capability.ts` — Zod schemas for create/update/binding/test forms.
  - `packages/types/capability-response.ts` — API response types.
  - `apps/web/constants/api-routers.ts` + `query-keys.ts` — capability router/query keys.
- **Data hooks** under `apps/web/hooks/transactions/use-capability/`:
  - `use-list-capabilities`, `use-get-capability`, `use-create-capability`,
    `use-update-capability`, `use-delete-capability`, `use-list-bindings`,
    `use-create-binding`, `use-delete-binding`, `use-test-capability` (react-query +
    axios via `services/axios`).
- **Private route components** under `apps/web/app/admin/capabilities/_components/`:
  capability table, capability form (dialog/drawer), bindings panel, test-invocation
  result panel.
- **Secret-safe UI** — only `credential_reference` is shown/edited; never secrets.

## Capabilities

### New Capabilities

- `capability/admin-ui`: a tenant- and project-scoped admin UI to register, list, view, edit, and
  disable capabilities.
- `capability/binding-ui`: an admin UI to create/list/delete capability bindings across
  tenant, workflow, and state scopes.
- `capability/test-ui`: an admin UI to test-invoke a capability in sandbox/mock mode and
  view the normalized result.

### Modified Capabilities

- None (new capabilities introduced by this epic).

## Impact

- **`apps/web/app/admin/capabilities/`** — new route pages + `_components/`.
- **`apps/web/hooks/transactions/use-capability/`** — new react-query hooks.
- **`apps/web/constants/api-routers.ts`**, `query-keys.ts` — add capability entries.
- **`packages/schemas/capability.ts`** — new shared Zod schemas.
- **`packages/types/capability-response.ts`** — new response types.
- **`apps/web/components/`** — reuse existing UI primitives where possible; add only if
  missing.
- **No** backend, worker, docker changes in this proposal.
- Quality gate: `bun run test`, `bun run build` in `apps/web`; Biome clean.

## Non-Goals

- The backend HTTP endpoints — in `capability-admin-api`.
- Capability execution internals — in `mcp-capability-execution`.
- MCP server/tools for LLM — in `mcp-server-runtime`.
- Policy management UI (later slice).

## Dependencies

- `capability-admin-api` (HTTP contract consumed by the hooks).
- `persistence-capabilities-policies`, `mcp-capability-execution` (transitively; not
  touched here).
- Epic #4.

## Notes

- The `web-slicing` skill covers UI implementation; `web-api-integrated` covers schema,
  types, constants, and hooks wiring; `web-code-review` covers final review.
