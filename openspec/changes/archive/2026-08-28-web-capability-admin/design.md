## Context

Epic #4 (MCP & Integrasi) needs a tenant-scoped admin UI to manage the Capability
Registry, bindings, and sandbox testing. The backend contract is defined in
`capability-admin-api`. This slice builds the Next.js App Router admin pages consuming
that contract. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- List/filter, create, view, edit, and disable capabilities.
- Manage bindings (tenant/workflow/state).
- Test-invoke a capability in mock mode with clear sandbox indication.
- Follow the repo's FE data flow: page-content → react-query hooks → axios → API.

**Non-Goals:**
- Backend endpoints (in `capability-admin-api`).
- Policy management UI (later slice).
- MCP server/tools (in `mcp-server-runtime`).

## Decisions

### D1. Route structure (web-slicing)
```
apps/web/app/admin/capabilities/
├── page.tsx                    → thin Suspense wrapper
├── capabilities-page-content.tsx → list + filter + create/disable (Client)
└── _components/
    ├── capability-table.tsx
    ├── capability-form.tsx     (dialog/drawer)
    ├── capability-detail.tsx
    ├── bindings-panel.tsx
    └── test-invocation-panel.tsx
```
Detail/edit and test live under `capabilities/[id]/`. Components used only here stay in
`_components/`; reused UI primitives come from `apps/web/components/` (PRD §75 React Flow
is optional; this admin uses tables/forms, not the graph builder).

### D2. Data layer (web-api-integrated)
- `packages/schemas/capability.ts` — Zod schemas (`createCapabilitySchema`,
  `updateCapabilitySchema`, `bindingSchema`, `testInvocationSchema`) with enums for
  provider type, status, scope type, permission (shared from `packages/schemas`).
- `packages/types/capability-response.ts` — API response types.
- `apps/web/constants/api-routers.ts` + `query-keys.ts` — capability entries.
- Hooks under `apps/web/hooks/transactions/use-capability/` using react-query +
  `services/axios` — mutations invalidate list/query keys on success.

### D3. Forms
`react-hook-form` + `zodResolver` with the shared schemas. Errors (validation 422,
conflict 409, not-found 404) surface inline from the error response mapping.

### D4. Secret-safe
Forms collect only `credential_reference`. Detail shows it as a reference; no secret
field exists anywhere in the UI (PRD §61, §91).

### D5. Test panel
The test-invocation panel calls `use-test-capability`, renders the normalized result
(duration, data) with an explicit sandbox/mock badge, and on failure shows the
classified error kind/code — never a raw provider error (PRD §63, §2064).

## Risks / Trade-offs

- [Large capability list] → Server/pagination or client filter; start with client filter
  + tenant-scoped list.
- [Schema JSONB editing in UI] → Edit as raw JSON text validated by the shared schema
  before submit.
- [Mock test misread as live] → Prominent sandbox badge and explanatory copy.

## Migration Plan

- Land after `capability-admin-api` provides the endpoints.
- Add route + hooks + shared contracts in one change.
- Rollback: remove the route pages and hooks; no server change.

## Open Questions

None — the backend contract is fixed by `capability-admin-api`.
