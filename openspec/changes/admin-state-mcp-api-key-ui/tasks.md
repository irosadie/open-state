## 1. Frontend API contract

- [x] 1.1 Add State MCP API-key request and response Zod schemas with scope
  constants and labels.
- [x] 1.2 Add API-key response types and package exports.
- [x] 1.3 Register API-key URLs and query keys.
- [x] 1.4 Add list, create, and revoke React Query hooks using the configured
  tenant header and response validation.
- [x] 1.5 Add the tenant-scoped project response contract and discovery hook for
  the API-key selector.

## 1A. Backend project discovery

- [x] 1A.1 Add the project list DTO, service, controller, and authenticated
  `GET /api/projects` route using the existing tenant-scoped repository.
- [x] 1A.2 Document the project discovery endpoint and add service/route tests.

## 2. Admin Console surface

- [x] 2.1 Add `/admin/api-keys` with the established thin page wrapper and
  client page content.
- [x] 2.2 Add create form for name, tenant-project multi-select, constrained
  default project select, scopes, and optional expiry.
- [x] 2.3 Show the one-time raw key with copy and dismiss actions; never show a
  raw key from list data.
- [x] 2.4 Render safe metadata, revoke confirmation, loading, empty, error,
  retry, and unauthorized states.
- [x] 2.5 Add permission-aware overview card, sidebar item, route policy, and
  action policies for API-key operations.

## 3. Verification

- [x] 3.1 Add schema, hook, and page tests for successful and denied flows.
- [x] 3.2 Run frontend tests, typecheck, lint, and production build.
- [ ] 3.3 Validate the OpenSpec change and perform an authenticated local smoke
  test against `/api/api-keys`.
