## 1. Backend intent catalog HTTP contract

- [x] 1.1 Add an application DTO/projection for the HTTP intent item and list response, including `id`, `key`, `name`, `description`, `examples`, `tenantId`, `projectId`, `workflowId`, and `workflowSlug`.
- [x] 1.2 Add the application service/use-case path that resolves an explicit project or the existing tenant `default` project without creating data, then delegates to the existing published-only intent catalog repository.
- [x] 1.3 Add an HTTP controller for `GET /api/intents` with optional `projectId`, preserving the existing tenant header, authentication, error handling, and read-only semantics.
- [x] 1.4 Register the route behind the existing `workflow:read` permission and wire the intent catalog service into `CreateApp` and the PostgreSQL composition root.

## 2. Backend and API contract verification

- [x] 2.1 Add service/controller/route tests for default-project scope, explicit project scope, tenant isolation, empty results, published-only filtering, missing tenant/authentication, and denied permission.
- [x] 2.2 Document `GET /intents` and its intent item schema in `docs/openapi/paths/intents.json` and `docs/openapi/schemas/intent.json`, reusing existing error schemas.
- [x] 2.3 Register the split OpenAPI files, run `bun run openapi:generate`, and verify the merged `docs/openapi.json` contains the new endpoint and response fields.

## 3. Frontend API integration

- [x] 3.1 Add the Zod response schema for the read-only intent catalog and export its inferred type from `packages/schemas/`.
- [x] 3.2 Add `IntentResponse` to `packages/types/` and re-export it from the package entry point.
- [x] 3.3 Register the intents API route and React Query key in `apps/web/constants/`.
- [x] 3.4 Add a `use-intent` transaction hook with a list operation that sends the configured tenant id and optional project id through the existing axios/BFF path, validates the response, and exposes retry/loading/error state.

## 4. Admin Console intent surface

- [x] 4.1 Add the thin Suspense route at `apps/web/app/admin/intents/page.tsx` and the client `intents-page-content.tsx` beside it.
- [x] 4.2 Render the current tenant/project context and a read-only intent table/card list with canonical key, description, examples, mapped workflow slug, and an `Open Builder` link using `workflowId`.
- [x] 4.3 Add clear loading, empty, unauthorized, API error, and retry states without rendering intent mutation controls.
- [x] 4.4 Add the `Intents` item to permission-aware shell navigation and an overview entry card, using `workflow:read` consistently.

## 5. Clarify the Admin Console hierarchy

- [x] 5.1 Extend `AdminFlowGuide` to five typed steps in the order `Tenant → Project → Intent → Workflow → Builder`, with Intent linked to `/admin/intents` and Project remaining an explicit non-link context step.
- [x] 5.2 Render the guide's Intent current state on the intent page and update tenant/workflow/overview copy so `Default Project (automatic)` and the intent-to-workflow relationship are explicit.
- [x] 5.3 Add or update frontend tests for navigation visibility, five-step order, intent link, default-project copy, `BOOKING_PADEL` rendering, Builder link, empty state, and API error/retry behavior.

## 6. Verification and handoff

- [x] 6.1 Run formatting, `go test ./...`, `go vet ./...`, and `go build ./...` for the backend.
- [x] 6.2 Run frontend tests, typecheck, lint, and production build; verify no direct API calls or `any` types are introduced in components.
- [x] 6.3 Run OpenAPI generation/validation and perform a local authenticated smoke test showing `BOOKING_PADEL` with `saya mau order lapangan` and its Builder destination.
