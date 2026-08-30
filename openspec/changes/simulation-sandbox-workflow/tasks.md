## 1. Domain sandbox and trace (Skill: api-feature)

- [x] 1.1 Read the backend architecture and API feature guidance; inspect the existing engine replay helpers, event pipeline, guard tests, and capability mock provider to preserve the established boundaries.
- [x] 1.2 Define domain simulation input/result/trace types, including entry state, event outcome, candidate guard outcomes, selected transition, context snapshot, rejection, and mock capability request plan.
- [x] 1.3 Extract shared, trace-aware transition-candidate evaluation so production `ProcessEvent` retains its current deterministic selection behavior while simulation can report every guard outcome.
- [x] 1.4 Implement in-memory draft simulation using fresh repositories, deterministic script event identities, supplied initial context, and stop-on-first-rejection behavior; ensure no production repository is read or written.
- [x] 1.5 Add domain tests for a successful path, priority selection, guard failure, disallowed event, initial context merge, mock capability planning, repeatability, and absence of persistent side effects.

## 2. Builder simulation API (Skill: api-feature)

- [x] 2.1 Add request/response DTOs for the snapshot, initial context, ordered events, and structured trace; validate required definition and bounded, well-formed script input.
- [x] 2.2 Add `SimulationService` to decode the Builder definition into the engine model, discard UI-only fields, assign an ephemeral sandbox project, and map domain trace values to the API contract.
- [x] 2.3 Add the workflow controller simulation action and register authenticated `POST /api/workflows/simulate`, deriving tenant exclusively from `X-Tenant-ID` and returning the standard `{ "data": ... }` envelope.
- [x] 2.4 Wire the service/controller dependency in the API composition root without changing persistent workflow or capability services.
- [x] 2.5 Add application/controller tests for an unsaved draft, missing tenant header, malformed payload, successful trace, and verification that simulation creates neither runtime nor audit data.

## 3. OpenAPI contract (Skill: docs-openapi)

- [x] 3.1 Document `POST /api/workflows/simulate`, its authenticated tenant header, input schema, trace response, and validation/error responses in the split OpenAPI files.
- [x] 3.2 Regenerate and validate the checked OpenAPI document; confirm the generated contract has no drift.

## 4. Frontend shared simulation contract (Skill: web-api-integrated)

- [x] 4.1 Add shared Zod schemas and TypeScript response types for initial context, scripted events, trace steps, guard outcomes, capability plans, and final simulation result.
- [x] 4.2 Add the simulation API route and query/mutation key constants, then implement and export a React Query simulation mutation using the configured axios client and tenant header.
- [x] 4.3 Add unit tests for valid/invalid simulation payload schemas and response typing boundaries.

## 5. State Builder simulation experience (Skill: web-slicing)

- [x] 5.1 Add transient simulation state and actions to the State Builder store: form data, request state, result, selected trace step, canvas focus target, and stale-result handling; keep it out of draft persistence and export/import.
- [x] 5.2 Add a toolbar `Simulate` action and an accessible simulation panel with initial-context JSON input, ordered event/payload editor, inline JSON errors, run/reset actions, and loading/error feedback.
- [x] 5.3 Render the returned entry/result timeline with state, event, guard outcome, selected or rejected transition, resulting context, and visibly mock-only capability requests.
- [x] 5.4 Connect trace-step selection to React Flow node/edge focus/highlighting, and mark the trace stale when workflow metadata, nodes, edges, transitions, undo/redo, import, reset, or new-workflow actions change the snapshot.
- [x] 5.5 Add focused frontend tests for JSON validation, running an unsaved canvas, successful/rejected trace display, mock labels, step-to-canvas focus, and stale-result behavior.

## 6. Quality gate

- [x] 6.1 Run `go build ./...`, `go vet ./...`, and `go test ./...` in `apps/api`; fix all regressions.
- [x] 6.2 Run `bun run test` and `bun run build` in `apps/web`; fix all regressions and Biome findings.
- [x] 6.3 Review that simulation never invokes live MCP/LLM/providers, writes no runtime/audit data, and shares rather than duplicates guard/priority semantics.
- [x] 6.4 Run `openspec validate simulation-sandbox-workflow --strict` and resolve every validation issue before implementation begins.
