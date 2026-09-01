## 1. OpenSpec and backend contract

- [x] 1.1 Update the project discovery route and OpenAPI description to use
  `workflow:read` as the shared read permission.
- [x] 1.2 Validate the complete `admin-project-flow` change in strict mode.

## 2. Project inventory and navigation

- [x] 2.1 Add `/admin/projects` with tenant-scoped project cards, downstream
  Intent/Workflow links, and loading/error/empty/unauthorized states.
- [x] 2.2 Add Projects to the Admin Console shell and overview, and make the
  Project step in `AdminFlowGuide` a real link.
- [x] 2.3 Add route-policy and component tests for Project visibility and the
  clickable flow step.

## 3. Propagate project scope

- [x] 3.1 Read `projectId` in Intent and Workflow pages and pass it to list
  queries and Builder destinations.
- [x] 3.2 Pass `projectId` through workflow creation and State Builder page
  props into the store.
- [x] 3.3 Preserve project scope in Builder load, save, version, and publish
  API requests.

## 4. Verification

- [x] 4.1 Run frontend tests, typecheck, lint, and production build.
- [x] 4.2 Run backend tests, vet, build, OpenAPI validation, and `git diff --check`.
